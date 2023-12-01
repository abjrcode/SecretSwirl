package utils

import "github.com/rs/xid"

func NewRequestId() string {
	return xid.New().String()
}
