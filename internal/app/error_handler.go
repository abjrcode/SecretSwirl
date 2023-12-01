package app

import (
	"errors"

	"github.com/awnumar/memguard"
	"github.com/rs/zerolog"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var ErrFatal = errors.New("FATAL_ERROR")

type ErrorHandler interface {
	Catch(ctx Context, logger zerolog.Logger, err error)
	CatchWithMsg(ctx Context, logger zerolog.Logger, err error, msg string)
}

type errorHandler struct {
}

func NewErrorHandler() ErrorHandler {
	return &errorHandler{}
}

func (eh *errorHandler) showFatalErrorDialog(ctx Context, logger zerolog.Logger) {
	wailsRuntime.MessageDialog(ctx, wailsRuntime.MessageDialogOptions{
		Type:    wailsRuntime.ErrorDialog,
		Title:   "Error",
		Message: "Unexpected error occurred. This is a bug & the application will exit to protect your data",
	})
}

func (eh *errorHandler) Catch(ctx Context, logger zerolog.Logger, err error) {
	eh.CatchWithMsg(ctx, logger, err, "")
}

func (eh *errorHandler) CatchWithMsg(ctx Context, logger zerolog.Logger, err error, msg string) {
	if err != nil {
		logger.Error().Stack().Err(err).Msgf(msg)

		if ctx != nil {
			eh.showFatalErrorDialog(ctx, logger)
		}

		memguard.SafeExit(1)
	}
}
