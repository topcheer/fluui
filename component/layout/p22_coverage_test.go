package layout

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P22-B: Layout coverage tests targeting uncovered code paths.
// Current coverage: 81.9% → Target: 95%+
//
// Uncovered paths identified:
// - center.go SetBounds: child larger than bounds (clamp), negative offset clamp
// - center.go Children: 0% (not called)
// - flex.go Measure: MaxHeight constraint on column mode
// - flex.go SetBounds: empty children early return
// - padding.go Measure: bounded constraints, negative inner clamp
// - padding.go SetBounds: negative innerW/innerH clamp
// - padding.go Children: 0% (not called)
// - stack.go Measure: MaxWidth/MaxHeight constraints
// - stack.go AddChild: 0% (not called)

// ── Center.Children ──

func TestCenterChildren(t *testing.T) {
	child := component.NewText("X")
	center := NewCenter(child)

	children := center.Children()
	if len(children) != 1 {
		t.Fatalf("Children() returned %d items, want 1", len(children))
	}
	if children[0] != child {
		t.Error("Children()[0] is not the wrapped child")
	}
}

// ── Center.SetBounds with child larger than bounds ──

func TestCenterSetBounds_ChildLargerThanBounds(t *testing.T) {
	// Child 20x5, bounds only 10x3 → should clamp to bounds
	child := newBox(20, 5)
	center := NewCenter(child)

	center.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 3})

	cb := child.Bounds()
	if cb.W != 10 {
		t.Errorf("child W = %d, want 10 (clamped to bounds)", cb.W)
	}
	if cb.H != 3 {
		t.Errorf("child H = %d, want 3 (clamped to bounds)", cb.H)
	}
	if cb.X != 0 {
		t.Errorf("child X = %d, want 0 (no offset when clamped)", cb.X)
	}
	if cb.Y != 0 {
		t.Errorf("child Y = %d, want 0 (no offset when clamped)", cb.Y)
	}
}

func TestCenterSetBounds_ChildExactFit(t *testing.T) {
	// Child exactly fits bounds → offset = 0
	child := newBox(5, 3)
	center := NewCenter(child)

	center.SetBounds(component.Rect{X: 2, Y: 4, W: 5, H: 3})

	cb := child.Bounds()
	if cb.X != 2 {
		t.Errorf("child X = %d, want 2", cb.X)
	}
	if cb.Y != 4 {
		t.Errorf("child Y = %d, want 4", cb.Y)
	}
}

func TestCenterSetBounds_NonZeroOrigin(t *testing.T) {
	// Bounds with non-zero X/Y → center offset should be relative
	child := newBox(2, 1)
	center := NewCenter(child)

	center.SetBounds(component.Rect{X: 10, Y: 20, W: 10, H: 5})

	cb := child.Bounds()
	// offsetX = (10-2)/2 = 4, so X = 10+4 = 14
	if cb.X != 14 {
		t.Errorf("child X = %d, want 14", cb.X)
	}
	// offsetY = (5-1)/2 = 2, so Y = 20+2 = 22
	if cb.Y != 22 {
		t.Errorf("child Y = %d, want 22", cb.Y)
	}
}

// ── Center.SetBounds via Measure path ──

func TestCenterSetBounds_MeasuresChildWithConstraints(t *testing.T) {
	// Use NewText (respects constraints) to exercise Measure within SetBounds
	child := component.NewText("Hi") // 2x1
	center := NewCenter(child)

	// Bounds much larger than child
	center.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	cb := child.Bounds()
	expectedX := (80 - 2) / 2 // 39
	expectedY := (24 - 1) / 2 // 11
	if cb.X != expectedX {
		t.Errorf("child X = %d, want %d", cb.X, expectedX)
	}
	if cb.Y != expectedY {
		t.Errorf("child Y = %d, want %d", cb.Y, expectedY)
	}
}

// ── Flex.Measure with MaxHeight constraint ──

func TestFlexColumnMeasure_MaxHeightConstraint(t *testing.T) {
	flex := NewFlexGap(FlexColumn, 1)
	flex.AddChild(newBox(10, 10))
	flex.AddChild(newBox(10, 10))

	// Natural height = 10 + 1 + 10 = 21, constrained to 15
	size := flex.Measure(component.Bounded(20, 15))
	if size.H != 15 {
		t.Errorf("height = %d, want 15 (MaxHeight cap)", size.H)
	}
}

func TestFlexRowMeasure_MaxHeightConstraint(t *testing.T) {
	flex := NewFlexGap(FlexRow, 1)
	flex.AddChild(newBox(10, 10))

	// Height = 10, but MaxHeight = 5
	size := flex.Measure(component.Bounded(50, 5))
	if size.H != 5 {
		t.Errorf("height = %d, want 5 (MaxHeight cap)", size.H)
	}
}

// ── Flex.SetBounds with empty children ──

func TestFlexSetBounds_Empty(t *testing.T) {
	flex := NewFlex(FlexRow)
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	// Should not panic, should return early
}

func TestFlexColumnSetBounds_Empty(t *testing.T) {
	flex := NewFlex(FlexColumn)
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
}

// ── Flex.SetBounds on column with offset ──

func TestFlexColumnSetBounds_WithOffset(t *testing.T) {
	flex := NewFlexGap(FlexColumn, 2)

	c1 := newBox(10, 3)
	c2 := newBox(5, 7)
	flex.AddChild(c1)
	flex.AddChild(c2)

	flex.SetBounds(component.Rect{X: 3, Y: 1, W: 20, H: 20})

	// c1: X=3, Y=1, W=10, H=3
	checkBounds(t, "c1", c1.bounds, component.Rect{X: 3, Y: 1, W: 10, H: 3})

	// c2: X=3, Y=1+3+2=6, W=5, H=7
	checkBounds(t, "c2", c2.bounds, component.Rect{X: 3, Y: 6, W: 5, H: 7})
}

// ── Flex.Paint with column ──

func TestFlexColumnPaint(t *testing.T) {
	flex := NewFlex(FlexColumn)
	flex.AddChild(newBox(5, 3))
	flex.AddChild(newBox(5, 3))
	flex.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 6})
	buf := buffer.NewBuffer(5, 6)
	flex.Paint(buf) // should not panic
}

// ── Flex single child column SetBounds ──

func TestFlexSingleChild_ColumnSetBounds(t *testing.T) {
	flex := NewFlex(FlexColumn)
	child := newBox(5, 3)
	flex.AddChild(child)

	flex.SetBounds(component.Rect{X: 2, Y: 3, W: 10, H: 20})

	cb := child.Bounds()
	if cb.X != 2 || cb.Y != 3 || cb.W != 5 || cb.H != 3 {
		t.Errorf("single child bounds = %+v, want {2,3,5,3}", cb)
	}
}

// ── Padding.Measure with bounded constraints ──

func TestPaddingMeasure_Bounded(t *testing.T) {
	child := component.NewText("Hello") // 5x1
	// padding: top=1, right=2, bottom=1, left=2
	padding := NewPadding(1, 2, 1, 2, child)

	// With bounded constraints (innerW = 20-4=16, innerH=10-2=8)
	sz := padding.Measure(component.Bounded(20, 10))
	// width = 5 + 2 + 2 = 9, height = 1 + 1 + 1 = 3
	if sz.W != 9 {
		t.Errorf("width = %d, want 9", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("height = %d, want 3", sz.H)
	}
}

func TestPaddingMeasure_BoundedExceedsPadding(t *testing.T) {
	child := newBox(10, 10) // Fixed 10x10
	// padding: top=5, right=5, bottom=5, left=5 → total padding = 10 each axis
	padding := NewPadding(5, 5, 5, 5, child)

	// Bounded to 8x8: innerW = 8-10 = -2 → clamped to 0
	// innerH = 8-10 = -2 → clamped to 0
	sz := padding.Measure(component.Bounded(8, 8))
	// child measured with 0x0 constraints, still returns 10x10
	// total = 10 + 10 = 20 each
	if sz.W != 20 {
		t.Errorf("width = %d, want 20 (child 10 + left 5 + right 5)", sz.W)
	}
	if sz.H != 20 {
		t.Errorf("height = %d, want 20", sz.H)
	}
}

func TestPaddingMeasure_ZeroConstraints(t *testing.T) {
	child := component.NewText("Hi") // 2x1
	padding := NewPadding(1, 1, 1, 1, child)

	// Unbounded: innerW/innerH stay 0, so Bounded(0, 0)
	sz := padding.Measure(component.Unbounded())
	// width = 2 + 1 + 1 = 4, height = 1 + 1 + 1 = 3
	if sz.W != 4 {
		t.Errorf("width = %d, want 4", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("height = %d, want 3", sz.H)
	}
}

// ── Padding.SetBounds with negative inner dimensions ──

func TestPaddingSetBounds_NegativeInner(t *testing.T) {
	child := newBox(5, 5)
	// padding: top=10, right=10, bottom=10, left=10
	padding := NewPadding(10, 10, 10, 10, child)

	// Bounds 5x5, but padding needs 20x20 → innerW = 5-20 = -15 → 0
	padding.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 5})

	cb := child.Bounds()
	if cb.W != 0 {
		t.Errorf("child W = %d, want 0 (negative clamped)", cb.W)
	}
	if cb.H != 0 {
		t.Errorf("child H = %d, want 0 (negative clamped)", cb.H)
	}
}

// ── Padding.Children ──

func TestPaddingChildren(t *testing.T) {
	child := component.NewText("X")
	padding := NewPadding(1, 1, 1, 1, child)

	children := padding.Children()
	if len(children) != 1 {
		t.Fatalf("Children() returned %d, want 1", len(children))
	}
	if children[0] != child {
		t.Error("Children()[0] is not the wrapped child")
	}
}

// ── Stack.Measure with constraints ──

func TestStackMeasure_MaxWidthConstraint(t *testing.T) {
	stack := NewStack(newBox(10, 5), newBox(20, 3))

	// Natural max = 20x5, constrained to 15 wide
	sz := stack.Measure(component.Bounded(15, 100))
	if sz.W != 15 {
		t.Errorf("width = %d, want 15 (MaxWidth cap)", sz.W)
	}
}

func TestStackMeasure_MaxHeightConstraint(t *testing.T) {
	stack := NewStack(newBox(10, 5), newBox(20, 15))

	// Natural max = 20x15, constrained to 10 high
	sz := stack.Measure(component.Bounded(100, 10))
	if sz.H != 10 {
		t.Errorf("height = %d, want 10 (MaxHeight cap)", sz.H)
	}
}

func TestStackMeasure_Empty(t *testing.T) {
	stack := NewStack()
	sz := stack.Measure(component.Unbounded())
	if sz.W != 0 || sz.H != 0 {
		t.Errorf("empty stack measure = %v, want {0,0}", sz)
	}
}

// ── Stack.AddChild ──

func TestStackAddChild(t *testing.T) {
	child1 := component.NewText("A")
	stack := NewStack(child1)

	// Initially 1 child
	if len(stack.Children()) != 1 {
		t.Fatalf("initial children = %d, want 1", len(stack.Children()))
	}

	child2 := component.NewText("B")
	stack.AddChild(child2)

	if len(stack.Children()) != 2 {
		t.Errorf("after AddChild: children = %d, want 2", len(stack.Children()))
	}
	if stack.Children()[1] != child2 {
		t.Error("Children()[1] is not the added child")
	}
}

func TestStackAddChild_Multiple(t *testing.T) {
	stack := NewStack()
	stack.AddChild(component.NewText("A"))
	stack.AddChild(component.NewText("B"))
	stack.AddChild(component.NewText("C"))

	if len(stack.Children()) != 3 {
		t.Errorf("children = %d, want 3", len(stack.Children()))
	}
}

// ── Stack.Paint with empty ──

func TestStackPaint_Empty(t *testing.T) {
	stack := NewStack()
	stack.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	stack.Paint(buf) // should not panic
}

// ── Stack.SetBounds propagates to all children ──

func TestStackSetBounds_AllChildrenSameBounds(t *testing.T) {
	c1 := newBox(3, 3)
	c2 := newBox(5, 5)
	c3 := newBox(2, 2)
	stack := NewStack(c1, c2, c3)

	r := component.Rect{X: 1, Y: 2, W: 20, H: 10}
	stack.SetBounds(r)

	for i, child := range []component.Component{c1, c2, c3} {
		cb := child.Bounds()
		if cb != r {
			t.Errorf("child[%d] bounds = %+v, want %+v", i, cb, r)
		}
	}
}
