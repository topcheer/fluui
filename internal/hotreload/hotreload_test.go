package hotreload

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// === Constructor ===

func TestNewWatcher_DefaultInterval(t *testing.T) {
	w := NewWatcher(0)
	if w.interval != 500*time.Millisecond {
		t.Errorf("expected default 500ms, got %v", w.interval)
	}
	w.Stop()
}

func TestNewWatcher_CustomInterval(t *testing.T) {
	w := NewWatcher(100 * time.Millisecond)
	if w.interval != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", w.interval)
	}
	w.Stop()
}

func TestNewWatcher_DefaultDebounce(t *testing.T) {
	w := NewWatcher(0)
	if w.debounce != 100*time.Millisecond {
		t.Errorf("expected default debounce 100ms, got %v", w.debounce)
	}
	w.Stop()
}

// === AddPath ===

func TestAddPath_SingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	if err := w.AddPath(f); err != nil {
		t.Fatalf("AddPath failed: %v", err)
	}
	if w.PathCount() == 0 {
		t.Error("expected at least 1 watched path")
	}
}

func TestAddPath_Directory(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	if err := w.AddPath(dir); err != nil {
		t.Fatalf("AddPath failed: %v", err)
	}
	if w.PathCount() < 2 {
		t.Errorf("expected at least 2 watched paths, got %d", w.PathCount())
	}
}

func TestAddPath_NestedDirectory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	os.Mkdir(sub, 0755)
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(sub, "b.txt"), []byte("b"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	if err := w.AddPath(dir); err != nil {
		t.Fatalf("AddPath failed: %v", err)
	}
	if w.PathCount() < 3 {
		t.Errorf("expected at least 3 paths (dir + sub + files), got %d", w.PathCount())
	}
}

func TestAddPath_NonExistent(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	err := w.AddPath("/nonexistent/path/to/file")
	if err == nil {
		t.Error("expected error for non-existent path")
	}
}

// === RemovePath ===

func TestRemovePath_File(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("hello"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	w.AddPath(f)
	before := w.PathCount()
	w.RemovePath(f)
	after := w.PathCount()
	if after >= before {
		t.Errorf("expected fewer paths after remove, got before=%d after=%d", before, after)
	}
}

func TestRemovePath_DirectoryRecursive(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("b"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	w.AddPath(dir)
	before := w.PathCount()
	w.RemovePath(dir)
	after := w.PathCount()
	if after >= before {
		t.Errorf("expected fewer paths after remove dir, got before=%d after=%d", before, after)
	}
}

func TestRemovePath_NotWatched(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	// Should not panic
	w.RemovePath("/not/watched")
}

// === OnChange ===

func TestOnChange_SingleFile(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("initial"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	var callCount int32
	w.OnChange(func(changed []string) {
		atomic.AddInt32(&callCount, 1)
	})
	w.AddPath(f)
	w.Start()

	// Modify the file
	time.Sleep(80 * time.Millisecond)
	os.WriteFile(f, []byte("modified"), 0644)

	// Wait for callback
	time.Sleep(150 * time.Millisecond)

	if atomic.LoadInt32(&callCount) == 0 {
		t.Error("expected OnChange callback to be called")
	}
}

func TestOnChange_Debounce(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "test.txt")
	os.WriteFile(f, []byte("initial"), 0644)

	w := NewWatcher(20 * time.Millisecond)
	w.SetDebounce(200 * time.Millisecond)
	defer w.Stop()

	var callCount int32
	w.OnChange(func(changed []string) {
		atomic.AddInt32(&callCount, 1)
	})
	w.AddPath(f)
	w.Start()

	// Rapidly modify the file multiple times
	time.Sleep(30 * time.Millisecond)
	for i := 0; i < 5; i++ {
		os.WriteFile(f, []byte("mod"+string(rune('0'+i))), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait enough for debounce to expire + one poll
	time.Sleep(300 * time.Millisecond)

	// With 200ms debounce, should have been called fewer than 5 times
	if atomic.LoadInt32(&callCount) > 3 {
		t.Errorf("expected debounce to limit calls, got %d", atomic.LoadInt32(&callCount))
	}
}

func TestOnChange_FileDeleted(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "deletable.txt")
	os.WriteFile(f, []byte("content"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	var changed []string
	var mu sync.Mutex
	w.OnChange(func(c []string) {
		mu.Lock()
		changed = append(changed, c...)
		mu.Unlock()
	})
	w.AddPath(f)
	w.Start()

	time.Sleep(80 * time.Millisecond)
	os.Remove(f)
	time.Sleep(150 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, p := range changed {
		if p == f {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected file deletion to trigger callback")
	}
}

func TestOnChange_NewFileInDir(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("data"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	var changed []string
	var mu sync.Mutex
	w.OnChange(func(c []string) {
		mu.Lock()
		changed = append(changed, c...)
		mu.Unlock()
	})
	w.AddPath(dir)
	w.Start()

	time.Sleep(80 * time.Millisecond)
	// Create new file in watched directory
	newFile := filepath.Join(dir, "new.txt")
	os.WriteFile(newFile, []byte("new"), 0644)
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	found := false
	for _, p := range changed {
		if p == newFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected new file creation in directory to trigger callback")
	}
}

// === Lifecycle ===

func TestStop_Idempotent(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	w.Stop()
	w.Stop() // should not panic
}

func TestStart_AfterStop(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	w.Start()
	w.Stop()
	// Cannot restart after stop
	w.Start() // should be a no-op
}

func TestStart_MultipleCalls(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()
	w.Start()
	w.Start() // second Start is a no-op (only one goroutine)
}

// === WatchedPaths ===

func TestWatchedPaths(t *testing.T) {
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	os.WriteFile(f1, []byte("a"), 0644)
	os.WriteFile(f2, []byte("b"), 0644)

	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	w.AddPath(f1)
	w.AddPath(f2)

	paths := w.WatchedPaths()
	if len(paths) < 2 {
		t.Errorf("expected at least 2 paths, got %d", len(paths))
	}
}

// === Concurrent access ===

func TestConcurrent_AddRemovePath(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	w.Start()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			f := filepath.Join(dir, "file"+string(rune('0'+n%10))+".txt")
			os.WriteFile(f, []byte("x"), 0644)
			w.AddPath(f)
			w.RemovePath(f)
		}(i)
	}
	wg.Wait()
}

func TestConcurrent_AddPathAndCount(t *testing.T) {
	dir := t.TempDir()
	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(2)

		go func(n int) {
			defer wg.Done()
			f := filepath.Join(dir, "f"+string(rune('0'+n))+".txt")
			os.WriteFile(f, []byte("x"), 0644)
			w.AddPath(f)
		}(i)

		go func() {
			defer wg.Done()
			_ = w.PathCount()
			_ = w.WatchedPaths()
		}()
	}
	wg.Wait()
}

// === SetDebounce ===

func TestSetDebounce(t *testing.T) {
	w := NewWatcher(50 * time.Millisecond)
	defer w.Stop()

	w.SetDebounce(500 * time.Millisecond)
	if w.debounce != 500*time.Millisecond {
		t.Errorf("expected 500ms, got %v", w.debounce)
	}
}

// === Benchmarks ===

func BenchmarkAddPath(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 100; i++ {
		os.WriteFile(filepath.Join(dir, "file"+string(rune('a'+i%26))+".txt"), []byte("x"), 0644)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWatcher(1 * time.Second)
		w.AddPath(dir)
		w.Stop()
	}
}

func BenchmarkCheck_100Files(b *testing.B) {
	dir := b.TempDir()
	for i := 0; i < 100; i++ {
		os.WriteFile(filepath.Join(dir, "f"+string(rune('a'+i%26))+string(rune('a'+(i/26)%26))+".txt"), []byte("x"), 0644)
	}
	w := NewWatcher(1 * time.Second)
	w.AddPath(dir)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.check()
	}
	w.Stop()
}
