package auth

import (
	"git.sr.ht/~icikowski/account-center/internal/model"
)

// SessionStore represents a store for transient login states and persisted auth sessions.
type SessionStore interface {
	LoginStates() model.Store[string, LoginState]
	Sessions() model.Store[string, StoredSession]
}
