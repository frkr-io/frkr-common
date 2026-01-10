package plugins

import (
	"context"
	"net/http"
)

// TokenType defines the type of authentication token
type TokenType string

const (
	// TokenTypeBasic indicates Basic Authentication (username:password)
	TokenTypeBasic TokenType = "basic"
	// TokenTypeBearer indicates Bearer token authentication (OIDC/JWT)
	TokenTypeBearer TokenType = "bearer"
)

// AuthResult contains authentication and authorization information
type AuthResult struct {
	// UserID is the authenticated user's ID (for user-based auth)
	UserID string

	// ClientID is the authenticated client's ID (for client credentials flow)
	ClientID string

	// TenantID is the tenant/organization ID
	TenantID string

	// ClientType indicates the type of client: "user", "client", or "service_account"
	ClientType string

	// AuthSource indicates the source/plugin that authenticated the user (e.g., "basic", "oidc")
	AuthSource string

	// Roles are the user's roles
	Roles []string

	// Permissions are the user's specific permissions
	Permissions []string
}

// AuthPlugin is the interface for authentication plugins
type AuthPlugin interface {
	// ValidateRequest validates the request (e.g. checking headers) and returns auth result
	// Uses SecretPlugin for credential lookup (passwords, client secrets)
	ValidateRequest(ctx context.Context, r *http.Request, secretPlugin SecretPlugin) (*AuthResult, error)

	// CanAccessStream checks if the user/client can access a specific stream
	CanAccessStream(ctx context.Context, authResult *AuthResult, streamID string, permission string) (bool, error)
}

