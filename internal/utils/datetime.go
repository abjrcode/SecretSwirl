package utils

import "time"

type Datetime interface {
	NowUnix() int64
}

type datetime struct{}

func NewDatetime() Datetime {
	return &datetime{}
}

// NowUnix returns the current Unix time in seconds (number of seconds since Epoch).
func (d *datetime) NowUnix() int64 {
	return time.Now().Unix()
}
