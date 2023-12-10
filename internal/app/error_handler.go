package app

import (
	"errors"

	"github.com/awnumar/memguard"
	"github.com/rs/zerolog"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type ValidationError struct {
	ActualError error
}

func (e *ValidationError) Error() string {
	return e.ActualError.Error()
}

func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

func (e *ValidationError) Unwrap() error {
	return e.ActualError
}

var (
	ErrValidation = &ValidationError{ActualError: errors.New("VALIDATION_ERROR")}
	ErrFatal      = errors.New("FATAL_ERROR")
)

func NewValidationError(msg string) error {
	return &ValidationError{ActualError: errors.New(msg)}
}

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
