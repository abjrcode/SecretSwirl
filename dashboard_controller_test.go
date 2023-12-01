package main

import (
	"testing"

	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

func initDashboardController(t *testing.T) *DashboardController {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "dashboard-controller-tests.db")
	require.NoError(t, err)

	favoritesRepo := favorites.NewFavorites(db)

	controller := &DashboardController{
		favoritesRepo: favoritesRepo,
	}

	return controller
}

func TestListFavoritesEmpty(t *testing.T) {
	controller := initDashboardController(t)
	ctx := testhelpers.NewMockAppContext()

	favorites, err := controller.ListFavorites(ctx)
	require.NoError(t, err)

	require.Len(t, favorites, 0)
}
