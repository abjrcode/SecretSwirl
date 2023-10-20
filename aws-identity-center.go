package main

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssooidc"
	"github.com/rs/zerolog"
)

type AwsIdentityCenterController struct {
	ctx     context.Context
	logger  *zerolog.Logger
	appData *AppData
}

func NewAwsIdentityCenterController(appData *AppData) *AwsIdentityCenterController {
	return &AwsIdentityCenterController{
		appData: appData,
	}
}

func (controller *AwsIdentityCenterController) startup(ctx context.Context) {
	controller.ctx = ctx
	controller.logger = zerolog.Ctx(ctx)
}

func (controller *AwsIdentityCenterController) RegisterClient(friendlyName string) error {
	if controller.appData.AwsIdc.ClientId != "" {
		return nil
	}

	session := session.Must(session.NewSession())
	controller.logger.Info().Msg("Creating SSO OIDC client")
	client := ssooidc.New(session, &aws.Config{Region: aws.String("eu-central-1")})

	output, err := client.RegisterClient(&ssooidc.RegisterClientInput{
		ClientName: aws.String(friendlyName),
		ClientType: aws.String("public"),
	})

	if err != nil {
		return err
	}

	controller.appData.AwsIdc.ClientId = *output.ClientId
	controller.appData.AwsIdc.ClientSecret = *output.ClientSecret
	controller.appData.AwsIdc.ClientIdIssuedAt = *output.ClientIdIssuedAt
	controller.appData.AwsIdc.ExpiresAt = *output.ClientSecretExpiresAt

	return nil
}

func (controller *AwsIdentityCenterController) AuthorizeDevice() error {
	session := session.Must(session.NewSession())
	controller.logger.Info().Msg("Authorizing Device")
	client := ssooidc.New(session, &aws.Config{Region: aws.String("eu-central-1")})

	output, err := client.StartDeviceAuthorization(&ssooidc.StartDeviceAuthorizationInput{
		ClientId:     aws.String(controller.appData.AwsIdc.ClientId),
		ClientSecret: aws.String(controller.appData.AwsIdc.ClientSecret),
		StartUrl:     aws.String("https://d-99670c0d3d.awsapps.com/start"),
	})

	if err != nil {
		return err
	}

	controller.appData.AwsIdc.DeviceCode = *output.DeviceCode

	controller.logger.Info().Msgf("Please visit %s and enter code %s", *output.VerificationUri, *output.UserCode)

	return nil
}

func (controller *AwsIdentityCenterController) CreateToken() error {
	session := session.Must(session.NewSession())
	controller.logger.Info().Msg("Creating Token")
	client := ssooidc.New(session, &aws.Config{Region: aws.String("eu-central-1")})

	output, err := client.CreateToken(&ssooidc.CreateTokenInput{
		ClientId:     aws.String(controller.appData.AwsIdc.ClientId),
		ClientSecret: aws.String(controller.appData.AwsIdc.ClientSecret),
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		DeviceCode:   aws.String(controller.appData.AwsIdc.DeviceCode),
	})

	if err != nil {
		log.Fatal(err)
	}

	controller.appData.AwsIdc.AccessToken = *output.AccessToken
	controller.appData.AwsIdc.AccessTokenExpiresInSeconds = *output.ExpiresIn
	//awdIdcController.appData.AwsIdc.RefreshToken = *output.RefreshToken
	controller.appData.AwsIdc.TokenType = *output.TokenType

	controller.logger.Info().Msgf("Access Token: %s", *output.AccessToken)

	return nil
}
