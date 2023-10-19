package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssooidc"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type AwsIdentityCenterController struct {
	ctx     context.Context
	appData *AppData
}

func NewAwsIdentityCenterController(appData *AppData) *AwsIdentityCenterController {
	return &AwsIdentityCenterController{
		appData: appData,
	}
}

func (awsIdcController *AwsIdentityCenterController) startup(ctx context.Context) {
	awsIdcController.ctx = ctx
}

func (awdIdcController *AwsIdentityCenterController) RegisterClient(friendlyName string) error {
	if awdIdcController.appData.AwsIdc.ClientId != "" {
		return nil
	}

	session := session.Must(session.NewSession())
	runtime.LogInfo(awdIdcController.ctx, "Creating SSO OIDC client")
	client := ssooidc.New(session, &aws.Config{Region: aws.String("eu-central-1")})

	output, err := client.RegisterClient(&ssooidc.RegisterClientInput{
		ClientName: aws.String(friendlyName),
		ClientType: aws.String("public"),
	})

	if err != nil {
		return err
	}

	awdIdcController.appData.AwsIdc.ClientId = *output.ClientId
	awdIdcController.appData.AwsIdc.ClientSecret = *output.ClientSecret
	awdIdcController.appData.AwsIdc.ClientIdIssuedAt = *output.ClientIdIssuedAt
	awdIdcController.appData.AwsIdc.ExpiresAt = *output.ClientSecretExpiresAt

	return nil
}

func (awdIdcController *AwsIdentityCenterController) AuthorizeDevice() error {
	session := session.Must(session.NewSession())
	runtime.LogInfo(awdIdcController.ctx, "Authorizing Device")
	client := ssooidc.New(session, &aws.Config{Region: aws.String("eu-central-1")})

	output, err := client.StartDeviceAuthorization(&ssooidc.StartDeviceAuthorizationInput{
		ClientId:     aws.String(awdIdcController.appData.AwsIdc.ClientId),
		ClientSecret: aws.String(awdIdcController.appData.AwsIdc.ClientSecret),
		StartUrl:     aws.String("https://my-app.awsapps.com/start"),
	})

	if err != nil {
		return err
	}

	awdIdcController.appData.AwsIdc.DeviceCode = *output.DeviceCode

	runtime.LogInfo(awdIdcController.ctx,
		fmt.Sprintf("Please visit %s and enter code %s", *output.VerificationUri, *output.UserCode))

	return nil
}

func (awdIdcController *AwsIdentityCenterController) CreateToken() error {
	session := session.Must(session.NewSession())
	runtime.LogInfo(awdIdcController.ctx, "Creating Token")
	client := ssooidc.New(session, &aws.Config{Region: aws.String("eu-central-1")})

	output, err := client.CreateToken(&ssooidc.CreateTokenInput{
		ClientId:     aws.String(awdIdcController.appData.AwsIdc.ClientId),
		ClientSecret: aws.String(awdIdcController.appData.AwsIdc.ClientSecret),
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		DeviceCode:   aws.String(awdIdcController.appData.AwsIdc.DeviceCode),
	})

	if err != nil {
		log.Fatal(err)
	}

	awdIdcController.appData.AwsIdc.AccessToken = *output.AccessToken
	awdIdcController.appData.AwsIdc.AccessTokenExpiresInSeconds = *output.ExpiresIn
	//awdIdcController.appData.AwsIdc.RefreshToken = *output.RefreshToken
	awdIdcController.appData.AwsIdc.TokenType = *output.TokenType

	runtime.LogInfo(awdIdcController.ctx, fmt.Sprintf("Access Token: %s", *output.AccessToken))

	return nil
}
