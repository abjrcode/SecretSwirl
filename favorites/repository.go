package favorites

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rs/zerolog"
)

type Favorite struct {
	ProviderCode string
	InstanceId   string
}

type FavoritesRepo interface {
	ListAll(ctx context.Context) ([]*Favorite, error)
	IsFavorite(ctx context.Context, favorite *Favorite) (bool, error)
	Add(ctx context.Context, favorite *Favorite) error
	Remove(ctx context.Context, favorite *Favorite) error
}

type favoritesImpl struct {
	logger *zerolog.Logger
	db     *sql.DB
}

func NewFavorites(db *sql.DB, logger *zerolog.Logger) FavoritesRepo {
	enrichedLogger := logger.With().Str("component", "favorites_repo").Logger()

	return &favoritesImpl{
		db:     db,
		logger: &enrichedLogger,
	}
}

func (f *favoritesImpl) ListAll(ctx context.Context) ([]*Favorite, error) {
	rows, err := f.db.QueryContext(ctx, `SELECT * FROM favorite_instances`)

	if err != nil {
		if err == sql.ErrNoRows {
			return []*Favorite{}, nil
		}

		return nil, err
	}

	favorites := make([]*Favorite, 0, 10)

	for rows.Next() {
		var favorite Favorite
		err := rows.Scan(&favorite.ProviderCode, &favorite.InstanceId)

		if err != nil {
			return nil, err
		}

		favorites = append(favorites, &favorite)
	}

	return favorites, nil
}

func (f *favoritesImpl) IsFavorite(ctx context.Context, favorite *Favorite) (bool, error) {
	row := f.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM favorite_instances WHERE provider_code = ? AND instance_id = ? `, favorite.ProviderCode, favorite.InstanceId)

	var count int
	err := row.Scan(&count)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return count > 0, nil
}

func (f *favoritesImpl) Add(ctx context.Context, favorite *Favorite) error {
	_, err := f.db.ExecContext(ctx, `INSERT INTO favorite_instances (provider_code, instance_id) VALUES (?, ?) `, favorite.ProviderCode, favorite.InstanceId)

	if err != nil {
		return err
	}

	return nil
}

func (f *favoritesImpl) Remove(ctx context.Context, favorite *Favorite) error {
	res, err := f.db.ExecContext(ctx, `DELETE FROM favorite_instances WHERE provider_code = ? AND instance_id = ? `, favorite.ProviderCode, favorite.InstanceId)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
