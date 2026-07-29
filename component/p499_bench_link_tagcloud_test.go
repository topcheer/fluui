package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownLink(b *testing.B) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("Visit [Fluui](https://fluui.dev) and [GitHub](https://github.com) then <https://example.com> for more. See [docs][1] and [api](https://api.fluui.dev) for details.\n\n[1]: https://docs.fluui.dev")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ml.Paint(buf)
	}
}

func BenchmarkPaintTagCloud(b *testing.B) {
	tc := NewTagCloud()
	tc.SetTags([]TagItem{
		{Name: "go", Weight: 80}, {Name: "tui", Weight: 60}, {Name: "ai", Weight: 90},
		{Name: "terminal", Weight: 40}, {Name: "library", Weight: 50}, {Name: "markdown", Weight: 30},
		{Name: "rendering", Weight: 25}, {Name: "component", Weight: 70}, {Name: "zero-alloc", Weight: 45},
		{Name: "benchmark", Weight: 20}, {Name: "protocol", Weight: 35}, {Name: "integration", Weight: 15},
	})
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tc.Paint(buf)
	}
}
