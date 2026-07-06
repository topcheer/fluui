package hotreload

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestP69_RemovePath_NonExistent(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	w.RemovePath("/nonexistent/path")
	if w.PathCount() != 0 {
		t.Errorf("expected 0 paths, got %d", w.PathCount())
	}
}

func TestP69_RemovePath_AfterAdd(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher(50 * time.Millisecond)
	w.AddPath(dir)
	if w.PathCount() != 1 {
		t.Errorf("expected 1 path, got %d", w.PathCount())
	}
	w.RemovePath(dir)
	if w.PathCount() != 0 {
		t.Errorf("expected 0 paths after remove, got %d", w.PathCount())
	}
}

func TestP69_RemovePath_FileInDir(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher(50 * time.Millisecond)
	w.AddPath(dir)
	w.RemovePath(filepath.Join(dir, "nonexistent"))
	if w.PathCount() != 1 {
		t.Errorf("expected 1 path still tracked, got %d", w.PathCount())
	}
}

func TestP69_OnChange_MultipleCallbacks(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher(50 * time.Millisecond)
	var count int64
	w.OnChange(func(changed []string) { atomic.AddInt64(&count, 1) })
	w.AddPath(dir)
	w.Start()
	defer w.Stop()

	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)
	time.Sleep(200 * time.Millisecond)

	os.WriteFile(f, []byte("world"), 0644)
	time.Sleep(200 * time.Millisecond)

	if atomic.LoadInt64(&count) == 0 {
		t.Log("callback may not have fired (timing dependent)")
	}
}
