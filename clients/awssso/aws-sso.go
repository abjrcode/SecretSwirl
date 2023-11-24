package awssso

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
)

var SupportedAwsRegions = map[string]string{
	"us-east-1":      "US East (N. Virginia)",
	"us-east-2":      "US East (Ohio)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"af-south-1":     "Africa (Cape Town)",
	"ap-east-1":      "Asia Pacific (Hong Kong)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"ap-northeast-3": "Asia Pacific (Osaka-Local)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ca-central-1":   "Canada (Central)",
	"cn-north-1":     "China (Beijing)",
	"cn-northwest-1": "China (Ningxia)",
	"eu-central-1":   "Europe (Frankfurt)",
	"eu-west-1":      "Europe (Ireland)",
	"eu-west-2":      "Europe (London)",
	"eu-south-1":     "Europe (Milan)",
	"eu-west-3":      "Europe (Paris)",
	"eu-north-1":     "Europe (Stockholm)",
	"me-south-1":     "Middle East (Bahrain)",
	"sa-east-1":      "South America (São Paulo)",
}

var (
	ErrInvalidRequest          = errors.New("request is not valid")
	ErrDeviceFlowNotAuthorized = errors.New("device flow not authorized")
	ErrDeviceCodeExpired       = errors.New("device code expired")
	ErrAccessTokenExpired      = errors.New("device code expired")
)

type AwsRegion string

type RegistrationResponse struct {
	ClientId, ClientSecret string
	CreatedAt, ExpiresAt   int64
}

type AuthorizationResponse struct {
	VerificationUri, VerificationUriComplete string
	UserCode, DeviceCode                     string
	Interval                                 int32
	ExpiresIn                                int32
}

type GetTokenResponse struct {
	IdToken, AccessToken, RefreshToken, TokenType string
	ExpiresIn                                     int32
}

type AwsAccountRole struct {
	RoleName string
}

type AwsAccount struct {
	AccountId, AccountEmail, AccountName string
	Roles                                []AwsAccountRole
}

type ListAccountsResponse struct {
	Accounts []AwsAccount
}

type AwsSsoOidcClient interface {
	RegisterClient(ctx context.Context, friendlyClientName string) (*RegistrationResponse, error)

	StartDeviceAuthorization(ctx context.Context, startUrl string, clientId, clientSecret string) (*AuthorizationResponse, error)

	CreateToken(ctx context.Context, clientId, clientSecret, userCode, deviceCode string) (*GetTokenResponse, error)

	ListAccounts(ctx context.Context, accessToken string) (*ListAccountsResponse, error)
}

type awsSsoClientImpl struct {
	oidcClient *ssooidc.Client
	ssoClient  *sso.Client
}

func NewAwsSsoOidcClient() AwsSsoOidcClient {
	return &awsSsoClientImpl{
		oidcClient: ssooidc.NewFromConfig(aws.Config{}),
		ssoClient:  sso.NewFromConfig(aws.Config{}),
	}
}

func (c *awsSsoClientImpl) RegisterClient(ctx context.Context, friendlyClientName string) (*RegistrationResponse, error) {
	output, err := c.oidcClient.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String(friendlyClientName),
		ClientType: aws.String("public"),
	}, func(options *ssooidc.Options) {
		options.Region = ctx.Value(AwsRegion("awsRegion")).(string)
	})

	if err != nil {
		return nil, err
	}

	return &RegistrationResponse{
		ClientId:     *output.ClientId,
		ClientSecret: *output.ClientSecret,
		CreatedAt:    output.ClientIdIssuedAt,
		ExpiresAt:    output.ClientSecretExpiresAt,
	}, nil
}

func (c *awsSsoClientImpl) StartDeviceAuthorization(ctx context.Context, startUrl string, clientId, clientSecret string) (*AuthorizationResponse, error) {
	output, err := c.oidcClient.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     aws.String(clientId),
		ClientSecret: aws.String(clientSecret),
		StartUrl:     aws.String(startUrl),
	}, func(options *ssooidc.Options) {
		options.Region = ctx.Value(AwsRegion("awsRegion")).(string)
	})

	if err != nil {
		var ire *types.InvalidRequestException

		if errors.As(err, &ire) {
			return nil, ErrInvalidRequest
		}
		return nil, err
	}

	return &AuthorizationResponse{
		VerificationUri:         *output.VerificationUri,
		VerificationUriComplete: *output.VerificationUriComplete,
		UserCode:                *output.UserCode,
		DeviceCode:              *output.DeviceCode,
		ExpiresIn:               output.ExpiresIn,
		Interval:                output.Interval,
	}, nil
}

func (c *awsSsoClientImpl) CreateToken(ctx context.Context, clientId, clientSecret, userCode, deviceCode string) (*GetTokenResponse, error) {
	output, err := c.oidcClient.CreateToken(ctx, &ssooidc.CreateTokenInput{
		ClientId:     aws.String(clientId),
		ClientSecret: aws.String(clientSecret),
		GrantType:    aws.String("urn:ietf:params:oauth:grant-type:device_code"),
		DeviceCode:   aws.String(deviceCode),
		Code:         aws.String(userCode),
	}, func(options *ssooidc.Options) {
		options.Region = ctx.Value(AwsRegion("awsRegion")).(string)
	})

	if err != nil {
		var ape *types.AuthorizationPendingException

		if errors.As(err, &ape) {
			return nil, ErrDeviceFlowNotAuthorized
		}

		return nil, err
	}

	return &GetTokenResponse{
		IdToken:      "", // Not supported by AWS SSO
		AccessToken:  *output.AccessToken,
		RefreshToken: "", // Not supported by AWS SSO
		TokenType:    *output.TokenType,
		ExpiresIn:    output.ExpiresIn,
	}, nil
}

func (c *awsSsoClientImpl) ListAccounts(ctx context.Context, accessToken string) (*ListAccountsResponse, error) {
	output, err := c.ssoClient.ListAccounts(ctx, &sso.ListAccountsInput{
		AccessToken: aws.String(accessToken),
	}, func(options *sso.Options) {
		options.Region = ctx.Value(AwsRegion("awsRegion")).(string)
	})

	if err != nil {
		var ete *types.ExpiredTokenException

		if errors.As(err, &ete) {
			return nil, ErrAccessTokenExpired
		}

		return nil, err
	}

	var accounts []AwsAccount

	for _, account := range output.AccountList {
		var roles []AwsAccountRole

		accountRoles, err := c.ssoClient.ListAccountRoles(ctx, &sso.ListAccountRolesInput{
			AccessToken: aws.String(accessToken),
			AccountId:   account.AccountId,
		}, func(options *sso.Options) {
			options.Region = ctx.Value(AwsRegion("awsRegion")).(string)
		})

		if err != nil {
			return nil, err
		}

		for _, role := range accountRoles.RoleList {
			roles = append(roles, AwsAccountRole{
				RoleName: *role.RoleName,
			})
		}

		accounts = append(accounts, AwsAccount{
			AccountId:    *account.AccountId,
			AccountEmail: *account.EmailAddress,
			AccountName:  *account.AccountName,
			Roles:        roles,
		})
	}

	return &ListAccountsResponse{
		Accounts: accounts,
	}, nil
}
