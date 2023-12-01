package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/utils"
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
	errorHandler app.ErrorHandler

	authController      *AuthController
	dashboardController *DashboardController
	awsIamIdcController *awsiamidc.AwsIdentityCenterController
}

func (c *AppController) Init(ctx context.Context, errorHandler app.ErrorHandler) {
	appMenu := menu.NewMenu()

	logger := zerolog.Ctx(ctx).With().Str("component", "app_controller").Logger()

	c.ctx = ctx
	c.logger = logger
	c.errorHandler = errorHandler
	c.mainMenu = appMenu

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

func (c *AppController) ShowErrorDialog(msg string) {
	wailsRuntime.MessageDialog(c.ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.ErrorDialog,
		Title:   "Error",
		Message: msg,
	})
}

func (c *AppController) ShowWarningDialog(msg string) {
	wailsRuntime.MessageDialog(c.ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.WarningDialog,
		Title:   "Warning",
		Message: msg,
	})
}

func (c *AppController) CatchUnhandledError(msg string) {
	reqId := utils.NewRequestId()

	userId := "root"

	ctx := app.NewContext(c.ctx, userId, reqId, reqId, reqId, &c.logger)
	c.errorHandler.Catch(ctx, c.logger, errors.New(msg))
}

func (c *AppController) RunAppCommand(command string, commandInput map[string]any) (any, error) {

	componentName, _, ok := strings.Cut(command, "_")

	if !ok {
		componentName = "unknown"
	}

	reqId := utils.NewRequestId()
	userId := "root"

	logger := c.logger.With().Str("component", componentName).Str("req_id", reqId).Str("user_id", userId).Logger()

	logger.Trace().Msgf("running command: [%s]", command)

	appContext := app.NewContext(
		logger.WithContext(c.ctx),
		userId,
		reqId,
		reqId,
		reqId,
		&logger,
	)

	var output any = nil
	var err error = nil

	switch command {
	case "Auth_IsVaultConfigured":
		output, err = c.authController.IsVaultConfigured(appContext)
	case "Auth_ConfigureVault":
		err = c.authController.ConfigureVault(
			appContext,
			Auth_ConfigureVaultCommandInput{
				Password: commandInput["password"].(string),
			})
	case "Auth_Unlock":
		output, err = c.authController.UnlockVault(
			appContext,
			Auth_UnlockCommandInput{
				Password: commandInput["password"].(string),
			})
	case "Auth_Lock":
		c.authController.LockVault()
	case "Dashboard_ListProviders":
		output = c.dashboardController.ListProviders()
	case "Dashboard_ListFavorites":
		output, err = c.dashboardController.ListFavorites(appContext)
	case "AwsIamIdc_ListInstances":
		output, err = c.awsIamIdcController.ListInstances(appContext)
	case "AwsIamIdc_GetInstanceData":
		output, err = c.awsIamIdcController.GetInstanceData(appContext,
			commandInput["instanceId"].(string),
			commandInput["forceRefresh"].(bool))
	case "AwsIamIdc_GetRoleCredentials":
		output, err = c.awsIamIdcController.GetRoleCredentials(appContext,
			awsiamidc.AwsIamIdc_GetRoleCredentialsCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				AccountId:  commandInput["accountId"].(string),
				RoleName:   commandInput["roleName"].(string),
			})
	case "AwsIamIdc_Setup":
		output, err = c.awsIamIdcController.Setup(appContext,
			awsiamidc.AwsIamIdc_SetupCommandInput{
				StartUrl:  commandInput["startUrl"].(string),
				AwsRegion: commandInput["awsRegion"].(string),
				Label:     commandInput["label"].(string),
			})
	case "AwsIamIdc_FinalizeSetup":
		output, err = c.awsIamIdcController.FinalizeSetup(appContext,
			awsiamidc.AwsIamIdc_FinalizeSetupCommandInput{
				ClientId:   commandInput["clientId"].(string),
				StartUrl:   commandInput["startUrl"].(string),
				AwsRegion:  commandInput["awsRegion"].(string),
				Label:      commandInput["label"].(string),
				UserCode:   commandInput["userCode"].(string),
				DeviceCode: commandInput["deviceCode"].(string),
			})
	case "AwsIamIdc_MarkAsFavorite":
		err = c.awsIamIdcController.MarkAsFavorite(appContext, commandInput["instanceId"].(string))
	case "AwsIamIdc_UnmarkAsFavorite":
		err = c.awsIamIdcController.UnmarkAsFavorite(appContext, commandInput["instanceId"].(string))
	case "AwsIamIdc_RefreshAccessToken":
		output, err = c.awsIamIdcController.RefreshAccessToken(appContext, commandInput["instanceId"].(string))
	case "AwsIamIdc_FinalizeRefreshAccessToken":
		err = c.awsIamIdcController.FinalizeRefreshAccessToken(appContext,
			awsiamidc.AwsIamIdc_FinalizeRefreshAccessTokenCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				Region:     commandInput["region"].(string),
				UserCode:   commandInput["userCode"].(string),
				DeviceCode: commandInput["deviceCode"].(string),
			})
	default:
		output, err = nil, errors.Join(ErrInvalidAppCommand, app.ErrFatal)
	}

	if errors.Is(err, app.ErrFatal) {
		c.errorHandler.Catch(appContext, c.logger, err)
	}

	return output, err
}
