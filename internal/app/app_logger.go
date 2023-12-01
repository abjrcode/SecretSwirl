package app

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/pkgerrors"

	"github.com/rs/zerolog"
)

// InitLogFile creates (or appends) a log file in the app data directory
func InitLogFile(appDataDir, logFileName string) (*os.File, error) {
	file, logFileErr := os.OpenFile(
		filepath.Join(appDataDir, "swervo_log.json"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if logFileErr != nil {
		return nil, logFileErr
	}

	return file, nil
}

func InitLogger(logFile io.Writer, appVersion, commitSha string) zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	logSink := zerolog.MultiLevelWriter(consoleWriter, logFile)

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	logger := zerolog.New(logSink).With().
		Timestamp().
		Str("version", appVersion).
		Str("commit_sha", commitSha).
		Logger()

	log.SetFlags(0)
	log.SetOutput(logger)

	return logger
}
