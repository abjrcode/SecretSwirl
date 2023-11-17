package awsiamidc

import (
	"context"
	"net/url"
	"testing"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAwsSsoOidcClient struct {
	mock.Mock
}

func (m *mockAwsSsoOidcClient) RegisterClient(ctx context.Context, friendlyClientName string) (*awssso.RegistrationResponse, error) {
	args := m.Called(ctx, friendlyClientName)
	res, _ := args.Get(0).(*awssso.RegistrationResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) StartDeviceAuthorization(ctx context.Context, startUrl url.URL, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	args := m.Called(ctx, startUrl, clientId, clientSecret)
	res, _ := args.Get(0).(*awssso.AuthorizationResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) CreateToken(ctx context.Context, clientId, clientSecret, userCode, deviceCode string) (*awssso.GetTokenResponse, error) {
	args := m.Called(ctx, clientId, clientSecret, userCode, deviceCode)
	res, _ := args.Get(0).(*awssso.GetTokenResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) ListAccounts(ctx context.Context, accessToken string) (*awssso.ListAccountsResponse, error) {
	args := m.Called(ctx, accessToken)
	res, _ := args.Get(0).(*awssso.ListAccountsResponse)
	return res, args.Error(1)
}

type mockDatetime struct {
	mock.Mock
}

func (m *mockDatetime) NowUnix() int64 {
	args := m.Called()
	return int64(args.Int(0))
}

func initController(t *testing.T) (*AwsIdentityCenterController, *mockAwsSsoOidcClient, *mockDatetime) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws-iam-idc-controller-tests.db")

	require.NoError(t, err)

	awsClient := new(mockAwsSsoOidcClient)
	mockDatetime := new(mockDatetime)
	logger := zerolog.Nop()
	errHandler := testhelpers.NewMockErrorHandler(t)
	ctx := logger.WithContext(context.Background())
	vault := vault.NewVault(db, mockDatetime, &logger, errHandler)
	timeSetCall := mockDatetime.On("NowUnix").Return(1)
	err = vault.ConfigureKey(context.Background(), "abc")
	require.NoError(t, err)
	controller := NewAwsIdentityCenterController(db, vault, awsClient, mockDatetime)
	controller.Init(ctx, errHandler)

	timeSetCall.Unset()

	return controller, awsClient, mockDatetime
}

func simulateSuccessfulSetup(t *testing.T, startUrl, region string) (*AwsIdentityCenterController, *mockAwsSsoOidcClient, *mockDatetime) {
	controller, mockAws, mockTimeProvider := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    1,
		ExpiresAt:    20,
	}
	regCall := mockAws.On("RegisterClient", mock.Anything, mock.Anything).Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               10,
	}
	deviceAuthCall := mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mockAuthRes, nil)

	setupResult, err := controller.Setup(startUrl, region)
	require.NoError(t, err)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "test-token-type",
		ExpiresIn:    5,
	}
	createTokenCall := mockAws.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mockTokenRes, nil)

	tokenCreatedAt := 2
	timeSetCall := mockTimeProvider.On("NowUnix").Return(tokenCreatedAt)

	err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.UserCode, setupResult.DeviceCode)
	require.NoError(t, err)

	regCall.Unset()
	deviceAuthCall.Unset()
	createTokenCall.Unset()
	timeSetCall.Unset()

	return controller, mockAws, mockTimeProvider
}

func TestNewAccountSetupErrorInvalidStartUrl(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup("test-account-id", "eu-west-1")

	require.Error(t, err)
}

func TestNewAccountSetupErrorInvalidRegion(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup("https://test-start-url.aws-apps.com/start", "mars")

	require.Error(t, err)
}

func TestNewAccountSetup(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    10,
		ExpiresAt:    200,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.AnythingOfType("string")).Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               5,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	setupResult, err := controller.Setup(startUrl, region)
	require.NoError(t, err)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "test-token-type",
		ExpiresIn:    300,
	}
	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockTokenRes, nil)

	mockTimeProvider.On("NowUnix").Return(1)

	err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.UserCode, setupResult.DeviceCode)

	require.NoError(t, err)
	require.Equal(t, setupResult, &AuthorizeDeviceFlowResult{
		ClientId:        "test-client-id",
		StartUrl:        "https://test-start-url.aws-apps.com/start",
		Region:          "eu-west-1",
		UserCode:        "test-user-code",
		DeviceCode:      "test-device-code",
		ExpiresIn:       5,
		VerificationUri: "https://test-verification-url",
	})
}

func TestNewAccountSetupErrorLoginTimeout(t *testing.T) {
	controller, mockAws, _ := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    10,
		ExpiresAt:    200,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.AnythingOfType("string")).Return(&mockRegRes, nil)

	expiresIn := 5

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "test-user-code",
		VerificationUri: "https://test-verification-url",
		ExpiresIn:       int32(expiresIn),
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	setupResult, err := controller.Setup(startUrl, region)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrDeviceCodeExpired)

	err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.UserCode, setupResult.DeviceCode)

	require.Error(t, err, ErrDeviceAuthFlowTimedOut)
}

func TestNewAccountSetupErrorUserDidNotAuthorizeDevice(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    10,
		ExpiresAt:    200,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.AnythingOfType("string")).Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "test-user-code",
		VerificationUri: "https://test-verification-url",
		ExpiresIn:       5,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	setupResult, err := controller.Setup(startUrl, region)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrDeviceFlowNotAuthorized)

	mockTimeProvider.On("NowUnix").Return(1)

	err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.UserCode, setupResult.DeviceCode)

	require.Error(t, err, ErrDeviceAuthFlowNotAuthorized)
}

func TestNewAccountSetupErrorDeviceAuthTimeout(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    10,
		ExpiresAt:    200,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.AnythingOfType("string")).Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "test-user-code",
		VerificationUri: "https://test-verification-url",
		ExpiresIn:       5,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	setupResult, err := controller.Setup(startUrl, region)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrDeviceCodeExpired)

	mockTimeProvider.On("NowUnix").Return(1)

	err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.UserCode, setupResult.DeviceCode)

	require.Error(t, err, ErrDeviceAuthFlowTimedOut)
}

func TestGetInstanceData(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	controller, mockAws, mockTimeProvider := simulateSuccessfulSetup(t, startUrl, region)

	mockTimeProvider.On("NowUnix").Return(3)

	mockListAccountsRes := awssso.ListAccountsResponse{
		Accounts: []awssso.AwsAccount{
			{
				AccountId:    "test-account-id",
				AccountName:  "test-account-name",
				AccountEmail: "test-account-email",
			},
			{
				AccountId:    "test-account-id-2",
				AccountName:  "test-account-name-2",
				AccountEmail: "test-account-email-2",
			},
		},
	}
	mockAws.On("ListAccounts", mock.Anything, mock.AnythingOfType("string")).Return(&mockListAccountsRes, nil)

	instanceData, err := controller.GetInstanceData(startUrl)
	require.NoError(t, err)

	require.Equal(t, "test-account-id", instanceData.Accounts[0].AccountId)
	require.Equal(t, "test-account-name", instanceData.Accounts[0].AccountName)

	require.Equal(t, "test-account-id-2", instanceData.Accounts[1].AccountId)
	require.Equal(t, "test-account-name-2", instanceData.Accounts[1].AccountName)
}

func TestGetInstanceTokenExpired(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	controller, mockAws, mockTimeProvider := simulateSuccessfulSetup(t, startUrl, region)

	mockTimeProvider.On("NowUnix").Return(10)

	mockAws.On("ListAccounts", mock.Anything, mock.AnythingOfType("string")).Return(nil, ErrAccessTokenExpired)

	_, err := controller.GetInstanceData(startUrl)

	require.Error(t, err)
}

func TestRefreshAccessToken(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	controller, mockAws, mockTimeProvider := simulateSuccessfulSetup(t, startUrl, region)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(startUrl)
	require.NoError(t, err)

	require.Equal(t, &AuthorizeDeviceFlowResult{
		ClientId:        "test-client-id",
		StartUrl:        "https://test-start-url.aws-apps.com/start",
		Region:          "eu-west-1",
		VerificationUri: "https://test-verification-url-2",
		UserCode:        "test-user-code-2",
		DeviceCode:      "test-device-code-2",
		ExpiresIn:       20,
	}, refreshRes)
}

func TestFinalizeRefreshAccessToken(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	controller, mockAws, mockTimeProvider := simulateSuccessfulSetup(t, startUrl, region)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(startUrl)
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Return(13)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token-2",
		AccessToken:  "test-access-token-2",
		RefreshToken: "test-refresh-token-2",
		TokenType:    "test-token-type-2",
		ExpiresIn:    15,
	}
	mockAws.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mockTokenRes, nil)

	err = controller.FinalizeRefreshAccessToken(refreshRes.ClientId, refreshRes.StartUrl, refreshRes.Region, refreshRes.UserCode, refreshRes.DeviceCode)
	require.NoError(t, err)
}

func TestAwsClientExpires(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    10,
		ExpiresAt:    200,
	}
	regCall := mockAws.On("RegisterClient", mock.Anything, mock.Anything).Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               5,
	}
	deviceAuthCall := mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	_, err := controller.Setup(startUrl, region)
	require.NoError(t, err)

	regCall.Unset()
	deviceAuthCall.Unset()

	mockTimeProvider.On("NowUnix").Return(int(mockRegRes.ExpiresAt + 1))

	mockRegRes = awssso.RegistrationResponse{
		ClientId:     "test-client-id-2",
		ClientSecret: "test-client-secret-2",
		CreatedAt:    201,
		ExpiresAt:    400,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.Anything).Return(&mockRegRes, nil)

	mockAuthRes = awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               10,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&mockAuthRes, nil)

	_, err = controller.Setup(startUrl, region)
	require.NoError(t, err)
}
