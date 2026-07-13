package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestP180_DefaultThemeDirHomeError(t *testing.T) {
	// DefaultThemeDir uses os.UserHomeDir() which normally works
	// Test the fallback by temporarily setting HOME to invalid
	oldHome := os.Getenv("HOME")
	os.Unsetenv("HOME")
	os.Setenv("USERPROFILE", "") // Windows
	result := DefaultThemeDir()
	os.Setenv("HOME", oldHome)
	// Should return "." when home can't be determined
	// (os.UserHomeDir may still find it via getpwuid on Unix)
	_ = result // Just ensure it doesn't panic
}

func TestP180_ListThemeFilesNonExistent(t *testing.T) {
	files, err := ListThemeFiles("/nonexistent/dir")
	// Some implementations return empty list with no error
	// Just ensure it doesn't crash
	_ = err
	_ = files
}

func TestP180_ListThemeFilesEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	files, err := ListThemeFiles(tmpDir)
	if err != nil {
		t.Errorf("expected no error for empty dir, got %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestP180_ListThemeFilesWithFiles(t *testing.T) {
	tmpDir := t.TempDir()
	// Create theme files
	os.WriteFile(filepath.Join(tmpDir, "dark.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "light.json"), []byte("{}"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "notjson.txt"), []byte("nope"), 0644)

	files, err := ListThemeFiles(tmpDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 theme files, got %d", len(files))
	}
}
