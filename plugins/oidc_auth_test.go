package plugins

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCAuthPlugin_ValidateRequest(t *testing.T) {
	plugin := NewOIDCAuthPlugin(nil)

	t.Run("bearer token not yet implemented", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer jwt-token")

		_, err := plugin.ValidateRequest(
			context.Background(),
			req,
			nil,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("wrong token type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		// Missing bearer, sending basic
		req.Header.Set("Authorization", "Basic user:pass")

		_, err := plugin.ValidateRequest(
			context.Background(),
			req,
			nil,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only supports bearer tokens")
	})
}

func TestOIDCAuthPlugin_CanAccessStream(t *testing.T) {
	plugin := NewOIDCAuthPlugin(nil)
	validAuthResult := &AuthResult{UserID: "sub1", AuthSource: "oidc"}

	t.Run("not yet implemented - checks database requirement first", func(t *testing.T) {
		_, err := plugin.CanAccessStream(context.Background(), validAuthResult, "stream", "read")
		require.Error(t, err)
		// CanAccessStream checks for database first before checking if implemented
		assert.Contains(t, err.Error(), "database connection required")
	})

	t.Run("nil database", func(t *testing.T) {
		plugin := NewOIDCAuthPlugin(nil)
		_, err := plugin.CanAccessStream(context.Background(), validAuthResult, "stream", "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database connection required")
	})

	t.Run("wrong auth source", func(t *testing.T) {
		res := &AuthResult{UserID: "user1", AuthSource: "basic"}
		allowed, err := plugin.CanAccessStream(context.Background(), res, "stream", "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot authorize user from source")
		assert.False(t, allowed)
	})
}
