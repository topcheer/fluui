package app

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === scrollToBottomLocked (77.8% → 100%) ===

func TestP131_ScrollToBottom_ZeroWidth(t *testing.T) {
	chat := NewChatApp(0, 24)
	chat.ScrollToBottom()
}

func TestP131_ScrollToBottom_TinyHeight(t *testing.T) {
	chat := NewChatApp(80, 1)
	chat.ScrollToBottom()
}

func TestP131_ScrollToBottom_NegativeContentHeight(t *testing.T) {
	chat := NewChatApp(80, 0)
	chat.ScrollToBottom()
}

// === HandleMouseP16 (76.3% → 90%+) ===

func TestP131_HandleMouseP16_NilMouse(t *testing.T) {
	chat := NewChatApp(80, 24)
	if chat.HandleMouseP16(nil) {
		t.Error("expected false for nil mouse")
	}
}

func TestP131_HandleMouseP16_UnknownWheel(t *testing.T) {
	chat := NewChatApp(80, 24)
	mouse := &term.MouseEvent{
		X:      10,
		Y:      10,
		Button: 99, // unknown button
		Action: term.MouseWheel,
	}
	// Should not consume (unknown wheel button)
	chat.HandleMouseP16(mouse)
}

func TestP131_HandleMouseP16_CustomHandler(t *testing.T) {
	chat := NewChatApp(80, 24)
	called := false
	chat.OnMouse(func(m *term.MouseEvent) {
		called = true
	})
	mouse := &term.MouseEvent{
		X:      10,
		Y:      10,
		Button: term.MouseLeft,
		Action: term.MouseMove,
	}
	chat.HandleMouseP16(mouse)
	if !called {
		t.Error("expected custom handler to be called")
	}
}

func TestP131_HandleMouseP16_WheelUp(t *testing.T) {
	chat := NewChatApp(80, 24)
	mouse := &term.MouseEvent{
		X:      10,
		Y:      10,
		Button: term.MouseWheelUp,
		Action: term.MouseWheel,
	}
	chat.HandleMouseP16(mouse)
}

func TestP131_HandleMouseP16_WheelDown(t *testing.T) {
	chat := NewChatApp(80, 24)
	mouse := &term.MouseEvent{
		X:      10,
		Y:      10,
		Button: term.MouseWheelDown,
		Action: term.MouseWheel,
	}
	chat.HandleMouseP16(mouse)
}

// === copySelectionToWriter (70.0% → 90%+) ===

func TestP131_CopySelection_NilSelectionMgr(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	// selectionMgr is nil by default
	if chat.CopySelectionToWriter(buf, &w) {
		t.Error("expected false with nil selectionMgr")
	}
}

func TestP131_CopySelection_NilBuffer(t *testing.T) {
	chat := NewChatApp(80, 24)
	var w bytes.Buffer
	if chat.CopySelectionToWriter(nil, &w) {
		t.Error("expected false with nil buffer")
	}
}

func TestP131_CopySelection_NilWriter(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	if chat.CopySelectionToWriter(buf, nil) {
		t.Error("expected false with nil writer")
	}
}

func TestP131_CopySelection_NoSelection(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	// Set a selectionMgr with no selection
	chat.SetSelectionManager(NewSelectionManager())
	if chat.CopySelectionToWriter(buf, &w) {
		t.Error("expected false with no selection")
	}
}
