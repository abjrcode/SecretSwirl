package providers

type ProviderMeta struct {
	Code          string
	Name          string
	IconSvgBase64 string
}

var (
	AwsIdc = "aws-idc"
)

var (
	SupportedProviders = map[string]ProviderMeta{
		AwsIdc: {
			Code: AwsIdc,
			Name: "AWS Identity Center",
		},
	}
)
