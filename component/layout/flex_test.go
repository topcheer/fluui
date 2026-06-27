package layout

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// box is a simple leaf component with a fixed size for testing.
type box struct {
	component.BaseComponent
	w, h   int
	bounds component.Rect
}

func newBox(w, h int) *box {
	return &box{w: w, h: h}
}

func (b *box) Measure(_ component.Constraints) component.Size {
	return component.Size{W: b.w, H: b.h}
}

func (b *box) SetBounds(r component.Rect) { b.bounds = r }
func (b *box) Bounds() component.Rect      { return b.bounds }
func (b *box) Paint(buf *buffer.Buffer)    {}

func TestFlexRowMeasure(t *testing.T) {
	flex := NewFlexGap(FlexRow, 1)
	flex.AddChild(newBox(10, 5))
	flex.AddChild(newBox(20, 3))
	flex.AddChild(newBox(15, 4))

	size := flex.Measure(component.Unbounded())

	// Width = 10 + 1(gap) + 20 + 1(gap) + 15 = 47
	wantW := 10 + 1 + 20 + 1 + 15
	if size.W != wantW {
		t.Errorf("width = %d, want %d", size.W, wantW)
	}

	// Height = max(5, 3, 4) = 5
	if size.H != 5 {
		t.Errorf("height = %d, want 5", size.H)
	}
}

func TestFlexColumnMeasure(t *testing.T) {
	flex := NewFlexGap(FlexColumn, 1)
	flex.AddChild(newBox(10, 5))
	flex.AddChild(newBox(20, 3))
	flex.AddChild(newBox(15, 4))

	size := flex.Measure(component.Unbounded())

	// Height = 5 + 1 + 3 + 1 + 4 = 14
	wantH := 5 + 1 + 3 + 1 + 4
	if size.H != wantH {
		t.Errorf("height = %d, want %d", size.H, wantH)
	}

	// Width = max(10, 20, 15) = 20
	if size.W != 20 {
		t.Errorf("width = %d, want 20", size.W)
	}
}

func TestFlexMeasureZeroGap(t *testing.T) {
	flex := NewFlex(FlexRow) // gap = 0
	flex.AddChild(newBox(10, 5))
	flex.AddChild(newBox(20, 5))

	size := flex.Measure(component.Unbounded())
	if size.W != 30 {
		t.Errorf("width = %d, want 30", size.W)
	}
}

func TestFlexMeasureEmpty(t *testing.T) {
	flex := NewFlex(FlexRow)
	size := flex.Measure(component.Unbounded())
	if size.W != 0 || size.H != 0 {
		t.Errorf("empty flex measure = %v, want {0,0}", size)
	}
}

func TestFlexRowSetBounds(t *testing.T) {
	flex := NewFlexGap(FlexRow, 1)

	c1 := newBox(10, 5)
	c2 := newBox(20, 3)
	c3 := newBox(15, 4)
	flex.AddChild(c1)
	flex.AddChild(c2)
	flex.AddChild(c3)

	flex.SetBounds(component.Rect{X: 5, Y: 2, W: 47, H: 5})

	// c1: X=5, Y=2, W=10, H=5
	checkBounds(t, "c1", c1.bounds, component.Rect{X: 5, Y: 2, W: 10, H: 5})

	// c2: X=5+10+1=16, Y=2, W=20, H=3
	checkBounds(t, "c2", c2.bounds, component.Rect{X: 16, Y: 2, W: 20, H: 3})

	// c3: X=16+20+1=37, Y=2, W=15, H=4
	checkBounds(t, "c3", c3.bounds, component.Rect{X: 37, Y: 2, W: 15, H: 4})
}

func TestFlexColumnSetBounds(t *testing.T) {
	flex := NewFlexGap(FlexColumn, 1)

	c1 := newBox(10, 5)
	c2 := newBox(20, 3)
	flex.AddChild(c1)
	flex.AddChild(c2)

	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 9})

	// c1: X=0, Y=0, W=10, H=5
	checkBounds(t, "c1", c1.bounds, component.Rect{X: 0, Y: 0, W: 10, H: 5})

	// c2: X=0, Y=5+1=6, W=20, H=3
	checkBounds(t, "c2", c2.bounds, component.Rect{X: 0, Y: 6, W: 20, H: 3})
}

func TestFlexChildren(t *testing.T) {
	flex := NewFlex(FlexRow)
	c1 := newBox(1, 1)
	c2 := newBox(2, 2)
	flex.AddChild(c1)
	flex.AddChild(c2)

	children := flex.Children()
	if len(children) != 2 {
		t.Fatalf("Children() returned %d, want 2", len(children))
	}
	if children[0] != c1 || children[1] != c2 {
		t.Error("Children() returned wrong components")
	}
}

func TestFlexPaint(t *testing.T) {
	// Paint should call Paint on all children without panic.
	flex := NewFlex(FlexRow)
	flex.AddChild(newBox(10, 5))
	flex.AddChild(newBox(20, 3))
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 5})

	buf := buffer.NewBuffer(30, 5)
	flex.Paint(buf)
	// If no panic, test passes.
}

func TestFlexMeasureWithMaxConstraints(t *testing.T) {
	flex := NewFlexGap(FlexRow, 1)
	flex.AddChild(newBox(10, 5))
	flex.AddChild(newBox(20, 5))

	// Natural width = 31, but constrained to 25
	size := flex.Measure(component.Bounded(25, 10))
	if size.W != 25 {
		t.Errorf("width = %d, want 25 (capped)", size.W)
	}
	if size.H != 5 {
		t.Errorf("height = %d, want 5", size.H)
	}
}

// --- helpers ---

func checkBounds(t *testing.T, name string, got, want component.Rect) {
	t.Helper()
	if got != want {
		t.Errorf("%s bounds = %+v, want %+v", name, got, want)
	}
}
