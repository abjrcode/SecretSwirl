package providers

type ProviderMeta struct {
	Code          string
	Name          string
	IconSvgBase64 string
}

var (
	AwsIamIdc = "aws-iam-idc"
)

var (
	SupportedProviders = map[string]ProviderMeta{
		AwsIamIdc: {
			Code: AwsIamIdc,
			Name: "AWS IAM IDC",
		},
	}
)
