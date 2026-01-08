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

func TestExtractAuthToken(t *testing.T) {
	t.Run("basic auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		credentials := base64.StdEncoding.EncodeToString([]byte("user:pass"))
		req.Header.Set("Authorization", "Basic "+credentials)

		tokenType, token, err := ExtractAuthToken(req)
		require.NoError(t, err)
		assert.Equal(t, plugins.TokenTypeBasic, tokenType)
		assert.Equal(t, "Basic "+credentials, token)
	})

	t.Run("bearer token header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer jwt-token-here")

		tokenType, token, err := ExtractAuthToken(req)
		require.NoError(t, err)
		assert.Equal(t, plugins.TokenTypeBearer, tokenType)
		assert.Equal(t, "jwt-token-here", token)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		_, _, err := ExtractAuthToken(req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("unsupported authorization type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Digest username=\"user\"")

		_, _, err := ExtractAuthToken(req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported")
	})
}

func TestAuthenticateRequest(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	tenant, err := dbcommon.CreateOrGetTenant(testDB, "test-tenant-middleware")
	require.NoError(t, err)

	// Setup users table and test user
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

	secretPlugin, err := plugins.NewDatabaseSecretPlugin(testDB)
	require.NoError(t, err)

	authPlugin := plugins.NewBasicAuthPlugin(testDB)

	stream, err := dbcommon.CreateStream(testDB, tenant.ID, "test-stream", "Test stream", 7)
	require.NoError(t, err)

	// Create test user
	password := "testpass123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	_, err = testDB.Exec(`
		INSERT INTO users (tenant_id, username, password_hash)
		VALUES ($1, $2, $3)
	`, tenant.ID, "testuser", string(passwordHash))
	require.NoError(t, err)

	t.Run("successful authentication and authorization", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:testpass123"))
		authHeader := "Basic " + credentials

		result, err := AuthenticateRequest(
			context.Background(),
			authPlugin,
			secretPlugin,
			plugins.TokenTypeBasic,
			authHeader,
			stream.ID,
			"read",
		)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "testuser", result.UserID)
		assert.Equal(t, tenant.ID, result.TenantID)
	})

	t.Run("authentication failure", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:wrongpass"))
		authHeader := "Basic " + credentials

		_, err := AuthenticateRequest(
			context.Background(),
			authPlugin,
			secretPlugin,
			plugins.TokenTypeBasic,
			authHeader,
			stream.ID,
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("authorization failure - wrong tenant", func(t *testing.T) {
		// Create another tenant and stream
		otherTenant, err := dbcommon.CreateOrGetTenant(testDB, "other-tenant")
		require.NoError(t, err)

		otherStream, err := dbcommon.CreateStream(testDB, otherTenant.ID, "other-stream", "Other", 7)
		require.NoError(t, err)

		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:testpass123"))
		authHeader := "Basic " + credentials

		_, err = AuthenticateRequest(
			context.Background(),
			authPlugin,
			secretPlugin,
			plugins.TokenTypeBasic,
			authHeader,
			otherStream.ID,
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")
	})

	t.Run("nil auth plugin", func(t *testing.T) {
		_, err := AuthenticateRequest(
			context.Background(),
			nil,
			secretPlugin,
			plugins.TokenTypeBasic,
			"Basic test",
			"stream-id",
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("nil secret plugin", func(t *testing.T) {
		_, err := AuthenticateRequest(
			context.Background(),
			authPlugin,
			nil,
			plugins.TokenTypeBasic,
			"Basic test",
			"stream-id",
			"read",
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("no stream ID - authentication only", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:testpass123"))
		authHeader := "Basic " + credentials

		result, err := AuthenticateRequest(
			context.Background(),
			authPlugin,
			secretPlugin,
			plugins.TokenTypeBasic,
			authHeader,
			"", // No stream ID
			"",
		)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestAuthenticateHTTPRequest(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

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
	})
}
