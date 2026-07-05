package auth

import (
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
)

const (
	paramNext  = "next"
	paramCode  = "code"
	paramState = "state"
)

var (
	errNotFound                        = errors.New("not found")
	errOIDCDiscoveryFailed             = errors.New("failed to discover OIDC provider")
	errOIDCMetadataReadFailed          = errors.New("failed to read OIDC provider metadata")
	errBaseURLParse                    = errors.New("failed to parse base URL")
	errReauthenticationRequired        = errors.New("re-authentication required")
	errUnexpectedRefreshResultType     = errors.New("unexpected refresh result type")
	errSaveLoginState                  = errors.New("failed to save login state")
	errLoadLoginState                  = errors.New("failed to load login state")
	errDeleteLoginState                = errors.New("failed to delete login state")
	errAuthorizationCodeExchangeFailed = errors.New("failed to exchange authorization code")
	errSaveSession                     = errors.New("failed to save session")
	errLoadSession                     = errors.New("failed to load session")
	errDeleteSession                   = errors.New("failed to delete session")
	errIDTokenMissing                  = errors.New("id_token missing from token response")
	errIDTokenDecode                   = errors.New("failed to decode id_token")
	errIDTokenVerification             = errors.New("failed to verify id_token")
	errIDTokenNonceMismatch            = errors.New("nonce mismatch when verifying id_token")
	errRevocationFailed                = errors.New("token revocation failed")
	errRevocationRequestBuild          = errors.New("failed to build token revocation request")
	errUserInfoClaimsDecode            = errors.New("failed to decode user info claims")
	errRandomIDGeneration              = errors.New("failed to generate random ID")

	defaultScopes = []string{
		oidc.ScopeOpenID,
		"profile",
		"email",
		"groups",
		oidc.ScopeOfflineAccess,
	}
)

type profileClaims struct {
	Subject string   `json:"sub"`
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	Groups  []string `json:"groups"`
	Nonce   string   `json:"nonce"`
}
