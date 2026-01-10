package plugins

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/frkr-io/frkr-common/auth"
	dbcommon "github.com/frkr-io/frkr-common/db"
	"golang.org/x/crypto/bcrypt"
)

// BasicAuthPlugin implements AuthPlugin for Basic Authentication (username/password)
type BasicAuthPlugin struct {
	db *sql.DB // Used for stream access checks
}

// NewBasicAuthPlugin creates a new BasicAuthPlugin
func NewBasicAuthPlugin(db *sql.DB) *BasicAuthPlugin {
	return &BasicAuthPlugin{db: db}
}

// ValidateRequest validates a Basic Auth token using SecretPlugin for credential lookup
func (p *BasicAuthPlugin) ValidateRequest(ctx context.Context, r *http.Request, secretPlugin SecretPlugin) (*AuthResult, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		return nil, fmt.Errorf("BasicAuthPlugin only supports basic auth")
	}

	token := strings.TrimPrefix(authHeader, "Basic ")
	// Parse Basic Auth header
	username, password, ok := auth.ValidateBasicAuth(token)
	if !ok {
		return nil, errors.New("invalid basic auth format")
	}

	if username == "" || password == "" {
		return nil, errors.New("username and password cannot be empty")
	}

	// Get password from SecretPlugin
	passwordHash, tenantID, err := secretPlugin.GetUserPassword(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Verify password
	// Try bcrypt first (for database-backed secrets), fall back to plain text (for K8s secrets)
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		// If bcrypt fails, try plain text comparison (for K8s secrets with plain passwords)
		if passwordHash != password {
			return nil, errors.New("invalid credentials")
		}
	}

	// If tenantID is empty from secret, we might need to query it from the database
	// For now, we'll use the username as a fallback identifier
	// In production, tenantID should always be available from the secret or user record
	if tenantID == "" && p.db != nil {
		// Try to get tenant ID from database user record
		var userTenantID sql.NullString
		err := p.db.QueryRowContext(ctx, `
			SELECT tenant_id FROM users 
			WHERE username = $1 AND deleted_at IS NULL
		`, username).Scan(&userTenantID)
		if err == nil && userTenantID.Valid {
			tenantID = userTenantID.String
		}
	}

	return &AuthResult{
		UserID:      username,
		TenantID:    tenantID,
		ClientType:  "user",
		AuthSource:  "basic",
		Roles:       []string{}, // Can be populated from user record if needed
		Permissions: []string{}, // Can be populated from user record if needed
	}, nil
}

// CanAccessStream checks if the user can access a specific stream
func (p *BasicAuthPlugin) CanAccessStream(ctx context.Context, authResult *AuthResult, streamID string, permission string) (bool, error) {
	if p.db == nil {
		return false, errors.New("database connection required for stream access checks")
	}

	// 1. Check AuthSource: this plugin only handles "basic" source users
	if authResult.AuthSource != "basic" {
		// Return error with filtered message or specific type if we want Composite to know it's "not my user"
		// For now, returning error is safe as Composite continues iteration on error.
		return false, fmt.Errorf("BasicAuthPlugin cannot authorize user from source: %s", authResult.AuthSource)
	}

	userID := authResult.UserID
	if userID == "" || streamID == "" {
		return false, errors.New("userID and streamID cannot be empty")
	}

	// Get user's tenant ID
	var userTenantID sql.NullString
	err := p.db.QueryRowContext(ctx, `
		SELECT tenant_id FROM users 
		WHERE username = $1 AND deleted_at IS NULL
	`, userID).Scan(&userTenantID)
	if err == sql.ErrNoRows {
		return false, fmt.Errorf("user %s not found", userID)
	}
	if err != nil {
		return false, fmt.Errorf("failed to query user: %w", err)
	}

	if !userTenantID.Valid {
		return false, fmt.Errorf("user %s has no tenant ID", userID)
	}

	// Get stream to verify tenant ownership and status
	stream, err := dbcommon.GetStream(p.db, userTenantID.String, streamID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil // Stream doesn't exist or user doesn't have access
		}
		return false, fmt.Errorf("failed to get stream: %w", err)
	}

	// Verify stream belongs to user's tenant
	if stream.TenantID != userTenantID.String {
		return false, nil
	}

	// Verify stream is active
	if stream.Status != "active" {
		return false, fmt.Errorf("stream %s is not active", streamID)
	}

	// Permission check: for now, we allow both read and write for all users in the tenant
	// This can be extended with role-based permissions later
	switch permission {
	case "read", "write":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported permission: %s", permission)
	}
}
