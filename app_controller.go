package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/abjrcode/swervo/internal/faults"
	awsiamidc "github.com/abjrcode/swervo/providers/aws_iam_idc"
	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	ErrInvalidAppCommand = errors.New("INVALID_APP_COMMAND")
)

type AppController struct {
	ctx          context.Context
	mainMenu     *menu.Menu
	logger       zerolog.Logger
	errorHandler faults.ErrorHandler

	authController      *AuthController
	dashboardController *DashboardController
	awsIamIdcController *awsiamidc.AwsIdentityCenterController
}

func (app *AppController) Init(ctx context.Context, errorHandler faults.ErrorHandler) {
	appMenu := menu.NewMenu()

	enrichedLogger := zerolog.Ctx(ctx).With().Str("component", "app_controller").Logger()

	app.ctx = ctx
	app.logger = enrichedLogger
	app.errorHandler = errorHandler
	app.mainMenu = appMenu

	FileMenu := appMenu.AddSubmenu("File")
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		wailsRuntime.Quit(ctx)
	})

	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.EditMenu())
	}

	HelpMenu := appMenu.AddSubmenu("Help")
	HelpMenu.AddText("About", keys.CmdOrCtrl("h"), func(_ *menu.CallbackData) {
		wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
			Title:   "About",
			Message: fmt.Sprintf("Swervo %s\nBuilt @ %s\nCommit SHA: %s\nBuild Link: %s", Version, BuildTimestamp, CommitSha, BuildLink),
		})
	})
}

func (app *AppController) ShowErrorDialog(msg string) {
	wailsRuntime.MessageDialog(app.ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.ErrorDialog,
		Title:   "Error",
		Message: msg,
	})
}

func (app *AppController) ShowWarningDialog(msg string) {
	wailsRuntime.MessageDialog(app.ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.WarningDialog,
		Title:   "Warning",
		Message: msg,
	})
}

func (app *AppController) CatchUnhandledError(msg string) {
	app.errorHandler.Catch(app.logger, errors.New(msg))
}

func (app *AppController) RunAppCommand(command string, commandInput map[string]any) (any, error) {
	app.logger.Trace().Msgf("running command: [%s]", command)

	var output any = nil
	var err error = nil

	switch command {
	case "Auth_IsVaultConfigured":
		output, err = app.authController.IsVaultConfigured(app.ctx)
	case "Auth_ConfigureVault":
		err = app.authController.ConfigureVault(
			app.ctx,
			Auth_ConfigureVaultCommandInput{
				Password: commandInput["password"].(string),
			})
	case "Auth_Unlock":
		output, err = app.authController.UnlockVault(
			app.ctx,
			Auth_UnlockCommandInput{
				Password: commandInput["password"].(string),
			})
	case "Auth_Lock":
		app.authController.LockVault()
	case "Dashboard_ListProviders":
		output = app.dashboardController.ListProviders()
	case "Dashboard_ListFavorites":
		output, err = app.dashboardController.ListFavorites(app.ctx)
	case "AwsIamIdc_ListInstances":
		output, err = app.awsIamIdcController.ListInstances(app.ctx)
	case "AwsIamIdc_GetInstanceData":
		output, err = app.awsIamIdcController.GetInstanceData(app.ctx,
			commandInput["instanceId"].(string),
			commandInput["forceRefresh"].(bool))
	case "AwsIamIdc_GetRoleCredentials":
		output, err = app.awsIamIdcController.GetRoleCredentials(app.ctx,
			awsiamidc.AwsIamIdc_GetRoleCredentialsCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				AccountId:  commandInput["accountId"].(string),
				RoleName:   commandInput["roleName"].(string),
			})
	case "AwsIamIdc_Setup":
		output, err = app.awsIamIdcController.Setup(app.ctx,
			awsiamidc.AwsIamIdc_SetupCommandInput{
				StartUrl:  commandInput["startUrl"].(string),
				AwsRegion: commandInput["awsRegion"].(string),
				Label:     commandInput["label"].(string),
			})
	case "AwsIamIdc_FinalizeSetup":
		output, err = app.awsIamIdcController.FinalizeSetup(app.ctx,
			awsiamidc.AwsIamIdc_FinalizeSetupCommandInput{
				ClientId:   commandInput["clientId"].(string),
				StartUrl:   commandInput["startUrl"].(string),
				AwsRegion:  commandInput["awsRegion"].(string),
				Label:      commandInput["label"].(string),
				UserCode:   commandInput["userCode"].(string),
				DeviceCode: commandInput["deviceCode"].(string),
			})
	case "AwsIamIdc_MarkAsFavorite":
		err = app.awsIamIdcController.MarkAsFavorite(app.ctx, commandInput["instanceId"].(string))
	case "AwsIamIdc_UnmarkAsFavorite":
		err = app.awsIamIdcController.UnmarkAsFavorite(app.ctx, commandInput["instanceId"].(string))
	case "AwsIamIdc_RefreshAccessToken":
		output, err = app.awsIamIdcController.RefreshAccessToken(app.ctx, commandInput["instanceId"].(string))
	case "AwsIamIdc_FinalizeRefreshAccessToken":
		err = app.awsIamIdcController.FinalizeRefreshAccessToken(app.ctx,
			awsiamidc.AwsIamIdc_FinalizeRefreshAccessTokenCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				Region:     commandInput["region"].(string),
				UserCode:   commandInput["userCode"].(string),
				DeviceCode: commandInput["deviceCode"].(string),
			})
	default:
		output, err = nil, errors.Join(ErrInvalidAppCommand, faults.ErrFatal)
	}

	if errors.Is(err, faults.ErrFatal) {
		app.errorHandler.Catch(app.logger, err)
	}

	return output, err
}
