package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

const (
	redisKindLoginState = "login-state"
	redisKindSession    = "session"
)

// redisStore keeps auth data in Redis.
type redisStore struct {
	loginStates *redisValueStore[LoginState]
	sessions    *redisValueStore[StoredSession]
}

// NewRedisStore creates a Redis-backed auth store.
func NewRedisStore(client redis.Cmdable, keyPrefix string) SessionStore {
	return &redisStore{
		loginStates: newRedisValueStore[LoginState](client, keyPrefix, redisKindLoginState),
		sessions:    newRedisValueStore[StoredSession](client, keyPrefix, redisKindSession),
	}
}

// LoginStates returns the login-state store.
func (s *redisStore) LoginStates() model.Store[string, LoginState] {
	return s.loginStates
}

// Sessions returns the persisted-session store.
func (s *redisStore) Sessions() model.Store[string, StoredSession] {
	return s.sessions
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
		return out, errNotFound
	}
	if err != nil {
		return out, fmt.Errorf("%w '%s': %w", errLoadFailed, key, err)
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return out, fmt.Errorf("%w '%s': %w", errUnmarshalFailed, key, err)
	}
	return out, nil
}

func (s *redisValueStore[T]) Set(
	ctx context.Context,
	key string,
	value T,
	ttl time.Duration,
) error {
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
	return s.keyPrefix + ":" + s.kind + ":" + key
}
