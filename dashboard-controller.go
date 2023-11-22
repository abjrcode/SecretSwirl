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

type FavoriteInstance struct {
	ProviderCode string `json:"providerCode"`
	InstanceId   string `json:"instanceId"`
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

func (c *DashboardController) ListFavorites() ([]FavoriteInstance, error) {
	rows, err := c.db.QueryContext(c.ctx, `SELECT * FROM favorite_instances`)

	if err != nil {
		if err == sql.ErrNoRows {
			return []FavoriteInstance{}, nil
		}

		c.errorHandler.Catch(c.logger, err)
	}

	favorites := make([]FavoriteInstance, 0, 10)

	for rows.Next() {
		var favorite FavoriteInstance
		err := rows.Scan(&favorite.ProviderCode, &favorite.InstanceId)
		if err != nil {
			c.errorHandler.Catch(c.logger, err)
		}
		favorites = append(favorites, favorite)
	}

	return favorites, nil
}

func (c *DashboardController) ListProviders() []Provider {
	return supportedProviders
}
