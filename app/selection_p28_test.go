package app

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// drawTextP28 fills a buffer with text at the given position using SetCell.
func drawTextP28(buf *buffer.Buffer, x, y int, text string) {
	for i, r := range text {
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    buffer.RGB(255, 255, 255),
			Bg:    buffer.RGB(0, 0, 0),
		})
	}
}

// --- Word selection ---

func TestP28_SelectWord_Middle(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world test")

	sm.SelectWord(2, 0, buf)
	sel := sm.Selection()
	if sel.Start.X != 0 || sel.End.X != 4 {
		t.Errorf("expected [0,4], got [%d,%d]", sel.Start.X, sel.End.X)
	}
}

func TestP28_SelectWord_SecondWord(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world test")

	sm.SelectWord(8, 0, buf)
	sel := sm.Selection()
	if sel.Start.X != 6 || sel.End.X != 10 {
		t.Errorf("expected [6,10], got [%d,%d]", sel.Start.X, sel.End.X)
	}
}

func TestP28_SelectWord_LastWord(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world test")

	sm.SelectWord(15, 0, buf)
	sel := sm.Selection()
	if sel.Start.X != 12 || sel.End.X != 15 {
		t.Errorf("expected [12,15], got [%d,%d]", sel.Start.X, sel.End.X)
	}
}

func TestP28_SelectWord_Whitespace(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world test")

	sm.SelectWord(5, 0, buf) // space between "hello" and "world"
	if sm.HasSelection() {
		t.Error("expected no selection on whitespace")
	}
}

func TestP28_SelectWord_NilBuffer(t *testing.T) {
	sm := NewSelectionManager()
	sm.SelectWord(0, 0, nil)
	if sm.HasSelection() {
		t.Error("expected no selection with nil buffer")
	}
}

// --- Line selection ---

func TestP28_SelectLine(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 3)
	drawTextP28(buf, 0, 0, "line one")
	drawTextP28(buf, 0, 1, "line two")
	drawTextP28(buf, 0, 2, "line three")

	sm.SelectLine(1, buf)
	sel := sm.Selection()
	if sel.Start.X != 0 || sel.End.X != 19 {
		t.Errorf("expected [0,19], got [%d,%d]", sel.Start.X, sel.End.X)
	}
	if sel.Start.Y != 1 || sel.End.Y != 1 {
		t.Errorf("expected y=1, got [%d,%d]", sel.Start.Y, sel.End.Y)
	}
}

func TestP28_SelectLine_OutOfBounds(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(10, 2)
	sm.SelectLine(5, buf)
	if sm.HasSelection() {
		t.Error("expected no selection for out-of-bounds")
	}
}

func TestP28_SelectLine_NilBuffer(t *testing.T) {
	sm := NewSelectionManager()
	sm.SelectLine(0, nil)
	if sm.HasSelection() {
		t.Error("expected no selection with nil buffer")
	}
}

// --- Select All ---

func TestP28_SelectAll(t *testing.T) {
	sm := NewSelectionManager()
	sm.SelectAll(10, 5)
	sel := sm.Selection()
	if sel.Start.X != 0 || sel.Start.Y != 0 {
		t.Errorf("expected start (0,0), got (%d,%d)", sel.Start.X, sel.Start.Y)
	}
	if sel.End.X != 9 || sel.End.Y != 4 {
		t.Errorf("expected end (9,4), got (%d,%d)", sel.End.X, sel.End.Y)
	}
}

func TestP28_SelectAll_Zero(t *testing.T) {
	sm := NewSelectionManager()
	sm.SelectAll(0, 0)
	if sm.HasSelection() {
		t.Error("expected no selection with zero dimensions")
	}
}

// --- HandleMouseClick ---

func TestP28_HandleMouseClick_DoubleClick(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world")

	sm.HandleMouseClick(2, 2, 0, buf)
	if !sm.HasSelection() {
		t.Fatal("expected selection after double-click")
	}
	sel := sm.Selection()
	if sel.Start.X != 0 || sel.End.X != 4 {
		t.Errorf("expected word [0,4], got [%d,%d]", sel.Start.X, sel.End.X)
	}
}

func TestP28_HandleMouseClick_TripleClick(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world")

	sm.HandleMouseClick(3, 5, 0, buf)
	if !sm.HasSelection() {
		t.Fatal("expected selection after triple-click")
	}
	sel := sm.Selection()
	if sel.Start.X != 0 || sel.End.X != 19 {
		t.Errorf("expected line [0,19], got [%d,%d]", sel.Start.X, sel.End.X)
	}
}

func TestP28_HandleMouseClick_SingleClick(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world")

	sm.HandleMouseClick(1, 5, 0, buf)
	if !sm.HasSelection() {
		t.Error("expected selection after single click")
	}
}

// --- Extended keyboard selection ---

func TestP28_HandleKey_ShiftHome(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(10, 2)

	key := &term.KeyEvent{Key: term.KeyHome, Modifiers: term.ModShift}
	consumed := sm.HandleKey(key, 10, 2, 20, 5)
	if !consumed {
		t.Error("Shift+Home should be consumed")
	}
	c := sm.Cursor()
	if c.X != 0 {
		t.Errorf("expected cursor X=0, got %d", c.X)
	}
}

func TestP28_HandleKey_ShiftEnd(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(5, 2)

	key := &term.KeyEvent{Key: term.KeyEnd, Modifiers: term.ModShift}
	consumed := sm.HandleKey(key, 5, 2, 20, 5)
	if !consumed {
		t.Error("Shift+End should be consumed")
	}
	c := sm.Cursor()
	if c.X != 19 {
		t.Errorf("expected cursor X=19, got %d", c.X)
	}
}

func TestP28_HandleKey_ShiftPageUp(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(0, 8)

	key := &term.KeyEvent{Key: term.KeyPageUp, Modifiers: term.ModShift}
	consumed := sm.HandleKey(key, 0, 8, 20, 10)
	if !consumed {
		t.Error("Shift+PageUp should be consumed")
	}
	c := sm.Cursor()
	if c.Y >= 8 {
		t.Errorf("expected cursor Y < 8, got %d", c.Y)
	}
}

func TestP28_HandleKey_ShiftPageDown(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(0, 0)

	key := &term.KeyEvent{Key: term.KeyPageDown, Modifiers: term.ModShift}
	consumed := sm.HandleKey(key, 0, 0, 20, 10)
	if !consumed {
		t.Error("Shift+PageDown should be consumed")
	}
	c := sm.Cursor()
	if c.Y <= 0 {
		t.Errorf("expected cursor Y > 0, got %d", c.Y)
	}
}

// --- ExtendKeyboardSelectionTo ---

func TestP28_ExtendSelectionTo(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(5, 5)

	sm.ExtendKeyboardSelectionTo(0, 0, 20, 10)
	anchor := sm.Anchor()
	cursor := sm.Cursor()
	if anchor.X != 5 || anchor.Y != 5 {
		t.Errorf("expected anchor (5,5), got (%d,%d)", anchor.X, anchor.Y)
	}
	if cursor.X != 0 || cursor.Y != 0 {
		t.Errorf("expected cursor (0,0), got (%d,%d)", cursor.X, cursor.Y)
	}
}

func TestP28_ExtendSelectionTo_Clamp(t *testing.T) {
	sm := NewSelectionManager()
	sm.StartKeyboardSelection(5, 5)

	sm.ExtendKeyboardSelectionTo(100, 100, 20, 10)
	c := sm.Cursor()
	if c.X != 19 || c.Y != 9 {
		t.Errorf("expected (19,9), got (%d,%d)", c.X, c.Y)
	}
}

func TestP28_ExtendSelectionTo_NotActive(t *testing.T) {
	sm := NewSelectionManager()
	sm.ExtendKeyboardSelectionTo(5, 5, 20, 10)
	if sm.HasSelection() {
		t.Error("expected no selection when not active")
	}
}

// --- CopySelectionToWriter (SelectionManager) ---

func TestP28_CopySelectionToWriter(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world")

	sm.StartSelection(0, 0)
	sm.ExtendSelection(4, 0)
	sm.EndSelection()

	var sb strings.Builder
	n := sm.CopySelectionToWriter(buf, &sb)
	if n == 0 {
		t.Error("expected non-zero bytes")
	}
	if !strings.Contains(sb.String(), "\x1b]52") {
		t.Error("expected OSC52 sequence")
	}
}

func TestP28_CopySelectionToWriter_NoSelection(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(10, 1)
	var sb strings.Builder
	n := sm.CopySelectionToWriter(buf, &sb)
	if n != 0 {
		t.Error("expected 0 bytes without selection")
	}
}

func TestP28_CopySelectionToWriter_NilWriter(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(10, 1)
	sm.StartSelection(0, 0)
	sm.ExtendSelection(5, 0)
	sm.EndSelection()

	n := sm.CopySelectionToWriter(buf, nil)
	if n != 0 {
		t.Error("expected 0 bytes with nil writer")
	}
}

// --- ChatApp clipboard integration ---

func TestP28_ChatApp_HandleClipboardKey_Copy(t *testing.T) {
	chat := NewChatApp(80, 24)
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)

	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "hello world")

	sm.StartSelection(0, 0)
	sm.ExtendSelection(4, 0)
	sm.EndSelection()

	var sb strings.Builder
	key := &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl | term.ModShift}
	consumed := chat.handleClipboardKey(key, buf, &sb)
	if !consumed {
		t.Error("Ctrl+Shift+C should be consumed")
	}
	if sb.Len() == 0 {
		t.Error("expected OSC52 output")
	}
}

func TestP28_ChatApp_HandleClipboardKey_CopyNoSelection(t *testing.T) {
	chat := NewChatApp(80, 24)
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)

	buf := buffer.NewBuffer(10, 1)
	var sb strings.Builder
	key := &term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl | term.ModShift}
	consumed := chat.handleClipboardKey(key, buf, &sb)
	if consumed {
		t.Error("should not be consumed without selection")
	}
}

func TestP28_ChatApp_HandleClipboardKey_Paste(t *testing.T) {
	chat := NewChatApp(80, 24)
	var sb strings.Builder
	key := &term.KeyEvent{Rune: 'v', Modifiers: term.ModCtrl | term.ModShift}
	consumed := chat.handleClipboardKey(key, nil, &sb)
	if !consumed {
		t.Error("Ctrl+Shift+V should be consumed")
	}
	if sb.Len() == 0 {
		t.Error("expected OSC52 paste query")
	}
}

func TestP28_ChatApp_HandleClipboardKey_OtherKey(t *testing.T) {
	chat := NewChatApp(80, 24)
	key := &term.KeyEvent{Rune: 'x', Modifiers: term.ModCtrl}
	var sb strings.Builder
	consumed := chat.handleClipboardKey(key, nil, &sb)
	if consumed {
		t.Error("non-clipboard key should not be consumed")
	}
}

func TestP28_ChatApp_HandleClipboardKey_NilKey(t *testing.T) {
	chat := NewChatApp(80, 24)
	var sb strings.Builder
	consumed := chat.handleClipboardKey(nil, nil, &sb)
	if consumed {
		t.Error("nil key should not be consumed")
	}
}

func TestP28_ChatApp_CopySelectionToWriter(t *testing.T) {
	chat := NewChatApp(80, 24)
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)

	buf := buffer.NewBuffer(20, 1)
	drawTextP28(buf, 0, 0, "test text")

	sm.StartSelection(0, 0)
	sm.ExtendSelection(3, 0)
	sm.EndSelection()

	var sb strings.Builder
	ok := chat.CopySelectionToWriter(buf, &sb)
	if !ok {
		t.Error("expected CopySelectionToWriter to succeed")
	}
}

func TestP28_ChatApp_CopySelectionToWriter_NoSelection(t *testing.T) {
	chat := NewChatApp(80, 24)
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)

	buf := buffer.NewBuffer(10, 1)
	var sb strings.Builder
	ok := chat.CopySelectionToWriter(buf, &sb)
	if ok {
		t.Error("expected false without selection")
	}
}

func TestP28_ChatApp_PasteFromClipboard(t *testing.T) {
	chat := NewChatApp(80, 24)
	var sb strings.Builder
	ok := chat.PasteFromClipboard(&sb)
	if !ok {
		t.Error("expected PasteFromClipboard to succeed")
	}
	if sb.Len() == 0 {
		t.Error("expected OSC52 paste query")
	}
}

// --- Concurrent ---

func TestP28_ConcurrentSelectAndCopy(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 5)
	drawTextP28(buf, 0, 0, "hello world test")
	drawTextP28(buf, 0, 1, "second line")

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			switch n % 4 {
			case 0:
				sm.SelectWord(2, 0, buf)
			case 1:
				sm.SelectLine(1, buf)
			case 2:
				var sb strings.Builder
				sm.CopySelectionToWriter(buf, &sb)
			case 3:
				sm.SelectAll(20, 5)
			}
		}(i)
	}
	wg.Wait()
}

// --- IsWhitespace ---

func TestP28_IsWhitespace(t *testing.T) {
	if !IsWhitespace(' ') {
		t.Error("' ' should be whitespace")
	}
	if !IsWhitespace('\t') {
		t.Error("'\\t' should be whitespace")
	}
	if !IsWhitespace(0) {
		t.Error("0 should be whitespace")
	}
	if IsWhitespace('a') {
		t.Error("'a' should not be whitespace")
	}
}
