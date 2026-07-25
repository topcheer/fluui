package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP356_Popover_Create(t *testing.T) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1}, "Tip", "Use /help")
	if !p.IsOpen() {
		t.Error("should start open")
	}
}

func TestP356_Popover_OpenClose(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "B")
	p.Close()
	if p.IsOpen() {
		t.Error("should be closed")
	}
	p.Open()
	if !p.IsOpen() {
		t.Error("should be open")
	}
}

func TestP356_Popover_Toggle(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "B")
	p.Toggle()
	if p.IsOpen() {
		t.Error("should be closed after toggle")
	}
	p.Toggle()
	if !p.IsOpen() {
		t.Error("should be open after second toggle")
	}
}

func TestP356_Popover_SetAnchor(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "B")
	p.SetAnchor(Rect{X: 20, Y: 10, W: 5, H: 2})
}

func TestP356_Popover_SetTitle(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "Old", "B")
	p.SetTitle("New")
}

func TestP356_Popover_SetBody(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "Old")
	p.SetBody("New body")
}

func TestP356_Popover_SetWidth(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "B")
	p.SetWidth(40)
	p.SetWidth(1) // should clamp
}

func TestP356_Popover_Measure_Closed(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "B")
	p.Close()
	s := p.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.W != 0 || s.H != 0 {
		t.Errorf("closed = %dx%d, want 0x0", s.W, s.H)
	}
}

func TestP356_Popover_Measure_Open(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "Title", "Line 1\nLine 2")
	s := p.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.H < 4 {
		t.Errorf("height = %d, expected at least 4", s.H)
	}
}

func TestP356_Popover_Paint(t *testing.T) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1}, "Quick Tip", "Press Tab to switch focus")
	p.SetWidth(30)
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)

	// Should render border below anchor
	cell := buf.GetCell(0, 6)
	if cell.Rune == 0 {
		t.Error("expected content below anchor")
	}
}

func TestP356_Popover_Paint_Closed(t *testing.T) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1}, "T", "B")
	p.Close()
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)
}

func TestP356_Popover_Paint_NoTitle(t *testing.T) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1}, "", "Just body text")
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)
}

func TestP356_Popover_Paint_LongBody(t *testing.T) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1}, "Tip",
		"This is a very long body line that exceeds the popover width\nAnd another line")
	p.SetWidth(15)
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)
}

func TestP356_Popover_Paint_AnchorBottom(t *testing.T) {
	// Anchor near bottom — popover should flip above
	p := NewPopover(Rect{X: 0, Y: 20, W: 4, H: 4}, "T", "B")
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)
}

func TestP356_Popover_Paint_LongTitle(t *testing.T) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1},
		"This is a very long title that needs truncation", "Body")
	p.SetWidth(15)
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)
}

func TestP356_Popover_Paint_ZeroBounds(t *testing.T) {
	p := NewPopover(Rect{X: 0, Y: 0, W: 4, H: 1}, "T", "B")
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 24)
	p.Paint(buf)
}

func BenchmarkPopover_Paint(b *testing.B) {
	p := NewPopover(Rect{X: 10, Y: 5, W: 4, H: 1}, "Quick Tip",
		"Press Tab to switch focus\nPress Enter to submit\nPress / for commands")
	p.SetWidth(30)
	p.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p.Paint(buf)
	}
}
