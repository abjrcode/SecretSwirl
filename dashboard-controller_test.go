package main

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func initDashboardController(t *testing.T) *DashboardController {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "dashboard-controller-tests.db")

	require.NoError(t, err)
	controller := &DashboardController{
		db: db,
	}
	ctx := zerolog.Nop().WithContext(context.Background())
	errHandler := testhelpers.NewMockErrorHandler(t)

	controller.Init(ctx, errHandler)

	return controller
}

func TestListFavoritesEmpty(t *testing.T) {
	controller := initDashboardController(t)

	favorites, err := controller.ListFavorites()
	require.NoError(t, err)

	require.Len(t, favorites, 0)
}
