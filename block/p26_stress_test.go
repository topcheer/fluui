package block

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P26 stress benchmarks: large-scale and extreme workloads.

// BenchmarkContainerPaint1000 benchmarks painting 1000 blocks — simulates
// a very long conversation with 1000 messages.
func BenchmarkContainerPaint1000(b *testing.B) {
	c := NewBlockContainer()
	for j := 0; j < 1000; j++ {
		blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
		blk.AppendDelta(fmt.Sprintf("Block #%d: Lorem ipsum dolor sit amet, consectetur adipiscing elit.", j))
		blk.Complete()
		c.AddBlock(blk)
	}

	buf := buffer.NewBuffer(80, 3000)
	bounds := component.Rect{X: 0, Y: 0, W: 80, H: 3000}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.SetBounds(bounds)
		c.Paint(buf)
	}
}

// BenchmarkContainerPaintMarkdown benchmarks painting blocks with complex markdown.
func BenchmarkContainerPaintMarkdown(b *testing.B) {
	c := NewBlockContainer()
	markdown := `## Heading

This has **bold** and *italic* text.

- Item 1
- Item 2
- Item 3

` + "```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```\n\n" +
		"> A blockquote with some text."

	for j := 0; j < 50; j++ {
		blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
		blk.AppendDelta(markdown)
		blk.Complete()
		c.AddBlock(blk)
	}

	buf := buffer.NewBuffer(80, 500)
	bounds := component.Rect{X: 0, Y: 0, W: 80, H: 500}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.SetBounds(bounds)
		c.Paint(buf)
	}
}

// BenchmarkStreamingDelta benchmarks appending streaming deltas — simulates
// real-time AI response with frequent small updates.
func BenchmarkStreamingDelta(b *testing.B) {
	// Pre-allocate words outside the loop to isolate streaming allocations
	words := strings.Split("The quick brown fox jumps over the lazy dog and runs through the forest", " ")
	wordPtrs := make([]string, len(words))
	for i, w := range words {
		wordPtrs[i] = w + " "
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		blk := NewAssistantTextBlock("stream")
		for _, w := range wordPtrs {
			blk.AppendDelta(w)
		}
		blk.Complete()
	}
}

// BenchmarkLargeBlock benchmarks a single very large block (50K chars).
func BenchmarkLargeBlock(b *testing.B) {
	large := strings.Repeat("Lorem ipsum dolor sit amet. ", 2000) // ~56K chars
	buf := buffer.NewBuffer(80, 1000)
	bounds := component.Rect{X: 0, Y: 0, W: 80, H: 1000}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		blk := NewAssistantTextBlock("large")
		blk.AppendDelta(large)
		blk.Complete()
		blk.SetBounds(bounds)
		blk.Paint(buf)
	}
}

// --- Memory leak detection ---

// TestP26_NoMemoryLeak_ContainerPaint verifies that repeated Paint calls
// don't accumulate memory. We measure heap growth over many iterations.
func TestP26_NoMemoryLeak_ContainerPaint(t *testing.T) {
	c := NewBlockContainer()
	for j := 0; j < 50; j++ {
		blk := NewAssistantTextBlock(fmt.Sprintf("blk-%d", j))
		blk.AppendDelta(fmt.Sprintf("Block #%d content with some text.", j))
		blk.Complete()
		c.AddBlock(blk)
	}

	buf := buffer.NewBuffer(80, 200)
	bounds := component.Rect{X: 0, Y: 0, W: 80, H: 200}

	// Warm up
	for i := 0; i < 100; i++ {
		c.SetBounds(bounds)
		c.Paint(buf)
	}
	runtime.GC()
	baseAllocs := getHeapAllocs()

	// Run many paint cycles
	for i := 0; i < 1000; i++ {
		c.SetBounds(bounds)
		c.Paint(buf)
	}
	runtime.GC()
	finalAllocs := getHeapAllocs()

	// Heap should not grow significantly (allow some GC slack)
	var growth int64
	if finalAllocs >= baseAllocs {
		growth = int64(finalAllocs - baseAllocs)
	}
	if growth > 1024*1024 { // > 1MB growth = potential leak
		t.Errorf("possible memory leak: heap grew %d bytes after 1000 paint cycles", growth)
	}
}

// TestP26_NoMemoryLeak_Streaming verifies streaming deltas don't leak.
func TestP26_NoMemoryLeak_Streaming(t *testing.T) {
	runtime.GC()
	baseAllocs := getHeapAllocs()

	for i := 0; i < 100; i++ {
		blk := NewAssistantTextBlock("test")
		for j := 0; j < 100; j++ {
			blk.AppendDelta("word ")
		}
		blk.Complete()
		// blk goes out of scope — should be GC'd
	}

	runtime.GC()
	finalAllocs := getHeapAllocs()

	var growth int64
	if finalAllocs >= baseAllocs {
		growth = int64(finalAllocs - baseAllocs)
	}
	if growth > 512*1024 { // > 512KB = potential leak
		t.Errorf("possible memory leak: heap grew %d bytes after 100 stream cycles", growth)
	}
}

// TestP26_NoGoroutineLeak verifies that block operations don't leak goroutines.
func TestP26_NoGoroutineLeak(t *testing.T) {
	baseGoroutines := runtime.NumGoroutine()

	for i := 0; i < 100; i++ {
		c := NewBlockContainer()
		blk := NewAssistantTextBlock("test")
		blk.AppendDelta("content")
		blk.Complete()
		c.AddBlock(blk)

		buf := buffer.NewBuffer(80, 24)
		c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
		c.Paint(buf)
	}

	// Allow goroutines to settle
	runtime.GC()
	time := 0
	for time < 100 {
		current := runtime.NumGoroutine()
		if current <= baseGoroutines+2 { // allow small slack
			return
		}
		runtime.Gosched()
		time += 10
	}

	current := runtime.NumGoroutine()
	if current > baseGoroutines+5 {
		t.Errorf("goroutine leak: started with %d, now %d", baseGoroutines, current)
	}
}

// getHeapAllocs returns the current heap allocation in bytes.
func getHeapAllocs() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}
