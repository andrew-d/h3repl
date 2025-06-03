// Package status provides status object management for h3repl.
package status

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/klauspost/compress/zstd"
)

// Status represents the status object stored in the source bucket
type Status struct {
	PeerID       string       `json:"peer_id"`
	SourceBucket string       `json:"source_bucket"`
	RunID        string       `json:"run_id"`
	LastSync     int64        `json:"last_sync"`
	TotalObjects int          `json:"total_objects"`
	TotalBytes   int64        `json:"total_bytes"`
	Objects      []ObjectInfo `json:"objects"`
}

// ObjectInfo represents information about a replicated object
type ObjectInfo struct {
	Key     string `json:"key"`
	Version string `json:"version"`
	Size    int64  `json:"size"`
	MTime   int64  `json:"mtime"`
}

// NewStatus creates a new status object
func NewStatus(peerID, sourceBucket, runID string) *Status {
	return &Status{
		PeerID:       peerID,
		SourceBucket: sourceBucket,
		RunID:        runID,
		LastSync:     time.Now().Unix(),
		Objects:      make([]ObjectInfo, 0),
	}
}

// AddObject adds an object to the status
func (s *Status) AddObject(key, version string, size, mtime int64) {
	s.Objects = append(s.Objects, ObjectInfo{
		Key:     key,
		Version: version,
		Size:    size,
		MTime:   mtime,
	})
	s.TotalObjects = len(s.Objects)
	s.TotalBytes += size
}

// RemoveObject removes an object from the status
func (s *Status) RemoveObject(key string) {
	for i, obj := range s.Objects {
		if obj.Key == key {
			s.TotalBytes -= obj.Size
			s.Objects = append(s.Objects[:i], s.Objects[i+1:]...)
			s.TotalObjects = len(s.Objects)
			break
		}
	}
}

// UpdateLastSync updates the last sync timestamp
func (s *Status) UpdateLastSync() {
	s.LastSync = time.Now().Unix()
}

// Marshal serializes the status object to zstd-compressed JSON
func (s *Status) Marshal() ([]byte, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal status to JSON: %w", err)
	}

	var buf bytes.Buffer
	encoder, err := zstd.NewWriter(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
	}
	defer encoder.Close()

	if _, err := encoder.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to compress status: %w", err)
	}

	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to close zstd encoder: %w", err)
	}

	return buf.Bytes(), nil
}

// Unmarshal deserializes a status object from zstd-compressed JSON
func Unmarshal(data []byte) (*Status, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
	}
	defer decoder.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(decoder); err != nil {
		return nil, fmt.Errorf("failed to decompress status: %w", err)
	}

	var status Status
	if err := json.Unmarshal(buf.Bytes(), &status); err != nil {
		return nil, fmt.Errorf("failed to unmarshal status from JSON: %w", err)
	}

	return &status, nil
}

// StatusKey returns the S3 key for storing the status object
func StatusKey(peerID string) string {
	return fmt.Sprintf(".replication-status/%s.json.zst", peerID)
}
