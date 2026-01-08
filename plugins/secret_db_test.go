package plugins

import (
	"context"
	"database/sql"
	"testing"

	dbcommon "github.com/frkr-io/frkr-common/db"
	"github.com/frkr-io/frkr-common/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// setupUsersTable creates the users table if it doesn't exist (for testing)
func setupUsersTable(t *testing.T, dbConn *sql.DB) {
	_, err := dbConn.Exec(`
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
}

// setupTestUser creates a test user in the database
func setupTestUser(t *testing.T, dbConn *sql.DB, tenantID, username, password string) {
	setupUsersTable(t, dbConn)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	_, err = dbConn.Exec(`
		INSERT INTO users (tenant_id, username, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (tenant_id, username) DO UPDATE SET password_hash = EXCLUDED.password_hash
	`, tenantID, username, string(passwordHash))
	require.NoError(t, err)
}

func TestDatabaseSecretPlugin_GetUserPassword(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	// Create a test tenant
	tenant, err := dbcommon.CreateOrGetTenant(testDB, "test-tenant")
	require.NoError(t, err)

	plugin, err := NewDatabaseSecretPlugin(testDB)
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		setupTestUser(t, testDB, tenant.ID, "testuser", "password123")

		passwordHash, tenantID, err := plugin.GetUserPassword(context.Background(), "testuser")
		require.NoError(t, err)
		assert.NotEmpty(t, passwordHash)
		assert.Equal(t, tenant.ID, tenantID)

		// Verify password hash can verify the password
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("password123"))
		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		_, _, err := plugin.GetUserPassword(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty username", func(t *testing.T) {
		_, _, err := plugin.GetUserPassword(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestDatabaseSecretPlugin_GetClientSecret(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	plugin, err := NewDatabaseSecretPlugin(testDB)
	require.NoError(t, err)

	t.Run("clients table does not exist", func(t *testing.T) {
		_, _, err := plugin.GetClientSecret(context.Background(), "testclient")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("empty clientID", func(t *testing.T) {
		_, _, err := plugin.GetClientSecret(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestDatabaseSecretPlugin_GetEncryptionKey(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	plugin, err := NewDatabaseSecretPlugin(testDB)
	require.NoError(t, err)

	t.Run("stream_keys table does not exist", func(t *testing.T) {
		_, _, err := plugin.GetEncryptionKey(context.Background(), "stream123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("empty streamID", func(t *testing.T) {
		_, _, err := plugin.GetEncryptionKey(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestDatabaseSecretPlugin_GetSecret(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	tenant, err := dbcommon.CreateOrGetTenant(testDB, "test-tenant-secret")
	require.NoError(t, err)

	plugin, err := NewDatabaseSecretPlugin(testDB)
	require.NoError(t, err)

	t.Run("get user password via GetSecret", func(t *testing.T) {
		setupTestUser(t, testDB, tenant.ID, "secretuser", "secretpass")

		data, err := plugin.GetSecret(context.Background(), SecretTypeUserPassword, "secretuser")
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify it's a valid bcrypt hash
		err = bcrypt.CompareHashAndPassword(data, []byte("secretpass"))
		assert.NoError(t, err)
	})

	t.Run("unsupported secret type", func(t *testing.T) {
		_, err := plugin.GetSecret(context.Background(), "unknown_type", "test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported")
	})
}

func TestNewDatabaseSecretPlugin(t *testing.T) {
	t.Run("nil database error", func(t *testing.T) {
		_, err := NewDatabaseSecretPlugin(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("valid database", func(t *testing.T) {
		testDB, _ := db.SetupTestDB(t, "../migrations")
		plugin, err := NewDatabaseSecretPlugin(testDB)
		require.NoError(t, err)
		assert.NotNil(t, plugin)
	})
}
