package model

import (
	"context"
	"time"
)

// Store defines a generic interface for a key-value store with TTL support.
type Store[K comparable, V any] interface {
	// Get retrieves the value associated with the given key.
	Get(ctx context.Context, key K) (V, error)
	// Set stores the value with the given key and TTL.
	Set(ctx context.Context, key K, value V, ttl time.Duration) error
	// Delete removes the value associated with the given key.
	Delete(ctx context.Context, key K) error
}
