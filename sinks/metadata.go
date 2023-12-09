package sinks

type SinkMeta struct {
	Code          string
	Name          string
	IconSvgBase64 string
}

var (
	AwsCredentialsFile = "aws-credentials-file"
)

var (
	SupportedSinks = map[string]SinkMeta{
		AwsCredentialsFile: {
			Code: AwsCredentialsFile,
			Name: "AWS Credentials File",
		},
	}
)
