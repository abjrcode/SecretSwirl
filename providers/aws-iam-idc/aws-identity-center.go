package awsiamidc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/logging"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/abjrcode/swervo/providers"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"
	"github.com/segmentio/ksuid"
)

var (
	ErrInvalidStartUrl             = errors.New("INVALID_START_URL")
	ErrInvalidAwsRegion            = errors.New("INVALID_AWS_REGION")
	ErrInvalidLabel                = errors.New("INVALID_LABEL")
	ErrDeviceAuthFlowNotAuthorized = errors.New("DEVICE_AUTH_FLOW_NOT_AUTHORIZED")
	ErrDeviceAuthFlowTimedOut      = errors.New("DEVICE_AUTH_FLOW_TIMED_OUT")
	ErrInstanceWasNotFound         = errors.New("INSTANCE_WAS_NOT_FOUND")
	ErrInstanceAlreadyRegistered   = errors.New("INSTANCE_ALREADY_REGISTERED")
	ErrTransientAwsClientError     = errors.New("TRANSIENT_AWS_CLIENT_ERROR")
)

type AwsIdentityCenterController struct {
	ctx               context.Context
	logger            *zerolog.Logger
	errHandler        logging.ErrorHandler
	db                *sql.DB
	favoritesRepo     favorites.FavoritesRepo
	encryptionService encryption.EncryptionService
	awsSsoClient      awssso.AwsSsoOidcClient
	timeHelper        utils.Clock
}

func NewAwsIdentityCenterController(db *sql.DB, favoritesRepo favorites.FavoritesRepo, encryptionService encryption.EncryptionService, awsSsoClient awssso.AwsSsoOidcClient, datetime utils.Clock) *AwsIdentityCenterController {
	return &AwsIdentityCenterController{
		db:                db,
		favoritesRepo:     favoritesRepo,
		encryptionService: encryptionService,
		awsSsoClient:      awsSsoClient,
		timeHelper:        datetime,
	}
}

func (controller *AwsIdentityCenterController) Init(ctx context.Context, errorHandler logging.ErrorHandler) {
	controller.ctx = ctx
	enrichedLogger := zerolog.Ctx(ctx).With().Str("component", "aws_idc_controller").Logger()
	controller.logger = &enrichedLogger
	controller.errHandler = errorHandler
}

type AwsIdentityCenterAccount struct {
	AccountId   string `json:"accountId"`
	AccountName string `json:"accountName"`
}

type AwsIdentityCenterCardData struct {
	InstanceId           string                     `json:"instanceId"`
	Enabled              bool                       `json:"enabled"`
	Label                string                     `json:"label"`
	IsFavorite           bool                       `json:"isFavorite"`
	IsAccessTokenExpired bool                       `json:"isAccessTokenExpired"`
	AccessTokenExpiresIn string                     `json:"accessTokenExpiresIn"`
	Accounts             []AwsIdentityCenterAccount `json:"accounts"`
}

func (c *AwsIdentityCenterController) ListInstances() ([]string, error) {
	rows, err := c.db.QueryContext(c.ctx, "SELECT instance_id FROM aws_iam_idc_instances ORDER BY instance_id DESC")

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]string, 0), nil
		}

		c.errHandler.Catch(c.logger, err)
	}

	instances := make([]string, 0)

	for rows.Next() {
		var instanceId string

		if err := rows.Scan(&instanceId); err != nil {
			c.errHandler.Catch(c.logger, err)
		}

		instances = append(instances, instanceId)
	}

	return instances, nil
}

func (c *AwsIdentityCenterController) GetInstanceData(instanceId string) (*AwsIdentityCenterCardData, error) {
	row := c.db.QueryRowContext(c.ctx, "SELECT region, label, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_iam_idc_instances WHERE instance_id = ?", instanceId)

	var region string
	var label string
	var accessTokenEnc string
	var accessTokenCreatedAt int64
	var accessTokenExpiresIn int64
	var encKeyId string

	if err := row.Scan(&region, &label, &accessTokenEnc, &accessTokenCreatedAt, &accessTokenExpiresIn, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceWasNotFound
		}

		c.errHandler.Catch(c.logger, err)
	}

	isFavorite, err := c.favoritesRepo.IsFavorite(c.ctx, &favorites.Favorite{
		ProviderCode: providers.AwsIamIdc,
		InstanceId:   instanceId,
	})

	c.errHandler.CatchWithMsg(c.logger, err, "failed to check if instance is favorite")

	now := c.timeHelper.NowUnix()
	if now > accessTokenCreatedAt+accessTokenExpiresIn {
		c.logger.Info().Msgf("token for instance [%s] has expired", instanceId)

		return &AwsIdentityCenterCardData{
			Enabled:              true,
			InstanceId:           instanceId,
			Label:                label,
			IsFavorite:           isFavorite,
			IsAccessTokenExpired: true,
			AccessTokenExpiresIn: humanize.Time(time.Unix(accessTokenCreatedAt+accessTokenExpiresIn, 0)),
			Accounts:             make([]AwsIdentityCenterAccount, 0),
		}, nil
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to decrypt access token")

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	accountsOut, err := c.awsSsoClient.ListAccounts(ctx, accessToken)

	if err != nil {
		c.logger.Error().Err(err).Msg("aws sso client failed to list accounts")

		return nil, ErrTransientAwsClientError
	}

	accounts := make([]AwsIdentityCenterAccount, 0)

	for _, account := range accountsOut.Accounts {
		accounts = append(accounts, AwsIdentityCenterAccount{
			AccountId:   account.AccountId,
			AccountName: account.AccountName,
		})
	}

	return &AwsIdentityCenterCardData{
		Enabled:              true,
		InstanceId:           instanceId,
		Label:                label,
		IsFavorite:           isFavorite,
		IsAccessTokenExpired: false,
		AccessTokenExpiresIn: humanize.Time(time.Unix(accessTokenCreatedAt+accessTokenExpiresIn, 0)),
		Accounts:             accounts,
	}, nil
}

func (c *AwsIdentityCenterController) validateStartUrl(startUrl string) error {
	_, err := url.ParseRequestURI(startUrl)

	if err != nil {
		return ErrInvalidStartUrl
	}

	return nil
}

func (c *AwsIdentityCenterController) validateAwsRegion(region string) error {
	if _, ok := awssso.SupportedAwsRegions[region]; !ok {
		return ErrInvalidAwsRegion
	}

	return nil
}

func (c *AwsIdentityCenterController) validateLabel(label string) error {
	if len(label) < 1 || len(label) > 50 {
		return ErrInvalidLabel
	}

	return nil
}

type AuthorizeDeviceFlowResult struct {
	InstanceId      string `json:"instanceId"`
	StartUrl        string `json:"startUrl"`
	Region          string `json:"region"`
	Label           string `json:"label"`
	ClientId        string `json:"clientId"`
	VerificationUri string `json:"verificationUri"`
	UserCode        string `json:"userCode"`
	ExpiresIn       int32  `json:"expiresIn"`
	DeviceCode      string `json:"deviceCode"`
}

func (c *AwsIdentityCenterController) Setup(startUrl, awsRegion, label string) (*AuthorizeDeviceFlowResult, error) {
	if err := c.validateStartUrl(startUrl); err != nil {
		return nil, err
	}

	if err := c.validateAwsRegion(awsRegion); err != nil {
		return nil, err
	}

	if err := c.validateLabel(label); err != nil {
		return nil, err
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), awsRegion)

	var exists bool
	err := c.db.QueryRowContext(ctx, "SELECT 1 FROM aws_iam_idc_instances WHERE start_url = ?", startUrl).Scan(&exists)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.errHandler.CatchWithMsg(c.logger, err, "failed to check if instance exists")
	}

	if exists {
		c.logger.Warn().Msgf("instance [%s] already exists", startUrl)
		return nil, ErrInstanceAlreadyRegistered
	}

	if len(label) < 1 || len(label) > 50 {
		c.logger.Debug().Msgf("invalid label [%s]", label)
		return nil, ErrInvalidLabel
	}

	regRes, err := c.getOrRegisterClient(ctx)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get or register client")
		return nil, ErrTransientAwsClientError
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		if errors.Is(err, awssso.ErrInvalidRequest) {
			c.logger.Debug().Err(err).Msg("failed to authorize device because start URL is invalid")
			return nil, ErrInvalidStartUrl
		}

		c.logger.Error().Err(err).Msg("failed to authorize device")
		return nil, ErrTransientAwsClientError
	}

	return &AuthorizeDeviceFlowResult{
		StartUrl:        startUrl,
		Region:          awsRegion,
		Label:           label,
		ClientId:        regRes.ClientId,
		VerificationUri: authorizeRes.VerificationUriComplete,
		UserCode:        authorizeRes.UserCode,
		ExpiresIn:       authorizeRes.ExpiresIn,
		DeviceCode:      authorizeRes.DeviceCode,
	}, nil
}

func (c *AwsIdentityCenterController) FinalizeSetup(clientId, startUrl, region, label, userCode, deviceCode string) (string, error) {
	if err := c.validateLabel(label); err != nil {
		return "", err
	}

	if err := c.validateStartUrl(startUrl); err != nil {
		return "", err
	}

	if err := c.validateAwsRegion(region); err != nil {
		return "", err
	}

	row := c.db.QueryRowContext(c.ctx, "SELECT client_secret_enc, enc_key_id FROM aws_iam_idc_clients")

	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientSecretEnc, &encKeyId); err != nil {
		c.errHandler.Catch(c.logger, err)
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to decrypt client secret")

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	tokenRes, err := c.getToken(ctx, clientId, clientSecret, deviceCode, userCode)

	if err != nil {
		if errors.Is(err, awssso.ErrDeviceFlowNotAuthorized) {
			c.logger.Debug().Err(err).Msg("failed to get token because user did not authorize device")
			return "", ErrDeviceAuthFlowNotAuthorized
		}

		if errors.Is(err, awssso.ErrDeviceCodeExpired) {
			c.logger.Debug().Err(err).Msg("failed to get token because user and device code expired")
			return "", ErrDeviceAuthFlowTimedOut
		}

		c.logger.Error().Err(err).Msg("failed to get token")
		return "", ErrTransientAwsClientError
	}

	idTokenEnc, keyId, err := c.encryptionService.Encrypt(tokenRes.IdToken)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt id token")

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt access token")

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)
	c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt refresh token")

	nowUnix := c.timeHelper.NowUnix()

	uniqueId, err := ksuid.NewRandomWithTime(time.Unix(nowUnix, 0))
	c.errHandler.CatchWithMsg(c.logger, err, "failed to generate instance ID")
	instanceId := uniqueId.String()

	sql := `INSERT INTO aws_iam_idc_instances
	(instance_id, start_url, region, label, enabled, id_token_enc, access_token_enc, token_type, access_token_created_at,
		access_token_expires_in, refresh_token_enc, enc_key_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = c.db.ExecContext(ctx, sql,
		instanceId,
		startUrl,
		region,
		label,
		true,
		idTokenEnc,
		accessTokenEnc,
		tokenRes.TokenType,
		nowUnix,
		tokenRes.ExpiresIn,
		refreshTokenEnc,
		keyId)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to save token to database")

	return instanceId, nil
}

func (c *AwsIdentityCenterController) MarkAsFavorite(instanceId string) error {
	return c.favoritesRepo.Add(c.ctx, &favorites.Favorite{
		ProviderCode: providers.AwsIamIdc,
		InstanceId:   instanceId,
	})
}

func (c *AwsIdentityCenterController) UnmarkAsFavorite(instanceId string) error {
	return c.favoritesRepo.Remove(c.ctx, &favorites.Favorite{
		ProviderCode: providers.AwsIamIdc,
		InstanceId:   instanceId,
	})
}

func (c *AwsIdentityCenterController) RefreshAccessToken(instanceId string) (*AuthorizeDeviceFlowResult, error) {
	var startUrl string
	var awsRegion string
	var label string
	row := c.db.QueryRowContext(c.ctx, "SELECT start_url, region, label FROM aws_iam_idc_instances WHERE instance_id = ?", instanceId)

	if err := row.Scan(&startUrl, &awsRegion, &label); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Debug().Msgf("instance [%s] was not found", instanceId)
			return nil, ErrInstanceWasNotFound
		}
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), awsRegion)
	regRes, err := c.getOrRegisterClient(ctx)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get or register client")
		return nil, ErrTransientAwsClientError
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to authorize device")
		return nil, ErrTransientAwsClientError
	}

	return &AuthorizeDeviceFlowResult{
		InstanceId:      instanceId,
		ClientId:        regRes.ClientId,
		StartUrl:        startUrl,
		Region:          awsRegion,
		Label:           label,
		VerificationUri: authorizeRes.VerificationUriComplete,
		UserCode:        authorizeRes.UserCode,
		ExpiresIn:       authorizeRes.ExpiresIn,
		DeviceCode:      authorizeRes.DeviceCode,
	}, nil
}

func (c *AwsIdentityCenterController) FinalizeRefreshAccessToken(instanceId, region, userCode, deviceCode string) error {
	if err := c.validateAwsRegion(region); err != nil {
		return err
	}

	row := c.db.QueryRowContext(c.ctx, "SELECT client_id, client_secret_enc, enc_key_id FROM aws_iam_idc_clients")

	var clientId string
	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientId, &clientSecretEnc, &encKeyId); err != nil {
		c.errHandler.Catch(c.logger, err)
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to decrypt client secret")

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	tokenRes, err := c.getToken(ctx, clientId, clientSecret, deviceCode, userCode)

	if err != nil {
		if errors.Is(err, awssso.ErrDeviceFlowNotAuthorized) {
			c.logger.Debug().Err(err).Msg("failed to get token because user did not authorize device")
			return ErrDeviceAuthFlowNotAuthorized
		}

		if errors.Is(err, awssso.ErrDeviceCodeExpired) {
			c.logger.Debug().Err(err).Msg("failed to get token because user and device code expired")
			return ErrDeviceAuthFlowTimedOut
		}

		c.logger.Error().Err(err).Msg("failed to get token")
		return ErrTransientAwsClientError
	}

	idTokenEnc, keyId, err := c.encryptionService.Encrypt(tokenRes.IdToken)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt id token")

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt access token")

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt refresh token")

	c.logger.Info().Msgf("refreshing access token for instance [%s]", instanceId)

	sql := `UPDATE aws_iam_idc_instances SET
		id_token_enc = ?,
		access_token_enc = ?,
		token_type = ?,
		access_token_created_at = ?,
		access_token_expires_in = ?,
		refresh_token_enc = ?,
		enc_key_id = ?
		WHERE instance_id = ?;
		
		SELECT changes();
		`

	res, err := c.db.ExecContext(ctx, sql,
		idTokenEnc,
		accessTokenEnc,
		tokenRes.TokenType,
		c.timeHelper.NowUnix(),
		tokenRes.ExpiresIn,
		refreshTokenEnc,
		keyId, instanceId)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to persist refreshed access token to database")

	rowsAffected, err := res.RowsAffected()

	c.errHandler.CatchWithMsg(c.logger, err, "failed to get number of rows affected")

	if rowsAffected != 1 {
		return ErrInstanceWasNotFound
	}

	return nil
}

func (c *AwsIdentityCenterController) getOrRegisterClient(ctx context.Context) (*awssso.RegistrationResponse, error) {
	row := c.db.QueryRowContext(ctx, "SELECT client_id, client_secret_enc, created_at, expires_at, enc_key_id FROM aws_iam_idc_clients")

	var encKeyId string
	var result awssso.RegistrationResponse

	shouldRegisterClient := false

	if err := row.Scan(&result.ClientId, &result.ClientSecret, &result.CreatedAt, &result.ExpiresAt, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			shouldRegisterClient = true
		} else {
			c.errHandler.Catch(c.logger, err)
		}
	}

	if shouldRegisterClient {
		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		c.logger.Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, friendlyClientName)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to register client")
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt client secret")

		_, err = c.db.ExecContext(ctx, `INSERT INTO aws_iam_idc_clients
			(client_id, client_secret_enc, created_at, expires_at, enc_key_id)
			VALUES (?, ?, ?, ?, ?)`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId)

		c.errHandler.CatchWithMsg(c.logger, err, "failed to save client to database")

		c.logger.Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	if c.timeHelper.NowUnix() > result.ExpiresAt {
		c.logger.Info().Msg("client expired. registering new client")

		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		c.logger.Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, friendlyClientName)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to register client")
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		c.errHandler.CatchWithMsg(c.logger, err, "failed to encrypt client secret")

		_, err = c.db.ExecContext(ctx, `UPDATE aws_iam_idc_clients SET
			client_id = ?,
			client_secret_enc = ?,
			created_at = ?,
			expires_at = ?,
			enc_key_id = ?
			WHERE client_id = ?`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId, result.ClientId)

		c.errHandler.CatchWithMsg(c.logger, err, "failed to save client to database")

		c.logger.Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	var err error
	result.ClientSecret, err = c.encryptionService.Decrypt(result.ClientSecret, encKeyId)

	c.errHandler.CatchWithMsg(c.logger, err, "failed to decrypt client secret")

	return &result, nil
}

func (c *AwsIdentityCenterController) authorizeDevice(ctx context.Context, startUrl string, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	c.logger.Info().Msg("Authorizing Device")
	output, err := c.awsSsoClient.StartDeviceAuthorization(ctx, startUrl, clientId, clientSecret)

	if err != nil {
		return nil, err
	}

	c.logger.Info().Msgf("please login at %s?user_code=%s. You have %d seconds to do so", output.VerificationUri, output.UserCode, output.ExpiresIn)

	return output, nil
}

func (c *AwsIdentityCenterController) getToken(ctx context.Context, clientId, clientSecret, deviceCode, userCode string) (*awssso.GetTokenResponse, error) {
	c.logger.Info().Msg("getting access token")
	output, err := c.awsSsoClient.CreateToken(ctx,
		clientId,
		clientSecret,
		userCode,
		deviceCode,
	)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get access token")
		return nil, err
	}

	c.logger.Info().Msg("got access token")

	return output, nil
}
