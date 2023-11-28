package faults

import (
	"context"
	"errors"

	"github.com/awnumar/memguard"
	"github.com/rs/zerolog"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var ErrFatal = errors.New("FATAL_ERROR")

type ErrorHandler interface {
	InitWailsContext(ctx *context.Context)

	Catch(logger zerolog.Logger, err error)
	CatchWithMsg(logger zerolog.Logger, err error, msg string)
}

type errorHandler struct {
	wailsCtx *context.Context
}

func NewErrorHandler() ErrorHandler {
	return &errorHandler{}
}

func (eh *errorHandler) InitWailsContext(ctx *context.Context) {
	eh.wailsCtx = ctx
}

func (eh *errorHandler) showFatalErrorDialog(logger zerolog.Logger) {
	wailsRuntime.MessageDialog(*eh.wailsCtx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.ErrorDialog,
		Title:   "Error",
		Message: "Unexpected error occurred. This is a bug & the application will exit to protect your data",
	})
}

func (eh *errorHandler) Catch(logger zerolog.Logger, err error) {
	eh.CatchWithMsg(logger, err, "")
}

func (eh *errorHandler) CatchWithMsg(logger zerolog.Logger, err error, msg string) {
	if err != nil {
		logger.Error().Stack().Err(err).Msgf(msg)

		if eh.wailsCtx != nil {
			eh.showFatalErrorDialog(logger)
		}

		memguard.SafeExit(1)
	}
}
