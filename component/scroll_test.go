package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// mockLeaf is a simple component with a fixed size for testing.
type mockLeaf struct {
	BaseComponent
	w, h int
}

func (m *mockLeaf) Measure(cs Constraints) Size { return Size{W: m.w, H: m.h} }
func (m *mockLeaf) Paint(buf *buffer.Buffer)    {}

func TestScrollViewMeasure(t *testing.T) {
	child := &mockLeaf{w: 20, h: 50}
	sv := NewScrollView(child)

	// No constraint → takes content size
	s := sv.Measure(Unbounded())
	if s.W != 20 || s.H != 50 {
		t.Errorf("unbounded: got %dx%d, want 20x50", s.W, s.H)
	}

	// Height constrained → clamped
	s = sv.Measure(Bounded(80, 10))
	if s.H != 10 {
		t.Errorf("bounded height: got %d, want 10", s.H)
	}
}

func TestScrollViewScroll(t *testing.T) {
	child := &mockLeaf{w: 20, h: 50}
	sv := NewScrollView(child)
	sv.Measure(Bounded(20, 10))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	// After Paint, maxOffset should be 50-10 = 40
	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)

	if sv.MaxOffset() != 40 {
		t.Errorf("maxOffset: got %d, want 40", sv.MaxOffset())
	}

	// Scroll down
	sv.ScrollDown(5)
	if sv.Offset() != 5 {
		t.Errorf("after scroll down 5: offset=%d, want 5", sv.Offset())
	}

	// Scroll past max
	sv.ScrollDown(100)
	if sv.Offset() != 40 {
		t.Errorf("after scroll down 100: offset=%d, want 40", sv.Offset())
	}

	// Scroll up
	sv.ScrollUp(10)
	if sv.Offset() != 30 {
		t.Errorf("after scroll up 10: offset=%d, want 30", sv.Offset())
	}

	// Scroll past top
	sv.ScrollUp(100)
	if sv.Offset() != 0 {
		t.Errorf("after scroll up 100: offset=%d, want 0", sv.Offset())
	}
}

func TestScrollViewScrollTo(t *testing.T) {
	child := &mockLeaf{w: 20, h: 50}
	sv := NewScrollView(child)
	sv.Measure(Bounded(20, 10))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)

	sv.ScrollTo(15)
	if sv.Offset() != 15 {
		t.Errorf("ScrollTo(15): got %d", sv.Offset())
	}

	sv.ScrollTo(-5)
	if sv.Offset() != 0 {
		t.Errorf("ScrollTo(-5): got %d, want 0", sv.Offset())
	}

	sv.ScrollTo(999)
	if sv.Offset() != 40 {
		t.Errorf("ScrollTo(999): got %d, want 40", sv.Offset())
	}
}

func TestScrollViewNoScrollNeeded(t *testing.T) {
	// Content smaller than viewport
	child := &mockLeaf{w: 20, h: 5}
	sv := NewScrollView(child)
	sv.Measure(Bounded(20, 10))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)

	if sv.MaxOffset() != 0 {
		t.Errorf("maxOffset should be 0 for small content, got %d", sv.MaxOffset())
	}
	if sv.Offset() != 0 {
		t.Errorf("offset should be 0, got %d", sv.Offset())
	}
}

func TestScrollViewScrollBar(t *testing.T) {
	child := &mockLeaf{w: 20, h: 100}
	sv := NewScrollView(child)
	sv.scrollBar = DefaultScrollBarStyle()
	sv.Measure(Bounded(20, 10))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)

	// Check scrollbar exists on the rightmost column
	cell := buf.GetCell(19, 0)
	if cell.Rune == 0 || cell.Rune == ' ' {
		t.Error("expected scrollbar character at (19,0)")
	}
}
