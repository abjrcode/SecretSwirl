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
	defer v.Close()

	isSetup, err := v.IsSetup(context.Background())

	assert.NoError(t, err)
	assert.False(t, isSetup)
}

func TestIsSetupAfterConfigure(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestIsSetupAfterConfigure")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	defer v.Close()

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
	defer v.Close()

	_, err = v.Open(context.Background(), "password")

	require.ErrorIs(t, err, ErrVaultNotConfigured)
}

func TestConfigureVault(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestConfigureVault")

	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	defer v.Close()

	err = v.ConfigureKey(context.Background(), "password")

	assert.NoError(t, err)
}

func TestConfigureVaultTwice(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestConfigureVaultTwice")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	defer v.Close()

	err = v.ConfigureKey(context.Background(), "password")
	require.NoError(t, err)

	err = v.ConfigureKey(context.Background(), "password")
	require.ErrorIs(t, err, ErrVaultAlreadyConfigured)
}

func TestOpenVaultCorrectPassword(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestOpenVault")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	defer v.Close()

	err = v.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)

	success, err := v.Open(context.Background(), "123")
	require.NoError(t, err)

	assert.True(t, success)
}

func TestOpenVaultWithWrongPassword(t *testing.T) {
	db, err := testhelpers.NewInMemoryMigratedDatabase("TestOpenVaultWithWrongPassword")
	require.NoError(t, err)

	v := NewVault(db, utils.NewDatetime())
	defer v.Close()

	err = v.ConfigureKey(context.Background(), "123")
	require.NoError(t, err)
	v.Seal()

	success, err := v.Open(context.Background(), "456")
	require.NoError(t, err)

	assert.False(t, success)
}
