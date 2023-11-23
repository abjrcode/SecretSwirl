package favorites

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestAddFavorite(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "favorites-repo-tests.db")
	require.NoError(t, err)

	logger := zerolog.Nop()
	repo := NewFavorites(db, &logger)

	favorite := &Favorite{
		ProviderCode: "aws-iam-idc",
		InstanceId:   "some-nice-id",
	}

	ctx := context.Background()
	err = repo.Add(ctx, favorite)
	require.NoError(t, err)

	favorites, err := repo.ListAll(ctx)
	require.NoError(t, err)

	require.Len(t, favorites, 1)
	require.Equal(t, favorite, favorites[0])
}

func TestRemoveFavorite(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "favorites-repo-tests.db")
	require.NoError(t, err)

	logger := zerolog.Nop()

	repo := NewFavorites(db, &logger)
	ctx := context.Background()

	favorite := &Favorite{
		ProviderCode: "aws-iam-idc",
		InstanceId:   "some-nice-id",
	}

	err = repo.Add(ctx, favorite)
	require.NoError(t, err)

	favorites, err := repo.ListAll(ctx)
	require.NoError(t, err)

	require.Len(t, favorites, 1)
	require.Equal(t, favorite, favorites[0])

	err = repo.Remove(ctx, favorite)
	require.NoError(t, err)

	favorites, err = repo.ListAll(ctx)
	require.NoError(t, err)

	require.Len(t, favorites, 0)
}