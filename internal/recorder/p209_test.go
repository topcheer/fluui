package recorder

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// P209: recorder coverage for sub-80% functions

func TestElapsedLocked_NotStarted_P209(t *testing.T) {
	r := NewRecorder()
	// start is zero → should return 0
	if r.elapsedLocked() != 0 {
		t.Error("should return 0 when not started")
	}
}

func TestElapsedLocked_Started_P209(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	time.Sleep(2 * time.Millisecond)
	elapsed := r.elapsedLocked()
	if elapsed <= 0 {
		t.Error("should return positive elapsed after start")
	}
	r.Stop()
}

func TestLoad_FileError_P209(t *testing.T) {
	_, err := Load("/nonexistent/path/file.json")
	if err == nil {
		t.Error("should return error for missing file")
	}
}

func TestLoad_InvalidJSON_P209(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(path, []byte("{invalid json}"), 0644)
	_, err := Load(path)
	if err == nil {
		t.Error("should return error for invalid JSON")
	}
}

func TestLoad_Valid_P209(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('x', 0, 0)
	r.Stop()
	data, _ := r.SaveBytes()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "valid.json")
	os.WriteFile(path, data, 0644)
	p, err := Load(path)
	if err != nil {
		t.Errorf("Load error: %v", err)
	}
	if p == nil {
		t.Error("should return non-nil player")
	}
}

func TestLoadFromBytes_Invalid_P209(t *testing.T) {
	_, err := LoadFromBytes([]byte("not json"))
	if err == nil {
		t.Error("should return error")
	}
}

func TestLoadFromBytes_Valid_P209(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)
	r.RecordKey('b', 0, 0)
	r.Stop()
	data, _ := r.SaveBytes()
	p, err := LoadFromBytes(data)
	if err != nil {
		t.Errorf("error: %v", err)
	}
	if p == nil {
		t.Error("should return player")
	}
}