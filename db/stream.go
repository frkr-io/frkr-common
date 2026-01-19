package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/frkr-io/frkr-common/models"
	"github.com/frkr-io/frkr-common/util"
	"github.com/lib/pq"
)

// CreateStream creates a new stream for a tenant
func CreateStream(db *sql.DB, tenantID, streamName, description string, retentionDays int) (*models.Stream, error) {
	// Validate inputs
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}
	// Use shared stream name validation
	if err := util.ValidateStreamName(streamName); err != nil {
		return nil, err
	}
	// Normalize and validate retention days
	normalizedDays, err := util.NormalizeRetentionDays(retentionDays)
	if err != nil {
		return nil, err
	}
	retentionDays = normalizedDays

	// Generate topic name (Kafka Protocol compliant)
	topicName := GenerateTopicName(tenantID, streamName)
	
	var stream models.Stream
	
	err = db.QueryRow(`
		INSERT INTO streams (tenant_id, name, description, retention_days, topic, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		RETURNING id, tenant_id, name, description, status, retention_days, topic, created_at, updated_at, deleted_at
	`, tenantID, streamName, description, retentionDays, topicName).Scan(
		&stream.ID,
		&stream.TenantID,
		&stream.Name,
		&stream.Description,
		&stream.Status,
		&stream.RetentionDays,
		&stream.Topic,
		&stream.CreatedAt,
		&stream.UpdatedAt,
		&stream.DeletedAt,
	)
	
	if err != nil {
		// Check if it's a unique constraint violation
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" { // unique_violation
				return nil, fmt.Errorf("stream '%s' already exists for this tenant", streamName)
			}
		}
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}
	
	return &stream, nil
}

// GetStream retrieves a stream by ID or name
func GetStream(db *sql.DB, tenantID, streamIdentifier string) (*models.Stream, error) {
	var stream models.Stream
	
	// Try by ID first (UUID format)
	var query string
	var args []interface{}
	
	if len(streamIdentifier) == 36 && strings.Contains(streamIdentifier, "-") {
		// Looks like a UUID
		query = `
			SELECT id, tenant_id, name, description, status, retention_days, topic, created_at, updated_at, deleted_at
			FROM streams
			WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		`
		args = []interface{}{streamIdentifier, tenantID}
	} else {
		// Try by name
		query = `
			SELECT id, tenant_id, name, description, status, retention_days, topic, created_at, updated_at, deleted_at
			FROM streams
			WHERE name = $1 AND tenant_id = $2 AND deleted_at IS NULL
		`
		args = []interface{}{streamIdentifier, tenantID}
	}
	
	err := db.QueryRow(query, args...).Scan(
		&stream.ID,
		&stream.TenantID,
		&stream.Name,
		&stream.Description,
		&stream.Status,
		&stream.RetentionDays,
		&stream.Topic,
		&stream.CreatedAt,
		&stream.UpdatedAt,
		&stream.DeletedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("stream '%s' not found", streamIdentifier)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stream: %w", err)
	}
	
	return &stream, nil
}

// ListStreams lists all streams for a tenant
func ListStreams(db *sql.DB, tenantID string) ([]*models.Stream, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant ID cannot be empty")
	}

	rows, err := db.Query(`
		SELECT id, tenant_id, name, description, status, retention_days, topic, created_at, updated_at, deleted_at
		FROM streams
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query streams: %w", err)
	}
	defer rows.Close()
	
	var streams []*models.Stream
	for rows.Next() {
		var stream models.Stream
		err := rows.Scan(
			&stream.ID,
			&stream.TenantID,
			&stream.Name,
			&stream.Description,
			&stream.Status,
			&stream.RetentionDays,
			&stream.Topic,
			&stream.CreatedAt,
			&stream.UpdatedAt,
			&stream.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stream: %w", err)
		}
		streams = append(streams, &stream)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating streams: %w", err)
	}
	
	return streams, nil
}

