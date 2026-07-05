package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === P35: Enhanced Border Tests ===

func TestP35_BorderRounded_Corners(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.Type = BorderRounded
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 4})

	buf := buffer.NewBuffer(6, 4)
	border.Paint(buf)

	if tl := buf.GetCell(0, 0); tl.Rune != '\u256d' {
		t.Errorf("rounded top-left: got %q, want \u256d", string(tl.Rune))
	}
	if tr := buf.GetCell(5, 0); tr.Rune != '\u256e' {
		t.Errorf("rounded top-right: got %q, want \u256e", string(tr.Rune))
	}
	if bl := buf.GetCell(0, 3); bl.Rune != '\u2570' {
		t.Errorf("rounded bottom-left: got %q, want \u2570", string(bl.Rune))
	}
	if br := buf.GetCell(5, 3); br.Rune != '\u256f' {
		t.Errorf("rounded bottom-right: got %q, want \u256f", string(br.Rune))
	}
}

func TestP35_BorderDouble_Corners(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.Type = BorderDouble
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 4})

	buf := buffer.NewBuffer(6, 4)
	border.Paint(buf)

	if tl := buf.GetCell(0, 0); tl.Rune != '\u2554' {
		t.Errorf("double top-left: got %q, want \u2554", string(tl.Rune))
	}
	if tr := buf.GetCell(5, 0); tr.Rune != '\u2557' {
		t.Errorf("double top-right: got %q, want \u2557", string(tr.Rune))
	}
	if h := buf.GetCell(2, 0); h.Rune != '\u2550' {
		t.Errorf("double horizontal: got %q, want \u2550", string(h.Rune))
	}
	if v := buf.GetCell(0, 1); v.Rune != '\u2551' {
		t.Errorf("double vertical: got %q, want \u2551", string(v.Rune))
	}
}

func TestP35_BorderHeavy_Corners(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.Type = BorderHeavy
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 4})

	buf := buffer.NewBuffer(6, 4)
	border.Paint(buf)

	if tl := buf.GetCell(0, 0); tl.Rune != '\u250f' {
		t.Errorf("heavy top-left: got %q, want \u250f", string(tl.Rune))
	}
	if h := buf.GetCell(2, 0); h.Rune != '\u2501' {
		t.Errorf("heavy horizontal: got %q, want \u2501", string(h.Rune))
	}
	if v := buf.GetCell(0, 1); v.Rune != '\u2503' {
		t.Errorf("heavy vertical: got %q, want \u2503", string(v.Rune))
	}
}

func TestP35_BorderNone_NoBorderChars(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.Type = BorderNone
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 4})

	buf := buffer.NewBuffer(6, 4)
	border.Paint(buf)

	// Corners should be spaces, not box-drawing chars
	for _, pos := range [][2]int{{0, 0}, {5, 0}, {0, 3}, {5, 3}} {
		c := buf.GetCell(pos[0], pos[1])
		if c.Rune != ' ' && c.Rune != 0 {
			t.Errorf("border-none corner [%d,%d]: got %q, want space", pos[0], pos[1], string(c.Rune))
		}
	}
}

func TestP35_BorderTitleLeft(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Title = "Title"
	border.TitleAlign = TitleLeft
	border.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})

	buf := buffer.NewBuffer(10, 3)
	border.Paint(buf)

	// Title should start at position 1 (after left border)
	for i, r := range "Title" {
		c := buf.GetCell(1+i, 0)
		if c.Rune != r {
			t.Errorf("left title [%d,0]: got %q, want %q", 1+i, string(c.Rune), string(r))
		}
	}
}

func TestP35_BorderTitleRight(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Title = "Title"
	border.TitleAlign = TitleRight
	border.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})

	buf := buffer.NewBuffer(10, 3)
	border.Paint(buf)

	// "Title" is 5 chars; right-aligned on width 10: starts at 10-1-5 = 4
	for i, r := range "Title" {
		c := buf.GetCell(4+i, 0)
		if c.Rune != r {
			t.Errorf("right title [%d,0]: got %q, want %q", 4+i, string(c.Rune), string(r))
		}
	}
}

func TestP35_BorderTitleCentered(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Title = "ABC"
	// Default TitleAlign is TitleCenter
	border.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})

	buf := buffer.NewBuffer(10, 3)
	border.Paint(buf)

	// Centered: (10-3)/2 = 3, so starts at 3
	for i, r := range "ABC" {
		c := buf.GetCell(3+i, 0)
		if c.Rune != r {
			t.Errorf("center title [%d,0]: got %q, want %q", 3+i, string(c.Rune), string(r))
		}
	}
}

func TestP35_BorderPartialSides_TopOnly(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Sides = BorderTop
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 3})

	buf := buffer.NewBuffer(6, 3)
	border.Paint(buf)

	// Top should have horizontal edge
	if h := buf.GetCell(2, 0); h.Rune != '\u2500' {
		t.Errorf("top side: got %q, want horizontal edge", string(h.Rune))
	}
	// Left/right verticals should NOT be drawn
	if v := buf.GetCell(0, 1); v.Rune == '\u2502' {
		t.Error("left side should not be drawn with BorderTop only")
	}
	if v := buf.GetCell(5, 1); v.Rune == '\u2502' {
		t.Error("right side should not be drawn with BorderTop only")
	}
}

func TestP35_BorderPartialSides_BottomOnly(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Sides = BorderBottom
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 3})

	buf := buffer.NewBuffer(6, 3)
	border.Paint(buf)

	// Bottom should have horizontal edge
	if h := buf.GetCell(2, 2); h.Rune != '\u2500' {
		t.Errorf("bottom side: got %q, want horizontal edge", string(h.Rune))
	}
	// Top should NOT be drawn
	if h := buf.GetCell(2, 0); h.Rune == '\u2500' {
		t.Error("top side should not be drawn with BorderBottom only")
	}
}

func TestP35_BorderPartialSides_LeftRightOnly(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Sides = BorderLeft | BorderRight
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 3})

	buf := buffer.NewBuffer(6, 3)
	border.Paint(buf)

	// Left/right verticals should be drawn
	if v := buf.GetCell(0, 1); v.Rune != '\u2502' {
		t.Errorf("left side: got %q, want vertical", string(v.Rune))
	}
	if v := buf.GetCell(5, 1); v.Rune != '\u2502' {
		t.Errorf("right side: got %q, want vertical", string(v.Rune))
	}
	// Top should NOT be drawn
	if h := buf.GetCell(2, 0); h.Rune == '\u2500' {
		t.Error("top should not be drawn with Left|Right only")
	}
}

func TestP35_BorderPadding(t *testing.T) {
	child := NewText("Hi")
	border := NewBorder(child)
	border.InnerPadding = UniformPadding(1)
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 5})

	buf := buffer.NewBuffer(6, 5)
	border.Paint(buf)

	// With padding(1): border(1) + pad(1) = 2 offset from each edge
	// Child at [2,2]
	if inner := buf.GetCell(2, 2); inner.Rune != 'H' {
		t.Errorf("inner with padding [2,2]: got %q, want 'H'", string(inner.Rune))
	}
}

func TestP35_BorderPaddingMeasure(t *testing.T) {
	child := NewText("Hello") // width=5
	border := NewBorder(child)
	border.InnerPadding = UniformPadding(1)

	size := border.Measure(Unbounded())
	// border(2) + padding(2) + child(5) = 9 wide
	// border(2) + padding(2) + child(1) = 5 tall
	if size.W != 9 {
		t.Errorf("W with padding: got %d, want 9", size.W)
	}
	if size.H != 5 {
		t.Errorf("H with padding: got %d, want 5", size.H)
	}
}

func TestP35_BorderAllSidesDefault(t *testing.T) {
	border := NewBorder(NewText("X"))
	if border.Sides != BorderAll {
		t.Errorf("default Sides: got %d, want %d", border.Sides, BorderAll)
	}
}

func TestP35_BorderTypeDefault(t *testing.T) {
	border := NewBorder(NewText("X"))
	if border.Type != BorderSingle {
		t.Errorf("default Type: got %d, want %d", border.Type, BorderSingle)
	}
}

func TestP35_BorderTitleAlignDefault(t *testing.T) {
	border := NewBorder(NewText("X"))
	if border.TitleAlign != TitleCenter {
		t.Errorf("default TitleAlign: got %d, want %d", border.TitleAlign, TitleCenter)
	}
}

func TestP35_PaddingHelpers(t *testing.T) {
	p := UniformPadding(2)
	if p.Horizontal() != 4 {
		t.Errorf("UniformPadding(2).Horizontal(): got %d, want 4", p.Horizontal())
	}
	if p.Vertical() != 4 {
		t.Errorf("UniformPadding(2).Vertical(): got %d, want 4", p.Vertical())
	}

	z := NoPadding()
	if z.Horizontal() != 0 || z.Vertical() != 0 {
		t.Errorf("NoPadding: got H=%d V=%d, want 0,0", z.Horizontal(), z.Vertical())
	}
}

func TestP35_BorderRoundedMeasure(t *testing.T) {
	child := NewText("Hello")
	border := NewBorder(child)
	border.Type = BorderRounded
	size := border.Measure(Unbounded())
	// Same as single: border doesn't change measurement, only rune style
	if size.W != 7 {
		t.Errorf("rounded W: got %d, want 7", size.W)
	}
	if size.H != 3 {
		t.Errorf("rounded H: got %d, want 3", size.H)
	}
}

func TestP35_BorderNoSidesMeasure(t *testing.T) {
	child := NewText("Hello")
	border := NewBorder(child)
	border.Sides = 0 // no sides at all
	size := border.Measure(Unbounded())
	// No border columns: child width + 0 = 5, height + 0 = 1
	if size.W != 5 {
		t.Errorf("no-sides W: got %d, want 5", size.W)
	}
	if size.H != 1 {
		t.Errorf("no-sides H: got %d, want 1", size.H)
	}
}

func TestP35_BorderTitleLeftWithPadding(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Title = "T"
	border.TitleAlign = TitleLeft
	border.InnerPadding = Padding{Left: 2}
	border.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4})

	buf := buffer.NewBuffer(10, 4)
	border.Paint(buf)

	// Title left with pad(2): starts at x + 1 + 2 = 3
	c := buf.GetCell(3, 0)
	if c.Rune != 'T' {
		t.Errorf("left title with pad: got %q at [3,0], want 'T'", string(c.Rune))
	}
}

func TestP35_BorderBottomRightOnly(t *testing.T) {
	child := NewText("X")
	border := NewBorder(child)
	border.Sides = BorderBottom | BorderRight
	border.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 3})

	buf := buffer.NewBuffer(6, 3)
	border.Paint(buf)

	// Bottom horizontal should be drawn
	if h := buf.GetCell(2, 2); h.Rune != '\u2500' {
		t.Errorf("bottom with bottom|right: got %q, want horizontal", string(h.Rune))
	}
	// Right vertical should be drawn
	if v := buf.GetCell(5, 1); v.Rune != '\u2502' {
		t.Errorf("right with bottom|right: got %q, want vertical", string(v.Rune))
	}
	// Top should NOT be drawn
	if h := buf.GetCell(2, 0); h.Rune == '\u2500' {
		t.Error("top should not be drawn with bottom|right only")
	}
}

func TestP35_BorderAllTypesPaint(t *testing.T) {
	// Ensure all border types paint without panicking
	types := []BorderType{BorderSingle, BorderRounded, BorderDouble, BorderHeavy, BorderNone}
	for _, bt := range types {
		border := NewBorder(NewText("Test"))
		border.Type = bt
		border.Title = "X"
		border.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 4})
		buf := buffer.NewBuffer(8, 4)
		border.Paint(buf) // should not panic
	}
}

func TestP35_BorderConcurrentPaint(t *testing.T) {
	// Concurrent paint with different border types should not race
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			border := NewBorder(NewText("Concurrent"))
			border.Type = BorderRounded
			border.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
			buf := buffer.NewBuffer(15, 5)
			border.Paint(buf)
			done <- true
		}()
	}
	for i := 0; i < 5; i++ {
		<-done
	}
}
