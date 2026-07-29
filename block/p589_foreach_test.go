package block

import (
	"fmt"
	"testing"
)

// BenchmarkForEachBlockZeroAlloc verifies ForEachBlock does not allocate.
func BenchmarkForEachBlockZeroAlloc(b *testing.B) {
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
		c.ForEachBlock(func(blk Block) bool {
			return true
		})
	}
}

// TestForEachBlock verifies correct iteration behavior.
func TestForEachBlock(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("a"))
	c.AddBlock(NewAssistantTextBlock("b"))
	c.AddBlock(NewAssistantTextBlock("c"))

	count := 0
	c.ForEachBlock(func(b Block) bool {
		count++
		return true
	})
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
}

// TestForEachBlockEarlyStop verifies iteration stops on false.
func TestForEachBlockEarlyStop(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("a"))
	c.AddBlock(NewAssistantTextBlock("b"))
	c.AddBlock(NewAssistantTextBlock("c"))

	count := 0
	c.ForEachBlock(func(b Block) bool {
		count++
		return count < 2 // stop after 2
	})
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// TestForEachBlockEmpty verifies no panic on empty container.
func TestForEachBlockEmpty(t *testing.T) {
	c := NewBlockContainer()
	c.ForEachBlock(func(b Block) bool {
		t.Error("callback should not be called on empty container")
		return true
	})
}
