package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownFootnote(b *testing.B) {
	mf := NewMarkdownFootnote()
	mf.SetMarkdown("First reference[^1] and second[^2] and third[^3].\n\n[^1]: First definition with details\n[^2]: Second explanation here\n[^3]: Third supplementary note")
	mf.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mf.Paint(buf)
	}
}

func BenchmarkPaintLegend(b *testing.B) {
	l := NewLegend()
	l.AddEntry("Revenue", buffer.RGB(34, 197, 94))
	l.AddEntry("Operating Costs", buffer.RGB(239, 68, 68))
	l.AddEntry("Net Profit", buffer.RGB(96, 165, 250))
	l.AddEntry("Taxes", buffer.RGB(234, 179, 8))
	l.AddEntry("R&D", buffer.RGB(167, 139, 250))
	l.AddEntry("Marketing", buffer.RGB(244, 114, 182))
	l.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		l.Paint(buf)
	}
}
