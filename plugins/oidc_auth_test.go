package plugins

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDCAuthPlugin_ValidateRequest(t *testing.T) {
	plugin := NewOIDCAuthPlugin(nil)

	t.Run("bearer token not yet implemented", func(t *testing.T) {
		_, err := plugin.ValidateRequest(
			context.Background(),
			"Bearer jwt-token",
			TokenTypeBearer,
			nil,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("wrong token type", func(t *testing.T) {
		_, err := plugin.ValidateRequest(
			context.Background(),
			"Basic user:pass",
			TokenTypeBasic,
			nil,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only supports bearer tokens")
	})
}

func TestOIDCAuthPlugin_CanAccessStream(t *testing.T) {
	plugin := NewOIDCAuthPlugin(nil)

	t.Run("not yet implemented - checks database requirement first", func(t *testing.T) {
		_, err := plugin.CanAccessStream(context.Background(), "user", "stream", "read")
		require.Error(t, err)
		// CanAccessStream checks for database first before checking if implemented
		assert.Contains(t, err.Error(), "database connection required")
	})

	t.Run("nil database", func(t *testing.T) {
		plugin := NewOIDCAuthPlugin(nil)
		_, err := plugin.CanAccessStream(context.Background(), "user", "stream", "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database connection required")
	})
}
