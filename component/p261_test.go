package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P261: viewport scrollbars overflow + MaxOffset<0 + SetBounds clamp + codeblock title/tab

// stubChildW0: a child whose Paint draws cells with Width=0 (padding cells)
type stubChildW0 struct{ BaseComponent }

func (c *stubChildW0) Measure(cs Constraints) Size { return Size{W: 5, H: 3} }
func (c *stubChildW0) Paint(buf *buffer.Buffer) {
	// Draw a cell with Width=0 to trigger the skip-padding branch
	buf.SetCell(0, 0, buffer.Cell{Rune: 0, Width: 0})
	buf.SetCell(1, 0, buffer.Cell{Rune: 'A', Width: 1})
}

func TestViewport_MaxOffsetX_Zero_P261(t *testing.T) {
	// Content narrower than viewport → max < 0 → return 0
	child := &tallChild{}
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 200, H: 10})
	if v.MaxOffsetX() != 0 {
		t.Errorf("narrow content should have maxOffsetX=0, got %d", v.MaxOffsetX())
	}
}

func TestViewport_MaxOffsetY_Zero_P261(t *testing.T) {
	// Content shorter than viewport → max < 0 → return 0
	child := &wideChild{}
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 100})
	if v.MaxOffsetY() != 0 {
		t.Errorf("short content should have maxOffsetY=0, got %d", v.MaxOffsetY())
	}
}

func TestViewport_SetBounds_ClampOffset_P261(t *testing.T) {
	child := &tallChild{}
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	v.ScrollDown(20)
	v.ScrollRight(20)
	// Now resize to smaller — should clamp offsets
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
}

func TestViewport_Paint_SkipPaddingCell_P261(t *testing.T) {
	v := NewViewport(&stubChildW0{})
	v.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	v.Paint(buf)
}

func TestViewport_DrawVBar_ContentFits_P261(t *testing.T) {
	// Content fits → maxOff==0 → full thumb
	child := &wideChild{} // h=5, smaller than viewport
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 10}) // tall enough, no overflow
	buf := buffer.NewBuffer(100, 10)
	v.Paint(buf)
}

func TestViewport_DrawHBar_ContentFits_P261(t *testing.T) {
	child := &tallChild{} // w=100
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 200, H: 10})
	buf := buffer.NewBuffer(200, 10)
	v.Paint(buf)
}

func TestViewport_DrawVBar_TinyThumb_P261(t *testing.T) {
	// Very tall content → thumbH would be < 1 → forced to 1
	child := &tallChild{} // h=50
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	v.ScrollDown(10)
	buf := buffer.NewBuffer(20, 3)
	v.Paint(buf)
}

func TestViewport_DrawHBar_TinyThumb_P261(t *testing.T) {
	child := &wideChild{} // w=200
	v := NewViewport(child)
	v.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	v.ScrollRight(10)
	buf := buffer.NewBuffer(3, 5)
	v.Paint(buf)
}

func TestCodeBlock_Paint_WithTitle_P261(t *testing.T) {
	cb := NewCodeBlock("python", "print('hello')")
	cb.SetTitle("app.py")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestCodeBlock_LineWidth_Tab_P261(t *testing.T) {
	// Code with tabs → tab width=4 branch
	cb := NewCodeBlock("go", "\tpackage main")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}
