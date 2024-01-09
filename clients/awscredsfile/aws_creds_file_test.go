package awscredsfile

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/abjrcode/swervo/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestWriteProfileCredentials(t *testing.T) {
	dirPath := t.TempDir()
	filePath := filepath.Join(dirPath, "credentials")

	manager := NewCredentialsFileManager(filePath)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    utils.AddressOf("test-session-token"),
	}

	err := manager.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err := os.Open(filePath)
	t.Cleanup(func() {
		file.Close()
	})
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, profileCredentials{
		Profile: "test-profile",

		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    utils.AddressOf("test-session-token"),
	}, credentials[0])
}

func TestWriteProfileCredentials_AlongExistingProfile(t *testing.T) {
	dirPath := t.TempDir()
	filePath := filepath.Join(dirPath, "credentials")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)

	_, err = file.WriteString(`
		[default]
		aws_access_key_id = test-access-key-id
		aws_secret_access_key = test-secret-access-key
		region = test-region
	`)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	manager := NewCredentialsFileManager(filePath)

	creds := ProfileCreds{
		AwsAccessKeyId:     "new-access-key-id",
		AwsSecretAccessKey: "new-secret-access-key",
	}

	err = manager.WriteProfileCredentials("new-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	t.Cleanup(func() {
		file.Close()
	})
	require.NoError(t, err)
	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 2, len(credentials))
	require.Equal(t, []profileCredentials{{
		Profile:         "default",
		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    nil,
	}, {
		Profile:         "new-profile",
		AccessKeyID:     "new-access-key-id",
		SecretAccessKey: "new-secret-access-key",
		SessionToken:    nil,
	}}, credentials)
}

func TestWriteProfileCredentials_OverrideExistingProfile(t *testing.T) {
	dirPath := t.TempDir()
	filePath := filepath.Join(dirPath, "credentials")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)

	_, err = file.WriteString(`
		[test-profile]
		aws_access_key_id = test-access-key-id
		aws_secret_access_key = test-secret-access-key
	`)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	manager := NewCredentialsFileManager(filePath)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    utils.AddressOf("test-session-token"),
	}

	err = manager.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	t.Cleanup(func() {
		file.Close()
	})
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, profileCredentials{
		Profile: "test-profile",

		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    utils.AddressOf("test-session-token"),
	}, credentials[0])
}

func TestWriteProfileCredentials_OverrideWithoutSessionToken(t *testing.T) {
	dirPath := t.TempDir()
	filePath := filepath.Join(dirPath, "credentials")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)

	_, err = file.WriteString(`
		[test-profile]
		aws_access_key_id = test-access-key-id
		aws_secret_access_key = test-secret-access-key
		aws_session_token = test-session-token
	`)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	manager := NewCredentialsFileManager(filePath)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    nil,
	}

	err = manager.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	t.Cleanup(func() {
		file.Close()
	})
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, profileCredentials{
		Profile: "test-profile",

		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    nil,
	}, credentials[0])
}

func TestWriteProfileCredentials_OverrideWithExistingRegion(t *testing.T) {
	dirPath := t.TempDir()
	filePath := filepath.Join(dirPath, "credentials")

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	require.NoError(t, err)

	_, err = file.WriteString(`
		[test-profile]
		aws_access_key_id = test-access-key-id
		aws_secret_access_key = test-secret-access-key
		region = test-region
	`)
	require.NoError(t, err)
	err = file.Close()
	require.NoError(t, err)

	manager := NewCredentialsFileManager(filePath)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    nil,
	}

	err = manager.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	t.Cleanup(func() {
		file.Close()
	})
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, profileCredentials{
		Profile: "test-profile",

		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    nil,
	}, credentials[0])
}
