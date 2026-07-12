package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestP160_SaveToFile_Basic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	th := Dracula()
	err := SaveToFile(th, path)
	if err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Error("expected file to exist")
	}
}

func TestP160_SaveToFile_WithDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "theme.json")
	th := Dracula()
	err := SaveToFile(th, path)
	if err != nil {
		t.Fatalf("SaveToFile with nested dirs: %v", err)
	}
}

func TestP160_SaveToFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.json")
	th := Dracula()
	if err := SaveToFile(th, path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Bg.Val != th.Bg.Val {
		t.Errorf("expected bg %d, got %d", th.Bg.Val, loaded.Bg.Val)
	}
}

func TestP160_SaveToFile_WriteError(t *testing.T) {
	// Try writing to a directory (should fail)
	dir := t.TempDir()
	err := SaveToFile(Dracula(), dir) // dir is not a file
	if err == nil {
		t.Error("expected write error")
	}
}

func TestP160_SaveActive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "active.json")
	original := Active
	Active = Dracula()
	defer func() { Active = original }()
	err := SaveActive(path)
	if err != nil {
		t.Fatalf("SaveActive: %v", err)
	}
}

func TestP160_LoadAndActivate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activate.json")
	original := Active
	defer func() { Active = original }()
	Active = Dracula()
	if err := SaveActive(path); err != nil {
		t.Fatalf("SaveActive: %v", err)
	}
	if err := LoadAndActivate(path); err != nil {
		t.Fatalf("LoadAndActivate: %v", err)
	}
	if Active == nil {
		t.Error("expected non-nil Active")
	}
}

func TestP160_DefaultThemeDir(t *testing.T) {
	dir := DefaultThemeDir()
	if dir == "" {
		t.Error("expected non-empty dir")
	}
}

func TestP160_ListThemeFiles_EmptyDir(t *testing.T) {
	files, err := ListThemeFiles("")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_ = files
}

func TestP160_ListThemeFiles_NonExistent(t *testing.T) {
	files, err := ListThemeFiles("/nonexistent/path/xyz")
	if err != nil {
		t.Errorf("expected nil error for non-existent dir, got %v", err)
	}
	if files != nil {
		t.Errorf("expected nil files for non-existent dir, got %v", files)
	}
}

func TestP160_ListThemeFiles_Valid(t *testing.T) {
	dir := t.TempDir()
	// Create test theme files
	th := Dracula()
	if err := SaveToFile(th, filepath.Join(dir, "dark.json")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	th2 := Default()
	if err := SaveToFile(th2, filepath.Join(dir, "light.json")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	// Create non-json file
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644)

	files, err := ListThemeFiles(dir)
	if err != nil {
		t.Fatalf("ListThemeFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 theme files, got %d", len(files))
	}
}

func TestP160_ListThemeFiles_EmptyDir2(t *testing.T) {
	dir := t.TempDir()
	files, err := ListThemeFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files in empty dir, got %d", len(files))
	}
}

func TestP160_LoadFromFile_NotExist(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/theme.json")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestP160_LoadFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	os.WriteFile(path, []byte("invalid json"), 0644)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP160_SaveToFile_256Color(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "256.json")
	th := Dracula()
	th.Fg.Type = 2
	th.Fg.Val = 196
	if err := SaveToFile(th, path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Fg.Type != 2 || loaded.Fg.Val != 196 {
		t.Errorf("expected 256 color 196, got type=%d val=%d", loaded.Fg.Type, loaded.Fg.Val)
	}
}