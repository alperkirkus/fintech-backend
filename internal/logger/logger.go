package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init(level, format string) {
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	zerolog.TimeFieldFormat = time.RFC3339

	var outputWriter io.Writer = os.Stdout
	if format == "pretty" {
		outputWriter = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}
	}

	log.Logger = zerolog.New(outputWriter).With().Timestamp().Caller().Logger()
}

func Get() zerolog.Logger {
	return log.Logger
}
