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
	"github.com/abjrcode/swervo/internal/security/encryption"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/mattn/go-sqlite3"
	"github.com/segmentio/ksuid"
)

var (
	ErrInvalidFilePath           = errors.New("INVALID_FILE_PATH")
	ErrInvalidLabel              = errors.New("INVALID_LABEL")
	ErrInstanceWasNotFound       = errors.New("INSTANCE_WAS_NOT_FOUND")
	ErrInstanceAlreadyRegistered = errors.New("INSTANCE_ALREADY_REGISTERED")
)

type ProfileCreds struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsSessionToken    *string
}

type AwsCredentialsFileController struct {
	db                *sql.DB
	bus               *eventing.Eventbus
	encryptionService encryption.EncryptionService
	clock             utils.Clock
}

func NewAwsCredentialsFileController(db *sql.DB, bus *eventing.Eventbus, encryptionService encryption.EncryptionService, clock utils.Clock) *AwsCredentialsFileController {
	return &AwsCredentialsFileController{
		db:                db,
		bus:               bus,
		encryptionService: encryptionService,
		clock:             clock,
	}
}

func (c *AwsCredentialsFileController) ListInstances(ctx app.Context) ([]string, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT instance_id FROM aws_credentials_file ORDER BY instance_id DESC")

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

type AwsCredentialsFileInstance struct {
	InstanceId string `json:"instanceId"`
	FilePath   string `json:"filePath"`
	Label      string `json:"label"`
}

func (c *AwsCredentialsFileController) GetInstanceData(ctx app.Context, instanceId string) (*AwsCredentialsFileInstance, error) {
	row := c.db.QueryRowContext(ctx, "SELECT file_path, label FROM aws_credentials_file WHERE instance_id = ?", instanceId)

	var filePath string
	var label string

	if err := row.Scan(&filePath, &label); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInstanceWasNotFound
		}

		return nil, errors.Join(err, app.ErrFatal)
	}

	return &AwsCredentialsFileInstance{

		InstanceId: instanceId,
		FilePath:   filePath,
		Label:      label,
	}, nil
}

func (c *AwsCredentialsFileController) validateLabel(label string) error {
	if len(label) < 1 || len(label) > 50 {
		return ErrInvalidLabel
	}

	return nil
}

type AwsCredentialsFile_NewInstanceCommandInput struct {
	FilePath string `json:"filePath"`
	Label    string `json:"label"`
}

func (c *AwsCredentialsFileController) NewInstance(ctx app.Context, input AwsCredentialsFile_NewInstanceCommandInput) (string, error) {
	filePath := filepath.Clean(input.FilePath)

	err := c.validateLabel(input.Label)
	if err != nil {
		return "", err
	}

	nowUnix := c.clock.NowUnix()

	uniqueId, err := ksuid.NewRandomWithTime(time.Unix(nowUnix, 0))
	if err != nil {
		return "", errors.Join(err, app.ErrFatal)
	}

	instanceId := uniqueId.String()
	version := 1

	_, err = c.db.ExecContext(ctx, "INSERT INTO aws_credentials_file (instance_id, version, file_path, label) VALUES (?, ?, ?, ?)", instanceId, version, filePath, input.Label)

	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return "", ErrInstanceAlreadyRegistered
			}
		}
		return "", errors.Join(err, app.ErrFatal)
	}

	return instanceId, nil
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

func (c *AwsCredentialsFileController) WriteProfileCredentials(profileName string, creds ProfileCreds) error {
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
