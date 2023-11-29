package migrations

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/rs/zerolog"
)

func NewInMemoryMigratedDatabase(t *testing.T, dbName string) (*sql.DB, error) {
	dbNameSuffix := utils.RandomString(4)
	store := datastore.NewInMemory(fmt.Sprintf("%s_%s.db", dbName, dbNameSuffix))

	migrationRunner, err := NewMigrationRunner(DefaultMigrationsFs, "scripts", store, zerolog.Nop())

	if err != nil {
		return nil, err
	}

	err = migrationRunner.RunSafe()

	if err != nil {
		return nil, err
	}

	db, err := store.Open()

	if err != nil {
		return nil, err
	}

	return db, nil
}
