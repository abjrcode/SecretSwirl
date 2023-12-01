package awsiamidc

import (
	"testing"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/app"
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

func (m *mockAwsSsoOidcClient) RegisterClient(ctx app.Context, region awssso.AwsRegion, friendlyClientName string) (*awssso.RegistrationResponse, error) {
	args := m.Called()
	res, _ := args.Get(0).(*awssso.RegistrationResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) StartDeviceAuthorization(ctx app.Context, region awssso.AwsRegion, startUrl string, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	args := m.Called()
	res, _ := args.Get(0).(*awssso.AuthorizationResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) CreateToken(ctx app.Context, region awssso.AwsRegion, clientId, clientSecret, userCode, deviceCode string) (*awssso.GetTokenResponse, error) {
	args := m.Called()
	res, _ := args.Get(0).(*awssso.GetTokenResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) ListAccounts(ctx app.Context, region awssso.AwsRegion, accessToken string) (*awssso.ListAccountsResponse, error) {
	args := m.Called()
	res, _ := args.Get(0).(*awssso.ListAccountsResponse)
	return res, args.Error(1)
}

func (m *mockAwsSsoOidcClient) GetRoleCredentials(ctx app.Context, region awssso.AwsRegion, accountId, roleName, accessToken string) (*awssso.GetRoleCredentialsResponse, error) {
	args := m.Called()
	res, _ := args.Get(0).(*awssso.GetRoleCredentialsResponse)
	return res, args.Error(1)
}

func initController(t *testing.T) (*AwsIdentityCenterController, *mockAwsSsoOidcClient, *testhelpers.MockClock) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws-iam-idc-controller-tests.db")
	require.NoError(t, err)

	awsClient := new(mockAwsSsoOidcClient)
	mockDatetime := testhelpers.NewMockClock()
	logger := zerolog.Nop()
	favoritesRepo := favorites.NewFavorites(db, logger)

	vault := vault.NewVault(db, mockDatetime, logger)
	timeSetCall := mockDatetime.On("NowUnix").Return(1)
	err = vault.Configure(testhelpers.NewMockAppContext(), "abc")
	require.NoError(t, err)
	controller := NewAwsIdentityCenterController(db, favoritesRepo, vault, awsClient, mockDatetime, logger)

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
	regCall := mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               10,
	}
	deviceAuthCall := mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	ctx := testhelpers.NewMockAppContext()

	mockTimeProvider.On("NowUnix").Once().Return(1)
	setupResult, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.NoError(t, err)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "test-token-type",
		ExpiresIn:    5,
	}
	createTokenCall := mockAws.On("CreateToken").Return(&mockTokenRes, nil)

	tokenCreatedAt := 2
	timeSetCall := mockTimeProvider.On("NowUnix").Return(tokenCreatedAt)

	instanceId, err := controller.FinalizeSetup(ctx, AwsIamIdc_FinalizeSetupCommandInput{
		ClientId:   setupResult.ClientId,
		StartUrl:   setupResult.StartUrl,
		AwsRegion:  setupResult.Region,
		Label:      setupResult.Label,
		UserCode:   setupResult.UserCode,
		DeviceCode: setupResult.DeviceCode,
	})
	require.NoError(t, err)

	regCall.Unset()
	deviceAuthCall.Unset()
	createTokenCall.Unset()
	timeSetCall.Unset()

	return instanceId
}

func TestNewAccountSetupErrorInvalidStartUrl(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup(testhelpers.NewMockAppContext(), AwsIamIdc_SetupCommandInput{
		StartUrl:  "test-account-id",
		AwsRegion: "eu-west-1",
		Label:     "test-label"})

	require.Error(t, err, ErrInvalidStartUrl)
}

func TestNewAccountSetupErrorInvalidRegion(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup(testhelpers.NewMockAppContext(), AwsIamIdc_SetupCommandInput{
		StartUrl:  "https://test-start-url.aws-apps.com/start",
		AwsRegion: "region_mars",
		Label:     "test-label"})

	require.Error(t, err, ErrInvalidAwsRegion)
}

func TestNewAccountSetup_Error_InvalidLabel(t *testing.T) {
	controller, _, _ := initController(t)

	_, err := controller.Setup(testhelpers.NewMockAppContext(), AwsIamIdc_SetupCommandInput{
		StartUrl:  "https://test-start-url.aws-apps.com/start",
		AwsRegion: "eu-west-1",
		Label:     "i_am_a_very_long_label_that_is_longer_than_50_characters_and_therefore_invalid"})

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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAws.On("StartDeviceAuthorization").Return(nil, awssso.ErrInvalidRequest)

	_, err := controller.Setup(testhelpers.NewMockAppContext(), AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               5,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	ctx := testhelpers.NewMockAppContext()

	setupResult, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.NoError(t, err)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "test-token-type",
		ExpiresIn:    300,
	}
	mockAws.On("CreateToken").Return(&mockTokenRes, nil)

	mockTimeProvider.On("NowUnix").Return(1)

	instanceId, err := controller.FinalizeSetup(ctx, AwsIamIdc_FinalizeSetupCommandInput{
		ClientId:   setupResult.ClientId,
		StartUrl:   setupResult.StartUrl,
		AwsRegion:  setupResult.Region,
		Label:      setupResult.Label,
		UserCode:   setupResult.UserCode,
		DeviceCode: setupResult.DeviceCode,
	})

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

	ctx := testhelpers.NewMockAppContext()

	controller, mockAws, mockTimeProvider := initController(t)

	_ = simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	_, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.Error(t, err, ErrInstanceAlreadyRegistered)
}

func TestFinalizeSetup_Error_InvalidStartUrl(t *testing.T) {
	controller, _, _ := initController(t)

	startUrl := "mama_mia_bla.com"
	region := "eu-west-1"
	label := "test_label"

	ctx := testhelpers.NewMockAppContext()

	_, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.Error(t, err, ErrInvalidStartUrl)
}

func TestFinalizeSetup_Error_InvalidRegion(t *testing.T) {
	controller, _, _ := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "region_mars"
	label := "test_label"

	ctx := testhelpers.NewMockAppContext()

	_, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.Error(t, err, ErrInvalidAwsRegion)
}

func TestFinalizeSetup_Error_InvalidLabel(t *testing.T) {
	controller, _, _ := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "i_am_a_very_long_label_that_is_longer_than_50_characters_and_therefore_invalid"

	ctx := testhelpers.NewMockAppContext()

	_, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	expiresIn := 5

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "test-user-code",
		VerificationUri: "https://test-verification-url",
		ExpiresIn:       int32(expiresIn),
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"

	ctx := testhelpers.NewMockAppContext()

	setupResult, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     "test_label",
	})
	require.NoError(t, err)

	mockAws.On("CreateToken").Return(nil, awssso.ErrDeviceCodeExpired)

	_, err = controller.FinalizeSetup(ctx, AwsIamIdc_FinalizeSetupCommandInput{
		ClientId:   setupResult.ClientId,
		StartUrl:   setupResult.StartUrl,
		AwsRegion:  setupResult.Region,
		Label:      setupResult.Label,
		UserCode:   setupResult.UserCode,
		DeviceCode: setupResult.DeviceCode,
	})

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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "test-user-code",
		VerificationUri: "https://test-verification-url",
		ExpiresIn:       5,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	ctx := testhelpers.NewMockAppContext()

	setupResult, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.NoError(t, err)

	mockAws.On("CreateToken").Return(nil, awssso.ErrDeviceFlowNotAuthorized)

	mockTimeProvider.On("NowUnix").Return(1)

	_, err = controller.FinalizeSetup(ctx, AwsIamIdc_FinalizeSetupCommandInput{
		ClientId:   setupResult.ClientId,
		StartUrl:   setupResult.StartUrl,
		AwsRegion:  setupResult.Region,
		Label:      setupResult.Label,
		UserCode:   setupResult.UserCode,
		DeviceCode: setupResult.DeviceCode,
	})

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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:      "test-device-code",
		UserCode:        "test-user-code",
		VerificationUri: "https://test-verification-url",
		ExpiresIn:       5,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	ctx := testhelpers.NewMockAppContext()

	setupResult, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.NoError(t, err)

	mockAws.On("CreateToken").Return(nil, awssso.ErrDeviceCodeExpired)

	mockTimeProvider.On("NowUnix").Return(1)

	_, err = controller.FinalizeSetup(ctx, AwsIamIdc_FinalizeSetupCommandInput{
		ClientId:   setupResult.ClientId,
		StartUrl:   setupResult.StartUrl,
		AwsRegion:  setupResult.Region,
		Label:      setupResult.Label,
		UserCode:   setupResult.UserCode,
		DeviceCode: setupResult.DeviceCode,
	})

	require.Error(t, err, ErrDeviceAuthFlowTimedOut)
}

func TestListInstances(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	ctx := testhelpers.NewMockAppContext()

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
	mockAws.On("StartDeviceAuthorization").Once().Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Once().Return(2)
	setupResult, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl2,
		AwsRegion: region2,
		Label:     label2,
	})
	require.NoError(t, err)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token",
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "test-token-type",
		ExpiresIn:    5,
	}
	mockAws.On("CreateToken").Once().Return(&mockTokenRes, nil)

	tokenCreatedAt := 3
	mockTimeProvider.On("NowUnix").Once().Return(tokenCreatedAt)

	instanceId2, err := controller.FinalizeSetup(ctx, AwsIamIdc_FinalizeSetupCommandInput{
		ClientId:   setupResult.ClientId,
		StartUrl:   setupResult.StartUrl,
		AwsRegion:  setupResult.Region,
		Label:      setupResult.Label,
		UserCode:   setupResult.UserCode,
		DeviceCode: setupResult.DeviceCode,
	})
	require.NoError(t, err)

	instances, err := controller.ListInstances(ctx)

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
				Roles: []awssso.AwsAccountRole{
					{
						RoleName: "test-role-name",
					},
				},
			},
			{
				AccountId:    "test-account-id-2",
				AccountName:  "test-account-name-2",
				AccountEmail: "test-account-email-2",
			},
		},
	}
	mockAws.On("ListAccounts").Return(&mockListAccountsRes, nil)

	ctx := testhelpers.NewMockAppContext()

	instanceData, err := controller.GetInstanceData(ctx, instanceId, false)
	require.NoError(t, err)

	require.Equal(t, instanceId, instanceData.InstanceId)
	require.Equal(t, label, instanceData.Label)
	require.Equal(t, false, instanceData.IsFavorite)
	require.Equal(t, false, instanceData.IsAccessTokenExpired)
	require.Equal(t, "test-account-id", instanceData.Accounts[0].AccountId)
	require.Equal(t, "test-account-name", instanceData.Accounts[0].AccountName)
	require.Equal(t, "test-role-name", instanceData.Accounts[0].Roles[0].RoleName)

	require.Equal(t, "test-account-id-2", instanceData.Accounts[1].AccountId)
	require.Equal(t, "test-account-name-2", instanceData.Accounts[1].AccountName)
}

func TestGetRoleCredentials(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	roleName := "test-role-name"
	accountId := "test-account-id"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockTimeProvider.On("NowUnix").Return(3)

	mockListAccountsRes := awssso.ListAccountsResponse{
		Accounts: []awssso.AwsAccount{
			{
				AccountId:    accountId,
				AccountName:  "test-account-name",
				AccountEmail: "test-account-email",
				Roles: []awssso.AwsAccountRole{
					{
						RoleName: roleName,
					},
				},
			},
		},
	}

	mockAws.On("ListAccounts").Return(&mockListAccountsRes, nil)

	mockGetRoleCredentialsRes := awssso.GetRoleCredentialsResponse{
		AccessKeyId:     "test-access-key-id",
		SecretAccessKey: "test-secret-key",
		SessionToken:    "test-session-token",
		Expiration:      100,
	}

	mockAws.On("GetRoleCredentials").Return(&mockGetRoleCredentialsRes, nil)

	ctx := testhelpers.NewMockAppContext()

	roleCredentials, err := controller.GetRoleCredentials(ctx, AwsIamIdc_GetRoleCredentialsCommandInput{
		InstanceId: instanceId,
		AccountId:  accountId,
		RoleName:   roleName,
	})

	require.NoError(t, err)
	require.Equal(t, roleCredentials, &AwsIdentityCenterAccountRoleCredentials{
		AccessKeyId:     mockGetRoleCredentialsRes.AccessKeyId,
		SecretAccessKey: mockGetRoleCredentialsRes.SecretAccessKey,
		SessionToken:    mockGetRoleCredentialsRes.SessionToken,
		Expiration:      mockGetRoleCredentialsRes.Expiration,
	})
}

func TestGetInstance_AccessTokenExpired(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockTimeProvider.On("NowUnix").Return(10)

	ctx := testhelpers.NewMockAppContext()

	data, err := controller.GetInstanceData(ctx, instanceId, false)

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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	ctx := testhelpers.NewMockAppContext()
	_, err := controller.GetInstanceData(ctx, "well-if-u-can-find-me-it-sucks", false)
	require.Error(t, err, ErrInstanceWasNotFound)
}

func TestMarkInstanceAsFavorite(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	ctx := testhelpers.NewMockAppContext()
	err := controller.MarkAsFavorite(ctx, instanceId)
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
	mockAws.On("ListAccounts").Return(&mockListAccountsRes, nil)

	instanceData, err := controller.GetInstanceData(ctx, instanceId, false)
	require.NoError(t, err)

	require.Equal(t, true, instanceData.IsFavorite)
}

func TestUnmarkInstanceAsFavorite(t *testing.T) {
	controller, mockAws, mockTimeProvider := initController(t)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	ctx := testhelpers.NewMockAppContext()

	err := controller.MarkAsFavorite(ctx, instanceId)
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
	mockAws.On("ListAccounts").Return(&mockListAccountsRes, nil)

	instanceData, err := controller.GetInstanceData(ctx, instanceId, false)
	require.NoError(t, err)
	require.Equal(t, true, instanceData.IsFavorite)

	err = controller.UnmarkAsFavorite(ctx, instanceId)
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Once().Return(5)

	instanceData, err = controller.GetInstanceData(ctx, instanceId, false)
	require.NoError(t, err)

	require.Equal(t, false, instanceData.IsFavorite)
}

func TestRefreshAccessToken(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	ctx := testhelpers.NewMockAppContext()

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(ctx, instanceId)
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

	ctx := testhelpers.NewMockAppContext()

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(ctx, instanceId)
	require.NoError(t, err)

	mockTimeProvider.On("NowUnix").Return(13)

	mockTokenRes := awssso.GetTokenResponse{
		IdToken:      "test-id-token-2",
		AccessToken:  "test-access-token-2",
		RefreshToken: "test-refresh-token-2",
		TokenType:    "test-token-type-2",
		ExpiresIn:    15,
	}
	mockAws.On("CreateToken").Return(&mockTokenRes, nil)

	err = controller.FinalizeRefreshAccessToken(ctx, AwsIamIdc_FinalizeRefreshAccessTokenCommandInput{
		InstanceId: instanceId,
		Region:     region,
		UserCode:   refreshRes.UserCode,
		DeviceCode: refreshRes.DeviceCode,
	})
	require.NoError(t, err)
}

func TestRefresh_NonExistentInstance(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	_ = simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	ctx := testhelpers.NewMockAppContext()

	_, err := controller.RefreshAccessToken(ctx, "well-if-u-can-find-me-it-sucks")
	require.Error(t, err, ErrInstanceWasNotFound)
}

func TestFinalizeRefreshAccessToken_InstanceDoesNotExist(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	ctx := testhelpers.NewMockAppContext()

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(ctx, instanceId)
	require.NoError(t, err)

	mockAws.On("CreateToken").Return(nil, awssso.ErrDeviceCodeExpired)

	incorrectInstanceId := "well-if-u-can-find-me-it-sucks"
	err = controller.FinalizeRefreshAccessToken(ctx, AwsIamIdc_FinalizeRefreshAccessTokenCommandInput{
		InstanceId: incorrectInstanceId,
		Region:     region,
		UserCode:   refreshRes.UserCode,
		DeviceCode: refreshRes.DeviceCode,
	})
	require.Error(t, err, ErrInstanceWasNotFound)
}

func TestFinalizeRefreshAccessToken_DeviceNotAuthorizedByUser(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	ctx := testhelpers.NewMockAppContext()

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(ctx, instanceId)
	require.NoError(t, err)

	mockAws.On("CreateToken").Return(nil, awssso.ErrDeviceFlowNotAuthorized)

	err = controller.FinalizeRefreshAccessToken(ctx, AwsIamIdc_FinalizeRefreshAccessTokenCommandInput{
		InstanceId: instanceId,
		Region:     refreshRes.Region,
		UserCode:   refreshRes.UserCode,
		DeviceCode: refreshRes.DeviceCode,
	})
	require.Error(t, err, ErrDeviceAuthFlowNotAuthorized)
}

func TestFinalizeRefreshAccessToken_DeviceAuthTimeout(t *testing.T) {
	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	controller, mockAws, mockTimeProvider := initController(t)

	ctx := testhelpers.NewMockAppContext()

	instanceId := simulateSuccessfulSetup(t, controller, mockAws, mockTimeProvider, startUrl, region, label)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               20,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	mockTimeProvider.On("NowUnix").Return(10)
	refreshRes, err := controller.RefreshAccessToken(ctx, instanceId)
	require.NoError(t, err)

	mockAws.On("CreateToken").Return(nil, awssso.ErrDeviceCodeExpired)

	err = controller.FinalizeRefreshAccessToken(ctx, AwsIamIdc_FinalizeRefreshAccessTokenCommandInput{
		InstanceId: instanceId,
		Region:     region,
		UserCode:   refreshRes.UserCode,
		DeviceCode: refreshRes.DeviceCode,
	})
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
	regCall := mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAuthRes := awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code",
		UserCode:                "test-user-code",
		VerificationUriComplete: "https://test-verification-url",
		ExpiresIn:               5,
	}
	deviceAuthCall := mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	startUrl := "https://test-start-url.aws-apps.com/start"
	region := "eu-west-1"
	label := "test_label"

	ctx := testhelpers.NewMockAppContext()

	_, err := controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
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
	mockAws.On("RegisterClient").Return(&mockRegRes, nil)

	mockAuthRes = awssso.AuthorizationResponse{
		DeviceCode:              "test-device-code-2",
		UserCode:                "test-user-code-2",
		VerificationUriComplete: "https://test-verification-url-2",
		ExpiresIn:               10,
	}
	mockAws.On("StartDeviceAuthorization").Return(&mockAuthRes, nil)

	_, err = controller.Setup(ctx, AwsIamIdc_SetupCommandInput{
		StartUrl:  startUrl,
		AwsRegion: region,
		Label:     label,
	})
	require.NoError(t, err)
}
