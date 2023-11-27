package testhelpers

import (
	"github.com/stretchr/testify/mock"
)

type MockClock struct {
	mock.Mock
}

func NewMockClock() *MockClock {
	return new(MockClock)
}

func (m *MockClock) NowUnix() int64 {
	args := m.Called()
	return int64(args.Int(0))
}
