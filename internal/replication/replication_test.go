package replication

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/andrew-d/h3repl/internal/config"
	"github.com/andrew-d/h3repl/internal/s3"
)

// mockS3Client implements s3.Client for testing
type mockS3Client struct {
	objects      []s3.Object
	getObjectErr error
	putObjectErr error
	listErr      error
}

func (m *mockS3Client) ListObjects(ctx context.Context, bucket string, prefix string) ([]s3.Object, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.objects, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, bucket string, key string) (io.ReadCloser, *s3.ObjectInfo, error) {
	if m.getObjectErr != nil {
		return nil, nil, m.getObjectErr
	}
	return io.NopCloser(strings.NewReader("{}")), &s3.ObjectInfo{}, nil
}

func (m *mockS3Client) PutObject(ctx context.Context, bucket string, key string, data io.Reader, size int64, contentType string, tags map[string]string) error {
	return m.putObjectErr
}

func (m *mockS3Client) DeleteObject(ctx context.Context, bucket string, key string) error {
	return nil
}

func (m *mockS3Client) GetObjectTags(ctx context.Context, bucket string, key string) (map[string]string, error) {
	return nil, nil
}

func (m *mockS3Client) GetObjectInfo(ctx context.Context, bucket string, key string) (*s3.ObjectInfo, error) {
	return nil, nil
}

func TestNewEngine(t *testing.T) {

	cfg := &config.Config{
		Peers: []config.Peer{
			{
				PeerID:   "test-peer",
				Endpoint: "https://test.com",
			},
		},
	}
	logger := slog.Default()
	runID := "test-run-123"

	engine := NewEngine(cfg, logger, false, runID)

	if engine == nil {
		t.Error("expected engine but got nil")
		return
	}

	if engine.config != cfg {
		t.Error("engine config not set correctly")
	}

	if engine.logger != logger {
		t.Error("engine logger not set correctly")
	}

	if engine.dryRun != false {
		t.Error("engine dryRun not set correctly")
	}

	if engine.runID != runID {
		t.Error("engine runID not set correctly")
	}

	if engine.remoteS3s == nil {
		t.Error("remoteS3s map not initialized")
	}
}

func TestEngineSetLocalS3Client(t *testing.T) {

	engine := NewEngine(&config.Config{}, slog.Default(), false, "test")
	client := &mockS3Client{}

	engine.SetLocalS3Client(client)

	if engine.localS3 != client {
		t.Error("local S3 client not set correctly")
	}
}

func TestEngineAddRemoteS3Client(t *testing.T) {

	engine := NewEngine(&config.Config{}, slog.Default(), false, "test")
	client := &mockS3Client{}
	peerID := "test-peer"

	engine.AddRemoteS3Client(peerID, client)

	if engine.remoteS3s[peerID] != client {
		t.Error("remote S3 client not added correctly")
	}
}

func TestEngineRun(t *testing.T) {

	tests := []struct {
		name        string
		config      *config.Config
		clients     map[string]s3.Client
		expectError bool
	}{
		{
			name: "successful replication",
			config: &config.Config{
				Peers: []config.Peer{
					{
						PeerID:   "peer1",
						Endpoint: "https://test.com",
						Buckets: []config.Bucket{
							{Name: "bucket1", QuotaBytes: 1000},
						},
					},
				},
			},
			clients: map[string]s3.Client{
				"peer1": &mockS3Client{
					objects: []s3.Object{
						{Key: "test-object", Size: 100},
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing remote client",
			config: &config.Config{
				Peers: []config.Peer{
					{
						PeerID:   "peer1",
						Endpoint: "https://test.com",
						Buckets: []config.Bucket{
							{Name: "bucket1", QuotaBytes: 1000},
						},
					},
				},
			},
			clients:     map[string]s3.Client{}, // No client for peer1
			expectError: true,
		},
		{
			name: "list objects error",
			config: &config.Config{
				Peers: []config.Peer{
					{
						PeerID:   "peer1",
						Endpoint: "https://test.com",
						Buckets: []config.Bucket{
							{Name: "bucket1", QuotaBytes: 1000},
						},
					},
				},
			},
			clients: map[string]s3.Client{
				"peer1": &mockS3Client{
					listErr: errors.New("list error"),
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEngine(tt.config, slog.Default(), false, "test-run")

			// Add remote clients
			for peerID, client := range tt.clients {
				engine.AddRemoteS3Client(peerID, client)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := engine.Run(ctx)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestEngineRunDryRun(t *testing.T) {

	cfg := &config.Config{
		Peers: []config.Peer{
			{
				PeerID:   "peer1",
				Endpoint: "https://test.com",
				Buckets: []config.Bucket{
					{Name: "bucket1", QuotaBytes: 1000},
				},
			},
		},
	}

	engine := NewEngine(cfg, slog.Default(), true, "test-run") // dry run = true
	engine.AddRemoteS3Client("peer1", &mockS3Client{
		objects: []s3.Object{
			{Key: "test-object", Size: 100},
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := engine.Run(ctx)
	if err != nil {
		t.Errorf("unexpected error in dry run: %v", err)
	}
}
