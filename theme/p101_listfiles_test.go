package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestP101_ListThemeFiles_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	themes, err := ListThemeFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListThemeFiles error: %v", err)
	}
	if len(themes) != 0 {
		t.Errorf("expected 0 themes in empty dir, got %d", len(themes))
	}
}

func TestP101_ListThemeFiles_NonExistentDir(t *testing.T) {
	// Non-existent dir should return nil without error.
	themes, err := ListThemeFiles("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Fatalf("expected nil error for non-existent dir, got: %v", err)
	}
	if themes != nil {
		t.Errorf("expected nil themes for non-existent dir, got %d", len(themes))
	}
}

func TestP101_ListThemeFiles_WithThemes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid theme file.
	t1 := &Theme{
		Name:      "TestDark",
		Bg:        Hex("#1a1a2e"),
		Fg:        Hex("#e0e0e0"),
		Accent:    Hex("#7f5af0"),
		Border:    Hex("#3a3a5c"),
		BorderActive: Hex("#7f5af0"),
		BorderMuted:  Hex("#2a2a3c"),
		Success:   Hex("#2cb67d"),
		Error:     Hex("#e63946"),
		Warning:   Hex("#f4a261"),
		Muted:     Hex("#6c6c80"),
		CodeBg:    Hex("#16161e"),
		CodeFg:    Hex("#e0e0e0"),
	}
	if err := SaveToFile(t1, filepath.Join(tmpDir, "dark.json")); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	// Create another valid theme file.
	t2 := &Theme{
		Name:   "TestLight",
		Bg:     Hex("#fafafa"),
		Fg:     Hex("#1a1a1a"),
		Accent: Hex("#0066ff"),
	}
	if err := SaveToFile(t2, filepath.Join(tmpDir, "light.json")); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	// Create a non-JSON file (should be skipped).
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("hello"), 0644)

	// Create a subdirectory (should be skipped).
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// Create an invalid JSON file (should use filename as name).
	os.WriteFile(filepath.Join(tmpDir, "broken.json"), []byte("{invalid"), 0644)

	themes, err := ListThemeFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListThemeFiles error: %v", err)
	}
	if len(themes) != 3 {
		t.Fatalf("expected 3 themes (2 valid + 1 broken), got %d", len(themes))
	}

	// Find each theme by name.
	names := make(map[string]string) // name → path
	for _, tf := range themes {
		names[tf.Name] = tf.Path
	}

	if _, ok := names["TestDark"]; !ok {
		t.Error("expected TestDark in theme list")
	}
	if _, ok := names["TestLight"]; !ok {
		t.Error("expected TestLight in theme list")
	}
	// broken.json should fall back to filename without extension.
	if _, ok := names["broken"]; !ok {
		t.Error("expected 'broken' (filename fallback) in theme list")
	}
}

func TestP101_DefaultThemeDir(t *testing.T) {
	dir := DefaultThemeDir()
	if dir == "" {
		t.Error("expected non-empty default theme dir")
	}
	// Should end with themes.
	if filepath.Base(dir) != "themes" {
		t.Errorf("expected dir to end with 'themes', got %s", dir)
	}
}

func TestP101_ListThemeFiles_DefaultDir(t *testing.T) {
	// When dir is empty, should use DefaultThemeDir().
	// This might not exist yet — should return nil without error.
	themes, err := ListThemeFiles("")
	if err != nil {
		t.Errorf("expected nil error for default dir, got: %v", err)
	}
	// themes might be nil or empty — either is fine.
	_ = themes
}

func TestP101_SaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "roundtrip.json")

	original := &Theme{
		Name:      "RoundTrip",
		Bg:        Hex("#0d1117"),
		Fg:        Hex("#c9d1d9"),
		Accent:    Hex("#58a6ff"),
		Border:    Hex("#30363d"),
		BorderActive: Hex("#58a6ff"),
		BorderMuted:  Hex("#21262d"),
		Success:   Hex("#3fb950"),
		Error:     Hex("#f85149"),
		Warning:   Hex("#d29922"),
		Muted:     Hex("#484f58"),
		CodeBg:    Hex("#161b22"),
		CodeFg:    Hex("#c9d1d9"),
	}

	if err := SaveToFile(original, path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("name: got %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Bg != original.Bg {
		t.Errorf("bg: got %v, want %v", loaded.Bg, original.Bg)
	}
	if loaded.Accent != original.Accent {
		t.Errorf("accent: got %v, want %v", loaded.Accent, original.Accent)
	}
}

func TestP101_LoadAndActivate(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "activate.json")

	// Save a known theme.
	t1 := &Theme{
		Name:   "Activate",
		Bg:     Hex("#abcdef"),
		Fg:     Hex("#123456"),
		Accent: Hex("#789abc"),
	}
	SaveToFile(t1, path)

	// Save current active, restore after test.
	oldActive := Active
	defer func() { Active = oldActive }()

	if err := LoadAndActivate(path); err != nil {
		t.Fatalf("LoadAndActivate: %v", err)
	}

	if Active.Name != "Activate" {
		t.Errorf("expected active name 'Activate', got %q", Active.Name)
	}
	if Active.Bg != t1.Bg {
		t.Errorf("expected active bg %v, got %v", t1.Bg, Active.Bg)
	}
}
