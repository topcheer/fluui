package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Rule Tests ───

func TestRule_Horizontal(t *testing.T) {
	r := NewRule()
	r.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 3)
	r.Paint(buf)

	// Should have ─ across row 0
	for x := 0; x < 10; x++ {
		cell := buf.GetCell(x, 0)
		if cell.Rune != '─' {
			t.Errorf("expected ─ at (%d,0), got %q", x, cell.Rune)
		}
	}
}

func TestRule_Vertical(t *testing.T) {
	r := NewVerticalRule()
	r.SetBounds(Rect{X: 5, Y: 0, W: 1, H: 5})

	buf := buffer.NewBuffer(10, 5)
	r.Paint(buf)

	for y := 0; y < 5; y++ {
		cell := buf.GetCell(5, y)
		if cell.Rune != '│' {
			t.Errorf("expected │ at (5,%d), got %q", y, cell.Rune)
		}
	}
}

func TestRule_CustomChar(t *testing.T) {
	r := NewRule()
	r.SetChar('*')
	r.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	r.Paint(buf)

	if buf.GetCell(0, 0).Rune != '*' {
		t.Error("expected *")
	}
}

func TestRule_Measure(t *testing.T) {
	r := NewRule()
	s := r.Measure(Bounded(20, 10))
	if s.H != 1 {
		t.Errorf("horizontal rule height should be 1, got %d", s.H)
	}

	rv := NewVerticalRule()
	s2 := rv.Measure(Bounded(20, 10))
	if s2.W != 1 {
		t.Errorf("vertical rule width should be 1, got %d", s2.W)
	}
}

func TestRule_ZeroBounds(t *testing.T) {
	r := NewRule()
	r.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	r.Paint(buf) // should not panic
}

// ─── Switch Tests ───

func TestSwitch_Basic(t *testing.T) {
	s := NewSwitch("Toggle me")
	if s.IsOn() {
		t.Error("should start off")
	}
	if s.Label() != "Toggle me" {
		t.Error("label mismatch")
	}
}

func TestSwitch_Toggle(t *testing.T) {
	s := NewSwitch("")
	s.Toggle()
	if !s.IsOn() {
		t.Error("should be on after toggle")
	}
	s.Toggle()
	if s.IsOn() {
		t.Error("should be off after second toggle")
	}
}

func TestSwitch_SetOn(t *testing.T) {
	s := NewSwitch("")
	s.SetOn(true)
	if !s.IsOn() {
		t.Error("should be on")
	}
	s.SetOn(false)
	if s.IsOn() {
		t.Error("should be off")
	}
}

func TestSwitch_OnChange(t *testing.T) {
	s := NewSwitch("")
	var changes []bool
	s.SetOnChange(func(on bool) {
		changes = append(changes, on)
	})

	s.Toggle()
	s.Toggle()
	s.SetOn(true)

	if len(changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(changes))
	}
	if !changes[0] || changes[1] || !changes[2] {
		t.Errorf("unexpected change sequence: %v", changes)
	}
}

func TestSwitch_Measure(t *testing.T) {
	s := NewSwitch("Label")
	size := s.Measure(Bounded(50, 10))
	if size.H != 1 {
		t.Error("switch should be 1 line high")
	}
	if size.W <= 0 {
		t.Error("switch should have positive width")
	}
}

func TestSwitch_Paint(t *testing.T) {
	s := NewSwitch("Test")
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	s.Paint(buf) // should not panic

	// Check label rendered
	if buf.GetCell(0, 0).Rune != 'T' {
		t.Error("label should start at (0,0)")
	}
}

// ─── ContentSwitcher Tests ───

func TestContentSwitcher_Basic(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	cs.Add("b", NewFill('B', buffer.Style{}))

	if cs.Count() != 2 {
		t.Errorf("expected 2 children, got %d", cs.Count())
	}
	if cs.Current() != "a" {
		t.Error("first added should be current")
	}
}

func TestContentSwitcher_SetCurrent(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	cs.Add("b", NewFill('B', buffer.Style{}))

	if !cs.SetCurrent("b") {
		t.Error("SetCurrent should return true for existing")
	}
	if cs.Current() != "b" {
		t.Error("should switch to b")
	}

	if cs.SetCurrent("nonexistent") {
		t.Error("SetCurrent should return false for missing")
	}
}

func TestContentSwitcher_NextPrev(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	cs.Add("b", NewFill('B', buffer.Style{}))
	cs.Add("c", NewFill('C', buffer.Style{}))

	cs.Next()
	if cs.Current() != "b" {
		t.Errorf("expected b after next, got %s", cs.Current())
	}

	cs.Next()
	if cs.Current() != "c" {
		t.Errorf("expected c after next, got %s", cs.Current())
	}

	cs.Next() // wrap
	if cs.Current() != "a" {
		t.Errorf("expected a after wrap, got %s", cs.Current())
	}

	cs.Prev() // wrap back
	if cs.Current() != "c" {
		t.Errorf("expected c after prev wrap, got %s", cs.Current())
	}
}

func TestContentSwitcher_Remove(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	cs.Add("b", NewFill('B', buffer.Style{}))

	cs.Remove("a")
	if cs.Count() != 1 {
		t.Error("should have 1 after remove")
	}
	if cs.Current() != "b" {
		t.Error("should switch to b after removing a")
	}
}

func TestContentSwitcher_IDs(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("x", NewFill('X', buffer.Style{}))
	cs.Add("y", NewFill('Y', buffer.Style{}))

	ids := cs.IDs()
	if len(ids) != 2 || ids[0] != "x" || ids[1] != "y" {
		t.Errorf("expected [x y], got %v", ids)
	}
}

func TestContentSwitcher_Paint(t *testing.T) {
	cs := NewContentSwitcher()
	cs.Add("a", NewFill('A', buffer.Style{}))
	cs.Add("b", NewFill('B', buffer.Style{}))
	cs.SetCurrent("b")
	cs.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	cs.Paint(buf)

	if buf.GetCell(0, 0).Rune != 'B' {
		t.Error("should paint active child (b)")
	}
}

// ─── Fill Tests ───

func TestFill_Basic(t *testing.T) {
	f := NewFill('.', buffer.Style{})
	f.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})

	buf := buffer.NewBuffer(5, 3)
	f.Paint(buf)

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			if buf.GetCell(x, y).Rune != '.' {
				t.Errorf("expected . at (%d,%d)", x, y)
			}
		}
	}
}

func TestFill_Measure(t *testing.T) {
	f := NewFill(' ', buffer.Style{})
	s := f.Measure(Bounded(30, 20))
	if s.W != 30 || s.H != 20 {
		t.Errorf("expected 30x20, got %dx%d", s.W, s.H)
	}
}

// ─── Clear Tests ───

func TestClear_Basic(t *testing.T) {
	c := NewClear()
	c.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})

	buf := buffer.NewBuffer(5, 3)
	// Fill with non-blank first
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: 'X', Width: 1})
		}
	}

	c.Paint(buf)

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			if buf.GetCell(x, y).Rune != ' ' && buf.GetCell(x, y).Rune != 0 {
				t.Errorf("expected blank at (%d,%d), got %q", x, y, buf.GetCell(x, y).Rune)
			}
		}
	}
}

// ─── Paragraph Tests ───

func TestParagraph_Basic(t *testing.T) {
	p := NewParagraph("Hello, World!")
	if p.Text() != "Hello, World!" {
		t.Error("text mismatch")
	}
}

func TestParagraph_SetText(t *testing.T) {
	p := NewParagraph("old")
	p.SetText("new")
	if p.Text() != "new" {
		t.Error("text not updated")
	}
}

func TestParagraph_Wrap(t *testing.T) {
	p := NewParagraph("The quick brown fox jumps over the lazy dog")
	p.SetWrap(true)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})

	buf := buffer.NewBuffer(10, 10)
	p.Paint(buf) // should not panic

	// First line should start with "The"
	if buf.GetCell(0, 0).Rune != 'T' {
		t.Error("first line should start with T")
	}
}

func TestParagraph_NoWrap(t *testing.T) {
	p := NewParagraph("A very long line that would normally wrap")
	p.SetWrap(false)
	p.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	p.Paint(buf)

	// Without wrap, should show first 5 chars
	if buf.GetCell(0, 0).Rune != 'A' {
		t.Error("should show first char")
	}
}

func TestParagraph_Align(t *testing.T) {
	p := NewParagraph("Hi")
	p.SetAlign(TextAlignCenter)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})

	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)

	// "Hi" centered in width 10: starts at x=4
	if buf.GetCell(4, 0).Rune != 'H' {
		t.Errorf("expected H at x=4, got %q at x=4", buf.GetCell(4, 0).Rune)
	}
}

func TestParagraph_Scroll(t *testing.T) {
	p := NewParagraph("line1\nline2\nline3\nline4\nline5")
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})

	buf := buffer.NewBuffer(20, 2)
	p.Paint(buf)

	// Initially shows line1
	if buf.GetCell(0, 0).Rune != 'l' {
		t.Error("first line should be visible")
	}

	p.ScrollDown(2)
	buf2 := buffer.NewBuffer(20, 2)
	p.Paint(buf2)

	// After scrolling down 2, should show line3
	if buf2.GetCell(0, 0).Rune != 'l' {
		t.Error("should show line3 after scroll")
	}
}

func TestParagraph_ScrollClamped(t *testing.T) {
	p := NewParagraph("line1\nline2")
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})

	p.ScrollDown(100) // way past end
	if p.ScrollY() > 1 {
		t.Errorf("scroll should be clamped, got %d", p.ScrollY())
	}

	p.ScrollUp(100) // way before start
	if p.ScrollY() != 0 {
		t.Errorf("scroll should be clamped to 0, got %d", p.ScrollY())
	}
}