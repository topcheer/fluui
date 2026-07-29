package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownList(b *testing.B) {
	ml := NewMarkdownList()
	ml.SetMarkdown("- Item alpha\n  - Nested alpha one\n  - Nested alpha two\n    - Deep nested\n- Item beta\n  1. Ordered one\n  2. Ordered two\n- Item gamma\n  - Nested gamma\n    - Deep gamma one\n    - Deep gamma two\n- Item delta")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 16})
	buf := buffer.NewBuffer(60, 16)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ml.Paint(buf)
	}
}

func BenchmarkPaintBreadcrumbTrail(b *testing.B) {
	bt := NewBreadcrumbTrail()
	bt.SetCrumbs([]string{"Root", "Workspace", "Projects", "Fluui", "Component", "Markdown", "Rendering", "Current"})
	bt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bt.Paint(buf)
	}
}
