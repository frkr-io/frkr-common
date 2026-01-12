package plugins

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// OIDCAuthPlugin implements AuthPlugin for OIDC Bearer token authentication
type OIDCAuthPlugin struct {
	db *sql.DB
}

// NewOIDCAuthPlugin creates a new OIDCAuthPlugin
func NewOIDCAuthPlugin(db *sql.DB) *OIDCAuthPlugin {
	return &OIDCAuthPlugin{db: db}
}

// ValidateRequest validates a Bearer token using OIDC
func (p *OIDCAuthPlugin) ValidateRequest(ctx context.Context, r *http.Request, secretPlugin SecretPlugin) (*AuthResult, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("OIDCAuthPlugin only supports bearer tokens")
	}

	// token := strings.TrimPrefix(authHeader, "Bearer ")
	return nil, errors.New("OIDC authentication not yet implemented - use BasicAuthPlugin for now")
}

// ValidateAuthHeader validates a Bearer token from the header string directly
func (p *OIDCAuthPlugin) ValidateAuthHeader(ctx context.Context, authHeader string, secretPlugin SecretPlugin) (*AuthResult, error) {
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("OIDCAuthPlugin only supports bearer tokens")
	}

	return nil, errors.New("OIDC authentication not yet implemented - use BasicAuthPlugin for now")
}

// CanAccessStream checks if the user/client can access a specific stream
// CanAccessStream checks if the user/client can access a specific stream
func (p *OIDCAuthPlugin) CanAccessStream(ctx context.Context, authResult *AuthResult, streamID string, permission string) (bool, error) {
	if authResult.AuthSource != "oidc" {
		return false, fmt.Errorf("OIDCAuthPlugin cannot authorize user from source: %s", authResult.AuthSource)
	}

	if p.db == nil {
		return false, errors.New("database connection required for stream access checks")
	}

	return false, errors.New("OIDC authorization not yet implemented")
}
