package awscredentialsfile

import (
	"testing"

	"github.com/abjrcode/swervo/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	fileContent := `
		[default]
		aws_access_key_id = YOUR_ACCESS_KEY
		aws_secret_access_key = YOUR_SECRET_KEY
		region = eu-west-1

		[profile_name]
		aws_access_key_id = ANOTHER_ACCESS_KEY
		aws_secret_access_key = ANOTHER_SECRET_KEY
		aws_session_token = ANOTHER_SESSION_TOKEN

		[yet_another_profile]
		aws_access_key_id = YET_ANOTHER_ACCESS_KEY
		aws_secret_access_key = YET_ANOTHER_SECRET_KEY
		aws_session_token = YET_ANOTHER_SESSION_TOKEN
		region = us-east-1
	`

	parser := newParser(fileContent)
	credentials, err := parser.parse()

	require.NoError(t, err)

	require.Equal(t, 3, len(credentials))

	require.Equal(t, credentials[0], credential{
		Profile:         "default",
		AccessKeyID:     "YOUR_ACCESS_KEY",
		SecretAccessKey: "YOUR_SECRET_KEY",
		SessionToken:    nil,
		Region:          utils.AddressOf("eu-west-1"),
	}, credential{
		Profile:         "profile_name",
		AccessKeyID:     "ANOTHER_ACCESS_KEY",
		SecretAccessKey: "ANOTHER_SECRET_KEY",
		SessionToken:    utils.AddressOf("ANOTHER_SESSION_TOKEN"),
		Region:          nil,
	},
		credential{
			Profile:         "yet_another_profile",
			AccessKeyID:     "yet_another_access_key",
			SecretAccessKey: "yet_another_secret_key",
			SessionToken:    utils.AddressOf("yet_another_session_token"),
			Region:          utils.AddressOf("us-east-1"),
		},
	)
}

func TestParse_Error_EmptyProfile(t *testing.T) {
	fileContent := `
		[]
		aws_access_key_id = YOUR_ACCESS_KEY
		aws_secret_access_key = YOUR_SECRET_KEY
		region = eu-west-1
	`

	parser := newParser(fileContent)
	_, err := parser.parse()

	require.Error(t, err, ErrEmptyProfile)
}

func TestParse_Error_EmptyKey(t *testing.T) {
	fileContent := `
		[default]
		= YOUR_ACCESS_KEY
		aws_secret_access_key = YOUR_SECRET_KEY
		region = eu-west-1
	`

	parser := newParser(fileContent)
	_, err := parser.parse()

	require.Error(t, err, ErrEmptyKey)
}

func TestParse_Error_EmptyKeyValue(t *testing.T) {
	fileContent := `
		[default]
		aws_access_key_id = 
		aws_secret_access_key = YOUR_SECRET_KEY
		region = eu-west-1
	`

	parser := newParser(fileContent)
	_, err := parser.parse()

	require.Error(t, err, ErrEmptyKeyValue)
}
