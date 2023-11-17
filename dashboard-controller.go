package main

import (
	"context"
	"database/sql"

	"github.com/abjrcode/swervo/internal/logging"
	"github.com/rs/zerolog"
)

type DashboardController struct {
	ctx          context.Context
	logger       *zerolog.Logger
	errorHandler logging.ErrorHandler
	db           *sql.DB
}

type Provider struct {
	Code          string `json:"code"`
	Name          string `json:"name"`
	IconSvgBase64 string `json:"iconSvgBase64"`
}

type ConfiguredProvider struct {
	Code          string `json:"code"`
	DisplayName   string `json:"displayName"`
	InstanceId    string `json:"instanceId"`
	IsFavorite    bool   `json:"isFavorite"`
	IconSvgBase64 string `json:"iconSvgBase64"`
}

var (
	SupportedProviders = map[string]Provider{
		"aws-iam-idc": {
			Code: "aws-iam-idc",
			Name: "AWS IAM IDC",
		},
	}
)

var supportedProviders []Provider

func NewDashboardController(db *sql.DB) *DashboardController {

	return &DashboardController{
		db: db,
	}
}

func (c *DashboardController) Init(ctx context.Context, errorHandler logging.ErrorHandler) {
	c.ctx = ctx
	enrichedLogger := zerolog.Ctx(ctx).With().Str("component", "dashboard_controller").Logger()
	c.logger = &enrichedLogger
	c.errorHandler = errorHandler

	supportedProviders = make([]Provider, 0, len(SupportedProviders))
	for _, provider := range SupportedProviders {
		supportedProviders = append(supportedProviders, provider)
	}
}

func (c *DashboardController) ListFavorites() ([]ConfiguredProvider, error) {
	rows, err := c.db.QueryContext(c.ctx, `SELECT * FROM providers WHERE is_favorite = ?;`, true)

	if err != nil {
		return []ConfiguredProvider{}, err
	}

	providers := make([]ConfiguredProvider, 0, 10)

	for rows.Next() {
		var provider ConfiguredProvider
		err := rows.Scan(&provider.Code, &provider.InstanceId, &provider.DisplayName, &provider.IsFavorite)
		if err != nil {
			c.errorHandler.Catch(c.logger, err)
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

func (c *DashboardController) ListProviders() []Provider {
	return supportedProviders
}
