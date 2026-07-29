package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func BenchmarkPaintMarkdownSubscript(b *testing.B) {
	ms := NewMarkdownSubscript()
	ms.SetMarkdown("Chemical formulas: H~2~O, CO~2~, H~2~SO~4~, NaCl, CaCO~3~, CH~3~COOH, C~6~H~12~O~6~")
	ms.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ms.Paint(buf)
	}
}

func BenchmarkPaintMultiSelect(b *testing.B) {
	ms := NewMultiSelect()
	ms.AddOption("Option Alpha")
	ms.AddOption("Option Beta")
	ms.AddOption("Option Gamma")
	ms.AddOption("Option Delta")
	ms.AddOption("Option Epsilon")
	ms.AddOption("Option Zeta")
	ms.AddOption("Option Eta")
	ms.AddOption("Option Theta")
	ms.Toggle(1)
	ms.Toggle(3)
	ms.Toggle(5)
	ms.SetCursor(2)
	ms.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 12})
	buf := buffer.NewBuffer(30, 12)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ms.Paint(buf)
	}
}
