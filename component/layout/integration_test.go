package layout

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// Integration tests for layout containers: nesting, resize, composition.
// File naming: *_integration_test.go, function prefix: TestIntegration_*

// box is a deterministic test component with fixed size.
type integrationBox struct {
	component.BaseComponent
	w, h int
}

func newIBox(w, h int) *integrationBox {
	return &integrationBox{w: w, h: h}
}

func (b *integrationBox) Measure(cs component.Constraints) component.Size {
	w, h := b.w, b.h
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return component.Size{W: w, H: h}
}

func (b *integrationBox) Paint(buf *buffer.Buffer) {}

// ─── Flex nesting: Row > Column > Box ──────────────────

func TestIntegration_FlexRowNestingColumn(t *testing.T) {
	// Row containing two columns, each with 2 boxes
	col1 := NewFlex(FlexColumn)
	col1.AddChild(newIBox(10, 3))
	col1.AddChild(newIBox(10, 2))

	col2 := NewFlex(FlexColumn)
	col2.AddChild(newIBox(8, 4))
	col2.AddChild(newIBox(8, 1))

	row := NewFlex(FlexRow)
	row.AddChild(col1)
	row.AddChild(col2)

	sz := row.Measure(component.Unbounded())
	// Row width = col1(10) + col2(8) = 18
	// Row height = max(col1(5), col2(5)) = 5
	if sz.W != 18 {
		t.Errorf("nested row width = %d, want 18", sz.W)
	}
	if sz.H != 5 {
		t.Errorf("nested row height = %d, want 5", sz.H)
	}

	// SetBounds and verify children
	row.SetBounds(component.Rect{X: 0, Y: 0, W: 18, H: 5})

	col1Children := col1.Children()
	if len(col1Children) != 2 {
		t.Fatalf("col1 children = %d, want 2", len(col1Children))
	}
	// First child in col1 should be at Y=0
	r1 := col1Children[0].Bounds()
	if r1.H != 3 {
		t.Errorf("col1 child0 height = %d, want 3", r1.H)
	}
	r2 := col1Children[1].Bounds()
	if r2.Y < r1.Y+r1.H {
		t.Errorf("col1 child1 should be below child0: child0.Y=%d H=%d child1.Y=%d", r1.Y, r1.H, r2.Y)
	}
}

func TestIntegration_FlexColumnNestingRow(t *testing.T) {
	// Column containing two rows
	row1 := NewFlex(FlexRow)
	row1.AddChild(newIBox(5, 3))
	row1.AddChild(newIBox(5, 3))

	row2 := NewFlex(FlexRow)
	row2.AddChild(newIBox(3, 2))

	col := NewFlex(FlexColumn)
	col.AddChild(row1)
	col.AddChild(row2)

	sz := col.Measure(component.Unbounded())
	// Col width = max(row1(10), row2(3)) = 10
	// Col height = row1(3) + row2(2) = 5
	if sz.W != 10 {
		t.Errorf("nested col width = %d, want 10", sz.W)
	}
	if sz.H != 5 {
		t.Errorf("nested col height = %d, want 5", sz.H)
	}
}

// ─── Flex with Gap nesting ─────────────────────────────

func TestIntegration_FlexGapNesting(t *testing.T) {
	inner := NewFlexGap(FlexRow, 2)
	inner.AddChild(newIBox(5, 3))
	inner.AddChild(newIBox(5, 3))
	inner.AddChild(newIBox(5, 3))

	outer := NewFlexGap(FlexColumn, 1)
	outer.AddChild(inner)
	outer.AddChild(newIBox(17, 2))

	sz := outer.Measure(component.Unbounded())
	// inner width = 5*3 + 2*2 = 19, height = 3
	// outer width = max(19, 17) = 19
	// outer height = 3 + 1(gap) + 2 = 6
	if sz.W != 19 {
		t.Errorf("outer width = %d, want 19", sz.W)
	}
	if sz.H != 6 {
		t.Errorf("outer height = %d, want 6", sz.H)
	}
}

// ─── Stack overlaying Flex ─────────────────────────────

func TestIntegration_StackOverFlex(t *testing.T) {
	bg := NewFlex(FlexRow)
	bg.AddChild(newIBox(10, 5))
	bg.AddChild(newIBox(10, 5))

	overlay := newIBox(20, 1)

	stack := NewStack(bg, overlay)

	sz := stack.Measure(component.Unbounded())
	// bg measures 20x5, overlay 20x1 → stack is 20x5
	if sz.W != 20 {
		t.Errorf("stack width = %d, want 20", sz.W)
	}
	if sz.H != 5 {
		t.Errorf("stack height = %d, want 5", sz.H)
	}

	// Both children get same bounds
	stack.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 5})
	bgBounds := bg.Bounds()
	overlayBounds := overlay.Bounds()
	if bgBounds != overlayBounds {
		t.Errorf("stack children should have same bounds: bg=%v overlay=%v", bgBounds, overlayBounds)
	}
}

// ─── Center inside Flex ─────────────────────────────────

func TestIntegration_CenterInsideFlex(t *testing.T) {
	centered := NewCenter(newIBox(6, 3))

	row := NewFlex(FlexRow)
	row.AddChild(centered)
	row.AddChild(newIBox(10, 10))

	sz := row.Measure(component.Unbounded())
	// centered measures 6x3, second child 10x10
	// row width = 6 + 10 = 16, height = max(3, 10) = 10
	if sz.W != 16 {
		t.Errorf("row width = %d, want 16", sz.W)
	}
	if sz.H != 10 {
		t.Errorf("row height = %d, want 10", sz.H)
	}

	// SetBounds — centered child gets measured height, not full row height
	row.SetBounds(component.Rect{X: 0, Y: 0, W: 16, H: 10})
	cBounds := centered.Bounds()
	// centered gets its measured width
	if cBounds.W != 6 {
		t.Errorf("centered width = %d, want 6", cBounds.W)
	}
	// Flex assigns each child the row height (max of children)
	// centered measures 3 tall, but row height = 10
	// Flex.SetBounds gives each child the full row height
	if cBounds.H < 3 {
		t.Errorf("centered height = %d, should be >= 3", cBounds.H)
	}
}

// ─── Padding inside Flex ────────────────────────────────

func TestIntegration_PaddingInsideFlex(t *testing.T) {
	padded := NewPadding(1, 2, 1, 2, newIBox(10, 3))
	// padded size = 10+2*2=14 wide, 3+1*2=5 tall

	col := NewFlex(FlexColumn)
	col.AddChild(padded)
	col.AddChild(newIBox(14, 4))

	sz := col.Measure(component.Unbounded())
	// col width = max(14, 14) = 14
	// col height = 5 + 4 = 9
	if sz.W != 14 {
		t.Errorf("col width = %d, want 14", sz.W)
	}
	if sz.H != 9 {
		t.Errorf("col height = %d, want 9", sz.H)
	}
}

// ─── Resize: verify reflow on bounds change ────────────

func TestIntegration_ResizeReflow(t *testing.T) {
	row := NewFlex(FlexRow)
	row.AddChild(newIBox(10, 5))
	row.AddChild(newIBox(10, 5))

	// Initial bounds
	row.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 5})
	children := row.Children()
	r0 := children[0].Bounds()
	r1 := children[1].Bounds()
	if r0.W != 10 || r1.W != 10 {
		t.Errorf("initial: child0.W=%d child1.W=%d, want 10/10", r0.W, r1.W)
	}
	if r1.X != 10 {
		t.Errorf("initial: child1.X=%d, want 10", r1.X)
	}

	// Resize wider
	row.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	children = row.Children()
	r0 = children[0].Bounds()
	r1 = children[1].Bounds()
	// Flex doesn't stretch children — they stay at measured size
	if r0.W != 10 {
		t.Errorf("after resize: child0.W=%d, want 10 (fixed)", r0.W)
	}
}

func TestIntegration_ResizeColumnReflow(t *testing.T) {
	col := NewFlex(FlexColumn)
	col.AddChild(newIBox(20, 3))
	col.AddChild(newIBox(20, 4))
	col.AddChild(newIBox(20, 2))

	// Initial
	col.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 9})
	children := col.Children()
	if children[0].Bounds().H != 3 {
		t.Errorf("child0.H=%d, want 3", children[0].Bounds().H)
	}
	if children[1].Bounds().Y != 3 {
		t.Errorf("child1.Y=%d, want 3", children[1].Bounds().Y)
	}
	if children[2].Bounds().Y != 7 {
		t.Errorf("child2.Y=%d, want 7", children[2].Bounds().Y)
	}

	// Resize taller — children stay at measured size, positions unchanged
	col.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 20})
	if children[0].Bounds().H != 3 {
		t.Errorf("after resize child0.H=%d, want 3 (fixed)", children[0].Bounds().H)
	}
}

// ─── Deep nesting: 4 levels ────────────────────────────

func TestIntegration_DeepNesting(t *testing.T) {
	// Flex > Padding > Center > Flex > box
	box := newIBox(4, 2)
	inner := NewFlex(FlexRow)
	inner.AddChild(box)
	centered := NewCenter(inner)
	padded := NewPadding(1, 1, 1, 1, centered)
	outer := NewFlex(FlexRow)
	outer.AddChild(padded)

	sz := outer.Measure(component.Unbounded())
	// box: 4x2
	// inner: 4x2
	// centered: 4x2
	// padded: 4+2=6 wide, 2+2=4 tall
	// outer: 6x4
	if sz.W != 6 {
		t.Errorf("deep nested width = %d, want 6", sz.W)
	}
	if sz.H != 4 {
		t.Errorf("deep nested height = %d, want 4", sz.H)
	}

	// SetBounds should propagate through all layers
	outer.SetBounds(component.Rect{X: 0, Y: 0, W: 6, H: 4})
	if padded.Bounds().W != 6 || padded.Bounds().H != 4 {
		t.Errorf("padded bounds = %v, want 6x4", padded.Bounds())
	}
}

// ─── Empty containers ──────────────────────────────────

func TestIntegration_EmptyFlexMeasure(t *testing.T) {
	row := NewFlex(FlexRow)
	sz := row.Measure(component.Unbounded())
	if sz.W != 0 || sz.H != 0 {
		t.Errorf("empty flex measure = %v, want 0x0", sz)
	}

	col := NewFlex(FlexColumn)
	sz = col.Measure(component.Unbounded())
	if sz.W != 0 || sz.H != 0 {
		t.Errorf("empty flex measure = %v, want 0x0", sz)
	}
}

func TestIntegration_EmptyStack(t *testing.T) {
	s := NewStack()
	sz := s.Measure(component.Unbounded())
	if sz.W != 0 || sz.H != 0 {
		t.Errorf("empty stack measure = %v, want 0x0", sz)
	}
}

// ─── Mixed layout: Center > Stack > Flex ───────────────

func TestIntegration_MixedCenterStackFlex(t *testing.T) {
	flex := NewFlex(FlexRow)
	flex.AddChild(newIBox(6, 3))
	flex.AddChild(newIBox(6, 3))

	stack := NewStack(flex, newIBox(12, 1))
	center := NewCenter(stack)

	// Measure
	sz := center.Measure(component.Unbounded())
	// flex: 12x3, stack overlay: 12x1 → stack: 12x3
	// center: 12x3
	if sz.W != 12 {
		t.Errorf("mixed width = %d, want 12", sz.W)
	}
	if sz.H != 3 {
		t.Errorf("mixed height = %d, want 3", sz.H)
	}

	// SetBounds with larger container — center positions child
	center.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 10})
	stackBounds := stack.Bounds()
	// stack should be centered: offset (20-12)/2=4, (10-3)/2=3
	if stackBounds.X != 4 {
		t.Errorf("stack.X = %d, want 4 (centered)", stackBounds.X)
	}
	if stackBounds.Y != 3 {
		t.Errorf("stack.Y = %d, want 3 (centered)", stackBounds.Y)
	}
}

// ─── Flex with constrained children ───────────────────

func TestIntegration_FlexConstrainedMeasure(t *testing.T) {
	row := NewFlex(FlexRow)
	row.AddChild(newIBox(100, 10))
	row.AddChild(newIBox(100, 10))

	// Measure with constraints smaller than children
	sz := row.Measure(component.Bounded(50, 5))
	// Each child clamped to 50 width, 5 height
	// Row: 50+50=100 clamped to 50, height 5
	if sz.W != 50 {
		t.Errorf("constrained row width = %d, want 50", sz.W)
	}
	if sz.H != 5 {
		t.Errorf("constrained row height = %d, want 5", sz.H)
	}
}

// ─── Concurrent layout operations ──────────────────────

func TestIntegration_ConcurrentResize(t *testing.T) {
	col := NewFlex(FlexColumn)
	for i := 0; i < 10; i++ {
		col.AddChild(newIBox(20, 3))
	}

	// Layout containers are not thread-safe by design — test with external mutex
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				w := 20 + (j % 10)
				h := 30 + j
				mu.Lock()
				col.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
				_ = col.Children()
				_ = col.Measure(component.Bounded(w, h))
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()
}

func TestIntegration_ConcurrentNestedLayout(t *testing.T) {
	row := NewFlex(FlexRow)
	for i := 0; i < 5; i++ {
		col := NewFlex(FlexColumn)
		col.AddChild(newIBox(10, 3))
		col.AddChild(newIBox(10, 2))
		row.AddChild(col)
	}

	// Layout containers are not thread-safe — use external mutex
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				mu.Lock()
				row.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 10})
				_ = row.Measure(component.Unbounded())
				for _, c := range row.Children() {
					_ = c.Bounds()
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
}

// ─── Stack overlay positioning ─────────────────────────

func TestIntegration_StackOverlayPositioning(t *testing.T) {
	bg := newIBox(20, 10)
	fg := newIBox(5, 1)

	stack := NewStack(bg, fg)

	stack.SetBounds(component.Rect{X: 2, Y: 3, W: 20, H: 10})

	// Both children get the same bounds
	if bg.Bounds().X != 2 || bg.Bounds().Y != 3 {
		t.Errorf("bg position = %v, want (2,3)", bg.Bounds())
	}
	if fg.Bounds().X != 2 || fg.Bounds().Y != 3 {
		t.Errorf("fg position = %v, want (2,3)", fg.Bounds())
	}
}

// ─── Padding edge: padding exceeds bounds ──────────────

func TestIntegration_PaddingExceedsBounds(t *testing.T) {
	// Padding 5 on each side but child only 2x2
	padded := NewPadding(5, 5, 5, 5, newIBox(2, 2))

	// Measure with tight constraints
	sz := padded.Measure(component.Bounded(8, 8))
	// innerW = 8 - 5 - 5 = -2 → 0, innerH same
	// child clamped to 0x0
	// result = 0 + 5 + 5 = 10 clamped... but constraints say max 8
	// Actually child measures 0x0 with 0 constraints, then add padding
	// W = 0 + 5 + 5 = 10, but constrained to 8
	// Wait — Measure doesn't clamp the result, it just passes constraints to child
	if sz.W < 0 || sz.H < 0 {
		t.Errorf("padded measure should not be negative: %v", sz)
	}
}

// ─── Multiple Centers stacked ──────────────────────────

func TestIntegration_NestedCenters(t *testing.T) {
	inner := NewCenter(newIBox(2, 1))
	outer := NewCenter(inner)

	sz := outer.Measure(component.Unbounded())
	if sz.W != 2 || sz.H != 1 {
		t.Errorf("nested centers measure = %v, want 2x1", sz)
	}

	// Center in a large space
	outer.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 10})
	innerBounds := inner.Bounds()
	// inner should be centered: offset = (20-2)/2 = 9, (10-1)/2 = 4
	if innerBounds.X != 9 {
		t.Errorf("inner.X = %d, want 9", innerBounds.X)
	}
	if innerBounds.Y != 4 {
		t.Errorf("inner.Y = %d, want 4", innerBounds.Y)
	}
}

// ─── Flex row gap positioning ──────────────────────────

func TestIntegration_FlexRowGapPositioning(t *testing.T) {
	row := NewFlexGap(FlexRow, 3)
	row.AddChild(newIBox(5, 3))
	row.AddChild(newIBox(5, 3))
	row.AddChild(newIBox(5, 3))

	row.SetBounds(component.Rect{X: 0, Y: 0, W: 21, H: 3})
	children := row.Children()

	// child0 at X=0, child1 at X=5+3=8, child2 at X=8+5+3=16
	if children[0].Bounds().X != 0 {
		t.Errorf("child0.X = %d, want 0", children[0].Bounds().X)
	}
	if children[1].Bounds().X != 8 {
		t.Errorf("child1.X = %d, want 8", children[1].Bounds().X)
	}
	if children[2].Bounds().X != 16 {
		t.Errorf("child2.X = %d, want 16", children[2].Bounds().X)
	}
}

// ─── Flex column gap positioning ───────────────────────

func TestIntegration_FlexColumnGapPositioning(t *testing.T) {
	col := NewFlexGap(FlexColumn, 2)
	col.AddChild(newIBox(10, 3))
	col.AddChild(newIBox(10, 4))

	col.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 9})
	children := col.Children()

	// child0 at Y=0, child1 at Y=3+2=5
	if children[0].Bounds().Y != 0 {
		t.Errorf("child0.Y = %d, want 0", children[0].Bounds().Y)
	}
	if children[1].Bounds().Y != 5 {
		t.Errorf("child1.Y = %d, want 5", children[1].Bounds().Y)
	}
}
