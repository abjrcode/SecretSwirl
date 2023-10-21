package migrations

import (
	"database/sql"
	"errors"
	"io/fs"
	"os"

	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
)

type MigrationRunner struct {
	migrationSource *source.Driver
	appStore        *datastore.AppStore
	logger          *zerolog.Logger
}

func New(migrationsFsys fs.FS, migrationsPath string, appStore *datastore.AppStore, logger zerolog.Logger) (*MigrationRunner, error) {
	migrationsSrc, err := iofs.New(migrationsFsys, migrationsPath)

	if err != nil {
		return nil, err
	}

	return &MigrationRunner{
		migrationSource: &migrationsSrc,
		appStore:        appStore,
		logger:          &logger,
	}, nil
}

type proxyLogger struct {
	logger *zerolog.Logger
}

func (l *proxyLogger) Printf(format string, v ...interface{}) {
	l.logger.Trace().Msgf(format, v...)
}

func (l *proxyLogger) Verbose() bool {
	return true
}

func (runner *MigrationRunner) shouldRunMigrations(db *sql.DB) (bool, uint, uint, error) {
	driver, driverErr := sqlite3.WithInstance(db, &sqlite3.Config{})

	if driverErr != nil {
		return false, 0, 0, driverErr
	}

	migrator, err := migrate.NewWithInstance("iofs", *runner.migrationSource, "sqlite3", driver)
	migrator.Log = &proxyLogger{
		logger: runner.logger,
	}

	if err != nil {
		return false, 0, 0, err
	}

	currentVersion, dirty, err := migrator.Version()

	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return false, 0, 0, err
	}

	if dirty {
		return false, 0, 0, nil
	}

	var nextUp uint = 0

	if currentVersion == 0 {
		nextUp, err = (*runner.migrationSource).First()
	} else {
		nextUp, err = (*runner.migrationSource).Next(currentVersion)
	}

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, 0, 0, nil
		}

		return false, 0, 0, err
	}

	return true, currentVersion, nextUp, nil
}

func (runner *MigrationRunner) up(db *sql.DB) error {
	driver, driverErr := sqlite3.WithInstance(db, &sqlite3.Config{})

	if driverErr != nil {
		return driverErr
	}

	if migrator, err := migrate.NewWithInstance("iofs", *runner.migrationSource, "sqlite3", driver); err != nil {
		return err
	} else {
		migrator.Log = &proxyLogger{
			logger: runner.logger,
		}

		if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
			return err
		}
	}

	return nil
}

// RunSafe runs migrations if necessary: new migrations are available and the database is not dirty.
// It returns the current version of the database or an error.
// If an error occurs, the database is restored to its previous state.
func (runner *MigrationRunner) RunSafe() error {
	db, err := runner.appStore.Open()

	if err != nil {
		runner.logger.Error().Err(err).Msgf("could not open database")
		return err
	}

	defer db.Close()

	shouldRunMigrations, currentVersion, nextUp, err := runner.shouldRunMigrations(db)

	if err != nil {
		runner.logger.Error().Err(err).Msgf("could not determine if migrations should be run")
		return err
	}

	runner.logger.Debug().Msgf("shouldRunMigrations: %t, currentVersion: %d, nextUp: %d", shouldRunMigrations, currentVersion, nextUp)

	if shouldRunMigrations {
		runner.logger.Info().Msg("taking a database backup")
		if err := runner.appStore.TakeBackup(); err != nil {
			runner.logger.Error().Err(err).Msgf("error taking database backup: %s", err)
			return nil
		} else {
			runner.logger.Info().Msgf("migrating database from @[%d] to @[%d]", currentVersion, nextUp)
			if upgradeError := runner.up(db); upgradeError != nil {
				runner.logger.Error().Err(upgradeError).Msgf("could not migrate database: %s", upgradeError)
				runner.logger.Info().Msg("restoring database backup")
				if restoreError := runner.appStore.RestoreBackup(); restoreError != nil {
					runner.logger.Error().Err(restoreError).Msgf("could not restore database backup: %s", restoreError)
					return errors.Join(upgradeError, restoreError)
				} else {
					return upgradeError
				}
			} else {
				runner.logger.Info().Msg("database migrated successfully")
				return nil
			}
		}
	}

	runner.logger.Info().Msg("no migrations to run")
	return nil
}
