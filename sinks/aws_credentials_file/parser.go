package awscredentialsfile

import (
	"bufio"
	"errors"
	"strings"
)

var (
	ErrInvalidCredentialsFile = errors.New("invalid credentials file")
	ErrEmptyProfile           = errors.New("empty profile")
	ErrEmptyKey               = errors.New("empty key")
	ErrEmptyKeyValue          = errors.New("empty value")
)

type credential struct {
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    *string
	Region          *string
}

type parser struct {
	scanner *bufio.Scanner
}

func newParser(input string) *parser {
	return &parser{
		scanner: bufio.NewScanner(strings.NewReader(input)),
	}
}

func (p *parser) parse() ([]credential, error) {
	var credentials []credential

	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())

		if line == "" || line[0] == '#' {
			continue // skip empty lines and comments
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			profile := strings.TrimSpace(line[1 : len(line)-1])
			if profile == "" {
				return nil, errors.Join(ErrEmptyProfile, ErrInvalidCredentialsFile)
			}
			credentials = append(credentials, credential{Profile: profile})
		} else {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				if key == "" {
					return nil, errors.Join(ErrEmptyKey, ErrInvalidCredentialsFile)
				}

				if value == "" {
					return nil, errors.Join(ErrEmptyKeyValue, ErrInvalidCredentialsFile)
				}

				switch key {
				case "aws_access_key_id":
					credentials[len(credentials)-1].AccessKeyID = value
				case "aws_secret_access_key":
					credentials[len(credentials)-1].SecretAccessKey = value
				case "aws_session_token":
					credentials[len(credentials)-1].SessionToken = &value
				case "region":
					credentials[len(credentials)-1].Region = &value
				}
			}
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, err
	}

	return credentials, nil
}
