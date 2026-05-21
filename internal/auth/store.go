package auth

import (
	"git.sr.ht/~icikowski/account-center/internal/model"
)

// SessionStore represents a store for transient login states and persisted auth sessions.
type SessionStore interface {
	// LoginStates returns the [model.Store] for transient login states.
	LoginStates() model.Store[string, LoginState]
	// Sessions returns the [model.Store] for persisted auth sessions.
	Sessions() model.Store[string, StoredSession]
}
