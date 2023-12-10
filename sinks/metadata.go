package sinks

import awscredentialsfile "github.com/abjrcode/swervo/sinks/aws_credentials_file"

type SinkMeta struct {
	Code          string
	Name          string
	IconSvgBase64 string
}

var (
	SupportedSinks = map[string]SinkMeta{
		awscredentialsfile.SinkCode: {
			Code: awscredentialsfile.SinkCode,
			Name: "AWS Credentials File",
		},
	}
)
