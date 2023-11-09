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

	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/awnumar/memguard"
)

var (
	ErrVaultAlreadyConfigured     = errors.New("vault is already configured")
	ErrVaultNotConfigured         = errors.New("vault is not configured")
	ErrVaultNotConfiguredOrSealed = errors.New("vault is not configured or sealed")
)

type Vault interface {
	// IsSetup returns true if the vault is configured with a key, false otherwise.
	IsSetup(ctx context.Context) (bool, error)

	// ConfigureKey configures the vault with a key derived from the given plainPassword.
	ConfigureKey(ctx context.Context, plainPassword string) error

	// Open opens the vault with the given plainPassword.
	// Allows the vault to be used for encryption and decryption.
	Open(ctx context.Context, plainPassword string) (bool, error)

	// Seal closes the vault and purges the key from memory.
	Seal()

	// Close closes the vault and purges the key from memory.
	// Typically used when the vault is no longer needed and the application is shutting down.
	Close()

	// Vault can be used as an encryption service.
	encryption.EncryptionService
}

type vaultImpl struct {
	timeSvc       utils.Datetime
	db            *sql.DB
	keyId         *string
	encryptionKey *memguard.Enclave
}

func NewVault(db *sql.DB, timeSvc utils.Datetime) Vault {
	memguard.CatchInterrupt()

	return &vaultImpl{
		timeSvc: timeSvc,
		db:      db,
	}
}

func (v *vaultImpl) IsSetup(ctx context.Context) (bool, error) {
	row := v.db.QueryRowContext(ctx, `SELECT "key_id" FROM "argon_key_material";`)

	var keyId string

	err := row.Scan(&keyId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (v *vaultImpl) ConfigureKey(ctx context.Context, plainPassword string) error {
	configured, err := v.IsSetup(ctx)

	if err != nil {
		return err
	}

	if configured {
		return ErrVaultAlreadyConfigured
	}

	keyId := utils.RandomString(4)

	derivedKey, salt, err := generateFromPassword(plainPassword, DefaultParameters)

	if err != nil {
		return err
	}

	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)

	hash := sha256Hash(derivedKey)

	_, err = v.db.ExecContext(ctx, `
	INSERT INTO "argon_key_material" (
		"key_id",
		"key_hash_sha256",
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
	);`, keyId, hash,
		DefaultParameters.Aargon2Version, DefaultParameters.Variant,
		v.timeSvc.NowUnix(), DefaultParameters.Memory,
		DefaultParameters.Iterations, DefaultParameters.Parallelism,
		DefaultParameters.SaltLength, saltBase64,
		DefaultParameters.KeyLength)

	if err != nil {
		return err
	}

	v.keyId = &keyId
	v.encryptionKey = memguard.NewEnclave(derivedKey)

	return nil
}

func (v *vaultImpl) Open(ctx context.Context, plainPassword string) (bool, error) {
	if v.encryptionKey != nil {
		return true, nil
	}

	row := v.db.QueryRowContext(ctx, `
	SELECT
		"key_id",
		"key_hash_sha256",
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
	var keyHash string
	var saltBase64 string
	var params ArgonParameters

	err := row.Scan(&keyId, &keyHash, &params.Aargon2Version, &params.Variant, &params.Memory,
		&params.Iterations, &params.Parallelism, &params.SaltLength, &saltBase64, &params.KeyLength)

	if err != nil {
		return false, errors.Join(ErrVaultNotConfigured, err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltBase64)

	if err != nil {
		return false, err
	}

	match, derivedKey, err := comparePasswordAndHash(plainPassword, salt, []byte(keyHash), &params)

	if err != nil {
		return false, err
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

func (v *vaultImpl) Close() {
	defer memguard.Purge()
}

func (v *vaultImpl) EncryptBinary(plaintext []byte) ([]byte, string, error) {
	if v.encryptionKey == nil {
		return nil, "", ErrVaultNotConfiguredOrSealed
	}

	key, err := v.encryptionKey.Open()
	if err != nil {
		return nil, "", err
	}
	defer key.Destroy()

	aesBlock, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, "", err
	}

	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, "", err
	}

	nonce := make([]byte, gcmInstance.NonceSize())
	_, _ = io.ReadFull(rand.Reader, nonce)

	ciphertext := gcmInstance.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, *v.keyId, nil

}

func (v *vaultImpl) DecryptBinary(ciphertext []byte, keyId string) ([]byte, error) {
	if v.encryptionKey == nil {
		return nil, ErrVaultNotConfiguredOrSealed
	}

	if keyId != *v.keyId {
		// TODO: try lookup deprecated or old keys in the database
		return nil, ErrVaultNotConfiguredOrSealed
	}

	key, err := v.encryptionKey.Open()
	if err != nil {
		return nil, err
	}
	defer key.Destroy()

	aesBlock, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return nil, err
	}
	gcmInstance, err := cipher.NewGCM(aesBlock)
	if err != nil {
		return nil, err
	}

	nonceSize := gcmInstance.NonceSize()
	nonce, encryptedText := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcmInstance.Open(nil, nonce, encryptedText, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (v *vaultImpl) Encrypt(plaintext string) (string, string, error) {
	if v.encryptionKey == nil {
		return "", "", ErrVaultNotConfiguredOrSealed
	}

	ciphertext, keyId, err := v.EncryptBinary([]byte(plaintext))

	if err != nil {
		return "", "", err
	}

	return string(ciphertext), keyId, nil
}

func (v *vaultImpl) Decrypt(ciphertext string, keyId string) (string, error) {
	if v.encryptionKey == nil {
		return "", ErrVaultNotConfiguredOrSealed
	}

	plaintext, err := v.DecryptBinary([]byte(ciphertext), keyId)

	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
