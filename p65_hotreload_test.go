package fluui

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// === Hot Reload App integration tests ===

func TestP65_WatchFile(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	dir := t.TempDir()
	f := filepath.Join(dir, "config.yaml")
	os.WriteFile(f, []byte("key: value"), 0644)

	err := app.WatchFile(f)
	if err != nil {
		t.Fatalf("WatchFile failed: %v", err)
	}
	paths := app.WatchedPaths()
	if len(paths) == 0 {
		t.Error("expected at least 1 watched path")
	}
}

func TestP65_WatchFile_NonExistent(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	err := app.WatchFile("/nonexistent/path/file.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestP65_WatchDir(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)

	err := app.WatchDir(dir)
	if err != nil {
		t.Fatalf("WatchDir failed: %v", err)
	}
	paths := app.WatchedPaths()
	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}
}

func TestP65_OnHotReload(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	dir := t.TempDir()
	f := filepath.Join(dir, "theme.yaml")
	os.WriteFile(f, []byte("fg: white"), 0644)

	var callCount int32
	app.OnHotReload(func(changed []string) {
		atomic.AddInt32(&callCount, 1)
	})

	app.WatchFile(f)

	// Modify the file after initial poll
	time.Sleep(200 * time.Millisecond)
	os.WriteFile(f, []byte("fg: red"), 0644)

	// Wait for at least 2 poll cycles (500ms each)
	time.Sleep(700 * time.Millisecond)
	if atomic.LoadInt32(&callCount) == 0 {
		t.Error("expected OnHotReload callback to fire")
	}
}

func TestP65_SetWatchInterval(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	app.SetWatchInterval(100 * time.Millisecond)
	// Should not panic
}

func TestP65_WatchedPaths_NoWatcher(t *testing.T) {
	app := newTestApp(80, 24)
	// Without initializing watcher, WatchedPaths should return nil
	paths := app.WatchedPaths()
	if paths != nil {
		t.Error("expected nil when no watcher initialized")
	}
}

func TestP65_StopWatching_NoWatcher(t *testing.T) {
	app := newTestApp(80, 24)
	// Should not panic
	app.StopWatching()
}

func TestP65_StopWatching_Idempotent(t *testing.T) {
	app := newTestApp(80, 24)
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("x"), 0644)
	app.WatchDir(dir)

	app.StopWatching()
	app.StopWatching() // should not panic
}

func TestP65_HotReload_DefaultMarksDirty(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	dir := t.TempDir()
	f := filepath.Join(dir, "data.txt")
	os.WriteFile(f, []byte("initial"), 0644)

	// Without OnHotReload, default callback just calls MarkDirty
	// This test verifies no panic and watcher works
	app.WatchFile(f)

	// Modify file
	time.Sleep(100 * time.Millisecond)
	os.WriteFile(f, []byte("changed"), 0644)
	time.Sleep(200 * time.Millisecond)

	// No assertion on dirty state — just verifies no panic
}

func TestP65_WatchFile_MultipleFiles(t *testing.T) {
	app := newTestApp(80, 24)
	defer app.StopWatching()

	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	os.WriteFile(f1, []byte("a"), 0644)
	os.WriteFile(f2, []byte("b"), 0644)

	app.WatchFile(f1)
	app.WatchFile(f2)

	paths := app.WatchedPaths()
	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}
}
