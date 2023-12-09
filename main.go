package main

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io"
	"log"
	"os"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/eventing"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/abjrcode/swervo/internal/utils"
	awsidc "github.com/abjrcode/swervo/providers/aws_idc"
	awscredentialsfile "github.com/abjrcode/swervo/sinks/aws_credentials_file"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

var BuildType string = "debug"
var Version string = "v0.0.0"
var BuildTimestamp string = "NOW"
var CommitSha string = "HEAD"
var BuildLink string = "http://localhost"

func main() {
	generateBindingsRun := app.IsWailsRunningAppToGenerateBindings(os.Args)

	pwd, err := os.Executable()

	if err != nil {
		log.Fatalf("failed to determine current working directory: [%s]", err)
	}

	appDataDir, appDataDirErr := app.GetAppDataDir(pwd, BuildType == "debug")

	if appDataDirErr != nil {
		log.Fatalf("failed to determine app data directory: [%s]", appDataDir)
	}

	var logFile = io.Discard

	if !generateBindingsRun {
		app.InitializeAppDataDir(appDataDir)

		file, logFileErr := app.InitLogFile(appDataDir, "swervo_log.json")

		if logFileErr != nil {
			log.Fatalf("failed to initialize log file: [%s]", logFileErr)
		}

		logFile = file

		defer file.Close()
	}

	logger := app.InitLogger(logFile, Version, CommitSha)

	errorHandler := app.NewErrorHandler()

	logger.Info().Msgf("Swervo version: %s, commit SHA: %s", Version, CommitSha)
	logger.Info().Msgf("app data directory: [%s]", appDataDir)

	dataStore := datastore.New(appDataDir, "swervo.db")
	var db *sql.DB

	if !generateBindingsRun {
		db, err = dataStore.Open()
		if err != nil {
			errorHandler.CatchWithMsg(nil, logger, err, "failed to open database")
		}
		defer dataStore.Close(db)
		migrationRunner, err := migrations.NewMigrationRunner(migrations.DefaultMigrationsFs, "scripts", dataStore, logger)

		errorHandler.CatchWithMsg(nil, logger, err, "could not read migrations from embedded filesystem")

		if err := migrationRunner.RunSafe(); err != nil {
			errorHandler.CatchWithMsg(nil, logger, err, "error when running migrations")
		}
	}

	clock := utils.NewClock()

	eventBus := eventing.NewEventbus(db, clock)

	vault := vault.NewVault(db, eventBus, clock)
	defer vault.Seal()

	authController := NewAuthController(vault)

	favoritesRepo := favorites.NewFavorites(db)
	dashboardController := NewDashboardController(favoritesRepo)

	awsIdcController := awsidc.NewAwsIdentityCenterController(db, eventBus, favoritesRepo, vault, awssso.NewAwsSsoOidcClient(), clock)

	awsCredentialsFileController := awscredentialsfile.NewAwsCredentialsFileController(db, eventBus, vault, clock)

	appController := &AppController{
		authController:               authController,
		dashboardController:          dashboardController,
		awsIdcController:             awsIdcController,
		awsCredentialsFileController: awsCredentialsFileController,
	}

	logger.Info().Msgf("PID [%d] - launching Swervo", os.Getpid())
	if err := wails.Run(&options.App{
		Title:  "Swervo",
		Width:  1024,
		Height: 768,
		Menu:   appController.mainMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger: app.NewWailsLoggerAdapter(&logger),
		OnStartup: func(ctx context.Context) {
			appController.Init(logger.WithContext(ctx), errorHandler)
		},
		Bind: []interface{}{
			appController,
			authController,
			dashboardController,
			awsIdcController,
			awsCredentialsFileController,
		},
	}); err != nil {
		errorHandler.Catch(nil, logger, errors.New("failed to launch Swervo"))
	}
}
