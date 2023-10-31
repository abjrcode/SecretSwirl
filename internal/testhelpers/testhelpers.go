package testhelpers

import (
	"database/sql"
	"fmt"

	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/rs/zerolog"
)

func NewInMemoryMigratedDatabase(dbName string) (*sql.DB, error) {
	dbNameSuffix := utils.RandomString(4)
	store := datastore.NewInMemory(fmt.Sprintf("%s_%s.db", dbName, dbNameSuffix))

	db, err := store.Open()

	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	migrationRunner, err := migrations.New(migrations.DefaultMigrationsFs, "scripts", store, zerolog.Nop())

	if err != nil {
		return nil, err
	}

	err = migrationRunner.RunSafe()

	if err != nil {
		return nil, err
	}

	return db, nil
}
