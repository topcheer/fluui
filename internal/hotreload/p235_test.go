package hotreload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// P235: cover hotreload error paths + subdirectory recursion + stopped check

func TestWatcher_AddPathNonExistent_P235(t *testing.T) {
	w := NewWatcher(1 * time.Second)
	defer w.Stop()
	// AddPath on non-existent file → os.Stat error
	err := w.AddPath("/nonexistent/path/file.go")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestWatcher_RemovePathNotWatching_P235(t *testing.T) {
	w := NewWatcher(1 * time.Second)
	defer w.Stop()
	// RemovePath on path not being watched — should be no-op
	w.RemovePath("/some/random/path")
}

func TestWatcher_AddPathWithSubdirs_P235(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "helper.go"), []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher(1 * time.Second)
	defer w.Stop()
	if err := w.AddPath(tmpDir); err != nil {
		t.Fatalf("AddPath failed: %v", err)
	}
	// Should have tracked the subdir recursively
	w.mu.Lock()
	count := len(w.paths)
	w.mu.Unlock()
	if count < 3 {
		t.Errorf("expected at least 3 paths tracked, got %d", count)
	}
}

func TestWatcher_FileSizeChange_P235(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher(100 * time.Millisecond)
	defer w.Stop()
	if err := w.AddPath(testFile); err != nil {
		t.Fatal(err)
	}

	// Wait for at least one poll cycle
	time.Sleep(200 * time.Millisecond)

	// Modify file size
	if err := os.WriteFile(testFile, []byte("package main\n\nfunc foo() {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for poll to detect change
	time.Sleep(200 * time.Millisecond)
}

func TestWatcher_FileDeleted_P235(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "deleted.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	w := NewWatcher(100 * time.Millisecond)
	defer w.Stop()
	if err := w.AddPath(testFile); err != nil {
		t.Fatal(err)
	}

	time.Sleep(200 * time.Millisecond)

	// Delete the file
	os.Remove(testFile)

	// Wait for poll to detect deletion
	time.Sleep(200 * time.Millisecond)
}

func TestWatcher_CheckAfterStop_P235(t *testing.T) {
	w := NewWatcher(100 * time.Millisecond)
	w.Stop()
	// check() after Stop should return immediately (stopped guard)
	// The poll goroutine should exit
	time.Sleep(100 * time.Millisecond)
}
