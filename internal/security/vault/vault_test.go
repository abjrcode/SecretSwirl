package vault

import (
	"testing"

	"github.com/abjrcode/swervo/internal/eventing"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/testhelpers"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSetup(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestIsSetup")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)
	t.Cleanup(v.Seal)

	isSetup, err := v.IsConfigured(testhelpers.NewMockAppContext())
	require.NoError(t, err)

	assert.False(t, isSetup)
}

func TestIsSetup_AfterConfigure(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestIsSetupAfterConfigure")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)
	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)
	t.Cleanup(v.Seal)

	ctx := testhelpers.NewMockAppContext()

	err = v.Configure(ctx, "password")
	require.NoError(t, err)

	isSetup, err := v.IsConfigured(ctx)
	require.NoError(t, err)

	assert.True(t, isSetup)
}

func TestOpen_NotConfiguredVault(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestConfigureVault")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	t.Cleanup(v.Seal)

	_, err = v.Open(testhelpers.NewMockAppContext(), "password")

	require.ErrorIs(t, err, ErrVaultNotConfigured)
}

func TestConfigureVault(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestConfigureVault")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)
	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	t.Cleanup(v.Seal)

	err = v.Configure(testhelpers.NewMockAppContext(), "password")

	assert.NoError(t, err)
}

func TestConfigureVault_Twice(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestConfigureVaultTwice")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)
	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	t.Cleanup(v.Seal)

	ctx := testhelpers.NewMockAppContext()

	err = v.Configure(ctx, "password")
	require.NoError(t, err)

	err = v.Configure(ctx, "password")
	require.ErrorIs(t, err, ErrVaultAlreadyConfigured)
}

func TestOpenVault_CorrectPassword(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestOpenVault")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)

	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	t.Cleanup(v.Seal)

	ctx := testhelpers.NewMockAppContext()

	err = v.Configure(ctx, "123")
	require.NoError(t, err)

	success, err := v.Open(ctx, "123")
	require.NoError(t, err)

	assert.True(t, success)
}

func TestVault_OpenTwice(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestOpenVault")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)

	v := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	t.Cleanup(v.Seal)

	ctx := testhelpers.NewMockAppContext()

	err = v.Configure(ctx, "123")
	require.NoError(t, err)

	success, err := v.Open(ctx, "123")
	require.NoError(t, err)
	assert.True(t, success)

	success, err = v.Open(ctx, "123")
	require.NoError(t, err)
	assert.True(t, success)
}

func TestOpenVault_WithWrongPassword(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestOpenVaultWithWrongPassword")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)

	vault := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)
	t.Cleanup(vault.Seal)

	ctx := testhelpers.NewMockAppContext()

	err = vault.Configure(ctx, "123")
	require.NoError(t, err)
	vault.Seal()

	success, err := vault.Open(ctx, "456")
	require.NoError(t, err)

	assert.False(t, success)
}

func TestEncryptDecrypt(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecrypt")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)

	vault := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)
	t.Cleanup(vault.Seal)

	err = vault.Configure(testhelpers.NewMockAppContext(), "123")
	require.NoError(t, err)

	encrypted, keyId, err := vault.Encrypt("hello")
	require.NoError(t, err)

	decrypted, err := vault.Decrypt(encrypted, keyId)
	require.NoError(t, err)

	assert.Equal(t, "hello", decrypted)
}

func TestEncryptDecrypt_AfterVaultSealAndOpen(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecrypt")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)

	vault := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	t.Cleanup(vault.Seal)

	ctx := testhelpers.NewMockAppContext()

	err = vault.Configure(ctx, "123")
	require.NoError(t, err)

	encrypted, keyId, err := vault.Encrypt("hello")
	require.NoError(t, err)

	vault.Seal()

	isOpenend, err := vault.Open(ctx, "123")
	require.True(t, isOpenend)
	require.NoError(t, err)

	decrypted, err := vault.Decrypt(encrypted, keyId)
	require.NoError(t, err)

	assert.Equal(t, "hello", decrypted)
}

func TestEncryptDecryptWithWrongPassword(t *testing.T) {
	ctx := testhelpers.NewMockAppContext()

	originalDb, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecryptWithWrongPasswordOriginal")
	require.NoError(t, err)
	originalClock := testhelpers.NewMockClock()
	originalClock.On("NowUnix").Return(1)
	originalVault := NewVault(originalDb, eventing.NewEventbus(originalDb, originalClock), originalClock)
	err = originalVault.Configure(ctx, "123")
	require.NoError(t, err)
	encrypted, keyId, err := originalVault.Encrypt("hello")
	require.NoError(t, err)
	originalVault.Seal()

	newDb, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecryptWithWrongPasswordNew")
	require.NoError(t, err)
	newClock := testhelpers.NewMockClock()
	newClock.On("NowUnix").Return(2)
	newVault := NewVault(newDb, eventing.NewEventbus(newDb, newClock), newClock)
	err = newVault.Configure(ctx, "321")
	require.NoError(t, err)
	_, err = newVault.Decrypt(encrypted, keyId)
	require.Error(t, err)
	newVault.Seal()
}

func TestEncryptDecryptErrorOnSealedVault(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "TestEncryptDecryptAfterVaultSeal")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()

	mockClock.On("NowUnix").Return(1)

	vault := NewVault(db, eventing.NewEventbus(db, mockClock), mockClock)

	err = vault.Configure(testhelpers.NewMockAppContext(), "123")
	require.NoError(t, err)
	vault.Seal()

	_, _, err = vault.Encrypt("hello")
	require.Error(t, err)
}
