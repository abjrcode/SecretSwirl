package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"

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

func InitLogger(logFile io.Writer) zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	logSink := zerolog.MultiLevelWriter(consoleWriter, logFile)

	logger := zerolog.New(logSink).With().Timestamp().Logger()

	log.SetFlags(0)
	log.SetOutput(logger)

	return logger
}
