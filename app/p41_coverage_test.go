package app

import (
	"testing"

	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/internal/termcompat"
)

func makeTestClient() *ai.Client {
	return ai.NewClient(&ai.Config{
		APIKey:  "test-key",
		BaseURL: "http://localhost:8080",
		Model:   "test-model",
	})
}

func TestAIBridge_ClearHistory(t *testing.T) {
	app := NewChatApp(80, 24)
	b := NewAIBridge(app, makeTestClient())

	b.ClearHistory()

	msgs := b.Messages()
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(msgs))
	}
}

func TestAIBridge_Messages_Empty(t *testing.T) {
	app := NewChatApp(80, 24)
	b := NewAIBridge(app, makeTestClient())

	msgs := b.Messages()
	if len(msgs) != 0 {
		t.Errorf("expected 0 messages, got %d", len(msgs))
	}
}

func TestAIBridge_SetSystemPrompt(t *testing.T) {
	app := NewChatApp(80, 24)
	b := NewAIBridge(app, makeTestClient())
	b.SetSystemPrompt("You are a helpful assistant")
}

func TestAIBridge_SetOnError(t *testing.T) {
	app := NewChatApp(80, 24)
	b := NewAIBridge(app, makeTestClient())
	b.SetOnError(func(err error) {})
}

func TestChatApp_OnMouse(t *testing.T) {
	app := NewChatApp(80, 24)
	called := false
	app.OnMouse(func(me *term.MouseEvent) {
		called = true
	})
	if called {
		t.Error("handler should not be called yet")
	}
}

func TestChatApp_OnQuit(t *testing.T) {
	app := NewChatApp(80, 24)
	called := false
	app.OnQuit(func() {
		called = true
	})
	if called {
		t.Error("handler should not be called yet")
	}
}

func TestChatApp_OnClipboard(t *testing.T) {
	app := NewChatApp(80, 24)
	var received string
	app.OnClipboard(func(text string) {
		received = text
	})
	if received != "" {
		t.Error("handler should not be called yet")
	}
}

func TestP41_ClipboardConfig_Capabilities(t *testing.T) {
	caps := termcompat.Capabilities{HasOSC52: true}
	cc := NewClipboardWithCapabilities(caps)
	got := cc.Capabilities()
	if !got.HasOSC52 {
		t.Error("expected HasOSC52=true")
	}
}

func TestP41_ClipboardConfig_SetCapabilities(t *testing.T) {
	cc := NewClipboardWithCapabilities(termcompat.Capabilities{})
	cc.SetCapabilities(termcompat.Capabilities{HasOSC52: true})
	got := cc.Capabilities()
	if !got.HasOSC52 {
		t.Error("expected HasOSC52=true after SetCapabilities")
	}
}

func TestInputLine_Measure(t *testing.T) {
	il := NewInputLine("> ")
	il.InsertText("hello")
	sz := il.Measure(component.Constraints{MaxWidth: 80, MaxHeight: 1})
	if sz.H != 1 {
		t.Errorf("expected height 1, got %d", sz.H)
	}
	if sz.W <= 0 {
		t.Errorf("expected positive width, got %d", sz.W)
	}
}

func TestInputLine_Measure_MaxWidthClamped(t *testing.T) {
	il := NewInputLine("> ")
	il.InsertText("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
	sz := il.Measure(component.Constraints{MaxWidth: 10, MaxHeight: 1})
	if sz.W > 10 {
		t.Errorf("expected width <= 10, got %d", sz.W)
	}
}

func TestMouseHandler_Handle_Click(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	mh.Handle(&term.MouseEvent{
		X:      10,
		Y:      5,
		Action: term.MouseDown,
		Button: term.MouseLeft,
	})
}

func TestMouseHandler_Handle_WheelUp(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	mh.Handle(&term.MouseEvent{
		X:      10,
		Y:      5,
		Action: term.MouseWheel,
		Button: term.MouseWheelUp,
	})
}

func TestMouseHandler_Handle_WheelDown(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	mh.Handle(&term.MouseEvent{
		X:      10,
		Y:      5,
		Action: term.MouseWheel,
		Button: term.MouseWheelDown,
	})
}
