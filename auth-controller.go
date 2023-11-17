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

func (c *AuthController) Init(ctx context.Context, errorHandler logging.ErrorHandler) {
	c.ctx = ctx
	c.logger = zerolog.Ctx(ctx)
	c.errorHandler = errorHandler
}

func (c *AuthController) IsVaultConfigured() bool {
	c.logger.Info().Msg("checking if vault is already configured")
	return c.vault.IsSetup(c.ctx)
}

func (c *AuthController) ConfigureVault(password string) error {
	c.logger.Info().Msg("setting up sault with a master password")
	return c.vault.ConfigureKey(c.ctx, password)
}

func (c *AuthController) UnlockVault(password string) (bool, error) {
	c.logger.Info().Msg("attempting to unlock vault with a master password")
	return c.vault.Open(c.ctx, password)
}

func (c *AuthController) LockVault() {
	c.logger.Info().Msg("locking Vault")
	c.vault.Seal()
}
