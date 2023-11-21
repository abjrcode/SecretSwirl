package vault

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/abjrcode/swervo/internal/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSetup(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestIsSetup")
	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	isSetup := v.IsConfigured(context.Background())

	assert.False(t, isSetup)
}

func TestIsSetupAfterConfigure(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestIsSetupAfterConfigure")
	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	err = v.Configure(context.Background(), "password")
	require.NoError(t, err)

	isSetup := v.IsConfigured(context.Background())

	assert.True(t, isSetup)
}

func TestOpenNotConfiguredVault(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestConfigureVault")

	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	_, err = v.Open(context.Background(), "password")

	require.ErrorIs(t, err, ErrVaultNotConfigured)
}

func TestConfigureVault(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestConfigureVault")

	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	err = v.Configure(context.Background(), "password")

	assert.NoError(t, err)
}

func TestConfigureVaultTwice(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestConfigureVaultTwice")
	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	err = v.Configure(context.Background(), "password")
	require.NoError(t, err)

	err = v.Configure(context.Background(), "password")
	require.ErrorIs(t, err, ErrVaultAlreadyConfigured)
}

func TestOpenVaultCorrectPassword(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestOpenVault")
	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	err = v.Configure(context.Background(), "123")
	require.NoError(t, err)

	success, err := v.Open(context.Background(), "123")
	require.NoError(t, err)

	assert.True(t, success)
}

func TestVault_OpenTwice(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestOpenVault")
	require.NoError(t, err)

	logger := zerolog.Nop()
	v := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(v.Seal)

	err = v.Configure(context.Background(), "123")
	require.NoError(t, err)

	success, err := v.Open(context.Background(), "123")
	require.NoError(t, err)
	assert.True(t, success)

	success, err = v.Open(context.Background(), "123")
	require.NoError(t, err)
	assert.True(t, success)
}

func TestOpenVaultWithWrongPassword(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestOpenVaultWithWrongPassword")
	require.NoError(t, err)

	logger := zerolog.Nop()
	vault := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(vault.Seal)

	err = vault.Configure(context.Background(), "123")
	require.NoError(t, err)
	vault.Seal()

	success, err := vault.Open(context.Background(), "456")
	require.NoError(t, err)

	assert.False(t, success)
}

func TestEncryptDecrypt(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecrypt")
	require.NoError(t, err)

	logger := zerolog.Nop()
	vault := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(vault.Seal)

	err = vault.Configure(context.Background(), "123")
	require.NoError(t, err)

	encrypted, keyId, err := vault.Encrypt("hello")
	require.NoError(t, err)

	decrypted, err := vault.Decrypt(encrypted, keyId)
	require.NoError(t, err)

	assert.Equal(t, "hello", decrypted)
}

func TestEncryptDecryptAfterVaultSealAndOpen(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecrypt")
	require.NoError(t, err)

	logger := zerolog.Nop()
	vault := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	t.Cleanup(vault.Seal)

	err = vault.Configure(context.Background(), "123")
	require.NoError(t, err)

	encrypted, keyId, err := vault.Encrypt("hello")
	require.NoError(t, err)

	vault.Seal()

	isOpenend, err := vault.Open(context.Background(), "123")
	require.True(t, isOpenend)
	require.NoError(t, err)

	decrypted, err := vault.Decrypt(encrypted, keyId)
	require.NoError(t, err)

	assert.Equal(t, "hello", decrypted)
}

func TestEncryptDecryptWithWrongPassword(t *testing.T) {
	originalDb, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecryptWithWrongPasswordOriginal")
	require.NoError(t, err)
	logger := zerolog.Nop()
	originalVault := NewVault(originalDb, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	err = originalVault.Configure(context.Background(), "123")
	require.NoError(t, err)
	encrypted, keyId, err := originalVault.Encrypt("hello")
	require.NoError(t, err)
	originalVault.Seal()

	newDb, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecryptWithWrongPasswordNew")
	require.NoError(t, err)
	newVault := NewVault(newDb, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))
	err = newVault.Configure(context.Background(), "321")
	require.NoError(t, err)
	_, err = newVault.Decrypt(encrypted, keyId)
	require.Error(t, err)
	newVault.Seal()
}

func TestEncryptDecryptErrorOnSealedVault(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecryptAfterVaultSeal")
	require.NoError(t, err)

	logger := zerolog.Nop()
	vault := NewVault(db, utils.NewClock(), &logger, testhelpers.NewMockErrorHandler(t))

	err = vault.Configure(context.Background(), "123")
	require.NoError(t, err)
	vault.Seal()

	_, _, err = vault.Encrypt("hello")
	require.Error(t, err)
}
