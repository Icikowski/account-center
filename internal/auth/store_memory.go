package auth

import (
	"context"
	"sync"
	"time"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

type memoryStore struct {
	loginStates *memoryValueStore[LoginState]
	sessions    *memoryValueStore[StoredSession]
}

// NewMemoryStore creates an in-memory auth store and starts periodic cleanup for expired entries.
func NewMemoryStore(ctx context.Context) SessionStore {
	loginStates := &memoryValueStore[LoginState]{}
	sessions := &memoryValueStore[StoredSession]{}

	go loginStates.cleanupLoop(ctx)
	go sessions.cleanupLoop(ctx)

	return &memoryStore{loginStates: loginStates, sessions: sessions}
}

// LoginStates implements [SessionStore].
func (s *memoryStore) LoginStates() model.Store[string, LoginState] {
	return s.loginStates
}

// Sessions implements [SessionStore].
func (s *memoryStore) Sessions() model.Store[string, StoredSession] {
	return s.sessions
}

type memoryValueStore[T any] struct {
	entries sync.Map
}

func (s *memoryValueStore[T]) Get(_ context.Context, key string) (T, error) {
	var zero T

	raw, ok := s.entries.Load(key)
	if !ok {
		return zero, errNotFound
	}

	entry, ok := raw.(memoryEntry[T])
	if !ok {
		s.entries.Delete(key)
		return zero, errNotFound
	}
	if !entry.expires.IsZero() && !time.Now().Before(entry.expires) {
		s.entries.Delete(key)
		return zero, errNotFound
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
	ticker := time.NewTicker(time.Minute)
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
