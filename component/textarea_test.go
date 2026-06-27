package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- Helpers ---

func taKeyChar(r rune) *term.KeyEvent {
	return &term.KeyEvent{Rune: r}
}

func taKeyCtrl(r rune) *term.KeyEvent {
	return &term.KeyEvent{Rune: r, Modifiers: term.ModCtrl}
}

func taKeySpecial(k term.KeyCode) *term.KeyEvent {
	return &term.KeyEvent{Key: k}
}

func taKeyAlt(k term.KeyCode) *term.KeyEvent {
	return &term.KeyEvent{Key: k, Modifiers: term.ModAlt}
}

func typeText(ta *TextArea, s string) {
	for _, r := range s {
		ta.HandleKey(taKeyChar(r))
	}
}

// --- Tests ---

func TestTextArea_New(t *testing.T) {
	ta := NewTextArea()
	if ta.LineCount() != 1 {
		t.Errorf("LineCount: got %d, want 1", ta.LineCount())
	}
	if ta.Text() != "" {
		t.Errorf("Text: got %q, want empty", ta.Text())
	}
	x, y := ta.CursorPos()
	if x != 0 || y != 0 {
		t.Errorf("Cursor: got (%d,%d), want (0,0)", x, y)
	}
}

func TestTextArea_InsertChars(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	if ta.Text() != "Hello" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Hello")
	}
	x, _ := ta.CursorPos()
	if x != 5 {
		t.Errorf("cursorX: got %d, want 5", x)
	}
}

func TestTextArea_EnterCreatesNewLine(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "AB")
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	typeText(ta, "CD")
	if ta.LineCount() != 2 {
		t.Errorf("LineCount: got %d, want 2", ta.LineCount())
	}
	if ta.Text() != "AB\nCD" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "AB\nCD")
	}
}

func TestTextArea_EnterAtMiddle(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	// Move cursor to after "He" (position 2)
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	if ta.Text() != "He\nllo" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "He\nllo")
	}
	x, y := ta.CursorPos()
	if x != 0 || y != 1 {
		t.Errorf("Cursor: got (%d,%d), want (0,1)", x, y)
	}
}

func TestTextArea_BackspaceInLine(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	ta.HandleKey(taKeySpecial(term.KeyBackspace))
	if ta.Text() != "Hell" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Hell")
	}
	x, _ := ta.CursorPos()
	if x != 4 {
		t.Errorf("cursorX: got %d, want 4", x)
	}
}

func TestTextArea_BackspaceAtStartJoinsLines(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "AB")
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	typeText(ta, "CD")
	// Cursor at (2, 1) — move to (0, 1)
	ta.HandleKey(taKeyCtrl('a'))
	// Now backspace should join
	ta.HandleKey(taKeySpecial(term.KeyBackspace))
	if ta.LineCount() != 1 {
		t.Errorf("LineCount: got %d, want 1", ta.LineCount())
	}
	if ta.Text() != "ABCD" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "ABCD")
	}
	x, y := ta.CursorPos()
	if x != 2 || y != 0 {
		t.Errorf("Cursor: got (%d,%d), want (2,0)", x, y)
	}
}

func TestTextArea_DeleteAtCursor(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	// Move cursor to position 2
	for i := 0; i < 3; i++ {
		ta.HandleKey(taKeySpecial(term.KeyLeft))
	}
	ta.HandleKey(taKeySpecial(term.KeyDelete))
	if ta.Text() != "Helo" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Helo")
	}
}

func TestTextArea_DeleteAtEndJoinsNextLine(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "AB")
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	typeText(ta, "CD")
	// Go back to end of line 0
	ta.HandleKey(taKeySpecial(term.KeyUp))
	ta.HandleKey(taKeyCtrl('e'))
	// Now delete should join line 1
	ta.HandleKey(taKeySpecial(term.KeyDelete))
	if ta.LineCount() != 1 {
		t.Errorf("LineCount: got %d, want 1", ta.LineCount())
	}
	if ta.Text() != "ABCD" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "ABCD")
	}
}

func TestTextArea_TabInsertsSpaces(t *testing.T) {
	ta := NewTextArea()
	ta.HandleKey(taKeySpecial(term.KeyTab))
	if ta.Text() != "    " {
		t.Errorf("Text: got %q, want %q", ta.Text(), "    ")
	}
	x, _ := ta.CursorPos()
	if x != TabWidth {
		t.Errorf("cursorX: got %d, want %d", x, TabWidth)
	}
}

func TestTextArea_ArrowNavigation(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "AB")
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	typeText(ta, "CD")
	// Cursor at (2, 1)
	// Up → should go to line 0
	ta.HandleKey(taKeySpecial(term.KeyUp))
	_, y := ta.CursorPos()
	if y != 0 {
		t.Errorf("After Up: y=%d, want 0", y)
	}
	// Down → back to line 1
	ta.HandleKey(taKeySpecial(term.KeyDown))
	_, y = ta.CursorPos()
	if y != 1 {
		t.Errorf("After Down: y=%d, want 1", y)
	}
}

func TestTextArea_LeftRightAcrossLines(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "AB")
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	typeText(ta, "CD")
	// Cursor at (2,1). Left 3 should bring us to end of line 0
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	x, y := ta.CursorPos()
	if x != 2 || y != 0 {
		t.Errorf("Cursor: got (%d,%d), want (2,0)", x, y)
	}
	// Right → start of line 1
	ta.HandleKey(taKeySpecial(term.KeyRight))
	x, y = ta.CursorPos()
	if x != 0 || y != 1 {
		t.Errorf("Cursor: got (%d,%d), want (0,1)", x, y)
	}
}

func TestTextArea_HomeEnd(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	ta.HandleKey(taKeySpecial(term.KeyEnd))
	x, _ := ta.CursorPos()
	if x != 5 {
		t.Errorf("After End: x=%d, want 5", x)
	}
	ta.HandleKey(taKeySpecial(term.KeyHome))
	x, _ = ta.CursorPos()
	if x != 0 {
		t.Errorf("After Home: x=%d, want 0", x)
	}
}

func TestTextArea_CtrlA_CtrlE(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	// Ctrl+A = home
	ta.HandleKey(taKeyCtrl('a'))
	x, _ := ta.CursorPos()
	if x != 0 {
		t.Errorf("Ctrl+A: x=%d, want 0", x)
	}
	// Ctrl+E = end
	ta.HandleKey(taKeyCtrl('e'))
	x, _ = ta.CursorPos()
	if x != 5 {
		t.Errorf("Ctrl+E: x=%d, want 5", x)
	}
}

func TestTextArea_CtrlK_DeleteToEndOfLine(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello World")
	// Move to position 5
	for i := 0; i < 6; i++ {
		ta.HandleKey(taKeySpecial(term.KeyLeft))
	}
	ta.HandleKey(taKeyCtrl('k'))
	if ta.Text() != "Hello" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Hello")
	}
}

func TestTextArea_CtrlU_DeleteToStartOfLine(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello World")
	// Move to position 5
	for i := 0; i < 6; i++ {
		ta.HandleKey(taKeySpecial(term.KeyLeft))
	}
	ta.HandleKey(taKeyCtrl('u'))
	if ta.Text() != " World" {
		t.Errorf("Text: got %q, want %q", ta.Text(), " World")
	}
	x, _ := ta.CursorPos()
	if x != 0 {
		t.Errorf("Ctrl+U cursorX: got %d, want 0", x)
	}
}

func TestTextArea_CtrlW_DeleteWordBack(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello World")
	ta.HandleKey(taKeyCtrl('w'))
	if ta.Text() != "Hello " {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Hello ")
	}
	ta.HandleKey(taKeyCtrl('w'))
	if ta.Text() != "" {
		t.Errorf("Text: got %q, want empty", ta.Text())
	}
}

func TestTextArea_CtrlW_DeletesTrailingSpaces(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello   ")
	ta.HandleKey(taKeyCtrl('w'))
	// Standard Ctrl+W: deletes trailing spaces + preceding word
	if ta.Text() != "" {
		t.Errorf("Text: got %q, want empty", ta.Text())
	}
}

func TestTextArea_SetText(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Line1\nLine2\nLine3")
	if ta.LineCount() != 3 {
		t.Errorf("LineCount: got %d, want 3", ta.LineCount())
	}
	if ta.Text() != "Line1\nLine2\nLine3" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Line1\nLine2\nLine3")
	}
	x, y := ta.CursorPos()
	if x != 0 || y != 0 {
		t.Errorf("Cursor: got (%d,%d), want (0,0)", x, y)
	}
}

func TestTextArea_SetTextEmpty(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	ta.SetText("")
	if ta.LineCount() != 1 {
		t.Errorf("LineCount: got %d, want 1", ta.LineCount())
	}
	if ta.Text() != "" {
		t.Errorf("Text: got %q, want empty", ta.Text())
	}
}

func TestTextArea_Clear(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hello")
	ta.HandleKey(taKeySpecial(term.KeyEnter))
	typeText(ta, "World")
	ta.Clear()
	if ta.LineCount() != 1 {
		t.Errorf("LineCount: got %d, want 1", ta.LineCount())
	}
	if ta.Text() != "" {
		t.Errorf("Text: got %q, want empty", ta.Text())
	}
}

func TestTextArea_InsertTextSingleLine(t *testing.T) {
	ta := NewTextArea()
	ta.InsertText("Hello")
	if ta.Text() != "Hello" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Hello")
	}
}

func TestTextArea_InsertTextMultiLine(t *testing.T) {
	ta := NewTextArea()
	ta.InsertText("AB\nCD")
	if ta.LineCount() != 2 {
		t.Errorf("LineCount: got %d, want 2", ta.LineCount())
	}
	if ta.Text() != "AB\nCD" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "AB\nCD")
	}
}

func TestTextArea_InsertTextAtCursor(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "AC")
	// Move cursor between A and C
	ta.HandleKey(taKeySpecial(term.KeyLeft))
	ta.InsertText("B")
	if ta.Text() != "ABC" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "ABC")
	}
}

func TestTextArea_Measure(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Hello\nWorld!\nX")
	size := ta.Measure(Unbounded())
	// Longest line = "World!" = 6 chars
	if size.W != 6 {
		t.Errorf("W: got %d, want 6", size.W)
	}
	if size.H != 3 {
		t.Errorf("H: got %d, want 3", size.H)
	}
}

func TestTextArea_MeasureClamped(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Very long line here")
	size := ta.Measure(Bounded(5, 1))
	if size.W != 5 {
		t.Errorf("W: got %d, want 5", size.W)
	}
	if size.H != 1 {
		t.Errorf("H: got %d, want 1", size.H)
	}
}

func TestTextArea_PaintEmpty(t *testing.T) {
	ta := NewTextArea()
	ta.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	// Should not panic
	ta.Paint(buf)
	// Cursor should be at (0,0)
	cell := buf.GetCell(0, 0)
	if cell.Flags&buffer.Reverse == 0 {
		t.Error("Expected reverse video cursor at (0,0)")
	}
}

func TestTextArea_PaintText(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Hello\nWorld")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	ta.Paint(buf)

	// Line 0: "Hello"
	for i, r := range "Hello" {
		cell := buf.GetCell(i, 0)
		if cell.Rune != r {
			t.Errorf("line0[%d]: got %q, want %q", i, string(cell.Rune), string(r))
		}
	}
	// Line 1: "World"
	for i, r := range "World" {
		cell := buf.GetCell(i, 1)
		if cell.Rune != r {
			t.Errorf("line1[%d]: got %q, want %q", i, string(cell.Rune), string(r))
		}
	}
}

func TestTextArea_PaintCursor(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "Hi")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	ta.Paint(buf)

	// Cursor at position 2 (end of "Hi")
	cell := buf.GetCell(2, 0)
	if cell.Flags&buffer.Reverse == 0 {
		t.Error("Expected cursor at position 2")
	}
}

func TestTextArea_PaintAtOffset(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("AB")
	ta.SetBounds(Rect{X: 3, Y: 2, W: 10, H: 3})
	buf := buffer.NewBuffer(15, 6)
	ta.Paint(buf)

	cell := buf.GetCell(3, 2)
	if cell.Rune != 'A' {
		t.Errorf("cell at (3,2): got %q, want 'A'", string(cell.Rune))
	}
	cell = buf.GetCell(4, 2)
	if cell.Rune != 'B' {
		t.Errorf("cell at (4,2): got %q, want 'B'", string(cell.Rune))
	}
}

func TestTextArea_ScrollBehavior(t *testing.T) {
	ta := NewTextArea()
	// Create 10 lines
	ta.SetText("Line0\nLine1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7\nLine8\nLine9")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	// Move cursor to line 9 (last)
	for i := 0; i < 9; i++ {
		ta.HandleKey(taKeySpecial(term.KeyDown))
	}
	_, y := ta.CursorPos()
	if y != 9 {
		t.Errorf("cursorY: got %d, want 9", y)
	}
	// scrollY should have adjusted so cursor is visible
	if ta.scrollY > 9 || ta.scrollY+3 <= 9 {
		t.Errorf("scrollY=%d, cursor at y=9 should be visible in H=3", ta.scrollY)
	}
}

func TestTextArea_PageUpPageDown(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Line0\nLine1\nLine2\nLine3\nLine4\nLine5\nLine6\nLine7\nLine8\nLine9")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	// Cursor starts at (0,0) after SetText
	// PageDown should advance cursor by 3
	ta.HandleKey(taKeySpecial(term.KeyPageDown))
	_, y := ta.CursorPos()
	if y != 3 {
		t.Errorf("After PageDown: y=%d, want 3", y)
	}
	// PageUp should go back by 3
	ta.HandleKey(taKeySpecial(term.KeyPageUp))
	_, y = ta.CursorPos()
	if y != 0 {
		t.Errorf("After PageUp: y=%d, want 0", y)
	}
}

func TestTextArea_PageDownClamped(t *testing.T) {
	ta := NewTextArea()
	for i := 0; i < 5; i++ {
		ta.InsertText("Line")
		if i < 4 {
			ta.HandleKey(taKeySpecial(term.KeyEnter))
		}
	}
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	ta.HandleKey(taKeySpecial(term.KeyPageDown))
	ta.HandleKey(taKeySpecial(term.KeyPageDown))
	_, y := ta.CursorPos()
	if y != 4 {
		t.Errorf("After 2x PageDown: y=%d, want 4 (clamped to last line)", y)
	}
}

func TestTextArea_AltUpMoveLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Line0\nLine1\nLine2")
	// Cursor at line 2
	ta.HandleKey(taKeySpecial(term.KeyDown))
	ta.HandleKey(taKeySpecial(term.KeyDown))
	_, y := ta.CursorPos()
	if y != 2 {
		t.Fatalf("setup: y=%d, want 2", y)
	}
	// Alt+Up should move line 2 up
	ta.HandleKey(taKeyAlt(term.KeyUp))
	_, y = ta.CursorPos()
	if y != 1 {
		t.Errorf("After Alt+Up: y=%d, want 1", y)
	}
	// Lines should be swapped: Line0, Line2, Line1
	if ta.Text() != "Line0\nLine2\nLine1" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Line0\nLine2\nLine1")
	}
}

func TestTextArea_AltDownMoveLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Line0\nLine1\nLine2")
	// Cursor at line 0
	// Alt+Down should move line 0 down
	ta.HandleKey(taKeyAlt(term.KeyDown))
	_, y := ta.CursorPos()
	if y != 1 {
		t.Errorf("After Alt+Down: y=%d, want 1", y)
	}
	// Lines should be: Line1, Line0, Line2
	if ta.Text() != "Line1\nLine0\nLine2" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "Line1\nLine0\nLine2")
	}
}

func TestTextArea_CursorClampsOnLineChange(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Hello\nX\nWorld")
	// Move cursor to end of line 0 (x=5)
	ta.HandleKey(taKeyCtrl('e'))
	x, _ := ta.CursorPos()
	if x != 5 {
		t.Fatalf("After Ctrl+E: x=%d, want 5", x)
	}
	// Down → cursor should be clamped to line 1 length (1)
	ta.HandleKey(taKeySpecial(term.KeyDown))
	x, y := ta.CursorPos()
	if x != 1 || y != 1 {
		t.Errorf("Cursor: got (%d,%d), want (1,1)", x, y)
	}
}

func TestTextArea_PaintTooSmall(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("Hello")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	// Should not panic
	ta.Paint(buf)
}

func TestTextArea_SpaceKey(t *testing.T) {
	ta := NewTextArea()
	ta.HandleKey(taKeySpecial(term.KeySpace))
	ta.HandleKey(taKeySpecial(term.KeySpace))
	if ta.Text() != "  " {
		t.Errorf("Text: got %q, want %q", ta.Text(), "  ")
	}
}

func TestTextArea_UnicodeInput(t *testing.T) {
	ta := NewTextArea()
	typeText(ta, "你好")
	if ta.Text() != "你好" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "你好")
	}
	x, _ := ta.CursorPos()
	if x != 2 {
		t.Errorf("cursorX: got %d, want 2", x)
	}
}

func TestTextArea_MultiLineBackspaceChain(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("AAA\nBBB\nCCC")
	// Cursor at (0,0), navigate to (0, 2)
	ta.HandleKey(taKeySpecial(term.KeyDown))
	ta.HandleKey(taKeySpecial(term.KeyDown))
	// Backspace at start of line 2 → join with line 1
	ta.HandleKey(taKeySpecial(term.KeyBackspace))
	if ta.Text() != "AAA\nBBBCCC" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "AAA\nBBBCCC")
	}
	// Another backspace at start → join with line 0
	ta.HandleKey(taKeyCtrl('a'))
	ta.HandleKey(taKeySpecial(term.KeyBackspace))
	if ta.Text() != "AAABBBCCC" {
		t.Errorf("Text: got %q, want %q", ta.Text(), "AAABBBCCC")
	}
	if ta.LineCount() != 1 {
		t.Errorf("LineCount: got %d, want 1", ta.LineCount())
	}
}

func TestTextArea_HandleKeyNil(t *testing.T) {
	ta := NewTextArea()
	if ta.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}
