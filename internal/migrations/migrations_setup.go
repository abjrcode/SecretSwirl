package migrations

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/rs/zerolog"
)

func NewInMemoryMigratedDatabase(t *testing.T, dbName string) (*sql.DB, error) {
	dbNameSuffix := utils.RandomString(4)
	store := datastore.NewInMemory(fmt.Sprintf("%s_%s.db", dbName, dbNameSuffix))

	db, err := store.Open()

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	migrationRunner, err := New(DefaultMigrationsFs, "scripts", store, zerolog.Nop(), testhelpers.NewMockErrorHandler(t))

	if err != nil {
		return nil, err
	}

	err = migrationRunner.RunSafe()

	if err != nil {
		return nil, err
	}

	return db, nil
}
