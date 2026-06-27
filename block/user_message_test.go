package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestUserMessageCreation(t *testing.T) {
	b := NewUserMessageBlock("user-1", "Hello AI")

	if b.ID() != "user-1" {
		t.Errorf("ID() = %q, want %q", b.ID(), "user-1")
	}
	if b.Type() != TypeUserMessage {
		t.Errorf("Type() = %v, want %v", b.Type(), TypeUserMessage)
	}
	if b.State() != BlockComplete {
		t.Errorf("State() = %v, want %v (user messages are immediately complete)", b.State(), BlockComplete)
	}
	if b.Content() != "Hello AI" {
		t.Errorf("Content() = %q, want %q", b.Content(), "Hello AI")
	}
}

func TestUserMessageMeasure(t *testing.T) {
	b := NewUserMessageBlock("user-2", "abcdefghijabcdefghij")
	// 20 chars at width 10 = 2 lines

	size := b.Measure(component.Bounded(10, 100))
	if size.H != 2 {
		t.Errorf("measure H = %d, want 2", size.H)
	}
}

func TestUserMessageMeasureWrap(t *testing.T) {
	b := NewUserMessageBlock("user-3", "abcdefghijklmnopqrstuvwxyz")

	size := b.Measure(component.Bounded(10, 100))
	if size.H < 2 {
		t.Errorf("wrapped measure H = %d, want >= 2", size.H)
	}
}

func TestUserMessagePaint(t *testing.T) {
	b := NewUserMessageBlock("user-4", "Hi")

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune != 'H' {
		t.Errorf("cell(0,0) rune = %q, want 'H'", cell.Rune)
	}

	// Verify cyan foreground color (0x8BE9FD)
	cyan := buffer.RGB(0x8B, 0xE9, 0xFD)
	if cell.Fg != cyan {
		t.Errorf("cell(0,0) Fg = %v, want cyan %v", cell.Fg, cyan)
	}
}

func TestUserMessagePaintMultiline(t *testing.T) {
	// Use a string that actually wraps at width 5: "ABCDEF" wraps to "ABCDE" + "F"
	b := NewUserMessageBlock("user-5", "ABCDEF")

	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 2})
	buf := buffer.NewBuffer(5, 2)
	b.Paint(buf)

	// Line 0: ABCDE
	if buf.GetCell(0, 0).Rune != 'A' {
		t.Errorf("cell(0,0) = %q, want 'A'", buf.GetCell(0, 0).Rune)
	}
	// Line 1: F
	if buf.GetCell(0, 1).Rune != 'F' {
		t.Errorf("cell(0,1) = %q, want 'F'", buf.GetCell(0, 1).Rune)
	}
}

func TestUserMessageEmpty(t *testing.T) {
	b := NewUserMessageBlock("user-6", "")

	// Measure should not panic
	size := b.Measure(component.Bounded(80, 100))
	if size.H != 1 {
		t.Errorf("empty measure H = %d, want 1", size.H)
	}

	// Paint should not panic
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.Paint(buf)

	// Buffer cells default to space — empty paint should leave default
	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("empty paint should leave blank, got %q", cell.Rune)
	}
}

func TestUserMessageDirtyAfterCreate(t *testing.T) {
	b := NewUserMessageBlock("user-7", "test")
	// NewBaseBlock sets dirty=true, and user message constructor doesn't clear it
	if !b.IsDirty() {
		t.Error("new user message should be dirty")
	}
	b.ClearDirty()
	if b.IsDirty() {
		t.Error("ClearDirty should clear dirty flag")
	}
}
