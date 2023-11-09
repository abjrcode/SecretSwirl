package vault

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/abjrcode/swervo/internal/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSetup(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestIsSetup")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	t.Cleanup(v.Close)

	isSetup, err := v.IsSetup(context.Background())

	assert.NoError(t, err)
	assert.False(t, isSetup)
}

func TestIsSetupAfterConfigure(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestIsSetupAfterConfigure")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	t.Cleanup(v.Close)

	err = v.ConfigureKey(context.Background(), "password")
	require.NoError(t, err)

	isSetup, err := v.IsSetup(context.Background())

	assert.NoError(t, err)
	assert.True(t, isSetup)
}

func TestOpenNotConfiguredVault(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestConfigureVault")

	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	t.Cleanup(v.Close)

	_, err = v.Open(context.Background(), "password")

	require.ErrorIs(t, err, ErrVaultNotConfigured)
}

func TestConfigureVault(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestConfigureVault")

	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	t.Cleanup(v.Close)

	err = v.ConfigureKey(context.Background(), "password")

	assert.NoError(t, err)
}

func TestConfigureVaultTwice(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestConfigureVaultTwice")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	t.Cleanup(v.Close)

	err = v.ConfigureKey(context.Background(), "password")
	require.NoError(t, err)

	err = v.ConfigureKey(context.Background(), "password")
	require.ErrorIs(t, err, ErrVaultAlreadyConfigured)
}

func TestOpenVaultCorrectPassword(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestOpenVault")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	t.Cleanup(v.Close)

	err = v.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)

	success, err := v.Open(context.Background(), "123")
	require.NoError(t, err)

	assert.True(t, success)
}

func TestOpenVaultWithWrongPassword(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestOpenVaultWithWrongPassword")
	require.NoError(t, err)

	vault := NewVault(db, utils.NewDatetime())
	t.Cleanup(vault.Close)

	err = vault.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)
	vault.Seal()

	success, err := vault.Open(context.Background(), "456")
	require.NoError(t, err)

	assert.False(t, success)
}

func TestEncryptDecrypt(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestEncryptDecrypt")
	require.NoError(t, err)

	vault := NewVault(db, utils.NewDatetime())
	t.Cleanup(vault.Close)

	err = vault.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)

	encrypted, keyId, err := vault.Encrypt("hello")
	require.NoError(t, err)

	decrypted, err := vault.Decrypt(encrypted, keyId)
	require.NoError(t, err)

	assert.Equal(t, "hello", decrypted)
}

func TestEncryptDecryptAfterVaultSealAndOpen(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestEncryptDecrypt")
	require.NoError(t, err)

	vault := NewVault(db, utils.NewDatetime())
	t.Cleanup(vault.Close)

	err = vault.ConfigureKey(context.Background(), "123")
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
	originalDb, err := testhelpers.NewInMemoryMigratedDatabase("TestEncryptDecryptWithWrongPasswordOriginal")
	require.NoError(t, err)
	originalVault := NewVault(originalDb, utils.NewDatetime())
	err = originalVault.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)
	encrypted, keyId, err := originalVault.Encrypt("hello")
	require.NoError(t, err)
	originalVault.Close()

	newDb, err := testhelpers.NewInMemoryMigratedDatabase("TestEncryptDecryptWithWrongPasswordNew")
	require.NoError(t, err)
	newVault := NewVault(newDb, utils.NewDatetime())
	err = newVault.ConfigureKey(context.Background(), "321")
	require.NoError(t, err)
	_, err = newVault.Decrypt(encrypted, keyId)
	require.Error(t, err)
	newVault.Close()
}

func TestEncryptDecryptErrorOnSealedVault(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestEncryptDecryptAfterVaultSeal")
	require.NoError(t, err)

	vault := NewVault(db, utils.NewDatetime())

	err = vault.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)
	vault.Seal()

	_, _, err = vault.Encrypt("hello")
	require.Error(t, err)
}
