package main

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func initAuthController(t *testing.T) (*AuthController, *testhelpers.MockClock) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "auth-controller-tests.db")
	require.NoError(t, err)

	logger := zerolog.Nop()
	mockClock := testhelpers.NewMockClock()
	mockErrHandler := testhelpers.NewMockErrorHandler(t)

	vault := vault.NewVault(db, mockClock, &logger, mockErrHandler)

	controller := &AuthController{
		vault: vault,
	}

	ctx := logger.WithContext(context.Background())

	controller.Init(ctx, mockErrHandler)

	return controller, mockClock
}

func TestAuthController_IsVaultConfigured(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)

	require.False(t, controller.IsVaultConfigured())

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault("password")
	require.NoError(t, err)

	require.True(t, controller.IsVaultConfigured())
}

func TestAuthController_ConfigureVault(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault("password")
	require.NoError(t, err)

	require.True(t, controller.IsVaultConfigured())
}

func TestAuthController_UnlockVault(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault("password")
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Return(2)
	unlocked, err := controller.UnlockVault("password")
	require.NoError(t, err)
	require.True(t, unlocked)
}

func TestAuthController_UnlockVault_WithWrongPassword(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault("password")
	require.NoError(t, err)

	controller.LockVault()

	mockTimeProvider.On("NowUnix").Return(2)
	unlocked, err := controller.UnlockVault("wrong-password")
	require.NoError(t, err)
	require.False(t, unlocked)
}

func TestAuthController_UnlockVault_WithoutConfiguringFirst(t *testing.T) {
	controller, _ := initAuthController(t)

	_, err := controller.UnlockVault("password")
	require.Error(t, err, vault.ErrVaultNotConfigured)
}

func TestAuthController_LockVault_WithoutConfiguringFirst(t *testing.T) {
	controller, _ := initAuthController(t)

	controller.LockVault()
}

func TestAuthController_LockVault_WithoutUnlockingFirst(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault("password")
	require.NoError(t, err)

	controller.LockVault()
}

func TestAuthController_UnlockVault_WithoutLockingFirst(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault("password")
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Return(2)
	unlocked, err := controller.UnlockVault("password")
	require.NoError(t, err)
	require.True(t, unlocked)
}
