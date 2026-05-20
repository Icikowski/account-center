package i18n

import (
	"context"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// Option is a functional option for configuring the translation behavior.
type Option interface {
	apply(cfg *i18n.LocalizeConfig)
}

type optionFunc func(*i18n.LocalizeConfig)

func (f optionFunc) apply(cfg *i18n.LocalizeConfig) {
	f(cfg)
}

// WithTemplateData sets the template data for the translation.
func WithTemplateData(data map[string]any) Option {
	return optionFunc(func(cfg *i18n.LocalizeConfig) {
		cfg.TemplateData = data
	})
}

// WithPluralCount sets the plural count for the translation.
func WithPluralCount(count any) Option {
	return optionFunc(func(cfg *i18n.LocalizeConfig) {
		cfg.PluralCount = count
	})
}

// T translates the given message ID using the [goi18n.Localizer] from the context.
//
// If the localizer is not found in the context, it returns the message ID as a fallback.
// It accepts optional [Option]s to configure the translation behavior, such as template data and plural count.
func T(ctx context.Context, key Key, opts ...Option) string {
	l := FromContext(ctx)
	if l == nil {
		return string(key)
	}

	cfg := &i18n.LocalizeConfig{MessageID: string(key)}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	s, err := l.Localize(cfg)
	if err != nil {
		return string(key)
	}
	return s
}
