package plugins

import "context"

// SecretPlugin is the interface for secret storage and retrieval
// It abstracts secret storage backends (K8s, DB, Vault, etc.)
type SecretPlugin interface {
	// GetUserPassword retrieves a user's password hash
	// Returns: password hash, tenant ID, error
	GetUserPassword(ctx context.Context, username string) (passwordHash string, tenantID string, err error)

	// GetClientSecret retrieves a client's secret for OAuth client credentials flow
	// Returns: client secret, tenant ID, error
	GetClientSecret(ctx context.Context, clientID string) (clientSecret string, tenantID string, err error)

	// GetEncryptionKey retrieves an encryption key for a stream
	// Returns: public key, private key (if available), error
	GetEncryptionKey(ctx context.Context, streamID string) (publicKey []byte, privateKey []byte, err error)

	// GetSecret retrieves a generic secret by type and identifier (for extensibility)
	// secretType can be: "user_password", "client_secret", "encryption_key", "service_account", etc.
	GetSecret(ctx context.Context, secretType string, identifier string) ([]byte, error)
}

// SecretType constants for GetSecret method
const (
	SecretTypeUserPassword  = "user_password"
	SecretTypeClientSecret  = "client_secret"
	SecretTypeEncryptionKey = "encryption_key"
	SecretTypeServiceAccount = "service_account"
)
