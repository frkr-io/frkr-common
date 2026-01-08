package plugins

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
func (p *OIDCAuthPlugin) ValidateRequest(ctx context.Context, token string, tokenType TokenType, secretPlugin SecretPlugin) (*AuthResult, error) {
	if tokenType != TokenTypeBearer {
		return nil, fmt.Errorf("OIDCAuthPlugin only supports bearer tokens, got token type: %s", tokenType)
	}

	return nil, errors.New("OIDC authentication not yet implemented - use BasicAuthPlugin for now")
}

// CanAccessStream checks if the user/client can access a specific stream
func (p *OIDCAuthPlugin) CanAccessStream(ctx context.Context, userID string, streamID string, permission string) (bool, error) {
	if p.db == nil {
		return false, errors.New("database connection required for stream access checks")
	}

	return false, errors.New("OIDC authorization not yet implemented")
}
