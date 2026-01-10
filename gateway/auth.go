package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/frkr-io/frkr-common/plugins"
)

// AuthenticateHTTPRequest authenticates and authorizes an HTTP request
// It validates the request using the auth plugin and checks stream access
func AuthenticateHTTPRequest(ctx context.Context, r *http.Request, authPlugin plugins.AuthPlugin, secretPlugin plugins.SecretPlugin, streamID string, permission string) (*plugins.AuthResult, error) {
	if authPlugin == nil {
		return nil, errors.New("auth plugin cannot be nil")
	}
	if secretPlugin == nil {
		return nil, errors.New("secret plugin cannot be nil")
	}

	// Validate the request (plugin handles token extraction)
	authResult, err := authPlugin.ValidateRequest(ctx, r, secretPlugin)
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

		allowed, err := authPlugin.CanAccessStream(ctx, authResult, streamID, permission)
		if err != nil {
			return nil, fmt.Errorf("authorization check failed: %w", err)
		}
		if !allowed {
			return nil, fmt.Errorf("access denied: user %s does not have %s permission for stream %s", userID, permission, streamID)
		}
	}

	return authResult, nil
}
