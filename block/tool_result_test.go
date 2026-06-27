package block

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestToolResultCreation(t *testing.T) {
	b := NewToolResultBlock("tr-1")

	if b.ID() != "tr-1" {
		t.Errorf("ID() = %q, want 'tr-1'", b.ID())
	}
	if b.Type() != TypeToolResult {
		t.Errorf("Type() = %v, want TypeToolResult", b.Type())
	}
	if b.State() != BlockStreaming {
		t.Errorf("State() = %v, want BlockStreaming", b.State())
	}
	if b.Output() != "" {
		t.Errorf("Output() = %q, want empty", b.Output())
	}
	if b.Collapsed() {
		t.Error("Collapsed() = true, want false (new block should be expanded)")
	}
}

func TestToolResultAppendDelta(t *testing.T) {
	b := NewToolResultBlock("tr-2")

	b.ClearDirty()
	b.AppendDelta("hello ")
	b.AppendDelta("world")

	if b.Output() != "hello world" {
		t.Errorf("Output() = %q, want 'hello world'", b.Output())
	}
	if !b.IsDirty() {
		t.Error("AppendDelta should mark dirty")
	}
}

func TestToolResultAutoCollapse(t *testing.T) {
	b := NewToolResultBlock("tr-3")

	// Append 6 lines of output (maxPreview is 5)
	for i := 0; i < 6; i++ {
		b.AppendDelta("line " + string(rune('A'+i)) + "\n")
	}

	if b.Collapsed() {
		t.Error("Collapsed() should be false before Complete")
	}

	b.Complete()

	if !b.Collapsed() {
		t.Error("Collapsed() should be true after Complete with >5 lines")
	}
}

func TestToolResultNoCollapseShortOutput(t *testing.T) {
	b := NewToolResultBlock("tr-3b")

	// Append only 3 lines — should NOT auto-collapse
	for i := 0; i < 3; i++ {
		b.AppendDelta("short line\n")
	}
	b.Complete()

	if b.Collapsed() {
		t.Error("Collapsed() should be false after Complete with <=5 lines")
	}
}

func TestToolResultToggle(t *testing.T) {
	b := NewToolResultBlock("tr-4")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\n")
	b.Complete()

	// Should be collapsed after Complete
	if !b.Collapsed() {
		t.Fatal("should be collapsed after Complete")
	}

	// Toggle to expand
	b.Toggle()
	if b.Collapsed() {
		t.Error("Collapsed() should be false after first Toggle")
	}

	// Toggle back to collapse
	b.Toggle()
	if !b.Collapsed() {
		t.Error("Collapsed() should be true after second Toggle")
	}
}

func TestToolResultMeasureCollapsed(t *testing.T) {
	b := NewToolResultBlock("tr-5")

	// Add enough lines to exceed maxPreview (5)
	for i := 0; i < 8; i++ {
		b.AppendDelta("output line " + string(rune('A'+i)) + "\n")
	}
	b.Complete()

	if !b.Collapsed() {
		t.Fatal("precondition: should be collapsed")
	}

	size := b.Measure(component.Bounded(80, 100))
	wantH := 5 + 2 // maxPreview + 2 (borders)
	if size.H != wantH {
		t.Errorf("Measure (collapsed) H = %d, want %d", size.H, wantH)
	}
}

func TestToolResultMeasureExpanded(t *testing.T) {
	b := NewToolResultBlock("tr-6")

	// Add 4 lines of output
	b.AppendDelta("line A\nline B\nline C\nline D")

	// Not collapsed (short output, not completed or toggled)
	size := b.Measure(component.Bounded(80, 100))
	wantH := 4 + 2 // 4 content lines + 2 borders
	if size.H != wantH {
		t.Errorf("Measure (expanded) H = %d, want %d", size.H, wantH)
	}
}

func TestToolResultMeasureEmpty(t *testing.T) {
	b := NewToolResultBlock("tr-6b")

	// No output at all
	size := b.Measure(component.Bounded(80, 100))
	wantH := 2 // just borders, no content
	if size.H != wantH {
		t.Errorf("Measure (empty) H = %d, want %d", size.H, wantH)
	}
}

func TestToolResultPaint(t *testing.T) {
	b := NewToolResultBlock("tr-7")
	b.AppendDelta("result line 1\nresult line 2")
	b.Complete()

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 4})

	buf := buffer.NewBuffer(40, 4)
	b.Paint(buf)

	// Top border should have ╭ at (0,0)
	topLeft := buf.GetCell(0, 0)
	if topLeft.Rune != '╭' {
		t.Errorf("Paint: top-left border rune = %q, want '╭'", topLeft.Rune)
	}

	// Bottom border should have ╰ at (0, H-1)
	botLeft := buf.GetCell(0, 3)
	if botLeft.Rune != '╰' {
		t.Errorf("Paint: bottom-left border rune = %q, want '╰'", botLeft.Rune)
	}

	// Content area should have "result" text somewhere in row 1
	contentFound := false
	for x := 0; x < 40; x++ {
		cell := buf.GetCell(x, 1)
		if cell.Rune == 'r' {
			// Check if it's "result"
			rest := readRunes(buf, x, 1, 6)
			if strings.HasPrefix(rest, "result") {
				contentFound = true
				break
			}
		}
	}
	if !contentFound {
		t.Error("Paint: expected 'result' content in row 1")
	}
}

// readRunes reads up to n runes from buf starting at (x, y) and returns them as a string.
// Does not stop on spaces — only stops on null runes.
func readRunes(buf *buffer.Buffer, x, y, n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		c := buf.GetCell(x+i, y)
		if c.Rune == 0 {
			break
		}
		sb.WriteRune(c.Rune)
	}
	return sb.String()
}
