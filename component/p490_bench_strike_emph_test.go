package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── P490: Zero-alloc benchmarks for MarkdownStrikethrough + MarkdownEmphasis ───

func BenchmarkPaintMarkdownStrikethrough(b *testing.B) {
	ms := NewMarkdownStrikethrough()
	ms.SetMarkdown("This ~~old method~~ is replaced by the ~~legacy API~~ which was ~~deprecated~~ in favor of the ~new approach~ and ~cleaner code~.")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ms.Paint(buf)
	}
}

func BenchmarkPaintMarkdownEmphasis(b *testing.B) {
	me := NewMarkdownEmphasis()
	me.SetMarkdown("This is **bold text** and *italic text* with ***bold italic*** plus __more bold__ and _more italic_ in the same line.")
	me.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		me.Paint(buf)
	}
}
