package awscredssink

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/eventing"
	"github.com/abjrcode/swervo/internal/plumbing"
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	awsidc "github.com/abjrcode/swervo/providers/aws_idc"
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

type AwsCredentialsSinkController struct {
	db                *sql.DB
	bus               *eventing.Eventbus
	encryptionService encryption.EncryptionService
	clock             utils.Clock

	instances map[string]*AwsCredentialsSinkInstance
}

func NewAwsCredentialsSinkController(db *sql.DB, bus *eventing.Eventbus, encryptionService encryption.EncryptionService, clock utils.Clock) *AwsCredentialsSinkController {
	return &AwsCredentialsSinkController{
		db:                db,
		bus:               bus,
		encryptionService: encryptionService,
		clock:             clock,

		instances: make(map[string]*AwsCredentialsSinkInstance),
	}
}

type AwsCredentialsSinkInstance struct {
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

func (c *AwsCredentialsSinkController) GetInstanceData(ctx app.Context, instanceId string) (*AwsCredentialsSinkInstance, error) {
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

	return &AwsCredentialsSinkInstance{
		InstanceId:     instanceId,
		FilePath:       filePath,
		AwsProfileName: awsProfileName,
		Label:          label,
	}, nil
}

func (c *AwsCredentialsSinkController) validateLabel(label string) error {
	if len(label) < 1 || len(label) > 50 {
		return ErrInvalidLabel
	}

	return nil
}

type AwsCredentialsSink_NewInstanceCommandInput struct {
	FilePath       string `json:"filePath"`
	AwsProfileName string `json:"awsProfileName"`
	Label          string `json:"label"`

	ProviderCode string `json:"providerCode"`
	ProviderId   string `json:"providerId"`
}

func (c *AwsCredentialsSinkController) NewInstance(ctx app.Context, input AwsCredentialsSink_NewInstanceCommandInput) (string, error) {
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

	c.instances[instanceId] = &AwsCredentialsSinkInstance{
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

func (c *AwsCredentialsSinkController) SinkCode() string {
	return SinkCode
}

func (c *AwsCredentialsSinkController) ListConnectedSinks(ctx app.Context, providerCode, providerId string) ([]plumbing.SinkInstance, error) {
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

func (c *AwsCredentialsSinkController) DisconnectSink(ctx app.Context, input plumbing.DisconnectSinkCommandInput) error {
	_, err := c.db.ExecContext(ctx, "DELETE FROM aws_credentials_file WHERE instance_id = ?", input.SinkId)

	if err != nil {
		return err
	}

	delete(c.instances, input.SinkId)

	return nil
}

func (c *AwsCredentialsSinkController) FlowData(ctx app.Context, creds awsidc.AwsCredentials, pipeId string) error {
	return nil
}
