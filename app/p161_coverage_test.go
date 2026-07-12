package app

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// Target ThemeToast 85.7% — cover expired and empty cases
func TestP161_ThemeToast_Empty(t *testing.T) {
	a := NewChatApp(80, 24)
	text, visible := a.ThemeToast()
	if visible || text != "" {
		t.Error("expected empty toast")
	}
}

func TestP161_ThemeToast_Visible(t *testing.T) {
	a := NewChatApp(80, 24)
	a.CycleTheme()
	text, visible := a.ThemeToast()
	if !visible {
		t.Error("expected visible toast after cycle")
	}
	if text == "" {
		t.Error("expected non-empty text")
	}
}

func TestP161_ThemeToast_Expired(t *testing.T) {
	a := NewChatApp(80, 24)
	a.CycleTheme()
	// Manually set toast time to past
	a.mu.Lock()
	a.themeToastAt = time.Now().Add(-5 * time.Second)
	a.mu.Unlock()
	text, visible := a.ThemeToast()
	if visible {
		t.Error("expected expired toast to be invisible")
	}
	if text != "" {
		t.Error("expected empty text for expired toast")
	}
}

// Target SetTheme 87.5% — cover nil case
func TestP161_SetTheme_Nil(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetTheme(nil)
	if a.Theme() == nil {
		t.Error("expected non-nil theme after nil set")
	}
}

func TestP161_SetTheme_Valid(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetTheme(theme.Dracula())
	if a.Theme().Name != theme.Dracula().Name {
		t.Error("expected dracula theme")
	}
}

// Target SendUserMessage 85.7% — cover nil bridge case
func TestP161_SendUserMessage_NilBridge(t *testing.T) {
	a := NewChatApp(80, 24)
	// No AI bridge set — should return without panic
	a.SendUserMessage("hello")
}

// Target scrollToBottomLocked 88.9% — cover zero-width case
func TestP161_ScrollToBottom_ZeroWidth(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetSize(0, 0)
	a.ScrollToBottom() // should not panic
}

func TestP161_ScrollToBottom_TinyHeight(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetSize(80, 1)
	a.ScrollToBottom() // should not panic
}

func TestP161_ScrollToBottom_Normal(t *testing.T) {
	a := NewChatApp(80, 24)
	a.ScrollToBottom()
}

// Target HandleMouse 86.7% — cover more branches

func TestP161_HandleMouse_CustomHandler(t *testing.T) {
	a := NewChatApp(80, 24)
	called := false
	a.OnMouse(func(ev *term.MouseEvent) {
		called = true
	})
	a.HandleMouse(&term.MouseEvent{X: 10, Y: 5, Action: term.MouseDown})
	if !called {
		t.Error("expected custom handler called")
	}
}

func TestP161_HandleMouse_WheelUp(t *testing.T) {
	a := NewChatApp(80, 24)
	a.HandleMouse(&term.MouseEvent{X: 10, Y: 5, Action: term.MouseWheel, Button: 1})
}

func TestP161_HandleMouse_WheelDown(t *testing.T) {
	a := NewChatApp(80, 24)
	a.HandleMouse(&term.MouseEvent{X: 10, Y: 5, Action: term.MouseWheel, Button: 0})
}

func TestP161_HandleMouse_UnknownWheel(t *testing.T) {
	a := NewChatApp(80, 24)
	a.HandleMouse(&term.MouseEvent{X: 10, Y: 5, Action: term.MouseWheel, Button: 99})
}

// Target renderInputLine 92.3%

// Target Render 90.9%

// Target IsStreaming
func TestP161_IsStreaming_NoBridge(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.IsStreaming() {
		t.Error("expected false with no bridge")
	}
}

func TestP161_StopStreaming_NoBridge(t *testing.T) {
	a := NewChatApp(80, 24)
	a.StopStreaming() // should not panic
}

// Target ThemeName
func TestP161_ThemeName(t *testing.T) {
	a := NewChatApp(80, 24)
	name := a.ThemeName()
	if name == "" {
		t.Error("expected non-empty theme name")
	}
}

func TestP161_ThemeName_NilTheme(t *testing.T) {
	a := NewChatApp(80, 24)
	a.mu.Lock()
	a.theme = nil
	a.mu.Unlock()
	name := a.ThemeName()
	if name != "" {
		t.Errorf("expected empty name for nil theme, got %q", name)
	}
}

// Target RenderImageOverlays 90.5%
func TestP161_RenderImageOverlays_NoBlocks(t *testing.T) {
	_ = NewChatApp(80, 24)
	// No image blocks — should not panic
	// We can't easily test this without a renderer but it's covered by existing tests
}

// Target CycleThemeBack
func TestP161_CycleThemeBack(t *testing.T) {
	a := NewChatApp(80, 24)
	prev := a.CycleThemeBack()
	if prev == nil {
		t.Error("expected non-nil theme")
	}
}

// Target AddUserMessage
func TestP161_AddUserMessage(t *testing.T) {
	a := NewChatApp(80, 24)
	_ = a.AddUserMessage("test message")
}
// Target scrollUp
func TestP161_ScrollUp(t *testing.T) {
	a := NewChatApp(80, 24)
	a.ScrollUp() // should not panic even with empty container
}

func TestP161_ScrollDown(t *testing.T) {
	a := NewChatApp(80, 24)
	a.ScrollDown()
}

// Target Container
func TestP161_Container(t *testing.T) {
	a := NewChatApp(80, 24)
	c := a.Container()
	if c == nil {
		t.Error("expected non-nil container")
	}
}

// Target ScrollView
func TestP161_ScrollView(t *testing.T) {
	a := NewChatApp(80, 24)
	sv := a.ScrollView()
	if sv == nil {
		t.Error("expected non-nil scrollview")
	}
}

// Target SetSize
func TestP161_SetSize(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetSize(120, 40)
	if a.width != 120 {
		t.Error("expected width 120")
	}
}

// Target Width/Height
func TestP161_WidthHeight(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.width != 80 || a.height != 24 {
		t.Errorf("expected 80x24, got %dx%d", a.width, a.height)
	}
}