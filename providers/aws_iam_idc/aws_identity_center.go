package awsiamidc

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/faults"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/abjrcode/swervo/providers"
	"github.com/coocood/freecache"
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
	logger            zerolog.Logger
	db                *sql.DB
	favoritesRepo     favorites.FavoritesRepo
	encryptionService encryption.EncryptionService
	awsSsoClient      awssso.AwsSsoOidcClient
	timeHelper        utils.Clock
	cache             *freecache.Cache
}

func NewAwsIdentityCenterController(db *sql.DB, favoritesRepo favorites.FavoritesRepo, encryptionService encryption.EncryptionService, awsSsoClient awssso.AwsSsoOidcClient, datetime utils.Clock, logger zerolog.Logger) *AwsIdentityCenterController {
	fiveHundredTwelveKilobytes := 512 * 1024
	cache := freecache.NewCache(fiveHundredTwelveKilobytes)

	logger = logger.With().Str("component", "aws_idc_controller").Logger()

	return &AwsIdentityCenterController{
		logger:            logger,
		db:                db,
		favoritesRepo:     favoritesRepo,
		encryptionService: encryptionService,
		awsSsoClient:      awsSsoClient,
		timeHelper:        datetime,
		cache:             cache,
	}
}

type AwsIdentityCenterAccountRole struct {
	RoleName string `json:"roleName"`
}

type AwsIdentityCenterAccount struct {
	AccountId   string                         `json:"accountId"`
	AccountName string                         `json:"accountName"`
	Roles       []AwsIdentityCenterAccountRole `json:"roles"`
}

type AwsIdentityCenterAccountRoleCredentials struct {
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken    string `json:"sessionToken"`
	Expiration      int64  `json:"expiration"`
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

func (c *AwsIdentityCenterController) ListInstances(ctx context.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT instance_id FROM aws_iam_idc_instances ORDER BY instance_id DESC")

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]string, 0), nil
		}

		return nil, errors.Join(err, faults.ErrFatal)
	}

	instances := make([]string, 0)

	for rows.Next() {
		var instanceId string

		if err := rows.Scan(&instanceId); err != nil {
			return nil, errors.Join(err, faults.ErrFatal)
		}

		instances = append(instances, instanceId)
	}

	return instances, nil
}

func (c *AwsIdentityCenterController) GetInstanceData(ctx context.Context, instanceId string, forceRefresh bool) (*AwsIdentityCenterCardData, error) {
	row := c.db.QueryRowContext(ctx, "SELECT region, label, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_iam_idc_instances WHERE instance_id = ?", instanceId)

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

		return nil, errors.Join(err, faults.ErrFatal)
	}

	isFavorite, err := c.favoritesRepo.IsFavorite(ctx, &favorites.Favorite{
		ProviderCode: providers.AwsIamIdc,
		InstanceId:   instanceId,
	})

	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}

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

	if !forceRefresh {
		got, err := c.cache.Get([]byte(instanceId))

		if err == nil {
			c.logger.Info().Msgf("cache hit for instance [%s]", instanceId)

			var accounts []AwsIdentityCenterAccount
			buffer := bytes.NewBuffer(got)

			err = gob.NewDecoder(buffer).Decode(&accounts)

			if err != nil {
				c.logger.Error().Err(err).Msg("failed to decode accounts from cache")
				return nil, errors.Join(err, faults.ErrFatal)
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
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}

	c.logger.Debug().Msgf("fetching accounts for instance [%s] with refresh=%t", instanceId, forceRefresh)
	accountsOut, err := c.awsSsoClient.ListAccounts(ctx, awssso.AwsRegion(region), accessToken)

	if err != nil {
		c.logger.Error().Err(err).Msg("aws sso client failed to list accounts")

		return nil, ErrTransientAwsClientError
	}

	buffer := bytes.NewBuffer([]byte{})

	err = gob.NewEncoder(buffer).Encode(accountsOut.Accounts)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to encode accounts to cache")
		return nil, errors.Join(err, faults.ErrFatal)
	}

	cacheTtl := accessTokenCreatedAt + accessTokenExpiresIn - now
	c.cache.Set([]byte(instanceId), buffer.Bytes(), int(cacheTtl))

	accounts := make([]AwsIdentityCenterAccount, 0)

	for _, account := range accountsOut.Accounts {
		accountRoles := make([]AwsIdentityCenterAccountRole, 0)

		for _, role := range account.Roles {
			accountRoles = append(accountRoles, AwsIdentityCenterAccountRole{
				RoleName: role.RoleName,
			})
		}

		accounts = append(accounts, AwsIdentityCenterAccount{
			AccountId:   account.AccountId,
			AccountName: account.AccountName,
			Roles:       accountRoles,
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

type AwsIamIdc_GetRoleCredentialsCommandInput struct {
	InstanceId string `json:"instanceId"`
	AccountId  string `json:"accountId"`
	RoleName   string `json:"roleName"`
}

func (c *AwsIdentityCenterController) GetRoleCredentials(ctx context.Context, input AwsIamIdc_GetRoleCredentialsCommandInput) (*AwsIdentityCenterAccountRoleCredentials, error) {
	row := c.db.QueryRowContext(ctx, "SELECT region, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_iam_idc_instances WHERE instance_id = ?", input.InstanceId)

	var region string
	var accessTokenEnc string
	var accessTokenCreatedAt int64
	var accessTokenExpiresIn int64
	var encKeyId string

	if err := row.Scan(&region, &accessTokenEnc, &accessTokenCreatedAt, &accessTokenExpiresIn, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceWasNotFound
		}

		return nil, errors.Join(err, faults.ErrFatal)
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	if err != nil {
		return nil, errors.Join(errors.New("failed to decrypt access token"), err, faults.ErrFatal)
	}

	res, err := c.awsSsoClient.GetRoleCredentials(ctx, awssso.AwsRegion(region), input.AccountId, input.RoleName, accessToken)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get role credentials")
		return nil, errors.Join(err, faults.ErrFatal)
	}

	return &AwsIdentityCenterAccountRoleCredentials{
		AccessKeyId:     res.AccessKeyId,
		SecretAccessKey: res.SecretAccessKey,
		SessionToken:    res.SessionToken,
		Expiration:      res.Expiration,
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

type AwsIamIdc_SetupCommandInput struct {
	StartUrl  string `json:"startUrl"`
	AwsRegion string `json:"awsRegion"`
	Label     string `json:"label"`
}

func (c *AwsIdentityCenterController) Setup(ctx context.Context, input AwsIamIdc_SetupCommandInput) (*AuthorizeDeviceFlowResult, error) {
	if err := c.validateStartUrl(input.StartUrl); err != nil {
		return nil, err
	}

	if err := c.validateAwsRegion(input.AwsRegion); err != nil {
		return nil, err
	}

	if err := c.validateLabel(input.Label); err != nil {
		return nil, err
	}

	startUrl := input.StartUrl
	awsRegion := input.AwsRegion
	label := input.Label

	var exists bool
	err := c.db.QueryRowContext(ctx, "SELECT 1 FROM aws_iam_idc_instances WHERE start_url = ?", startUrl).Scan(&exists)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(err, faults.ErrFatal)
	}

	if exists {
		c.logger.Warn().Msgf("instance [%s] already exists", startUrl)
		return nil, ErrInstanceAlreadyRegistered
	}

	if len(label) < 1 || len(label) > 50 {
		c.logger.Debug().Msgf("invalid label [%s]", label)
		return nil, ErrInvalidLabel
	}

	regRes, err := c.getOrRegisterClient(ctx, awsRegion)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get or register client")
		return nil, ErrTransientAwsClientError
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, awsRegion, regRes.ClientId, regRes.ClientSecret)

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

type AwsIamIdc_FinalizeSetupCommandInput struct {
	ClientId   string `json:"clientId"`
	StartUrl   string `json:"startUrl"`
	AwsRegion  string `json:"awsRegion"`
	Label      string `json:"label"`
	UserCode   string `json:"userCode"`
	DeviceCode string `json:"deviceCode"`
}

func (c *AwsIdentityCenterController) FinalizeSetup(ctx context.Context, input AwsIamIdc_FinalizeSetupCommandInput) (string, error) {
	if err := c.validateLabel(input.Label); err != nil {
		return "", err
	}

	if err := c.validateStartUrl(input.StartUrl); err != nil {
		return "", err
	}

	if err := c.validateAwsRegion(input.AwsRegion); err != nil {
		return "", err
	}

	row := c.db.QueryRowContext(ctx, "SELECT client_secret_enc, enc_key_id FROM aws_iam_idc_clients")

	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientSecretEnc, &encKeyId); err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	if err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}

	tokenRes, err := c.getToken(ctx, input.AwsRegion, input.ClientId, clientSecret, input.DeviceCode, input.UserCode)

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

	if err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)

	if err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)
	if err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}

	nowUnix := c.timeHelper.NowUnix()

	uniqueId, err := ksuid.NewRandomWithTime(time.Unix(nowUnix, 0))
	if err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}
	instanceId := uniqueId.String()

	sql := `INSERT INTO aws_iam_idc_instances
	(instance_id, start_url, region, label, enabled, id_token_enc, access_token_enc, token_type, access_token_created_at,
		access_token_expires_in, refresh_token_enc, enc_key_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = c.db.ExecContext(ctx, sql,
		instanceId,
		input.StartUrl,
		input.AwsRegion,
		input.Label,
		true,
		idTokenEnc,
		accessTokenEnc,
		tokenRes.TokenType,
		nowUnix,
		tokenRes.ExpiresIn,
		refreshTokenEnc,
		keyId)

	if err != nil {
		return "", errors.Join(err, faults.ErrFatal)
	}

	return instanceId, nil
}

func (c *AwsIdentityCenterController) MarkAsFavorite(ctx context.Context, instanceId string) error {
	return c.favoritesRepo.Add(ctx, &favorites.Favorite{
		ProviderCode: providers.AwsIamIdc,
		InstanceId:   instanceId,
	})
}

func (c *AwsIdentityCenterController) UnmarkAsFavorite(ctx context.Context, instanceId string) error {
	return c.favoritesRepo.Remove(ctx, &favorites.Favorite{
		ProviderCode: providers.AwsIamIdc,
		InstanceId:   instanceId,
	})
}

func (c *AwsIdentityCenterController) RefreshAccessToken(ctx context.Context, instanceId string) (*AuthorizeDeviceFlowResult, error) {
	var startUrl string
	var awsRegion string
	var label string
	row := c.db.QueryRowContext(ctx, "SELECT start_url, region, label FROM aws_iam_idc_instances WHERE instance_id = ?", instanceId)

	if err := row.Scan(&startUrl, &awsRegion, &label); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Debug().Msgf("instance [%s] was not found", instanceId)
			return nil, ErrInstanceWasNotFound
		}
	}

	regRes, err := c.getOrRegisterClient(ctx, awsRegion)

	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get or register client")
		return nil, ErrTransientAwsClientError
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, awsRegion, regRes.ClientId, regRes.ClientSecret)

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

type AwsIamIdc_FinalizeRefreshAccessTokenCommandInput struct {
	InstanceId string `json:"instanceId"`
	Region     string `json:"region"`
	UserCode   string `json:"userCode"`
	DeviceCode string `json:"deviceCode"`
}

func (c *AwsIdentityCenterController) FinalizeRefreshAccessToken(ctx context.Context, input AwsIamIdc_FinalizeRefreshAccessTokenCommandInput) error {
	if err := c.validateAwsRegion(input.Region); err != nil {
		return err
	}

	row := c.db.QueryRowContext(ctx, "SELECT client_id, client_secret_enc, enc_key_id FROM aws_iam_idc_clients")

	var clientId string
	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientId, &clientSecretEnc, &encKeyId); err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	tokenRes, err := c.getToken(ctx, input.Region, clientId, clientSecret, input.DeviceCode, input.UserCode)

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

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	c.logger.Info().Msgf("refreshing access token for instance [%s]", input.InstanceId)

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
		keyId, input.InstanceId)

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return errors.Join(err, faults.ErrFatal)
	}

	if rowsAffected != 1 {
		return ErrInstanceWasNotFound
	}

	return nil
}

func (c *AwsIdentityCenterController) getOrRegisterClient(ctx context.Context, awsRegion string) (*awssso.RegistrationResponse, error) {
	row := c.db.QueryRowContext(ctx, "SELECT client_id, client_secret_enc, created_at, expires_at, enc_key_id FROM aws_iam_idc_clients")

	var encKeyId string
	var result awssso.RegistrationResponse

	shouldRegisterClient := false

	if err := row.Scan(&result.ClientId, &result.ClientSecret, &result.CreatedAt, &result.ExpiresAt, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			shouldRegisterClient = true
		} else {
			return nil, errors.Join(err, faults.ErrFatal)
		}
	}

	if shouldRegisterClient {
		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		c.logger.Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, awssso.AwsRegion(awsRegion), friendlyClientName)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to register client")
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		if err != nil {
			return nil, errors.Join(err, faults.ErrFatal)
		}

		_, err = c.db.ExecContext(ctx, `INSERT INTO aws_iam_idc_clients
			(client_id, client_secret_enc, created_at, expires_at, enc_key_id)
			VALUES (?, ?, ?, ?, ?)`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId)

		if err != nil {
			return nil, errors.Join(err, faults.ErrFatal)
		}

		c.logger.Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	if c.timeHelper.NowUnix() > result.ExpiresAt {
		c.logger.Info().Msg("client expired. registering new client")

		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		c.logger.Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, awssso.AwsRegion(awsRegion), friendlyClientName)
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to register client")
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		if err != nil {
			return nil, errors.Join(err, faults.ErrFatal)
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
			return nil, errors.Join(err, faults.ErrFatal)
		}

		c.logger.Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	var err error
	result.ClientSecret, err = c.encryptionService.Decrypt(result.ClientSecret, encKeyId)

	if err != nil {
		return nil, errors.Join(err, faults.ErrFatal)
	}

	return &result, nil
}

func (c *AwsIdentityCenterController) authorizeDevice(ctx context.Context, startUrl, region string, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	c.logger.Info().Msg("Authorizing Device")

	output, err := c.awsSsoClient.StartDeviceAuthorization(ctx, awssso.AwsRegion(region), startUrl, clientId, clientSecret)

	if err != nil {
		return nil, err
	}

	c.logger.Info().Msgf("please login at %s?user_code=%s. You have %d seconds to do so", output.VerificationUri, output.UserCode, output.ExpiresIn)

	return output, nil
}

func (c *AwsIdentityCenterController) getToken(ctx context.Context, awsRegion, clientId, clientSecret, deviceCode, userCode string) (*awssso.GetTokenResponse, error) {
	c.logger.Info().Msg("getting access token")
	output, err := c.awsSsoClient.CreateToken(ctx,
		awssso.AwsRegion(awsRegion),
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
