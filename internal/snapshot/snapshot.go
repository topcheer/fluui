// Package snapshot provides deterministic buffer serialization and comparison
// utilities for testing TUI rendering.
//
// Snapshot testing captures the exact rendered output of a component (both
// visible characters and style attributes) into a comparable string. This
// enables golden-file testing patterns where changes to rendering are
// immediately visible in code review.
//
// Basic usage:
//
//	buf := buffer.NewBuffer(20, 5)
//	myComponent.Paint(buf)
//	snapshot.AssertEqual(t, expected, buf)
//
// Golden file usage:
//
//	buf := buffer.NewBuffer(20, 5)
//	myComponent.Paint(buf)
//	snapshot.Golden(t, "my_component_default", buf)
//
// Set SNAPSHOT_UPDATE=1 to update golden files.
package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Serialize converts a buffer's visible runes into a text grid.
// Each row becomes a line, cells are separated by nothing (continuous text).
// Trailing spaces are trimmed per line for cleaner output.
func Serialize(buf *buffer.Buffer) string {
	if buf == nil || buf.Width <= 0 || buf.Height <= 0 {
		return ""
	}

	var sb strings.Builder
	for y := 0; y < buf.Height; y++ {
		if y > 0 {
			sb.WriteByte('\n')
		}
		var line strings.Builder
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y*buf.Width+x]
			if cell.Rune != 0 && cell.Rune != ' ' {
				line.WriteRune(cell.Rune)
			} else {
				line.WriteByte(' ')
			}
		}
		sb.WriteString(strings.TrimRight(line.String(), " "))
	}
	return sb.String()
}

// SerializeRaw is like Serialize but preserves trailing spaces.
func SerializeRaw(buf *buffer.Buffer) string {
	if buf == nil || buf.Width <= 0 || buf.Height <= 0 {
		return ""
	}

	var sb strings.Builder
	for y := 0; y < buf.Height; y++ {
		if y > 0 {
			sb.WriteByte('\n')
		}
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y*buf.Width+x]
			if cell.Rune != 0 && cell.Rune != ' ' {
				sb.WriteRune(cell.Rune)
			} else {
				sb.WriteByte(' ')
			}
		}
	}
	return sb.String()
}

// SerializeDetailed converts a buffer into a detailed representation including
// foreground color, background color, and style flags for each cell.
// This is useful for verifying exact styling in tests.
//
// Format per cell: [rune:fg=X,bg=Y,flags=Z]
// Empty cells are shown as [ ].
func SerializeDetailed(buf *buffer.Buffer) string {
	if buf == nil || buf.Width <= 0 || buf.Height <= 0 {
		return ""
	}

	var sb strings.Builder
	for y := 0; y < buf.Height; y++ {
		if y > 0 {
			sb.WriteByte('\n')
		}
		for x := 0; x < buf.Width; x++ {
			cell := buf.Cells[y*buf.Width+x]
			sb.WriteString(formatCell(cell))
		}
	}
	return sb.String()
}

// formatCell formats a single cell for detailed serialization.
func formatCell(c buffer.Cell) string {
	if c.Rune == 0 || (c.Rune == ' ' && c.Flags == 0 && c.Fg.Type == buffer.ColorNone) {
		return "[ ]"
	}

	flags := ""
	if c.Flags&buffer.Bold != 0 {
		flags += "B"
	}
	if c.Flags&buffer.Italic != 0 {
		flags += "I"
	}
	if c.Flags&buffer.Underline != 0 {
		flags += "U"
	}
	if c.Flags&buffer.Strikethrough != 0 {
		flags += "S"
	}
	if c.Flags&buffer.Reverse != 0 {
		flags += "R"
	}
	if c.Flags&buffer.Dim != 0 {
		flags += "D"
	}
	if c.Flags&buffer.Blink != 0 {
		flags += "K"
	}
	if flags == "" {
		flags = "-"
	}

	fg := colorName(c.Fg)
	bg := colorName(c.Bg)

	return fmt.Sprintf("[%c:f=%s,b=%s,%s]", c.Rune, fg, bg, flags)
}

// colorName returns a human-readable name for common colors.
func colorName(c buffer.Color) string {
	if c.Type == buffer.ColorNone {
		return "default"
	}
	if c.Type == buffer.ColorNamed && c.Val <= 15 {
		names := []string{
			"Black", "Red", "Green", "Yellow",
			"Blue", "Magenta", "Cyan", "White",
			"BrightBlack", "BrightRed", "BrightGreen", "BrightYellow",
			"BrightBlue", "BrightMagenta", "BrightCyan", "BrightWhite",
		}
		return names[c.Val]
	}
	if c.Type == buffer.Color256 {
		return fmt.Sprintf("256:%d", c.Val)
	}
	if c.Type == buffer.ColorTrue {
		r := (c.Val >> 16) & 0xFF
		g := (c.Val >> 8) & 0xFF
		b := c.Val & 0xFF
		return fmt.Sprintf("#%02X%02X%02X", r, g, b)
	}
	return fmt.Sprintf("type=%d,val=%d", c.Type, c.Val)
}

// Diff compares two buffers and returns a visual diff string.
// Lines that differ are prefixed with:
//   "  " for matching lines
//   "- " for expected-only lines
//   "+ " for actual-only lines
//   "! " for lines that differ in content
func Diff(expected, actual *buffer.Buffer) string {
	expStr := Serialize(expected)
	actStr := Serialize(actual)

	if expStr == actStr {
		return ""
	}

	expLines := strings.Split(expStr, "\n")
	actLines := strings.Split(actStr, "\n")

	maxLines := len(expLines)
	if len(actLines) > maxLines {
		maxLines = len(actLines)
	}

	var sb strings.Builder
	for i := 0; i < maxLines; i++ {
		var exp, act string
		if i < len(expLines) {
			exp = expLines[i]
		}
		if i < len(actLines) {
			act = actLines[i]
		}

		if exp == act {
			sb.WriteString("  " + exp + "\n")
		} else if i >= len(expLines) {
			sb.WriteString("+ " + act + "\n")
		} else if i >= len(actLines) {
			sb.WriteString("- " + exp + "\n")
		} else {
			sb.WriteString("! " + act + "\n")
		}
	}
	return sb.String()
}

// AssertEqual asserts that two buffers have identical visible output.
// On failure, it prints a visual diff and fails the test.
func AssertEqual(t *testing.T, expected, actual *buffer.Buffer) {
	t.Helper()

	diff := Diff(expected, actual)
	if diff != "" {
		t.Errorf("buffer snapshot mismatch:\n%s", diff)
	}
}

// AssertEqualStr asserts that a buffer matches an expected string.
// The expected string should be in the same format as Serialize().
func AssertEqualStr(t *testing.T, expected string, actual *buffer.Buffer) {
	t.Helper()

	got := Serialize(actual)
	if expected != got {
		t.Errorf("snapshot mismatch:\nexpected:\n%s\ngot:\n%s\ndiff:\n%s",
			expected, got, strDiff(expected, got))
	}
}

// strDiff produces a line-by-line diff of two strings.
func strDiff(expected, actual string) string {
	expLines := strings.Split(expected, "\n")
	actLines := strings.Split(actual, "\n")

	maxLines := len(expLines)
	if len(actLines) > maxLines {
		maxLines = len(actLines)
	}

	var sb strings.Builder
	for i := 0; i < maxLines; i++ {
		var exp, act string
		hasExp := i < len(expLines)
		hasAct := i < len(actLines)
		if hasExp {
			exp = expLines[i]
		}
		if hasAct {
			act = actLines[i]
		}

		if hasExp && hasAct && exp == act {
			sb.WriteString("  " + exp + "\n")
		} else if hasExp && !hasAct {
			sb.WriteString("- " + exp + "\n")
		} else if !hasExp && hasAct {
			sb.WriteString("+ " + act + "\n")
		} else {
			sb.WriteString("! exp: " + exp + "\n")
			sb.WriteString("! got: " + act + "\n")
		}
	}
	return sb.String()
}

// GoldenDir is the default directory for golden snapshot files.
var GoldenDir = "testdata/snapshots"

// Golden compares a buffer against a golden snapshot file.
// If SNAPSHOT_UPDATE=1 is set in the environment, the golden file is
// created/updated instead of compared.
//
// The golden file is stored at testdata/snapshots/<name>.txt relative
// to the test file's directory.
func Golden(t *testing.T, name string, buf *buffer.Buffer) {
	t.Helper()

	actual := Serialize(buf)
	goldenPath := filepath.Join(GoldenDir, name+".txt")

	if os.Getenv("SNAPSHOT_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
			t.Fatalf("failed to create snapshot dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden snapshot: %s", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("golden snapshot not found: %s\n"+
				"Run with SNAPSHOT_UPDATE=1 to create it.\n"+
				"Actual output:\n%s", goldenPath, actual)
			return
		}
		t.Errorf("failed to read golden file %s: %v", goldenPath, err)
		return
	}

	expStr := string(expected)
	if expStr != actual {
		t.Errorf("snapshot mismatch (%s):\n%s", goldenPath, strDiff(expStr, actual))
	}
}

// GoldenDetailed is like Golden but uses SerializeDetailed for full style info.
func GoldenDetailed(t *testing.T, name string, buf *buffer.Buffer) {
	t.Helper()

	actual := SerializeDetailed(buf)
	goldenPath := filepath.Join(GoldenDir, name+".txt")

	if os.Getenv("SNAPSHOT_UPDATE") == "1" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0755); err != nil {
			t.Fatalf("failed to create snapshot dir: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actual), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("updated golden snapshot: %s", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Errorf("golden snapshot not found: %s\n"+
				"Run with SNAPSHOT_UPDATE=1 to create it.\n"+
				"Actual output:\n%s", goldenPath, actual)
			return
		}
		t.Errorf("failed to read golden file %s: %v", goldenPath, err)
		return
	}

	expStr := string(expected)
	if expStr != actual {
		t.Errorf("snapshot mismatch (%s):\n%s", goldenPath, strDiff(expStr, actual))
	}
}

// CountNonEmpty returns the number of cells with a non-zero rune.
// Useful for verifying that rendering produced output.
func CountNonEmpty(buf *buffer.Buffer) int {
	if buf == nil {
		return 0
	}
	count := 0
	for _, c := range buf.Cells {
		if c.Rune != 0 && c.Rune != ' ' {
			count++
		}
	}
	return count
}

// Region extracts a sub-region of a buffer as a serialized string.
// Useful for testing specific areas of a large buffer.
func Region(buf *buffer.Buffer, x, y, w, h int) string {
	if buf == nil || w <= 0 || h <= 0 {
		return ""
	}

	var sb strings.Builder
	for row := 0; row < h; row++ {
		if row > 0 {
			sb.WriteByte('\n')
		}
		for col := 0; col < w; col++ {
			bx := x + col
			by := y + row
			if bx < 0 || bx >= buf.Width || by < 0 || by >= buf.Height {
				sb.WriteByte(' ')
				continue
			}
			cell := buf.Cells[by*buf.Width+bx]
			if cell.Rune != 0 {
				sb.WriteRune(cell.Rune)
			} else {
				sb.WriteByte(' ')
			}
		}
	}
	return sb.String()
}

// AssertRegion asserts that a specific region of a buffer matches expected text.
func AssertRegion(t *testing.T, expected string, buf *buffer.Buffer, x, y, w, h int) {
	t.Helper()
	got := Region(buf, x, y, w, h)
	if expected != got {
		t.Errorf("region mismatch at (%d,%d,%dx%d):\nexpected:\n%s\ngot:\n%s",
			x, y, w, h, expected, got)
	}
}
