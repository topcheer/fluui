package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownDefinitionList(b *testing.B) {
	dl := NewMarkdownDefinitionList()
	dl.SetMarkdown("Go\n: A compiled, statically typed language\nRust\n: A systems programming language with memory safety\nPython\n: A high-level interpreted language\nTypeScript\n: A typed superset of JavaScript\nZig\n: A general-purpose programming language")
	dl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 14})
	buf := buffer.NewBuffer(60, 14)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dl.Paint(buf)
	}
}

func BenchmarkPaintStatusBarSegment(b *testing.B) {
	sb := NewStatusBarSegment()
	sb.AddSegment("Branch", "main", buffer.RGB(255, 255, 255), buffer.RGB(34, 197, 94))
	sb.AddSegment("Errors", "0", buffer.RGB(255, 255, 255), buffer.RGB(239, 68, 68))
	sb.AddSegment("Warnings", "3", buffer.RGB(255, 255, 255), buffer.RGB(234, 179, 8))
	sb.AddSegment("Encoding", "UTF-8", buffer.RGB(255, 255, 255), buffer.RGB(96, 165, 250))
	sb.AddSegment("LN", "42", buffer.RGB(255, 255, 255), buffer.RGB(148, 163, 184))
	sb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sb.Paint(buf)
	}
}
