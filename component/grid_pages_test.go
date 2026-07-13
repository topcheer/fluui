package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Grid Tests ───

func TestGrid_Basic(t *testing.T) {
	g := NewGrid()
	g.SetRows(3, 0, 3)
	g.SetColumns(20, 0, 20)

	a := NewFill('A', buffer.Style{})
	b := NewFill('B', buffer.Style{})
	g.AddItem(a, 0, 0, 3, 1)
	g.AddItem(b, 0, 1, 1, 1)

	if g.ItemCount() != 2 {
		t.Errorf("expected 2 items, got %d", g.ItemCount())
	}
}

func TestGrid_RemoveItem(t *testing.T) {
	g := NewGrid()
	a := NewFill('A', buffer.Style{})
	g.AddItem(a, 0, 0, 1, 1)

	g.RemoveItem(a)
	if g.ItemCount() != 0 {
		t.Error("should have 0 items after remove")
	}
}

func TestGrid_Clear(t *testing.T) {
	g := NewGrid()
	g.AddItem(NewFill('A', buffer.Style{}), 0, 0, 1, 1)
	g.AddItem(NewFill('B', buffer.Style{}), 0, 1, 1, 1)
	g.Clear()

	if g.ItemCount() != 0 {
		t.Error("should be empty after clear")
	}
}

func TestGrid_Items(t *testing.T) {
	g := NewGrid()
	a := NewFill('A', buffer.Style{})
	g.AddItem(a, 0, 0, 1, 1)

	items := g.Items()
	if len(items) != 1 || items[0].Component != a {
		t.Error("Items should return copy with correct component")
	}
}

func TestGrid_Measure(t *testing.T) {
	g := NewGrid()
	s := g.Measure(Bounded(50, 20))
	if s.W != 50 || s.H != 20 {
		t.Errorf("expected 50x20, got %dx%d", s.W, s.H)
	}
}

func TestGrid_Paint(t *testing.T) {
	g := NewGrid()
	g.SetRows(5, 5)
	g.SetColumns(10, 10)
	a := NewFill('A', buffer.Style{})
	b := NewFill('B', buffer.Style{})
	g.AddItem(a, 0, 0, 1, 1)
	g.AddItem(b, 0, 1, 1, 1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	g.Paint(buf)

	if buf.GetCell(0, 0).Rune != 'A' {
		t.Error("should paint A in first cell")
	}
	if buf.GetCell(10, 0).Rune != 'B' {
		t.Error("should paint B in second column")
	}
}

func TestGrid_PaintWithSpan(t *testing.T) {
	g := NewGrid()
	g.SetRows(5, 5)
	g.SetColumns(10, 10)
	a := NewFill('A', buffer.Style{})
	g.AddItem(a, 0, 0, 2, 2) // span all
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	buf := buffer.NewBuffer(20, 10)
	g.Paint(buf)

	// Should fill entire area with A
	if buf.GetCell(0, 0).Rune != 'A' || buf.GetCell(15, 5).Rune != 'A' {
		t.Error("spanning item should fill entire area")
	}
}

func TestGrid_PaintZeroBounds(t *testing.T) {
	g := NewGrid()
	g.AddItem(NewFill('A', buffer.Style{}), 0, 0, 1, 1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	g.Paint(buf) // should not panic
}

func TestGrid_Children(t *testing.T) {
	g := NewGrid()
	a := NewFill('A', buffer.Style{})
	b := NewFill('B', buffer.Style{})
	g.AddItem(a, 0, 0, 1, 1)
	g.AddItem(b, 0, 1, 1, 1)

	children := g.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestGrid_AutoDetectSize(t *testing.T) {
	g := NewGrid()
	// No rows/cols set — should auto-detect from items
	g.AddItem(NewFill('A', buffer.Style{}), 0, 0, 1, 1)
	g.AddItem(NewFill('B', buffer.Style{}), 0, 1, 1, 1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	g.Paint(buf) // should not panic with auto-detected sizes
}

func TestGrid_SetGaps(t *testing.T) {
	g := NewGrid()
	g.SetRowGap(1)
	g.SetColGap(1)
	g.SetRows(3, 3)
	g.SetColumns(5, 5)
	g.AddItem(NewFill('A', buffer.Style{}), 0, 0, 1, 1)
	g.AddItem(NewFill('B', buffer.Style{}), 1, 1, 1, 1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 8})
	buf := buffer.NewBuffer(12, 8)
	g.Paint(buf) // should handle gaps
}

func TestGrid_NilComponent(t *testing.T) {
	g := NewGrid()
	g.AddItem(nil, 0, 0, 1, 1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	g.Paint(buf) // should not panic
}

// ─── Pages Tests ───

func TestPages_Basic(t *testing.T) {
	p := NewPages()
	p.AddPage("home", NewFill('H', buffer.Style{}))
	p.AddPage("settings", NewFill('S', buffer.Style{}))

	if p.PageCount() != 2 {
		t.Error("should have 2 pages")
	}
	if p.CurrentPage() != "home" {
		t.Error("first added should be current")
	}
}

func TestPages_SwitchTo(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	p.AddPage("b", NewFill('B', buffer.Style{}))

	if !p.SwitchTo("b") {
		t.Error("SwitchTo should return true for existing")
	}
	if p.CurrentPage() != "b" {
		t.Error("should switch to b")
	}
	if p.SwitchTo("nonexistent") {
		t.Error("SwitchTo should return false for missing")
	}
}

func TestPages_HasPage(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))

	if !p.HasPage("a") {
		t.Error("should have page a")
	}
	if p.HasPage("b") {
		t.Error("should not have page b")
	}
}

func TestPages_RemovePage(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	p.AddPage("b", NewFill('B', buffer.Style{}))

	p.RemovePage("a")
	if p.HasPage("a") {
		t.Error("should not have page a after remove")
	}
	if p.CurrentPage() != "b" {
		t.Error("should switch to b after removing current")
	}
}

func TestPages_NextPrev(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	p.AddPage("b", NewFill('B', buffer.Style{}))
	p.AddPage("c", NewFill('C', buffer.Style{}))

	p.NextPage()
	if p.CurrentPage() != "b" {
		t.Errorf("expected b, got %s", p.CurrentPage())
	}
	p.NextPage()
	if p.CurrentPage() != "c" {
		t.Errorf("expected c, got %s", p.CurrentPage())
	}
	p.NextPage() // wrap
	if p.CurrentPage() != "a" {
		t.Errorf("expected a, got %s", p.CurrentPage())
	}
	p.PrevPage() // wrap back
	if p.CurrentPage() != "c" {
		t.Errorf("expected c, got %s", p.CurrentPage())
	}
}

func TestPages_PageNames(t *testing.T) {
	p := NewPages()
	p.AddPage("x", NewFill('X', buffer.Style{}))
	p.AddPage("y", NewFill('Y', buffer.Style{}))

	names := p.PageNames()
	if len(names) != 2 || names[0] != "x" || names[1] != "y" {
		t.Errorf("expected [x y], got %v", names)
	}
}

func TestPages_Paint(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	p.AddPage("b", NewFill('B', buffer.Style{}))
	p.SwitchTo("b")
	p.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})

	buf := buffer.NewBuffer(5, 1)
	p.Paint(buf)

	if buf.GetCell(0, 0).Rune != 'B' {
		t.Error("should paint current page (b)")
	}
}

func TestPages_PaintEmpty(t *testing.T) {
	p := NewPages()
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	p.Paint(buf) // should not panic with no pages
}

func TestPages_Measure(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	s := p.Measure(Bounded(30, 20))
	if s.W != 30 || s.H != 20 {
		t.Error("should return child measure")
	}
}

func TestPages_Children(t *testing.T) {
	p := NewPages()
	a := NewFill('A', buffer.Style{})
	b := NewFill('B', buffer.Style{})
	p.AddPage("a", a)
	p.AddPage("b", b)

	children := p.Children()
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

func TestPages_NextPrevEmpty(t *testing.T) {
	p := NewPages()
	p.NextPage() // should not panic
	p.PrevPage()
}