package i18n

import (
	"context"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

type contextKeyI18n struct{}

// NewContext returns a new context with the given [goi18n.Localizer] set.
func NewContext(ctx context.Context, l *goi18n.Localizer) context.Context {
	return context.WithValue(ctx, contextKeyI18n{}, l)
}

// FromContext retrieves the [goi18n.Localizer] from the context.
func FromContext(ctx context.Context) *goi18n.Localizer {
	l, _ := ctx.Value(contextKeyI18n{}).(*goi18n.Localizer)
	return l
}
