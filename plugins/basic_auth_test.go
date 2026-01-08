package plugins

import (
	"context"
	"encoding/base64"
	"testing"

	dbcommon "github.com/frkr-io/frkr-common/db"
	"github.com/frkr-io/frkr-common/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthPlugin_ValidateRequest(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	tenant, err := dbcommon.CreateOrGetTenant(testDB, "test-tenant-auth")
	require.NoError(t, err)

	setupUsersTable(t, testDB)
	setupTestUser(t, testDB, tenant.ID, "testuser", "password123")

	secretPlugin, err := NewDatabaseSecretPlugin(testDB)
	require.NoError(t, err)

	authPlugin := NewBasicAuthPlugin(testDB)

	t.Run("valid credentials with bcrypt hash", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:password123"))
		authHeader := "Basic " + credentials

		result, err := authPlugin.ValidateRequest(context.Background(), authHeader, TokenTypeBasic, secretPlugin)
		require.NoError(t, err)
		assert.Equal(t, "testuser", result.UserID)
		assert.Equal(t, tenant.ID, result.TenantID)
		assert.Equal(t, "user", result.ClientType)
	})

	t.Run("invalid password", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:wrongpassword"))
		authHeader := "Basic " + credentials

		_, err := authPlugin.ValidateRequest(context.Background(), authHeader, TokenTypeBasic, secretPlugin)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("non-existent user", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("nonexistent:password"))
		authHeader := "Basic " + credentials

		_, err := authPlugin.ValidateRequest(context.Background(), authHeader, TokenTypeBasic, secretPlugin)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := authPlugin.ValidateRequest(context.Background(), "", TokenTypeBasic, secretPlugin)
		require.Error(t, err)
	})

	t.Run("invalid token format", func(t *testing.T) {
		_, err := authPlugin.ValidateRequest(context.Background(), "InvalidFormat", TokenTypeBasic, secretPlugin)
		require.Error(t, err)
	})

	t.Run("wrong token type", func(t *testing.T) {
		credentials := base64.StdEncoding.EncodeToString([]byte("testuser:password123"))
		authHeader := "Basic " + credentials

		_, err := authPlugin.ValidateRequest(context.Background(), authHeader, TokenTypeBearer, secretPlugin)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "only supports basic auth")
	})

	t.Run("plain password from K8s secret", func(t *testing.T) {
		// Create a mock K8s client with plain password
		mockClient := newMockK8sSecretClient()
		mockClient.addSecret("default", "frkr-user-k8suser", map[string][]byte{
			"username":  []byte("k8suser"),
			"password":  []byte("plainpass123"),
			"tenant_id": []byte(tenant.ID),
		})

		k8sSecretPlugin, err := NewK8sSecretPlugin(mockClient, "default")
		require.NoError(t, err)

		credentials := base64.StdEncoding.EncodeToString([]byte("k8suser:plainpass123"))
		authHeader := "Basic " + credentials

		result, err := authPlugin.ValidateRequest(context.Background(), authHeader, TokenTypeBasic, k8sSecretPlugin)
		require.NoError(t, err)
		assert.Equal(t, "k8suser", result.UserID)
	})
}

func TestBasicAuthPlugin_CanAccessStream(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	tenant1, err := dbcommon.CreateOrGetTenant(testDB, "tenant-1")
	require.NoError(t, err)

	tenant2, err := dbcommon.CreateOrGetTenant(testDB, "tenant-2")
	require.NoError(t, err)

	setupUsersTable(t, testDB)
	setupTestUser(t, testDB, tenant1.ID, "user1", "pass1")

	stream1, err := dbcommon.CreateStream(testDB, tenant1.ID, "stream-1", "Test stream 1", 7)
	require.NoError(t, err)

	stream2, err := dbcommon.CreateStream(testDB, tenant2.ID, "stream-2", "Test stream 2", 7)
	require.NoError(t, err)

	authPlugin := NewBasicAuthPlugin(testDB)

	t.Run("user can access stream in same tenant with read permission", func(t *testing.T) {
		allowed, err := authPlugin.CanAccessStream(context.Background(), "user1", stream1.ID, "read")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("user can access stream in same tenant with write permission", func(t *testing.T) {
		allowed, err := authPlugin.CanAccessStream(context.Background(), "user1", stream1.ID, "write")
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("user cannot access stream in different tenant", func(t *testing.T) {
		allowed, err := authPlugin.CanAccessStream(context.Background(), "user1", stream2.ID, "read")
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("non-existent user", func(t *testing.T) {
		_, err := authPlugin.CanAccessStream(context.Background(), "nonexistent", stream1.ID, "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("non-existent stream", func(t *testing.T) {
		allowed, err := authPlugin.CanAccessStream(context.Background(), "user1", "nonexistent-stream-id", "read")
		require.NoError(t, err)
		assert.False(t, allowed) // Returns false, not error
	})

	t.Run("nil database", func(t *testing.T) {
		plugin := NewBasicAuthPlugin(nil)
		_, err := plugin.CanAccessStream(context.Background(), "user1", stream1.ID, "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database connection required")
	})

	t.Run("empty userID", func(t *testing.T) {
		_, err := authPlugin.CanAccessStream(context.Background(), "", stream1.ID, "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("empty streamID", func(t *testing.T) {
		_, err := authPlugin.CanAccessStream(context.Background(), "user1", "", "read")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("invalid permission", func(t *testing.T) {
		allowed, err := authPlugin.CanAccessStream(context.Background(), "user1", stream1.ID, "delete")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported permission")
		assert.False(t, allowed)
	})
}

func TestBasicAuthPlugin_TenantIDLookup(t *testing.T) {
	testDB, _ := db.SetupTestDB(t, "../migrations")

	tenant, err := dbcommon.CreateOrGetTenant(testDB, "test-tenant-lookup")
	require.NoError(t, err)

	setupUsersTable(t, testDB)
	setupTestUser(t, testDB, tenant.ID, "lookupuser", "pass123")

	// Test with K8s secret that has tenant_id
	mockClient := newMockK8sSecretClient()
	mockClient.addSecret("default", "frkr-user-lookupuser", map[string][]byte{
		"username":  []byte("lookupuser"),
		"password":  []byte("pass123"),
		"tenant_id": []byte(tenant.ID),
	})

	k8sSecretPlugin, err := NewK8sSecretPlugin(mockClient, "default")
	require.NoError(t, err)

	authPlugin := NewBasicAuthPlugin(testDB)

	credentials := base64.StdEncoding.EncodeToString([]byte("lookupuser:pass123"))
	authHeader := "Basic " + credentials

	result, err := authPlugin.ValidateRequest(context.Background(), authHeader, TokenTypeBasic, k8sSecretPlugin)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, result.TenantID)
}

// Helper to create basic auth header
func createBasicAuthHeader(username, password string) string {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return "Basic " + credentials
}
