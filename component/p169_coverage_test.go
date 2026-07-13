package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Coverage tests for new components (P169) ───

func TestContentSwitcher_CurrentComponent(t *testing.T) {
	cs := NewContentSwitcher()
	f := NewFill('A', buffer.Style{})
	cs.Add("a", f)

	if cs.CurrentComponent() != f {
		t.Error("CurrentComponent should return the active child")
	}

	cs.SetCurrent("nonexistent")
	// Still returns the old current
	if cs.CurrentComponent() != f {
		t.Error("should still return a after failed SetCurrent")
	}
}

func TestContentSwitcher_Measure(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	s := cs.Measure(Bounded(30, 20))
	// Fill returns MaxWidth/MaxHeight
	if s.W != 30 || s.H != 20 {
		t.Errorf("expected 30x20, got %dx%d", s.W, s.H)
	}
}

func TestContentSwitcher_Children(t *testing.T) {
	cs := NewContentSwitcher()
	f1 := NewFill('A', buffer.Style{})
	f2 := NewFill('B', buffer.Style{})
	cs.Add("a", f1)
	cs.Add("b", f2)

	children := cs.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestContentSwitcher_EmptyMeasure(t *testing.T) {
	cs := NewContentSwitcher()
	s := cs.Measure(Bounded(30, 20))
	if s.W != 0 || s.H != 0 {
		t.Error("empty switcher should measure 0x0")
	}
}

func TestContentSwitcher_PaintEmpty(t *testing.T) {
	cs := NewContentSwitcher()
	cs.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	cs.Paint(buf) // should not panic with no children
}

func TestContentSwitcher_NextEmpty(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Next() // should not panic
}

func TestContentSwitcher_PrevEmpty(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Prev() // should not panic
}

func TestContentSwitcher_NextWrap(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	cs.SetCurrent("a")
	cs.Next() // single item, should wrap to "a"
	if cs.Current() != "a" {
		t.Error("single item Next should stay on same")
	}
}

func TestFill_SetChar(t *testing.T) {
	f := NewFill('A', buffer.Style{})
	f.SetChar('B')
	f.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	f.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'B' {
		t.Error("SetChar should change fill char")
	}
}

func TestFill_SetStyle(t *testing.T) {
	f := NewFill(' ', buffer.Style{})
	newStyle := buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	f.SetStyle(newStyle)
	f.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	f.Paint(buf)
	if buf.GetCell(0, 0).Fg.Val != newStyle.Fg.Val {
		t.Error("SetStyle should change fill style")
	}
}

func TestFill_ZeroBounds(t *testing.T) {
	f := NewFill(' ', buffer.Style{})
	f.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	f.Paint(buf) // should not panic
}

func TestClear_Measure(t *testing.T) {
	c := NewClear()
	s := c.Measure(Bounded(20, 10))
	if s.W != 20 || s.H != 10 {
		t.Error("Clear should fill bounds")
	}
}

func TestClear_ZeroBounds(t *testing.T) {
	c := NewClear()
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(5, 5)
	c.Paint(buf) // should not panic
}

func TestParagraph_SetWrap(t *testing.T) {
	p := NewParagraph("test")
	p.SetWrap(false)
	p.SetWrap(true)
	// just verify it doesn't crash
}

func TestParagraph_SetFg(t *testing.T) {
	p := NewParagraph("test")
	p.SetFg(buffer.RGB(255, 0, 0))
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
	if buf.GetCell(0, 0).Fg.Val != buffer.RGB(255, 0, 0).Val {
		t.Error("SetFg should set foreground")
	}
}

func TestParagraph_SetBg(t *testing.T) {
	p := NewParagraph("test")
	p.SetBg(buffer.RGB(0, 0, 255))
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
	// Check bg fill on space cell
	if buf.GetCell(4, 0).Bg.Val != buffer.RGB(0, 0, 255).Val {
		t.Error("SetBg should fill background")
	}
}

func TestParagraph_Measure(t *testing.T) {
	p := NewParagraph("Hello\nWorld")
	s := p.Measure(Bounded(50, 10))
	if s.H < 2 {
		t.Errorf("expected at least 2 lines, got %d", s.H)
	}
}

func TestParagraph_MeasureEmpty(t *testing.T) {
	p := NewParagraph("")
	s := p.Measure(Bounded(50, 10))
	// Empty text should still produce at least 1 line height
	if s.H < 1 {
		t.Errorf("expected at least 1, got %d", s.H)
	}
}

func TestParagraph_PaintZeroBounds(t *testing.T) {
	p := NewParagraph("test")
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	p.Paint(buf) // should not panic
}

func TestParagraph_AlignRight(t *testing.T) {
	p := NewParagraph("Hi")
	p.SetAlign(TextAlignRight)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
	// "Hi" right-aligned in width 10: starts at x=8
	if buf.GetCell(8, 0).Rune != 'H' {
		t.Errorf("expected H at x=8, got %q", buf.GetCell(8, 0).Rune)
	}
}

func TestParagraph_LongWordWrap(t *testing.T) {
	// Word longer than width should hard-break
	p := NewParagraph("abcdefghijklmnopqrstuvwxyz")
	p.SetWrap(true)
	p.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	p.Paint(buf) // should not panic, should hard-break
}

func TestParagraph_MultilineWrap(t *testing.T) {
	p := NewParagraph("hello world\nfoo bar baz")
	p.SetWrap(true)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	p.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'h' {
		t.Error("should start with 'h'")
	}
}

func TestRule_SetOrientation(t *testing.T) {
	r := NewRule()
	r.SetOrientation(VerticalRule)
	r.SetChar('│')
	r.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	buf := buffer.NewBuffer(1, 5)
	r.Paint(buf)
	for y := 0; y < 5; y++ {
		if buf.GetCell(0, y).Rune != '│' {
			t.Errorf("expected │ at y=%d", y)
		}
	}
}

func TestRule_SetStyle(t *testing.T) {
	r := NewRule()
	r.SetStyle(buffer.Style{Fg: buffer.RGB(255, 0, 0)})
	r.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	r.Paint(buf)
	if buf.GetCell(0, 0).Fg.Val != buffer.RGB(255, 0, 0).Val {
		t.Error("SetStyle should set color")
	}
}

func TestSwitch_SetStyle(t *testing.T) {
	s := NewSwitch("Test")
	style := SwitchStyle{
		OnBg:   buffer.RGB(0, 255, 0),
		OffBg:  buffer.RGB(60, 60, 60),
		KnobFg: buffer.NamedColor(buffer.NamedWhite),
		LabelFg: buffer.NamedColor(buffer.NamedYellow),
		DisabledFg: buffer.RGB(100, 100, 100),
	}
	s.SetStyle(style)
	s.SetOn(true)
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf)
	// Should render with OnBg
	if buf.GetCell(6, 0).Bg.Val != buffer.RGB(0, 255, 0).Val {
		t.Error("should use OnBg when on")
	}
}

func TestSwitch_HandleKey(t *testing.T) {
	s := NewSwitch("")
	if s.HandleKey(nil) {
		t.Error("nil key should not be consumed")
	}
}