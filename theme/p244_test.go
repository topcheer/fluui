package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToFile_WithSubdir_P244(t *testing.T) {
	tmpDir := t.TempDir()
	subPath := filepath.Join(tmpDir, "subdir", "theme.json")
	t2 := Dracula()
	err := SaveToFile(t2, subPath)
	if err != nil {
		t.Fatalf("SaveToFile with subdir failed: %v", err)
	}
	if _, err := os.Stat(subPath); err != nil {
		t.Error("file should exist")
	}
}

func TestListThemeFiles_NonJSON_P244(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a theme"), 0644)
	t2 := Dracula()
	SaveToFile(t2, filepath.Join(tmpDir, "my.json"))
	files, err := ListThemeFiles(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range files {
		if f.Name != "" {
			found = true
		}
	}
	if !found {
		t.Error("should find my theme")
	}
}
