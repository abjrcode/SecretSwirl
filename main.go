package main

import (
	"embed"
	"fmt"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

var Version string = "v0.0.0"
var BuildTimestamp string = "NOW"
var CommitSha string = "HEAD"
var BuildLink string = "http://localhost"

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

	err := wails.Run(&options.App{
		Title:  "Swervo",
		Width:  1024,
		Height: 768,
		Menu:   AppMenu,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
