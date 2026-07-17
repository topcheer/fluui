package theme

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveToFile_CreateDir_P279(t *testing.T) {
	dir := t.TempDir()
	nestedPath := filepath.Join(dir, "subdir", "deep", "theme.json")
	err := SaveToFile(Get(), nestedPath)
	if err != nil {
		t.Fatalf("should create nested dirs and save: %v", err)
	}
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("file should exist")
	}
}

func TestSaveToFile_CurrentDir_P279(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.json")
	err := SaveToFile(Get(), path)
	if err != nil {
		t.Fatalf("should save to current dir: %v", err)
	}
}

func TestLoadFromFile_InvalidJSON_P279(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	os.WriteFile(path, []byte("{invalid json"), 0644)
	_, err := LoadFromFile(path)
	if err == nil {
		t.Error("should return error for invalid JSON")
	}
}

func TestListThemeFiles_NonExistentDir_P279(t *testing.T) {
	// Non-existent dir should return nil, nil (not error)
	result, err := ListThemeFiles("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Errorf("should not error for non-existent dir: %v", err)
	}
	if result != nil {
		t.Error("should return nil for non-existent dir")
	}
}

func TestListThemeFiles_WithFiles_P279(t *testing.T) {
	dir := t.TempDir()
	// Save a theme to the dir
	SaveToFile(Get(), filepath.Join(dir, "test_theme.json"))
	// Also add a non-json file
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("hello"), 0644)
	result, err := ListThemeFiles(dir)
	if err != nil {
		t.Fatalf("should not error: %v", err)
	}
	if len(result) == 0 {
		t.Error("should find at least one theme")
	}
}
