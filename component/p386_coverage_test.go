package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P386: Coverage for avatar Measure edge cases + initials override path + textarea stubs

func TestP386_Avatar_Measure_TinyHeight(t *testing.T) {
	a := NewAvatar("Alice")
	// h > cs.MaxHeight path (cs.HasHeight && h > MaxHeight)
	s := a.Measure(Constraints{MaxWidth: 80, MaxHeight: 0})
	if s.H != 1 {
		t.Errorf("H = %d, want 1 (clamped from 0)", s.H)
	}
}

func TestP386_Avatar_Measure_NegativeWidth(t *testing.T) {
	a := NewAvatar("Alice")
	a.SetSize(AvatarSmall)
	// w < 1 path — constraints with negative max width
	s := a.Measure(Constraints{MaxWidth: -5, MaxHeight: 5})
	if s.W != 1 {
		t.Errorf("W = %d, want 1 (clamped)", s.W)
	}
}

func TestP386_Avatar_Measure_NegativeHeight(t *testing.T) {
	a := NewAvatar("Alice")
	// h < 1 path
	s := a.Measure(Constraints{MaxWidth: 5, MaxHeight: -1})
	if s.H != 1 {
		t.Errorf("H = %d, want 1 (clamped)", s.H)
	}
}

func TestP386_Avatar_Paint_InitialsOverride(t *testing.T) {
	// Set initials override → exercises the initials loop in Paint
	a := NewAvatar("Alice Brown")
	a.SetInitials("XY")
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'X' {
		t.Errorf("cell[0] = %q, want 'X'", string(c.Rune))
	}
}

func TestP386_Avatar_Paint_InitialsLowercase(t *testing.T) {
	// Lowercase initials override → exercises uppercase conversion
	a := NewAvatar("Alice")
	a.SetInitials("ab")
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'A' {
		t.Errorf("cell[0] = %q, want 'A' (uppercased)", string(c.Rune))
	}
}

func TestP386_Avatar_Paint_InitialsNoLetters(t *testing.T) {
	// Initials with only non-letter chars → initN==0 → fallback '?'
	a := NewAvatar("Alice")
	a.SetInitials("123")
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '?' {
		t.Errorf("cell[0] = %q, want '?' (fallback)", string(c.Rune))
	}
}

func TestP386_Avatar_Paint_InitialsSpecialChars(t *testing.T) {
	// Mixed: special chars + letters → only letters extracted
	a := NewAvatar("Alice")
	a.SetInitials("#X")
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'X' {
		t.Errorf("cell[0] = %q, want 'X'", string(c.Rune))
	}
}

// TextArea stubs (bubbles-compatible no-ops)

func TestP386_TextArea_SetPrompt(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("$ ") // no-op, should not panic
	if ta.Prompt() != "" {
		t.Errorf("Prompt = %q, want empty (no-op stub)", ta.Prompt())
	}
}

func TestP386_TextArea_SetPlaceholder(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("Enter text...") // no-op
	if ta.Placeholder() != "" {
		t.Errorf("Placeholder = %q, want empty (no-op stub)", ta.Placeholder())
	}
}

func TestP386_TextArea_FocusBlur(t *testing.T) {
	ta := NewTextArea()
	ta.Focus() // no-op
	ta.Blur()  // no-op
	if ta.Blink() {
		t.Error("Blink should return false")
	}
}

func TestP386_TextArea_SetCharLimit(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100) // no-op
	// Just verify it doesn't panic
}

// DiffPreview stubs

func TestP386_DiffPreview_Stubs(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(false)
	if !dp.ShowLineNumbers() {
		t.Error("ShowLineNumbers should always return true (stub)")
	}
	dp.SetShowStats(true)  // no-op
	dp.SetShowStats(false) // no-op
}

// DiffStatBar Measure edge cases

func TestP386_DiffStatBar_Measure_TextStyle(t *testing.T) {
	d := NewDiffStatBar(12, 5)
	d.SetStyle(DiffStatStyleText)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if s.H != 1 {
		t.Errorf("Text style H = %d, want 1", s.H)
	}
}

func TestP386_DiffStatBar_Measure_SingleDigit(t *testing.T) {
	d := NewDiffStatBar(5, 3)
	d.SetStyle(DiffStatStyleText)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// "+5 -3" = 5 chars: '+' '5' ' ' '-' '3'
	if s.W != 5 {
		t.Errorf("W = %d, want 5", s.W)
	}
}
