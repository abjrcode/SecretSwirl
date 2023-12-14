package awscredentialsfile

import (
	"path/filepath"
	"testing"

	"github.com/abjrcode/swervo/internal/eventing"
	"github.com/abjrcode/swervo/internal/migrations"
	"github.com/abjrcode/swervo/internal/plumbing"
	"github.com/abjrcode/swervo/internal/security/vault"
	"github.com/abjrcode/swervo/internal/testhelpers"
	"github.com/stretchr/testify/require"
)

func TestNewInstance(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	credFile := NewAwsCredentialsFileSinkController(db, bus, encryptionService, mockClock)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credentials")
	mockClock.On("NowUnix").Return(1)

	instanceId, err := credFile.NewInstance(ctx, AwsCredentialsFile_NewInstanceCommandInput{
		FilePath:       filePath,
		AwsProfileName: "test-profile",
		Label:          "default",
		ProviderCode:   "some-provider-code",
		ProviderId:     "some-provider-id",
	})
	require.NoError(t, err)

	instance, err := credFile.GetInstanceData(ctx, instanceId)
	require.NoError(t, err)

	require.Equal(t, &AwsCredentialsFileInstance{
		InstanceId:     instanceId,
		FilePath:       filePath,
		AwsProfileName: "test-profile",
		Label:          "default",
	}, instance)
}

func Test_DeleteInstance(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	credFile := NewAwsCredentialsFileSinkController(db, bus, encryptionService, mockClock)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credentials")
	mockClock.On("NowUnix").Return(1)

	instanceId, err := credFile.NewInstance(ctx, AwsCredentialsFile_NewInstanceCommandInput{
		FilePath:       filePath,
		AwsProfileName: "default",
		Label:          "default",
		ProviderCode:   "some-provider-code",
		ProviderId:     "some-provider-id",
	})
	require.NoError(t, err)

	err = credFile.DisconnectSink(ctx, plumbing.DisconnectSinkCommandInput{
		SinkCode: credFile.SinkCode(),
		SinkId:   instanceId,
	})
	require.NoError(t, err)

	_, err = credFile.GetInstanceData(ctx, instanceId)
	require.ErrorIs(t, err, ErrInstanceWasNotFound)
}

func Test_ListConnectedSinks(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	credFile := NewAwsCredentialsFileSinkController(db, bus, encryptionService, mockClock)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credentials")
	mockClock.On("NowUnix").Return(1)

	instanceId, err := credFile.NewInstance(ctx, AwsCredentialsFile_NewInstanceCommandInput{
		FilePath:       filePath,
		AwsProfileName: "default",
		Label:          "default",
		ProviderCode:   "some-provider-code",
		ProviderId:     "some-provider-id",
	})
	require.NoError(t, err)

	instances, err := credFile.ListConnectedSinks(ctx, "some-provider-code", "some-provider-id")
	require.NoError(t, err)

	require.Equal(t, []plumbing.SinkInstance{{
		SinkCode: SinkCode,
		SinkId:   instanceId,
	}}, instances)
}

func Test_ListConnectedSinks_WhenNoneExist(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	credFile := NewAwsCredentialsFileSinkController(db, bus, encryptionService, mockClock)

	instances, err := credFile.ListConnectedSinks(ctx, "some-provider-code", "some-provider-id")
	require.NoError(t, err)

	require.Equal(t, []plumbing.SinkInstance{}, instances)
}

/*
func TestWriteProfileCredentials(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	clock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, clock)
	encryptionService := vault.NewVault(db, bus, clock)

	dirPath := t.TempDir()
	filePath := filepath.Join(dirPath, "credentials")
	credFile := NewAwsCredentialsFileControllerFromPath(db, bus, encryptionService, clock)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    utils.AddressOf("test-session-token"),
	}

	err = credFile.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err := os.Open(filePath)
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, credential{
		Profile:         "test-profile",
		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    utils.AddressOf("test-session-token"),
		Region:          nil,
	}, credentials[0])
}

func TestWriteProfileCredentials_AlongExistingProfile(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	clock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, clock)
	encryptionService := vault.NewVault(db, bus, clock)

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

	credFile := NewAwsCredentialsFileControllerFromPath(filePath, db, bus, encryptionService, clock)

	creds := ProfileCreds{
		AwsAccessKeyId:     "new-access-key-id",
		AwsSecretAccessKey: "new-secret-access-key",
	}

	err = credFile.WriteProfileCredentials("new-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)
	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 2, len(credentials))
	require.Equal(t, []credential{{
		Profile:         "default",
		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    nil,
		Region:          utils.AddressOf("test-region"),
	}, {
		Profile:         "new-profile",
		AccessKeyID:     "new-access-key-id",
		SecretAccessKey: "new-secret-access-key",
		SessionToken:    nil,
		Region:          nil,
	}}, credentials)
}

func TestWriteProfileCredentials_OverrideExistingProfile(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	clock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, clock)
	encryptionService := vault.NewVault(db, bus, clock)

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

	credFile := NewAwsCredentialsFileControllerFromPath(filePath, db, bus, encryptionService, clock)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    utils.AddressOf("test-session-token"),
	}

	err = credFile.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, credential{
		Profile:         "test-profile",
		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    utils.AddressOf("test-session-token"),
		Region:          nil,
	}, credentials[0])
}

func TestWriteProfileCredentials_OverrideWithoutSessionToken(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	clock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, clock)
	encryptionService := vault.NewVault(db, bus, clock)

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

	credFile := NewAwsCredentialsFileControllerFromPath(filePath, db, bus, encryptionService, clock)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    nil,
	}

	err = credFile.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, credential{
		Profile:         "test-profile",
		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    nil,
		Region:          nil,
	}, credentials[0])
}

func TestWriteProfileCredentials_OverrideWithExistingRegion(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_file_tests")
	require.NoError(t, err)
	clock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, clock)
	encryptionService := vault.NewVault(db, bus, clock)

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

	credFile := NewAwsCredentialsFileControllerFromPath(filePath, db, bus, encryptionService, clock)

	creds := ProfileCreds{
		AwsAccessKeyId:     "test-access-key-id",
		AwsSecretAccessKey: "test-secret-access-key",
		AwsSessionToken:    nil,
	}

	err = credFile.WriteProfileCredentials("test-profile", creds)
	require.NoError(t, err)

	file, err = os.Open(filePath)
	require.NoError(t, err)

	token, err := io.ReadAll(file)
	require.NoError(t, err)

	parser := newParser(string(token))
	credentials, err := parser.parse()
	require.NoError(t, err)

	require.Equal(t, 1, len(credentials))
	require.Equal(t, credential{
		Profile:         "test-profile",
		AccessKeyID:     "test-access-key-id",
		SecretAccessKey: "test-secret-access-key",
		SessionToken:    nil,
		Region:          utils.AddressOf("test-region"),
	}, credentials[0])
}
*/
