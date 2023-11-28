package vault

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"io"

	"github.com/abjrcode/swervo/internal/faults"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/awnumar/memguard"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

var (
	ErrVaultAlreadyConfigured     = errors.New("vault is already configured")
	ErrVaultNotConfigured         = errors.New("vault is not configured")
	ErrVaultNotConfiguredOrSealed = errors.New("vault is not configured or sealed")
)

type Vault interface {
	// IsConfigured returns true if the vault is configured with a key, false otherwise.
	IsConfigured(ctx context.Context) (bool, error)

	// Configure configures the vault with a key derived from the given plainPassword.
	Configure(ctx context.Context, plainPassword string) error

	// Open opens the vault with the given plainPassword.
	// Allows the vault to be used for encryption and decryption.
	Open(ctx context.Context, plainPassword string) (bool, error)

	// Seal closes the vault and purges the key from memory.
	Seal()

	// Vault can be used as an encryption service.
	encryption.EncryptionService
}

type vaultImpl struct {
	timeSvc       utils.Clock
	db            *sql.DB
	logger        zerolog.Logger
	keyId         *string
	encryptionKey *memguard.Enclave
}

func NewVault(db *sql.DB, timeSvc utils.Clock, logger zerolog.Logger) Vault {
	memguard.CatchInterrupt()

	enrichedLogger := logger.With().Str("component", "vault").Logger()

	return &vaultImpl{
		timeSvc: timeSvc,
		db:      db,
		logger:  enrichedLogger,
	}
}

func (v *vaultImpl) IsConfigured(ctx context.Context) (bool, error) {
	row := v.db.QueryRowContext(ctx, `SELECT "key_id" FROM "argon_key_material";`)

	var keyId string

	err := row.Scan(&keyId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, errors.Join(err, faults.ErrFatal)
	}

	return true, nil
}

func (v *vaultImpl) Configure(ctx context.Context, plainPassword string) error {
	configured, err := v.IsConfigured(ctx)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	if configured {
		return ErrVaultAlreadyConfigured
	}

	uniqueId, err := ksuid.NewRandom()
	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	keyId := uniqueId.String()

	derivedKey, salt, err := generateFromPassword(plainPassword, DefaultParameters)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)

	encKeyHash := sha3_512Hash(derivedKey)

	_, err = v.db.ExecContext(ctx, `
	INSERT INTO "argon_key_material" (
		"key_id",
		"key_hash_sha3_512",
		"argon2_version",
		"argon2_variant",
		"created_at",
		"memory",
		"iterations",
		"parallelism",
		"salt_length",
		"salt_base64",
		"key_length"
	) VALUES (
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?
	);`, keyId, encKeyHash,
		DefaultParameters.Aargon2Version, DefaultParameters.Variant,
		v.timeSvc.NowUnix(), DefaultParameters.Memory,
		DefaultParameters.Iterations, DefaultParameters.Parallelism,
		DefaultParameters.SaltLength, saltBase64,
		DefaultParameters.KeyLength)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	v.keyId = &keyId
	v.encryptionKey = memguard.NewEnclave(derivedKey)

	return nil
}

func (v *vaultImpl) IsOpen() bool {
	return v.encryptionKey != nil
}

func (v *vaultImpl) Open(ctx context.Context, plainPassword string) (bool, error) {
	if v.IsOpen() {
		return true, nil
	}

	row := v.db.QueryRowContext(ctx, `
	SELECT
		"key_id",
		"key_hash_sha3_512",
		"argon2_version",
		"argon2_variant",
		"memory",
		"iterations",
		"parallelism",
		"salt_length",
		"salt_base64",
		"key_length"
	FROM "argon_key_material";`)

	var keyId string
	var keyHash []byte
	var saltBase64 string
	var params ArgonParameters

	err := row.Scan(&keyId, &keyHash, &params.Aargon2Version, &params.Variant, &params.Memory,
		&params.Iterations, &params.Parallelism, &params.SaltLength, &saltBase64, &params.KeyLength)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrVaultNotConfigured
		}
		return false, errors.Join(err, faults.ErrFatal)
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltBase64)

	if err != nil {
		return false, errors.Join(err, faults.ErrFatal)
	}

	match, derivedKey, err := comparePasswordAndHash(plainPassword, salt, keyHash, &params)

	if err != nil {
		return false, errors.Join(err, faults.ErrFatal)
	}

	if match {
		v.keyId = &keyId
		v.encryptionKey = memguard.NewEnclave(derivedKey)
	}

	return match, nil
}

func (v *vaultImpl) Seal() {
	v.keyId = nil
	v.encryptionKey = nil
	memguard.Purge()
}

func (v *vaultImpl) EncryptBinary(plaintext []byte) ([]byte, string, error) {
	if !v.IsOpen() {
		return nil, "", ErrVaultNotConfiguredOrSealed
	}

	key, err := v.encryptionKey.Open()
	if err != nil {
		return nil, "", errors.Join(err, faults.ErrFatal)
	}
	defer key.Destroy()

	aesBlock, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, "", errors.Join(err, faults.ErrFatal)
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, "", errors.Join(err, faults.ErrFatal)
	}

	nonce := make([]byte, gcmInstance.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	ciphertext := gcmInstance.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, *v.keyId, nil

}

func (v *vaultImpl) DecryptBinary(ciphertext []byte, keyId string) ([]byte, error) {
	if !v.IsOpen() {
		return nil, ErrVaultNotConfiguredOrSealed
	}

	if keyId != *v.keyId {
		// TODO: try lookup deprecated or old keys in the database
		return nil, ErrVaultNotConfiguredOrSealed
	}

	key, err := v.encryptionKey.Open()
	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}
	defer key.Destroy()

	aesBlock, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}
	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}

	nonceSize := gcmInstance.NonceSize()
	nonce, encryptedText := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcmInstance.Open(nil, nonce, encryptedText, nil)
	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}

	return plaintext, nil
}

func (v *vaultImpl) Encrypt(plaintext string) (string, string, error) {
	if !v.IsOpen() {
		return "", "", ErrVaultNotConfiguredOrSealed
	}

	ciphertext, keyId, err := v.EncryptBinary([]byte(plaintext))

	if err != nil {
		return "", "", errors.Join(err, faults.ErrFatal)
	}

	return string(ciphertext), keyId, nil
}

func (v *vaultImpl) Decrypt(ciphertext string, keyId string) (string, error) {
	if !v.IsOpen() {
		return "", ErrVaultNotConfiguredOrSealed
	}

	plaintext, err := v.DecryptBinary([]byte(ciphertext), keyId)

	if err != nil {
		if errors.Is(err, ErrVaultNotConfiguredOrSealed) {
			return "", err
		}

		return "", errors.Join(err, faults.ErrFatal)
	}

	return string(plaintext), nil
}
