package plugins

import (
	"testing"
)

// Test SecretPlugin interface compliance (compile-time check)
func TestSecretPlugin_InterfaceCompliance(t *testing.T) {
	var _ SecretPlugin = (*DatabaseSecretPlugin)(nil)
	var _ SecretPlugin = (*K8sSecretPlugin)(nil)
}

// Test SecretType constants
func TestSecretTypeConstants(t *testing.T) {
	if SecretTypeUserPassword != "user_password" {
		t.Errorf("SecretTypeUserPassword = %v, want user_password", SecretTypeUserPassword)
	}
	if SecretTypeClientSecret != "client_secret" {
		t.Errorf("SecretTypeClientSecret = %v, want client_secret", SecretTypeClientSecret)
	}
	if SecretTypeEncryptionKey != "encryption_key" {
		t.Errorf("SecretTypeEncryptionKey = %v, want encryption_key", SecretTypeEncryptionKey)
	}
	if SecretTypeServiceAccount != "service_account" {
		t.Errorf("SecretTypeServiceAccount = %v, want service_account", SecretTypeServiceAccount)
	}
}
