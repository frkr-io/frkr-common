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

// ValidateStreamName validates a stream name according to frkr standards.
// Stream names must:
// - Not be empty
// - Not exceed 100 characters
func ValidateStreamName(streamName string) error {
	if streamName == "" {
		return fmt.Errorf("stream name cannot be empty")
	}
	if len(streamName) > 100 {
		return fmt.Errorf("stream name cannot exceed 100 characters")
	}
	return nil
}

// NormalizeRetentionDays normalizes and validates retention days.
// - If days <= 0, returns 7 (default)
// - If days > 365, returns an error
// - Otherwise returns the provided value
func NormalizeRetentionDays(days int) (int, error) {
	if days <= 0 {
		return 7, nil
	}
	if days > 365 {
		return 0, fmt.Errorf("retention days cannot exceed 365")
	}
	return days, nil
}

