package main

import (
	"errors"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/security/vault"
)

type AuthController struct {
	vault vault.Vault
}

func NewAuthController(vault vault.Vault) *AuthController {
	return &AuthController{
		vault: vault,
	}
}

func check(err error) error {
	if !errors.Is(err, vault.ErrVaultAlreadyConfigured) {
		return err
	}

	if !errors.Is(err, vault.ErrVaultNotConfigured) {
		return err
	}

	if !errors.Is(err, vault.ErrVaultNotConfiguredOrSealed) {
		return err
	}

	return nil
}

func (c *AuthController) IsVaultConfigured(ctx app.Context) (bool, error) {
	ctx.Logger().Info().Msg("checking if vault is already configured")
	isConfigured, err := c.vault.IsConfigured(ctx)

	if check(err) != nil {
		return false, errors.Join(errors.New("failed to check if vault is configured"), err, app.ErrFatal)
	}

	return isConfigured, err
}

type Auth_ConfigureVaultCommandInput struct {
	Password string `json:"password"`
}

// ConfigureVault sets up the vault with a master password. It is called when the user sets up the app for the first time.
// After configuration, the vault is unsealed and ready to be used.
func (c *AuthController) ConfigureVault(ctx app.Context, input Auth_ConfigureVaultCommandInput) error {
	ctx.Logger().Info().Msg("setting up vault with a master password")
	err := c.vault.Configure(ctx, input.Password)

	if check(err) != nil {
		return errors.Join(errors.New("failed to check if vault is configured"), err, app.ErrFatal)
	}

	return err
}

type Auth_UnlockCommandInput struct {
	Password string `json:"password"`
}

// UnlockVault opens the vault with the given password. It is called when the user logs in.
func (c *AuthController) UnlockVault(ctx app.Context, input Auth_UnlockCommandInput) (bool, error) {
	ctx.Logger().Info().Msg("attempting to unlock vault with a master password")
	success, err := c.vault.Open(ctx, input.Password)

	if check(err) != nil {
		return false, errors.Join(errors.New("failed to unlock vault"), err, app.ErrFatal)
	}

	return success, err
}

// LockVault closes the vault and purges the key from memory. It is called when the user logs out.
func (c *AuthController) LockVault(ctx app.Context) {
	ctx.Logger().Info().Msg("locking Vault")
	c.vault.Seal()
}
