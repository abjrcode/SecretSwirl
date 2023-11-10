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
	"github.com/rs/zerolog"
)

type AwsIdentityCenterController struct {
	ctx               context.Context
	logger            *zerolog.Logger
	db                *sql.DB
	encryptionService encryption.EncryptionService
	awsSsoClient      awssso.AwsSsoOidcClient
	timeHelper        utils.Datetime
	syncChan          chan bool
	errChan           chan error
}

func NewAwsIdentityCenterController(db *sql.DB, encryptionService encryption.EncryptionService, awsSsoClient awssso.AwsSsoOidcClient, datetime utils.Datetime) *AwsIdentityCenterController {
	return &AwsIdentityCenterController{
		db:                db,
		encryptionService: encryptionService,
		awsSsoClient:      awsSsoClient,
		timeHelper:        datetime,
		syncChan:          make(chan bool),
		errChan:           make(chan error),
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
	Enabled  bool                       `json:"enabled"`
	Accounts []AwsIdentityCenterAccount `json:"accounts"`
}

func (c *AwsIdentityCenterController) GetInstanceData(startUrl string) (AwsIdentityCenterCardData, error) {
	row := c.db.QueryRowContext(c.ctx, "SELECT region, access_token_enc, access_token_created_at, access_token_expires_in, enc_key_id FROM aws_iam_idc WHERE start_url = ?", startUrl)

	var region string
	var accessTokenEnc string
	var accessTokenCreatedAt int64
	var accessTokenExpiresIn int64
	var encKeyId string

	if err := row.Scan(&region, &accessTokenEnc, &accessTokenCreatedAt, &accessTokenExpiresIn, &encKeyId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.logger.Debug().Msgf("No token found for start URL [%s]", startUrl)
			return AwsIdentityCenterCardData{
				Enabled:  false,
				Accounts: []AwsIdentityCenterAccount{},
			}, nil
		}

		return AwsIdentityCenterCardData{
			Enabled:  false,
			Accounts: []AwsIdentityCenterAccount{},
		}, err
	}

	now := c.timeHelper.NowUnix()
	if now > accessTokenCreatedAt+accessTokenExpiresIn {
		c.logger.Debug().Msgf("Token for start URL [%s] is expired", startUrl)
		return AwsIdentityCenterCardData{}, errors.New("access token expired")
	}

	accessToken, err := c.encryptionService.Decrypt(accessTokenEnc, encKeyId)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to decrypt access token")
		return AwsIdentityCenterCardData{
			Enabled:  false,
			Accounts: []AwsIdentityCenterAccount{},
		}, err
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), region)
	accountsOut, err := c.awsSsoClient.ListAccounts(ctx, accessToken)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to list accounts")
		return AwsIdentityCenterCardData{
			Enabled:  false,
			Accounts: []AwsIdentityCenterAccount{},
		}, err
	}

	accounts := make([]AwsIdentityCenterAccount, 0)

	for _, account := range accountsOut.Accounts {
		accounts = append(accounts, AwsIdentityCenterAccount{
			AccountId:   account.AccountId,
			AccountName: account.AccountName,
		})
	}

	return AwsIdentityCenterCardData{
		Enabled:  true,
		Accounts: accounts,
	}, nil
}

func (c *AwsIdentityCenterController) Setup(startUrlStr, awsRegion string) (string, error) {
	if _, ok := awssso.SupportedAwsRegions[awsRegion]; !ok {
		c.logger.Error().Msgf("Unsupported AWS region [%s]", awsRegion)
		return "", errors.New("unsupported AWS region. Supported regions are: " + fmt.Sprintf("%v", awssso.SupportedAwsRegions))
	}

	ctx := context.WithValue(c.ctx, awssso.AwsRegion("awsRegion"), awsRegion)
	startUrl, err := url.Parse(startUrlStr)

	if startUrl.Scheme == "" || startUrl.Host == "" {
		c.logger.Error().Msgf("Invalid start URL [%s]", startUrlStr)
		return "", errors.New("invalid start URL")
	}

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to parse start URL")
		return "", err
	}

	regRes, err := c.registerClient(ctx, startUrl, awsRegion)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to register client")
		return "", err
	}

	authorizeRes, err := c.authorizeDevice(ctx, startUrl, regRes.ClientId, regRes.ClientSecret)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to authorize device")
		return "", err
	}

	go func(syncChan chan bool, errChan chan error) {
		select {
		case <-time.After(time.Duration(authorizeRes.ExpiresIn) * time.Second):
			c.logger.Error().Msg("Timeout waiting for user to login")
			return
		case <-c.syncChan:
		}

		clientId := regRes.ClientId
		clientSecret := regRes.ClientSecret
		expiresAt := regRes.ExpiresAt
		userCode := authorizeRes.UserCode
		deviceCode := authorizeRes.DeviceCode

		tokenRes, err := c.getToken(ctx, clientId, clientSecret, deviceCode, userCode)

		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to get token")
			errChan <- err
			return
		}

		tx, err := c.db.BeginTx(ctx, nil)

		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to start transaction")
			errChan <- err
			return
		}

		defer tx.Rollback()

		clientSecretEnc, keyId, err := c.encryptionService.Encrypt(clientSecret)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to encrypt client secret")
			errChan <- err
			return
		}

		idTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.IdToken)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to encrypt id token")
			errChan <- err
			return
		}

		accessTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.AccessToken)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to encrypt access token")
			errChan <- err
			return
		}

		refreshTokenEnc, _, err := c.encryptionService.Encrypt(tokenRes.RefreshToken)
		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to encrypt refresh token")
			errChan <- err
			return
		}

		sql := `INSERT INTO aws_iam_idc
	  (start_url, enabled, region, client_id, client_secret_enc, created_at, expires_at, id_token_enc,
			access_token_enc, token_type, access_token_created_at, access_token_expires_in, refresh_token_enc, enc_key_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		_, err = tx.ExecContext(ctx, sql,
			startUrl.String(),
			true,
			awsRegion,
			clientId,
			clientSecretEnc,
			c.timeHelper.NowUnix(),
			expiresAt,
			idTokenEnc,
			accessTokenEnc,
			tokenRes.TokenType,
			c.timeHelper.NowUnix(),
			tokenRes.ExpiresIn,
			refreshTokenEnc,
			keyId)

		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to create token")
			errChan <- err
			return
		}

		_, err = tx.ExecContext(ctx, `INSERT INTO providers (code, instance_id, display_name, is_favorite) VALUES (?, ?, ?, ?) `, "aws-iam-idc", startUrl.String(), "AWS IAM IDC", true)

		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to add provider to list of configured providers")
			errChan <- err
			return
		}

		err = tx.Commit()

		if err != nil {
			c.logger.Error().Err(err).Msg("Failed to commit transaction")
			errChan <- err
			return
		}

		errChan <- nil
	}(c.syncChan, c.errChan)

	return fmt.Sprintf("%s?user_code=%s&expires_in=%d", authorizeRes.VerificationUri, authorizeRes.UserCode, authorizeRes.ExpiresIn), nil
}

func (c *AwsIdentityCenterController) FinalizeSetup(timeoutSec uint8) error {
	c.logger.Debug().Msgf("attempting to finalize setup with a timeout of [%d] seconds", timeoutSec)
	select {
	case c.syncChan <- true:
		{
			c.logger.Debug().Msg("sync signal sent")
			select {
			case <-time.After(time.Duration(timeoutSec) * time.Second):
				return errors.New("timeout waiting for setup to finish")
			case err := <-c.errChan:
				{
					if err != nil {
						c.logger.Error().Msgf("finalizing setup failed with: %s", err)
						return err
					}
				}
			}
		}
	case <-time.After(time.Duration(timeoutSec) * time.Second):
		c.logger.Debug().Msg("timed out waiting for sync signal")
		close(c.syncChan)
		c.syncChan = make(chan bool)
		return errors.New("timeout waiting for receiver to finish")
	}

	return nil
}

func (c *AwsIdentityCenterController) registerClient(ctx context.Context, startUrl *url.URL, awsRegion string) (*awssso.RegistrationResponse, error) {
	row := c.db.QueryRowContext(ctx, "SELECT client_id, created_at, expires_at FROM aws_iam_idc WHERE start_url = ?", startUrl.String())

	var result awssso.RegistrationResponse

	if err := row.Scan(&result.ClientId, &result.CreatedAt, &result.ExpiresAt); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		c.logger.Error().Msgf("Client [%s] aws already registered at [%d]", startUrl, result.CreatedAt)
		return nil, errors.New("client already registered")
	}

	friendlyClientName := fmt.Sprintf("swervo_%s_%s", startUrl.Hostname(), utils.RandomString(2))
	c.logger.Info().Msgf("Registering client [%s] for start URL [%s]", friendlyClientName, startUrl)

	output, err := c.awsSsoClient.RegisterClient(ctx, friendlyClientName)

	if err != nil {
		return nil, err
	}

	c.logger.Info().Msgf("Client [%s] registered for Start URL: [%s]", friendlyClientName, startUrl)

	return output, nil
}

func (c *AwsIdentityCenterController) authorizeDevice(ctx context.Context, startUrl *url.URL, clientId, clientSecret string) (*awssso.AuthorizationResponse, error) {
	c.logger.Info().Msg("Authorizing Device")
	output, err := c.awsSsoClient.StartDeviceAuthorization(ctx, *startUrl, clientId, clientSecret)

	if err != nil {
		return nil, err
	}

	c.logger.Info().Msgf("Please login at %s?user_code=%s. You have %d seconds to do so", output.VerificationUri, output.UserCode, output.ExpiresIn)

	return output, nil
}

func (c *AwsIdentityCenterController) getToken(ctx context.Context, clientId, clientSecret, deviceCode, userCode string) (*awssso.GetTokenResponse, error) {
	c.logger.Info().Msg("Getting Access Token")
	output, err := c.awsSsoClient.CreateToken(ctx,
		clientId,
		clientSecret,
		userCode,
		deviceCode,
	)

	if err != nil {
		c.logger.Error().Err(err).Msg("Failed to get access token")
		return nil, err
	}

	c.logger.Info().Msg("Got access token")

	return output, nil
}
