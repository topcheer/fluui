package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP300_NewChatComposer(t *testing.T) {
	c := NewChatComposer()
	if c.Text() != "" {
		t.Errorf("Text = %q, want empty", c.Text())
	}
	if c.ID() == "" {
		t.Error("ID should not be empty")
	}
	if c.MaxRows() != 5 {
		t.Errorf("MaxRows = %d, want 5", c.MaxRows())
	}
	if c.Placeholder() == "" {
		t.Error("placeholder should not be empty")
	}
	if c.Hint() == "" {
		t.Error("hint should not be empty")
	}
}

func TestP300_SetText(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hello world")
	if c.Text() != "hello world" {
		t.Errorf("Text = %q", c.Text())
	}
}

func TestP300_SetText_SlashMode(t *testing.T) {
	c := NewChatComposer()
	c.SetText("/help")
	if !c.SlashMode() {
		t.Error("should detect slash mode")
	}
	c.SetText("normal text")
	if c.SlashMode() {
		t.Error("should not be slash mode")
	}
}

func TestP300_Clear(t *testing.T) {
	c := NewChatComposer()
	c.SetText("some text")
	c.Clear()
	if c.Text() != "" {
		t.Errorf("Text = %q after clear", c.Text())
	}
	if c.SlashMode() {
		t.Error("slash mode should be false after clear")
	}
}

func TestP300_SetPlaceholder(t *testing.T) {
	c := NewChatComposer()
	c.SetPlaceholder("Ask anything…")
	if c.Placeholder() != "Ask anything…" {
		t.Errorf("Placeholder = %q", c.Placeholder())
	}
}

func TestP300_SetHint(t *testing.T) {
	c := NewChatComposer()
	c.SetHint("Press Enter")
	if c.Hint() != "Press Enter" {
		t.Errorf("Hint = %q", c.Hint())
	}
}

func TestP300_SetDisabled(t *testing.T) {
	c := NewChatComposer()
	c.SetDisabled(true)
	if !c.IsDisabled() {
		t.Error("should be disabled")
	}
	c.SetDisabled(false)
	if c.IsDisabled() {
		t.Error("should not be disabled")
	}
}

func TestP300_SetMaxRows(t *testing.T) {
	c := NewChatComposer()
	c.SetMaxRows(10)
	if c.MaxRows() != 10 {
		t.Errorf("MaxRows = %d, want 10", c.MaxRows())
	}
	// clamp
	c.SetMaxRows(0)
	if c.MaxRows() != 1 {
		t.Errorf("MaxRows = %d, want 1 (clamped)", c.MaxRows())
	}
}

func TestP300_SetTokenCount(t *testing.T) {
	c := NewChatComposer()
	c.SetTokenCount(150, 80)
	c.mu.Lock()
	if c.tokenIn != 150 || c.tokenOut != 80 {
		t.Errorf("tokens = %d/%d, want 150/80", c.tokenIn, c.tokenOut)
	}
	c.mu.Unlock()
}

func TestP300_HandleKey_Enter_Submit(t *testing.T) {
	c := NewChatComposer()
	submitted := ""
	c.SetOnSubmit(func(s string) { submitted = s })
	c.SetText("hello")

	ev := &term.KeyEvent{Key: term.KeyEnter}
	if !c.HandleKey(ev) {
		t.Error("Enter should be handled")
	}
	if submitted != "hello" {
		t.Errorf("submitted = %q, want 'hello'", submitted)
	}
	if c.Text() != "" {
		t.Error("text should be cleared after submit")
	}
}

func TestP300_HandleKey_Enter_NoCallback(t *testing.T) {
	c := NewChatComposer()
	c.SetText("test")
	ev := &term.KeyEvent{Key: term.KeyEnter}
	c.HandleKey(ev) // should not panic even without callback
	if c.Text() != "" {
		t.Error("should clear text even without callback")
	}
}

func TestP300_HandleKey_ShiftEnter_Newline(t *testing.T) {
	c := NewChatComposer()
	c.SetText("line1")
	ev := &term.KeyEvent{Key: term.KeyEnter, Modifiers: term.ModShift}
	c.HandleKey(ev)
	// Shift+Enter inserts a newline before the text (cursor at position 0)
	if !strings.Contains(c.Text(), "line1") {
		t.Errorf("text should contain 'line1', got %q", c.Text())
	}
	// Should have a newline
	if !strings.Contains(c.Text(), "\n") {
		t.Error("expected a newline in the text")
	}
}

func TestP300_HandleKey_Disabled(t *testing.T) {
	c := NewChatComposer()
	c.SetDisabled(true)
	ev := &term.KeyEvent{Key: term.KeyEnter}
	if c.HandleKey(ev) {
		t.Error("disabled composer should not handle keys")
	}
}

func TestP300_HandleKey_Nil(t *testing.T) {
	c := NewChatComposer()
	if c.HandleKey(nil) {
		t.Error("nil event should return false")
	}
}

func TestP300_HandleKey_Typing(t *testing.T) {
	c := NewChatComposer()
	ev := &term.KeyEvent{Key: 0, Rune: 'x'}
	c.HandleKey(ev)
	if c.Text() != "x" {
		t.Errorf("Text = %q, want 'x'", c.Text())
	}
}

func TestP300_Measure_Default(t *testing.T) {
	c := NewChatComposer()
	s := c.Measure(Constraints{MaxWidth: 80})
	// hint(1) + border(1) + 1 line(1) + border(1) = 4
	if s.H != 4 {
		t.Errorf("H = %d, want 4", s.H)
	}
	if s.W != 80 {
		t.Errorf("W = %d, want 80", s.W)
	}
}

func TestP300_Measure_AutoGrow(t *testing.T) {
	c := NewChatComposer()
	c.SetText("line1\nline2\nline3")
	s := c.Measure(Constraints{MaxWidth: 80})
	// hint(1) + border(1) + 3 lines + border(1) = 6
	if s.H != 6 {
		t.Errorf("H = %d, want 6", s.H)
	}
}

func TestP300_Measure_MaxRowsClamp(t *testing.T) {
	c := NewChatComposer()
	c.SetMaxRows(2)
	c.SetText("l1\nl2\nl3\nl4\nl5")
	s := c.Measure(Constraints{MaxWidth: 80})
	// hint(1) + border(1) + 2 lines (clamped) + border(1) = 5
	if s.H != 5 {
		t.Errorf("H = %d, want 5 (clamped to maxRows)", s.H)
	}
}

func TestP300_Measure_Disabled(t *testing.T) {
	c := NewChatComposer()
	c.SetDisabled(true)
	s := c.Measure(Constraints{MaxWidth: 80})
	// hint(1) + "Thinking"(1) = 2
	if s.H != 2 {
		t.Errorf("H = %d, want 2 (disabled)", s.H)
	}
}

func TestP300_Measure_DefaultWidth(t *testing.T) {
	c := NewChatComposer()
	s := c.Measure(Constraints{})
	if s.W != 80 {
		t.Errorf("default W = %d, want 80", s.W)
	}
}

func TestP300_Paint_Active(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hello world")
	c.SetTokenCount(100, 50)
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	c.Paint(buf)
}

func TestP300_Paint_EmptyPlaceholder(t *testing.T) {
	c := NewChatComposer()
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	c.Paint(buf)
}

func TestP300_Paint_Disabled(t *testing.T) {
	c := NewChatComposer()
	c.SetDisabled(true)
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 2})
	buf := buffer.NewBuffer(60, 2)
	c.Paint(buf)
}

func TestP300_Paint_SlashMode(t *testing.T) {
	c := NewChatComposer()
	c.SetText("/help me")
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	c.Paint(buf)
}

func TestP300_Paint_Multiline(t *testing.T) {
	c := NewChatComposer()
	c.SetText("line1\nline2\nline3")
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})
	buf := buffer.NewBuffer(60, 6)
	c.Paint(buf)
}

func TestP300_Paint_ZeroBounds(t *testing.T) {
	c := NewChatComposer()
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	c.Paint(buf) // should not panic
}

func TestP300_Paint_NarrowWidth(t *testing.T) {
	c := NewChatComposer()
	c.SetText("very long text")
	c.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 4})
	buf := buffer.NewBuffer(5, 4)
	c.Paint(buf)
}

func TestP300_Paint_NonZeroOffset(t *testing.T) {
	c := NewChatComposer()
	c.SetText("hello")
	c.SetBounds(Rect{X: 5, Y: 10, W: 50, H: 4})
	buf := buffer.NewBuffer(60, 15)
	c.Paint(buf)
}

func TestP300_Concurrent(t *testing.T) {
	c := NewChatComposer()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			c.SetText("update")
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		_ = c.Text()
		_ = c.IsDisabled()
	}
	<-done
}

func TestP300_SatisfiesComponent(t *testing.T) {
	// ChatComposer doesn't implement Paint directly as Component because it
	// has its own HandleKey, but it does implement Measure/Paint
	var _ interface {
		Measure(Constraints) Size
		Paint(*buffer.Buffer)
	} = (*ChatComposer)(nil)
}
