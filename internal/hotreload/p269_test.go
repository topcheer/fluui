package hotreload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAddPath_NonExistent_P269(t *testing.T) {
	w := NewWatcher(time.Second)
	defer w.Stop()
	err := w.AddPath("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("should return error for non-existent path")
	}
}

func TestAddPath_FileNotDir_P269(t *testing.T) {
	w := NewWatcher(time.Second)
	defer w.Stop()
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	err := w.AddPath(tmpFile)
	if err != nil {
		t.Errorf("should succeed for a file, got: %v", err)
	}
}

func TestAddPath_DirWithSubdirs_P269(t *testing.T) {
	w := NewWatcher(time.Second)
	defer w.Stop()
	dir := t.TempDir()
	subDir := filepath.Join(dir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "root.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	err := w.AddPath(dir)
	if err != nil {
		t.Errorf("should succeed for directory with subdirs, got: %v", err)
	}
}

func TestRemovePath_NonExistent_P269(t *testing.T) {
	w := NewWatcher(time.Second)
	defer w.Stop()
	w.RemovePath("/nonexistent/path")
}

func TestRemovePath_Existing_P269(t *testing.T) {
	w := NewWatcher(time.Second)
	defer w.Stop()
	dir := t.TempDir()
	tmpFile := filepath.Join(dir, "remove_me.txt")
	os.WriteFile(tmpFile, []byte("x"), 0644)
	w.AddPath(tmpFile)
	w.RemovePath(tmpFile)
}

func TestStop_AlreadyStopped_P269(t *testing.T) {
	w := NewWatcher(time.Second)
	w.Stop()
	w.Stop()
}
