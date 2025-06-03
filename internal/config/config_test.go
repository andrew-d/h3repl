package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {

	tests := []struct {
		name        string
		configData  string
		expectError bool
		expectedErr string
	}{
		{
			name: "valid config",
			configData: `
[[peer]]
peer_id = "peer1"
endpoint = "https://s3.example.com"
access_key_id = "ACCESS_KEY"
secret_access_key_file = "/path/to/secret"

[[peer.bucket]]
name = "test-bucket"
quota_bytes = 1073741824
`,
			expectError: false,
		},
		{
			name: "multiple peers",
			configData: `
[[peer]]
peer_id = "peer1"
endpoint = "https://s3.example.com"
access_key_id = "ACCESS_KEY"
secret_access_key_file = "/path/to/secret1"

[[peer.bucket]]
name = "bucket1"
quota_bytes = 1073741824

[[peer]]
peer_id = "peer2"
endpoint = "https://s3.other.com"
access_key_id = "OTHER_KEY"
secret_access_key_file = "/path/to/secret2"

[[peer.bucket]]
name = "bucket2"
quota_bytes = 2147483648
`,
			expectError: false,
		},
		{
			name: "missing peer_id",
			configData: `
[[peer]]
endpoint = "https://s3.example.com"
access_key_id = "ACCESS_KEY"
secret_access_key_file = "/path/to/secret"

[[peer.bucket]]
name = "test-bucket"
quota_bytes = 1073741824
`,
			expectError: true,
			expectedErr: "peer_id is required",
		},
		{
			name: "duplicate peer_id",
			configData: `
[[peer]]
peer_id = "peer1"
endpoint = "https://s3.example.com"
access_key_id = "ACCESS_KEY1"
secret_access_key_file = "/path/to/secret1"

[[peer.bucket]]
name = "bucket1"
quota_bytes = 1073741824

[[peer]]
peer_id = "peer1"
endpoint = "https://s3.other.com"
access_key_id = "ACCESS_KEY2"
secret_access_key_file = "/path/to/secret2"

[[peer.bucket]]
name = "bucket2"
quota_bytes = 2147483648
`,
			expectError: true,
			expectedErr: "duplicate peer_id",
		},
		{
			name: "missing endpoint",
			configData: `
[[peer]]
peer_id = "peer1"
access_key_id = "ACCESS_KEY"
secret_access_key_file = "/path/to/secret"

[[peer.bucket]]
name = "test-bucket"
quota_bytes = 1073741824
`,
			expectError: true,
			expectedErr: "endpoint is required",
		},
		{
			name: "zero quota_bytes",
			configData: `
[[peer]]
peer_id = "peer1"
endpoint = "https://s3.example.com"
access_key_id = "ACCESS_KEY"
secret_access_key_file = "/path/to/secret"

[[peer.bucket]]
name = "test-bucket"
quota_bytes = 0
`,
			expectError: true,
			expectedErr: "quota_bytes must be positive",
		},
		{
			name: "empty bucket name",
			configData: `
[[peer]]
peer_id = "peer1"
endpoint = "https://s3.example.com"
access_key_id = "ACCESS_KEY"
secret_access_key_file = "/path/to/secret"

[[peer.bucket]]
name = ""
quota_bytes = 1073741824
`,
			expectError: true,
			expectedErr: "name is required",
		},
		{
			name:        "no peers",
			configData:  ``,
			expectError: true,
			expectedErr: "at least one peer must be configured",
		},
		{
			name: "invalid toml",
			configData: `
[[peer]
peer_id = "invalid
`,
			expectError: true,
			expectedErr: "failed to parse config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.toml")

			err := os.WriteFile(configPath, []byte(tt.configData), 0644)
			if err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			config, err := Load(configPath)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
					return
				}
				if tt.expectedErr != "" && err.Error() == "" {
					t.Errorf("expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config == nil {
				t.Error("expected config but got nil")
				return
			}

			// Verify the config was parsed correctly for valid cases
			if len(config.Peers) == 0 {
				t.Error("expected at least one peer")
			}
		})
	}
}

func TestLoadNonexistentFile(t *testing.T) {

	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestPeerGetSecretAccessKey(t *testing.T) {

	// Create a temporary file with secret
	tmpDir := t.TempDir()
	secretFile := filepath.Join(tmpDir, "secret")
	secretContent := "my-secret-key"

	err := os.WriteFile(secretFile, []byte(secretContent), 0600)
	if err != nil {
		t.Fatalf("failed to create test secret file: %v", err)
	}

	peer := Peer{
		SecretAccessKeyFile: secretFile,
	}

	secret, err := peer.GetSecretAccessKey()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if secret != secretContent {
		t.Errorf("expected secret %q, got %q", secretContent, secret)
	}
}

func TestPeerGetSecretAccessKeyNonexistent(t *testing.T) {

	peer := Peer{
		SecretAccessKeyFile: "/nonexistent/secret",
	}

	_, err := peer.GetSecretAccessKey()
	if err == nil {
		t.Error("expected error for nonexistent secret file")
	}
}
