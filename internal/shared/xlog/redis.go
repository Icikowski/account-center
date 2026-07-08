package xlog

import (
	"context"

	"github.com/rs/zerolog"
)

// RedisLogger adapts a [zerolog.Logger] to [github.com/redis/go-redis/v9/internal.Logging] interface.
type RedisLogger interface {
	// Printf logs a formatted message with the given context.
	Printf(ctx context.Context, format string, v ...any)
}

type redisLogger struct {
	log zerolog.Logger
}

// NewRedisLogger creates a [RedisLogger] backed by [zerolog.Logger].
func NewRedisLogger(log zerolog.Logger) RedisLogger {
	return redisLogger{log: log}
}

// Printf implements [RedisLogger].
func (l redisLogger) Printf(_ context.Context, format string, v ...any) {
	l.log.Debug().Msgf(format, v...)
}
