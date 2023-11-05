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

func (c *AuthController) IsPasswordSetup() (bool, error) {
	return c.vault.IsSetup(c.ctx)
}

func (c *AuthController) SetupMasterPassword(password string) error {
	return c.vault.ConfigureKey(c.ctx, password)
}

func (c *AuthController) Login(password string) (bool, error) {
	return c.vault.Open(c.ctx, password)
}

func (c *AuthController) Logout() {
	c.vault.Seal()
}
