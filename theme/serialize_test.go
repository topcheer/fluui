package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test_theme.json")

	original := Get()
	if err := SaveToFile(original, path); err != nil {
		t.Fatalf("SaveToFile error: %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	if loaded.Name != original.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, original.Name)
	}
	if loaded.Bg != original.Bg {
		t.Errorf("Bg mismatch")
	}
	if loaded.Fg != original.Fg {
		t.Errorf("Fg mismatch")
	}
	if loaded.Accent != original.Accent {
		t.Errorf("Accent mismatch")
	}
}

func TestSaveLoad_AllColors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "all_colors.json")

	original := Get()
	if err := SaveToFile(original, path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}

	// Check every field
	checks := []struct {
		name string
		a, b Color
	}{
		{"Bg", original.Bg, loaded.Bg},
		{"Fg", original.Fg, loaded.Fg},
		{"Accent", original.Accent, loaded.Accent},
		{"Border", original.Border, loaded.Border},
		{"BorderActive", original.BorderActive, loaded.BorderActive},
		{"Success", original.Success, loaded.Success},
		{"Error", original.Error, loaded.Error},
		{"Warning", original.Warning, loaded.Warning},
		{"Muted", original.Muted, loaded.Muted},
		{"CodeBg", original.CodeBg, loaded.CodeBg},
		{"CodeFg", original.CodeFg, loaded.CodeFg},
		{"DiffAdd", original.DiffAdd, loaded.DiffAdd},
		{"DiffDel", original.DiffDel, loaded.DiffDel},
		{"PromptFg", original.PromptFg, loaded.PromptFg},
		{"Separator", original.Separator, loaded.Separator},
		{"SearchBarBg", original.SearchBarBg, loaded.SearchBarBg},
		{"AssistantFg", original.AssistantFg, loaded.AssistantFg},
	}
	for _, c := range checks {
		if c.a != c.b {
			t.Errorf("%s: original=%v, loaded=%v", c.name, c.a, c.b)
		}
	}
}

func TestSaveToFile_CreateDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "theme.json")

	err := SaveToFile(Get(), path)
	if err != nil {
		t.Fatalf("SaveToFile should create dirs: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}

func TestLoadFromFile_NotExist(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/theme.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "invalid.json")
	os.WriteFile(path, []byte("not valid json"), 0644)

	_, err := LoadFromFile(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestSaveActive(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "active.json")

	original := Active
	err := SaveActive(path)
	if err != nil {
		t.Fatalf("SaveActive: %v", err)
	}

	// Verify the file was written
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(data) == 0 {
		t.Error("file should not be empty")
	}
	_ = original
}

func TestLoadAndActivate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activate.json")

	// Save current theme
	if err := SaveActive(path); err != nil {
		t.Fatalf("SaveActive: %v", err)
	}

	// Load and activate
	if err := LoadAndActivate(path); err != nil {
		t.Fatalf("LoadAndActivate: %v", err)
	}

	// Verify it loaded correctly (name should match)
	if Active.Name != Get().Name {
		t.Error("theme name mismatch after LoadAndActivate")
	}
}

func TestColorToHexStr(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"true color", C(255, 0, 0), "#FF0000"},
		{"true color cyan", C(0, 255, 255), "#00FFFF"},
		{"none", NoColor(), ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := colorToHexStr(tc.c)
			if got != tc.want {
				t.Errorf("colorToHexStr(%v) = %q, want %q", tc.c, got, tc.want)
			}
		})
	}
}

func TestHexStrToColor(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Color
	}{
		{"hex red", "#FF0000", C(255, 0, 0)},
		{"hex cyan", "#00ffff", C(0, 255, 255)},
		{"empty", "", NoColor()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hexStrToColor(tc.s)
			if got != tc.want {
				t.Errorf("hexStrToColor(%q) = %v, want %v", tc.s, got, tc.want)
			}
		})
	}
}
