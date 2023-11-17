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
	"github.com/abjrcode/swervo/internal/config"
	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/logging"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/abjrcode/swervo/internal/utils"
	awsiamidc "github.com/abjrcode/swervo/providers/aws-iam-idc"

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

	errorHandler := logging.NewErrorHandler()

	logger.Info().Msgf("Swervo version: %s, commit SHA: %s", Version, CommitSha)
	logger.Info().Msgf("app data directory: [%s]", appDataDir)

	var sqlDb *sql.DB

	if !generateBindingsRun {
		dataStore := datastore.New(appDataDir, "swervo.db")
		migrationRunner, err := migrations.New(migrations.DefaultMigrationsFs, "scripts", dataStore, logger, errorHandler)

		errorHandler.CatchWithMsg(&logger, err, "could not read migrations from embedded filesystem")

		if err := migrationRunner.RunSafe(); err != nil {
			errorHandler.CatchWithMsg(&logger, err, "error when running migrations")
		}

		sqlDb, err = dataStore.Open()

		errorHandler.CatchWithMsg(&logger, err, "could not open database")

		defer sqlDb.Close()
	}

	appController := NewAppController()
	timeProvider := utils.NewDatetime()
	vault := vault.NewVault(sqlDb, timeProvider, &logger, errorHandler)
	defer vault.Close()

	authController := NewAuthController(vault)
	dashboardController := NewDashboardController(sqlDb)

	awsIdcController := awsiamidc.NewAwsIdentityCenterController(sqlDb, vault, awssso.NewAwsSsoOidcClient(), timeProvider)

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
			authController.Init(logger.WithContext(ctx), errorHandler)
			dashboardController.Init(logger.WithContext(ctx), errorHandler)
			awsIdcController.Init(logger.WithContext(ctx), errorHandler)
		},
		Bind: []interface{}{
			appController,
			authController,
			dashboardController,
			awsIdcController,
		},
	}); err != nil {
		errorHandler.Catch(&logger, errors.New("failed to launch Swervo"))
	}
}
