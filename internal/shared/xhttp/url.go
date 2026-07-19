package xhttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"go.yaml.in/yaml/v3"

	"git.sr.ht/~icikowski/account-center/internal/consts"
)

var errSchemeNotAllowed = errors.New("scheme not allowed")

// URL is a wrapper around [url.URL] that implements JSON & YAML marshaling and unmarshaling.
type URL struct {
	url.URL
}

// MarshalJSON implements [json.Marshaler].
func (u *URL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON implements [json.Unmarshaler].
func (u *URL) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if err := u.parse(s); err != nil {
		return err
	}
	return nil
}

// MarshalYAML implements [yaml.Marshaler].
func (u *URL) MarshalYAML() (any, error) {
	return yaml.Marshal(u.String())
}

// UnmarshalYAML implements [yaml.Unmarshaler].
func (u *URL) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	if err := u.parse(s); err != nil {
		return err
	}
	return nil
}

func (u *URL) parse(input string) error {
	parsed, err := url.Parse(input)
	if err != nil {
		return err
	}

	if parsed.Scheme != consts.SchemeHTTP && parsed.Scheme != consts.SchemeHTTPS {
		return fmt.Errorf("%w: %s", errSchemeNotAllowed, parsed.Scheme)
	}

	u.URL = *parsed
	return nil
}

var (
	_ json.Marshaler   = (*URL)(nil)
	_ json.Unmarshaler = (*URL)(nil)
	_ yaml.Marshaler   = (*URL)(nil)
	_ yaml.Unmarshaler = (*URL)(nil)
)
