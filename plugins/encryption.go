package plugins

// EncryptionPlugin is the interface for encryption plugins
type EncryptionPlugin interface {
	// EncryptStream encrypts data for a specific stream
	// Returns: encrypted data, encrypted symmetric key, IV, auth tag
	EncryptStream(streamID string, payload []byte) (encryptedData []byte, encryptedKey []byte, iv []byte, authTag []byte, err error)
	
	// DecryptStream decrypts data for a specific stream
	// Requires: encrypted data, encrypted symmetric key, IV, auth tag
	DecryptStream(streamID string, encryptedData []byte, encryptedKey []byte, iv []byte, authTag []byte) ([]byte, error)
	
	// GetPublicKey retrieves the public key for a stream (for encryption)
	GetPublicKey(streamID string) ([]byte, error)
	
	// GetPrivateKey retrieves the private key for a stream (for decryption)
	GetPrivateKey(streamID string) ([]byte, error)
}

