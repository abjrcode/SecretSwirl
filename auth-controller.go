package main

import (
	"context"

	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/rs/zerolog"
)

type AuthController struct {
	ctx    context.Context
	logger *zerolog.Logger
	vault  vault.Vault
}

func (c *AuthController) Init(ctx context.Context) {
	c.ctx = ctx
	c.logger = zerolog.Ctx(ctx)
}

func (c *AuthController) IsVaultConfigured() (bool, error) {
	c.logger.Info().Msg("Checking if Vault is already configured")
	return c.vault.IsSetup(c.ctx)
}

func (c *AuthController) ConfigureVault(password string) error {
	c.logger.Info().Msg("Setting up Vault with a master password")
	return c.vault.ConfigureKey(c.ctx, password)
}

func (c *AuthController) UnlockVault(password string) (bool, error) {
	c.logger.Info().Msg("Attempting to unlock Vault with a master password")
	return c.vault.Open(c.ctx, password)
}

func (c *AuthController) LockVault() {
	c.logger.Info().Msg("Locking Vault")
	c.vault.Seal()
}
