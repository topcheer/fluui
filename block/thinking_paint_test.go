package block

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestThinkingPaintCollapsedContent(t *testing.T) {
	b := NewThinkingBlock("think-1")
	b.AppendDelta("Let me reason about this problem.")

	// Default: collapsed
	if !b.Collapsed() {
		t.Fatal("new ThinkingBlock should be collapsed by default")
	}

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.Paint(buf)

	// Collapsed streaming state should contain "Thinking"
	text := rowToString(buf, 0, 80)
	if !strings.Contains(text, "Thinking") && !strings.Contains(text, "Thought") {
		t.Errorf("collapsed text = %q, want to contain 'Thinking' or 'Thought'", text)
	}
}

func TestThinkingPaintExpandedHeader(t *testing.T) {
	b := NewThinkingBlock("think-2")
	b.AppendDelta("some reasoning")

	// Expand
	b.Toggle()
	if b.Collapsed() {
		t.Fatal("should be expanded after Toggle")
	}

	// Expanded: header at row 0 + content at row 1+
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)

	// Header row should contain ▼
	header := rowToString(buf, 0, 80)
	if !strings.Contains(header, "▼") {
		t.Errorf("expanded header = %q, want to contain '▼'", header)
	}
}

func TestThinkingPaintExpandedContent(t *testing.T) {
	b := NewThinkingBlock("think-3")
	b.AppendDelta("Deep reasoning about the universe and everything.")

	b.Toggle() // expand

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)

	// Row 1+ should contain thinking content (starts with │ prefix)
	contentRow := rowToString(buf, 1, 80)
	if contentRow == "" || strings.TrimSpace(contentRow) == "" {
		t.Errorf("expanded content row 1 is empty, expected thinking content")
	}
	if !strings.Contains(contentRow, "│") {
		t.Errorf("content row 1 = %q, want to contain '│' prefix", contentRow)
	}
	if !strings.Contains(contentRow, "Deep") {
		t.Errorf("content row 1 = %q, want to contain 'Deep'", contentRow)
	}
}

func TestThinkingPaintComplete(t *testing.T) {
	b := NewThinkingBlock("think-4")
	b.AppendDelta("analysis complete")
	b.Complete()

	// Should be collapsed by default after complete
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.Paint(buf)

	// After Complete, should show "Thought for"
	text := rowToString(buf, 0, 80)
	if !strings.Contains(text, "Thought") {
		t.Errorf("completed collapsed text = %q, want 'Thought'", text)
	}
	if strings.Contains(text, "Thinking...") {
		t.Errorf("completed text should not contain 'Thinking...', got %q", text)
	}
}

func TestThinkingMeasureCollapsedVsExpanded(t *testing.T) {
	b := NewThinkingBlock("think-5")
	b.AppendDelta("abcdefghijabcdefghij") // 20 chars

	// Collapsed: H=1
	collapsedSize := b.Measure(component.Bounded(80, 100))
	if collapsedSize.H != 1 {
		t.Errorf("collapsed H = %d, want 1", collapsedSize.H)
	}

	// Expanded: H = 1 (header) + content lines
	b.Toggle()
	expandedSize := b.Measure(component.Bounded(10, 100))
	// 20 chars at width 10 → at least 2 content lines + 1 header = 3
	if expandedSize.H < 3 {
		t.Errorf("expanded H = %d, want >= 3 (header + wrapped content)", expandedSize.H)
	}
}

func TestThinkingToggleDirty(t *testing.T) {
	b := NewThinkingBlock("think-6")
	b.ClearDirty()

	b.Toggle()
	if !b.IsDirty() {
		t.Error("should be dirty after Toggle")
	}

	b.ClearDirty()
	b.Toggle() // toggle back
	if !b.IsDirty() {
		t.Error("should be dirty after second Toggle")
	}
}

func TestThinkingEmptyContent(t *testing.T) {
	b := NewThinkingBlock("think-7")
	// No content appended

	// Collapsed measure
	size := b.Measure(component.Bounded(80, 100))
	if size.H != 1 {
		t.Errorf("empty collapsed H = %d, want 1", size.H)
	}

	// Expanded measure with no content
	b.Toggle()
	size = b.Measure(component.Bounded(80, 100))
	if size.H != 1 {
		t.Errorf("empty expanded H = %d, want 1 (header only)", size.H)
	}

	// Paint should not panic
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.Paint(buf)
}

func TestThinkingCompleteExpandedPaint(t *testing.T) {
	b := NewThinkingBlock("think-8")
	b.AppendDelta("finished reasoning")
	b.Complete()
	b.Toggle() // expand

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)

	// Header should have ▼ and "Thought for" (not "Thinking...")
	header := rowToString(buf, 0, 80)
	if !strings.Contains(header, "▼") {
		t.Errorf("expanded completed header = %q, want '▼'", header)
	}
	if !strings.Contains(header, "Thought") {
		t.Errorf("expanded completed header = %q, want 'Thought'", header)
	}
}

// --- helpers ---

// rowToString reads a full row from the buffer into a string.
func rowToString(buf *buffer.Buffer, y, w int) string {
	var sb strings.Builder
	for x := 0; x < w; x++ {
		cell := buf.GetCell(x, y)
		if cell.Rune != 0 {
			sb.WriteRune(cell.Rune)
		} else {
			sb.WriteRune(' ')
		}
	}
	return strings.TrimRight(sb.String(), " ")
}
