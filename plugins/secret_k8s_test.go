package plugins

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestK8sSecretPlugin_GetUserPassword(t *testing.T) {
	mockClient := newMockK8sSecretClient()
	mockClient.addSecret("default", "frkr-user-testuser", map[string][]byte{
		"username":  []byte("testuser"),
		"password":  []byte("testpassword123"),
		"tenant_id": []byte("test-tenant-123"),
	})

	plugin, err := NewK8sSecretPlugin(mockClient, "default")
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		password, tenantID, err := plugin.GetUserPassword(context.Background(), "testuser")
		require.NoError(t, err)
		assert.Equal(t, "testpassword123", password)
		assert.Equal(t, "test-tenant-123", tenantID)
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

	t.Run("missing password key", func(t *testing.T) {
		mockClient.addSecret("default", "frkr-user-nopass", map[string][]byte{
			"username": []byte("nopass"),
		})
		_, _, err := plugin.GetUserPassword(context.Background(), "nopass")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not contain password key")
	})
}

func TestK8sSecretPlugin_GetClientSecret(t *testing.T) {
	mockClient := newMockK8sSecretClient()
	mockClient.addSecret("default", "frkr-client-testclient", map[string][]byte{
		"client_id":     []byte("testclient"),
		"client_secret": []byte("secret123"),
		"tenant_id":     []byte("tenant-123"),
	})

	plugin, err := NewK8sSecretPlugin(mockClient, "default")
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		secret, tenantID, err := plugin.GetClientSecret(context.Background(), "testclient")
		require.NoError(t, err)
		assert.Equal(t, "secret123", secret)
		assert.Equal(t, "tenant-123", tenantID)
	})

	t.Run("client not found", func(t *testing.T) {
		_, _, err := plugin.GetClientSecret(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("empty clientID", func(t *testing.T) {
		_, _, err := plugin.GetClientSecret(context.Background(), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})
}

func TestK8sSecretPlugin_GetEncryptionKey(t *testing.T) {
	mockClient := newMockK8sSecretClient()
	mockClient.addSecret("default", "frkr-stream-stream123", map[string][]byte{
		"public_key":  []byte("public-key-data"),
		"private_key": []byte("private-key-data"),
	})

	plugin, err := NewK8sSecretPlugin(mockClient, "default")
	require.NoError(t, err)

	t.Run("successful retrieval", func(t *testing.T) {
		publicKey, privateKey, err := plugin.GetEncryptionKey(context.Background(), "stream123")
		require.NoError(t, err)
		assert.Equal(t, []byte("public-key-data"), publicKey)
		assert.Equal(t, []byte("private-key-data"), privateKey)
	})

	t.Run("global encryption key", func(t *testing.T) {
		mockClient.addSecret("default", "frkr-encryption-key", map[string][]byte{
			"public_key": []byte("global-public-key"),
		})
		publicKey, privateKey, err := plugin.GetEncryptionKey(context.Background(), "")
		require.NoError(t, err)
		assert.Equal(t, []byte("global-public-key"), publicKey)
		assert.Nil(t, privateKey) // Private key is optional
	})

	t.Run("key not found", func(t *testing.T) {
		_, _, err := plugin.GetEncryptionKey(context.Background(), "nonexistent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestK8sSecretPlugin_GetSecret(t *testing.T) {
	mockClient := newMockK8sSecretClient()
	mockClient.addSecret("default", "frkr-user-password-testuser", map[string][]byte{
		"value": []byte("password123"),
	})

	plugin, err := NewK8sSecretPlugin(mockClient, "default")
	require.NoError(t, err)

	t.Run("get user password via GetSecret", func(t *testing.T) {
		mockClient.addSecret("default", "frkr-user-testuser2", map[string][]byte{
			"username": []byte("testuser2"),
			"password": []byte("pass123"),
		})
		data, err := plugin.GetSecret(context.Background(), SecretTypeUserPassword, "testuser2")
		require.NoError(t, err)
		assert.Equal(t, "pass123", string(data))
	})

	t.Run("unsupported secret type", func(t *testing.T) {
		_, err := plugin.GetSecret(context.Background(), "unknown_type", "test")
		require.Error(t, err)
	})
}

func TestNewK8sSecretPlugin_NamespaceAutoDetection(t *testing.T) {
	t.Run("with provided namespace", func(t *testing.T) {
		mockClient := newMockK8sSecretClient()
		plugin, err := NewK8sSecretPlugin(mockClient, "test-namespace")
		require.NoError(t, err)
		assert.Equal(t, "test-namespace", plugin.namespace)
	})

	t.Run("nil client error", func(t *testing.T) {
		_, err := NewK8sSecretPlugin(nil, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}
