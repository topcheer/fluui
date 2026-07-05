package snapshot

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func makeBuffer(w, h int, text []string, fg buffer.Color) *buffer.Buffer {
	buf := buffer.NewBuffer(w, h)
	for y, line := range text {
		if y >= h {
			break
		}
		for x, r := range line {
			if x >= w {
				break
			}
			buf.SetCell(x, y, buffer.NewCell(r, buffer.Style{Fg: fg}))
		}
	}
	return buf
}

func TestSerialize_Basic(t *testing.T) {
	buf := makeBuffer(10, 3, []string{"Hello", "World", "Test"}, buffer.White)
	got := Serialize(buf)

	expected := "Hello\nWorld\nTest"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerialize_TrailingSpacesTrimmed(t *testing.T) {
	buf := buffer.NewBuffer(10, 2)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))
	// Row 0: "A         " -> trimmed to "A"
	// Row 1: all spaces -> trimmed to ""

	got := Serialize(buf)
	expected := "A\n"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerialize_NilBuffer(t *testing.T) {
	got := Serialize(nil)
	if got != "" {
		t.Errorf("expected empty string for nil buffer, got %q", got)
	}
}

func TestSerialize_ZeroSize(t *testing.T) {
	buf := buffer.NewBuffer(0, 0)
	got := Serialize(buf)
	if got != "" {
		t.Errorf("expected empty string for zero-size buffer, got %q", got)
	}
}

func TestSerializeRaw_PreservesSpaces(t *testing.T) {
	buf := buffer.NewBuffer(5, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.DefaultStyle))
	// "X    " (X + 4 spaces)

	got := SerializeRaw(buf)
	if got != "X    " {
		t.Errorf("expected 'X    ', got %q", got)
	}
}

func TestSerialize_SpecialChars(t *testing.T) {
	buf := buffer.NewBuffer(10, 1)
	for x, r := range []rune("┌─┐│└┘") {
		buf.SetCell(x, 0, buffer.NewCell(r, buffer.DefaultStyle))
	}
	got := Serialize(buf)
	expected := "┌─┐│└┘"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerializeDetailed_Basic(t *testing.T) {
	buf := buffer.NewBuffer(3, 1)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedRed),
		Flags: buffer.Bold,
	}))

	got := SerializeDetailed(buf)
	if !strings.Contains(got, "A") {
		t.Errorf("expected 'A' in output, got %q", got)
	}
	if !strings.Contains(got, "Red") {
		t.Errorf("expected 'Red' in output, got %q", got)
	}
	if !strings.Contains(got, "B") {
		t.Errorf("expected 'B' (bold) in output, got %q", got)
	}
}

func TestSerializeDetailed_EmptyCell(t *testing.T) {
	buf := buffer.NewBuffer(2, 1)
	got := SerializeDetailed(buf)
	if !strings.Contains(got, "[ ]") {
		t.Errorf("expected empty cell marker, got %q", got)
	}
}

func TestSerializeDetailed_AllFlags(t *testing.T) {
	buf := buffer.NewBuffer(1, 1)

	flags := []buffer.StyleFlags{
		buffer.Bold, buffer.Italic, buffer.Underline,
		buffer.Strikethrough, buffer.Reverse, buffer.Dim, buffer.Blink,
	}
	letters := []string{"B", "I", "U", "S", "R", "D", "K"}

	for i, f := range flags {
		buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{Flags: f}))
		got := SerializeDetailed(buf)
		if !strings.Contains(got, letters[i]) {
			t.Errorf("expected flag %s in output, got %q", letters[i], got)
		}
	}
}

func TestSerializeDetailed_NoFlags(t *testing.T) {
	buf := buffer.NewBuffer(1, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{
		Fg: buffer.NamedColor(buffer.NamedWhite),
	}))
	got := SerializeDetailed(buf)
	if !strings.Contains(got, ",-]") {
		t.Errorf("expected '-' for no flags, got %q", got)
	}
}

func TestSerializeDetailed_RGBColor(t *testing.T) {
	buf := buffer.NewBuffer(1, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{
		Fg: buffer.RGB(0xFF, 0x00, 0x80),
	}))
	got := SerializeDetailed(buf)
	if !strings.Contains(got, "#FF0080") {
		t.Errorf("expected RGB hex in output, got %q", got)
	}
}

func TestSerializeDetailed_256Color(t *testing.T) {
	buf := buffer.NewBuffer(1, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{
		Fg: buffer.Color256Val(42),
	}))
	got := SerializeDetailed(buf)
	if !strings.Contains(got, "256:42") {
		t.Errorf("expected 256:42 in output, got %q", got)
	}
}

func TestDiff_IdenticalBuffers(t *testing.T) {
	buf1 := makeBuffer(5, 2, []string{"Hello", "World"}, buffer.White)
	buf2 := makeBuffer(5, 2, []string{"Hello", "World"}, buffer.White)

	diff := Diff(buf1, buf2)
	if diff != "" {
		t.Errorf("expected empty diff for identical buffers, got %q", diff)
	}
}

func TestDiff_DifferentContent(t *testing.T) {
	buf1 := makeBuffer(5, 2, []string{"Hello", "World"}, buffer.White)
	buf2 := makeBuffer(5, 2, []string{"Hello", "Test!"}, buffer.White)

	diff := Diff(buf1, buf2)
	if diff == "" {
		t.Error("expected non-empty diff for different buffers")
	}
	if !strings.Contains(diff, "!") {
		t.Error("expected '!' marker for differing line")
	}
}

func TestDiff_DifferentHeights(t *testing.T) {
	buf1 := makeBuffer(5, 3, []string{"A", "B", "C"}, buffer.White)
	buf2 := makeBuffer(5, 2, []string{"A", "B"}, buffer.White)

	diff := Diff(buf1, buf2)
	if diff == "" {
		t.Error("expected non-empty diff")
	}
	if !strings.Contains(diff, "- ") {
		t.Error("expected removal marker")
	}
}

func TestAssertEqual_Matching(t *testing.T) {
	buf1 := makeBuffer(5, 1, []string{"Hello"}, buffer.White)
	buf2 := makeBuffer(5, 1, []string{"Hello"}, buffer.White)

	// Should not fail
	mockT := &testing.T{}
	AssertEqual(mockT, buf1, buf2)
	if mockT.Failed() {
		t.Error("expected no failure for matching buffers")
	}
}

func TestAssertEqual_Differing(t *testing.T) {
	buf1 := makeBuffer(5, 1, []string{"Hello"}, buffer.White)
	buf2 := makeBuffer(5, 1, []string{"World"}, buffer.White)

	mockT := &testing.T{}
	AssertEqual(mockT, buf1, buf2)
	if !mockT.Failed() {
		t.Error("expected failure for differing buffers")
	}
}

func TestAssertEqualStr_Matching(t *testing.T) {
	buf := makeBuffer(5, 1, []string{"Hello"}, buffer.White)
	mockT := &testing.T{}
	AssertEqualStr(mockT, "Hello", buf)
	if mockT.Failed() {
		t.Error("expected no failure for matching string")
	}
}

func TestAssertEqualStr_Differing(t *testing.T) {
	buf := makeBuffer(5, 1, []string{"Hello"}, buffer.White)
	mockT := &testing.T{}
	AssertEqualStr(mockT, "World", buf)
	if !mockT.Failed() {
		t.Error("expected failure for differing string")
	}
}

func TestCountNonEmpty(t *testing.T) {
	buf := buffer.NewBuffer(10, 2)
	buf.SetCell(0, 0, buffer.NewCell('A', buffer.DefaultStyle))
	buf.SetCell(1, 0, buffer.NewCell('B', buffer.DefaultStyle))
	buf.SetCell(5, 1, buffer.NewCell('C', buffer.DefaultStyle))

	count := CountNonEmpty(buf)
	if count != 3 {
		t.Errorf("expected 3 non-empty cells, got %d", count)
	}
}

func TestCountNonEmpty_Nil(t *testing.T) {
	count := CountNonEmpty(nil)
	if count != 0 {
		t.Errorf("expected 0 for nil buffer, got %d", count)
	}
}

func TestRegion_Basic(t *testing.T) {
	buf := makeBuffer(10, 5, []string{
		"AAAAAAAAAA",
		"BBBBBBBBBB",
		"CCCCCCCCCC",
		"DDDDDDDDDD",
		"EEEEEEEEEE",
	}, buffer.White)

	got := Region(buf, 2, 1, 4, 3)
	expected := "BBBB\nCCCC\nDDDD"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRegion_PartialOutOfBounds(t *testing.T) {
	buf := makeBuffer(5, 2, []string{"ABCDE", "FGHIJ"}, buffer.White)

	// Request region starting at x=3: chars 'I', 'J' + 3 spaces
	got := Region(buf, 3, 1, 5, 1)
	if got != "IJ   " {
		t.Errorf("expected 'IJ   ', got %q", got)
	}
}

func TestRegion_NilBuffer(t *testing.T) {
	got := Region(nil, 0, 0, 5, 5)
	if got != "" {
		t.Errorf("expected empty string for nil buffer, got %q", got)
	}
}

func TestRegion_ZeroSize(t *testing.T) {
	buf := buffer.NewBuffer(10, 10)
	got := Region(buf, 0, 0, 0, 0)
	if got != "" {
		t.Errorf("expected empty string for zero size region, got %q", got)
	}
}

func TestAssertRegion_Matching(t *testing.T) {
	buf := makeBuffer(5, 2, []string{"ABCDE", "FGHIJ"}, buffer.White)
	mockT := &testing.T{}
	AssertRegion(mockT, "BCD", buf, 1, 0, 3, 1)
	if mockT.Failed() {
		t.Error("expected no failure for matching region")
	}
}

func TestAssertRegion_Differing(t *testing.T) {
	buf := makeBuffer(5, 2, []string{"ABCDE", "FGHIJ"}, buffer.White)
	mockT := &testing.T{}
	AssertRegion(mockT, "XXX", buf, 1, 0, 3, 1)
	if !mockT.Failed() {
		t.Error("expected failure for differing region")
	}
}

func TestGolden_CreateAndUpdate(t *testing.T) {
	// Set up temp dir for golden files
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	buf := makeBuffer(5, 2, []string{"Hello", "World"}, buffer.White)

	// Set SNAPSHOT_UPDATE=1 to create golden file
	os.Setenv("SNAPSHOT_UPDATE", "1")
	defer os.Unsetenv("SNAPSHOT_UPDATE")

	mockT := &testing.T{}
	Golden(mockT, "test_basic", buf)

	// Verify file was created
	goldenPath := filepath.Join(GoldenDir, "test_basic.txt")
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("expected golden file to be created: %v", err)
	}

	expected := "Hello\nWorld"
	if string(data) != expected {
		t.Errorf("expected %q in golden file, got %q", expected, string(data))
	}
}

func TestGolden_Compare(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	buf := makeBuffer(5, 2, []string{"Hello", "World"}, buffer.White)

	// Create golden file
	os.Setenv("SNAPSHOT_UPDATE", "1")
	mockT1 := &testing.T{}
	Golden(mockT1, "test_compare", buf)
	os.Unsetenv("SNAPSHOT_UPDATE")

	// Now compare — should pass
	mockT2 := &testing.T{}
	Golden(mockT2, "test_compare", buf)
	if mockT2.Failed() {
		t.Error("expected golden comparison to pass")
	}
}

func TestGolden_Mismatch(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	// Create golden with first content
	buf1 := makeBuffer(5, 1, []string{"Hello"}, buffer.White)
	os.Setenv("SNAPSHOT_UPDATE", "1")
	mockT1 := &testing.T{}
	Golden(mockT1, "test_mismatch", buf1)
	os.Unsetenv("SNAPSHOT_UPDATE")

	// Compare with different content
	buf2 := makeBuffer(5, 1, []string{"World"}, buffer.White)
	mockT2 := &testing.T{}
	Golden(mockT2, "test_mismatch", buf2)
	if !mockT2.Failed() {
		t.Error("expected golden comparison to fail on mismatch")
	}
}

func TestGolden_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	buf := makeBuffer(5, 1, []string{"Test"}, buffer.White)
	// Don't set SNAPSHOT_UPDATE, don't create file

	mockT := &testing.T{}
	Golden(mockT, "nonexistent", buf)
	if !mockT.Failed() {
		t.Error("expected failure when golden file doesn't exist")
	}
}

func TestGoldenDetailed_Create(t *testing.T) {
	tmpDir := t.TempDir()
	origDir := GoldenDir
	GoldenDir = filepath.Join(tmpDir, "snapshots")
	defer func() { GoldenDir = origDir }()

	buf := buffer.NewBuffer(2, 1)
	buf.SetCell(0, 0, buffer.NewCell('X', buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedRed),
		Flags: buffer.Bold,
	}))

	os.Setenv("SNAPSHOT_UPDATE", "1")
	defer os.Unsetenv("SNAPSHOT_UPDATE")

	mockT := &testing.T{}
	GoldenDetailed(mockT, "test_detailed", buf)

	goldenPath := filepath.Join(GoldenDir, "test_detailed.txt")
	data, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("expected golden file: %v", err)
	}

	got := string(data)
	if !strings.Contains(got, "X") {
		t.Errorf("expected X in golden, got %q", got)
	}
	if !strings.Contains(got, "Red") {
		t.Errorf("expected Red in golden, got %q", got)
	}
}

func TestColorName_Default(t *testing.T) {
	c := buffer.Color{Type: buffer.ColorNone}
	got := colorName(c)
	if got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestColorName_Named(t *testing.T) {
	tests := []struct {
		n    int
		name string
	}{
		{buffer.NamedBlack, "Black"},
		{buffer.NamedRed, "Red"},
		{buffer.NamedGreen, "Green"},
		{buffer.NamedWhite, "White"},
		{buffer.NamedBrightRed, "BrightRed"},
		{buffer.NamedBrightWhite, "BrightWhite"},
	}
	for _, tc := range tests {
		c := buffer.NamedColor(tc.n)
		got := colorName(c)
		if got != tc.name {
			t.Errorf("colorName(NamedColor(%d)) = %q, want %q", tc.n, got, tc.name)
		}
	}
}

func TestColorName_256(t *testing.T) {
	c := buffer.Color256Val(128)
	got := colorName(c)
	if got != "256:128" {
		t.Errorf("expected '256:128', got %q", got)
	}
}

func TestColorName_RGB(t *testing.T) {
	c := buffer.RGB(0xFF, 0x80, 0x00)
	got := colorName(c)
	if got != "#FF8000" {
		t.Errorf("expected '#FF8000', got %q", got)
	}
}

func TestStrDiff_Identical(t *testing.T) {
	diff := strDiff("Hello\nWorld", "Hello\nWorld")
	if diff != "  Hello\n  World\n" {
		t.Errorf("unexpected diff output: %q", diff)
	}
}

func TestStrDiff_Added(t *testing.T) {
	diff := strDiff("Hello", "Hello\nWorld")
	if !strings.Contains(diff, "+ World") {
		t.Errorf("expected '+ World' in diff: %q", diff)
	}
}

func TestStrDiff_Removed(t *testing.T) {
	diff := strDiff("Hello\nWorld", "Hello")
	if !strings.Contains(diff, "- World") {
		t.Errorf("expected '- World' in diff: %q", diff)
	}
}

func TestSerialize_MultilineUnicode(t *testing.T) {
	buf := buffer.NewBuffer(6, 2)
	// Row 0: "café"
	for x, r := range []rune("café") {
		buf.SetCell(x, 0, buffer.NewCell(r, buffer.DefaultStyle))
	}
	// Row 1: "naïve"
	for x, r := range []rune("naïve") {
		buf.SetCell(x, 1, buffer.NewCell(r, buffer.DefaultStyle))
	}
	got := Serialize(buf)
	expected := "café\nnaïve"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestSerialize_DetailedWithLink(t *testing.T) {
	buf := buffer.NewBuffer(5, 1)
	cell := buffer.NewCell('L', buffer.Style{
		Fg: buffer.NamedColor(buffer.NamedBlue),
	})
	cell.Link = &buffer.Link{URL: "https://example.com"}
	buf.SetCell(0, 0, cell)

	got := SerializeDetailed(buf)
	// Should still render properly
	if !strings.Contains(got, "L") {
		t.Errorf("expected L in output, got %q", got)
	}
}
