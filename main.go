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
	"github.com/abjrcode/swervo/internal/config"
	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/faults"
	"github.com/abjrcode/swervo/internal/logging"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/abjrcode/swervo/internal/utils"
	awsiamidc "github.com/abjrcode/swervo/providers/aws_iam_idc"

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
	generateBindingsRun := config.IsWailsRunningAppToGenerateBindings(os.Args)

	pwd, err := os.Executable()

	if err != nil {
		log.Fatalf("failed to determine current working directory: [%s]", err)
	}

	appDataDir, appDataDirErr := config.GetAppDataDir(pwd, BuildType == "debug")

	if appDataDirErr != nil {
		log.Fatalf("failed to determine app data directory: [%s]", appDataDir)
	}

	var logFile = io.Discard

	if !generateBindingsRun {
		config.InitializeAppDataDir(appDataDir)

		file, logFileErr := logging.InitLogFile(appDataDir, "swervo_log.json")

		if logFileErr != nil {
			log.Fatalf("failed to initialize log file: [%s]", logFileErr)
		}

		logFile = file

		defer file.Close()
	}

	logger := logging.InitLogger(logFile, Version, CommitSha)

	errorHandler := faults.NewErrorHandler()

	logger.Info().Msgf("Swervo version: %s, commit SHA: %s", Version, CommitSha)
	logger.Info().Msgf("app data directory: [%s]", appDataDir)

	var sqlDb *sql.DB

	if !generateBindingsRun {
		dataStore := datastore.New(appDataDir, "swervo.db")
		migrationRunner, err := migrations.New(migrations.DefaultMigrationsFs, "scripts", dataStore, logger)

		errorHandler.CatchWithMsg(logger, err, "could not read migrations from embedded filesystem")

		if err := migrationRunner.RunSafe(); err != nil {
			errorHandler.CatchWithMsg(logger, err, "error when running migrations")
		}

		sqlDb, err = dataStore.Open()

		errorHandler.CatchWithMsg(logger, err, "could not open database")

		defer sqlDb.Close()
	}

	timeProvider := utils.NewClock()
	vault := vault.NewVault(sqlDb, timeProvider, logger)
	defer vault.Seal()

	authController := NewAuthController(vault, logger)

	favoritesRepo := favorites.NewFavorites(sqlDb, logger)
	dashboardController := NewDashboardController(favoritesRepo, logger)

	awsIdcController := awsiamidc.NewAwsIdentityCenterController(sqlDb, favoritesRepo, vault, awssso.NewAwsSsoOidcClient(), timeProvider, logger)

	appController := &AppController{
		authController:      authController,
		dashboardController: dashboardController,
		awsIamIdcController: awsIdcController,
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
		Logger: logging.NewWailsLoggerAdapter(&logger),
		OnStartup: func(ctx context.Context) {
			errorHandler.InitWailsContext(&ctx)

			appController.Init(logger.WithContext(ctx), errorHandler)
		},
		Bind: []interface{}{
			appController,
			authController,
			dashboardController,
			awsIdcController,
		},
	}); err != nil {
		errorHandler.Catch(logger, errors.New("failed to launch Swervo"))
	}
}
