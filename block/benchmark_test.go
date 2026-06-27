package block

import (
	"fmt"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// BenchmarkContainerAddBlocks100 benchmarks adding 100 blocks to a container.
func BenchmarkContainerAddBlocks100(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := NewBlockContainer()
		for j := 0; j < 100; j++ {
			blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
			blk.AppendDelta(fmt.Sprintf("Block number %d content", j))
			blk.Complete()
			c.AddBlock(blk)
		}
	}
}

// BenchmarkContainerAddBlocks1000 benchmarks adding 1000 blocks.
func BenchmarkContainerAddBlocks1000(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := NewBlockContainer()
		for j := 0; j < 1000; j++ {
			blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
			blk.AppendDelta(fmt.Sprintf("Block number %d", j))
			blk.Complete()
			c.AddBlock(blk)
		}
	}
}

// BenchmarkContainerPaint100 benchmarks painting 100 blocks into a buffer.
// This measures the full Measure + SetBounds + Paint cycle.
func BenchmarkContainerPaint100(b *testing.B) {
	c := NewBlockContainer()
	for j := 0; j < 100; j++ {
		blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
		blk.AppendDelta(fmt.Sprintf("Block #%d: Lorem ipsum dolor sit amet, consectetur adipiscing elit.", j))
		blk.Complete()
		c.AddBlock(blk)
	}

	buf := buffer.NewBuffer(80, 300)
	bounds := component.Rect{X: 0, Y: 0, W: 80, H: 300}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.SetBounds(bounds)
		c.Paint(buf)
	}
}

// BenchmarkContainerBlocks benchmarks taking a snapshot of all blocks.
// Blocks() returns a copy of the slice, so this measures the copy cost.
func BenchmarkContainerBlocks(b *testing.B) {
	c := NewBlockContainer()
	for j := 0; j < 100; j++ {
		blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
		blk.AppendDelta("content")
		blk.Complete()
		c.AddBlock(blk)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = c.Blocks()
	}
}
