package render

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// dummyWriter implements io.Writer for benchmarking the renderer
// without writing to a real terminal.
type benchWriter struct{}

func (benchWriter) Write(p []byte) (int, error) { return len(p), nil }

// fillBufferWithText fills the buffer with varied text content to simulate
// a realistic terminal state (mixed styles, colors, text).
func fillBufferWithText(buf *buffer.Buffer) {
	styles := []buffer.Style{
		buffer.DefaultStyle.WithFg(buffer.RGB(255, 121, 198)),
		buffer.DefaultStyle.WithFg(buffer.RGB(139, 233, 253)).AddFlags(buffer.Bold),
		buffer.DefaultStyle.WithFg(buffer.RGB(80, 250, 123)),
		buffer.DefaultStyle.WithFg(buffer.RGB(248, 248, 242)).AddFlags(buffer.Italic),
		buffer.DefaultStyle.WithFg(buffer.RGB(189, 147, 249)).AddFlags(buffer.Underline),
	}
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			c := byte('A' + (x*7+y*3)%52)
			if c > 'Z' {
				c += 'a' - 'A' - 5
			}
			buf.SetCell(x, y, buffer.Cell{
				Rune:  rune(c),
				Width: 1,
				Fg:    styles[(x+y)%len(styles)].Fg,
				Bg:    buffer.RGB(40, 42, 54),
				Flags: styles[(x+y)%len(styles)].Flags,
			})
		}
	}
}

// BenchmarkRenderFull measures rendering a full 80x24 screen from scratch
// (front buffer is blank, back buffer has full content).
func BenchmarkRenderFull(b *testing.B) {
	tw := term.NewWriter(benchWriter{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.BeginFrame()
		fillBufferWithText(r.Back())
		_ = r.EndFrame()
	}
}

// BenchmarkRenderDiff measures rendering with small changes (1-5 cells changed)
// between frames. This is the most common real-world scenario.
func BenchmarkRenderDiff(b *testing.B) {
	tw := term.NewWriter(benchWriter{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// Initial frame to populate front buffer.
	r.BeginFrame()
	fillBufferWithText(r.Back())
	_ = r.EndFrame()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.BeginFrame()
		fillBufferWithText(r.Back())
		// Change 1-5 cells to simulate cursor movement / small update.
		for j := 0; j < 5; j++ {
			r.Back().SetCell(j, 0, buffer.Cell{
				Rune:  rune('X'),
				Width: 1,
				Fg:    buffer.RGB(255, 255, 255),
				Bg:    buffer.RGB(40, 42, 54),
			})
		}
		_ = r.EndFrame()
	}
}

// BenchmarkRenderNoChange measures rendering with zero changes between frames.
// The diff should short-circuit on identical rows.
func BenchmarkRenderNoChange(b *testing.B) {
	tw := term.NewWriter(benchWriter{}, term.ProfileTrue)
	r := New(tw, 80, 24)

	// Initial frame.
	r.BeginFrame()
	fillBufferWithText(r.Back())
	_ = r.EndFrame()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.BeginFrame()
		// Redraw identical content.
		fillBufferWithText(r.Back())
		_ = r.EndFrame()
	}
}

// BenchmarkRenderLargeScreen measures rendering a large 200x50 terminal.
func BenchmarkRenderLargeScreen(b *testing.B) {
	tw := term.NewWriter(benchWriter{}, term.ProfileTrue)
	r := New(tw, 200, 50)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.BeginFrame()
		fillBufferWithText(r.Back())
		_ = r.EndFrame()
	}
}

// BenchmarkRenderSequentialBytes measures output byte count for sequential
// text rendering — the cursor tracking optimization should reduce bytes
// by skipping MoveTo for adjacent cells.
func BenchmarkRenderSequentialBytes(b *testing.B) {
	counter := &countWriter{}
	tw := term.NewWriter(counter, term.ProfileTrue)
	r := New(tw, 80, 24)

	// Initial fill.
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			r.Back().SetCell(x, y, buffer.Cell{
				Rune:  'A',
				Width: 1,
				Fg:    buffer.RGB(255, 255, 255),
			})
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	var total int
	for i := 0; i < b.N; i++ {
		counter.n = 0
		r.BeginFrame()
		for y := 0; y < 24; y++ {
			for x := 0; x < 80; x++ {
				r.Back().SetCell(x, y, buffer.Cell{
					Rune:  rune('A' + (i % 26)),
					Width: 1,
					Fg:    buffer.RGB(255, 255, 255),
				})
			}
		}
		_ = r.EndFrame()
		total = counter.n
	}
	b.ReportMetric(float64(total)/1024, "KB/op")
}

type countWriter struct {
	n int
}

func (c *countWriter) Write(p []byte) (int, error) {
	c.n += len(p)
	return len(p), nil
}
