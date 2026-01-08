package plugins

import (
	"context"
	"fmt"
)

// mockK8sSecretClient implements K8sSecretClient for testing
type mockK8sSecretClient struct {
	secrets map[string]map[string][]byte
	err     error
}

func (m *mockK8sSecretClient) GetSecret(ctx context.Context, namespace, name string) (map[string][]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := fmt.Sprintf("%s/%s", namespace, name)
	secret, ok := m.secrets[key]
	if !ok {
		return nil, fmt.Errorf("secret %s not found", key)
	}
	return secret, nil
}

// newMockK8sSecretClient creates a new mock K8s secret client
func newMockK8sSecretClient() *mockK8sSecretClient {
	return &mockK8sSecretClient{
		secrets: make(map[string]map[string][]byte),
	}
}

// addSecret adds a secret to the mock
func (m *mockK8sSecretClient) addSecret(namespace, name string, data map[string][]byte) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	m.secrets[key] = data
}

// setError sets an error to return on next GetSecret call
func (m *mockK8sSecretClient) setError(err error) {
	m.err = err
}

// clearError clears the error
func (m *mockK8sSecretClient) clearError() {
	m.err = nil
}
