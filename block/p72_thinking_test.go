package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestP72_ThinkingSpinner_AdvanceSpinner(t *testing.T) {
	b := NewThinkingBlock("th-sp-1")
	initial := b.SpinnerFrame()
	b.AdvanceSpinner()
	after := b.SpinnerFrame()
	if initial == after {
		t.Error("expected spinner frame to change after AdvanceSpinner")
	}
}

func TestP72_ThinkingSpinner_CycleAllFrames(t *testing.T) {
	b := NewThinkingBlock("th-sp-2")
	seen := map[string]bool{}
	for i := 0; i < len(thinkingSpinnerFrames)*2; i++ {
		frame := b.SpinnerFrame()
		seen[frame] = true
		b.AdvanceSpinner()
	}
	if len(seen) < 5 {
		t.Errorf("expected at least 5 distinct frames over cycle, got %d", len(seen))
	}
}

func TestP72_ThinkingSpinner_FrameInBounds(t *testing.T) {
	b := NewThinkingBlock("th-sp-3")
	// Advance many times
	for i := 0; i < 100; i++ {
		b.AdvanceSpinner()
	}
	frame := b.SpinnerFrame()
	found := false
	for _, f := range thinkingSpinnerFrames {
		if f == frame {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("spinner frame %q not in known frames", frame)
	}
}

func TestP72_Thinking_CharCount(t *testing.T) {
	b := NewThinkingBlock("th-cc-1")
	if b.CharCount() != 0 {
		t.Errorf("expected 0 chars for new block, got %d", b.CharCount())
	}
	b.AppendDelta("hello")
	if b.CharCount() != 5 {
		t.Errorf("expected 5 chars, got %d", b.CharCount())
	}
	b.AppendDelta(" world")
	if b.CharCount() != 11 {
		t.Errorf("expected 11 chars, got %d", b.CharCount())
	}
}

func TestP72_Thinking_PreviewLine_Empty(t *testing.T) {
	b := NewThinkingBlock("th-pl-1")
	if b.PreviewLine(40) != "" {
		t.Errorf("expected empty preview for no content")
	}
}

func TestP72_Thinking_PreviewLine_Short(t *testing.T) {
	b := NewThinkingBlock("th-pl-2")
	b.AppendDelta("I should analyze this problem step by step.")
	preview := b.PreviewLine(80)
	if preview == "" {
		t.Error("expected non-empty preview")
	}
}

func TestP72_Thinking_PreviewLine_Truncated(t *testing.T) {
	b := NewThinkingBlock("th-pl-3")
	longText := "This is a very long thinking content that should be truncated when previewing"
	b.AppendDelta(longText)
	preview := b.PreviewLine(20)
	r := []rune(preview)
	if len(r) > 20 {
		t.Errorf("expected preview <= 20 chars, got %d", len(r))
	}
	// Should end with ellipsis if truncated
	if len([]rune(longText)) > 20 {
		lastRune := r[len(r)-1]
		if lastRune != '…' {
			t.Errorf("expected ellipsis at end of truncated preview, got %c", lastRune)
		}
	}
}

func TestP72_Thinking_PreviewLine_Multiline(t *testing.T) {
	b := NewThinkingBlock("th-pl-4")
	b.AppendDelta("First line of thinking\nSecond line\nThird line")
	preview := b.PreviewLine(80)
	if preview != "First line of thinking" {
		t.Errorf("expected first line only, got %q", preview)
	}
}

func TestP72_Thinking_PreviewLine_LeadingWhitespace(t *testing.T) {
	b := NewThinkingBlock("th-pl-5")
	b.AppendDelta("  \n  \n  Actual content here")
	preview := b.PreviewLine(80)
	if preview != "Actual content here" {
		t.Errorf("expected trimmed content, got %q", preview)
	}
}

func TestP72_Thinking_Expand(t *testing.T) {
	b := NewThinkingBlock("th-ex-1")
	if !b.Collapsed() {
		t.Error("expected collapsed by default")
	}
	b.Expand()
	if b.Collapsed() {
		t.Error("expected expanded after Expand()")
	}
	// Expand when already expanded - should be no-op
	b.Expand()
	if b.Collapsed() {
		t.Error("expected still expanded")
	}
}

func TestP72_Thinking_Collapse(t *testing.T) {
	b := NewThinkingBlock("th-co-1")
	b.Expand()
	b.Collapse()
	if !b.Collapsed() {
		t.Error("expected collapsed after Collapse()")
	}
	// Collapse when already collapsed - should be no-op
	b.Collapse()
	if !b.Collapsed() {
		t.Error("expected still collapsed")
	}
}

func TestP72_Thinking_SetCollapsed(t *testing.T) {
	b := NewThinkingBlock("th-sc-1")
	b.SetCollapsed(false)
	if b.Collapsed() {
		t.Error("expected expanded")
	}
	b.SetCollapsed(true)
	if !b.Collapsed() {
		t.Error("expected collapsed")
	}
}

func TestP72_Thinking_PaintCollapsed_Streaming(t *testing.T) {
	b := NewThinkingBlock("th-pc-1")
	b.AppendDelta("analyzing input")
	b.SetState(BlockStreaming)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.Paint(buf)
	// Should have content in first row
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}

func TestP72_Thinking_PaintCollapsed_Complete(t *testing.T) {
	b := NewThinkingBlock("th-pc-2")
	b.AppendDelta("analyzing input")
	b.SetState(BlockComplete)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}

func TestP72_Thinking_PaintCollapsed_EmptyContent(t *testing.T) {
	b := NewThinkingBlock("th-pc-3")
	b.SetState(BlockStreaming)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.Paint(buf) // should not panic
}

func TestP72_Thinking_PaintCollapsed_NarrowWidth(t *testing.T) {
	b := NewThinkingBlock("th-pc-4")
	b.AppendDelta("very long thinking content")
	b.SetState(BlockComplete)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)
	// Content should be truncated
	_ = buf.GetCell(0, 0) // should not panic
}

func TestP72_Thinking_PaintExpanded_WithContent(t *testing.T) {
	b := NewThinkingBlock("th-pe-1")
	b.AppendDelta("## Analysis\n\nLet me think about this step by step.\n\n1. First consideration\n2. Second consideration")
	b.SetState(BlockComplete)
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 20})
	buf := buffer.NewBuffer(80, 20)
	b.Paint(buf)
	// Header should be at row 0
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected header content at (0,0)")
	}
	// Content should start at row 1+
	cell = buf.GetCell(2, 1)
	if cell.Rune == 0 {
		t.Error("expected content at row 1")
	}
}

func TestP72_Thinking_PaintExpanded_EmptyContent(t *testing.T) {
	b := NewThinkingBlock("th-pe-2")
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf) // should not panic
}

func TestP72_Thinking_PaintExpanded_Streaming(t *testing.T) {
	b := NewThinkingBlock("th-pe-3")
	b.AppendDelta("thinking...")
	b.SetState(BlockStreaming)
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	// Should have streaming content + cursor
}

func TestP72_Thinking_PaintExpanded_ZeroBounds(t *testing.T) {
	b := NewThinkingBlock("th-pe-4")
	b.AppendDelta("content")
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // should not panic with zero bounds
}

func TestP72_Thinking_SerializeState(t *testing.T) {
	b := NewThinkingBlock("th-ss-1")
	b.AppendDelta("thinking content")
	b.AdvanceSpinner()
	b.AdvanceSpinner()
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty serialized data")
	}
}

func TestP72_Thinking_DeserializeState_RoundTrip(t *testing.T) {
	b := NewThinkingBlock("th-ds-1")
	b.AppendDelta("original thinking")
	b.Expand()
	b.AdvanceSpinner()
	b.AdvanceSpinner()
	data, _ := b.SerializeState()

	b2 := NewThinkingBlock("th-ds-2")
	err := b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if b2.Content() != "original thinking" {
		t.Errorf("expected 'original thinking', got %q", b2.Content())
	}
	if b2.Collapsed() {
		t.Error("expected expanded (not collapsed)")
	}
}

func TestP72_Thinking_SetContentInvalidatesCache(t *testing.T) {
	b := NewThinkingBlock("th-ci-1")
	b.AppendDelta("initial content")
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // fill cache

	b.SetContent("replaced content")
	if b.Content() != "replaced content" {
		t.Errorf("expected 'replaced content', got %q", b.Content())
	}
}

func TestP72_Thinking_MeasureExpanded(t *testing.T) {
	b := NewThinkingBlock("th-me-1")
	b.AppendDelta("short content")
	b.Expand()
	s := b.Measure(component.Bounded(80, 100))
	if s.H <= 1 {
		t.Errorf("expected height > 1 for expanded, got %d", s.H)
	}
}

func TestP72_Thinking_MeasureCollapsed(t *testing.T) {
	b := NewThinkingBlock("th-me-2")
	b.AppendDelta("content")
	s := b.Measure(component.Bounded(80, 100))
	if s.H != 1 {
		t.Errorf("expected height 1 for collapsed, got %d", s.H)
	}
}

func TestP72_Thinking_MeasureExpanded_Empty(t *testing.T) {
	b := NewThinkingBlock("th-me-3")
	b.Expand()
	s := b.Measure(component.Bounded(80, 100))
	if s.H != 1 {
		t.Errorf("expected height 1 for empty expanded, got %d", s.H)
	}
}

func TestP72_Thinking_ConcurrentPaintAndAdvanceSpinner(t *testing.T) {
	b := NewThinkingBlock("th-conc-1")
	b.AppendDelta("thinking content here")
	b.SetState(BlockStreaming)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf1 := buffer.NewBuffer(80, 10)
	buf2 := buffer.NewBuffer(80, 10)

	done := make(chan bool, 2)
	// Concurrent paint
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 50; i++ {
			b.Paint(buf1)
		}
	}()
	// Concurrent spinner advance
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 50; i++ {
			b.AdvanceSpinner()
		}
	}()
	<-done
	<-done
	_ = buf2
}
