package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestChatAppCreation(t *testing.T) {
	app := NewChatApp(80, 24)
	if app == nil {
		t.Fatal("NewChatApp returned nil")
	}
	w, h := app.Size()
	if w != 80 || h != 24 {
		t.Errorf("Size() = %d,%d want 80,24", w, h)
	}
	if app.Container() == nil {
		t.Error("Container() should not be nil")
	}
	if app.Overlays() == nil {
		t.Error("Overlays() should not be nil")
	}
}

func TestChatAppAddUserMessage(t *testing.T) {
	app := NewChatApp(80, 24)
	msg := app.AddUserMessage("Hello world")
	if msg == nil {
		t.Fatal("AddUserMessage returned nil")
	}
	if msg.Content() != "Hello world" {
		t.Errorf("Content() = %q, want 'Hello world'", msg.Content())
	}
	if app.Container().Len() != 1 {
		t.Errorf("Container.Len() = %d, want 1", app.Container().Len())
	}
}

func TestChatAppAddThinking(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := app.AddThinking()
	if tb == nil {
		t.Fatal("AddThinking returned nil")
	}
	if tb.Type() != block.TypeThinking {
		t.Errorf("Type() = %v, want TypeThinking", tb.Type())
	}
}

func TestChatAppAddAssistantText(t *testing.T) {
	app := NewChatApp(80, 24)
	at := app.AddAssistantText()
	if at == nil {
		t.Fatal("AddAssistantText returned nil")
	}
	if at.Type() != block.TypeAssistantText {
		t.Errorf("Type() = %v, want TypeAssistantText", at.Type())
	}
}

func TestChatAppAddToolCall(t *testing.T) {
	app := NewChatApp(80, 24)
	tc := app.AddToolCall("read_file", `{path:"foo.go"}`)
	if tc == nil {
		t.Fatal("AddToolCall returned nil")
	}
	if tc.ToolName() != "read_file" {
		t.Errorf("ToolName() = %q", tc.ToolName())
	}
}

func TestChatAppAddToolResult(t *testing.T) {
	app := NewChatApp(80, 24)
	tr := app.AddToolResult()
	if tr == nil {
		t.Fatal("AddToolResult returned nil")
	}
	if tr.Type() != block.TypeToolResult {
		t.Errorf("Type() = %v, want TypeToolResult", tr.Type())
	}
}

func TestChatAppStreamDelta(t *testing.T) {
	app := NewChatApp(80, 24)
	app.StreamDelta(block.StreamDelta{Type: "text", Content: "Hello"})
	app.StreamDelta(block.StreamDelta{Type: "text", Content: " world"})

	// Should have at least 1 block
	if app.Container().Len() < 1 {
		t.Error("StreamDelta should create blocks")
	}
}

func TestChatAppRender(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("Test message")
	app.AddAssistantText()

	buf := buffer.NewBuffer(80, 24)
	app.Render(buf) // should not panic

	// Buffer should have some content
	hasContent := false
	for x := 0; x < 80; x++ {
		for y := 0; y < 24; y++ {
			if c := buf.GetCell(x, y); c.Rune != 0 && c.Rune != ' ' {
				hasContent = true
				break
			}
		}
		if hasContent {
			break
		}
	}
	if !hasContent {
		t.Error("Render should write content to buffer")
	}
}

func TestChatAppHandleKeyQuit(t *testing.T) {
	app := NewChatApp(80, 24)
	quitCalled := false
	app.OnQuit(func() { quitCalled = true })

	app.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !quitCalled {
		t.Error("Escape should trigger quit")
	}
}

func TestChatAppHandleKeyScroll(t *testing.T) {
	app := NewChatApp(80, 24)

	// Add enough content to scroll
	for i := 0; i < 30; i++ {
		app.AddUserMessage("Line " + string(rune('A'+i%26)))
	}

	// Down should not panic
	app.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	// Up should not panic
	app.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

func TestChatAppHandleMouseWheel(t *testing.T) {
	app := NewChatApp(80, 24)

	// Wheel down should not panic
	consumed := app.HandleMouse(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: term.MouseWheelDown,
		X:      10, Y: 10,
	})
	if !consumed {
		t.Error("Mouse wheel should be consumed")
	}

	// Wheel up
	consumed = app.HandleMouse(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: term.MouseWheelUp,
		X:      10, Y: 10,
	})
	if !consumed {
		t.Error("Mouse wheel should be consumed")
	}
}

func TestChatAppSetSize(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetSize(120, 40)
	w, h := app.Size()
	if w != 120 || h != 40 {
		t.Errorf("Size() = %d,%d want 120,40", w, h)
	}
}

func TestChatAppClear(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("msg1")
	app.AddUserMessage("msg2")
	app.AddUserMessage("msg3")

	if app.Container().Len() != 3 {
		t.Fatalf("Len = %d, want 3", app.Container().Len())
	}

	app.Clear()

	if app.Container().Len() != 0 {
		t.Errorf("Len after Clear = %d, want 0", app.Container().Len())
	}
}

func TestChatAppSetInputHeight(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetInputHeight(2)

	buf := buffer.NewBuffer(80, 24)
	app.AddUserMessage("Test")
	app.Render(buf) // should render input line without panic
}

func TestChatAppIsDirty(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("test")
	if !app.IsDirty() {
		t.Error("should be dirty after AddUserMessage")
	}
	app.ClearDirty()
	if app.IsDirty() {
		t.Error("should not be dirty after ClearDirty")
	}
}
