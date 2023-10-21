package main

import (
	"context"
	"embed"
	"io"
	"log"
	"os"

	"github.com/abjrcode/swervo/internal/config"
	"github.com/abjrcode/swervo/internal/datastore"
	"github.com/abjrcode/swervo/internal/logging"
	"github.com/abjrcode/swervo/internal/migrations"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed db/migrations
var migrationSqlAssets embed.FS

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

	logger := logging.InitLogger(logFile)

	logger.Info().Msgf("Swervo version: %s, commit SHA: %s", Version, CommitSha)
	logger.Info().Msgf("app data directory: [%s]", appDataDir)

	dataStore := datastore.New(appDataDir, "swervo.db")

	if !generateBindingsRun {
		migrationRunner, err := migrations.New(migrationSqlAssets, "db/migrations", dataStore, logger)

		if err != nil {
			logger.Fatal().Err(err).Msg("could not read migrations from embedded filesystem")
		}

		if err := migrationRunner.RunSafe(); err != nil {
			// TODO: report it!
			logger.Error().Err(err).Msg("migrations run failed")
		}
	}

	var appdata AppData

	appController := NewAppController()
	awsIdcController := NewAwsIdentityCenterController(&appdata)

	logger.Info().Msgf("PID [%d] - launching Swervo", os.Getpid())
	if err := wails.Run(&options.App{
		Title:  "Swervo",
		Width:  1024,
		Height: 768,
		Menu:   appController.mainMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			appController.init(logger.WithContext(ctx))
			awsIdcController.startup(logger.WithContext(ctx))
		},
		Bind: []interface{}{
			appController,
			awsIdcController,
		},
	}); err != nil {
		logger.Fatal().Err(err).Msg("Could not run Wails app")
	}
}
