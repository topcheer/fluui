package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestNewTerminalPanel(t *testing.T) {
	tp := NewTerminalPanel(100)
	if tp == nil {
		t.Fatal("NewTerminalPanel returned nil")
	}
	if tp.LineCount() != 0 {
		t.Errorf("LineCount = %d, want 0", tp.LineCount())
	}
	if tp.maxLines != 100 {
		t.Errorf("maxLines = %d, want 100", tp.maxLines)
	}
}

func TestNewTerminalPanel_DefaultMaxLines(t *testing.T) {
	tp := NewTerminalPanel(0)
	if tp.maxLines != 1000 {
		t.Errorf("maxLines = %d, want 1000 (default)", tp.maxLines)
	}
}

func TestTerminalPanel_WritePlain(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("Hello\nWorld")

	if tp.LineCount() != 2 {
		t.Errorf("LineCount = %d, want 2", tp.LineCount())
	}

	lines := tp.Lines()
	if len(lines) != 2 {
		t.Fatalf("len(lines) = %d, want 2", len(lines))
	}
	if len(lines[0].Cells) != 5 {
		t.Errorf("line 0 cells = %d, want 5", len(lines[0].Cells))
	}
}

func TestTerminalPanel_WriteSGR(t *testing.T) {
	tp := NewTerminalPanel(100)
	// Red foreground
	tp.WriteString("\x1b[31mRed\x1b[0m")

	lines := tp.Lines()
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	if len(lines[0].Cells) != 3 {
		t.Errorf("cells = %d, want 3", len(lines[0].Cells))
	}

	// First cell should have red fg
	c := lines[0].Cells[0]
	if c.Fg.Type != buffer.ColorNamed {
		t.Errorf("Fg.Type = %v, want ColorNamed", c.Fg.Type)
	}
	if c.Fg.Val != 1 { // red = 30-30=1
		t.Errorf("Fg.Val = %d, want 1 (red)", c.Fg.Val)
	}

	// Check rune is correct
	if c.Rune != 'R' {
		t.Errorf("Rune = %c, want 'R'", c.Rune)
	}
}

func TestTerminalPanel_WriteTrueColor(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[38;2;255;128;0mX\x1b[0m")

	lines := tp.Lines()
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	c := lines[0].Cells[0]
	if c.Fg.Type != buffer.ColorTrue {
		t.Errorf("Fg.Type = %v, want ColorTrue", c.Fg.Type)
	}
}

func TestTerminalPanel_Write256Color(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[38;5;196mX\x1b[0m")

	lines := tp.Lines()
	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	c := lines[0].Cells[0]
	if c.Fg.Type != buffer.Color256 {
		t.Errorf("Fg.Type = %v, want Color256", c.Fg.Type)
	}
}

func TestTerminalPanel_WriteBold(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[1mBold\x1b[0m")

	lines := tp.Lines()
	c := lines[0].Cells[0]
	if c.Flags&buffer.Bold == 0 {
		t.Error("Bold flag not set")
	}
}

func TestTerminalPanel_WriteUnderline(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[4mU\x1b[24m")

	lines := tp.Lines()
	c := lines[0].Cells[0]
	if c.Flags&buffer.Underline == 0 {
		t.Error("Underline flag not set")
	}
}

func TestTerminalPanel_WriteMultipleAttrs(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[1;3;4mB\x1b[0m")

	lines := tp.Lines()
	c := lines[0].Cells[0]
	if c.Flags&buffer.Bold == 0 {
		t.Error("Bold not set")
	}
	if c.Flags&buffer.Italic == 0 {
		t.Error("Italic not set")
	}
	if c.Flags&buffer.Underline == 0 {
		t.Error("Underline not set")
	}
}

func TestTerminalPanel_WriteBrightColors(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[91mX") // bright red

	lines := tp.Lines()
	c := lines[0].Cells[0]
	if c.Fg.Type != buffer.ColorNamed {
		t.Error("expected ColorNamed")
	}
	if c.Fg.Val != 9 { // bright red = 91-90+8 = 9
		t.Errorf("Fg.Val = %d, want 9", c.Fg.Val)
	}
}

func TestTerminalPanel_WriteBgColor(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[42mX") // green bg

	lines := tp.Lines()
	c := lines[0].Cells[0]
	if c.Bg.Type != buffer.ColorNamed {
		t.Error("expected ColorNamed for Bg")
	}
}

func TestTerminalPanel_WriteTabs(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("A\tB")

	lines := tp.Lines()
	// A + 4 spaces + B = 6 cells
	if len(lines[0].Cells) != 6 {
		t.Errorf("cells = %d, want 6 (A + 4 spaces + B)", len(lines[0].Cells))
	}
}

func TestTerminalPanel_WriteControlChars(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x00\x01\x02visible")

	lines := tp.Lines()
	if len(lines[0].Cells) != 7 {
		t.Errorf("cells = %d, want 7 (control chars ignored)", len(lines[0].Cells))
	}
}

func TestTerminalPanel_WriteOSC(t *testing.T) {
	tp := NewTerminalPanel(100)
	// OSC title sequence should be skipped
	tp.WriteString("\x1b]0;Title\x07Hello")

	lines := tp.Lines()
	if len(lines[0].Cells) != 5 {
		t.Errorf("cells = %d, want 5 (OSC skipped)", len(lines[0].Cells))
	}
}

func TestTerminalPanel_WriteOSCESCBackslash(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b]0;Title\x1b\\Hello")

	lines := tp.Lines()
	if len(lines[0].Cells) != 5 {
		t.Errorf("cells = %d, want 5", len(lines[0].Cells))
	}
}

func TestTerminalPanel_WriteResetAll(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("\x1b[1;31mBold\x1b[0mNormal")

	lines := tp.Lines()
	if len(lines[0].Cells) != 10 {
		t.Errorf("cells = %d, want 10", len(lines[0].Cells))
	}
	// After reset, flags should be clear
	c := lines[0].Cells[4] // 'N'
	if c.Flags != 0 {
		t.Error("Flags should be 0 after reset")
	}
}

func TestTerminalPanel_MaxLines(t *testing.T) {
	tp := NewTerminalPanel(5)
	for i := 0; i < 10; i++ {
		tp.WriteString("Line\n")
	}
	if tp.LineCount() > 5 {
		t.Errorf("LineCount = %d, should be <= 5", tp.LineCount())
	}
}

func TestTerminalPanel_SetMaxLines(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 10; i++ {
		tp.WriteString("Line\n")
	}
	tp.SetMaxLines(3)
	if tp.LineCount() > 3 {
		t.Errorf("LineCount = %d, should be <= 3 after SetMaxLines", tp.LineCount())
	}
}

func TestTerminalPanel_SetMaxLinesZero(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.SetMaxLines(0)
	if tp.maxLines != 1000 {
		t.Errorf("maxLines = %d, want 1000 (default)", tp.maxLines)
	}
}

func TestTerminalPanel_Clear(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("Hello\n")
	tp.Clear()
	if tp.LineCount() != 0 {
		t.Errorf("LineCount = %d after Clear, want 0", tp.LineCount())
	}
}

func TestTerminalPanel_ScrollUp(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 20; i++ {
		tp.WriteString("Line\n")
	}
	tp.ScrollUp(5)
	if tp.ScrollOffset() != 5 {
		t.Errorf("ScrollOffset = %d, want 5", tp.ScrollOffset())
	}
}

func TestTerminalPanel_ScrollDown(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 20; i++ {
		tp.WriteString("Line\n")
	}
	tp.ScrollUp(10)
	tp.ScrollDown(3)
	if tp.ScrollOffset() != 7 {
		t.Errorf("ScrollOffset = %d, want 7", tp.ScrollOffset())
	}
}

func TestTerminalPanel_ScrollToBottom(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 20; i++ {
		tp.WriteString("Line\n")
	}
	tp.ScrollUp(10)
	tp.ScrollToBottom()
	if tp.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %d, want 0", tp.ScrollOffset())
	}
}

func TestTerminalPanel_ScrollClamp(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 10; i++ {
		tp.WriteString("Line\n")
	}
	tp.ScrollUp(100) // way past max
	if tp.ScrollOffset() > 10 {
		t.Errorf("ScrollOffset = %d, should be clamped", tp.ScrollOffset())
	}

	tp.ScrollDown(100) // way past min
	if tp.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %d, should be clamped to 0", tp.ScrollOffset())
	}
}

func TestTerminalPanel_Paint(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("Hello World\nSecond line")

	buf := buffer.NewBuffer(40, 10)
	buf.Fill(buffer.BlankCell)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	tp.Paint(buf)

	// Should have rendered "Hello World" on first row
	c := buf.GetCell(0, 0)
	if c.Rune != 'H' {
		t.Errorf("cell(0,0) = %q, want 'H'", c.Rune)
	}
	c = buf.GetCell(1, 0)
	if c.Rune != 'e' {
		t.Errorf("cell(1,0) = %q, want 'e'", c.Rune)
	}
}

func TestTerminalPanel_PaintNilBuffer(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.Paint(nil) // should not panic
}

func TestTerminalPanel_PaintZeroBounds(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("Hello")
	buf := buffer.NewBuffer(40, 10)
	tp.SetBounds(Rect{0, 0, 0, 0})
	tp.Paint(buf) // should not crash
}

func TestTerminalPanel_PaintScrollOffset(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 20; i++ {
		tp.WriteString("L" + string(rune('0'+i)) + "\n")
	}
	tp.ScrollUp(5)

	buf := buffer.NewBuffer(40, 10)
	buf.Fill(buffer.BlankCell)
	tp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	tp.Paint(buf)

	// Should show lines before the last 5 (scrolled back)
	c := buf.GetCell(0, 0)
	if c.Rune == 0 || c.Rune == ' ' {
		t.Error("Paint with scroll offset should show content")
	}
}

func TestTerminalPanel_Measure(t *testing.T) {
	tp := NewTerminalPanel(100)
	s := tp.Measure(Constraints{MaxWidth: 120, MaxHeight: 40})
	if s.W != 80 {
		t.Errorf("W = %d, want 80", s.W)
	}
	if s.H != 24 {
		t.Errorf("H = %d, want 24", s.H)
	}
}

func TestTerminalPanel_HandleKeyScroll(t *testing.T) {
	tp := NewTerminalPanel(100)
	for i := 0; i < 20; i++ {
		tp.WriteString("Line\n")
	}

	// Up should scroll
	tp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if tp.ScrollOffset() != 1 {
		t.Errorf("ScrollOffset = %d after KeyUp, want 1", tp.ScrollOffset())
	}

	// PageUp should scroll more
	tp.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if tp.ScrollOffset() < 5 {
		t.Errorf("ScrollOffset = %d after PageUp, want >= 5", tp.ScrollOffset())
	}

	// Down should scroll back
	off := tp.ScrollOffset()
	tp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if tp.ScrollOffset() != off-1 {
		t.Errorf("ScrollOffset = %d after Down, want %d", tp.ScrollOffset(), off-1)
	}

	// End should reset
	tp.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if tp.ScrollOffset() != 0 {
		t.Errorf("ScrollOffset = %d after End, want 0", tp.ScrollOffset())
	}

	// Home should scroll to top
	tp.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if tp.ScrollOffset() == 0 {
		t.Error("ScrollOffset should be > 0 after Home")
	}
}

func TestTerminalPanel_HandleKeyInput(t *testing.T) {
	tp := NewTerminalPanel(100)

	var received string
	tp.SetOnInput(func(s string) {
		received = s
	})

	// Type 'hi'
	tp.HandleKey(&term.KeyEvent{Rune: 'h'})
	tp.HandleKey(&term.KeyEvent{Rune: 'i'})

	// Enter should trigger callback
	tp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	if received != "hi" {
		t.Errorf("received = %q, want 'hi'", received)
	}
}

func TestTerminalPanel_HandleKeyBackspace(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.HandleKey(&term.KeyEvent{Rune: 'a'})
	tp.HandleKey(&term.KeyEvent{Rune: 'b'})
	tp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})

	tp.mu.RLock()
	if len(tp.inputBuf) != 1 {
		t.Errorf("inputBuf len = %d, want 1 after backspace", len(tp.inputBuf))
	}
	tp.mu.RUnlock()
}

func TestTerminalPanel_HandleKeyNil(t *testing.T) {
	tp := NewTerminalPanel(100)
	if tp.HandleKey(nil) {
		t.Error("HandleKey(nil) should return false")
	}
}

func TestTerminalPanel_HandleKeyOnKey(t *testing.T) {
	tp := NewTerminalPanel(100)
	var called bool
	tp.OnKey = func(k *term.KeyEvent) bool {
		called = true
		return true
	}
	tp.HandleKey(&term.KeyEvent{Key: 0xFF}) // unknown key
	if !called {
		t.Error("OnKey callback not called")
	}
}

func TestTerminalPanel_Children(t *testing.T) {
	tp := NewTerminalPanel(100)
	if tp.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestTerminalPanel_SetOnInput(t *testing.T) {
	tp := NewTerminalPanel(100)
	var called bool
	tp.SetOnInput(func(s string) {
		called = true
	})
	tp.HandleKey(&term.KeyEvent{Rune: 'x'})
	tp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !called {
		t.Error("OnInput callback not called on Enter")
	}
}

func TestTerminalPanel_Lines(t *testing.T) {
	tp := NewTerminalPanel(100)
	tp.WriteString("A\nB\nC")

	lines := tp.Lines()
	if len(lines) != 3 {
		t.Fatalf("len(lines) = %d, want 3", len(lines))
	}
	// Verify it's a copy (modifying shouldn't affect original)
	lines[0] = nil
	orig := tp.Lines()
	if orig[0] == nil {
		t.Error("Lines() should return a defensive copy")
	}
}

func TestTerminalPanel_Concurrent(t *testing.T) {
	tp := NewTerminalPanel(100)
	done := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			tp.WriteString("concurrent write\n")
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			buf := buffer.NewBuffer(40, 10)
			buf.Fill(buffer.BlankCell)
			tp.SetBounds(Rect{0, 0, 40, 10})
			tp.Paint(buf)
		}
		done <- struct{}{}
	}()

	go func() {
		for i := 0; i < 100; i++ {
			tp.LineCount()
			tp.Lines()
		}
		done <- struct{}{}
	}()

	<-done
	<-done
	<-done
}

func TestTerminalPanel_WriteLargeContent(t *testing.T) {
	tp := NewTerminalPanel(50)
	// Write many lines
	for i := 0; i < 100; i++ {
		tp.WriteString(strings.Repeat("x", 80) + "\n")
	}
	// Should be trimmed to maxLines
	if tp.LineCount() > 50 {
		t.Errorf("LineCount = %d, should be <= 50", tp.LineCount())
	}
}

func TestTerminalPanel_WritePartialSGR(t *testing.T) {
	tp := NewTerminalPanel(100)
	// Incomplete escape at end
	tp.WriteString("\x1b[31mRe")
	// Reset
	tp.WriteString("d\x1b[0m")

	lines := tp.Lines()
	if len(lines[0].Cells) != 3 {
		t.Errorf("cells = %d, want 3", len(lines[0].Cells))
	}
}
