package component

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- Test Helpers ---

// mockDirReader returns a fixed list of entries for testing.
func mockDirReader(entries []FileEntry) func(string) ([]FileEntry, error) {
	return func(dir string) ([]FileEntry, error) {
		return entries, nil
	}
}

// mockEntries creates a set of test directory entries.
func mockEntries() []FileEntry {
	return []FileEntry{
		{Name: "docs", Path: "/test/docs", IsDir: true},
		{Name: "src", Path: "/test/src", IsDir: true},
		{Name: "main.go", Path: "/test/main.go", IsDir: false, Size: 1024},
		{Name: "readme.md", Path: "/test/readme.md", IsDir: false, Size: 512},
		{Name: "test.go", Path: "/test/test.go", IsDir: false, Size: 256},
	}
}

// newTestFilePicker creates a FilePicker with mock entries.
func newTestFilePicker(entries []FileEntry) *FilePicker {
	fp := NewFilePicker("/test")
	fp.SetDirReader(mockDirReader(entries))
	fp.loadDir("/test")
	return fp
}

// makeKeyEvent creates a KeyEvent with the given key and rune.
func makeKeyEvent(key term.KeyCode, r rune) *term.KeyEvent {
	return &term.KeyEvent{Key: key, Rune: r}
}

// --- Tests ---

// 1. Construction and defaults
func TestFilePicker_New(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	if fp == nil {
		t.Fatal("NewFilePicker returned nil")
	}
	if fp.Cwd() != "/test" {
		t.Errorf("Cwd = %q, want %q", fp.Cwd(), "/test")
	}
	if fp.FilteredCount() != 5 {
		t.Errorf("FilteredCount = %d, want 5", fp.FilteredCount())
	}
	if fp.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", fp.Cursor())
	}
}

// 2. Empty directory
func TestFilePicker_EmptyDir(t *testing.T) {
	fp := newTestFilePicker([]FileEntry{})
	if fp.FilteredCount() != 0 {
		t.Errorf("FilteredCount = %d, want 0", fp.FilteredCount())
	}
	entry, ok := fp.CurrentEntry()
	if ok {
		t.Errorf("CurrentEntry on empty dir should return ok=false, got %+v", entry)
	}
}

// 3. Directory sorting (dirs first, then alphabetical)
func TestFilePicker_Sorting(t *testing.T) {
	// Unsorted input
	entries := []FileEntry{
		{Name: "zebra.go", IsDir: false},
		{Name: "alpha", IsDir: true},
		{Name: "beta.go", IsDir: false},
		{Name: "gamma", IsDir: true},
	}
	fp := newTestFilePicker(entries)
	result := fp.Entries()
	if len(result) != 4 {
		t.Fatalf("Entries len = %d, want 4", len(result))
	}
	// Directories first: alpha, gamma, then files: beta.go, zebra.go
	if result[0].Name != "alpha" {
		t.Errorf("result[0].Name = %q, want alpha", result[0].Name)
	}
	if result[1].Name != "gamma" {
		t.Errorf("result[1].Name = %q, want gamma", result[1].Name)
	}
	if result[2].Name != "beta.go" {
		t.Errorf("result[2].Name = %q, want beta.go", result[2].Name)
	}
	if result[3].Name != "zebra.go" {
		t.Errorf("result[3].Name = %q, want zebra.go", result[3].Name)
	}
}

// 4. MoveDown / MoveUp
func TestFilePicker_MoveDownUp(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.MoveDown()
	if fp.Cursor() != 1 {
		t.Errorf("after MoveDown, Cursor = %d, want 1", fp.Cursor())
	}
	fp.MoveDown()
	fp.MoveDown()
	if fp.Cursor() != 3 {
		t.Errorf("after 3x MoveDown, Cursor = %d, want 3", fp.Cursor())
	}
	fp.MoveUp()
	if fp.Cursor() != 2 {
		t.Errorf("after MoveUp, Cursor = %d, want 2", fp.Cursor())
	}
}

// 5. MoveDown wraps around
func TestFilePicker_MoveDownWrap(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// Move to last entry
	fp.SetCursor(fp.FilteredCount() - 1)
	// Move down past end → should wrap to 0
	fp.MoveDown()
	if fp.Cursor() != 0 {
		t.Errorf("after MoveDown at end, Cursor = %d, want 0 (wrap)", fp.Cursor())
	}
}

// 6. MoveUp wraps around
func TestFilePicker_MoveUpWrap(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// At position 0, move up → should wrap to last
	fp.MoveUp()
	if fp.Cursor() != fp.FilteredCount()-1 {
		t.Errorf("after MoveUp at start, Cursor = %d, want %d (wrap)", fp.Cursor(), fp.FilteredCount()-1)
	}
}

// 7. SetCursor clamping
func TestFilePicker_SetCursorClamp(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetCursor(-10)
	if fp.Cursor() != 0 {
		t.Errorf("SetCursor(-10): Cursor = %d, want 0", fp.Cursor())
	}
	fp.SetCursor(999)
	if fp.Cursor() != fp.FilteredCount()-1 {
		t.Errorf("SetCursor(999): Cursor = %d, want %d", fp.Cursor(), fp.FilteredCount()-1)
	}
}

// 8. CurrentEntry
func TestFilePicker_CurrentEntry(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	entry, ok := fp.CurrentEntry()
	if !ok {
		t.Fatal("CurrentEntry returned ok=false")
	}
	// First entry should be first directory "docs"
	if entry.Name != "docs" {
		t.Errorf("CurrentEntry Name = %q, want docs", entry.Name)
	}
	if !entry.IsDir {
		t.Error("CurrentEntry should be a directory")
	}
}

// 9. EnterDir enters a directory
func TestFilePicker_EnterDir(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// Cursor is at "docs" (first dir), EnterDir should navigate into it

	// Set up entries for the subdirectory
	subEntries := []FileEntry{
		{Name: "file1.txt", Path: "/test/docs/file1.txt", IsDir: false},
	}
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		if dir == "/test/docs" {
			return subEntries, nil
		}
		return mockEntries(), nil
	})

	fp.EnterDir()
	if fp.Cwd() != "/test/docs" {
		t.Errorf("Cwd after EnterDir = %q, want /test/docs", fp.Cwd())
	}
}

// 10. EnterDir on file fires OnConfirm
func TestFilePicker_EnterOnFile(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	confirmed := false
	fp.SetOnConfirm(func(entry FileEntry) {
		if entry.Name == "main.go" {
			confirmed = true
		}
	})

	// Move cursor to first file ("main.go" is at index 2 after 2 dirs)
	fp.SetCursor(2)
	fp.EnterDir()
	if !confirmed {
		t.Error("EnterDir on file did not fire OnConfirm")
	}
}

// 11. GoUp navigates to parent
func TestFilePicker_GoUp(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return mockEntries(), nil
	})

	fp.GoUp()
	// Should navigate to parent of /test which is /
	parent := fp.Cwd()
	if parent == "/test" {
		t.Error("GoUp did not change directory")
	}
}

// 12. Filter: set and query
func TestFilePicker_SetFilter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFilter("go")
	// Only files containing "go" should match: main.go, test.go
	if fp.FilteredCount() != 2 {
		t.Errorf("FilteredCount with 'go' = %d, want 2", fp.FilteredCount())
	}
	if fp.Filter() != "go" {
		t.Errorf("Filter = %q, want 'go'", fp.Filter())
	}
}

// 13. Filter: clear
func TestFilePicker_ClearFilter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFilter("go")
	fp.SetFilter("")
	if fp.FilteredCount() != 5 {
		t.Errorf("FilteredCount after clear = %d, want 5", fp.FilteredCount())
	}
}

// 14. Filter mode: enter/exit
func TestFilePicker_FilterMode(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFiltering(true)
	if !fp.IsFiltering() {
		t.Error("IsFiltering should be true after SetFiltering(true)")
	}
	fp.SetFiltering(false)
	if fp.IsFiltering() {
		t.Error("IsFiltering should be false after SetFiltering(false)")
	}
}

// 15. Filter: append and backspace
func TestFilePicker_AppendBackspaceFilter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.AppendFilter('t')
	fp.AppendFilter('e')
	fp.AppendFilter('s')
	fp.AppendFilter('t')
	// "test" should match: test.go, (docs/src don't match fuzzy)
	if fp.Filter() != "test" {
		t.Errorf("Filter = %q, want 'test'", fp.Filter())
	}
	if fp.FilteredCount() != 1 {
		t.Errorf("FilteredCount with 'test' = %d, want 1", fp.FilteredCount())
	}

	fp.BackspaceFilter()
	if fp.Filter() != "tes" {
		t.Errorf("Filter after backspace = %q, want 'tes'", fp.Filter())
	}
}

// 16. Multi-select: toggle
func TestFilePicker_ToggleSelect(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// Move to first file (index 2 = main.go)
	fp.SetCursor(2)
	fp.ToggleSelect()
	if !fp.IsSelected("/test/main.go") {
		t.Error("main.go should be selected after toggle")
	}
	// Toggle again to deselect
	fp.ToggleSelect()
	if fp.IsSelected("/test/main.go") {
		t.Error("main.go should be deselected after second toggle")
	}
}

// 17. Multi-select: SelectedFiles
func TestFilePicker_SelectedFiles(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetCursor(2) // main.go
	fp.ToggleSelect()
	fp.SetCursor(4) // test.go
	fp.ToggleSelect()
	sel := fp.SelectedFiles()
	if len(sel) != 2 {
		t.Fatalf("SelectedFiles len = %d, want 2", len(sel))
	}
}

// 18. Multi-select: ClearSelection
func TestFilePicker_ClearSelection(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetCursor(2)
	fp.ToggleSelect()
	fp.ClearSelection()
	if len(fp.SelectedFiles()) != 0 {
		t.Error("SelectedFiles should be empty after ClearSelection")
	}
}

// 19. Multi-select: toggle on directory does nothing
func TestFilePicker_ToggleSelectDir(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetCursor(0) // "docs" directory
	fp.ToggleSelect()
	sel := fp.SelectedFiles()
	if len(sel) != 0 {
		t.Errorf("Selecting a directory should not add to selected, got %d", len(sel))
	}
}

// 20. HandleKey: j/k navigation
func TestFilePicker_HandleKey_Navigation(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// 'j' should move down
	fp.HandleKey(makeKeyEvent(0, 'j'))
	if fp.Cursor() != 1 {
		t.Errorf("after 'j', Cursor = %d, want 1", fp.Cursor())
	}
	// 'k' should move up
	fp.HandleKey(makeKeyEvent(0, 'k'))
	if fp.Cursor() != 0 {
		t.Errorf("after 'k', Cursor = %d, want 0", fp.Cursor())
	}
}

// 21. HandleKey: Enter key
func TestFilePicker_HandleKey_Enter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// Cursor at 0 (docs dir), Enter should navigate
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		if dir == "/test/docs" {
			return []FileEntry{{Name: "a.txt", Path: "/test/docs/a.txt"}}, nil
		}
		return mockEntries(), nil
	})
	consumed := fp.HandleKey(makeKeyEvent(term.KeyEnter, 0))
	if !consumed {
		t.Error("HandleKey Enter should be consumed")
	}
	if fp.Cwd() != "/test/docs" {
		t.Errorf("Cwd after Enter on dir = %q, want /test/docs", fp.Cwd())
	}
}

// 22. HandleKey: filter mode '/' enters filter
func TestFilePicker_HandleKey_EnterFilter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	consumed := fp.HandleKey(makeKeyEvent(0, '/'))
	if !consumed {
		t.Error("HandleKey '/' should be consumed")
	}
	if !fp.IsFiltering() {
		t.Error("Filtering should be true after '/'")
	}
}

// 23. HandleKey: Escape exits filter mode
func TestFilePicker_HandleKey_EscapeFilter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFiltering(true)
	fp.AppendFilter('g')
	fp.HandleKey(makeKeyEvent(term.KeyEscape, 0))
	if fp.IsFiltering() {
		t.Error("Filtering should be false after Escape")
	}
}

// 24. HandleKey: nil key returns false
func TestFilePicker_HandleKey_Nil(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	if fp.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

// 25. Paint: no panic
func TestFilePicker_Paint_NoPanic(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Paint panicked: %v", r)
		}
	}()
	fp.Paint(buf)
}

// 26. Paint: zero bounds
func TestFilePicker_Paint_ZeroBounds(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	buf := buffer.NewBuffer(40, 20)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Paint with zero bounds panicked: %v", r)
		}
	}()
	fp.Paint(buf)
}

// 27. Measure: returns reasonable size
func TestFilePicker_Measure(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	s := fp.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W < 30 {
		t.Errorf("Measure W = %d, want >= 30", s.W)
	}
	if s.H < 5 {
		t.Errorf("Measure H = %d, want >= 5", s.H)
	}
}

// 28. Measure: respects constraints
func TestFilePicker_Measure_Constraints(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	s := fp.Measure(Constraints{MaxWidth: 30, MaxHeight: 10})
	if s.W > 30 {
		t.Errorf("Measure W = %d, want <= 30 (MaxWidth)", s.W)
	}
	if s.H > 10 {
		t.Errorf("Measure H = %d, want <= 10 (MaxHeight)", s.H)
	}
}

// 29. Children returns nil
func TestFilePicker_Children(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	if fp.Children() != nil {
		t.Error("Children should return nil for leaf component")
	}
}

// 30. String returns description
func TestFilePicker_String(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	s := fp.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}

// 31. SetStyle / Style
func TestFilePicker_SetGetStyle(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	custom := DefaultFilePickerStyle()
	custom.Normal.Fg = buffer.RGB(255, 255, 255)
	fp.SetStyle(custom)
	got := fp.Style()
	if got.Normal.Fg != buffer.RGB(255, 255, 255) {
		t.Error("Style round-trip failed")
	}
}

// 32. DefaultFilePickerStyle returns valid styles
func TestFilePicker_DefaultStyle(t *testing.T) {
	s := DefaultFilePickerStyle()
	if s.Normal.Fg == s.Normal.Bg {
		t.Error("Default style Normal Fg == Bg, likely uninitialized")
	}
}

// 33. HandleKey: g/G vim navigation
func TestFilePicker_HandleKey_VimNav(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	// Move to middle
	fp.SetCursor(2)
	// 'G' should jump to last
	fp.HandleKey(makeKeyEvent(0, 'G'))
	if fp.Cursor() != fp.FilteredCount()-1 {
		t.Errorf("after 'G', Cursor = %d, want %d", fp.Cursor(), fp.FilteredCount()-1)
	}
	// 'g' should jump to first
	fp.HandleKey(makeKeyEvent(0, 'g'))
	if fp.Cursor() != 0 {
		t.Errorf("after 'g', Cursor = %d, want 0", fp.Cursor())
	}
}

// 34. HandleKey: Backspace goes to parent
func TestFilePicker_HandleKey_BackspaceParent(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return mockEntries(), nil
	})
	fp.HandleKey(makeKeyEvent(term.KeyBackspace, 0))
	// Cwd should change (go to parent)
	if fp.Cwd() == "/test" {
		t.Error("Backspace should navigate to parent")
	}
}

// 35. HandleKey: Space toggles select
func TestFilePicker_HandleKey_Space(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetCursor(2) // first file (main.go)
	fp.HandleKey(makeKeyEvent(term.KeySpace, ' '))
	if !fp.IsSelected("/test/main.go") {
		t.Error("Space should toggle selection on current file")
	}
}

// 36. OnSelect callback fires
func TestFilePicker_OnSelectCallback(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	called := false
	fp.SetOnSelect(func(entry FileEntry) {
		called = true
	})
	fp.SetCursor(2) // main.go
	fp.ToggleSelect()
	if !called {
		t.Error("OnSelect callback should fire on ToggleSelect")
	}
}

// 37. OnError callback fires on bad directory
func TestFilePicker_OnErrorCallback(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	errFired := false
	fp.SetOnError(func(err error) {
		errFired = true
	})
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return nil, os.ErrNotExist
	})
	fp.loadDir("/nonexistent")
	if !errFired {
		t.Error("OnError callback should fire on read failure")
	}
}

// 38. Concurrency: concurrent access
func TestFilePicker_ConcurrentAccess(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fp.MoveDown()
			fp.MoveUp()
			fp.SetCursor(0)
			_ = fp.Cursor()
			_ = fp.Entries()
			_ = fp.FilteredCount()
			fp.SetFilter("go")
			fp.SetFilter("")
			fp.ToggleSelect()
			fp.ClearSelection()
		}()
	}
	wg.Wait()
}

// 39. Concurrency: concurrent paint
func TestFilePicker_ConcurrentPaint(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(40, 20)
			fp.Paint(buf)
		}()
	}
	wg.Wait()
}

// 40. SetBounds updates bounds
func TestFilePicker_SetBounds(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	r := Rect{X: 5, Y: 10, W: 50, H: 30}
	fp.SetBounds(r)
	if fp.Bounds() != r {
		t.Errorf("Bounds = %+v, want %+v", fp.Bounds(), r)
	}
}

// 41. Real directory test (uses actual filesystem)
func TestFilePicker_RealDir(t *testing.T) {
	// Create temp dir with known files
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.go"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("test"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	fp := NewFilePicker(tmpDir)
	if fp.FilteredCount() != 3 {
		t.Errorf("FilteredCount = %d, want 3", fp.FilteredCount())
	}
	// First entry should be subdir (directory first)
	entry, ok := fp.CurrentEntry()
	if !ok {
		t.Fatal("CurrentEntry returned false")
	}
	if entry.Name != "subdir" {
		t.Errorf("First entry Name = %q, want subdir", entry.Name)
	}
	if !entry.IsDir {
		t.Error("First entry should be a directory")
	}
}

// 42. Filter handles empty query
func TestFilePicker_EmptyFilter(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFilter("")
	if fp.FilteredCount() != 5 {
		t.Errorf("Empty filter: FilteredCount = %d, want 5", fp.FilteredCount())
	}
}

// 43. Filter: no matches
func TestFilePicker_FilterNoMatch(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFilter("zzzznomatch")
	if fp.FilteredCount() != 0 {
		t.Errorf("No-match filter: FilteredCount = %d, want 0", fp.FilteredCount())
	}
	entry, ok := fp.CurrentEntry()
	if ok {
		t.Errorf("CurrentEntry on empty filter should return false, got %+v", entry)
	}
}

// 44. BackspaceFilter on empty query does nothing
func TestFilePicker_BackspaceEmpty(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.BackspaceFilter() // should not panic
	if fp.Filter() != "" {
		t.Errorf("Filter after backspace on empty = %q, want empty", fp.Filter())
	}
}

// 45. Filter mode: typing then exiting
func TestFilePicker_FilterTypingFlow(t *testing.T) {
	fp := newTestFilePicker(mockEntries())
	fp.SetFiltering(true)
	fp.AppendFilter('g')
	fp.AppendFilter('o')
	if fp.FilteredCount() != 2 {
		t.Errorf("After typing 'go': FilteredCount = %d, want 2", fp.FilteredCount())
	}
	// Exit filter mode clears filter
	fp.SetFiltering(false)
	if fp.FilteredCount() != 5 {
		t.Errorf("After exit filter mode: FilteredCount = %d, want 5", fp.FilteredCount())
	}
}
