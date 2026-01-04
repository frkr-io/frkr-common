package auth

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/frkr-io/frkr-common/internal/plugins"
)

// BasicAuthPlugin implements basic username/password authentication
type BasicAuthPlugin struct {
	// UserStore provides user credentials (username, password hash)
	UserStore UserStore
}

// UserStore provides user credential storage
type UserStore interface {
	GetPasswordHash(username string) ([]byte, error)
	GetUserInfo(username string) (*UserInfo, error)
}

// UserInfo contains user information
type UserInfo struct {
	UserID   string
	TenantID string
	Roles    []string
}

// ValidateRequest validates a basic auth token
func (p *BasicAuthPlugin) ValidateRequest(ctx context.Context, token string) (*plugins.AuthResult, error) {
	// Decode basic auth token
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("invalid basic auth token: %w", err)
	}

	// Parse username:password
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid basic auth format")
	}

	username := parts[0]
	password := parts[1]

	// Get password hash from store
	hash, err := p.UserStore.GetPasswordHash(username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify password (using constant-time comparison)
	// TODO: Use bcrypt or Argon2 for password hashing
	if subtle.ConstantTimeCompare(hash, []byte(password)) != 1 {
		return nil, fmt.Errorf("invalid password")
	}

	// Get user info
	userInfo, err := p.UserStore.GetUserInfo(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return &plugins.AuthResult{
		UserID:      userInfo.UserID,
		TenantID:    userInfo.TenantID,
		Roles:       userInfo.Roles,
		Permissions: []string{}, // Basic auth has no fine-grained permissions
	}, nil
}

// CanAccessStream checks if user can access a stream
func (p *BasicAuthPlugin) CanAccessStream(ctx context.Context, userID string, streamID string, permission string) (bool, error) {
	// Basic auth: all authenticated users can access all streams
	// More fine-grained control would require OIDC/OAuth2
	return true, nil
}

