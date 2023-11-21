package utils

import "time"

type Clock interface {
	NowUnix() int64
}

type clock struct{}

func NewClock() Clock {
	return &clock{}
}

// NowUnix returns the current Unix time in seconds (number of seconds since Epoch).
func (d *clock) NowUnix() int64 {
	return time.Now().Unix()
}
