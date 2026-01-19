package gateway

import (
	"context"
	"encoding/base64"
	"net/http/httptest"
	"testing"

	"github.com/frkr-io/frkr-common/db"
	dbcommon "github.com/frkr-io/frkr-common/db"
	"github.com/frkr-io/frkr-common/plugins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthenticateHTTPRequest(t *testing.T) {
	testDB, _ := db.SetupTestDB(t)

	tenant, err := dbcommon.CreateOrGetTenant(testDB, "test-tenant-http")
	require.NoError(t, err)

	// Setup users table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			username STRING(255) NOT NULL,
			password_hash STRING(255) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
			deleted_at TIMESTAMPTZ,
			UNIQUE (tenant_id, username)
		)
	`)
	require.NoError(t, err)

	secretPlugin, _ := plugins.NewDatabaseSecretPlugin(testDB)
	authPlugin := plugins.NewBasicAuthPlugin(testDB)

	stream, err := dbcommon.CreateStream(testDB, tenant.ID, "http-stream", "Test", 7)
	require.NoError(t, err)

	// Create test user
	password := "httppass123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	_, err = testDB.Exec(`
		INSERT INTO users (tenant_id, username, password_hash)
		VALUES ($1, $2, $3)
	`, tenant.ID, "httpuser", string(passwordHash))
	require.NoError(t, err)

	t.Run("successful request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stream?stream_id="+stream.ID, nil)
		credentials := base64.StdEncoding.EncodeToString([]byte("httpuser:httppass123"))
		req.Header.Set("Authorization", "Basic "+credentials)

		result, err := AuthenticateHTTPRequest(
			context.Background(),
			req,
			authPlugin,
			secretPlugin,
			stream.ID,
			"read",
		)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "httpuser", result.UserID)
		assert.Equal(t, "basic", result.AuthSource)
	})

	t.Run("successful request - no stream ID (auth only)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stream", nil)
		credentials := base64.StdEncoding.EncodeToString([]byte("httpuser:httppass123"))
		req.Header.Set("Authorization", "Basic "+credentials)

		result, err := AuthenticateHTTPRequest(
			context.Background(),
			req,
			authPlugin,
			secretPlugin,
			"", // No stream ID
			"",
		)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "httpuser", result.UserID)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stream", nil)

		_, err := AuthenticateHTTPRequest(
			context.Background(),
			req,
			authPlugin,
			secretPlugin,
			stream.ID,
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing Authorization header")
	})

	t.Run("authentication failure (wrong password)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stream", nil)
		credentials := base64.StdEncoding.EncodeToString([]byte("httpuser:wrongpass"))
		req.Header.Set("Authorization", "Basic "+credentials)

		_, err := AuthenticateHTTPRequest(
			context.Background(),
			req,
			authPlugin,
			secretPlugin,
			stream.ID,
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("authorization failure - wrong tenant", func(t *testing.T) {
		// Create another tenant and stream
		otherTenant, err := dbcommon.CreateOrGetTenant(testDB, "other-tenant-http")
		require.NoError(t, err)
		otherStream, err := dbcommon.CreateStream(testDB, otherTenant.ID, "other-stream-http", "Other", 7)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/stream", nil)
		credentials := base64.StdEncoding.EncodeToString([]byte("httpuser:httppass123"))
		req.Header.Set("Authorization", "Basic "+credentials)

		_, err = AuthenticateHTTPRequest(
			context.Background(),
			req,
			authPlugin,
			secretPlugin,
			otherStream.ID, // User doesn't have access to this stream
			"read",
		)
		require.Error(t, err)
		// BasicAuthPlugin.CanAccessStream returns "false, nil" if unauthorized/not found (to allow fallback),
		// BUT AuthenticateHTTPRequest wraps the return. 
		// Wait, BasicAuthPlugin returns "false, nil" if *stream* logic fails conditionally?
		// "if strings.Contains(err.Error(), "not found") { return false, nil }"
		// "if stream.TenantID != userTenantID.String { return false, nil }"
		//
		// AuthenticateHTTPRequest logic:
		// allowed, err := authPlugin.CanAccessStream(...)
		// if err != nil { return err }
		// if !allowed { return error("access denied...") }
		
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("nil auth plugin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/stream", nil)
		_, err := AuthenticateHTTPRequest(
			context.Background(),
			req,
			nil,
			secretPlugin,
			stream.ID,
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}
