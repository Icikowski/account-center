package xhttp

import (
	"errors"
	"fmt"
	"net/url"

	"go.yaml.in/yaml/v3"
)

var errSchemeNotAllowed = errors.New("scheme not allowed")

// URL is a wrapper around [url.URL] that implements YAML marshaling and unmarshaling.
type URL struct {
	url.URL
}

// MarshalYAML implements [yaml.Marshaler].
func (u *URL) MarshalYAML() (any, error) {
	return u.String(), nil
}

// UnmarshalYAML implements [yaml.Unmarshaler].
func (u *URL) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	parsed, err := url.Parse(s)
	if err != nil {
		return err
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%w: %s", errSchemeNotAllowed, parsed.Scheme)
	}

	u.URL = *parsed
	return nil
}
