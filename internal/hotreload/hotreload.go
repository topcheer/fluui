// Package hotreload provides file-watching based hot reload capabilities.
// It monitors files and directories for changes and triggers callbacks
// when modifications are detected.
//
// Hot Reload is a key developer experience feature for TUI applications.
// It enables:
//   - Theme hot-reloading during development
//   - Component definition watching and refreshing
//   - Configuration file auto-reload
//   - Development workflows with instant feedback
//
// Implementation uses polling (stdlib only, zero external dependencies).
// Polling is cross-platform, works on all Go-supported platforms, and
// has negligible CPU overhead for typical file counts (<1000 files).
//
// Competitive advantage: First Go TUI library with built-in hot reload.
// Textual (Python) has CSS hot-reload via external tools.
// Bubble Tea, tview, Ratatui lack this entirely.
package hotreload

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Watcher monitors files and directories for changes.
// When a change is detected, the registered callback is invoked.
type Watcher struct {
	mu       sync.Mutex
	paths    map[string]fileState // path → last known state
	callback func(changed []string)
	interval time.Duration
	stopCh   chan struct{}
	stopped  bool
	wg       sync.WaitGroup
	debounce time.Duration
	lastFire time.Time
}

// fileState tracks the modification time and size of a watched file.
type fileState struct {
	modTime time.Time
	size    int64
	isDir   bool
}

// NewWatcher creates a new file watcher with the given poll interval.
// Default interval is 500ms which provides responsive feedback with
// negligible CPU usage.
func NewWatcher(interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &Watcher{
		paths:    make(map[string]fileState),
		interval: interval,
		stopCh:   make(chan struct{}),
		debounce: 100 * time.Millisecond,
	}
}

// AddPath adds a file or directory to the watch list.
// If the path is a directory, all files within it are watched recursively.
// Returns an error if the path does not exist.
func (w *Watcher) AddPath(path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	return w.addPathLocked(absPath)
}

// addPathLocked adds a path without locking (caller must hold the lock).
func (w *Watcher) addPathLocked(absPath string) error {
	info, err := os.Stat(absPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		// Watch directory and all files within it
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return err
		}
		w.paths[absPath] = fileState{
			modTime: info.ModTime(),
			size:    info.Size(),
			isDir:   true,
		}
		for _, entry := range entries {
			full := filepath.Join(absPath, entry.Name())
			if entry.IsDir() {
				// Recurse into subdirectories
				if err := w.addPathLocked(full); err != nil {
					continue // skip unreadable subdirs
				}
			} else {
				w.paths[full] = fileState{
					modTime: modTimeOrZero(entry),
					size:    sizeOrZero(entry),
				}
			}
		}
	} else {
		w.paths[absPath] = fileState{
			modTime: info.ModTime(),
			size:    info.Size(),
			isDir:   false,
		}
	}
	return nil
}

// RemovePath removes a file or directory from the watch list.
func (w *Watcher) RemovePath(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}

	// Remove the exact path
	delete(w.paths, absPath)

	// Remove any paths under this directory
	prefix := absPath + string(filepath.Separator)
	for p := range w.paths {
		if len(p) > len(prefix) && p[:len(prefix)] == prefix {
			delete(w.paths, p)
		}
	}
}

// OnChange sets the callback to invoke when files change.
// The callback receives a list of changed file paths.
func (w *Watcher) OnChange(fn func(changed []string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callback = fn
}

// SetDebounce sets the minimum time between callback invocations.
// This prevents rapid successive callbacks when many files change
// at once (e.g., during a git checkout). Default is 100ms.
func (w *Watcher) SetDebounce(d time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.debounce = d
}

// WatchedPaths returns the list of currently watched file paths.
func (w *Watcher) WatchedPaths() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	paths := make([]string, 0, len(w.paths))
	for p := range w.paths {
		paths = append(paths, p)
	}
	return paths
}

// PathCount returns the number of watched paths.
func (w *Watcher) PathCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.paths)
}

// Start begins watching for file changes in a background goroutine.
// The watcher runs until Stop() is called.
// Start should only be called once per watcher instance.
func (w *Watcher) Start() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.mu.Unlock()

	w.wg.Add(1)
	go w.watchLoop()
}

// Stop stops the watcher and waits for the background goroutine to exit.
// After Stop, the watcher cannot be restarted.
func (w *Watcher) Stop() {
	w.mu.Lock()
	if w.stopped {
		w.mu.Unlock()
		return
	}
	w.stopped = true
	close(w.stopCh)
	w.mu.Unlock()

	w.wg.Wait()
}

// watchLoop is the main polling loop.
func (w *Watcher) watchLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.check()
		}
	}
}

// check polls all watched files for changes.
func (w *Watcher) check() {
	w.mu.Lock()

	if w.stopped {
		w.mu.Unlock()
		return
	}

	var changed []string
	var newPaths []string

	for path, state := range w.paths {
		info, err := os.Stat(path)
		if err != nil {
			// File was deleted
			if !state.modTime.IsZero() {
				changed = append(changed, path)
				delete(w.paths, path)
			}
			continue
		}

		newModTime := info.ModTime()
		newSize := info.Size()

		if newModTime != state.modTime || newSize != state.size {
			changed = append(changed, path)
			w.paths[path] = fileState{
				modTime: newModTime,
				size:    newSize,
				isDir:   info.IsDir(),
			}

			// If directory changed, check for new files
			if info.IsDir() {
				entries, err := os.ReadDir(path)
				if err == nil {
					for _, entry := range entries {
						full := filepath.Join(path, entry.Name())
						if _, exists := w.paths[full]; !exists {
							newPaths = append(newPaths, full)
						}
					}
				}
			}
		}
	}

	// Add newly discovered files
	for _, p := range newPaths {
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		w.paths[p] = fileState{
			modTime: info.ModTime(),
			size:    info.Size(),
			isDir:   info.IsDir(),
		}
		changed = append(changed, p)
	}

	// Check debounce
	now := time.Now()
	if len(changed) > 0 && w.callback != nil && now.Sub(w.lastFire) >= w.debounce {
		w.lastFire = now
		callback := w.callback
		w.mu.Unlock()
		callback(changed)
	} else {
		w.mu.Unlock()
	}
}

// modTimeOrZero extracts the modification time from a directory entry.
func modTimeOrZero(entry os.DirEntry) time.Time {
	info, err := entry.Info()
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// sizeOrZero extracts the size from a directory entry.
func sizeOrZero(entry os.DirEntry) int64 {
	info, err := entry.Info()
	if err != nil {
		return 0
	}
	return info.Size()
}
