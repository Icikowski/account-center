package config

import (
	"net"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"git.sr.ht/~icikowski/account-center/internal/shared/xerror"
)

// Config holds the configuration loaded from environment variables.
type Config struct {
	Instance      InstanceConfig      `envPrefix:"INSTANCE_"`
	Server        ServerConfig        `envPrefix:"SERVER_"`
	Catalog       CatalogConfig       `envPrefix:"CATALOG_"`
	KnowledgeBase KnowledgeBaseConfig `envPrefix:"KB_"`
	Auth          AuthConfig          `envPrefix:"AUTH_"`
	OIDC          OIDCConfig          `envPrefix:"OIDC_"`
	Redis         RedisConfig         `envPrefix:"REDIS_"`
	Log           LogConfig           `envPrefix:"LOG_"`
}

// Validate checks if the whole [Config] is valid.
func (c Config) Validate() error {
	errs := make([]error, 0, 8)

	if err := c.Instance.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Server.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Catalog.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.KnowledgeBase.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Auth.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.OIDC.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Redis.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Log.Validate(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid configuration", errs...)
	}
	return nil
}

// InstanceConfig holds the configuration for the instance.
type InstanceConfig struct {
	Name    string `env:"NAME"`
	BaseURL string `env:"BASE_URL"`
}

// Validate checks if the [InstanceConfig] is valid.
func (c InstanceConfig) Validate() error {
	errs := make([]error, 0, 1)

	if c.BaseURL != "" {
		validSchemes := []string{"http", "https"}
		baseURL, err := url.Parse(c.BaseURL)
		if err != nil || !slices.Contains(validSchemes, baseURL.Scheme) || baseURL.Host == "" {
			errs = append(errs, xerror.NewValidationError("invalid instance base URL"))
		}
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid instance configuration", errs...)
	}
	return nil
}

// ServerConfig holds the configuration for the server.
type ServerConfig struct {
	Address           string   `env:"ADDRESS"         envDefault:""`
	Port              uint16   `env:"PORT"            envDefault:"8080"`
	TrustedProxyCIDRs []string `env:"TRUSTED_PROXIES"`
}

// Validate checks if the [ServerConfig] is valid.
func (c ServerConfig) Validate() error {
	errs := make([]error, 0, 3)

	if c.Address != "" {
		if ip := net.ParseIP(c.Address); ip == nil {
			errs = append(errs, xerror.NewValidationError("invalid server address"))
		}
	}
	if c.Port == 0 {
		errs = append(errs, xerror.NewValidationError("server port must be between 1 and 65535"))
	}

	trustedProxiesErrs := make([]error, 0, len(c.TrustedProxyCIDRs))
	for i, trustedProxyCIDR := range c.TrustedProxyCIDRs {
		if _, _, err := parseTrustedProxyCIDR(trustedProxyCIDR); err != nil {
			trustedProxiesErrs = append(trustedProxiesErrs, xerror.NewItemValidationError(
				i,
				xerror.NewValidationError("invalid trusted proxy CIDR", err),
			))
		}
	}
	if len(trustedProxiesErrs) != 0 {
		errs = append(errs, xerror.NewValidationError("invalid trusted proxy CIDRs", trustedProxiesErrs...))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid server configuration", errs...)
	}
	return nil
}

// CatalogConfig holds the configuration for the services catalog.
type CatalogConfig struct {
	Path           string        `env:"PATH"            envDefault:"./catalog.yaml"`
	ReloadDebounce time.Duration `env:"RELOAD_DEBOUNCE" envDefault:"500ms"`
}

// Validate checks if the [CatalogConfig] is valid.
func (c CatalogConfig) Validate() error {
	errs := make([]error, 0, 2)

	if c.Path == "" {
		errs = append(errs, xerror.NewValidationError("catalog path is required"))
	}
	if c.ReloadDebounce <= 0 {
		errs = append(errs, xerror.NewValidationError("catalog reload debounce must be greater than zero"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid catalog configuration", errs...)
	}
	return nil
}

// KnowledgeBaseConfig holds the configuration for the knowledge base articles.
type KnowledgeBaseConfig struct {
	Enabled        bool          `env:"ENABLED"         envDefault:"false"`
	Path           string        `env:"PATH"            envDefault:"./kb"`
	ReloadDebounce time.Duration `env:"RELOAD_DEBOUNCE" envDefault:"500ms"`
}

// Validate checks if the [KnowledgeBaseConfig] is valid.
func (c KnowledgeBaseConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	errs := make([]error, 0, 2)

	if c.Path == "" {
		errs = append(errs, xerror.NewValidationError("knowledge base path is required when enabled"))
	}
	if c.ReloadDebounce <= 0 {
		errs = append(errs, xerror.NewValidationError("knowledge base reload debounce must be greater than zero"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid knowledge base configuration", errs...)
	}
	return nil
}

// AuthConfig holds the configuration for auth session persistence.
type AuthConfig struct {
	SessionTTL        time.Duration `env:"SESSION_TTL"         envDefault:"24h"`
	LoginStateTTL     time.Duration `env:"LOGIN_STATE_TTL"     envDefault:"10m"`
	SessionCookieName string        `env:"SESSION_COOKIE_NAME" envDefault:"account-center-session"`
}

// Validate checks if the [AuthConfig] is valid.
func (c AuthConfig) Validate() error {
	errs := make([]error, 0, 3)

	if c.SessionTTL <= 0 {
		errs = append(errs, xerror.NewValidationError("session TTL must be greater than zero"))
	}
	if c.LoginStateTTL <= 0 {
		errs = append(errs, xerror.NewValidationError("login state TTL must be greater than zero"))
	}
	if c.SessionCookieName == "" {
		errs = append(errs, xerror.NewValidationError("session cookie name is required"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid auth configuration", errs...)
	}
	return nil
}

// OIDCConfig holds the configuration for OpenID Connect authentication.
type OIDCConfig struct {
	ProviderURL   string        `env:"PROVIDER_URL"`
	ClientID      string        `env:"CLIENT_ID"`
	ClientSecret  string        `env:"CLIENT_SECRET"`
	RefreshBefore time.Duration `env:"REFRESH_BEFORE" envDefault:"1m"`
}

// Validate checks if the [OIDCConfig] is valid.
func (c OIDCConfig) Validate() error {
	errs := make([]error, 0, 4)

	if c.ProviderURL == "" {
		errs = append(errs, xerror.NewValidationError("OIDC provider URL is required"))
	}
	if c.ClientID == "" {
		errs = append(errs, xerror.NewValidationError("OIDC client ID is required"))
	}
	if c.ClientSecret == "" {
		errs = append(errs, xerror.NewValidationError("OIDC client secret is required"))
	}
	if c.RefreshBefore <= 0 {
		errs = append(errs, xerror.NewValidationError("OIDC refresh before duration must be greater than zero"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid OIDC configuration", errs...)
	}
	return nil
}

// RedisConfig holds the configuration for Redis-backed auth storage.
type RedisConfig struct {
	Enabled   bool   `env:"ENABLED"    envDefault:"false"`
	Address   string `env:"ADDRESS"`
	Username  string `env:"USERNAME"`
	Password  string `env:"PASSWORD"`
	Database  int    `env:"DATABASE"   envDefault:"0"`
	KeyPrefix string `env:"KEY_PREFIX" envDefault:"account-center"`
}

// Client creates a Redis client based on the configuration (if enabled).
func (c RedisConfig) Client() *redis.Client {
	if !c.Enabled {
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Username: c.Username,
		Password: c.Password,
		DB:       c.Database,
	})
}

// Validate checks if the [RedisConfig] is valid.
func (c RedisConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	errs := make([]error, 0, 2)

	if c.Address == "" {
		errs = append(errs, xerror.NewValidationError("Redis address is required when enabled"))
	}
	if c.KeyPrefix == "" {
		errs = append(errs, xerror.NewValidationError("Redis key prefix is required when enabled"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid Redis configuration", errs...)
	}
	return nil
}

// LogConfig holds the configuration for logging.
type LogConfig struct {
	Level  zerolog.Level `env:"LEVEL"  envDefault:"info"`
	Pretty bool          `env:"PRETTY" envDefault:"false"`
}

// Validate checks if the [LogConfig] is valid.
func (c LogConfig) Validate() error {
	errs := make([]error, 0, 1)

	if c.Level < zerolog.TraceLevel || c.Level > zerolog.PanicLevel {
		errs = append(errs, xerror.NewValidationError("invalid log level"))
	}

	if len(errs) != 0 {
		return xerror.NewValidationError("invalid log configuration", errs...)
	}
	return nil
}

func parseTrustedProxyCIDR(value string) (net.IP, *net.IPNet, error) {
	if ip := net.ParseIP(value); ip != nil {
		bits := 32
		if ip.To4() == nil {
			bits = 128
		}
		return net.ParseCIDR(value + "/" + strconv.Itoa(bits))
	}

	return net.ParseCIDR(value)
}
