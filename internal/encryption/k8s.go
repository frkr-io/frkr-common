package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/frkr-io/frkr-common/internal/plugins"
)

// K8sEncryptionPlugin implements encryption using Kubernetes secrets
type K8sEncryptionPlugin struct {
	// KeyStore provides encryption keys from Kubernetes secrets
	KeyStore KeyStore
}

// KeyStore provides encryption key storage
type KeyStore interface {
	GetPublicKey(streamID string) (*rsa.PublicKey, error)
	GetPrivateKey(streamID string) (*rsa.PrivateKey, error)
}

// EncryptStream encrypts data using hybrid encryption (RSA + AES)
func (p *K8sEncryptionPlugin) EncryptStream(streamID string, payload []byte) ([]byte, []byte, []byte, []byte, error) {
	// Get stream public key
	pubKey, err := p.KeyStore.GetPublicKey(streamID)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Generate random symmetric key (AES-256)
	symmetricKey := make([]byte, 32) // 256 bits
	if _, err := io.ReadFull(rand.Reader, symmetricKey); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate symmetric key: %w", err)
	}

	// Generate random IV (96 bits for AES-GCM)
	iv := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt payload with AES-256-GCM
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	encryptedData := aesgcm.Seal(nil, iv, payload, nil)
	authTag := encryptedData[len(encryptedData)-16:] // Last 16 bytes are auth tag
	encryptedData = encryptedData[:len(encryptedData)-16] // Remove auth tag from data

	// Encrypt symmetric key with RSA-OAEP
	encryptedKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pubKey,
		symmetricKey,
		nil,
	)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to encrypt symmetric key: %w", err)
	}

	return encryptedData, encryptedKey, iv, authTag, nil
}

// DecryptStream decrypts data using hybrid decryption
func (p *K8sEncryptionPlugin) DecryptStream(streamID string, encryptedData []byte, encryptedKey []byte, iv []byte, authTag []byte) ([]byte, error) {
	// Get stream private key
	privKey, err := p.KeyStore.GetPrivateKey(streamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	// Decrypt symmetric key with RSA-OAEP
	symmetricKey, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		privKey,
		encryptedKey,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt symmetric key: %w", err)
	}

	// Decrypt payload with AES-256-GCM
	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Combine encrypted data and auth tag
	ciphertext := append(encryptedData, authTag...)

	decryptedData, err := aesgcm.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return decryptedData, nil
}

// GetPublicKey retrieves the public key for a stream
func (p *K8sEncryptionPlugin) GetPublicKey(streamID string) ([]byte, error) {
	pubKey, err := p.KeyStore.GetPublicKey(streamID)
	if err != nil {
		return nil, err
	}

	// Encode public key as PEM
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return pubKeyPEM, nil
}

// GetPrivateKey retrieves the private key for a stream
func (p *K8sEncryptionPlugin) GetPrivateKey(streamID string) ([]byte, error) {
	privKey, err := p.KeyStore.GetPrivateKey(streamID)
	if err != nil {
		return nil, err
	}

	// Encode private key as PEM
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return privKeyPEM, nil
}

