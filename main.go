package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"
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

type MyLogger struct{}

func (myLogger *MyLogger) Printf(format string, v ...interface{}) {
	fmt.Printf(format, v...)
}

func (myLogger *MyLogger) Verbose() bool {
	return true
}

func isWailsRunningAppToGenerateBindings(osArgs []string) bool {
	for _, arg := range osArgs {
		if strings.HasSuffix(arg, "wailsbindings") {
			return true
		}
	}

	return false
}

func getAppDataDir() (string, error) {
	if BuildType == "debug" {
		if pwd, err := os.Executable(); err != nil {
			return "", err
		} else {
			if strings.Contains(strings.ToLower(pwd), "swervo.app") {
				pwd = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(pwd))))
			} else {
				pwd = filepath.Dir(pwd)
			}
			return filepath.Join(pwd, "swervo_data"), nil
		}
	} else {
		if userHomeDir, err := os.UserHomeDir(); err != nil {
			return "", err
		} else {
			return filepath.Join(userHomeDir, "swervo_data"), nil
		}
	}
}

func initializeAppDataDir(appDataDir string) {
	if err := os.MkdirAll(appDataDir, 0700); err != nil {
		if err != nil {
			log.Fatal("Could not create app data directory")
		}
	}
}

func initializeLogFile(appDataDir, logFileName string) (*os.File, error) {
	file, logFileErr := os.OpenFile(
		filepath.Join(appDataDir, "swervo_log.json"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if logFileErr != nil {
		return nil, logFileErr
	}

	return file, nil
}

func main() {
	generateBindingsRun := isWailsRunningAppToGenerateBindings(os.Args)

	appDataDir, appDataDirErr := getAppDataDir()

	if appDataDirErr != nil {
		log.Fatalf("Failed to determine app data directory: [%s]", appDataDir)
	}

	var logFile = io.Discard

	if !generateBindingsRun {
		initializeAppDataDir(appDataDir)

		file, logFileErr := initializeLogFile(appDataDir, "swervo_log.json")

		if logFileErr != nil {
			log.Fatalf("Failed to initialize log file: [%s]", logFileErr)
		}

		logFile = file

		defer file.Close()
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	logSink := zerolog.MultiLevelWriter(consoleWriter, logFile)

	logger := zerolog.New(logSink).With().Timestamp().Logger()

	log.SetFlags(0)
	log.SetOutput(logger)

	logger.Info().Msgf("Swervo version: %s, commit SHA: %s", Version, CommitSha)
	logger.Info().Msgf("App data directory: [%s]", appDataDir)

	appController := NewAppController()

	if !generateBindingsRun {
		dbFile := filepath.Join(appDataDir, "swervo.db")

		if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
			errDir := os.MkdirAll(appDataDir, 0700)

			if errDir != nil {
				logger.Fatal().Err(errDir).Msg("Could not create app storage directory")
			}
		}

		migrationsSrc, err := iofs.New(migrationSqlAssets, "db/migrations")
		if err != nil {
			logger.Fatal().Err(err).Msg("Could not create migrations source")
		}

		dbConnectionString := fmt.Sprintf("sqlite3://%s", strings.ReplaceAll(dbFile, "\\", "/"))

		if m, err := migrate.NewWithSourceInstance("iofs", migrationsSrc, dbConnectionString); err != nil {
			logger.Fatal().Err(err).Msg("Could not create migrations instance")
		} else {
			version, dirty, err := m.Version()

			if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
				logger.Fatal().Err(err).Msg("Could not get migrations version")
			}

			logger.Printf("Running migrations against dirty?[%t] database@[%d]", dirty, version)
			if err := m.Up(); err != nil && err != migrate.ErrNoChange {
				logger.Fatal().Err(err).Msg("Migrations run failed")
			}
		}
	}

	var appdata AppData

	awsIdcController := NewAwsIdentityCenterController(&appdata)

	var onStartup = func(ctx context.Context) {
		appController.init(logger.WithContext(ctx))
		awsIdcController.startup(logger.WithContext(ctx))
	}

	if err := wails.Run(&options.App{
		Title:  "Swervo",
		Width:  1024,
		Height: 768,
		Menu:   appController.mainMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: onStartup,
		Bind: []interface{}{
			appController,
			awsIdcController,
		},
	}); err != nil {
		logger.Fatal().Err(err).Msg("Could not run Wails app")
	}
}
