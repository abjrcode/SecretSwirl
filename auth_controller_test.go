package main

import (
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

	vault := vault.NewVault(db, mockClock, logger)

	controller := NewAuthController(vault, logger)

	return controller, mockClock
}

func TestAuthController_IsVaultConfigured(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	isVaultConfigured, err := controller.IsVaultConfigured(ctx)
	require.NoError(t, err)

	require.False(t, isVaultConfigured)

	mockTimeProvider.On("NowUnix").Return(1)
	err = controller.ConfigureVault(ctx, Auth_ConfigureVaultCommandInput{Password: "password"})
	require.NoError(t, err)

	isVaultConfigured, err = controller.IsVaultConfigured(ctx)
	require.NoError(t, err)
	require.True(t, isVaultConfigured)
}

func TestAuthController_ConfigureVault(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault(ctx, Auth_ConfigureVaultCommandInput{Password: "password"})
	require.NoError(t, err)

	isVaultConfigured, err := controller.IsVaultConfigured(ctx)
	require.NoError(t, err)
	require.True(t, isVaultConfigured)
}

func TestAuthController_UnlockVault(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault(ctx, Auth_ConfigureVaultCommandInput{Password: "password"})
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Return(2)
	unlocked, err := controller.UnlockVault(ctx, Auth_UnlockCommandInput{Password: "password"})
	require.NoError(t, err)
	require.True(t, unlocked)
}

func TestAuthController_UnlockVault_WithWrongPassword(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault(ctx, Auth_ConfigureVaultCommandInput{Password: "password"})
	require.NoError(t, err)

	controller.LockVault()

	mockTimeProvider.On("NowUnix").Return(2)
	unlocked, err := controller.UnlockVault(ctx, Auth_UnlockCommandInput{Password: "wrong-password"})
	require.NoError(t, err)
	require.False(t, unlocked)
}

func TestAuthController_UnlockVault_WithoutConfiguringFirst(t *testing.T) {
	controller, _ := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	_, err := controller.UnlockVault(ctx, Auth_UnlockCommandInput{
		Password: "password",
	})

	require.Error(t, err, vault.ErrVaultNotConfigured)
}

func TestAuthController_LockVault_WithoutConfiguringFirst(t *testing.T) {
	controller, _ := initAuthController(t)

	controller.LockVault()
}

func TestAuthController_LockVault_WithoutUnlockingFirst(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault(ctx, Auth_ConfigureVaultCommandInput{Password: "password"})
	require.NoError(t, err)

	controller.LockVault()
}

func TestAuthController_UnlockVault_WithoutLockingFirst(t *testing.T) {
	controller, mockTimeProvider := initAuthController(t)
	ctx := testhelpers.NewMockAppContext()

	mockTimeProvider.On("NowUnix").Return(1)
	err := controller.ConfigureVault(ctx, Auth_ConfigureVaultCommandInput{Password: "password"})
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Return(2)
	unlocked, err := controller.UnlockVault(ctx, Auth_UnlockCommandInput{Password: "password"})
	require.NoError(t, err)
	require.True(t, unlocked)
}
