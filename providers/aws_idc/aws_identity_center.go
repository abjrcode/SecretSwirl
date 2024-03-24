package awsidc

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"errors"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/abjrcode/swervo/clients/awscredsfile"
	"github.com/abjrcode/swervo/clients/awssso"
	"github.com/abjrcode/swervo/favorites"
	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/eventing"
	"github.com/abjrcode/swervo/internal/plumbing"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/coocood/freecache"
	"github.com/dustin/go-humanize"
	"github.com/segmentio/ksuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var ProviderCode = "aws-idc"

var (
	ErrInvalidStartUrl             = app.NewValidationError("INVALID_START_URL")
	ErrInvalidAwsRegion            = app.NewValidationError("INVALID_AWS_REGION")
	ErrInvalidLabel                = app.NewValidationError("INVALID_LABEL")
	ErrDeviceAuthFlowNotAuthorized = app.NewValidationError("DEVICE_AUTH_FLOW_NOT_AUTHORIZED")
	ErrDeviceAuthFlowTimedOut      = app.NewValidationError("DEVICE_AUTH_FLOW_TIMED_OUT")
	ErrInstanceWasNotFound         = app.NewValidationError("INSTANCE_WAS_NOT_FOUND")
	ErrInstanceAlreadyRegistered   = app.NewValidationError("INSTANCE_ALREADY_REGISTERED")
	ErrStaleAwsAccessToken         = app.NewValidationError("STALE_AWS_ACCESS_TOKEN")
	ErrTransientAwsClientError     = app.NewValidationError("TRANSIENT_AWS_CLIENT_ERROR")
)

var AwsIdcEventSource = eventing.EventSource("AwsIdc")

type AwsIdcInstanceCreatedEvent struct {
	InstanceId string

	StartUrl string
	Region   string
	Label    string
}

type AwsCredentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

type AwsIdentityCenterController struct {
	db                *sql.DB
	bus               *eventing.Eventbus
	favoritesRepo     favorites.FavoritesRepo
	encryptionService encryption.EncryptionService
	awsSsoClient      awssso.AwsSsoOidcClient
	clock             utils.Clock
	cache             *freecache.Cache

	plumbers []plumbing.Plumber[AwsCredentials]
}

func NewAwsIdentityCenterController(db *sql.DB, bus *eventing.Eventbus, favoritesRepo favorites.FavoritesRepo, encryptionService encryption.EncryptionService, awsSsoClient awssso.AwsSsoOidcClient, datetime utils.Clock) *AwsIdentityCenterController {
	fiveHundredTwelveKilobytes := 512 * 1024
	cache := freecache.NewCache(fiveHundredTwelveKilobytes)

	return &AwsIdentityCenterController{
		db:                db,
		bus:               bus,
		favoritesRepo:     favoritesRepo,
		encryptionService: encryptionService,
		awsSsoClient:      awsSsoClient,
		clock:             datetime,
		cache:             cache,
		plumbers:          make([]plumbing.Plumber[AwsCredentials], 0),
	}
}

func (c *AwsIdentityCenterController) AddPlumbers(plumbers ...plumbing.Plumber[AwsCredentials]) {
	c.plumbers = append(c.plumbers, plumbers...)
}

type AwsIdentityCenterAccountRole struct {
	RoleName string `json:"roleName"`
}

type AwsIdentityCenterAccount struct {
	AccountId   string                         `json:"accountId"`
	AccountName string                         `json:"accountName"`
	Roles       []AwsIdentityCenterAccountRole `json:"roles"`
}

type AwsIdentityCenterCardData struct {
	InstanceId           string                     `json:"instanceId"`
	Enabled              bool                       `json:"enabled"`
	Label                string                     `json:"label"`
	IsFavorite           bool                       `json:"isFavorite"`
	IsAccessTokenExpired bool                       `json:"isAccessTokenExpired"`
	AccessTokenExpiresIn string                     `json:"accessTokenExpiresIn"`
	Accounts             []AwsIdentityCenterAccount `json:"accounts"`

	Sinks []plumbing.SinkInstance `json:"sinks"`
}

func (c *AwsIdentityCenterController) ListInstances(ctx app.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT instance_id FROM aws_idc ORDER BY instance_id DESC")

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make([]string, 0), nil
		}

		return nil, errors.Join(err, app.ErrFatal)
	}

	instances := make([]string, 0)

	for rows.Next() {
		var instanceId string

		if err := rows.Scan(&instanceId); err != nil {
			return nil, errors.Join(err, app.ErrFatal)
		}

		instances = append(instances, instanceId)
	}

	return instances, nil
}

func (c *AwsIdentityCenterController) invalidateStaleAccessToken(ctx app.Context, instanceId string) error {
	ctx.Logger().Trace().Msgf("invalidating stale access token for instance [%s]", instanceId)

	c.cache.Del([]byte(instanceId))

	_, err := c.db.ExecContext(ctx, "UPDATE aws_idc SET access_token_expires_in = 0 WHERE instance_id = ?", instanceId)

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	return nil
}

func (c *AwsIdentityCenterController) GetInstanceData(ctx app.Context, instanceId string, forceRefresh bool) (*AwsIdentityCenterCardData, error) {
	row := c.db.QueryRowContext(ctx, "SELECT region, label, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_idc WHERE instance_id = ?", instanceId)

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

		return nil, errors.Join(err, app.ErrFatal)
	}

	isFavorite, err := c.favoritesRepo.IsFavorite(ctx, &favorites.Favorite{
		ProviderCode: ProviderCode,
		InstanceId:   instanceId,
	})

	if err != nil {
		return nil, errors.Join(err, app.ErrFatal)
	}

	sinks := make([]plumbing.SinkInstance, 0)

	now := c.clock.NowUnix()
	if now > accessTokenCreatedAt+accessTokenExpiresIn {
		ctx.Logger().Info().Msgf("token for instance [%s] has expired", instanceId)

		return &AwsIdentityCenterCardData{
			Enabled:              true,
			InstanceId:           instanceId,
			Label:                label,
			IsFavorite:           isFavorite,
			IsAccessTokenExpired: true,
			AccessTokenExpiresIn: humanize.Time(time.Unix(accessTokenCreatedAt+accessTokenExpiresIn, 0)),
			Accounts:             make([]AwsIdentityCenterAccount, 0),

			Sinks: sinks,
		}, nil
	}

	for _, plumber := range c.plumbers {
		connectedSinks, err := plumber.ListConnectedSinks(ctx, ProviderCode, instanceId)

		if err != nil {
			ctx.Logger().Error().Err(err).Msg("failed to list pipes")
			return nil, errors.Join(err, app.ErrFatal)
		}

		for _, pipe := range connectedSinks {
			sinks = append(sinks, plumbing.SinkInstance{
				SinkCode: pipe.SinkCode,
				SinkId:   pipe.SinkId,
			})
		}
	}

	if !forceRefresh {
		got, err := c.cache.Get([]byte(instanceId))

		if err == nil {
			ctx.Logger().Info().Msgf("cache hit for instance [%s]", instanceId)

			var accounts []AwsIdentityCenterAccount
			buffer := bytes.NewBuffer(got)

			err = gob.NewDecoder(buffer).Decode(&accounts)

			if err != nil {
				ctx.Logger().Error().Err(err).Msg("failed to decode accounts from cache")
				return nil, errors.Join(err, app.ErrFatal)
			}

			return &AwsIdentityCenterCardData{
				Enabled:              true,
				InstanceId:           instanceId,
				Label:                label,
				IsFavorite:           isFavorite,
				IsAccessTokenExpired: false,
				AccessTokenExpiresIn: humanize.Time(time.Unix(accessTokenCreatedAt+accessTokenExpiresIn, 0)),
				Accounts:             accounts,

				Sinks: sinks,
			}, nil
		}
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	if err != nil {
		return nil, errors.Join(err, app.ErrFatal)
	}

	ctx.Logger().Debug().Msgf("fetching accounts for instance [%s] with refresh=%t", instanceId, forceRefresh)
	accountsOut, err := c.awsSsoClient.ListAccounts(ctx, awssso.AwsRegion(region), accessToken)

	if err != nil {
		if errors.Is(err, awssso.ErrAccessTokenExpired) {
			ctx.Logger().Debug().Msgf("token for instance [%s] has expired", instanceId)

			c.invalidateStaleAccessToken(ctx, instanceId)

			return &AwsIdentityCenterCardData{
				Enabled:              true,
				InstanceId:           instanceId,
				Label:                label,
				IsFavorite:           isFavorite,
				IsAccessTokenExpired: true,
				AccessTokenExpiresIn: "stale",
				Accounts:             make([]AwsIdentityCenterAccount, 0),

				Sinks: sinks,
			}, nil
		}

		ctx.Logger().Error().Err(err).Msg("aws sso client failed to list accounts")

		return nil, ErrTransientAwsClientError
	}

	buffer := bytes.NewBuffer([]byte{})

	err = gob.NewEncoder(buffer).Encode(accountsOut.Accounts)

	if err != nil {
		ctx.Logger().Error().Err(err).Msg("failed to encode accounts to cache")
		return nil, errors.Join(err, app.ErrFatal)
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

		Sinks: sinks,
	}, nil
}

type awsRoleCredentials struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	Expiration      int64
}

func (c *AwsIdentityCenterController) getRoleCredentials(ctx app.Context, instanceId, accountId, roleName string) (*awsRoleCredentials, error) {
	row := c.db.QueryRowContext(ctx, "SELECT region, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_idc WHERE instance_id = ?", instanceId)

	var region string
	var accessTokenEnc string
	var accessTokenCreatedAt int64
	var accessTokenExpiresIn int64
	var encKeyId string

	if err := row.Scan(&region, &accessTokenEnc, &accessTokenCreatedAt, &accessTokenExpiresIn, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceWasNotFound
		}

		return nil, errors.Join(err, app.ErrFatal)
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	if err != nil {
		return nil, errors.Join(errors.New("failed to decrypt access token"), err, app.ErrFatal)
	}

	res, err := c.awsSsoClient.GetRoleCredentials(ctx, awssso.AwsRegion(region), accountId, roleName, accessToken)

	if err != nil {
		if errors.Is(err, awssso.ErrAccessTokenExpired) {
			ctx.Logger().Debug().Msgf("token for instance [%s] has expired", instanceId)

			c.invalidateStaleAccessToken(ctx, instanceId)

			return nil, ErrStaleAwsAccessToken
		}

		ctx.Logger().Error().Err(err).Msg("failed to get role credentials")
		return nil, errors.Join(err, app.ErrFatal)
	}

	return &awsRoleCredentials{
		AccessKeyId:     res.AccessKeyId,
		SecretAccessKey: res.SecretAccessKey,
		SessionToken:    res.SessionToken,
		Expiration:      res.Expiration,
	}, nil
}

type AwsIdc_CopyRoleCredentialsCommandInput struct {
	InstanceId string `json:"instanceId"`
	AccountId  string `json:"accountId"`
	RoleName   string `json:"roleName"`
}

func (c *AwsIdentityCenterController) CopyRoleCredentials(ctx app.Context, input AwsIdc_CopyRoleCredentialsCommandInput) error {
	res, err := c.getRoleCredentials(ctx, input.InstanceId, input.AccountId, input.RoleName)

	if err != nil {
		return err
	}

	output := ""

	if runtime.GOOS == "windows" {
		output = fmt.Sprintf(`$Env:AWS_ACCESS_KEY_ID="%s"
$Env:AWS_SECRET_ACCESS_KEY="%s"
$Env:AWS_SESSION_TOKEN="%s"`, res.AccessKeyId, res.SecretAccessKey, res.SessionToken)
	} else {
		output = fmt.Sprintf(`export AWS_ACCESS_KEY_ID="%s"
export AWS_SECRET_ACCESS_KEY="%s"
export AWS_SESSION_TOKEN="%s"`, res.AccessKeyId, res.SecretAccessKey, res.SessionToken)
	}

	err = wailsRuntime.ClipboardSetText(ctx, output)

	if err != nil {
		ctx.Logger().Error().Err(err).Msg("failed to copy to clipboard")
		return errors.Join(err, app.ErrFatal)
	}

	return nil
}

type AwsIdc_SaveRoleCredentialsCommandInput struct {
	InstanceId string `json:"instanceId"`

	AccountId  string `json:"accountId"`
	RoleName   string `json:"roleName"`
	AwsProfile string `json:"awsProfile"`
}

func (c *AwsIdentityCenterController) SaveRoleCredentials(ctx app.Context, input AwsIdc_SaveRoleCredentialsCommandInput) error {
	result, err := c.getRoleCredentials(ctx, input.InstanceId, input.AccountId, input.RoleName)

	if err != nil {
		return err
	}

	credsFile := awscredsfile.NewDefaultCredentialsFileManager()

	err = credsFile.WriteProfileCredentials(input.AwsProfile, awscredsfile.ProfileCreds{
		AwsAccessKeyId:     result.AccessKeyId,
		AwsSecretAccessKey: result.SecretAccessKey,
		AwsSessionToken:    &result.SessionToken,
	})

	if err != nil {
		if errors.Is(err, app.ErrValidation) {
			return err
		} else {
			return errors.Join(err, app.ErrFatal)
		}
	}

	return nil
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

type AwsIdc_SetupCommandInput struct {
	StartUrl  string `json:"startUrl"`
	AwsRegion string `json:"awsRegion"`
	Label     string `json:"label"`
}

func (c *AwsIdentityCenterController) Setup(ctx app.Context, input AwsIdc_SetupCommandInput) (*AuthorizeDeviceFlowResult, error) {
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
	err := c.db.QueryRowContext(ctx, "SELECT 1 FROM aws_idc WHERE start_url = ?", startUrl).Scan(&exists)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(err, app.ErrFatal)
	}

	if exists {
		ctx.Logger().Warn().Msgf("instance [%s] already exists", startUrl)
		return nil, ErrInstanceAlreadyRegistered
	}

	regRes, err := c.getOrRegisterClient(ctx, awsRegion)

	if err != nil {
		ctx.Logger().Error().Err(err).Msg("failed to get or register client")
		return nil, ErrTransientAwsClientError
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, awsRegion, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		if errors.Is(err, awssso.ErrInvalidRequest) {
			ctx.Logger().Debug().Err(err).Msg("failed to authorize device because start URL is invalid")
			return nil, ErrInvalidStartUrl
		}

		ctx.Logger().Error().Err(err).Msg("failed to authorize device")
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

type AwsIdc_FinalizeSetupCommandInput struct {
	ClientId   string `json:"clientId"`
	StartUrl   string `json:"startUrl"`
	AwsRegion  string `json:"awsRegion"`
	Label      string `json:"label"`
	UserCode   string `json:"userCode"`
	DeviceCode string `json:"deviceCode"`
}

func (c *AwsIdentityCenterController) FinalizeSetup(ctx app.Context, input AwsIdc_FinalizeSetupCommandInput) (string, error) {
	if err := c.validateLabel(input.Label); err != nil {
		return "", err
	}

	if err := c.validateStartUrl(input.StartUrl); err != nil {
		return "", err
	}

	if err := c.validateAwsRegion(input.AwsRegion); err != nil {
		return "", err
	}

	row := c.db.QueryRowContext(ctx, "SELECT client_secret_enc, enc_key_id FROM aws_sso_clients")

	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientSecretEnc, &encKeyId); err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	tokenRes, err := c.getToken(ctx, input.AwsRegion, input.ClientId, clientSecret, input.DeviceCode, input.UserCode)

	if err != nil {
		if errors.Is(err, awssso.ErrDeviceFlowNotAuthorized) {
			ctx.Logger().Debug().Err(err).Msg("failed to get token because user did not authorize device")
			return "", ErrDeviceAuthFlowNotAuthorized
		}

		if errors.Is(err, awssso.ErrDeviceCodeExpired) {
			ctx.Logger().Debug().Err(err).Msg("failed to get token because user and device code expired")
			return "", ErrDeviceAuthFlowTimedOut
		}

		ctx.Logger().Error().Err(err).Msg("failed to get token")
		return "", ErrTransientAwsClientError
	}

	idTokenEnc, keyId, err := c.encryptionService.Encrypt(tokenRes.IdToken)

	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)

	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)
	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	nowUnix := c.clock.NowUnix()

	uniqueId, err := ksuid.NewRandomWithTime(time.Unix(nowUnix, 0))
	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}
	instanceId := uniqueId.String()
	version := 1

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}
	defer tx.Rollback()

	sql := `INSERT INTO aws_idc
	(instance_id, version, start_url, region, label, enabled, id_token_enc, access_token_enc, token_type, access_token_created_at,
		access_token_expires_in, refresh_token_enc, enc_key_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.ExecContext(ctx, sql,
		instanceId,
		version,
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
		return "", errors.Join(err, app.ErrFatal)
	}

	publish, err := c.bus.PublishTx(ctx, AwsIdcInstanceCreatedEvent{
		InstanceId: instanceId,
		StartUrl:   input.StartUrl,
		Region:     input.AwsRegion,
		Label:      input.Label,
	}, eventing.EventMeta{
		SourceType:   AwsIdcEventSource,
		SourceId:     instanceId,
		EventVersion: uint(version),
	}, tx)

	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	err = tx.Commit()
	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	publish()

	return instanceId, nil
}

func (c *AwsIdentityCenterController) MarkAsFavorite(ctx app.Context, instanceId string) error {
	return c.favoritesRepo.Add(ctx, &favorites.Favorite{
		ProviderCode: ProviderCode,
		InstanceId:   instanceId,
	})
}

func (c *AwsIdentityCenterController) UnmarkAsFavorite(ctx app.Context, instanceId string) error {
	return c.favoritesRepo.Remove(ctx, &favorites.Favorite{
		ProviderCode: ProviderCode,
		InstanceId:   instanceId,
	})
}

func (c *AwsIdentityCenterController) RefreshAccessToken(ctx app.Context, instanceId string) (*AuthorizeDeviceFlowResult, error) {
	var startUrl string
	var awsRegion string
	var label string

	row := c.db.QueryRowContext(ctx, "SELECT start_url, region, label FROM aws_idc WHERE instance_id = ?", instanceId)

	if err := row.Scan(&startUrl, &awsRegion, &label); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Logger().Debug().Msgf("instance [%s] was not found", instanceId)
			return nil, ErrInstanceWasNotFound
		}
	}

	regRes, err := c.getOrRegisterClient(ctx, awsRegion)

	if err != nil {
		ctx.Logger().Error().Err(err).Msg("failed to get or register client")
		return nil, ErrTransientAwsClientError
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, awsRegion, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		ctx.Logger().Error().Err(err).Msg("failed to authorize device")
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

type AwsIdc_FinalizeRefreshAccessTokenCommandInput struct {
	InstanceId string `json:"instanceId"`
	Region     string `json:"region"`
	UserCode   string `json:"userCode"`
	DeviceCode string `json:"deviceCode"`
}

func (c *AwsIdentityCenterController) FinalizeRefreshAccessToken(ctx app.Context, input AwsIdc_FinalizeRefreshAccessTokenCommandInput) error {
	if err := c.validateAwsRegion(input.Region); err != nil {
		return err
	}

	row := c.db.QueryRowContext(ctx, "SELECT client_id, client_secret_enc, enc_key_id FROM aws_sso_clients")

	var clientId string
	var clientSecretEnc string
	var encKeyId string

	if err := row.Scan(&clientId, &clientSecretEnc, &encKeyId); err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	clientSecret, err := c.encryptionService.Decrypt(clientSecretEnc, encKeyId)

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	tokenRes, err := c.getToken(ctx, input.Region, clientId, clientSecret, input.DeviceCode, input.UserCode)

	if err != nil {
		if errors.Is(err, awssso.ErrDeviceFlowNotAuthorized) {
			ctx.Logger().Debug().Err(err).Msg("failed to get token because user did not authorize device")
			return ErrDeviceAuthFlowNotAuthorized
		}

		if errors.Is(err, awssso.ErrDeviceCodeExpired) {
			ctx.Logger().Debug().Err(err).Msg("failed to get token because user and device code expired")
			return ErrDeviceAuthFlowTimedOut
		}

		ctx.Logger().Error().Err(err).Msg("failed to get token")
		return ErrTransientAwsClientError
	}

	idTokenEnc, keyId, err := c.encryptionService.Encrypt(tokenRes.IdToken)

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	ctx.Logger().Info().Msgf("refreshing access token for instance [%s]", input.InstanceId)

	sql := `UPDATE aws_idc SET
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
		c.clock.NowUnix(),
		tokenRes.ExpiresIn,
		refreshTokenEnc,
		keyId, input.InstanceId)

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	rowsAffected, err := res.RowsAffected()

	if err != nil {
		return errors.Join(err, app.ErrFatal)
	}

	if rowsAffected != 1 {
		return ErrInstanceWasNotFound
	}

	return nil
}

func (c *AwsIdentityCenterController) getOrRegisterClient(ctx app.Context, awsRegion string) (*awssso.RegistrationResponse, error) {
	row := c.db.QueryRowContext(ctx, "SELECT client_id, client_secret_enc, created_at, expires_at, enc_key_id FROM aws_sso_clients")

	var encKeyId string
	var result awssso.RegistrationResponse

	shouldRegisterClient := false

	if err := row.Scan(&result.ClientId, &result.ClientSecret, &result.CreatedAt, &result.ExpiresAt, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			shouldRegisterClient = true
		} else {
			return nil, errors.Join(err, app.ErrFatal)
		}
	}

	if shouldRegisterClient {
		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		ctx.Logger().Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, awssso.AwsRegion(awsRegion), friendlyClientName)
		if err != nil {
			ctx.Logger().Error().Err(err).Msg("failed to register client")
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		if err != nil {
			return nil, errors.Join(err, app.ErrFatal)
		}

		_, err = c.db.ExecContext(ctx, `INSERT INTO aws_sso_clients
			(client_id, client_secret_enc, created_at, expires_at, enc_key_id)
			VALUES (?, ?, ?, ?, ?)`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId)

		if err != nil {
			return nil, errors.Join(err, app.ErrFatal)
		}

		ctx.Logger().Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	if c.clock.NowUnix() > result.ExpiresAt {
		ctx.Logger().Info().Msg("client expired. registering new client")

		friendlyClientName := fmt.Sprintf("swervo_%s", utils.RandomString(6))
		ctx.Logger().Info().Msgf("registering new client [%s]", friendlyClientName)

		output, err := c.awsSsoClient.RegisterClient(ctx, awssso.AwsRegion(awsRegion), friendlyClientName)
		if err != nil {
			ctx.Logger().Error().Err(err).Msg("failed to register client")
			return nil, err
		}

		clientSecretEnc, encKeyId, err := c.encryptionService.Encrypt(output.ClientSecret)

		if err != nil {
			return nil, errors.Join(err, app.ErrFatal)
		}

		_, err = c.db.ExecContext(ctx, `UPDATE aws_sso_clients SET
			client_id = ?,
			client_secret_enc = ?,
			created_at = ?,
			expires_at = ?,
			enc_key_id = ?
			WHERE client_id = ?`,
			output.ClientId, clientSecretEnc, output.CreatedAt, output.ExpiresAt, encKeyId, result.ClientId)

		if err != nil {
			return nil, errors.Join(err, app.ErrFatal)
		}

		ctx.Logger().Info().Msgf("client [%s] registered successfully", friendlyClientName)

		return output, nil
	}

	var err error
	result.ClientSecret, err = c.encryptionService.Decrypt(result.ClientSecret, encKeyId)

	if err != nil {
		return nil, errors.Join(err, app.ErrFatal)
	}

	return &result, nil
}

func (c *AwsIdentityCenterController) authorizeDevice(ctx app.Context, startUrl, region string, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	ctx.Logger().Info().Msg("Authorizing Device")

	output, err := c.awsSsoClient.StartDeviceAuthorization(ctx, awssso.AwsRegion(region), startUrl, clientId, clientSecret)

	if err != nil {
		return nil, err
	}

	ctx.Logger().Info().Msgf("please login at %s?user_code=%s. You have %d seconds to do so", output.VerificationUri, output.UserCode, output.ExpiresIn)

	return output, nil
}

func (c *AwsIdentityCenterController) getToken(ctx app.Context, awsRegion, clientId, clientSecret, deviceCode, userCode string) (*awssso.GetTokenResponse, error) {
	ctx.Logger().Info().Msg("getting access token")
	output, err := c.awsSsoClient.CreateToken(ctx,
		awssso.AwsRegion(awsRegion),
		clientId,
		clientSecret,
		userCode,
		deviceCode,
	)

	if err != nil {
		ctx.Logger().Error().Err(err).Msg("failed to get access token")
		return nil, err
	}

	ctx.Logger().Info().Msg("got access token")

	return output, nil
}
