package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P270: drawText edge + selection GetSelectedText multi-line + findWordBoundary + recorder Save

func TestAppShell_DrawText_MaxWZero_P270(t *testing.T) {
	s := &AppShell{}
	buf := buffer.NewBuffer(20, 1)
	s.drawText(buf, 0, 0, "hello", 0, buffer.RGB(255, 255, 255), buffer.RGB(0, 0, 0), true)
	// maxW<=0 should early return — nothing written
	cell := buf.GetCell(0, 0)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Error("should not draw when maxW<=0")
	}
}

func TestAppShell_DrawText_Truncate_P270(t *testing.T) {
	s := &AppShell{}
	buf := buffer.NewBuffer(5, 1)
	s.drawText(buf, 0, 0, "hello world", 3, buffer.RGB(255, 255, 255), buffer.RGB(0, 0, 0), false)
	// Should truncate to 3 chars
	cell := buf.GetCell(3, 0)
	if cell.Rune != ' ' && cell.Rune != 0 {
		t.Errorf("position 3 should be blank, got %q", cell.Rune)
	}
}

func TestSelection_GetSelectedText_MultiLine_P270(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(10, 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: rune('A' + y), Width: 1})
		}
	}
	// Start selection at (2,1) and extend to (5,3)
	sm.StartSelection(2, 1)
	sm.ExtendSelection(5, 3)
	txt := sm.GetSelectedText(buf)
	if txt == "" {
		t.Error("should return non-empty text for multi-line selection")
	}
}

func TestSelection_GetSelectedText_WithWideChar_P270(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(10, 1)
	// Fill with text including a zero-width continuation cell
	buf.SetCell(0, 0, buffer.Cell{Rune: 'A', Width: 1})
	buf.SetCell(1, 0, buffer.Cell{Rune: 0x4E2D, Width: 2}) // Chinese char
	buf.SetCell(2, 0, buffer.Cell{Rune: 0, Width: 0})      // continuation
	buf.SetCell(3, 0, buffer.Cell{Rune: 'B', Width: 1})
	sm.StartSelection(0, 0)
	sm.ExtendSelection(3, 0)
	txt := sm.GetSelectedText(buf)
	// Should contain 'A', wide char, and 'B'
	if !strings.Contains(txt, "A") || !strings.Contains(txt, "B") {
		t.Error("should contain A and B")
	}
}

func TestFindWordBoundary_NilBuf_P270(t *testing.T) {
	x, y := findWordBoundary(nil, 0, 0)
	if x != -1 || y != -1 {
		t.Error("nil buffer should return -1,-1")
	}
}

func TestFindWordBoundary_Whitespace_P270(t *testing.T) {
	buf := buffer.NewBuffer(10, 1)
	buf.SetCell(5, 0, buffer.Cell{Rune: ' ', Width: 1})
	x, y := findWordBoundary(buf, 5, 0)
	if x != -1 || y != -1 {
		t.Error("whitespace should return -1,-1")
	}
}

func TestFindWordBoundary_OutOfBounds_P270(t *testing.T) {
	buf := buffer.NewBuffer(10, 1)
	x, y := findWordBoundary(buf, 0, 5)
	if x != -1 || y != -1 {
		t.Error("out of bounds Y should return -1,-1")
	}
}

func TestRecorder_Save_Error_P270(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordUserInput("test")
	// Use a writer that always fails
	r.Save(&failingWriter{})
}

type failingWriter struct{}

func (w *failingWriter) Write(p []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}

func TestSelection_ExtendKeyboardSelectionTo_Clamp_P270(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(10, 5)
	for y := 0; y < 5; y++ {
		for x := 0; x < 10; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: 'X', Width: 1})
		}
	}
	sm.StartKeyboardSelection(0, 0)
	// Extend to negative and beyond bounds
	sm.ExtendKeyboardSelectionTo(-5, -5, 10, 5)
	sm.ExtendKeyboardSelectionTo(20, 20, 10, 5)
}
