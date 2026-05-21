package auth

import "context"

type contextKeySession struct{}

// NewContext returns a new context with the given [Session] set.
func NewContext(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, contextKeySession{}, session)
}

// FromContext retrieves the [Session] from the context.
func FromContext(ctx context.Context) *Session {
	session, ok := ctx.Value(contextKeySession{}).(*Session)
	if !ok {
		return nil
	}
	return session
}
