package vault

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"

	"github.com/abjrcode/swervo/internal/utils"
	"github.com/awnumar/memguard"
)

var (
	ErrVaultAlreadyConfigured = errors.New("vault is already configured")
	ErrVaultNotConfigured     = errors.New("vault is not configured")
)

type Vault interface {
	IsSetup(ctx context.Context) (bool, error)
	ConfigureKey(ctx context.Context, plainPassword string) error
	Open(ctx context.Context, plainPassword string) (bool, error)
	Seal()
	Close()
	Encrypt(ctx context.Context, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error)
}

type vaultImpl struct {
	timeSvc       utils.Datetime
	db            *sql.DB
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

	v.encryptionKey = memguard.NewEnclave(derivedKey)

	return nil
}

func (v *vaultImpl) Open(ctx context.Context, plainPassword string) (bool, error) {
	if v.encryptionKey != nil {
		return true, nil
	}

	row := v.db.QueryRowContext(ctx, `
	SELECT
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

	var keyHash string
	var saltBase64 string
	var params ArgonParameters

	err := row.Scan(&keyHash, &params.Aargon2Version, &params.Variant, &params.Memory,
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
		v.encryptionKey = memguard.NewEnclave(derivedKey)
	}

	return match, nil
}

func (v *vaultImpl) Seal() {
	v.encryptionKey = nil
	memguard.Purge()
}

func (v *vaultImpl) Close() {
	defer memguard.Purge()
}

func (v *vaultImpl) Encrypt(ctx context.Context, plaintext []byte) ([]byte, error) {
	return nil, nil
}

func (v *vaultImpl) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	return nil, nil
}
