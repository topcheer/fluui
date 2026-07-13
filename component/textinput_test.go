package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP164_NewTextInput(t *testing.T) {
	ti := NewTextInput()
	if ti == nil { t.Fatal("expected non-nil") }
}

func TestP164_Value_SetValue(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	if ti.Value() != "hello" { t.Errorf("expected 'hello', got %q", ti.Value()) }
}

func TestP164_InsertText(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	ti.SetCursor(2)
	ti.InsertText("XX")
	if ti.Value() != "heXXllo" { t.Errorf("got %q", ti.Value()) }
	if ti.Cursor() != 4 { t.Errorf("got cursor %d", ti.Cursor()) }
}

func TestP164_CharLimit(t *testing.T) {
	ti := NewTextInput()
	ti.SetCharLimit(5)
	ti.SetValue("123456789")
	if ti.Value() != "12345" { t.Errorf("expected '12345', got %q", ti.Value()) }
	ti.SetCursor(5)
	ti.InsertText("X") // should be truncated
	if ti.Value() != "12345" { t.Errorf("expected still '12345', got %q", ti.Value()) }
}

func TestP164_Clear(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("test")
	ti.Clear()
	if ti.Value() != "" || !ti.Empty() { t.Error("expected empty") }
}

func TestP164_Cursor(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	ti.SetCursor(3)
	if ti.Cursor() != 3 { t.Errorf("got %d", ti.Cursor()) }
	ti.SetCursor(-1)
	if ti.Cursor() != 0 { t.Error("expected 0") }
	ti.SetCursor(99)
	if ti.Cursor() != 5 { t.Error("expected 5") }
	ti.CursorEnd()
	if ti.Cursor() != 5 { t.Error("expected 5") }
	ti.CursorStart()
	if ti.Cursor() != 0 { t.Error("expected 0") }
}

func TestP164_Prompt(t *testing.T) {
	ti := NewTextInput()
	ti.SetPrompt("> ")
	if ti.Prompt() != "> " { t.Errorf("got %q", ti.Prompt()) }
}

func TestP164_Placeholder(t *testing.T) {
	ti := NewTextInput()
	ti.SetPlaceholder("Enter text...")
	if ti.Placeholder() != "Enter text..." { t.Errorf("got %q", ti.Placeholder()) }
}

func TestP164_EchoPassword(t *testing.T) {
	ti := NewTextInput()
	ti.EchoPassword()
	if ti.EchoMode() != EchoPassword { t.Error("expected EchoPassword") }
	ti.SetValue("secret")
	if ti.displayValue() != "******" { t.Errorf("expected '******', got %q", ti.displayValue()) }
}

func TestP164_EchoNone(t *testing.T) {
	ti := NewTextInput()
	ti.SetEchoMode(EchoNone)
	ti.SetValue("hidden")
	if ti.displayValue() != "" { t.Error("expected empty display") }
}

func TestP164_FocusBlur(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	if !ti.Focused() { t.Error("expected focused") }
	ti.Blur()
	if ti.Focused() { t.Error("expected not focused") }
}

func TestP164_Width(t *testing.T) {
	ti := NewTextInput()
	ti.SetWidth(40)
	if ti.Width() != 40 { t.Errorf("got %d", ti.Width()) }
	ti.SetWidth(0)
	if ti.Width() != 1 { t.Error("expected min 1") }
}

func TestP164_HandleKey(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	// Type characters
	if !ti.HandleKey(&term.KeyEvent{Rune: 'a', Key: term.KeyUnknown}) { t.Error("expected handled") }
	if !ti.HandleKey(&term.KeyEvent{Rune: 'b', Key: term.KeyUnknown}) { t.Error("expected handled") }
	if ti.Value() != "ab" { t.Errorf("got %q", ti.Value()) }
	// Backspace
	if !ti.HandleKey(&term.KeyEvent{Key: term.KeyBackspace}) { t.Error("expected handled") }
	if ti.Value() != "a" { t.Errorf("got %q", ti.Value()) }
	// Left
	ti.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if ti.Cursor() != 0 { t.Error("expected 0") }
	// Right
	ti.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if ti.Cursor() != 1 { t.Error("expected 1") }
	// Home
	ti.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if ti.Cursor() != 0 { t.Error("expected 0") }
	// End
	ti.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if ti.Cursor() != 1 { t.Error("expected 1") }
	// Delete (at end, should not change)
	ti.HandleKey(&term.KeyEvent{Key: term.KeyDelete})
	if ti.Value() != "a" { t.Errorf("got %q", ti.Value()) }
}

func TestP164_HandleKey_NilKey(t *testing.T) {
	ti := NewTextInput()
	if ti.HandleKey(nil) { t.Error("expected false for nil key") }
}

func TestP164_HandleKey_Enter(t *testing.T) {
	ti := NewTextInput()
	called := false
	ti.SetOnSubmit(func(s string) { called = true; if s != "test" { t.Errorf("expected 'test', got %q", s) } })
	ti.SetValue("test")
	ti.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !called { t.Error("expected submit called") }
}

func TestP164_OnChange(t *testing.T) {
	ti := NewTextInput()
	changes := 0
	ti.SetOnChange(func(s string) { changes++ })
	ti.SetValue("a")
	ti.InsertText("b")
	if changes != 2 { t.Errorf("expected 2 changes, got %d", changes) }
}

func TestP164_History(t *testing.T) {
	ti := NewTextInput()
	ti.AddHistory("cmd1")
	ti.AddHistory("cmd2")
	h := ti.History()
	if len(h) != 2 { t.Errorf("expected 2, got %d", len(h)) }
	ti.SetHistory([]string{"x", "y", "z"})
	if len(ti.History()) != 3 { t.Error("expected 3") }
	// Navigate up (older)
	ti.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if ti.Value() != "z" { t.Errorf("expected 'z', got %q", ti.Value()) }
	// Navigate up again
	ti.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if ti.Value() != "y" { t.Errorf("expected 'y', got %q", ti.Value()) }
	// Navigate down (newer)
	ti.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if ti.Value() != "z" { t.Errorf("expected 'z', got %q", ti.Value()) }
	// Navigate down past end → empty
	ti.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if ti.Value() != "" { t.Errorf("expected '', got %q", ti.Value()) }
}

func TestP164_Position(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	ti.SetCursor(3)
	line, col := ti.Position()
	if line != 0 || col != 3 { t.Errorf("expected (0,3), got (%d,%d)", line, col) }
}

func TestP164_Paint(t *testing.T) {
	ti := NewTextInput()
	ti.SetPrompt("> ")
	ti.SetValue("hello")
	ti.Focus()
	ti.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	ti.Paint(buf)
}

func TestP164_Paint_Placeholder(t *testing.T) {
	ti := NewTextInput()
	ti.SetPrompt("> ")
	ti.SetPlaceholder("type here")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	ti.Paint(buf)
}

func TestP164_Paint_Password(t *testing.T) {
	ti := NewTextInput()
	ti.EchoPassword()
	ti.SetValue("secret")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	ti.Paint(buf)
}

func TestP164_Paint_NilBuf(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("test")
	ti.Paint(nil) // should not panic
}

func TestP164_Measure(t *testing.T) {
	ti := NewTextInput()
	ti.SetWidth(30)
	size := ti.Measure(Bounded(50, 10))
	if size.H != 1 { t.Errorf("expected H=1, got %d", size.H) }
	if size.W != 30 { t.Errorf("expected W=30, got %d", size.W) }
}

func TestP164_Measure_ClampMaxWidth(t *testing.T) {
	ti := NewTextInput()
	ti.SetWidth(100)
	size := ti.Measure(Bounded(50, 10))
	if size.W != 50 { t.Errorf("expected W=50, got %d", size.W) }
}

func TestP164_SetBounds(t *testing.T) {
	ti := NewTextInput()
	ti.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	if ti.Width() != 40 { t.Error("expected 40") }
}

func TestP164_Children(t *testing.T) {
	ti := NewTextInput()
	if ti.Children() != nil { t.Error("expected nil") }
}

func TestP164_CtrlLeftRight(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello world foo")
	ti.CursorEnd()
	// Ctrl+Left: move to start of "foo"
	ti.HandleKey(&term.KeyEvent{Key: term.KeyLeft, Modifiers: term.ModCtrl})
	if ti.Cursor() != 12 { t.Errorf("expected 11, got %d", ti.Cursor()) }
	// Ctrl+Left again: move to start of "world"
	ti.HandleKey(&term.KeyEvent{Key: term.KeyLeft, Modifiers: term.ModCtrl})
	if ti.Cursor() != 6 { t.Errorf("expected 6, got %d", ti.Cursor()) }
	// Ctrl+Right: move to start of "foo"
	ti.HandleKey(&term.KeyEvent{Key: term.KeyRight, Modifiers: term.ModCtrl})
	if ti.Cursor() != 11 { t.Errorf("expected 11, got %d", ti.Cursor()) }
}

func TestP164_Style(t *testing.T) {
	ti := NewTextInput()
	ti.SetStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedRed)})
	ti.SetPromptStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlue)})
	ti.SetPlaceholderStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow)})
	ti.SetFocusedStyle(buffer.Style{Fg: buffer.NamedColor(buffer.NamedGreen)})
}

func TestP164_Blink(t *testing.T) {
	ti := NewTextInput()
	ti.Focus()
	_ = ti.Blink() // should return some bool
	ti.SetBlink(false)
	ti.Blur()
	_ = ti.Blink()
}

func TestP164_String(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("test")
	if ti.String() != "test" { t.Errorf("got %q", ti.String()) }
}

func TestP164_SetCursor_MidText(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	ti.SetCursor(2)
	ti.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	if ti.Value() != "hllo" { t.Errorf("got %q", ti.Value()) }
	if ti.Cursor() != 1 { t.Errorf("got %d", ti.Cursor()) }
}

func TestP164_Delete_MidText(t *testing.T) {
	ti := NewTextInput()
	ti.SetValue("hello")
	ti.SetCursor(2)
	ti.HandleKey(&term.KeyEvent{Key: term.KeyDelete})
	if ti.Value() != "helo" { t.Errorf("got %q", ti.Value()) }
	if ti.Cursor() != 2 { t.Errorf("got %d", ti.Cursor()) }
}