package util

import "fmt"

// ValidateUsername validates a username according to frkr standards.
// Usernames must:
// - Not be empty
// - Not exceed 100 characters
// - Only contain alphanumeric characters, dashes, and underscores
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}
	if len(username) > 100 {
		return fmt.Errorf("username cannot exceed 100 characters")
	}
	// Basic validation: alphanumeric, dash, underscore only
	for _, r := range username {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return fmt.Errorf("username can only contain alphanumeric characters, dashes, and underscores")
		}
	}
	return nil
}

