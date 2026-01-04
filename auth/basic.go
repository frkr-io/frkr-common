package auth

import (
	"database/sql"
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

// ValidateBasicAuthForStream validates Basic Auth and checks stream access.
// Currently accepts any non-empty basic auth credentials for development/testing.
// Production implementations should extend this to verify credentials against
// the database and check user access to the specified stream.
func ValidateBasicAuthForStream(authHeader, streamID string, db *sql.DB) bool {
	username, password, ok := ValidateBasicAuth(authHeader)
	if !ok {
		return false
	}

	// Accept any non-empty basic auth credentials
	// Note: Production implementations should verify credentials against database
	// and check user has access to the specified stream
	return username != "" && password != ""
}

