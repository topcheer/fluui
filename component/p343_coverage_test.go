package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// TestP343_Accordion_Expand_SingleOpen covers the singleOpen branch in Expand.
func TestP343_Accordion_Expand_SingleOpen(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a", Expanded: true},
		{Title: "B", Content: "b"},
		{Title: "C", Content: "c"},
	})
	a.SetSingleOpen(true)

	// Expand C — A should collapse (singleOpen behavior)
	a.Expand(2)

	if a.IsExpanded(0) {
		t.Error("A should collapse when C expands (singleOpen)")
	}
	if !a.IsExpanded(2) {
		t.Error("C should be expanded")
	}
	if a.IsExpanded(1) {
		t.Error("B should remain collapsed")
	}
}

// TestP343_Accordion_Expand_Invalid covers the out-of-range guard.
func TestP343_Accordion_Expand_Invalid(t *testing.T) {
	a := NewAccordion([]AccordionItem{{Title: "A", Content: "a"}})
	a.Expand(-1)
	a.Expand(99)
	if a.IsExpanded(0) {
		t.Error("invalid Expand should not change state")
	}
}

// TestP343_DrawStyledText_Truncation covers the truncation path.
func TestP343_DrawStyledText_Truncation(t *testing.T) {
	tc := NewToolCallView("test", "{}")
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)

	// Text longer than maxW should truncate
	tc.drawStyledText(buf, 0, 0, 5, "This is a very long text", buffer.Style{})

	// Text shorter than maxW should pass through
	tc.drawStyledText(buf, 0, 1, 60, "short", buffer.Style{})

	// Cell at position 0 should be non-empty
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell after truncation")
	}
}
