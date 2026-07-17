package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P258: CodeBlock Measure clamp + highlight fallback + Viewport visible scrollbars

func TestCodeBlock_Measure_ClampSmall_P258(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc main() {}")
	s := cb.Measure(Constraints{MaxWidth: 3, MaxHeight: 2})
	if s.W > 3 {
		t.Errorf("width should be clamped to 3, got %d", s.W)
	}
	if s.H > 2 {
		t.Errorf("height should be clamped to 2, got %d", s.H)
	}
}

func TestCodeBlock_Measure_Tiny_P258(t *testing.T) {
	cb := NewCodeBlock("go", "x")
	s := cb.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W < 1 || s.H < 1 {
		t.Error("width and height should be at least 1")
	}
}

func TestCodeBlock_Paint_NegativeCodeWidth_P258(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 3}) // very narrow
	buf := buffer.NewBuffer(2, 3)
	cb.Paint(buf)
}

func TestCodeBlock_Paint_UnknownLanguage_P258(t *testing.T) {
	// Unknown language → highlighter returns error → plain text fallback
	cb := NewCodeBlock("xyz_unknown_lang", "some code here")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	cb.Paint(buf)
}

func TestViewport_DrawVScrollBar_ThumbClamp_P258(t *testing.T) {
	// Content much taller than viewport with scroll offset → thumb clamped to bottom
	child := &tallChild{}
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	v.ScrollToBottom()
	buf := buffer.NewBuffer(20, 5)
	v.Paint(buf)
}

func TestViewport_DrawHScrollBar_ThumbClamp_P258(t *testing.T) {
	child := &wideChild{}
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	v.ScrollToRight()
	buf := buffer.NewBuffer(5, 5)
	v.Paint(buf)
}
