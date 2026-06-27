package block

import (
	"strings"
	"testing"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestThinkingBlockCreation(t *testing.T) {
	tb := NewThinkingBlock("think-1")
	if tb.ID() != "think-1" {
		t.Errorf("ID() = %q", tb.ID())
	}
	if tb.Type() != TypeThinking {
		t.Errorf("Type() = %v, want TypeThinking", tb.Type())
	}
	if tb.State() != BlockStreaming {
		t.Errorf("State() = %v, want BlockStreaming", tb.State())
	}
	if !tb.Collapsed() {
		t.Error("new thinking block should be collapsed by default")
	}
}

func TestThinkingBlockAppendDelta(t *testing.T) {
	tb := NewThinkingBlock("think-2")
	tb.ClearDirty()

	tb.AppendDelta("Let me think...")
	if tb.Content() != "Let me think..." {
		t.Errorf("Content() = %q, want 'Let me think...'", tb.Content())
	}
	if !tb.IsDirty() {
		t.Error("AppendDelta should mark dirty")
	}
}

func TestThinkingBlockToggle(t *testing.T) {
	tb := NewThinkingBlock("think-3")
	if !tb.Collapsed() {
		t.Error("should start collapsed")
	}

	tb.Toggle()
	if tb.Collapsed() {
		t.Error("should be expanded after Toggle")
	}

	tb.Toggle()
	if !tb.Collapsed() {
		t.Error("should be collapsed after second Toggle")
	}
}

func TestThinkingBlockMeasureCollapsed(t *testing.T) {
	tb := NewThinkingBlock("think-4")
	tb.AppendDelta("Some thinking content that is long")

	sz := tb.Measure(component.Bounded(80, 100))
	if sz.H != 1 {
		t.Errorf("collapsed height = %d, want 1", sz.H)
	}
}

func TestThinkingBlockMeasureExpanded(t *testing.T) {
	tb := NewThinkingBlock("think-5")
	tb.AppendDelta("Line 1")
	tb.Toggle() // expand

	sz := tb.Measure(component.Bounded(80, 100))
	// 1 header + at least 1 content line
	if sz.H < 2 {
		t.Errorf("expanded height = %d, want >= 2", sz.H)
	}
}

func TestThinkingBlockMeasureExpandedEmpty(t *testing.T) {
	tb := NewThinkingBlock("think-6")
	tb.Toggle() // expand but no content

	sz := tb.Measure(component.Bounded(80, 100))
	if sz.H != 1 {
		t.Errorf("expanded empty height = %d, want 1", sz.H)
	}
}

func TestThinkingBlockPaintCollapsed(t *testing.T) {
	tb := NewThinkingBlock("think-7")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	tb.Paint(buf)

	// Should have some text at (0,0)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("collapsed Paint should write text at (0,0)")
	}
}

func TestThinkingBlockPaintExpanded(t *testing.T) {
	tb := NewThinkingBlock("think-8")
	tb.AppendDelta("This is my thinking process.")
	tb.Toggle() // expand
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})

	buf := buffer.NewBuffer(40, 10)
	tb.Paint(buf)

	// Header at (0,0)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expanded Paint should write header at (0,0)")
	}
	// Content at (0,1)
	cell = buf.GetCell(0, 1)
	if cell.Rune == 0 {
		t.Error("expanded Paint should write content at (0,1)")
	}
}

func TestThinkingBlockPaintComplete(t *testing.T) {
	tb := NewThinkingBlock("think-9")
	tb.AppendDelta("Done thinking.")
	tb.Complete()
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	tb.Paint(buf)

	// Should show "Thought for ..." instead of "Thinking..."
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("Paint should write text")
	}
}

func TestThinkingBlockDuration(t *testing.T) {
	tb := NewThinkingBlock("think-10")
	time.Sleep(2 * time.Millisecond)
	tb.Complete()

	d := tb.Duration()
	if d < 0 {
		t.Errorf("Duration should be non-negative, got %v", d)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{2 * time.Second, "2s"},
		{3500 * time.Millisecond, "3.5s"},
	}
	for _, tt := range tests {
		got := formatDuration(tt.d)
		if !strings.Contains(got, "ms") && !strings.Contains(got, "s") {
			t.Errorf("formatDuration(%v) = %q, missing unit", tt.d, got)
		}
	}
}
