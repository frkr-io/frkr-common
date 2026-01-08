package plugins

import "context"

// CryptoPlugin is the interface for encryption and decryption operations
// It uses SecretPlugin to retrieve encryption keys, separating crypto operations from secret storage
type CryptoPlugin interface {
	// EncryptStream encrypts data for a specific stream
	// Uses SecretPlugin to retrieve encryption key
	// Returns: encrypted data, encrypted symmetric key, IV, auth tag, error
	EncryptStream(ctx context.Context, streamID string, payload []byte, secretPlugin SecretPlugin) (encryptedData []byte, encryptedKey []byte, iv []byte, authTag []byte, err error)

	// DecryptStream decrypts data for a specific stream
	// Uses SecretPlugin to retrieve decryption key
	// Requires: encrypted data, encrypted symmetric key, IV, auth tag
	DecryptStream(ctx context.Context, streamID string, encryptedData []byte, encryptedKey []byte, iv []byte, authTag []byte, secretPlugin SecretPlugin) ([]byte, error)
}
