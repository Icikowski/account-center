package config

import (
	"reflect"

	"github.com/caarlos0/env/v11"
)

// Load reads the configuration from environment variables and returns a [Config] struct.
func Load() (*Config, error) {
	cfg, err := env.ParseAsWithOptions[Config](env.Options{
		Prefix:  "AC_",
		FuncMap: map[reflect.Type]env.ParserFunc{},
	})
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
