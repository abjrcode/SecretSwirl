package main

import (
	"errors"

	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/providers"
)

type DashboardController struct {
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

var supportedProviders []Provider

func NewDashboardController(favoritesRepo favorites.FavoritesRepo) *DashboardController {
	supportedProviders = make([]Provider, 0, len(providers.SupportedProviders))
	for _, provider := range providers.SupportedProviders {
		supportedProviders = append(supportedProviders, Provider{
			Code: provider.Code,
			Name: provider.Name,
		})
	}

	return &DashboardController{
		favoritesRepo: favoritesRepo,
	}
}

func (c *DashboardController) ListFavorites(ctx app.Context) ([]FavoriteInstance, error) {
	favorites, err := c.favoritesRepo.ListAll(ctx)

	if err != nil {
		return nil, errors.Join(err, app.ErrFatal)
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
