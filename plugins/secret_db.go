package plugins

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// DatabaseSecretPlugin implements SecretPlugin using PostgreSQL database
type DatabaseSecretPlugin struct {
	db *sql.DB
}

// NewDatabaseSecretPlugin creates a new DatabaseSecretPlugin
func NewDatabaseSecretPlugin(db *sql.DB) (*DatabaseSecretPlugin, error) {
	if db == nil {
		return nil, errors.New("database connection cannot be nil")
	}
	return &DatabaseSecretPlugin{db: db}, nil
}

// GetUserPassword retrieves a user's password hash from the database
// Queries the users table: SELECT password_hash, tenant_id FROM users WHERE username = $1 AND deleted_at IS NULL
func (p *DatabaseSecretPlugin) GetUserPassword(ctx context.Context, username string) (passwordHash string, tenantID string, err error) {
	if username == "" {
		return "", "", errors.New("username cannot be empty")
	}

	var passwordHashVal, tenantIDVal sql.NullString
	err = p.db.QueryRowContext(ctx, `
		SELECT password_hash, tenant_id 
		FROM users 
		WHERE username = $1 AND deleted_at IS NULL
	`, username).Scan(&passwordHashVal, &tenantIDVal)

	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("user %s not found", username)
	}
	if err != nil {
		return "", "", fmt.Errorf("failed to query user: %w", err)
	}

	if !passwordHashVal.Valid {
		return "", "", fmt.Errorf("user %s has no password hash", username)
	}

	passwordHash = passwordHashVal.String
	if tenantIDVal.Valid {
		tenantID = tenantIDVal.String
	}

	return passwordHash, tenantID, nil
}

// GetClientSecret retrieves a client's secret from the database
func (p *DatabaseSecretPlugin) GetClientSecret(ctx context.Context, clientID string) (clientSecret string, tenantID string, err error) {
	if clientID == "" {
		return "", "", errors.New("clientID cannot be empty")
	}

	var secretVal, tenantIDVal sql.NullString
	err = p.db.QueryRowContext(ctx, `
		SELECT client_secret, tenant_id 
		FROM clients 
		WHERE client_id = $1 AND deleted_at IS NULL
	`, clientID).Scan(&secretVal, &tenantIDVal)

	if err == sql.ErrNoRows {
		return "", "", fmt.Errorf("client %s not found", clientID)
	}
	if err != nil {
		// Check if table doesn't exist (42P01 is PostgreSQL error code for undefined_table)
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "undefined_table") {
			return "", "", fmt.Errorf("clients table does not exist - migration may not have been run")
		}
		return "", "", fmt.Errorf("failed to query client: %w", err)
	}

	if !secretVal.Valid {
		return "", "", fmt.Errorf("client %s has no secret", clientID)
	}

	clientSecret = secretVal.String
	if tenantIDVal.Valid {
		tenantID = tenantIDVal.String
	}

	return clientSecret, tenantID, nil
}

// GetEncryptionKey retrieves encryption keys from the database
func (p *DatabaseSecretPlugin) GetEncryptionKey(ctx context.Context, streamID string) (publicKey []byte, privateKey []byte, err error) {
	if streamID == "" {
		return nil, nil, errors.New("streamID cannot be empty")
	}

	var publicKeyVal, privateKeyVal sql.NullString
	err = p.db.QueryRowContext(ctx, `
		SELECT public_key, private_key 
		FROM stream_keys 
		WHERE stream_id = $1 AND deleted_at IS NULL
	`, streamID).Scan(&publicKeyVal, &privateKeyVal)

	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("encryption key for stream %s not found", streamID)
	}
	if err != nil {
		// Check if table doesn't exist
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "undefined_table") {
			return nil, nil, fmt.Errorf("stream_keys table does not exist - migration may not have been run")
		}
		return nil, nil, fmt.Errorf("failed to query encryption key: %w", err)
	}

	if !publicKeyVal.Valid {
		return nil, nil, fmt.Errorf("stream %s has no public key", streamID)
	}

	publicKey = []byte(publicKeyVal.String)
	if privateKeyVal.Valid {
		privateKey = []byte(privateKeyVal.String)
	}

	return publicKey, privateKey, nil
}

// GetSecret retrieves a generic secret by type and identifier
func (p *DatabaseSecretPlugin) GetSecret(ctx context.Context, secretType string, identifier string) ([]byte, error) {
	switch secretType {
	case SecretTypeUserPassword:
		password, _, err := p.GetUserPassword(ctx, identifier)
		return []byte(password), err
	case SecretTypeClientSecret:
		secret, _, err := p.GetClientSecret(ctx, identifier)
		return []byte(secret), err
	case SecretTypeEncryptionKey:
		publicKey, _, err := p.GetEncryptionKey(ctx, identifier)
		return publicKey, err
	default:
		return nil, fmt.Errorf("unsupported secret type: %s", secretType)
	}
}
