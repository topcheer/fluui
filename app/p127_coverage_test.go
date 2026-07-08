package app

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === scrollToBottomLocked edge cases (77.8% → 100%) ===
// The uncovered branch is contentH < 1 / contentW < 1 clamping

func TestP127_ScrollToBottom_ZeroWidth(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.SetSize(0, 0)
	// Should not panic with zero dimensions
	chat.mu.Lock()
	chat.scrollToBottomLocked()
	chat.mu.Unlock()
}

func TestP127_ScrollToBottom_TinyHeight(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.SetSize(80, 1) // height so small that contentH < 1 after input
	chat.mu.Lock()
	chat.scrollToBottomLocked()
	chat.mu.Unlock()
}

// === HandleMouseP16 (76.3% → 90%+) ===

func TestP127_HandleMouseP16_NilMouse(t *testing.T) {
	chat := NewChatApp(80, 24)
	result := chat.HandleMouseP16(nil)
	if result {
		t.Error("expected false for nil mouse")
	}
}

func TestP127_HandleMouseP16_CustomHandler(t *testing.T) {
	chat := NewChatApp(80, 24)
	called := false
	chat.OnMouse(func(m *term.MouseEvent) {
		called = true
	})
	mouse := &term.MouseEvent{X: 5, Y: 5, Action: term.MouseMove, Button: 0}
	chat.HandleMouseP16(mouse)
	if !called {
		t.Error("expected custom handler to be called")
	}
}

func TestP127_HandleMouseP16_UnknownWheelButton(t *testing.T) {
	chat := NewChatApp(80, 24)
	mouse := &term.MouseEvent{X: 5, Y: 5, Action: term.MouseWheel, Button: 99}
	result := chat.HandleMouseP16(mouse)
	if !result {
		t.Error("expected true for wheel event (falls to custom handler)")
	}
}

// === copySelectionToWriter (70.0% → 90%+) ===

func TestP127_CopySelection_NilSelectionMgr(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	result := chat.copySelectionToWriter(buf, &w)
	if result {
		t.Error("expected false with nil selectionMgr")
	}
}

func TestP127_CopySelection_NilBuf(t *testing.T) {
	chat := NewChatApp(80, 24)
	var w bytes.Buffer
	result := chat.copySelectionToWriter(nil, &w)
	if result {
		t.Error("expected false with nil buf")
	}
}

func TestP127_CopySelection_NilWriter(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	result := chat.copySelectionToWriter(buf, nil)
	if result {
		t.Error("expected false with nil writer")
	}
}

func TestP127_CopySelection_NoSelection(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.selectionMgr = NewSelectionManager()
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	result := chat.copySelectionToWriter(buf, &w)
	if result {
		t.Error("expected false with no selection")
	}
}

func TestP127_CopySelection_WithSelectionOSC52(t *testing.T) {
	chat := NewChatApp(80, 24)
	sm := NewSelectionManager()
	chat.selectionMgr = sm
	buf := buffer.NewBuffer(80, 24)
	// Put some text in buffer
	buf.DrawText(0, 0, "Hello World", buffer.Style{})
	// Select word
	sm.SelectWord(2, 0, buf)
	var w bytes.Buffer
	result := chat.copySelectionToWriter(buf, &w)
	// Should succeed with OSC52 fallback
	if !result {
		t.Error("expected true with selection and no clipboardConfig")
	}
	if w.Len() == 0 {
		t.Error("expected output in writer")
	}
}
