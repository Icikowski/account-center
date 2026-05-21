package layouts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
)

// BaseData represents the base data structure that is passed to all templates in the application.
type BaseData struct {
	InstanceName string
	Counters     Counters
	User         *User
}

// Counters holds the number of services and knowledge base articles.
type Counters struct {
	Services              int
	KnowledgeBaseArticles int
}

// User hold information about the currently logged in user, such as their full name and email address.
type User struct {
	FullName string
	Email    string
}

// GravatarURL generates a Gravatar URL for the user based on their e-mail address.
func (u *User) GravatarURL() string {
	emailHash := "000000000000000000000000000000000000000000000000000000"
	if u != nil && u.Email != "" {
		hasher := sha256.Sum256([]byte(strings.TrimSpace(u.Email)))
		emailHash = hex.EncodeToString(hasher[:])
	}

	values := url.Values{}
	values.Add("size", "256")
	values.Add("rating", "pg")
	if u != nil && u.FullName != "" {
		values.Add("default", "initials")
		values.Add("name", u.FullName)
	} else {
		values.Add("default", "mp")
	}

	return "https://gravatar.com/avatar/" + emailHash + "?" + values.Encode()
}

type contextKeyLayoutData struct{}

// NewContext creates a new context with the given [BaseData] set.
func NewContext(ctx context.Context, data BaseData) context.Context {
	return context.WithValue(ctx, contextKeyLayoutData{}, data)
}

// FromContext retrieves the [BaseData] from the context.
func FromContext(ctx context.Context) BaseData {
	data, ok := ctx.Value(contextKeyLayoutData{}).(BaseData)
	if !ok {
		return BaseData{}
	}
	return data
}
