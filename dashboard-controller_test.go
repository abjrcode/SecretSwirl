package main

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func initController(t *testing.T) *DashboardController {
	db, err := testhelpers.NewInMemoryMigratedDatabase("dashboard-controller-tests.db")

	require.NoError(t, err)
	controller := &DashboardController{
		db: db,
	}
	ctx := zerolog.Nop().WithContext(context.Background())
	controller.Init(ctx)

	return controller
}

func TestListFavoritesEmpty(t *testing.T) {
	controller := initController(t)

	favorites, err := controller.ListFavorites()
	require.NoError(t, err)

	require.Len(t, favorites, 0)
}
