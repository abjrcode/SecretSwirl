package main

import (
	"context"
	"database/sql"
	"embed"
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

	logger := logging.InitLogger(logFile)

	logger.Info().Msgf("Swervo version: %s, commit SHA: %s", Version, CommitSha)
	logger.Info().Msgf("app data directory: [%s]", appDataDir)

	var sqlDb *sql.DB

	if !generateBindingsRun {
		dataStore := datastore.New(appDataDir, "swervo.db")
		migrationRunner, err := migrations.New(migrations.DefaultMigrationsFs, "scripts", dataStore, logger)

		if err != nil {
			logger.Fatal().Err(err).Msg("could not read migrations from embedded filesystem")
		}

		if err := migrationRunner.RunSafe(); err != nil {
			// TODO: report it!
			logger.Error().Err(err).Msg("migrations run failed")
		}

		sqlDb, err = dataStore.Open()

		if err != nil {
			logger.Fatal().Err(err).Msg("could not open database")
		}

		defer sqlDb.Close()
	}

	appController := NewAppController()
	timeProvider := utils.NewDatetime()
	vault := vault.NewVault(sqlDb, timeProvider)
	defer vault.Close()

	authController := &AuthController{
		logger: &logger,
		vault:  vault,
	}
	dashboardController := &DashboardController{
		logger: &logger,
		db:     sqlDb,
	}

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
		OnStartup: func(ctx context.Context) {
			appController.Init(logger.WithContext(ctx))
			authController.Init(logger.WithContext(ctx))
			dashboardController.Init(logger.WithContext(ctx))
			awsIdcController.Init(logger.WithContext(ctx))
		},
		Bind: []interface{}{
			appController,
			authController,
			dashboardController,
			awsIdcController,
		},
	}); err != nil {
		logger.Fatal().Err(err).Msg("Could not run Wails app")
	}
}
