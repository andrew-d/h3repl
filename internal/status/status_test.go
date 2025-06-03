package status

import (
	"testing"
	"time"
)

func TestNewStatus(t *testing.T) {

	peerID := "test-peer"
	sourceBucket := "test-bucket"
	runID := "run-123"

	status := NewStatus(peerID, sourceBucket, runID)

	if status == nil {
		t.Error("expected status but got nil")
		return
	}

	if status.PeerID != peerID {
		t.Errorf("expected PeerID %q, got %q", peerID, status.PeerID)
	}

	if status.SourceBucket != sourceBucket {
		t.Errorf("expected SourceBucket %q, got %q", sourceBucket, status.SourceBucket)
	}

	if status.RunID != runID {
		t.Errorf("expected RunID %q, got %q", runID, status.RunID)
	}

	if status.TotalObjects != 0 {
		t.Errorf("expected TotalObjects 0, got %d", status.TotalObjects)
	}

	if status.TotalBytes != 0 {
		t.Errorf("expected TotalBytes 0, got %d", status.TotalBytes)
	}

	if status.Objects == nil {
		t.Error("expected Objects slice to be initialized")
	}

	if len(status.Objects) != 0 {
		t.Errorf("expected empty Objects slice, got length %d", len(status.Objects))
	}

	// Verify LastSync is recent (within last 10 seconds)
	now := time.Now().Unix()
	if status.LastSync < now-10 || status.LastSync > now {
		t.Errorf("expected LastSync to be recent, got %d (now: %d)", status.LastSync, now)
	}
}

func TestStatusAddObject(t *testing.T) {

	status := NewStatus("peer1", "bucket1", "run1")

	key := "test/object.txt"
	version := "version123"
	size := int64(1024)
	mtime := time.Now().Unix()

	status.AddObject(key, version, size, mtime)

	if status.TotalObjects != 1 {
		t.Errorf("expected TotalObjects 1, got %d", status.TotalObjects)
	}

	if status.TotalBytes != size {
		t.Errorf("expected TotalBytes %d, got %d", size, status.TotalBytes)
	}

	if len(status.Objects) != 1 {
		t.Errorf("expected 1 object, got %d", len(status.Objects))
	}

	obj := status.Objects[0]
	if obj.Key != key {
		t.Errorf("expected Key %q, got %q", key, obj.Key)
	}

	if obj.Version != version {
		t.Errorf("expected Version %q, got %q", version, obj.Version)
	}

	if obj.Size != size {
		t.Errorf("expected Size %d, got %d", size, obj.Size)
	}

	if obj.MTime != mtime {
		t.Errorf("expected MTime %d, got %d", mtime, obj.MTime)
	}

	// Add another object
	status.AddObject("test/object2.txt", "version456", 2048, mtime+1)

	if status.TotalObjects != 2 {
		t.Errorf("expected TotalObjects 2, got %d", status.TotalObjects)
	}

	if status.TotalBytes != size+2048 {
		t.Errorf("expected TotalBytes %d, got %d", size+2048, status.TotalBytes)
	}
}

func TestStatusRemoveObject(t *testing.T) {

	status := NewStatus("peer1", "bucket1", "run1")

	// Add some objects
	status.AddObject("object1.txt", "v1", 1000, 12345)
	status.AddObject("object2.txt", "v2", 2000, 12346)
	status.AddObject("object3.txt", "v3", 3000, 12347)

	if status.TotalObjects != 3 {
		t.Errorf("expected 3 objects, got %d", status.TotalObjects)
	}

	if status.TotalBytes != 6000 {
		t.Errorf("expected 6000 bytes, got %d", status.TotalBytes)
	}

	// Remove middle object
	status.RemoveObject("object2.txt")

	if status.TotalObjects != 2 {
		t.Errorf("expected 2 objects after removal, got %d", status.TotalObjects)
	}

	if status.TotalBytes != 4000 {
		t.Errorf("expected 4000 bytes after removal, got %d", status.TotalBytes)
	}

	// Verify correct objects remain
	keys := make(map[string]bool)
	for _, obj := range status.Objects {
		keys[obj.Key] = true
	}

	if !keys["object1.txt"] {
		t.Error("object1.txt should still be present")
	}

	if keys["object2.txt"] {
		t.Error("object2.txt should have been removed")
	}

	if !keys["object3.txt"] {
		t.Error("object3.txt should still be present")
	}

	// Remove non-existent object (should not change anything)
	status.RemoveObject("nonexistent.txt")

	if status.TotalObjects != 2 {
		t.Errorf("expected 2 objects after removing nonexistent, got %d", status.TotalObjects)
	}
}

func TestStatusUpdateLastSync(t *testing.T) {

	status := NewStatus("peer1", "bucket1", "run1")
	originalSync := status.LastSync

	// Sleep a full second to ensure time difference at unix timestamp granularity
	time.Sleep(time.Second + 10*time.Millisecond)

	status.UpdateLastSync()

	if status.LastSync <= originalSync {
		t.Errorf("expected LastSync to be updated, got %d (was %d)", status.LastSync, originalSync)
	}

	// Verify it's recent
	now := time.Now().Unix()
	if status.LastSync < now-10 || status.LastSync > now {
		t.Errorf("expected LastSync to be recent, got %d (now: %d)", status.LastSync, now)
	}
}

func TestStatusMarshalUnmarshal(t *testing.T) {

	// Create a status with some data
	status := NewStatus("test-peer", "test-bucket", "test-run")
	status.AddObject("file1.txt", "v1", 1024, 1640995200)
	status.AddObject("file2.txt", "v2", 2048, 1640995300)

	// Marshal to compressed JSON
	data, err := status.Marshal()
	if err != nil {
		t.Errorf("unexpected error marshaling: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected marshaled data but got empty slice")
	}

	// Unmarshal back
	unmarshaled, err := Unmarshal(data)
	if err != nil {
		t.Errorf("unexpected error unmarshaling: %v", err)
	}

	if unmarshaled == nil {
		t.Error("expected unmarshaled status but got nil")
		return
	}

	// Verify all fields match
	if unmarshaled.PeerID != status.PeerID {
		t.Errorf("expected PeerID %q, got %q", status.PeerID, unmarshaled.PeerID)
	}

	if unmarshaled.SourceBucket != status.SourceBucket {
		t.Errorf("expected SourceBucket %q, got %q", status.SourceBucket, unmarshaled.SourceBucket)
	}

	if unmarshaled.RunID != status.RunID {
		t.Errorf("expected RunID %q, got %q", status.RunID, unmarshaled.RunID)
	}

	if unmarshaled.TotalObjects != status.TotalObjects {
		t.Errorf("expected TotalObjects %d, got %d", status.TotalObjects, unmarshaled.TotalObjects)
	}

	if unmarshaled.TotalBytes != status.TotalBytes {
		t.Errorf("expected TotalBytes %d, got %d", status.TotalBytes, unmarshaled.TotalBytes)
	}

	if len(unmarshaled.Objects) != len(status.Objects) {
		t.Errorf("expected %d objects, got %d", len(status.Objects), len(unmarshaled.Objects))
	}

	// Check objects match
	for i, obj := range status.Objects {
		uObj := unmarshaled.Objects[i]
		if uObj.Key != obj.Key {
			t.Errorf("object %d: expected Key %q, got %q", i, obj.Key, uObj.Key)
		}
		if uObj.Version != obj.Version {
			t.Errorf("object %d: expected Version %q, got %q", i, obj.Version, uObj.Version)
		}
		if uObj.Size != obj.Size {
			t.Errorf("object %d: expected Size %d, got %d", i, obj.Size, uObj.Size)
		}
		if uObj.MTime != obj.MTime {
			t.Errorf("object %d: expected MTime %d, got %d", i, obj.MTime, uObj.MTime)
		}
	}
}

func TestUnmarshalInvalidData(t *testing.T) {

	// Test with invalid compressed data
	_, err := Unmarshal([]byte("invalid data"))
	if err == nil {
		t.Error("expected error unmarshaling invalid data")
	}

	// Test with empty data
	_, err = Unmarshal([]byte{})
	if err == nil {
		t.Error("expected error unmarshaling empty data")
	}
}

func TestStatusKey(t *testing.T) {

	tests := []struct {
		peerID   string
		expected string
	}{
		{
			peerID:   "peer1",
			expected: ".replication-status/peer1.json.zst",
		},
		{
			peerID:   "another-peer-id",
			expected: ".replication-status/another-peer-id.json.zst",
		},
		{
			peerID:   "peer_with_underscores",
			expected: ".replication-status/peer_with_underscores.json.zst",
		},
	}

	for _, tt := range tests {
		t.Run(tt.peerID, func(t *testing.T) {
			key := StatusKey(tt.peerID)
			if key != tt.expected {
				t.Errorf("expected key %q, got %q", tt.expected, key)
			}
		})
	}
}
