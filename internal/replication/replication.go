// Package replication provides the core replication engine for h3repl.
package replication

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andrew-d/h3repl/internal/config"
	"github.com/andrew-d/h3repl/internal/s3"
	"github.com/andrew-d/h3repl/internal/status"
)

// Engine handles the replication process
type Engine struct {
	config    *config.Config
	logger    *slog.Logger
	dryRun    bool
	runID     string
	localS3   s3.Client
	remoteS3s map[string]s3.Client
}

// NewEngine creates a new replication engine
func NewEngine(cfg *config.Config, logger *slog.Logger, dryRun bool, runID string) *Engine {
	return &Engine{
		config:    cfg,
		logger:    logger,
		dryRun:    dryRun,
		runID:     runID,
		remoteS3s: make(map[string]s3.Client),
	}
}

// SetLocalS3Client sets the local S3 client
func (e *Engine) SetLocalS3Client(client s3.Client) {
	e.localS3 = client
}

// AddRemoteS3Client adds a remote S3 client for a peer
func (e *Engine) AddRemoteS3Client(peerID string, client s3.Client) {
	e.remoteS3s[peerID] = client
}

// Run executes the replication process
func (e *Engine) Run(ctx context.Context) error {
	e.logger.Info("starting replication run", "run_id", e.runID, "dry_run", e.dryRun)

	for _, peer := range e.config.Peers {
		if err := e.replicatePeer(ctx, peer); err != nil {
			e.logger.Error("peer replication failed", "peer_id", peer.PeerID, "error", err)
			return fmt.Errorf("failed to replicate peer %s: %w", peer.PeerID, err)
		}
	}

	e.logger.Info("replication run completed", "run_id", e.runID)
	return nil
}

// replicatePeer replicates all buckets for a specific peer
func (e *Engine) replicatePeer(ctx context.Context, peer config.Peer) error {
	e.logger.Info("starting peer replication", "peer_id", peer.PeerID)

	remoteClient, exists := e.remoteS3s[peer.PeerID]
	if !exists {
		return fmt.Errorf("no remote S3 client configured for peer %s", peer.PeerID)
	}

	for _, bucket := range peer.Buckets {
		if err := e.replicateBucket(ctx, peer, bucket, remoteClient); err != nil {
			e.logger.Error("bucket replication failed",
				"peer_id", peer.PeerID,
				"bucket", bucket.Name,
				"error", err)
			return fmt.Errorf("failed to replicate bucket %s: %w", bucket.Name, err)
		}
	}

	e.logger.Info("peer replication completed", "peer_id", peer.PeerID)
	return nil
}

// replicateBucket replicates a specific bucket
func (e *Engine) replicateBucket(ctx context.Context, peer config.Peer, bucket config.Bucket, remoteClient s3.Client) error {
	e.logger.Info("starting bucket replication",
		"peer_id", peer.PeerID,
		"bucket", bucket.Name,
		"quota_bytes", bucket.QuotaBytes)

	// Load existing status
	currentStatus, err := e.loadStatus(ctx, remoteClient, peer.PeerID, bucket.Name)
	if err != nil {
		e.logger.Warn("failed to load existing status",
			"peer_id", peer.PeerID,
			"bucket", bucket.Name,
			"error", err)
		currentStatus = status.NewStatus(peer.PeerID, bucket.Name, e.runID)
	}

	// List objects in remote bucket
	objects, err := remoteClient.ListObjects(ctx, bucket.Name, "")
	if err != nil {
		return fmt.Errorf("failed to list objects in bucket %s: %w", bucket.Name, err)
	}

	e.logger.Info("listed objects",
		"peer_id", peer.PeerID,
		"bucket", bucket.Name,
		"object_count", len(objects))

	// TODO: Implement actual replication logic:
	// 1. Filter objects by ReplicationPriority tag
	// 2. Select objects based on priority and quota
	// 3. Download and upload objects
	// 4. Delete objects no longer eligible
	// 5. Update status periodically

	// Save status
	if !e.dryRun {
		if err := e.saveStatus(ctx, remoteClient, currentStatus); err != nil {
			e.logger.Error("failed to save status",
				"peer_id", peer.PeerID,
				"bucket", bucket.Name,
				"error", err)
		}
	}

	e.logger.Info("bucket replication completed",
		"peer_id", peer.PeerID,
		"bucket", bucket.Name)

	return nil
}

// loadStatus loads the current status for a peer
func (e *Engine) loadStatus(ctx context.Context, remoteClient s3.Client, peerID, bucketName string) (*status.Status, error) {
	statusKey := status.StatusKey(peerID)

	reader, _, err := remoteClient.GetObject(ctx, bucketName, statusKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get status object: %w", err)
	}
	defer reader.Close()

	// TODO: Read data from reader and unmarshal
	// For now, return a new status
	return status.NewStatus(peerID, bucketName, e.runID), nil
}

// saveStatus saves the current status for a peer
func (e *Engine) saveStatus(ctx context.Context, remoteClient s3.Client, st *status.Status) error {
	statusKey := status.StatusKey(st.PeerID)

	data, err := st.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	// TODO: Implement saving status to S3
	e.logger.Debug("would save status",
		"key", statusKey,
		"size", len(data))

	return nil
}
