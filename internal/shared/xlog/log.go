package xlog

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New creates a new [zerolog.Logger] with the specified log level and pretty printing option.
func New(level zerolog.Level, pretty bool) zerolog.Logger {
	var w io.Writer = os.Stderr
	if pretty {
		w = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.DateTime,
		}
	}

	return zerolog.New(w).Level(level).With().Timestamp().Logger()
}

// InitialLogger creates a minimal [zerolog.Logger] for initial logging before the full configuration is available.
func InitialLogger() zerolog.Logger {
	return New(zerolog.DebugLevel, false)
}
