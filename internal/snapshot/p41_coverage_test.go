package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- GoldenDetailed coverage tests ---

func TestGoldenDetailed_Compare(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	buf := buffer.NewBuffer(2, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedRed),
		Flags: buffer.Bold,
	}))

	// Create golden
	os.Setenv("SNAPSHOT_UPDATE", "1")
	GoldenDetailed(&testing.T{}, "gd_compare", buf)
	os.Unsetenv("SNAPSHOT_UPDATE")

	// Compare should pass
	mockT := &testing.T{}
	GoldenDetailed(mockT, "gd_compare", buf)
	if mockT.Failed() {
		t.Error("expected GoldenDetailed comparison to pass")
	}
}

func TestGoldenDetailed_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	// Create golden with one style
	buf1 := buffer.NewBuffer(1, 1)
	buf1.SetCell(0, 0, buffer.NewCell('A', buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)}))
	os.Setenv("SNAPSHOT_UPDATE", "1")
	GoldenDetailed(&testing.T{}, "gd_mismatch", buf1)
	os.Unsetenv("SNAPSHOT_UPDATE")

	// Compare with different style
	buf2 := buffer.NewBuffer(1, 1)
	buf2.SetCell(0, 0, buffer.NewCell('B', buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlue)}))

	mockT := &testing.T{}
	GoldenDetailed(mockT, "gd_mismatch", buf2)
	if !mockT.Failed() {
		t.Error("expected GoldenDetailed comparison to fail on mismatch")
	}
}

func TestGoldenDetailed_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	buf := buffer.NewBuffer(1, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))

	mockT := &testing.T{}
	GoldenDetailed(mockT, "gd_nonexistent", buf)
	if !mockT.Failed() {
		t.Error("expected failure when golden detailed file doesn't exist")
	}
}

// --- SerializeRaw coverage ---

func TestSerializeRaw_NilBuffer(t *testing.T) {
	got := SerializeRaw(nil)
	if got != "" {
		t.Errorf("expected empty for nil buffer, got %q", got)
	}
}

func TestSerializeRaw_ZeroSize(t *testing.T) {
	buf := buffer.NewBuffer(0, 0)
	got := SerializeRaw(buf)
	if got != "" {
		t.Errorf("expected empty for zero-size, got %q", got)
	}
}

func TestSerializeRaw_MultiLine(t *testing.T) {
	buf := buffer.NewBuffer(3, 2)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))
	buf.SetCell(0, 1, buffer.NewCell('B', buffer.DefaultStyle))

	got := SerializeRaw(buf)
	expected := "A  \nB  "
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

// --- SerializeDetailed coverage ---

func TestSerializeDetailed_NilBuffer(t *testing.T) {
	got := SerializeDetailed(nil)
	if got != "" {
		t.Errorf("expected empty for nil, got %q", got)
	}
}

func TestSerializeDetailed_ZeroSize(t *testing.T) {
	buf := buffer.NewBuffer(0, 0)
	got := SerializeDetailed(buf)
	if got != "" {
		t.Errorf("expected empty for zero-size, got %q", got)
	}
}

func TestSerializeDetailed_DefaultFgBg(t *testing.T) {
	buf := buffer.NewBuffer(1, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{}))

	got := SerializeDetailed(buf)
	if !strings.Contains(got, "default") {
		t.Errorf("expected 'default' for no color, got %q", got)
	}
}

// --- Golden coverage ---

func TestGolden_ReadError(t *testing.T) {
	// Point GoldenDir to a path that's a file (not a directory) to trigger read error
	tmpFile := filepath.Join(t.TempDir(), "blocker")
	os.WriteFile(tmpFile, []byte("blocker"), 0644)

	origDir := GoldenDir
	GoldenDir = tmpFile // This is a file, not a directory, so ReadFile will fail differently
	defer func() { GoldenDir = origDir }()

	buf := buffer.NewBuffer(1, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))

	// Don't set SNAPSHOT_UPDATE — should try to read
	mockT := &testing.T{}
	Golden(mockT, "test_readerr", buf)
	// This should either fail (file in path) or not-fail (if it can create subdirs)
	// Either way, no panic
}

// --- Diff coverage ---

func TestDiff_NilBuffer1(t *testing.T) {
	buf := buffer.NewBuffer(5, 1)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))

	diff := Diff(nil, buf)
	if diff == "" {
		t.Error("expected non-empty diff when comparing nil to buffer")
	}
}

func TestDiff_NilBuffer2(t *testing.T) {
	buf := buffer.NewBuffer(5, 1)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))

	diff := Diff(buf, nil)
	if diff == "" {
		t.Error("expected non-empty diff when comparing buffer to nil")
	}
}

func TestDiff_BothNil(t *testing.T) {
	diff := Diff(nil, nil)
	if diff != "" {
		t.Errorf("expected empty diff for both nil, got %q", diff)
	}
}

// --- colorName coverage ---

func TestColorName_UnknownType(t *testing.T) {
	c := buffer.Color{Type: 99, Val: 42}
	got := colorName(c)
	if !strings.Contains(got, "99") {
		t.Errorf("expected type=99 in output, got %q", got)
	}
}

func TestColorName_AllNamed(t *testing.T) {
	names := []string{"Black", "Red", "Green", "Yellow",
		"Blue", "Magenta", "Cyan", "White",
		"BrightBlack", "BrightRed", "BrightGreen", "BrightYellow",
		"BrightBlue", "BrightMagenta", "BrightCyan", "BrightWhite"}

	for i, expected := range names {
		c := buffer.NamedColor(i)
		got := colorName(c)
		if got != expected {
			t.Errorf("colorName(NamedColor(%d)) = %q, want %q", i, got, expected)
		}
	}
}

// --- Region coverage ---

func TestRegion_NegativeCoords(t *testing.T) {
	buf := buffer.NewBuffer(5, 5)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))

	got := Region(buf, -1, -1, 3, 3)
	// Negative x,y means some cells are out of bounds (rendered as spaces)
	if !strings.Contains(got, " ") {
		t.Error("expected spaces for out-of-bounds negative coords")
	}
}

func TestRegion_BufferSmallerThanRegion(t *testing.T) {
	buf := buffer.NewBuffer(2, 2)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))
	buf.SetCell(1, 0, buffer.NewCell('B', buffer.DefaultStyle))

	got := Region(buf, 0, 0, 10, 10)
	// Should return content padded with spaces
	lines := strings.Split(got, "\n")
	if len(lines) != 10 {
		t.Errorf("expected 10 lines, got %d", len(lines))
	}
}

// --- strDiff coverage ---

func TestStrDiff_BothEmpty(t *testing.T) {
	diff := strDiff("", "")
	// Two empty strings: one line comparison, should match
	if !strings.Contains(diff, "  ") {
		t.Errorf("expected matching marker for empty strings: %q", diff)
	}
}

func TestStrDiff_LineDifferInPlace(t *testing.T) {
	diff := strDiff("hello", "world")
	if !strings.Contains(diff, "! exp:") {
		t.Errorf("expected '! exp:' marker, got %q", diff)
	}
	if !strings.Contains(diff, "! got:") {
		t.Errorf("expected '! got:' marker, got %q", diff)
	}
}

// --- CountNonEmpty coverage ---

func TestCountNonEmpty_AllSpaces(t *testing.T) {
	buf := buffer.NewBuffer(5, 2)
	// All blank cells (spaces), no non-empty
	count := CountNonEmpty(buf)
	if count != 0 {
		t.Errorf("expected 0 non-empty for blank buffer, got %d", count)
	}
}

func TestCountNonEmpty_AllFilled(t *testing.T) {
	buf := buffer.NewBuffer(3, 1)
	for x := 0; x < 3; x++ {
		buf.SetCell(x, 0, buffer.NewCell('A', buffer.DefaultStyle))
	}
	count := CountNonEmpty(buf)
	if count != 3 {
		t.Errorf("expected 3 non-empty, got %d", count)
	}
}
