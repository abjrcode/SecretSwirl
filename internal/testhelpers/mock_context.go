package testhelpers

import (
	"context"

	"github.com/abjrcode/swervo/internal/app"
)

func NewMockAppContext() app.Context {
	return app.NewContext(context.TODO(), "test_user_id", "test_request_id", "test_causation_id", "test_correlation_id")
}
