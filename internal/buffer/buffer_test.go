package buffer

import "testing"

func TestBufferSetGet(t *testing.T) {
	b := NewBuffer(10, 5)
	b.SetCell(3, 2, Cell{Rune: 'X', Width: 1})

	cell := b.GetCell(3, 2)
	if cell.Rune != 'X' {
		t.Errorf("got rune %c, want X", cell.Rune)
	}

	// Out of bounds
	cell = b.GetCell(100, 100)
	if cell.Rune != ' ' {
		t.Errorf("out-of-bounds should return blank")
	}
}

func TestBufferDrawText(t *testing.T) {
	b := NewBuffer(20, 5)
	style := Style{}.WithFg(Red)
	end := b.DrawText(0, 0, "hello", style)

	if end != 5 {
		t.Errorf("end x: got %d, want 5", end)
	}

	for i, want := range "hello" {
		cell := b.GetCell(i, 0)
		if cell.Rune != want {
			t.Errorf("cell %d: got %c, want %c", i, cell.Rune, want)
		}
		if !cell.Fg.Equal(Red) {
			t.Errorf("cell %d: expected red fg", i)
		}
	}
}

func TestBufferFill(t *testing.T) {
	b := NewBuffer(5, 3)
	fill := Cell{Rune: '#', Width: 1, Fg: Green}
	b.Fill(fill)

	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			cell := b.GetCell(x, y)
			if cell.Rune != '#' {
				t.Errorf("(%d,%d): got %c, want #", x, y, cell.Rune)
			}
		}
	}
}

func TestBufferDiff(t *testing.T) {
	front := NewBuffer(5, 3)
	back := NewBuffer(5, 3)

	// Change one cell in back
	back.SetCell(2, 1, Cell{Rune: 'X', Width: 1})

	ops := Diff(front, back)
	if len(ops) != 1 {
		t.Fatalf("expected 1 diff op, got %d", len(ops))
	}
	if ops[0].X != 2 || ops[0].Y != 1 {
		t.Errorf("diff at (%d,%d), want (2,1)", ops[0].X, ops[0].Y)
	}
}

func TestBufferDiffSkipsRows(t *testing.T) {
	front := NewBuffer(10, 5)
	back := NewBuffer(10, 5)

	// Only change row 4
	back.SetCell(0, 4, Cell{Rune: 'Z', Width: 1})

	ops := Diff(front, back)
	if len(ops) != 1 {
		t.Fatalf("expected 1 diff op (row skip), got %d", len(ops))
	}
}

func TestBufferDiffNoChanges(t *testing.T) {
	a := NewBuffer(10, 5)
	b := NewBuffer(10, 5)

	ops := Diff(a, b)
	if len(ops) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(ops))
	}
}

func TestColorHex(t *testing.T) {
	c := Hex("#ff6600")
	if c.Type != ColorTrue {
		t.Errorf("expected ColorTrue, got %v", c.Type)
	}
	if c.R() != 255 || c.G() != 102 || c.B() != 0 {
		t.Errorf("RGB: got (%d,%d,%d), want (255,102,0)", c.R(), c.G(), c.B())
	}
}

func TestColorEqual(t *testing.T) {
	a := RGB(255, 128, 0)
	b := RGB(255, 128, 0)
	c := RGB(255, 128, 1)

	if !a.Equal(b) {
		t.Error("identical colors should be equal")
	}
	if a.Equal(c) {
		t.Error("different colors should not be equal")
	}
}

func TestStyleSGR(t *testing.T) {
	s := Style{}.
		WithFg(RGB(255, 0, 0)).
		WithBg(RGB(0, 0, 255)).
		AddFlags(Bold | Underline)

	seq := s.SGRSequence()
	if !contains(seq, "1") { // Bold
		t.Errorf("SGR should contain Bold (1): %s", seq)
	}
	if !contains(seq, "4") { // Underline
		t.Errorf("SGR should contain Underline (4): %s", seq)
	}
	if !contains(seq, "38;2;255;0;0") { // TrueColor FG
		t.Errorf("SGR should contain FG: %s", seq)
	}
	if !contains(seq, "48;2;0;0;255") { // TrueColor BG
		t.Errorf("SGR should contain BG: %s", seq)
	}
}

func TestCellEqual(t *testing.T) {
	a := Cell{Rune: 'a', Width: 1, Fg: Red}
	b := Cell{Rune: 'a', Width: 1, Fg: Red}
	c := Cell{Rune: 'b', Width: 1, Fg: Red}

	if !a.Equal(b) {
		t.Error("identical cells should be equal")
	}
	if a.Equal(c) {
		t.Error("different runes should not be equal")
	}
}

func TestRuneWidth(t *testing.T) {
	tests := []struct {
		name string
		r    rune
		want int
	}{
		{"ASCII", 'a', 1},
		{"CJK", '你', 2},
		{"Katakana", 'ア', 2},
		{"Emoji", '\U0001f600', 2},
		{"Combining", '\u0301', 0},
		{"Space", ' ', 1},
		{"Digit", '5', 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RuneWidth(tt.r)
			if got != tt.want {
				t.Errorf("RuneWidth(%U): got %d, want %d", tt.r, got, tt.want)
			}
		})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
