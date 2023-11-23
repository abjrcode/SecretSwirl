package main

import (
	"context"

	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/logging"
	"github.com/rs/zerolog"
)

type DashboardController struct {
	ctx           context.Context
	logger        *zerolog.Logger
	errorHandler  logging.ErrorHandler
	favoritesRepo favorites.FavoritesRepo
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

func NewDashboardController(favoritesRepo favorites.FavoritesRepo) *DashboardController {

	return &DashboardController{
		favoritesRepo: favoritesRepo,
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
	favorites, err := c.favoritesRepo.ListAll(c.ctx)

	if err != nil {
		c.errorHandler.Catch(c.logger, err)
	}

	favoriteInstances := make([]FavoriteInstance, 0, len(favorites))

	for _, favorite := range favorites {
		favoriteInstances = append(favoriteInstances, FavoriteInstance{
			ProviderCode: favorite.ProviderCode,
			InstanceId:   favorite.InstanceId,
		})
	}

	return favoriteInstances, nil
}

func (c *DashboardController) ListProviders() []Provider {
	return supportedProviders
}
