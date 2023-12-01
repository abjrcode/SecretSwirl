package app

import (
	"github.com/rs/zerolog"
	"github.com/wailsapp/wails/v2/pkg/logger"
)

type wailsLoggerAdapter struct {
	logger *zerolog.Logger
}

func (l *wailsLoggerAdapter) Print(message string) {
	l.logger.Info().Msg(message)
}

func (l *wailsLoggerAdapter) Trace(message string) {
	l.logger.Trace().Msg(message)
}

func (l *wailsLoggerAdapter) Debug(message string) {
	l.logger.Debug().Msg(message)
}

func (l *wailsLoggerAdapter) Info(message string) {
	l.logger.Info().Msg(message)
}

func (l *wailsLoggerAdapter) Warning(message string) {
	l.logger.Warn().Msg(message)
}

func (l *wailsLoggerAdapter) Error(message string) {
	l.logger.Error().Msg(message)
}

func (l *wailsLoggerAdapter) Fatal(message string) {
	l.logger.Fatal().Msg(message)
}

func NewWailsLoggerAdapter(logger *zerolog.Logger) logger.Logger {
	return &wailsLoggerAdapter{
		logger: logger,
	}
}
