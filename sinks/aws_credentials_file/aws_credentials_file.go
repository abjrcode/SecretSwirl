package awscredentialsfile

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
)

type ProfileCreds struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	AwsSessionToken    *string
}

type awsCredentialsFile struct {
	filePath string
}

func NewAwsCredentialsFile() awsCredentialsFile {
	return awsCredentialsFile{
		filePath: config.DefaultSharedCredentialsFilename(),
	}
}

func NewAwsCredentialsFileFromPath(filePath string) awsCredentialsFile {
	return awsCredentialsFile{
		filePath: filePath,
	}
}

func (f *awsCredentialsFile) WriteProfileCredentials(profileName string, creds ProfileCreds) error {
	credFilePath := f.filePath

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

		err = safelyOverwriteFile(file.Name(), serializeCredentialsToString(credentials))

		if err != nil {
			return err
		}
	}

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

func safelyOverwriteFile(filePath string, content string) error {
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "swervo_temp")
	if err != nil {
		return err
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(content)
	if err != nil {
		return err
	}

	err = tempFile.Sync()
	if err != nil {
		return err
	}

	err = os.Rename(tempFile.Name(), filePath)
	if err != nil {
		return err
	}

	return nil
}
