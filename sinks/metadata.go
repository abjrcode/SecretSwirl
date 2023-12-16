package sinks

import awscredssink "github.com/abjrcode/swervo/sinks/awscredssink"

type SinkMeta struct {
	Code          string
	Name          string
	IconSvgBase64 string
}

var (
	SupportedSinks = map[string]SinkMeta{
		awscredssink.SinkCode: {
			Code: awscredssink.SinkCode,
			Name: "AWS Credentials File",
		},
	}
)
