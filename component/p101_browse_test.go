package component

import (
	"path/filepath"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

func TestP101_ThemeStudio_BrowseOpen(t *testing.T) {
	ts := NewThemeStudio(theme.Get())

	// Browse should be closed initially.
	ts.mu.RLock()
	open := ts.browseOpen
	ts.mu.RUnlock()
	if open {
		t.Fatal("browse should be closed initially")
	}

	// Press 'o' to open browse.
	handled := ts.HandleKey(&term.KeyEvent{Rune: 'o'})
	if !handled {
		t.Fatal("HandleKey should return true for 'o'")
	}

	ts.mu.RLock()
	open = ts.browseOpen
	ts.mu.RUnlock()
	if !open {
		t.Fatal("browse should be open after pressing 'o'")
	}
}

func TestP101_ThemeStudio_BrowseClose(t *testing.T) {
	ts := NewThemeStudio(theme.Get())

	// Open browse.
	ts.HandleKey(&term.KeyEvent{Rune: 'o'})

	// Press Escape to close.
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEscape})

	ts.mu.RLock()
	open := ts.browseOpen
	ts.mu.RUnlock()
	if open {
		t.Fatal("browse should be closed after Escape")
	}
}

func TestP101_ThemeStudio_BrowseNavigation(t *testing.T) {
	ts := NewThemeStudio(theme.Get())

	// Set up fake browse files.
	tmpDir := t.TempDir()
	t1 := &theme.Theme{Name: "Dark", Bg: theme.Hex("#111111"), Fg: theme.Hex("#eeeeee")}
	theme.SaveToFile(t1, filepath.Join(tmpDir, "dark.json"))
	t2 := &theme.Theme{Name: "Light", Bg: theme.Hex("#ffffff"), Fg: theme.Hex("#000000")}
	theme.SaveToFile(t2, filepath.Join(tmpDir, "light.json"))

	ts.mu.Lock()
	ts.browseDir = tmpDir
	ts.browseFiles, _ = theme.ListThemeFiles(tmpDir)
	ts.browseOpen = true
	ts.browseCursor = 0
	ts.mu.Unlock()

	if len(ts.browseFiles) < 2 {
		t.Fatalf("expected at least 2 browse files, got %d", len(ts.browseFiles))
	}

	// Press Down to move cursor.
	ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})

	ts.mu.RLock()
	cursor := ts.browseCursor
	ts.mu.RUnlock()
	if cursor != 1 {
		t.Errorf("expected cursor=1 after Down, got %d", cursor)
	}

	// Press Up to move back.
	ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})

	ts.mu.RLock()
	cursor = ts.browseCursor
	ts.mu.RUnlock()
	if cursor != 0 {
		t.Errorf("expected cursor=0 after Up, got %d", cursor)
	}
}

func TestP101_ThemeStudio_BrowseLoadTheme(t *testing.T) {
	// Save current active theme to restore later.
	oldActive := theme.Active
	defer func() { theme.Active = oldActive }()

	ts := NewThemeStudio(theme.Get())

	tmpDir := t.TempDir()
	custom := &theme.Theme{
		Name:   "CustomLoaded",
		Bg:     theme.Hex("#0a0b0c"),
		Fg:     theme.Hex("#aabbcc"),
		Accent: theme.Hex("#ff00ff"),
	}
	theme.SaveToFile(custom, filepath.Join(tmpDir, "custom.json"))

	ts.mu.Lock()
	ts.browseDir = tmpDir
	ts.browseFiles, _ = theme.ListThemeFiles(tmpDir)
	ts.browseOpen = true
	ts.browseCursor = 0
	ts.mu.Unlock()

	// Press Enter to load.
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	// Browse should be closed after loading.
	ts.mu.RLock()
	open := ts.browseOpen
	ts.mu.RUnlock()
	if open {
		t.Error("browse should be closed after loading")
	}

	// Active theme should be the loaded one.
	if theme.Get().Name != "CustomLoaded" {
		t.Errorf("expected active theme 'CustomLoaded', got %q", theme.Get().Name)
	}
}

func TestP101_ThemeStudio_BrowseEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	ts := NewThemeStudio(theme.Get())
	ts.mu.Lock()
	ts.browseDir = tmpDir
	ts.browseFiles, _ = theme.ListThemeFiles(tmpDir)
	ts.browseOpen = true
	ts.browseCursor = 0
	ts.mu.Unlock()

	if len(ts.browseFiles) != 0 {
		t.Errorf("expected 0 files in empty dir, got %d", len(ts.browseFiles))
	}

	// Paint should not crash with empty file list.
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 24})
	buf := buffer.NewBuffer(60, 24)
	ts.Paint(buf)

	// Verify "No saved themes" text is rendered.
	found := false
	for y := 0; y < 24; y++ {
		for x := 0; x < 60; x++ {
			if buf.GetCell(x, y).Rune == 'N' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'No saved themes' text in output")
	}
}

func TestP101_ThemeStudio_BrowsePaint(t *testing.T) {
	ts := NewThemeStudio(theme.Get())

	// Set up with files.
	tmpDir := t.TempDir()
	t1 := &theme.Theme{Name: "PaintTest", Bg: theme.Hex("#222222")}
	theme.SaveToFile(t1, filepath.Join(tmpDir, "paint.json"))

	ts.mu.Lock()
	ts.browseDir = tmpDir
	ts.browseFiles, _ = theme.ListThemeFiles(tmpDir)
	ts.browseOpen = true
	ts.browseCursor = 0
	ts.mu.Unlock()

	// Paint should not crash.
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 24})
	buf := buffer.NewBuffer(60, 24)
	ts.Paint(buf)

	// Should contain "Open Theme" title.
	found := false
	for y := 0; y < 24; y++ {
		for x := 0; x < 60-4; x++ {
			if buf.GetCell(x, y).Rune == 'O' &&
				buf.GetCell(x+1, y).Rune == 'p' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'Open Theme' title in browse overlay")
	}
}

func TestP101_ThemeStudio_OnLoadCallback(t *testing.T) {
	oldActive := theme.Active
	defer func() { theme.Active = oldActive }()

	loaded := false
	ts := NewThemeStudio(theme.Get())
	ts.OnLoad = func() { loaded = true }

	tmpDir := t.TempDir()
	t1 := &theme.Theme{Name: "Callback", Bg: theme.Hex("#333333")}
	theme.SaveToFile(t1, filepath.Join(tmpDir, "cb.json"))

	ts.mu.Lock()
	ts.browseDir = tmpDir
	ts.browseFiles, _ = theme.ListThemeFiles(tmpDir)
	ts.browseOpen = true
	ts.browseCursor = 0
	ts.mu.Unlock()

	// Load via Enter.
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if !loaded {
		t.Error("OnLoad callback should have been called")
	}
}
