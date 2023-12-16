package main

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/plumbing"
	"github.com/abjrcode/swervo/internal/utils"
	awsidc "github.com/abjrcode/swervo/providers/aws_idc"
	awscredssink "github.com/abjrcode/swervo/sinks/awscredssink"
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

	awsIdcController *awsidc.AwsIdentityCenterController

	awsCredentialsSinkController *awscredssink.AwsCredentialsSinkController
}

func (c *AppController) init(ctx context.Context, errorHandler app.ErrorHandler) {
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
		c.authController.LockVault(appContext)
	case "Dashboard_ListProviders":
		output = c.dashboardController.ListProviders()
	case "Dashboard_ListCompatibleSinks":
		output = c.dashboardController.ListCompatibleSinks(appContext, commandInput["providerCode"].(string))
	case "Dashboard_ListFavorites":
		output, err = c.dashboardController.ListFavorites(appContext)
	case "AwsIdc_ListInstances":
		output, err = c.awsIdcController.ListInstances(appContext)
	case "AwsIdc_GetInstanceData":
		output, err = c.awsIdcController.GetInstanceData(appContext,
			commandInput["instanceId"].(string),
			commandInput["forceRefresh"].(bool))
	case "AwsIdc_CopyRoleCredentials":
		err = c.awsIdcController.CopyRoleCredentials(appContext,
			awsidc.AwsIdc_CopyRoleCredentialsCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				AccountId:  commandInput["accountId"].(string),
				RoleName:   commandInput["roleName"].(string),
			})
	case "AwsIdc_SaveRoleCredentials":
		err = c.awsIdcController.SaveRoleCredentials(appContext,
			awsidc.AwsIdc_SaveRoleCredentialsCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				AccountId:  commandInput["accountId"].(string),
				RoleName:   commandInput["roleName"].(string),
				AwsProfile: commandInput["awsProfile"].(string),
			})
	case "AwsIdc_Setup":
		output, err = c.awsIdcController.Setup(appContext,
			awsidc.AwsIdc_SetupCommandInput{
				StartUrl:  commandInput["startUrl"].(string),
				AwsRegion: commandInput["awsRegion"].(string),
				Label:     commandInput["label"].(string),
			})
	case "AwsIdc_FinalizeSetup":
		output, err = c.awsIdcController.FinalizeSetup(appContext,
			awsidc.AwsIdc_FinalizeSetupCommandInput{
				ClientId:   commandInput["clientId"].(string),
				StartUrl:   commandInput["startUrl"].(string),
				AwsRegion:  commandInput["awsRegion"].(string),
				Label:      commandInput["label"].(string),
				UserCode:   commandInput["userCode"].(string),
				DeviceCode: commandInput["deviceCode"].(string),
			})
	case "AwsIdc_MarkAsFavorite":
		err = c.awsIdcController.MarkAsFavorite(appContext, commandInput["instanceId"].(string))
	case "AwsIdc_UnmarkAsFavorite":
		err = c.awsIdcController.UnmarkAsFavorite(appContext, commandInput["instanceId"].(string))
	case "AwsIdc_RefreshAccessToken":
		output, err = c.awsIdcController.RefreshAccessToken(appContext, commandInput["instanceId"].(string))
	case "AwsIdc_FinalizeRefreshAccessToken":
		err = c.awsIdcController.FinalizeRefreshAccessToken(appContext,
			awsidc.AwsIdc_FinalizeRefreshAccessTokenCommandInput{
				InstanceId: commandInput["instanceId"].(string),
				Region:     commandInput["region"].(string),
				UserCode:   commandInput["userCode"].(string),
				DeviceCode: commandInput["deviceCode"].(string),
			})
	case "AwsCredentialsSink_NewInstance":
		output, err = c.awsCredentialsSinkController.NewInstance(appContext,
			awscredssink.AwsCredentialsSink_NewInstanceCommandInput{
				FilePath:       commandInput["filePath"].(string),
				AwsProfileName: commandInput["awsProfileName"].(string),
				Label:          commandInput["label"].(string),
				ProviderCode:   commandInput["providerCode"].(string),
				ProviderId:     commandInput["providerId"].(string),
			})
	case "AwsCredentialsSink_GetInstanceData":
		output, err = c.awsCredentialsSinkController.GetInstanceData(appContext,
			commandInput["instanceId"].(string),
		)
	case "AwsCredentialsSink_DisconnectSink":
		err = c.awsCredentialsSinkController.DisconnectSink(appContext, plumbing.DisconnectSinkCommandInput{
			SinkCode: commandInput["sinkCode"].(string),
			SinkId:   commandInput["sinkId"].(string),
		})
	default:
		output, err = nil, errors.Join(ErrInvalidAppCommand, app.ErrFatal)
	}

	if errors.Is(err, app.ErrFatal) {
		c.errorHandler.Catch(appContext, c.logger, err)
	}

	if errors.Is(err, app.ErrValidation) {
		return output, errors.Unwrap(err)
	}

	return output, err
}
