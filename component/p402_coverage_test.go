package component

import (
	"testing"
	"github.com/topcheer/fluui/internal/buffer"
)

// drawWrappedText coverage
func TestP402_DrawWrappedText_Basic(t *testing.T) {
	buf := buffer.NewBuffer(10, 3)
	drawWrappedText(buf, Rect{X: 0, Y: 0, W: 10, H: 3}, "Hello World", buffer.DefaultStyle)
}

func TestP402_DrawWrappedText_Multiline(t *testing.T) {
	buf := buffer.NewBuffer(5, 5)
	drawWrappedText(buf, Rect{X: 0, Y: 0, W: 5, H: 5}, "ab\ncd\nef", buffer.DefaultStyle)
}

func TestP402_DrawWrappedText_Empty(t *testing.T) {
	buf := buffer.NewBuffer(5, 3)
	drawWrappedText(buf, Rect{X: 0, Y: 0, W: 5, H: 3}, "", buffer.DefaultStyle)
}

func TestP402_DrawWrappedText_Wrap(t *testing.T) {
	buf := buffer.NewBuffer(3, 4)
	drawWrappedText(buf, Rect{X: 0, Y: 0, W: 3, H: 4}, "abcdef", buffer.DefaultStyle)
}

// Measure edge cases
func TestP402_Measure_Edges(t *testing.T) {
	tests := []struct {
		name string
		fn   func() Size
	}{
		{"SearchBar", func() Size { s := NewSearchBar("x"); return s.Measure(Constraints{}) }},
		{"StatCard", func() Size { return NewStatCard("X", "1").Measure(Constraints{}) }},
		{"Toast", func() Size { return NewToast("x", ToastInfo).Measure(Constraints{}) }},
		{"HintLabel", func() Size { return NewHintLabel("x").Measure(Constraints{}) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fn()
			if s.W < 1 { t.Error("W should be >= 1") }
			if s.H < 1 { t.Error("H should be >= 1") }
		})
	}
}

// MarkdownStream drawWrappedText via empty source
func TestP402_MDStream_DrawWrappedText(t *testing.T) {
	m := NewMarkdownStream()
	m.SetSource("")
	m.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	m.Paint(buf)
}
