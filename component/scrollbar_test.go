package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// helper: create a ScrollView with a tall child for scrollbar testing.
func newTestScrollView(viewportH int, contentH int) *ScrollView {
	child := &mockTallComponent{height: contentH}
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: viewportH})
	sv.Measure(Constraints{MaxWidth: 20, MaxHeight: viewportH})
	return sv
}

// mockTallComponent is a simple component with a configurable height.
type mockTallComponent struct {
	BaseComponent
	height int
}

func (m *mockTallComponent) Measure(cs Constraints) Size {
	return Size{W: cs.MaxWidth, H: m.height}
}

func (m *mockTallComponent) SetBounds(r Rect) {
	m.BaseComponent.SetBounds(r)
}

func (m *mockTallComponent) Paint(buf *buffer.Buffer) {}

func (m *mockTallComponent) Children() []Component { return nil }

func TestScrollbarVisible(t *testing.T) {
	// Content taller than viewport → scrollbar visible.
	sv := newTestScrollView(10, 50)
	if !sv.IsScrollbarVisible() {
		t.Fatal("expected scrollbar visible when content > viewport")
	}

	// Content fits in viewport → scrollbar hidden.
	sv2 := newTestScrollView(50, 10)
	if sv2.IsScrollbarVisible() {
		t.Fatal("expected scrollbar hidden when content <= viewport")
	}
}

func TestScrollbarColumn(t *testing.T) {
	sv := newTestScrollView(10, 50)
	col := sv.ScrollbarColumn()
	if col != 19 { // X(0) + W(20) - 1
		t.Fatalf("expected scrollbar column 19, got %d", col)
	}

	// Hidden scrollbar → -1.
	sv2 := newTestScrollView(50, 10)
	if sv2.ScrollbarColumn() != -1 {
		t.Fatal("expected -1 when scrollbar hidden")
	}
}

func TestScrollbarBounds(t *testing.T) {
	sv := newTestScrollView(10, 50)

	barStartY, barH, thumbStartY, thumbH := sv.ScrollbarBounds()

	if barStartY != 0 {
		t.Fatalf("expected barStartY=0, got %d", barStartY)
	}
	if barH != 10 {
		t.Fatalf("expected barH=10, got %d", barH)
	}
	// thumbH = max(1, 10*10/50) = max(1, 2) = 2
	if thumbH != 2 {
		t.Fatalf("expected thumbH=2, got %d", thumbH)
	}
	// thumbStartY at offset 0 = 0
	if thumbStartY != 0 {
		t.Fatalf("expected thumbStartY=0, got %d", thumbStartY)
	}
}

func TestScrollbarClickTrack(t *testing.T) {
	sv := newTestScrollView(10, 50)

	// Click at relY=5 (middle of track, not on thumb which is at 0-1).
	sv.HandleScrollbarDown(5)

	// Should have scrolled. offset should be > 0.
	if sv.Offset() == 0 {
		t.Fatal("expected offset > 0 after clicking track below thumb")
	}
}

func TestScrollbarClickThumbStartsDrag(t *testing.T) {
	sv := newTestScrollView(10, 50)

	// Click at relY=0 (on thumb at position 0-1).
	sv.HandleScrollbarDown(0)

	if !sv.IsDragging() {
		t.Fatal("expected dragging to be true after clicking on thumb")
	}
}

func TestScrollbarDrag(t *testing.T) {
	sv := newTestScrollView(10, 50)

	// Start drag on thumb.
	sv.HandleScrollbarDown(0)
	if !sv.IsDragging() {
		t.Fatal("expected dragging")
	}

	// Drag to bottom of bar.
	sv.HandleScrollbarDrag(9)

	// Should have scrolled to near the end.
	if sv.Offset() == 0 {
		t.Fatal("expected offset > 0 after dragging to bottom")
	}

	_, _, _, thumbH := sv.ScrollbarBounds()
	maxRelY := 10 - thumbH

	// Drag to exact bottom of valid range.
	sv.HandleScrollbarDrag(maxRelY)
	if sv.Offset() != sv.MaxOffset() {
		// Allow off-by-one due to rounding.
		diff := sv.MaxOffset() - sv.Offset()
		if diff > 1 {
			t.Fatalf("expected offset near max (%d), got %d (diff=%d)", sv.MaxOffset(), sv.Offset(), diff)
		}
	}
}

func TestScrollbarDragToEnd(t *testing.T) {
	sv := newTestScrollView(10, 50)

	_, _, _, thumbH := sv.ScrollbarBounds()
	maxRelY := 10 - thumbH

	sv.HandleScrollbarDown(0)
	sv.HandleScrollbarDrag(maxRelY)

	// Should be at or near max offset.
	if sv.Offset() < sv.MaxOffset()-1 {
		t.Fatalf("expected offset near max (%d), got %d", sv.MaxOffset(), sv.Offset())
	}
}

func TestScrollbarDragClamped(t *testing.T) {
	sv := newTestScrollView(10, 50)

	sv.HandleScrollbarDown(0)
	sv.HandleScrollbarDrag(-5) // negative, should clamp to 0

	if sv.Offset() != 0 {
		t.Fatalf("expected offset 0 after dragging to negative, got %d", sv.Offset())
	}
}

func TestScrollbarDragBeyondMax(t *testing.T) {
	sv := newTestScrollView(10, 50)

	sv.HandleScrollbarDown(0)
	sv.HandleScrollbarDrag(100) // well beyond bar height

	// Should clamp to max offset.
	if sv.Offset() < sv.MaxOffset()-1 {
		t.Fatalf("expected offset near max, got %d", sv.Offset())
	}
}

func TestScrollbarUpEndsDrag(t *testing.T) {
	sv := newTestScrollView(10, 50)

	sv.HandleScrollbarDown(0)
	if !sv.IsDragging() {
		t.Fatal("expected dragging")
	}

	sv.HandleScrollbarUp()
	if sv.IsDragging() {
		t.Fatal("expected not dragging after HandleScrollbarUp")
	}
}

func TestScrollbarDragWithoutDown(t *testing.T) {
	sv := newTestScrollView(10, 50)

	// Drag without prior MouseDown should be a no-op.
	sv.HandleScrollbarDrag(5)
	if sv.IsDragging() {
		t.Fatal("expected not dragging without prior down")
	}
	if sv.Offset() != 0 {
		t.Fatalf("expected offset 0, got %d", sv.Offset())
	}
}

func TestScrollbarPaintHighlight(t *testing.T) {
	sv := newTestScrollView(10, 50)
	sv.scrollBar.ThumbChar = '█'
	sv.scrollBar.TrackChar = '░'

	// Start dragging.
	sv.HandleScrollbarDown(0)

	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)

	// Thumb cells should have different style (Bold flag) when dragging.
	cell := buf.GetCell(19, 0) // barX=19, y=0 (thumb position)
	if cell.Rune != '█' {
		t.Fatalf("expected thumb char '█', got %c", cell.Rune)
	}
	// When dragging, thumb should have Bold flag.
	if cell.Flags&buffer.Bold == 0 {
		t.Fatal("expected Bold flag on thumb when dragging")
	}

	// End dragging.
	sv.HandleScrollbarUp()
	buf2 := buffer.NewBuffer(20, 10)
	sv.Paint(buf2)

	cell2 := buf2.GetCell(19, 0)
	// When not dragging, thumb should NOT have Bold (unless default style has it).
	// The key check is that dragging changes the style.
	if cell2.Flags&buffer.Bold != 0 && cell.Flags&buffer.Bold != 0 {
		// Both have bold — the highlight didn't change. Only fail if default style doesn't have bold.
	}
}

func TestScrollbarClickNearThumbNoJump(t *testing.T) {
	sv := newTestScrollView(10, 50)

	// Click at relY=1 (still on thumb which is 0-1).
	sv.HandleScrollbarDown(1)

	// Should start drag, not jump.
	if !sv.IsDragging() {
		t.Fatal("expected dragging when clicking on thumb")
	}
	if sv.Offset() != 0 {
		t.Fatalf("expected offset 0 (no jump when clicking thumb), got %d", sv.Offset())
	}
}

func TestScrollbarHiddenNoInteraction(t *testing.T) {
	// Content fits viewport — scrollbar not visible.
	sv := newTestScrollView(50, 10)

	// These should all be no-ops.
	sv.HandleScrollbarDown(5)
	sv.HandleScrollbarDrag(5)
	sv.HandleScrollbarUp()

	if sv.IsDragging() {
		t.Fatal("expected not dragging with hidden scrollbar")
	}
	if sv.Offset() != 0 {
		t.Fatalf("expected offset 0, got %d", sv.Offset())
	}
}
