package model

import "time"

// AuthorizationRequest describes a newly prepared OIDC login request.
type AuthorizationRequest struct {
	LoginID string `json:"login_id"`
	URL     string `json:"url"`
}

// LoginState represents the transient state of an ongoing login flow.
type LoginState struct {
	ID           string    `json:"id"`
	Next         string    `json:"next"`
	RedirectURL  string    `json:"redirect_url"`
	CodeVerifier string    `json:"code_verifier"`
	Nonce        string    `json:"nonce"`
	CreatedAt    time.Time `json:"created_at"`
}

// StoredSession represents a persisted authenticated session.
type StoredSession struct {
	ID           string    `json:"id"`
	Subject      string    `json:"subject"`
	User         User      `json:"user"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	IDToken      string    `json:"id_token"`
	TokenType    string    `json:"token_type"`
	Expiry       time.Time `json:"expiry"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Session returns the public view of an authenticated session.
func (r StoredSession) Session() Session {
	return Session{
		ID:   r.ID,
		User: r.User,
	}
}

// Session represents the public view of an authenticated session, without any token data.
type Session struct {
	ID   string `json:"id"`
	User User   `json:"user"`
}

// User contains the identity data fetched from OIDC.
type User struct {
	Subject string   `json:"sub"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Groups  []string `json:"groups"`
}
