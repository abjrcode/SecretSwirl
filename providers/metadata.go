package providers

import awsidc "github.com/abjrcode/swervo/providers/aws_idc"

type ProviderMeta struct {
	Code          string
	Name          string
	IconSvgBase64 string
}

var (
	SupportedProviders = map[string]ProviderMeta{
		awsidc.ProviderCode: {
			Code: awsidc.ProviderCode,
			Name: "AWS Identity Center",
		},
	}
)
