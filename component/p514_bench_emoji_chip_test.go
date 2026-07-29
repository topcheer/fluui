package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownEmoji(b *testing.B) {
	me := NewMarkdownEmoji()
	me.SetMarkdown("Hello :smile: world :heart: :fire: :rocket: :thumbsup: :tada: :bug: :zap: :book: :code: :wrench: :star: :trophy: :chart: done!")
	me.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		me.Paint(buf)
	}
}

func BenchmarkPaintChipBadge(b *testing.B) {
	cb := NewChipBadge()
	cb.AddChip("production")
	cb.AddChip("staging")
	cb.AddChip("development")
	cb.AddChip("testing")
	cb.AddChip("experimental")
	cb.ToggleSelected(2)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Paint(buf)
	}
}
