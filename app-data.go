package main

type AwsIdentityCenter struct {
	ClientId                    string `toml:"client_id"`
	ClientSecret                string `toml:"client_secret"`
	ExpiresAt                   int64  `toml:"expires_at"`
	ClientIdIssuedAt            int64  `toml:"client_id_issued_at"`
	DeviceCode                  string `toml:"device_code,omitempty"`
	AccessToken                 string `toml:"access_token,omitempty"`
	TokenType                   string `toml:"token_type,omitempty"`
	AccessTokenExpiresInSeconds int64  `toml:"access_token_expires_in_seconds,omitempty"`
	RefreshToken                string `toml:"refresh_token,omitempty"`
}

type AppData struct {
	AwsIdc AwsIdentityCenter `toml:"aws_identity_center,omitempty"`
}
