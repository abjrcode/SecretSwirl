package awsiamidc

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/favorites"
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

func (m *mockAwsSsoOidcClient) StartDeviceAuthorization(ctx context.Context, startUrl string, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
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

func initController(t *testing.T) (*AwsIdentityCenterController, *mockAwsSsoOidcClient, *testhelpers.MockClock) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws-iam-idc-controller-tests.db")
	require.NoError(t, err)

	awsClient := new(mockAwsSsoOidcClient)
	mockDatetime := testhelpers.NewMockClock()
	logger := zerolog.Nop()
	favoritesRepo := favorites.NewFavorites(db, &logger)
	errHandler := testhelpers.NewMockErrorHandler(t)
	ctx := logger.WithContext(context.Background())
	vault := vault.NewVault(db, mockDatetime, &logger, errHandler)
	timeSetCall := mockDatetime.On("NowUnix").Return(1)
	err = vault.Configure(context.Background(), "abc")
	require.NoError(t, err)
	controller := NewAwsIdentityCenterController(db, favoritesRepo, vault, awsClient, mockDatetime)
	controller.Init(ctx, errHandler)

	timeSetCall.Unset()

	return controller, awsClient, mockDatetime
}

func simulateSuccessfulSetup(t *testing.T, controller *AwsIdentityCenterController, mockAws *mockAwsSsoOidcClient, mockTimeProvider *testhelpers.MockClock, startUrl, region, label string) string {
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

	mockTimeProvider.On("NowUnix").Once().Return(1)
	setupResult, err := controller.Setup(startUrl, region, label)
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

	instanceId, err := controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.Label, setupResult.UserCode, setupResult.DeviceCode)
	require.NoError(t, err)

	regCall.Unset()
	deviceAuthCall.Unset()
	createTokenCall.Unset()
	timeSetCall.Unset()

	return instanceId
}

func TestNewAccountSetupErrorInvalidStartUrl(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup("test-account-id", "eu-west-1", "test-label")

	require.Error(t, err, ErrInvalidStartUrl)
}

func TestNewAccountSetupErrorInvalidRegion(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup("https://test-start-url.aws-apps.com/start", "region_mars", "test_label")

	require.Error(t, err, ErrInvalidAwsRegion)
}

func TestNewAccountSetup_Error_InvalidLabel(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup("https://test-start-url.aws-apps.com/start", "region_mars", "i_am_a_very_long_label_that_is_longer_than_50_characters_and_therefore_invalid")

	require.Error(t, err, ErrInvalidLabel)
}

func TestNewAccountSetup_Error_AwsInvalidRequest(t *testing.T) {
	startUrl := "https://wth/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, _ := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    10,
		ExpiresAt:    200,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.AnythingOfType("string")).Return(&mockRegRes, nil)

	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrInvalidRequest)

	_, err := controller.Setup(startUrl, region, label)
	require.Error(t, err, ErrInvalidStartUrl)
}

func TestNewAccount_FullSetup_Success(t *testing.T) {
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
	label := "test_label"

	setupResult, err := controller.Setup(startUrl, region, label)
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

	instanceId, err := controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.Label, setupResult.UserCode, setupResult.DeviceCode)

	require.NoError(t, err)
	require.NotEmpty(t, instanceId)
	require.Equal(t, setupResult, &AuthorizeDeviceFlowResult{
		StartUrl:        "https://test-start-url.aws-apps.com/start",
		Region:          "eu-west-1",
		Label:           label,
		ClientId:        "test-client-id",
		UserCode:        "test-user-code",
		DeviceCode:      "test-device-code",
		ExpiresIn:       5,
		VerificationUri: "https://test-verification-url",
	})
}

func TestNewAccountSetupErrorDoubleRegistration(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	_ = simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	_, err := controller.Setup(startUrl, region, label)
	require.Error(t, err, ErrInstanceAlreadyRegistered)
}

func TestFinalizeSetup_Error_InvalidStartUrl(t *testing.T) {
	controller, _, _ := initController(t)

	startUrl := "mama_mia_bla.com"
	region := "eu-west-1"
	label := "test_label"

	_, err := controller.Setup(startUrl, region, label)
	require.Error(t, err, ErrInvalidStartUrl)
}

func TestFinalizeSetup_Error_InvalidRegion(t *testing.T) {
	controller, _, _ := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "region_mars"
	label := "test_label"

	_, err := controller.Setup(startUrl, region, label)
	require.Error(t, err, ErrInvalidAwsRegion)
}

func TestFinalizeSetup_Error_InvalidLabel(t *testing.T) {
	controller, _, _ := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "i_am_a_very_long_label_that_is_longer_than_50_characters_and_therefore_invalid"

	_, err := controller.Setup(startUrl, region, label)
	require.Error(t, err, ErrInvalidLabel)
}

func TestFinalizeSetup_ErrorLoginTimeout(t *testing.T) {
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

	setupResult, err := controller.Setup(startUrl, region, "test_label")
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrDeviceCodeExpired)

	_, err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.Label, setupResult.UserCode, setupResult.DeviceCode)

	require.Error(t, err, ErrDeviceAuthFlowTimedOut)
}

func TestFinalizeSetup_Error_UserDidNotAuthorizeDevice(t *testing.T) {
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
	label := "test_label"

	setupResult, err := controller.Setup(startUrl, region, label)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrDeviceFlowNotAuthorized)

	mockTimeProvider.On("NowUnix").Return(1)

	_, err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.Label, setupResult.UserCode, setupResult.DeviceCode)

	require.Error(t, err, ErrDeviceAuthFlowNotAuthorized)
}

func TestFinalizeSetup_Error_DeviceAuthTimeout(t *testing.T) {
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
	label := "test_label"

	setupResult, err := controller.Setup(startUrl, region, label)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, awssso.ErrDeviceCodeExpired)

	mockTimeProvider.On("NowUnix").Return(1)

	_, err = controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.Label, setupResult.UserCode, setupResult.DeviceCode)

	require.Error(t, err, ErrDeviceAuthFlowTimedOut)
}

func TestListInstances(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	startUrl2 := "https://test-start-url-2.aws-apps.com/start"
	region2 := "eu-west-2"
	label2 := "test_label_2"

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               10,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Once().Return(2)
	setupResult, err := controller.Setup(startUrl2, region2, label2)
	require.NoError(t, err)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "test-token-type",
		ExpiresIn:    5,
	}
	mockAws.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(&mockTokenRes, nil)

	tokenCreatedAt := 3
	mockTimeProvider.On("NowUnix").Once().Return(tokenCreatedAt)

	instanceId2, err := controller.FinalizeSetup(setupResult.ClientId, setupResult.StartUrl, setupResult.Region, setupResult.Label, setupResult.UserCode, setupResult.DeviceCode)
	require.NoError(t, err)

	instances, err := controller.ListInstances()

	require.NoError(t, err)

	require.Equal(t, instances, []string{instanceId2, instanceId})
}

func TestGetInstanceData(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

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

	instanceData, err := controller.GetInstanceData(instanceId)
	require.NoError(t, err)

	require.Equal(t, instanceId, instanceData.InstanceId)
	require.Equal(t, label, instanceData.Label)
	require.Equal(t, false, instanceData.IsFavorite)
	require.Equal(t, false, instanceData.IsAccessTokenExpired)
	require.Equal(t, "test-account-id", instanceData.Accounts[0].AccountId)
	require.Equal(t, "test-account-name", instanceData.Accounts[0].AccountName)

	require.Equal(t, "test-account-id-2", instanceData.Accounts[1].AccountId)
	require.Equal(t, "test-account-name-2", instanceData.Accounts[1].AccountName)
}

func TestGetInstance_AccessTokenExpired(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockTimeProvider.On("NowUnix").Return(10)

	// mockAws.On("ListAccounts", mock.Anything, mock.AnythingOfType("string")).Return(nil, awssso.ErrAccessTokenExpired)

	data, err := controller.GetInstanceData(instanceId)

	require.NoError(t, err)

	require.Equal(t, instanceId, data.InstanceId)
	require.Equal(t, label, data.Label)
	require.Equal(t, true, data.IsAccessTokenExpired)
	require.Empty(t, data.Accounts)
}

func TestGetNonExistentInstance(t *testing.T) {
	controller, mockAws, _ := initController(t)

	mockRegRes := awssso.RegistrationResponse{
		ClientId:     "test-client-id",
		ClientSecret: "test-client-secret",
		CreatedAt:    1,
		ExpiresAt:    20,
	}
	mockAws.On("RegisterClient", mock.Anything, mock.Anything).Return(&mockRegRes, nil)

	_, err := controller.GetInstanceData("well-if-u-can-find-me-it-sucks")
	require.Error(t, err, ErrInstanceWasNotFound)
}

func TestMarkInstanceAsFavorite(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	err := controller.MarkAsFavorite(instanceId)
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Once().Return(4)
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

	instanceData, err := controller.GetInstanceData(instanceId)
	require.NoError(t, err)

	require.Equal(t, true, instanceData.IsFavorite)
}

func TestUnmarkInstanceAsFavorite(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	err := controller.MarkAsFavorite(instanceId)
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Once().Return(4)
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

	instanceData, err := controller.GetInstanceData(instanceId)
	require.NoError(t, err)
	require.Equal(t, true, instanceData.IsFavorite)

	err = controller.UnmarkAsFavorite(instanceId)
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Once().Return(5)

	instanceData, err = controller.GetInstanceData(instanceId)
	require.NoError(t, err)

	require.Equal(t, false, instanceData.IsFavorite)
}

func TestRefreshAccessToken(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(instanceId)
	require.NoError(t, err)

	require.Equal(t, &AuthorizeDeviceFlowResult{
		InstanceId:      instanceId,
		ClientId:        "test-client-id",
		StartUrl:        "https://test-start-url.aws-apps.com/start",
		Region:          "eu-west-1",
		Label:           label,
		VerificationUri: "https://test-verification-url-2",
		UserCode:        "test-user-code-2",
		DeviceCode:      "test-device-code-2",
		ExpiresIn:       20,
	}, refreshRes)
}

func TestFinalizeRefreshAccessToken(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(instanceId)
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

	err = controller.FinalizeRefreshAccessToken(instanceId, refreshRes.Region, refreshRes.UserCode, refreshRes.DeviceCode)
	require.NoError(t, err)
}

func TestRefresh_NonExistentInstance(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	_ = simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	_, err := controller.RefreshAccessToken("well-if-u-can-find-me-it-sucks")
	require.Error(t, err, ErrInstanceWasNotFound)
}

func TestFinalizeRefreshAccessToken_InstanceDoesNotExist(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(instanceId)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, awssso.ErrDeviceCodeExpired)

	incorrectInstanceId := "well-if-u-can-find-me-it-sucks"
	err = controller.FinalizeRefreshAccessToken(incorrectInstanceId, region, refreshRes.UserCode, refreshRes.DeviceCode)
	require.Error(t, err, ErrInstanceWasNotFound)
}

func TestFinalizeRefreshAccessToken_DeviceNotAuthorizedByUser(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(instanceId)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, awssso.ErrDeviceFlowNotAuthorized)

	err = controller.FinalizeRefreshAccessToken(instanceId, refreshRes.Region, refreshRes.UserCode, refreshRes.DeviceCode)
	require.Error(t, err, ErrDeviceAuthFlowNotAuthorized)
}

func TestFinalizeRefreshAccessToken_DeviceAuthTimeout(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization", mock.Anything, mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(instanceId)
	require.NoError(t, err)

	mockAws.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, awssso.ErrDeviceCodeExpired)

	err = controller.FinalizeRefreshAccessToken(instanceId, region, refreshRes.UserCode, refreshRes.DeviceCode)
	require.Error(t, err, ErrDeviceAuthFlowTimedOut)
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
	label := "test_label"

	_, err := controller.Setup(startUrl, region, label)
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

	_, err = controller.Setup(startUrl, region, label)
	require.NoError(t, err)
}
