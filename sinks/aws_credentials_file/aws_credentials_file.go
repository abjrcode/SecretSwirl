package awscredentialsfile

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/eventing"
	"github.com/abjrcode/swervo/internal/plumbing"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	awsidc "github.com/abjrcode/swervo/providers/aws_idc"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/segmentio/ksuid"
)

var SinkCode = "aws-credentials-file"

var (
	ErrInvalidAwsProfileName     = app.NewValidationError("INVALID_AWS_PROFILE_NAME")
	ErrInvalidLabel              = app.NewValidationError("INVALID_LABEL")
	ErrInvalidProviderCode       = app.NewValidationError("INVALID_PROVIDER_CODE")
	ErrInvalidProviderId         = app.NewValidationError("INVALID_PROVIDER_ID")
	ErrInstanceWasNotFound       = app.NewValidationError("INSTANCE_WAS_NOT_FOUND")
	ErrInstanceAlreadyRegistered = app.NewValidationError("INSTANCE_ALREADY_REGISTERED")
)

type ProfileCreds struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsSessionToken    *string
}

type AwsCredentialsFileSinkController struct {
	db                *sql.DB
	bus               *eventing.Eventbus
	encryptionService encryption.EncryptionService
	clock             utils.Clock

	instances map[string]*AwsCredentialsFileInstance
}

func NewAwsCredentialsFileSinkController(db *sql.DB, bus *eventing.Eventbus, encryptionService encryption.EncryptionService, clock utils.Clock) *AwsCredentialsFileSinkController {
	return &AwsCredentialsFileSinkController{
		db:                db,
		bus:               bus,
		encryptionService: encryptionService,
		clock:             clock,

		instances: make(map[string]*AwsCredentialsFileInstance),
	}
}

type AwsCredentialsFileInstance struct {
	InstanceId     string `json:"instanceId"`
	Version        int    `json:"version"`
	FilePath       string `json:"filePath"`
	AwsProfileName string `json:"awsProfileName"`
	Label          string `json:"label"`
	ProviderCode   string `json:"providerCode"`
	ProviderId     string `json:"providerId"`
	CreatedAt      int64  `json:"createdAt"`
	LastDrainedAt  *int64 `json:"lastDrainedAt"`
}

func (c *AwsCredentialsFileSinkController) GetInstanceData(ctx app.Context, instanceId string) (*AwsCredentialsFileInstance, error) {
	row := c.db.QueryRowContext(ctx, "SELECT file_path, aws_profile_name, label FROM aws_credentials_file WHERE instance_id = ?", instanceId)

	var filePath string
	var awsProfileName string
	var label string

	if err := row.Scan(&filePath, &awsProfileName, &label); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceWasNotFound
		}

		return nil, errors.Join(err, app.ErrFatal)
	}

	return &AwsCredentialsFileInstance{
		InstanceId:     instanceId,
		FilePath:       filePath,
		AwsProfileName: awsProfileName,
		Label:          label,
	}, nil
}

func (c *AwsCredentialsFileSinkController) validateLabel(label string) error {
	if len(label) < 1 || len(label) > 50 {
		return ErrInvalidLabel
	}

	return nil
}

type AwsCredentialsFile_NewInstanceCommandInput struct {
	FilePath       string `json:"filePath"`
	AwsProfileName string `json:"awsProfileName"`
	Label          string `json:"label"`

	ProviderCode string `json:"providerCode"`
	ProviderId   string `json:"providerId"`
}

func (c *AwsCredentialsFileSinkController) NewInstance(ctx app.Context, input AwsCredentialsFile_NewInstanceCommandInput) (string, error) {
	filePath := filepath.Clean(input.FilePath)

	err := c.validateLabel(input.Label)
	if err != nil {
		return "", err
	}

	awsProfileName := strings.Trim(input.AwsProfileName, " ")

	if len(awsProfileName) < 1 || len(awsProfileName) > 50 {
		return "", ErrInvalidAwsProfileName
	}

	if len(input.ProviderCode) < 1 {
		return "", ErrInvalidProviderCode
	}

	if len(input.ProviderId) < 1 {
		return "", ErrInvalidProviderId
	}

	nowUnix := c.clock.NowUnix()

	uniqueId, err := ksuid.NewRandomWithTime(time.Unix(nowUnix, 0))
	if err != nil {
		return "", err
	}

	instanceId := uniqueId.String()
	version := 1

	_, err = c.db.ExecContext(ctx,
		"INSERT INTO aws_credentials_file (instance_id, version, file_path, aws_profile_name, label, provider_code, provider_id, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		instanceId, version, filePath, awsProfileName, input.Label, input.ProviderCode, input.ProviderId, nowUnix)

	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	c.instances[instanceId] = &AwsCredentialsFileInstance{
		InstanceId:     instanceId,
		Version:        version,
		FilePath:       filePath,
		AwsProfileName: awsProfileName,
		Label:          input.Label,
		ProviderCode:   input.ProviderCode,
		ProviderId:     input.ProviderId,
		CreatedAt:      nowUnix,
	}

	return instanceId, nil
}

func (c *AwsCredentialsFileSinkController) SinkCode() string {
	return SinkCode
}

func (c *AwsCredentialsFileSinkController) ListConnectedSinks(ctx app.Context, providerCode, providerId string) ([]plumbing.SinkInstance, error) {
	pipes := make([]plumbing.SinkInstance, 0)

	for _, instance := range c.instances {
		if instance.ProviderCode == providerCode && instance.ProviderId == providerId {
			pipes = append(pipes, plumbing.SinkInstance{
				SinkCode: SinkCode,
				SinkId:   instance.InstanceId,
			})
		}
	}

	return pipes, nil
}

func (c *AwsCredentialsFileSinkController) DisconnectSink(ctx app.Context, input plumbing.DisconnectSinkCommandInput) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM aws_credentials_file WHERE instance_id = ?", input.SinkId)

	if err != nil {
		return err
	}

	delete(c.instances, input.SinkId)

	return nil
}

func (c *AwsCredentialsFileSinkController) FlowData(ctx app.Context, creds awsidc.AwsCredentials, pipeId string) error {
	return nil
}

func serializeCredentialsToString(credentials []credential) string {
	var result strings.Builder

	for _, cred := range credentials {
		result.WriteString(fmt.Sprintf("[%s]\n", cred.Profile))

		result.WriteString(fmt.Sprintf("aws_access_key_id = %s\n", cred.AccessKeyID))

		result.WriteString(fmt.Sprintf("aws_secret_access_key = %s\n", cred.SecretAccessKey))

		if cred.SessionToken != nil {
			result.WriteString(fmt.Sprintf("aws_session_token = %s\n", *cred.SessionToken))
		}

		if cred.Region != nil {
			result.WriteString(fmt.Sprintf("region = %s\n", *cred.Region))
		}

		result.WriteString("\n")
	}

	return result.String()
}

func (c *AwsCredentialsFileSinkController) WriteProfileCredentials(profileName string, creds ProfileCreds) error {
	credFilePath := config.DefaultSharedCredentialsFilename()

	fileWasCreated := false

	var file *os.File

	if _, err := os.Stat(credFilePath); os.IsNotExist(err) {
		file, err = os.Create(credFilePath)
		fileWasCreated = true

		if err != nil {
			return err
		}
	} else {
		file, err = os.OpenFile(credFilePath, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
	}

	defer file.Close()

	if fileWasCreated {
		var builder strings.Builder

		builder.WriteString(fmt.Sprintf("[%s]\n", profileName))
		builder.WriteString(fmt.Sprintf("aws_access_key_id = %s\n", creds.AwsAccessKeyId))
		builder.WriteString(fmt.Sprintf("aws_secret_access_key = %s\n", creds.AwsSecretAccessKey))

		if creds.AwsSessionToken != nil {
			builder.WriteString(fmt.Sprintf("aws_session_token = %s\n", *creds.AwsSessionToken))
		}

		credentialsProfile := builder.String()

		_, err := file.WriteString(credentialsProfile)

		if err != nil {
			return err
		}
	} else {
		credentialsProfile, err := io.ReadAll(file)

		if err != nil {
			return err
		}

		parser := newParser(string(credentialsProfile))
		credentials, err := parser.parse()

		if err != nil {
			return err
		}

		index := slices.IndexFunc(credentials, func(cred credential) bool {
			return cred.Profile == profileName
		})

		if index == -1 {
			credentials = append(credentials, credential{
				Profile:         profileName,
				AccessKeyID:     creds.AwsAccessKeyId,
				SecretAccessKey: creds.AwsSecretAccessKey,
				SessionToken:    creds.AwsSessionToken,
			})
		} else {
			credentials[index].AccessKeyID = creds.AwsAccessKeyId
			credentials[index].SecretAccessKey = creds.AwsSecretAccessKey
			credentials[index].SessionToken = creds.AwsSessionToken
		}

		if err != nil {
			return err
		}

		err = file.Close()

		if err != nil {
			return err
		}

		err = utils.SafelyOverwriteFile(file.Name(), serializeCredentialsToString(credentials))

		if err != nil {
			return err
		}
	}

	return nil
}
