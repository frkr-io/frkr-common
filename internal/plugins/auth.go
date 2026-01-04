package plugins

import "context"

// AuthResult contains authentication and authorization information
type AuthResult struct {
	// UserID is the authenticated user's ID
	UserID string
	
	// TenantID is the tenant/organization ID
	TenantID string
	
	// Roles are the user's roles
	Roles []string
	
	// Permissions are the user's specific permissions
	Permissions []string
}

// AuthPlugin is the interface for authentication plugins
type AuthPlugin interface {
	// ValidateRequest validates an authentication token and returns auth result
	ValidateRequest(ctx context.Context, token string) (*AuthResult, error)
	
	// CanAccessStream checks if the user can access a specific stream
	CanAccessStream(ctx context.Context, userID string, streamID string, permission string) (bool, error)
}

