package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestP99_ColorToHexStr_256Color(t *testing.T) {
	c := Color{Type: 2, Val: 42} // Color256
	got := colorToHexStr(c)
	if got != "256:42" {
		t.Errorf("colorToHexStr(256:42) = %q, want '256:42'", got)
	}
}

func TestP99_ColorToHexStr_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		c    Color
		want string
	}{
		{"true color red", C(255, 0, 0), "#FF0000"},
		{"true color cyan", C(0, 255, 255), "#00FFFF"},
		{"256 color", Color{Type: 2, Val: 196}, "256:196"},
		{"256 color zero", Color{Type: 2, Val: 0}, "256:0"},
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

func TestP99_HexStrToColor_256Prefix(t *testing.T) {
	c := hexStrToColor("256:42")
	if c.Type != 2 {
		t.Errorf("Type = %d, want 2", c.Type)
	}
	if c.Val != 42 {
		t.Errorf("Val = %d, want 42", c.Val)
	}
}

func TestP99_HexStrToColor_AllFormats(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want Color
	}{
		{"hex red", "#FF0000", C(255, 0, 0)},
		{"hex cyan", "#00ffff", C(0, 255, 255)},
		{"hex black", "#000000", C(0, 0, 0)},
		{"hex white", "#FFFFFF", C(255, 255, 255)},
		{"256 prefix", "256:196", Color{Type: 2, Val: 196}},
		{"256 zero", "256:0", Color{Type: 2, Val: 0}},
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

func TestP99_SaveToFile_CurrentDir(t *testing.T) {
	// Save to current directory (dir == ".")
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tmpDir)

	path := "theme.json" // no directory component
	err := SaveToFile(Get(), path)
	if err != nil {
		t.Fatalf("SaveToFile to current dir: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}

func TestP99_LoadAndActivate_Error(t *testing.T) {
	err := LoadAndActivate("/nonexistent/path/theme.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestP99_LoadAndActivate_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activate.json")

	// Save current theme
	if err := SaveActive(path); err != nil {
		t.Fatalf("SaveActive: %v", err)
	}

	// Load and activate should succeed
	if err := LoadAndActivate(path); err != nil {
		t.Fatalf("LoadAndActivate: %v", err)
	}
}

func TestP99_SaveLoad_256ColorRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "256_theme.json")

	original := Get()
	// Modify a color to be 256
	original.Bg = Color{Type: 2, Val: 235}

	err := SaveToFile(original, path)
	if err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	loaded, err := LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}

	if loaded.Bg.Type != 2 || loaded.Bg.Val != 235 {
		t.Errorf("Loaded Bg = %v, want Type=2 Val=235", loaded.Bg)
	}
}
