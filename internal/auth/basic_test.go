package auth

import (
	"context"
	"testing"
)

// MockUserStore implements UserStore for testing
type MockUserStore struct {
	users map[string]*UserInfo
	hashes map[string][]byte
}

func (m *MockUserStore) GetPasswordHash(username string) ([]byte, error) {
	return m.hashes[username], nil
}

func (m *MockUserStore) GetUserInfo(username string) (*UserInfo, error) {
	return m.users[username], nil
}

func TestBasicAuthPlugin_ValidateRequest(t *testing.T) {
	store := &MockUserStore{
		users: map[string]*UserInfo{
			"testuser": {
				UserID:   "user-123",
				TenantID: "tenant-456",
				Roles:    []string{"admin"},
			},
		},
		hashes: map[string][]byte{
			"testuser": []byte("password123"),
		},
	}

	plugin := &BasicAuthPlugin{
		UserStore: store,
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid credentials",
			token:   "dGVzdHVzZXI6cGFzc3dvcmQxMjM=", // base64("testuser:password123")
			wantErr: false,
		},
		{
			name:    "invalid password",
			token:   "dGVzdHVzZXI6d3Jvbmc=", // base64("testuser:wrong")
			wantErr: true,
		},
		{
			name:    "invalid format",
			token:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := plugin.ValidateRequest(context.Background(), tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result == nil {
				t.Error("ValidateRequest() returned nil result for valid token")
			}
		})
	}
}

