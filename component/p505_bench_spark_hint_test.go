package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintKeyHintBar(b *testing.B) {
	kb := NewKeyHintBar()
	kb.AddHint("Q", "Quit")
	kb.AddHint("S", "Save")
	kb.AddHint("/", "Search")
	kb.AddHint("R", "Refresh")
	kb.AddHint("T", "Theme")
	kb.AddHint("H", "Help")
	kb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		kb.Paint(buf)
	}
}

func BenchmarkPaintDataLabel(b *testing.B) {
	dl := NewDataLabel()
	dl.SetLabel("Monthly Revenue")
	dl.SetValue(154320.5)
	dl.SetUnit(" USD")
	dl.SetTrend(DataTrendUp)
	dl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dl.Paint(buf)
	}
}
