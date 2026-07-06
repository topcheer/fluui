package fluui

import (
	"time"

	"github.com/topcheer/fluui/internal/hotreload"
)

// HotReload provides file-watching based hot reload for TUI applications.
// It monitors files/directories for changes and triggers re-renders.

// WatchFile adds a file path to the hot reload watch list.
// When the file changes, the app is marked dirty (triggering a re-render).
// Returns an error if the file does not exist.
func (a *App) WatchFile(path string) error {
	a.initWatcher()
	return a.watcher.AddPath(path)
}

// WatchDir adds a directory to the hot reload watch list.
// All files in the directory are watched recursively.
// When any file changes, the app is marked dirty.
func (a *App) WatchDir(path string) error {
	a.initWatcher()
	return a.watcher.AddPath(path)
}

// OnHotReload sets a callback that is invoked when watched files change.
// The callback receives the list of changed file paths.
// If not set, the app simply marks itself dirty (re-renders).
func (a *App) OnHotReload(fn func(changed []string)) {
	a.initWatcher()
	a.watcher.OnChange(func(changed []string) {
		fn(changed)
		a.MarkDirty()
	})
}

// SetWatchInterval sets the file polling interval.
// Default is 500ms. Lower values provide faster feedback but use more CPU.
// SetWatchInterval sets the debounce for file change callbacks.
// Lower values provide faster feedback but may fire more frequently.
// Default debounce is 100ms.
func (a *App) SetWatchInterval(d time.Duration) {
	a.initWatcher()
	a.watcher.SetDebounce(d)
}

// WatchedPaths returns the list of currently watched file paths.
func (a *App) WatchedPaths() []string {
	if a.watcher == nil {
		return nil
	}
	return a.watcher.WatchedPaths()
}

// StopWatching stops the hot reload watcher.
func (a *App) StopWatching() {
	if a.watcher != nil {
		a.watcher.Stop()
	}
}

// initWatcher lazily creates and starts the watcher on first use.
func (a *App) initWatcher() {
	if a.watcher != nil {
		return
	}
	a.watcher = hotreload.NewWatcher(500 * time.Millisecond)
	a.watcher.OnChange(func(changed []string) {
		a.MarkDirty()
	})
	a.watcher.Start()
}
