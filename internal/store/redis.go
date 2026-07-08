package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

const (
	redisKindLoginState = "login-state"
	redisKindSession    = "session"
	redisKindEvaluation = "evaluation"
)

type redisStore struct {
	client      redis.Cmdable
	loginStates *redisValueStore[model.LoginState]
	sessions    *redisValueStore[model.StoredSession]
	evaluations *redisValueStore[model.Evaluation]
}

func newRedisStore(client redis.Cmdable, keyPrefix string) StorageBackend {
	return &redisStore{
		client:      client,
		loginStates: newRedisValueStore[model.LoginState](client, keyPrefix, redisKindLoginState),
		sessions:    newRedisValueStore[model.StoredSession](client, keyPrefix, redisKindSession),
		evaluations: newRedisValueStore[model.Evaluation](client, keyPrefix, redisKindEvaluation),
	}
}

// LoginStates implements [SessionStore].
func (s *redisStore) LoginStates() model.Store[string, model.LoginState] {
	return s.loginStates
}

// Sessions implements [SessionStore].
func (s *redisStore) Sessions() model.Store[string, model.StoredSession] {
	return s.sessions
}

// Evaluations implements [EvaluationStore].
func (s *redisStore) Evaluations() model.Store[string, model.Evaluation] {
	return s.evaluations
}

// Ping implements [StorageBackend].
func (s *redisStore) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

type redisValueStore[T any] struct {
	client    redis.Cmdable
	keyPrefix string
	kind      string
}

func newRedisValueStore[T any](client redis.Cmdable, keyPrefix, kind string) *redisValueStore[T] {
	return &redisValueStore[T]{
		client:    client,
		keyPrefix: keyPrefix,
		kind:      kind,
	}
}

func (s *redisValueStore[T]) Get(ctx context.Context, key string) (T, error) {
	var out T

	payload, err := s.client.Get(ctx, s.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return out, ErrNotFound
	}
	if err != nil {
		return out, fmt.Errorf("failed to load '%s': %w", key, err)
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return out, fmt.Errorf("failed to unmarshal '%s': %w", key, err)
	}
	return out, nil
}

func (s *redisValueStore[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal %s %q: %w", s.kind, key, err)
	}
	if err := s.client.Set(ctx, s.key(key), payload, ttl).Err(); err != nil {
		return fmt.Errorf("save %s %q: %w", s.kind, key, err)
	}
	return nil
}

func (s *redisValueStore[T]) Delete(ctx context.Context, key string) error {
	return s.client.Del(ctx, s.key(key)).Err()
}

func (s *redisValueStore[T]) key(key string) string {
	return strings.Join([]string{s.keyPrefix, s.kind, key}, ":")
}
