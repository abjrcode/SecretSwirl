package awscredssink

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
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_sink_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	controller := NewAwsCredentialsSinkController(db, bus, encryptionService, mockClock)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credentials")
	mockClock.On("NowUnix").Return(1)

	instanceId, err := controller.NewInstance(ctx, AwsCredentialsSink_NewInstanceCommandInput{
		FilePath:       filePath,
		AwsProfileName: "test-profile",
		Label:          "default",
		ProviderCode:   "some-provider-code",
		ProviderId:     "some-provider-id",
	})
	require.NoError(t, err)

	instance, err := controller.GetInstanceData(ctx, instanceId)
	require.NoError(t, err)

	require.Equal(t, &AwsCredentialsSinkInstance{
		InstanceId:     instanceId,
		FilePath:       filePath,
		AwsProfileName: "test-profile",
		Label:          "default",
	}, instance)
}

func Test_DeleteInstance(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_sink_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	controller := NewAwsCredentialsSinkController(db, bus, encryptionService, mockClock)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credentials")
	mockClock.On("NowUnix").Return(1)

	instanceId, err := controller.NewInstance(ctx, AwsCredentialsSink_NewInstanceCommandInput{
		FilePath:       filePath,
		AwsProfileName: "default",
		Label:          "default",
		ProviderCode:   "some-provider-code",
		ProviderId:     "some-provider-id",
	})
	require.NoError(t, err)

	err = controller.DisconnectSink(ctx, plumbing.DisconnectSinkCommandInput{
		SinkCode: controller.SinkCode(),
		SinkId:   instanceId,
	})
	require.NoError(t, err)

	_, err = controller.GetInstanceData(ctx, instanceId)
	require.ErrorIs(t, err, ErrInstanceWasNotFound)
}

func Test_ListConnectedSinks(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_sink_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	controller := NewAwsCredentialsSinkController(db, bus, encryptionService, mockClock)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "credentials")
	mockClock.On("NowUnix").Return(1)

	instanceId, err := controller.NewInstance(ctx, AwsCredentialsSink_NewInstanceCommandInput{
		FilePath:       filePath,
		AwsProfileName: "default",
		Label:          "default",
		ProviderCode:   "some-provider-code",
		ProviderId:     "some-provider-id",
	})
	require.NoError(t, err)

	instances, err := controller.ListConnectedSinks(ctx, "some-provider-code", "some-provider-id")
	require.NoError(t, err)

	require.Equal(t, []plumbing.SinkInstance{{
		SinkCode: SinkCode,
		SinkId:   instanceId,
	}}, instances)
}

func Test_ListConnectedSinks_WhenNoneExist(t *testing.T) {
	db, err := migrations.NewInMemoryMigratedDatabase(t, "aws_credentials_sink_tests")
	require.NoError(t, err)
	mockClock := testhelpers.NewMockClock()
	bus := eventing.NewEventbus(db, mockClock)
	encryptionService := vault.NewVault(db, bus, mockClock)

	ctx := testhelpers.NewMockAppContext()

	controller := NewAwsCredentialsSinkController(db, bus, encryptionService, mockClock)

	instances, err := controller.ListConnectedSinks(ctx, "some-provider-code", "some-provider-id")
	require.NoError(t, err)

	require.Equal(t, []plumbing.SinkInstance{}, instances)
}
