// Package config provides configuration parsing for h3repl.
package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the root configuration structure
type Config struct {
	Peers []Peer `toml:"peer"`
}

// Peer represents a single peer configuration
type Peer struct {
	PeerID              string   `toml:"peer_id"`
	Endpoint            string   `toml:"endpoint"`
	AccessKeyID         string   `toml:"access_key_id"`
	SecretAccessKeyFile string   `toml:"secret_access_key_file"`
	InsecureSkipVerify  bool     `toml:"insecure_skip_verify"`
	Buckets             []Bucket `toml:"bucket"`
}

// Bucket represents a bucket configuration for a peer
type Bucket struct {
	Name       string `toml:"name"`
	QuotaBytes int64  `toml:"quota_bytes"`
}

// Load reads and parses a TOML configuration file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validate performs basic validation on the configuration
func (c *Config) validate() error {
	if len(c.Peers) == 0 {
		return fmt.Errorf("at least one peer must be configured")
	}

	peerIDs := make(map[string]bool)
	for i, peer := range c.Peers {
		if peer.PeerID == "" {
			return fmt.Errorf("peer[%d]: peer_id is required", i)
		}
		if peerIDs[peer.PeerID] {
			return fmt.Errorf("peer[%d]: duplicate peer_id '%s'", i, peer.PeerID)
		}
		peerIDs[peer.PeerID] = true

		if peer.Endpoint == "" {
			return fmt.Errorf("peer[%d]: endpoint is required", i)
		}
		if peer.AccessKeyID == "" {
			return fmt.Errorf("peer[%d]: access_key_id is required", i)
		}
		if peer.SecretAccessKeyFile == "" {
			return fmt.Errorf("peer[%d]: secret_access_key_file is required", i)
		}

		for j, bucket := range peer.Buckets {
			if bucket.Name == "" {
				return fmt.Errorf("peer[%d].bucket[%d]: name is required", i, j)
			}
			if bucket.QuotaBytes <= 0 {
				return fmt.Errorf("peer[%d].bucket[%d]: quota_bytes must be positive", i, j)
			}
		}
	}

	return nil
}

// GetSecretAccessKey reads the secret access key from the file
func (p *Peer) GetSecretAccessKey() (string, error) {
	data, err := os.ReadFile(p.SecretAccessKeyFile)
	if err != nil {
		return "", fmt.Errorf("failed to read secret access key file: %w", err)
	}
	return string(data), nil
}
