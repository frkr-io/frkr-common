package db

import (
	"database/sql"
	"fmt"
)

// GetStreamTopic retrieves the topic name for a stream (Kafka Protocol compliant)
func GetStreamTopic(db *sql.DB, streamName string) (string, error) {
	var topic string
	err := db.QueryRow(`
		SELECT topic
		FROM streams
		WHERE name = $1 AND deleted_at IS NULL
	`, streamName).Scan(&topic)
	if err != nil {
		return "", fmt.Errorf("stream '%s' not found: %w", streamName, err)
	}
	return topic, nil
}

