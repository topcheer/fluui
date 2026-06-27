package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// --- Edge case tests for PaintVisible binary search ---

func TestPaintVisibleEmptyContainer(t *testing.T) {
	c := NewBlockContainer()
	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})

	buf := buffer.NewBuffer(10, 10)
	// Should not panic
	c.PaintVisible(buf, 0, 10)
}

func TestPaintVisibleSingleBlock(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(0)
	b1 := newCountingBlock("b1", 10, 5)
	c.AddBlock(b1)

	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})

	buf := buffer.NewBuffer(10, 5)

	// Block fully visible
	c.PaintVisible(buf, 0, 5)
	if b1.paintCount != 1 {
		t.Errorf("b1 painted %d, want 1", b1.paintCount)
	}

	// Block above visible range
	b1.paintCount = 0
	c.PaintVisible(buf, 5, 10)
	if b1.paintCount != 0 {
		t.Errorf("b1 painted %d, want 0 (above visible)", b1.paintCount)
	}
}

func TestPaintVisibleOffsetZero(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(1)

	b1 := newCountingBlock("b1", 10, 3)
	b2 := newCountingBlock("b2", 10, 3)
	b3 := newCountingBlock("b3", 10, 3)
	c.AddBlock(b1)
	c.AddBlock(b2)
	c.AddBlock(b3)

	// Layout: b1 [0,3), gap [3,4), b2 [4,7), gap [7,8), b3 [8,11)
	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 11})

	buf := buffer.NewBuffer(10, 11)

	// Offset=0, visible [0, 4) → b1 fully + gap → b1 only
	c.PaintVisible(buf, 0, 4)
	if b1.paintCount != 1 {
		t.Errorf("b1 painted %d, want 1", b1.paintCount)
	}
	if b2.paintCount != 0 {
		t.Errorf("b2 painted %d, want 0", b2.paintCount)
	}
}

func TestPaintVisibleAllVisible(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(0)

	b1 := newCountingBlock("b1", 10, 3)
	b2 := newCountingBlock("b2", 10, 3)
	b3 := newCountingBlock("b3", 10, 3)
	c.AddBlock(b1)
	c.AddBlock(b2)
	c.AddBlock(b3)

	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 9})

	buf := buffer.NewBuffer(10, 9)

	// All blocks visible [0, 9)
	c.PaintVisible(buf, 0, 9)
	if b1.paintCount != 1 || b2.paintCount != 1 || b3.paintCount != 1 {
		t.Errorf("paint counts: b1=%d b2=%d b3=%d, want all 1", b1.paintCount, b2.paintCount, b3.paintCount)
	}
}

func TestPaintVisibleBinarySearchCorrectness(t *testing.T) {
	// Test with many blocks to exercise binary search path
	c := NewBlockContainer()
	c.SetSpacing(0)

	const n = 100
	blocks := make([]*countingBlock, n)
	for i := 0; i < n; i++ {
		blocks[i] = newCountingBlock("b", 10, 3)
		c.AddBlock(blocks[i])
	}

	c.Measure(component.Unbounded())
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 300})

	buf := buffer.NewBuffer(10, 24)

	// Visible range [150, 174) → blocks 50-57 (each height 3)
	c.PaintVisible(buf, 150, 174)

	painted := 0
	for i, b := range blocks {
		if b.paintCount > 0 {
			painted++
			// Verify painted blocks are in visible range
			pos := c.positions[i]
			if pos.Y+pos.H <= 150 || pos.Y >= 174 {
				t.Errorf("block %d painted but Y=%d H=%d not in [150,174)", i, pos.Y, pos.H)
			}
		}
	}

	if painted == 0 {
		t.Error("expected at least 1 block painted")
	}
	// Only ~8 blocks should be visible (24/3=8)
	if painted > 10 {
		t.Errorf("painted %d blocks, expected ~8", painted)
	}
}

// --- Benchmark: binary search vs linear scan comparison ---

func BenchmarkPaintVisibleBinarySearch(b *testing.B) {
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
		// Scrolling in the middle: offset 1500, viewport 24 rows
		c.PaintVisible(buf, 1500, 1524)
	}
}

// BenchmarkPaintVisibleLinearScan simulates the old O(n) approach for comparison.
func BenchmarkPaintVisibleLinearScan(b *testing.B) {
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
		// Simulate linear scan: iterate all blocks, paint only visible ones
		positions := c.BlockPositions()
		for idx := 0; idx < len(c.blocks); idx++ {
			pos := positions[idx]
			if pos.Y+pos.H > 1500 && pos.Y < 1524 {
				c.blocks[idx].Paint(buf)
			}
		}
	}
}
