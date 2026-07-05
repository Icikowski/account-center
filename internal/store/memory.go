package store

import (
	"context"
	"sync"
	"time"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

const memoryCleanupInterval = time.Minute

type memoryStore struct {
	loginStates *memoryValueStore[model.LoginState]
	sessions    *memoryValueStore[model.StoredSession]
	evaluations *memoryValueStore[model.Evaluation]
}

func newMemoryStore(ctx context.Context) StorageBackend {
	loginStates := &memoryValueStore[model.LoginState]{}
	sessions := &memoryValueStore[model.StoredSession]{}
	evaluations := &memoryValueStore[model.Evaluation]{}

	go loginStates.cleanupLoop(ctx)
	go sessions.cleanupLoop(ctx)
	go evaluations.cleanupLoop(ctx)

	return &memoryStore{loginStates: loginStates, sessions: sessions, evaluations: evaluations}
}

// LoginStates implements [SessionStore].
func (s *memoryStore) LoginStates() model.Store[string, model.LoginState] {
	return s.loginStates
}

// Sessions implements [SessionStore].
func (s *memoryStore) Sessions() model.Store[string, model.StoredSession] {
	return s.sessions
}

// Evaluations implements [EvaluationStore].
func (s *memoryStore) Evaluations() model.Store[string, model.Evaluation] {
	return s.evaluations
}

type memoryValueStore[T any] struct {
	entries sync.Map
}

func (s *memoryValueStore[T]) Get(_ context.Context, key string) (T, error) {
	var zero T

	raw, ok := s.entries.Load(key)
	if !ok {
		return zero, ErrNotFound
	}

	entry, ok := raw.(memoryEntry[T])
	if !ok {
		s.entries.Delete(key)
		return zero, ErrNotFound
	}
	if !entry.expires.IsZero() && !time.Now().Before(entry.expires) {
		s.entries.Delete(key)
		return zero, ErrNotFound
	}

	return entry.value, nil
}

func (s *memoryValueStore[T]) Set(_ context.Context, key string, value T, ttl time.Duration) error {
	s.entries.Store(key, memoryEntry[T]{
		value:   value,
		expires: time.Now().Add(ttl),
	})
	return nil
}

func (s *memoryValueStore[T]) Delete(_ context.Context, key string) error {
	s.entries.Delete(key)
	return nil
}

func (s *memoryValueStore[T]) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(memoryCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.deleteExpired()
		}
	}
}

func (s *memoryValueStore[T]) deleteExpired() {
	s.entries.Range(func(key, value any) bool {
		entry, ok := value.(memoryEntry[T])
		if ok && !entry.expires.IsZero() && !time.Now().Before(entry.expires) {
			s.entries.Delete(key)
		}
		return true
	})
}

type memoryEntry[T any] struct {
	value   T
	expires time.Time
}
