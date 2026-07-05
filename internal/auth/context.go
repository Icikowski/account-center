package auth

import (
	"context"

	"git.sr.ht/~icikowski/account-center/internal/model"
)

type contextKeySession struct{}

// NewContext returns a new context with the given [model.Session] set.
func NewContext(ctx context.Context, session *model.Session) context.Context {
	return context.WithValue(ctx, contextKeySession{}, session)
}

// FromContext retrieves the [model.Session] from the context.
func FromContext(ctx context.Context) *model.Session {
	session, ok := ctx.Value(contextKeySession{}).(*model.Session)
	if !ok {
		return nil
	}
	return session
}
