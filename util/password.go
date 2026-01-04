package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GeneratePassword generates a cryptographically secure random password.
// The password is 32 bytes of random data encoded as base64 URL-safe string,
// resulting in approximately 43 characters.
func GeneratePassword() (string, error) {
	passwordBytes := make([]byte, 32)
	if _, err := rand.Read(passwordBytes); err != nil {
		return "", fmt.Errorf("failed to generate password: %w", err)
	}
	return base64.URLEncoding.EncodeToString(passwordBytes), nil
}

