package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// --- BlockPosition cache tests ---

func TestBlockPositionsCached(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(1)
	c.AddBlock(newMockBlock("b1", TypeThinking, 10, 3))
	c.AddBlock(newMockBlock("b2", TypeAssistantText, 10, 5))

	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 20})

	positions := c.BlockPositions()
	if len(positions) != 2 {
		t.Fatalf("len(positions) = %d, want 2", len(positions))
	}

	// b1: localY=0, H=3
	if positions[0].Y != 0 || positions[0].H != 3 {
		t.Errorf("positions[0] = {Y:%d, H:%d}, want {0, 3}", positions[0].Y, positions[0].H)
	}
	// b2: localY=3+1(spacing)=4, H=5
	if positions[1].Y != 4 || positions[1].H != 5 {
		t.Errorf("positions[1] = {Y:%d, H:%d}, want {4, 5}", positions[1].Y, positions[1].H)
	}
}

func TestTotalHeight(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(2)
	c.AddBlock(newMockBlock("b1", TypeThinking, 10, 3))
	c.AddBlock(newMockBlock("b2", TypeAssistantText, 10, 5))

	c.Measure(component.Unbounded())
	// 3 + 2(spacing) + 5 = 10
	if c.TotalHeight() != 10 {
		t.Errorf("TotalHeight() = %d, want 10", c.TotalHeight())
	}
}

// --- PaintVisible tests ---

func TestPaintVisibleAllBlocks(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(0)
	c.AddBlock(newMockBlock("b1", TypeThinking, 10, 3))
	c.AddBlock(newMockBlock("b2", TypeAssistantText, 10, 5))

	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})

	buf := buffer.NewBuffer(10, 10)
	// Visible range covers all blocks
	c.PaintVisible(buf, 0, 10)

	// Should not panic — both blocks painted
	// (mockBlock.Paint is a no-op, so we just verify no panic)
}

func TestPaintVisibleSkipsOffscreen(t *testing.T) {
	// Create countingBlock to track which blocks get painted
	c := NewBlockContainer()
	c.SetSpacing(0)

	b1 := newCountingBlock("b1", 10, 3)
	b2 := newCountingBlock("b2", 10, 5)
	b3 := newCountingBlock("b3", 10, 4)
	c.AddBlock(b1)
	c.AddBlock(b2)
	c.AddBlock(b3)

	// Total height = 3 + 5 + 4 = 12, no spacing
	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 12})

	buf := buffer.NewBuffer(10, 12)

	// Visible range: only b2 (Y=3, H=5 → [3,8)) intersects [4, 7)
	c.PaintVisible(buf, 4, 7)

	if b1.paintCount != 0 {
		t.Errorf("b1 painted %d times, want 0 (offscreen above)", b1.paintCount)
	}
	if b2.paintCount != 1 {
		t.Errorf("b2 painted %d times, want 1 (visible)", b2.paintCount)
	}
	if b3.paintCount != 0 {
		t.Errorf("b3 painted %d times, want 0 (offscreen below)", b3.paintCount)
	}
}

func TestPaintVisibleBoundaryBlock(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(0)

	// Spacer block: Y=0, H=5 → occupies [0, 5)
	spacer := newCountingBlock("spacer", 10, 5)
	// Target block: Y=5, H=10 → occupies [5, 15)
	b1 := newCountingBlock("b1", 10, 10)
	c.AddBlock(spacer)
	c.AddBlock(b1)

	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 20})

	buf := buffer.NewBuffer(10, 20)

	// Visible [0, 5) — b1 starts at 5, so blockTop=5 >= visibleY1=5 → skip
	c.PaintVisible(buf, 0, 5)
	if b1.paintCount != 0 {
		t.Errorf("b1 painted %d, want 0 (blockTop=5 >= visibleY1=5)", b1.paintCount)
	}

	// Visible [0, 6) — b1 [5,15) intersects → paint
	b1.paintCount = 0
	c.PaintVisible(buf, 0, 6)
	if b1.paintCount != 1 {
		t.Errorf("b1 painted %d, want 1 (intersects)", b1.paintCount)
	}

	// Visible [15, 20) — b1 blockBottom=15 <= visibleY0=15 → skip
	b1.paintCount = 0
	c.PaintVisible(buf, 15, 20)
	if b1.paintCount != 0 {
		t.Errorf("b1 painted %d, want 0 (blockBottom=15 <= visibleY0=15)", b1.paintCount)
	}
}

// --- VisiblePainter interface test ---

func TestBlockContainerImplementsVisiblePainter(t *testing.T) {
	c := NewBlockContainer()
	if !component.IsVisiblePainter(c) {
		t.Error("BlockContainer should implement component.VisiblePainter")
	}
}

// --- Benchmark ---

func BenchmarkPaintVisible1000Blocks(b *testing.B) {
	c := NewBlockContainer()
	c.SetSpacing(0)
	for i := 0; i < 1000; i++ {
		c.AddBlock(newMockBlock("b", TypeAssistantText, 80, 3))
	}

	c.Measure(component.Bounded(80, 0))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3000})

	buf := buffer.NewBuffer(80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate scrolling: only ~8 blocks visible (24 rows / 3 height)
		c.PaintVisible(buf, 1500, 1524)
	}
}

func BenchmarkPaintVisible10000Blocks(b *testing.B) {
	c := NewBlockContainer()
	c.SetSpacing(0)
	for i := 0; i < 10000; i++ {
		c.AddBlock(newMockBlock("b", TypeAssistantText, 80, 3))
	}

	c.Measure(component.Bounded(80, 0))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 30000})

	buf := buffer.NewBuffer(80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Only ~8 blocks visible in the middle of 10000
		c.PaintVisible(buf, 15000, 15024)
	}
}

func BenchmarkPaintAll1000Blocks(b *testing.B) {
	c := NewBlockContainer()
	c.SetSpacing(0)
	for i := 0; i < 1000; i++ {
		c.AddBlock(newMockBlock("b", TypeAssistantText, 80, 3))
	}

	c.Measure(component.Bounded(80, 0))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3000})

	buf := buffer.NewBuffer(80, 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

// --- countingBlock: tracks Paint calls for virtual scroll testing ---

type countingBlock struct {
	BaseBlock
	w, h       int
	paintCount int
}

func newCountingBlock(id string, w, h int) *countingBlock {
	return &countingBlock{
		BaseBlock: NewBaseBlock(id, TypeAssistantText),
		w:         w,
		h:         h,
	}
}

func (c *countingBlock) Measure(cs component.Constraints) component.Size {
	w, h := c.w, c.h
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return component.Size{W: w, H: h}
}

func (c *countingBlock) SetBounds(r component.Rect) {
	c.BaseComponent.SetBounds(r)
}

func (c *countingBlock) Paint(buf *buffer.Buffer) {
	c.paintCount++
}
