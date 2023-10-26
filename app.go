package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type AppController struct {
	ctx      context.Context
	mainMenu *menu.Menu
	logger   *zerolog.Logger
}

func NewAppController() *AppController {
	appMenu := menu.NewMenu()

	controller := &AppController{
		mainMenu: appMenu,
	}

	FileMenu := appMenu.AddSubmenu("File")
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsRuntime.Quit(controller.ctx)
	})

	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.EditMenu())
	}

	HelpMenu := appMenu.AddSubmenu("Help")
	HelpMenu.AddText("About", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		wailsRuntime.MessageDialog(controller.ctx, wailsRuntime.MessageDialogOptions{
			Title:   "About",
			Message: fmt.Sprintf("Swervo %s\nBuilt @ %s\nCommit SHA: %s\nBuild Link: %s", Version, BuildTimestamp, CommitSha, BuildLink),
		})
	})

	return controller
}

func (app *AppController) Init(ctx context.Context) {
	app.ctx = ctx
	app.logger = zerolog.Ctx(ctx)
}
