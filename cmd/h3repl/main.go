package main

import (
	"context"
	"fmt"
	"os"

	"github.com/andrew-d/h3repl/internal/cli"
	"github.com/andrew-d/h3repl/internal/config"
	"github.com/andrew-d/h3repl/internal/logging"
	"github.com/andrew-d/h3repl/internal/replication"
	"github.com/andrew-d/h3repl/internal/s3"
	"github.com/google/uuid"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	opts, err := cli.Parse()
	if err != nil {
		return err
	}

	logger := logging.Setup(opts.Verbose)

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	runID := uuid.New().String()
	logger.Info("h3repl starting", "run_id", runID, "config", opts.ConfigPath)

	engine := replication.NewEngine(cfg, logger, opts.DryRun, runID)

	// TODO: Initialize and configure S3 clients
	// For now, create dummy clients to demonstrate the structure
	for _, peer := range cfg.Peers {
		secretKey, err := peer.GetSecretAccessKey()
		if err != nil {
			logger.Error("failed to read secret key", "peer_id", peer.PeerID, "error", err)
			continue
		}

		s3Config := s3.Config{
			Endpoint:           peer.Endpoint,
			AccessKeyID:        peer.AccessKeyID,
			SecretAccessKey:    secretKey,
			InsecureSkipVerify: peer.InsecureSkipVerify,
		}

		client, err := s3.NewClient(s3Config)
		if err != nil {
			return fmt.Errorf("failed to create S3 client for peer %s: %w", peer.PeerID, err)
		}

		engine.AddRemoteS3Client(peer.PeerID, client)
	}

	// TODO: Initialize local S3 client
	// For now, use a dummy client
	localClient, err := s3.NewClient(s3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create local S3 client: %w", err)
	}
	engine.SetLocalS3Client(localClient)

	ctx := context.Background()
	if err := engine.Run(ctx); err != nil {
		return fmt.Errorf("replication failed: %w", err)
	}

	logger.Info("h3repl completed successfully", "run_id", runID)
	return nil
}
