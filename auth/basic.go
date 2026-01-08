package auth

import (
	"encoding/base64"
	"strings"
)

// ValidateBasicAuth validates a Basic Auth header
func ValidateBasicAuth(authHeader string) (username, password string, ok bool) {
	if authHeader == "" {
		return "", "", false
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return "", "", false
	}

	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}

	return parts[0], parts[1], true
}

