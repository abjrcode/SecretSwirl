package main

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func initDashboardController(t *testing.T) *DashboardController {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "dashboard-controller-tests.db")
	require.NoError(t, err)

	logger := zerolog.Nop()
	favoritesRepo := favorites.NewFavorites(db, logger)

	controller := &DashboardController{
		favoritesRepo: favoritesRepo,
		logger:        logger,
	}

	return controller
}

func TestListFavoritesEmpty(t *testing.T) {
	controller := initDashboardController(t)

	favorites, err := controller.ListFavorites(context.TODO())
	require.NoError(t, err)

	require.Len(t, favorites, 0)
}
