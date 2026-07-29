package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── P482: Zero-alloc Paint benchmarks for P480-P481 components ───

func BenchmarkPaintMarkdownBlockquote(b *testing.B) {
	bq := NewMarkdownBlockquote()
	bq.SetMarkdown("> First quote line.\n> Second line of quote.\n> Third line.\n>> Nested quote here.\n>>> Deeply nested.\n> Back to outer.\n> Another line.")
	bq.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})
	buf := buffer.NewBuffer(60, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bq.Paint(buf)
	}
}

func BenchmarkPaintCarousel(b *testing.B) {
	c := NewCarousel()
	c.AddSlide("Welcome", "Get started with Fluui TUI library!")
	c.AddSlide("Components", "Over 160 components available")
	c.AddSlide("AI Native", "Built-in AI response rendering")
	c.AddSlide("Performance", "Zero-allocation Paint verified")
	c.AddSlide("Terminal Protocols", "80+ protocol functions")
	c.AddSlide("Examples", "160+ runnable examples")
	c.SetCurrent(2)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}
