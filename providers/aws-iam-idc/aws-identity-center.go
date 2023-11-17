package awsiamidc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/dustin/go-humanize"
	"github.com/rs/zerolog"
)

var (
	ErrInvalidStartUrl             = errors.New("INVALID_START_URL")
	ErrInvalidAwsRegion            = errors.New("INVALID_AWS_REGION")
	ErrDeviceAuthFlowNotAuthorized = errors.New("DEVICE_AUTH_FLOW_NOT_AUTHORIZED")
	ErrDeviceAuthFlowTimedOut      = errors.New("DEVICE_AUTH_FLOW_TIMED_OUT")
	ErrAccessTokenExpired          = errors.New("ACCESS_TOKEN_EXPIRED")
	ErrInstanceAlreadyRegistered   = errors.New("INSTANCE_ALREADY_REGISTERED")
)

type AwsIdentityCenterController struct {
	ctx               context.Context
	logger            *zerolog.Logger
	db                *sql.DB
	encryptionService encryption.EncryptionService
	awsSsoClient      awssso.AwsSsoOidcClient
	timeHelper        utils.Datetime
}

func NewAwsIdentityCenterController(db *sql.DB, encryptionService encryption.EncryptionService, awsSsoClient awssso.AwsSsoOidcClient, datetime utils.Datetime) *AwsIdentityCenterController {
	return &AwsIdentityCenterController{
		db:                db,
		encryptionService: encryptionService,
		awsSsoClient:      awsSsoClient,
		timeHelper:        datetime,
	}
}

func (controller *AwsIdentityCenterController) Init(ctx context.Context) {
	controller.ctx = ctx
	controller.logger = zerolog.Ctx(ctx)
}

type AwsIdentityCenterAccount struct {
	AccountId   string `json:"accountId"`
	AccountName string `json:"accountName"`
}

type AwsIdentityCenterCardData struct {
	Enabled              bool                       `json:"enabled"`
	AccessTokenExpiresIn string                     `json:"accessTokenExpiresIn"`
	Accounts             []AwsIdentityCenterAccount `json:"accounts"`
}

func (c *AwsIdentityCenterController) GetInstanceData(startUrl string) (*AwsIdentityCenterCardData, error) {
	row := c.db.QueryRowContext(c.ctx, "SELECT region, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_iam_idc_instances WHERE start_url = ?", startUrl)

	var region string
	var accessTokenEnc string
	var accessTokenCreatedAt int64
	var accessTokenExpiresIn int64
	var encKeyId string

	if err := row.Scan(&region, &accessTokenEnc, &accessTokenCreatedAt, &accessTokenExpiresIn, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Debug().Msgf("no token found for start URL [%s]", startUrl)
			return nil, nil
		}

		return nil, err
	}

	now := c.timeHelper.NowUnix()
	if now > accessTokenCreatedAt+accessTokenExpiresIn {
		c.logger.Debug().Msgf("token for start URL [%s] is expired", startUrl)

		return nil, ErrAccessTokenExpired
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to decrypt access token")
		return nil, err
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	accountsOut, err := c.awsSsoClient.ListAccounts(ctx, accessToken)

	if err != nil {
		if errors.Is(err, awssso.ErrAccessTokenExpired) {
			return nil, ErrAccessTokenExpired
		}

		c.logger.Error().Err(err).Msg("failed to list accounts")
		return nil, err
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
		AccessTokenExpiresIn: humanize.Time(time.Unix(accessTokenCreatedAt+accessTokenExpiresIn, 0)),
		Accounts:             accounts,
	}, nil
}

type AuthorizeDeviceFlowResult struct {
	ClientId        string `json:"clientId"`
	StartUrl        string `json:"startUrl"`
	Region          string `json:"region"`
	VerificationUri string `json:"verificationUri"`
	UserCode        string `json:"userCode"`
	ExpiresIn       int32  `json:"expiresIn"`
	DeviceCode      string `json:"deviceCode"`
}

func (c *AwsIdentityCenterController) Setup(startUrlStr, awsRegion string) (*AuthorizeDeviceFlowResult, error) {
	if _, ok := awssso.SupportedAwsRegions[awsRegion]; !ok {
		c.logger.Error().Msgf("unsupported AWS region [%s]", awsRegion)
		return nil, ErrInvalidAwsRegion
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), awsRegion)
	startUrl, err := url.Parse(startUrlStr)

	if startUrl.Scheme == "" || startUrl.Host == "" {
		c.logger.Error().Msgf("invalid start URL [%s]", startUrlStr)
		return nil, ErrInvalidStartUrl
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to parse start URL")
		return nil, err
	}

	var exists bool
	err = c.db.QueryRowContext(ctx, "SELECT 1 FROM aws_iam_idc_instances WHERE start_url = ?", startUrlStr).Scan(&exists)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		c.logger.Error().Err(err).Msg("failed to check if instance exists")
		return nil, err
	}

	if exists {
		c.logger.Error().Msgf("instance [%s] already exists", startUrlStr)
		return nil, ErrInstanceAlreadyRegistered
	}

	regRes, err := c.getOrRegisterClient(ctx)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to register client")
		return nil, err
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to authorize device")
		return nil, err
	}

	return &AuthorizeDeviceFlowResult{
		ClientId:        regRes.ClientId,
		StartUrl:        startUrlStr,
		Region:          awsRegion,
		VerificationUri: authorizeRes.VerificationUriComplete,
		UserCode:        authorizeRes.UserCode,
		ExpiresIn:       authorizeRes.ExpiresIn,
		DeviceCode:      authorizeRes.DeviceCode,
	}, nil
}

func (c *AwsIdentityCenterController) FinalizeSetup(clientId, startUrl, region, userCode, deviceCode string) error {
	row := c.db.QueryRowContext(c.ctx, "SELECT client_secret_enc, enc_key_id FROM aws_iam_idc_clients")

	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientSecretEnc, &encKeyId); err != nil {
		return err
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to decrypt client secret")
		return err
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	tokenRes, err := c.getToken(ctx, clientId, clientSecret, deviceCode, userCode)

	if err != nil {
		if errors.Is(err, awssso.ErrDeviceFlowNotAuthorized) {
			c.logger.Error().Err(err).Msg("failed to get token because user did not authorize device")
			return ErrDeviceAuthFlowNotAuthorized
		}

		if errors.Is(err, awssso.ErrDeviceCodeExpired) {
			c.logger.Error().Err(err).Msg("failed to get token because user and device code expired")
			return ErrDeviceAuthFlowTimedOut
		}
		c.logger.Error().Err(err).Msg("failed to get token")
		return err
	}

	tx, err := c.db.BeginTx(ctx, nil)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to start transaction")
		return err
	}

	defer tx.Rollback()

	idTokenEnc, keyId, err := c.encryptionService.Encrypt(tokenRes.IdToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encrypt id token")
		return err
	}

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encrypt access token")
		return err
	}

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encrypt refresh token")
		return err
	}

	sql := `INSERT INTO aws_iam_idc_instances
	(start_url, region, enabled, id_token_enc, access_token_enc, token_type, access_token_created_at,
		access_token_expires_in, refresh_token_enc, enc_key_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, sql,
		startUrl,
		region,
		true,
		idTokenEnc,
		accessTokenEnc,
		tokenRes.TokenType,
		c.timeHelper.NowUnix(),
		tokenRes.ExpiresIn,
		refreshTokenEnc,
		keyId)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create token")
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO providers (code, instance_id, display_name, is_favorite) VALUES (?, ?, ?, ?) `, "aws-iam-idc", startUrl, "AWS IAM Identity Center", true)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to add provider to list of configured providers")
		return err
	}

	err = tx.Commit()

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to commit transaction")
		return err
	}

	return nil
}

func (c *AwsIdentityCenterController) RefreshAccessToken(startUrlStr string) (*AuthorizeDeviceFlowResult, error) {
	startUrl, err := url.Parse(startUrlStr)

	if startUrl.Scheme == "" || startUrl.Host == "" {
		c.logger.Error().Msgf("invalid start URL [%s]", startUrlStr)
		return nil, ErrInvalidStartUrl
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to parse start URL")
		return nil, err
	}

	var awsRegion string
	row := c.db.QueryRowContext(c.ctx, "SELECT region FROM aws_iam_idc_instances WHERE start_url = ?", startUrlStr)

	if err := row.Scan(&awsRegion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Debug().Msgf("no instance found for start URL [%s]", startUrlStr)
			// TODO: probably should panic
			return nil, errors.New("no instance found for start URL")
		}
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), awsRegion)
	regRes, err := c.getOrRegisterClient(ctx)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get or register client")
		return nil, err
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to authorize device")
		return nil, err
	}

	return &AuthorizeDeviceFlowResult{
		ClientId:        regRes.ClientId,
		StartUrl:        startUrlStr,
		Region:          awsRegion,
		VerificationUri: authorizeRes.VerificationUriComplete,
		UserCode:        authorizeRes.UserCode,
		ExpiresIn:       authorizeRes.ExpiresIn,
		DeviceCode:      authorizeRes.DeviceCode,
	}, nil
}

func (c *AwsIdentityCenterController) FinalizeRefreshAccessToken(clientId, startUrl, region, userCode, deviceCode string) error {
	row := c.db.QueryRowContext(c.ctx, "SELECT client_secret_enc, enc_key_id FROM aws_iam_idc_clients")

	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientSecretEnc, &encKeyId); err != nil {
		return err
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to decrypt client secret")
		return err
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	tokenRes, err := c.getToken(ctx, clientId, clientSecret, deviceCode, userCode)

	if err != nil {
		if errors.Is(err, awssso.ErrDeviceFlowNotAuthorized) {
			c.logger.Error().Err(err).Msg("failed to get token because user did not authorize device")
			return ErrDeviceAuthFlowNotAuthorized
		}

		if errors.Is(err, awssso.ErrDeviceCodeExpired) {
			c.logger.Error().Err(err).Msg("failed to get token because user and device code expired")
			return ErrDeviceAuthFlowTimedOut
		}

		c.logger.Error().Err(err).Msg("failed to get token")
		return err
	}

	idTokenEnc, keyId, err := c.encryptionService.Encrypt(tokenRes.IdToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encrypt id token")
		return err
	}

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encrypt access token")
		return err
	}

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encrypt refresh token")
		return err
	}

	sql := `UPDATE aws_iam_idc_instances SET
		id_token_enc = ?,
		access_token_enc = ?,
		token_type = ?,
		access_token_created_at = ?,
		access_token_expires_in = ?,
		refresh_token_enc = ?,
		enc_key_id = ?
		WHERE start_url = ?`
	_, err = c.db.ExecContext(ctx, sql,
		idTokenEnc,
		accessTokenEnc,
		tokenRes.TokenType,
		c.timeHelper.NowUnix(),
		tokenRes.ExpiresIn,
		refreshTokenEnc,
		keyId, startUrl)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to refresh access token")
		return err
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
			return nil, err
		}
	}

	if shouldRegisterClient {
		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		c.logger.Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, friendlyClientName)
		if err != nil {
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		if err != nil {
			c.logger.Error().Err(err).Msg("failed to encrypt client secret")
			return nil, err
		}

		_, err = c.db.ExecContext(ctx, `INSERT INTO aws_iam_idc_clients
			(client_id, client_secret_enc, created_at, expires_at, enc_key_id)
			VALUES (?, ?, ?, ?, ?)`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId)

		if err != nil {
			c.logger.Error().Err(err).Msg("failed to insert client into database")
			return nil, err
		}

		c.logger.Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	if c.timeHelper.NowUnix() > result.ExpiresAt {
		c.logger.Info().Msg("client expired. registering new client")

		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		c.logger.Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, friendlyClientName)
		if err != nil {
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		if err != nil {
			c.logger.Error().Err(err).Msg("failed to encrypt client secret")
			return nil, err
		}

		_, err = c.db.ExecContext(ctx, `UPDATE aws_iam_idc_clients SET
			client_id = ?,
			client_secret_enc = ?,
			created_at = ?,
			expires_at = ?,
			enc_key_id = ?
			WHERE client_id = ?`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId, result.ClientId)

		if err != nil {
			c.logger.Error().Err(err).Msg("failed to update client in database")
			return nil, err
		}

		c.logger.Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	var err error
	result.ClientSecret, err = c.encryptionService.Decrypt(result.ClientSecret, encKeyId)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to decrypt client secret")
		return nil, err
	}

	return &result, nil
}

func (c *AwsIdentityCenterController) authorizeDevice(ctx context.Context, startUrl *url.URL, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	c.logger.Info().Msg("Authorizing Device")
	output, err := c.awsSsoClient.StartDeviceAuthorization(ctx, *startUrl, clientId, clientSecret)

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
