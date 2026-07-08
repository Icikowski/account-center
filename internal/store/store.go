package store

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

// ErrNotFound is returned when a requested entry is not found in the store.
var ErrNotFound = errors.New("not found")

// SessionStore represents a store for transient login states and persisted auth sessions.
type SessionStore interface {
	// LoginStates returns the [model.Store] for transient login states.
	LoginStates() model.Store[string, model.LoginState]
	// Sessions returns the [model.Store] for persisted auth sessions.
	Sessions() model.Store[string, model.StoredSession]
}

// EvaluationStore represents a store for evaluator results.
type EvaluationStore interface {
	// Evaluations returns the [model.Store] for cached evaluator results.
	Evaluations() model.Store[string, model.Evaluation]
}

// StorageBackend represents a store that implements both [SessionStore] and [EvaluationStore].
type StorageBackend interface {
	SessionStore
	EvaluationStore

	// Ping checks the health of the storage backend. It should return an error if the backend is not reachable or not
	// functioning properly.
	Ping(ctx context.Context) error
}

// NewMemory creates an in-memory [StorageBackend] and starts periodic cleanup for expired entries.
func NewMemory(ctx context.Context) StorageBackend {
	return newMemoryStore(ctx)
}

// NewRedis creates an Redis [StorageBackend].
func NewRedis(client redis.Cmdable, keyPrefix string) StorageBackend {
	return newRedisStore(client, keyPrefix)
}
