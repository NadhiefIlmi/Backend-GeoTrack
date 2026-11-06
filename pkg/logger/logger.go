package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Log is the global logger instance
var Log zerolog.Logger

// Init initializes the structured logger based on environment
func Init(env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = os.Stdout

	if env == "development" || env == "dev" || env == "" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	Log = zerolog.New(output).With().Timestamp().Caller().Logger()
	log.Logger = Log
}

// Info logs with info level
func Info() *zerolog.Event {
	return Log.Info()
}

// Error logs with error level
func Error() *zerolog.Event {
	return Log.Error()
}

// Warn logs with warn level
func Warn() *zerolog.Event {
	return Log.Warn()
}

// Debug logs with debug level
func Debug() *zerolog.Event {
	return Log.Debug()
}

// Fatal logs with fatal level
func Fatal() *zerolog.Event {
	return Log.Fatal()
}

