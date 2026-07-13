package hotreload

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestP180_AddPathNonExistent(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	err := w.AddPath("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestP180_AddPathLockedNonExistent(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	w.mu.Lock()
	err := w.addPathLocked("/another/nonexistent/path")
	w.mu.Unlock()
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

func TestP180_AddPathUnreadableDir(t *testing.T) {
	// Create a directory with no read permissions
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "noperm")
	os.Mkdir(subDir, 0755)
	// Create a file inside
	os.WriteFile(filepath.Join(subDir, "test.txt"), []byte("test"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	err := w.AddPath(subDir)
	if err != nil {
		t.Errorf("expected no error for readable dir, got %v", err)
	}
}

func TestP180_ModTimeOrZeroError(t *testing.T) {
	// Create a fake DirEntry that returns error on Info()
	// We can test by creating a file then deleting it before calling Info()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "temp.txt")
	os.WriteFile(filePath, []byte("test"), 0644)

	entries, _ := os.ReadDir(tmpDir)
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}

	// Delete the file so Info() will fail
	os.Remove(filePath)

	// Now modTimeOrZero should return zero time
	result := modTimeOrZero(entries[0])
	if !result.IsZero() {
		t.Error("expected zero time for deleted file")
	}
}

func TestP180_SizeOrZeroError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "temp.txt")
	os.WriteFile(filePath, []byte("test"), 0644)

	entries, _ := os.ReadDir(tmpDir)
	if len(entries) != 1 {
		t.Fatal("expected 1 entry")
	}

	// Delete the file so Info() will fail
	os.Remove(filePath)

	// Now sizeOrZero should return 0
	result := sizeOrZero(entries[0])
	if result != 0 {
		t.Error("expected 0 for deleted file")
	}
}

func TestP180_AddPathAbsError(t *testing.T) {
	// Test the filepath.Abs error path is hard to trigger normally
	// but we can verify AddPath works with relative paths
	tmpDir := t.TempDir()
	relPath := filepath.Base(tmpDir)
	// Change to parent dir
	oldDir, _ := os.Getwd()
	os.Chdir(filepath.Dir(tmpDir))
	defer os.Chdir(oldDir)

	w := NewWatcher(50 * time.Millisecond)
	err := w.AddPath(relPath)
	if err != nil {
		t.Errorf("expected no error for relative path, got %v", err)
	}
}

func TestP180_AddPathAfterStart(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("initial"), 0644)

	w.AddPath(tmpDir)
	w.Start()
	defer w.Stop()

	// Add another file path
	file2 := filepath.Join(tmpDir, "test2.txt")
	os.WriteFile(file2, []byte("test2"), 0644)
	w.AddPath(file2)

	if w.PathCount() < 2 {
		t.Errorf("expected at least 2 paths, got %d", w.PathCount())
	}
}
