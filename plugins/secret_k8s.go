package plugins

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

// K8sSecretClient is an interface for reading Kubernetes secrets
type K8sSecretClient interface {
	GetSecret(ctx context.Context, namespace, name string) (map[string][]byte, error)
}

// K8sSecretPlugin implements SecretPlugin using Kubernetes Secrets
type K8sSecretPlugin struct {
	client    K8sSecretClient
	namespace string
}

// NewK8sSecretPlugin creates a new K8sSecretPlugin
// If namespace is empty, it will try to read from the pod's namespace
// (via /var/run/secrets/kubernetes.io/serviceaccount/namespace) or default to "default"
func NewK8sSecretPlugin(client K8sSecretClient, namespace string) (*K8sSecretPlugin, error) {
	if client == nil {
		return nil, errors.New("k8s client cannot be nil")
	}

	ns := namespace
	if ns == "" {
		// Try to read namespace from pod service account
		if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
			ns = strings.TrimSpace(string(data))
		}
		if ns == "" {
			ns = "default"
		}
	}

	return &K8sSecretPlugin{
		client:    client,
		namespace: ns,
	}, nil
}

// GetUserPassword retrieves a user's password from a Kubernetes secret
// Secret naming pattern: frkr-user-{username}
// Secret keys: username, password
func (p *K8sSecretPlugin) GetUserPassword(ctx context.Context, username string) (passwordHash string, tenantID string, err error) {
	if username == "" {
		return "", "", errors.New("username cannot be empty")
	}

	secretName := fmt.Sprintf("frkr-user-%s", username)
	secretData, err := p.client.GetSecret(ctx, p.namespace, secretName)
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	password, ok := secretData["password"]
	if !ok {
		return "", "", fmt.Errorf("secret %s does not contain password key", secretName)
	}

	// Get tenant ID if available, otherwise use username as identifier
	tenantIDVal, _ := secretData["tenant_id"]
	if len(tenantIDVal) > 0 {
		tenantID = string(tenantIDVal)
	}

	// For K8s secrets, password might be plain or hashed
	// Return as-is and let BasicAuthPlugin handle verification
	return string(password), tenantID, nil
}

// GetClientSecret retrieves a client's secret from a Kubernetes secret
// Secret naming pattern: frkr-client-{clientid}
// Secret keys: client_id, client_secret
func (p *K8sSecretPlugin) GetClientSecret(ctx context.Context, clientID string) (clientSecret string, tenantID string, err error) {
	if clientID == "" {
		return "", "", errors.New("clientID cannot be empty")
	}

	secretName := fmt.Sprintf("frkr-client-%s", clientID)
	secretData, err := p.client.GetSecret(ctx, p.namespace, secretName)
	if err != nil {
		return "", "", fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	secret, ok := secretData["client_secret"]
	if !ok {
		return "", "", fmt.Errorf("secret %s does not contain client_secret key", secretName)
	}

	// Get tenant ID if available
	tenantIDVal, _ := secretData["tenant_id"]
	if len(tenantIDVal) > 0 {
		tenantID = string(tenantIDVal)
	}

	return string(secret), tenantID, nil
}

// GetEncryptionKey retrieves encryption keys from a Kubernetes secret
// Secret naming pattern: frkr-stream-{streamid} or frkr-encryption-key (global)
// Secret keys: public_key, private_key
func (p *K8sSecretPlugin) GetEncryptionKey(ctx context.Context, streamID string) (publicKey []byte, privateKey []byte, err error) {
	secretName := "frkr-encryption-key"
	if streamID != "" {
		secretName = fmt.Sprintf("frkr-stream-%s", streamID)
	}

	secretData, err := p.client.GetSecret(ctx, p.namespace, secretName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}

	publicKey, ok := secretData["public_key"]
	if !ok {
		return nil, nil, fmt.Errorf("secret %s does not contain public_key key", secretName)
	}

	// Private key is optional (may not be available in all contexts)
	privateKey, _ = secretData["private_key"]

	return publicKey, privateKey, nil
}

// GetSecret retrieves a generic secret by type and identifier
func (p *K8sSecretPlugin) GetSecret(ctx context.Context, secretType string, identifier string) ([]byte, error) {
	switch secretType {
	case SecretTypeUserPassword:
		password, _, err := p.GetUserPassword(ctx, identifier)
		return []byte(password), err
	case SecretTypeClientSecret:
		secret, _, err := p.GetClientSecret(ctx, identifier)
		return []byte(secret), err
	case SecretTypeEncryptionKey:
		publicKey, _, err := p.GetEncryptionKey(ctx, identifier)
		return publicKey, err
	default:
		// Generic secret lookup: assume secret name format frkr-{type}-{identifier}
		secretName := fmt.Sprintf("frkr-%s-%s", strings.ReplaceAll(secretType, "_", "-"), identifier)
		secretData, err := p.client.GetSecret(ctx, p.namespace, secretName)
		if err != nil {
			return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
		}
		// Return the first value found, or empty if no values
		for _, v := range secretData {
			return v, nil
		}
		return nil, fmt.Errorf("secret %s is empty", secretName)
	}
}
