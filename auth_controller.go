package main

import (
	"context"

	"github.com/abjrcode/swervo/internal/logging"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/rs/zerolog"
)

type AuthController struct {
	ctx          context.Context
	logger       *zerolog.Logger
	errorHandler logging.ErrorHandler
	vault        vault.Vault
}

func NewAuthController(vault vault.Vault) *AuthController {
	return &AuthController{
		vault: vault,
	}
}

func (c *AuthController) Init(ctx context.Context, errorHandler logging.ErrorHandler) {
	c.ctx = ctx
	enrichedLogger := zerolog.Ctx(ctx).With().Str("component", "auth_controller").Logger()
	c.logger = &enrichedLogger
	c.errorHandler = errorHandler
}

func (c *AuthController) IsVaultConfigured() bool {
	c.logger.Info().Msg("checking if vault is already configured")
	return c.vault.IsConfigured(c.ctx)
}

// ConfigureVault sets up the vault with a master password. It is called when the user sets up the app for the first time.
// After configuration, the vault is unsealed and ready to be used.
func (c *AuthController) ConfigureVault(password string) error {
	c.logger.Info().Msg("setting up vault with a master password")
	return c.vault.Configure(c.ctx, password)
}

// UnlockVault opens the vault with the given password. It is called when the user logs in.
func (c *AuthController) UnlockVault(password string) (bool, error) {
	c.logger.Info().Msg("attempting to unlock vault with a master password")
	return c.vault.Open(c.ctx, password)
}

// LockVault closes the vault and purges the key from memory. It is called when the user logs out.
func (c *AuthController) LockVault() {
	c.logger.Info().Msg("locking Vault")
	c.vault.Seal()
}
