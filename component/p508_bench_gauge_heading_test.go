package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMiniGauge(b *testing.B) {
	mg := NewMiniGauge()
	mg.SetLabel("Memory Usage")
	mg.SetValue(73.5)
	mg.SetMax(100)
	mg.SetWidth(20)
	mg.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mg.Paint(buf)
	}
}

func BenchmarkPaintMarkdownHeading(b *testing.B) {
	mh := NewMarkdownHeading()
	mh.SetMarkdown("## Important Section Title Here")
	mh.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2})
	buf := buffer.NewBuffer(60, 2)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mh.Paint(buf)
	}
}
