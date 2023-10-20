package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed db/migrations
var migrationSqlAssets embed.FS

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

func main() {
	app := NewApp()

	AppMenu := menu.NewMenu()

	FileMenu := AppMenu.AddSubmenu("File")
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsRuntime.Quit(app.ctx)
	})

	if runtime.GOOS == "darwin" {
		AppMenu.Append(menu.EditMenu())
	}

	HelpMenu := AppMenu.AddSubmenu("Help")
	HelpMenu.AddText("About", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		wailsRuntime.MessageDialog(app.ctx, wailsRuntime.MessageDialogOptions{
			Title:   "About",
			Message: fmt.Sprintf("Swervo %s\nBuilt @ %s\nCommit SHA: %s\nBuild Link: %s", Version, BuildTimestamp, CommitSha, BuildLink),
		})
	})

	var appDataDir string

	if userHomeDir, err := os.UserHomeDir(); err != nil {
		log.Fatal(err)
	} else {
		appDataDir = userHomeDir
	}

	appStorageDir := filepath.Join(appDataDir, "swervo")

	// appConfigFile := filepath.Join(appStorageDir, "swervo.toml")
	appDbFile := filepath.Join(appStorageDir, "swervo.db")

	if _, err := os.Stat(appStorageDir); os.IsNotExist(err) {
		errDir := os.MkdirAll(appStorageDir, 0700)

		if errDir != nil {
			log.Fatal(errDir)
		}
	}

	migrationsSrc, err := iofs.New(migrationSqlAssets, "db/migrations")
	if err != nil {
		log.Fatal(err)
	}

	dbConnectionString := fmt.Sprintf("sqlite3://%s", strings.ReplaceAll(appDbFile, "\\", "/"))

	if m, err := migrate.NewWithSourceInstance("iofs", migrationsSrc, dbConnectionString); err != nil {
		log.Fatal(err)
	} else {
		log.Println("Running migrations against database: ", appDbFile)
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal(err)
		}
	}

	var appdata AppData

	awsIdcc := NewAwsIdentityCenterController(&appdata)

	var onStartup = func(ctx context.Context) {
		app.startup(ctx)
		awsIdcc.startup(ctx)
	}

	if err := wails.Run(&options.App{
		Title:  "Swervo",
		Width:  1024,
		Height: 768,
		Menu:   AppMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: onStartup,
		Bind: []interface{}{
			app,
			awsIdcc,
		},
	}); err != nil {
		println("Error:", err.Error())
	}
}
