package block

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestAssistantTextCreation(t *testing.T) {
	b := NewAssistantTextBlock("asst-1")

	if b.ID() != "asst-1" {
		t.Errorf("ID() = %q, want %q", b.ID(), "asst-1")
	}
	if b.Type() != TypeAssistantText {
		t.Errorf("Type() = %v, want %v", b.Type(), TypeAssistantText)
	}
	// NewBaseBlock starts in BlockStreaming state
	if b.State() != BlockStreaming {
		t.Errorf("State() = %v, want %v", b.State(), BlockStreaming)
	}
	if b.Content() != "" {
		t.Errorf("Content() = %q, want empty", b.Content())
	}
	if !b.IsDirty() {
		t.Error("new block should be dirty")
	}
}

func TestAssistantTextAppendDelta(t *testing.T) {
	b := NewAssistantTextBlock("asst-2")

	b.AppendDelta("Hello")
	if b.Content() != "Hello" {
		t.Errorf("after first delta, Content() = %q, want %q", b.Content(), "Hello")
	}

	b.AppendDelta(" World")
	if b.Content() != "Hello World" {
		t.Errorf("after second delta, Content() = %q, want %q", b.Content(), "Hello World")
	}

	// Each AppendDelta should mark dirty
	b.ClearDirty()
	b.AppendDelta("!")
	if !b.IsDirty() {
		t.Error("AppendDelta should mark block dirty")
	}
}

func TestAssistantTextMeasureEmpty(t *testing.T) {
	b := NewAssistantTextBlock("asst-3")

	size := b.Measure(component.Unbounded())
	// Empty content should return H=1
	if size.H != 1 {
		t.Errorf("empty measure H = %d, want 1", size.H)
	}
}

func TestAssistantTextMeasureContent(t *testing.T) {
	b := NewAssistantTextBlock("asst-4")
	// Use a long string that will wrap at width 10
	b.AppendDelta("abcdefghij")
	b.AppendDelta("abcdefghij")
	// 20 chars at width 10 = 2 lines

	size := b.Measure(component.Bounded(10, 100))
	if size.H != 2 {
		t.Errorf("measure H = %d, want 2", size.H)
	}
}

func TestAssistantTextMeasureWrap(t *testing.T) {
	b := NewAssistantTextBlock("asst-5")
	// 25-char string at width 10 should wrap to 3 lines
	b.AppendDelta("abcdefghijklmnopqrstuvwxyz")

	size := b.Measure(component.Bounded(10, 100))
	if size.H < 2 {
		t.Errorf("wrapped measure H = %d, want >= 2", size.H)
	}
}

func TestAssistantTextPaint(t *testing.T) {
	b := NewAssistantTextBlock("asst-6")
	b.AppendDelta("Hi")

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("cell(0,0) rune = %q, want 'H'", cell.Rune)
	}
	cell2 := buf.GetCell(1, 0)
	if cell2.Rune != 'i' {
		t.Errorf("cell(1,0) rune = %q, want 'i'", cell2.Rune)
	}
}

func TestAssistantTextPaintEmpty(t *testing.T) {
	b := NewAssistantTextBlock("asst-7")

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)

	// Should not panic on empty content
	b.Paint(buf)

	// Buffer cells are initialized to spaces — empty paint should not write any text
	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("empty paint should leave cell blank, got rune %q", cell.Rune)
	}
}

func TestAssistantTextComplete(t *testing.T) {
	b := NewAssistantTextBlock("asst-8")
	b.AppendDelta("done")
	b.Complete()

	if b.State() != BlockComplete {
		t.Errorf("State() = %v, want %v", b.State(), BlockComplete)
	}
	if b.Duration() <= 0 {
		t.Error("Duration() should be > 0 after Complete")
	}
}

func TestAssistantTextStreamingFlow(t *testing.T) {
	b := NewAssistantTextBlock("asst-stream")

	// Simulate streaming deltas
	deltas := []string{"Hello", ", ", "world", "!"}
	expected := strings.Join(deltas, "")

	for _, d := range deltas {
		b.AppendDelta(d)
	}

	if b.Content() != expected {
		t.Errorf("Content() = %q, want %q", b.Content(), expected)
	}

	b.Complete()
	if b.State() != BlockComplete {
		t.Errorf("State() = %v, want BlockComplete", b.State())
	}
}
