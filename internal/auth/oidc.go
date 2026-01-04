package auth

import (
	"context"
	"fmt"

	"github.com/frkr-io/frkr-common/plugins"
	"golang.org/x/oauth2"
)

// OIDCAuthPlugin implements generic OIDC authentication
type OIDCAuthPlugin struct {
	// OAuth2 config for token validation
	Config *oauth2.Config
	
	// TokenValidator validates OIDC tokens
	TokenValidator TokenValidator
}

// TokenValidator validates OIDC tokens
type TokenValidator interface {
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

// TokenClaims contains OIDC token claims
type TokenClaims struct {
	Subject   string   // User ID
	Email     string   // User email
	TenantID  string   // Tenant/organization ID
	Roles     []string // User roles
	Groups    []string // User groups
	ExpiresAt int64    // Token expiration timestamp
}

// ValidateRequest validates an OIDC token
func (p *OIDCAuthPlugin) ValidateRequest(ctx context.Context, token string) (*plugins.AuthResult, error) {
	// Validate token
	claims, err := p.TokenValidator.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return &plugins.AuthResult{
		UserID:      claims.Subject,
		TenantID:    claims.TenantID,
		Roles:       claims.Roles,
		Permissions: []string{}, // Permissions derived from roles/groups
	}, nil
}

// CanAccessStream checks if user can access a stream
func (p *OIDCAuthPlugin) CanAccessStream(ctx context.Context, userID string, streamID string, permission string) (bool, error) {
	// OIDC: Check permissions based on roles/groups
	// This is a placeholder - actual implementation would check against policy engine
	return true, nil
}

