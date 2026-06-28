package app

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// helper: create a buffer filled with sequential letters per row
func newSelectionTestBuffer(w, h int) *buffer.Buffer {
	buf := buffer.NewBuffer(w, h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, buffer.Cell{
				Rune:  rune('A' + (x % 26)),
				Width: 1,
				Fg:    buffer.RGB(255, 255, 255),
				Bg:    buffer.RGB(0, 0, 0),
			})
		}
	}
	return buf
}

// --- Construction ---

func TestNewSelectionManager(t *testing.T) {
	sm := NewSelectionManager()
	if sm == nil {
		t.Fatal("NewSelectionManager returned nil")
	}
	if sm.HasSelection() {
		t.Error("new SelectionManager should not have an active selection")
	}
	if sm.IsSelecting() {
		t.Error("new SelectionManager should not be in selecting state")
	}
}

// --- Mouse selection ---

func TestSelection_MouseStartAndExtend(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(5, 3)

	if !sm.HasSelection() {
		t.Error("expected active selection after StartSelection")
	}
	if !sm.IsSelecting() {
		t.Error("expected selecting=true after StartSelection")
	}

	start, end, ok := sm.SelectionRange()
	if !ok {
		t.Fatal("expected SelectionRange to return ok=true")
	}
	if start.X != 5 || start.Y != 3 {
		t.Errorf("start = (%d,%d), want (5,3)", start.X, start.Y)
	}
	if end.X != 5 || end.Y != 3 {
		t.Errorf("end = (%d,%d), want (5,3)", end.X, end.Y)
	}
}

func TestSelection_MouseDragExtend(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(2, 1)
	sm.ExtendSelection(8, 3)

	start, end, _ := sm.SelectionRange()
	if start.X != 2 || start.Y != 1 {
		t.Errorf("start = (%d,%d), want (2,1)", start.X, start.Y)
	}
	if end.X != 8 || end.Y != 3 {
		t.Errorf("end = (%d,%d), want (8,3)", end.X, end.Y)
	}
}

func TestSelection_MouseDragReverse(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(8, 3)
	sm.ExtendSelection(2, 1)

	start, end, _ := sm.SelectionRange()
	// SelectionRange should normalize.
	if start.X != 2 || start.Y != 1 {
		t.Errorf("normalized start = (%d,%d), want (2,1)", start.X, start.Y)
	}
	if end.X != 8 || end.Y != 3 {
		t.Errorf("normalized end = (%d,%d), want (8,3)", end.X, end.Y)
	}
}

func TestSelection_MouseEndClearsClick(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(5, 5)
	// No extend = click, not drag.
	sm.EndSelection()
	if sm.HasSelection() {
		t.Error("selection should be cleared when start==end (click)")
	}
}

func TestSelection_MouseEndKeepsDrag(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(10, 2)
	sm.EndSelection()
	if !sm.HasSelection() {
		t.Error("selection should remain after drag end")
	}
	if sm.IsSelecting() {
		t.Error("selecting should be false after EndSelection")
	}
}

// --- HandleMouse ---

func TestSelection_HandleMouseDownStarts(t *testing.T) {
	sm := NewSelectionManager()
	mouse := &term.MouseEvent{
		X: 3, Y: 2, Button: term.MouseLeft, Action: term.MouseDown,
	}
	consumed := sm.HandleMouse(mouse)
	if !consumed {
		t.Error("HandleMouse should consume MouseDown")
	}
	if !sm.HasSelection() {
		t.Error("expected selection after MouseDown")
	}
}

func TestSelection_HandleMouseDragExtends(t *testing.T) {
	sm := NewSelectionManager()
	sm.HandleMouse(&term.MouseEvent{
		X: 0, Y: 0, Button: term.MouseLeft, Action: term.MouseDown,
	})
	sm.HandleMouse(&term.MouseEvent{
		X: 5, Y: 2, Button: term.MouseLeft, Action: term.MouseDrag,
	})
	_, end, _ := sm.SelectionRange()
	if end.X != 5 || end.Y != 2 {
		t.Errorf("end = (%d,%d), want (5,2)", end.X, end.Y)
	}
}

func TestSelection_HandleMouseUpFinalizes(t *testing.T) {
	sm := NewSelectionManager()
	sm.HandleMouse(&term.MouseEvent{
		X: 0, Y: 0, Button: term.MouseLeft, Action: term.MouseDown,
	})
	sm.HandleMouse(&term.MouseEvent{
		X: 3, Y: 1, Button: term.MouseLeft, Action: term.MouseDrag,
	})
	consumed := sm.HandleMouse(&term.MouseEvent{
		X: 3, Y: 1, Button: term.MouseLeft, Action: term.MouseUp,
	})
	if !consumed {
		t.Error("MouseUp should be consumed during drag")
	}
	if !sm.HasSelection() {
		t.Error("selection should remain after drag+up")
	}
}

func TestSelection_HandleMouseRightButtonIgnored(t *testing.T) {
	sm := NewSelectionManager()
	mouse := &term.MouseEvent{
		X: 3, Y: 2, Button: term.MouseRight, Action: term.MouseDown,
	}
	consumed := sm.HandleMouse(mouse)
	if consumed {
		t.Error("right button should not be consumed")
	}
	if sm.HasSelection() {
		t.Error("right button should not start selection")
	}
}

func TestSelection_HandleMouseNil(t *testing.T) {
	sm := NewSelectionManager()
	if sm.HandleMouse(nil) {
		t.Error("nil mouse should not be consumed")
	}
}

// --- Keyboard selection ---

func TestSelection_KeyboardStartAndExtend(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(5, 5)
	sm.ExtendKeyboardSelection(3, 0, 80, 24)

	start, end, ok := sm.SelectionRange()
	if !ok {
		t.Fatal("expected selection")
	}
	if start.X != 5 || start.Y != 5 {
		t.Errorf("start = (%d,%d), want (5,5)", start.X, start.Y)
	}
	if end.X != 8 || end.Y != 5 {
		t.Errorf("end = (%d,%d), want (8,5)", end.X, end.Y)
	}
}

func TestSelection_KeyboardClamp(t *testing.T) {
	sm := NewSelectionManager()

	// Negative clamp.
	sm.StartKeyboardSelection(0, 0)
	sm.ExtendKeyboardSelection(-5, -5, 80, 24)
	cursor := sm.Cursor()
	if cursor.X != 0 || cursor.Y != 0 {
		t.Errorf("cursor = (%d,%d), want (0,0) after negative clamp", cursor.X, cursor.Y)
	}

	// Positive clamp.
	sm.StartKeyboardSelection(79, 23)
	sm.ExtendKeyboardSelection(10, 10, 80, 24)
	cursor = sm.Cursor()
	if cursor.X != 79 || cursor.Y != 23 {
		t.Errorf("cursor = (%d,%d), want (79,23) after positive clamp", cursor.X, cursor.Y)
	}
}

func TestSelection_HandleKeyShiftArrow(t *testing.T) {
	sm := NewSelectionManager()
	key := &term.KeyEvent{
		Key:       term.KeyRight,
		Modifiers: term.ModShift,
	}
	consumed := sm.HandleKey(key, 5, 5, 80, 24)
	if !consumed {
		t.Error("Shift+Right should be consumed")
	}
	if !sm.HasSelection() {
		t.Error("expected selection after Shift+Right")
	}
	cursor := sm.Cursor()
	if cursor.X != 6 {
		t.Errorf("cursor.X = %d, want 6", cursor.X)
	}
}

func TestSelection_HandleKeyEscapeClears(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(5, 5)

	key := &term.KeyEvent{Key: term.KeyEscape}
	consumed := sm.HandleKey(key, 0, 0, 80, 24)
	if !consumed {
		t.Error("Escape should be consumed when selection active")
	}
	if sm.HasSelection() {
		t.Error("selection should be cleared after Escape")
	}
}

func TestSelection_HandleKeyEscapeNoSelection(t *testing.T) {
	sm := NewSelectionManager()
	key := &term.KeyEvent{Key: term.KeyEscape}
	if sm.HandleKey(key, 0, 0, 80, 24) {
		t.Error("Escape should not be consumed when no selection")
	}
}

func TestSelection_HandleKeyNilKey(t *testing.T) {
	sm := NewSelectionManager()
	if sm.HandleKey(nil, 0, 0, 80, 24) {
		t.Error("nil key should not be consumed")
	}
}

func TestSelection_HandleKeyNonShiftIgnored(t *testing.T) {
	sm := NewSelectionManager()
	key := &term.KeyEvent{Key: term.KeyRight, Modifiers: 0}
	if sm.HandleKey(key, 5, 5, 80, 24) {
		t.Error("non-Shift key should not be consumed")
	}
}

// --- Text extraction ---

func TestSelection_GetSelectedTextSingleRow(t *testing.T) {
	buf := newSelectionTestBuffer(10, 5)
	sm := NewSelectionManager()
	sm.StartSelection(2, 1)
	sm.ExtendSelection(5, 1)

	text := sm.GetSelectedText(buf)
	if text != "CDEF" {
		t.Errorf("got %q, want %q", text, "CDEF")
	}
}

func TestSelection_GetSelectedTextMultiRow(t *testing.T) {
	buf := newSelectionTestBuffer(10, 5)
	sm := NewSelectionManager()
	sm.StartSelection(2, 1)
	sm.ExtendSelection(3, 2)

	text := sm.GetSelectedText(buf)
	lines := strings.Split(text, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), text)
	}
	if lines[0] != "CDEFGHIJ" {
		t.Errorf("line 0 = %q, want %q", lines[0], "CDEFGHIJ")
	}
	if lines[1] != "ABCD" {
		t.Errorf("line 1 = %q, want %q", lines[1], "ABCD")
	}
}

func TestSelection_GetSelectedTextNoSelection(t *testing.T) {
	buf := newSelectionTestBuffer(10, 5)
	sm := NewSelectionManager()
	if sm.GetSelectedText(buf) != "" {
		t.Error("no selection should return empty string")
	}
}

func TestSelection_GetSelectedTextNilBuffer(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(5, 5)
	if sm.GetSelectedText(nil) != "" {
		t.Error("nil buffer should return empty string")
	}
}

// --- OSC52 copy ---

func TestSelection_CopySelection(t *testing.T) {
	buf := newSelectionTestBuffer(10, 5)
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(3, 0)

	seq := sm.CopySelection(buf)
	if seq == "" {
		t.Fatal("expected non-empty OSC52 sequence")
	}
	if !strings.HasPrefix(seq, "\x1b]52;") {
		t.Errorf("expected OSC52 prefix, got %q", seq[:min(10, len(seq))])
	}
}

func TestSelection_CopySelectionNoSelection(t *testing.T) {
	buf := newSelectionTestBuffer(10, 5)
	sm := NewSelectionManager()
	if sm.CopySelection(buf) != "" {
		t.Error("no selection should return empty string")
	}
}

func TestSelection_CopySelectionSource(t *testing.T) {
	buf := newSelectionTestBuffer(10, 5)
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(2, 0)

	seq := sm.CopySelectionSource(buf, term.ClipboardPrimary)
	if seq == "" {
		t.Fatal("expected non-empty OSC52 sequence")
	}
}

// --- Highlight rendering ---

func TestSelection_ApplyHighlightSwapsColors(t *testing.T) {
	buf := buffer.NewBuffer(5, 1)
	white := buffer.RGB(255, 255, 255)
	black := buffer.RGB(0, 0, 0)
	for x := 0; x < 5; x++ {
		buf.SetCell(x, 0, buffer.Cell{
			Rune:  'X',
			Width: 1,
			Fg:    white,
			Bg:    black,
		})
	}

	sm := NewSelectionManager()
	// Select cells 1-3.
	sm.StartSelection(1, 0)
	sm.ExtendSelection(3, 0)
	sm.EndSelection()

	sm.ApplyHighlight(buf)

	// Selected cells should have swapped Fg/Bg.
	for x := 1; x <= 3; x++ {
		cell := buf.GetCell(x, 0)
		if !cell.Fg.Equal(black) {
			t.Errorf("cell %d Fg should be swapped to black, got %v", x, cell.Fg)
		}
		if !cell.Bg.Equal(white) {
			t.Errorf("cell %d Bg should be swapped to white, got %v", x, cell.Bg)
		}
	}

	// Non-selected cell 0 should be unchanged.
	cell0 := buf.GetCell(0, 0)
	if !cell0.Fg.Equal(white) {
		t.Error("cell 0 Fg should be unchanged")
	}
}

func TestSelection_ApplyHighlightNoSelection(t *testing.T) {
	buf := newSelectionTestBuffer(5, 3)
	sm := NewSelectionManager()
	sm.ApplyHighlight(buf) // should not panic
}

func TestSelection_ApplyHighlightNilBuffer(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(5, 5)
	sm.ApplyHighlight(nil) // should not panic
}

// --- Clear ---

func TestSelection_Clear(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(0, 0)
	sm.ExtendSelection(5, 5)

	if !sm.HasSelection() {
		t.Fatal("expected selection")
	}

	sm.Clear()
	if sm.HasSelection() {
		t.Error("selection should be cleared")
	}
}

// --- Cursor/Anchor ---

func TestSelection_CursorAnchor(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartSelection(3, 4)

	anchor := sm.Anchor()
	if anchor.X != 3 || anchor.Y != 4 {
		t.Errorf("anchor = (%d,%d), want (3,4)", anchor.X, anchor.Y)
	}

	sm.ExtendSelection(8, 10)
	cursor := sm.Cursor()
	if cursor.X != 8 || cursor.Y != 10 {
		t.Errorf("cursor = (%d,%d), want (8,10)", cursor.X, cursor.Y)
	}

	// Anchor should not move.
	anchor = sm.Anchor()
	if anchor.X != 3 || anchor.Y != 4 {
		t.Errorf("anchor = (%d,%d), want (3,4) after extend", anchor.X, anchor.Y)
	}
}

// --- Concurrency ---

func TestSelection_ConcurrentAccess(t *testing.T) {
	sm := NewSelectionManager()
	buf := newSelectionTestBuffer(80, 24)

	var wg sync.WaitGroup

	// Writer: start/extend selection.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			sm.StartSelection(i%20, i%10)
			sm.ExtendSelection((i+1)%20, (i+1)%10)
		}
	}()

	// Reader: query selection.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			sm.HasSelection()
			sm.SelectionRange()
			sm.Cursor()
			sm.Anchor()
		}
	}()

	// Text extractor + painter.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			sm.GetSelectedText(buf)
			sm.ApplyHighlight(buf)
		}
	}()

	wg.Wait()
}

func TestSelection_ConcurrentMouseAndKey(t *testing.T) {
	sm := NewSelectionManager()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			sm.HandleMouse(&term.MouseEvent{
				X: i % 20, Y: i % 10,
				Button: term.MouseLeft, Action: term.MouseDown,
			})
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			sm.HandleKey(&term.KeyEvent{
				Key:       term.KeyRight,
				Modifiers: term.ModShift,
			}, 5, 5, 80, 24)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			sm.Clear()
		}
	}()

	wg.Wait()
}

// --- min helper ---

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
