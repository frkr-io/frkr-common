package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/frkr-io/frkr-common/plugins"
)

// ExtractAuthToken extracts the authentication token from the HTTP request
// Returns the token type (Basic or Bearer) and the raw token string
func ExtractAuthToken(r *http.Request) (plugins.TokenType, string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", errors.New("missing Authorization header")
	}

	// Check for Basic Auth
	if strings.HasPrefix(authHeader, "Basic ") {
		return plugins.TokenTypeBasic, authHeader, nil
	}

	// Check for Bearer token
	if strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		return plugins.TokenTypeBearer, token, nil
	}

	return "", "", fmt.Errorf("unsupported authorization type: %s", authHeader)
}

// AuthenticateRequest authenticates and authorizes an HTTP request
// It extracts the token, validates it using the auth plugin, and checks stream access
func AuthenticateRequest(ctx context.Context, authPlugin plugins.AuthPlugin, secretPlugin plugins.SecretPlugin, tokenType plugins.TokenType, token string, streamID string, permission string) (*plugins.AuthResult, error) {
	if authPlugin == nil {
		return nil, errors.New("auth plugin cannot be nil")
	}
	if secretPlugin == nil {
		return nil, errors.New("secret plugin cannot be nil")
	}

	// Validate the request token
	authResult, err := authPlugin.ValidateRequest(ctx, token, tokenType, secretPlugin)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if authResult == nil {
		return nil, errors.New("authentication returned nil result")
	}

	// If streamID is provided, check authorization
	if streamID != "" {
		// Determine user/client ID for authorization check
		userID := authResult.UserID
		if userID == "" {
			userID = authResult.ClientID
		}
		if userID == "" {
			return nil, errors.New("auth result has no user ID or client ID")
		}

		allowed, err := authPlugin.CanAccessStream(ctx, userID, streamID, permission)
		if err != nil {
			return nil, fmt.Errorf("authorization check failed: %w", err)
		}
		if !allowed {
			return nil, fmt.Errorf("access denied: user %s does not have %s permission for stream %s", userID, permission, streamID)
		}
	}

	return authResult, nil
}

// AuthenticateHTTPRequest is a convenience function that combines ExtractAuthToken and AuthenticateRequest
func AuthenticateHTTPRequest(ctx context.Context, r *http.Request, authPlugin plugins.AuthPlugin, secretPlugin plugins.SecretPlugin, streamID string, permission string) (*plugins.AuthResult, error) {
	tokenType, token, err := ExtractAuthToken(r)
	if err != nil {
		return nil, err
	}

	return AuthenticateRequest(ctx, authPlugin, secretPlugin, tokenType, token, streamID, permission)
}
