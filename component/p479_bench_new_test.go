package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── P479: Zero-alloc Paint benchmarks for P477-P478 components ───

func BenchmarkPaintMarkdownTable(b *testing.B) {
	mt := NewMarkdownTable()
	mt.SetMarkdown("| Name | Score | Rank |\n|:-----|-----:|:----:|\n| Alpha | 95 | A |\n| Beta | 87 | B |\n| Gamma | 72 | C |\n| Delta | 65 | D |")
	mt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mt.Paint(buf)
	}
}
