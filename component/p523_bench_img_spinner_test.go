package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownImage(b *testing.B) {
	mi := NewMarkdownImage()
	mi.SetMarkdown("Gallery: ![photo1](https://example.com/p1.jpg) and ![screenshot](https://example.com/shot.png) plus ![diagram](https://example.com/diag.svg) in this text.")
	mi.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mi.Paint(buf)
	}
}

func BenchmarkPaintSpinnerDots(b *testing.B) {
	sd := NewSpinnerDots()
	sd.SetLabel("Loading data from server")
	sd.SetDotCount(5)
	sd.Advance()
	sd.Advance()
	sd.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sd.Paint(buf)
	}
}
