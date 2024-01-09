package awscredsfile

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/abjrcode/swervo/internal/app"
	"github.com/abjrcode/swervo/internal/utils"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	ErrInvalidCredentialsFile = app.NewValidationError("INVALID_CREDENTIALS_FILE")
	ErrEmptyProfile           = app.NewValidationError("EMPTY_PROFILE")
	ErrEmptyKey               = app.NewValidationError("EMPTY_KEY")
	ErrEmptyKeyValue          = app.NewValidationError("EMPTY_KEY_VALUE")
)

type credentialsFileManager struct {
	filePath string
}

func NewCredentialsFileManager(filePath string) *credentialsFileManager {
	return &credentialsFileManager{
		filePath: filePath,
	}
}

func NewDefaultCredentialsFileManager() *credentialsFileManager {
	return NewCredentialsFileManager(config.DefaultSharedCredentialsFilename())
}

type ProfileCreds struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsSessionToken    *string
}

func (manager *credentialsFileManager) WriteProfileCredentials(profileName string, creds ProfileCreds) error {
	credFilePath := manager.filePath

	fileWasCreated := false

	var credentialsFileHandle *os.File

	if _, err := os.Stat(credFilePath); os.IsNotExist(err) {
		credentialsFileHandle, err = os.Create(credFilePath)
		fileWasCreated = true

		if err != nil {
			return err
		}
	} else {
		credentialsFileHandle, err = os.OpenFile(credFilePath, os.O_RDWR, 0644)
		if err != nil {
			return err
		}
	}

	if fileWasCreated {
		defer credentialsFileHandle.Close()

		var builder strings.Builder

		builder.WriteString(fmt.Sprintf("[%s]\n", profileName))
		builder.WriteString(fmt.Sprintf("aws_access_key_id = %s\n", creds.AwsAccessKeyId))
		builder.WriteString(fmt.Sprintf("aws_secret_access_key = %s\n", creds.AwsSecretAccessKey))

		if creds.AwsSessionToken != nil {
			builder.WriteString(fmt.Sprintf("aws_session_token = %s\n", *creds.AwsSessionToken))
		}

		credentialsProfile := builder.String()

		_, err := credentialsFileHandle.WriteString(credentialsProfile)

		if err != nil {
			return err
		}
	} else {
		credentialsProfile, err := io.ReadAll(credentialsFileHandle)

		if err != nil {
			return err
		}

		parser := newParser(string(credentialsProfile))
		credentials, err := parser.parse()

		if err != nil {
			return err
		}

		index := slices.IndexFunc(credentials, func(cred profileCredentials) bool {
			return cred.Profile == profileName
		})

		if index == -1 {
			credentials = append(credentials, profileCredentials{
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

		credentialsFileName := credentialsFileHandle.Name()

		err = credentialsFileHandle.Close()
		if err != nil {
			return err
		}

		err = utils.SafelyOverwriteFile(credentialsFileName, serializeCredentialsToString(credentials))

		if err != nil {
			return err
		}
	}

	return nil
}

func serializeCredentialsToString(credentials []profileCredentials) string {
	var result strings.Builder

	for _, cred := range credentials {
		result.WriteString(fmt.Sprintf("[%s]\n", cred.Profile))

		result.WriteString(fmt.Sprintf("aws_access_key_id = %s\n", cred.AccessKeyID))

		result.WriteString(fmt.Sprintf("aws_secret_access_key = %s\n", cred.SecretAccessKey))

		if cred.SessionToken != nil {
			result.WriteString(fmt.Sprintf("aws_session_token = %s\n", *cred.SessionToken))
		}

		result.WriteString("\n")
	}

	return result.String()
}
