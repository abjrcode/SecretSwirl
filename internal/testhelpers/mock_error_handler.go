package testhelpers

import (
	"context"
	"testing"

	"github.com/abjrcode/swervo/internal/faults"
	"github.com/rs/zerolog"
)

type mockErrorHandler struct {
	t        *testing.T
	wailsCtx *context.Context
}

func NewMockErrorHandler(t *testing.T) faults.ErrorHandler {
	return &mockErrorHandler{
		t: t,
	}
}

func (eh *mockErrorHandler) InitWailsContext(ctx *context.Context) {
	eh.wailsCtx = ctx
}

func (eh *mockErrorHandler) Catch(logger zerolog.Logger, err error) {
	eh.CatchWithMsg(logger, err, "")
}

func (eh *mockErrorHandler) CatchWithMsg(logger zerolog.Logger, err error, msg string) {
	if err != nil {
		eh.t.Fatalf("Error: %v, Msg: %s", err, msg)
	}
}
