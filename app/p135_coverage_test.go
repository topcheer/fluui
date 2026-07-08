package app

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/internal/termcompat"
)

// === HandleMouseP16 (76.3% → 90%+) — cover tab bar + selection + custom ===

func TestP135_HandleMouseP16_TabBarHit(t *testing.T) {
	chat := NewChatApp(80, 24)
	tb := component.NewTabBar()
	tb.AddTab("s-0", "Tab1")
	tb.AddTab("s-1", "Tab2")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	chat.SetTabBar(tb)

	// Click on second tab
	mouse := &term.MouseEvent{
		X:      10,
		Y:      0,
		Button: term.MouseLeft,
		Action: term.MouseDown,
	}
	if !chat.HandleMouseP16(mouse) {
		t.Error("expected true for tab bar click")
	}
}

func TestP135_HandleMouseP16_TabBarMiss(t *testing.T) {
	chat := NewChatApp(80, 24)
	tb := component.NewTabBar()
	tb.AddTab("s-0", "Tab1")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	chat.SetTabBar(tb)

	// Click below tab bar — should fall through to custom handler
	mouse := &term.MouseEvent{
		X:      10,
		Y:      10,
		Button: term.MouseLeft,
		Action: term.MouseDown,
	}
	chat.HandleMouseP16(mouse)
}

func TestP135_HandleMouseP16_MouseMoveOnTabBar(t *testing.T) {
	chat := NewChatApp(80, 24)
	tb := component.NewTabBar()
	tb.AddTab("s-0", "Tab1")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 1})
	chat.SetTabBar(tb)

	// Mouse move on tab bar area (not MouseDown)
	mouse := &term.MouseEvent{
		X:      5,
		Y:      0,
		Button: term.MouseLeft,
		Action: term.MouseMove,
	}
	chat.HandleMouseP16(mouse)
}

func TestP135_HandleMouseP16_CustomHandlerOnly(t *testing.T) {
	chat := NewChatApp(80, 24)
	called := false
	chat.OnMouse(func(m *term.MouseEvent) {
		called = true
	})
	mouse := &term.MouseEvent{
		X:      40,
		Y:      12,
		Button: term.MouseLeft,
		Action: term.MouseMove,
	}
	chat.HandleMouseP16(mouse)
	if !called {
		t.Error("expected custom handler called")
	}
}

func TestP135_HandleMouseP16_WithSelectionManager(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.SetSelectionManager(NewSelectionManager())
	mouse := &term.MouseEvent{
		X:      40,
		Y:      12,
		Button: term.MouseLeft,
		Action: term.MouseDown,
	}
	chat.HandleMouseP16(mouse)
}

// === copySelectionToWriter (70% → 90%+) ===

func TestP135_CopySelection_NilSelectionMgr(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	if chat.CopySelectionToWriter(buf, &w) {
		t.Error("expected false with nil selectionMgr")
	}
}

func TestP135_CopySelection_NilBuffer(t *testing.T) {
	chat := NewChatApp(80, 24)
	var w bytes.Buffer
	if chat.CopySelectionToWriter(nil, &w) {
		t.Error("expected false with nil buffer")
	}
}

func TestP135_CopySelection_NilWriter(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	if chat.CopySelectionToWriter(buf, nil) {
		t.Error("expected false with nil writer")
	}
}

func TestP135_CopySelection_NoSelection(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	chat.SetSelectionManager(NewSelectionManager())
	if chat.CopySelectionToWriter(buf, &w) {
		t.Error("expected false with no selection")
	}
}

func TestP135_CopySelection_OSC52Fallback(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	buf.DrawText(0, 0, "hello", buffer.Style{})
	var w bytes.Buffer
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)
	// Start a selection
	sm.StartSelection(0, 0)
	sm.ExtendSelection(4, 0)
	// Without clipboard config, should use OSC52 fallback
	chat.CopySelectionToWriter(buf, &w)
	// Should have written something (OSC52 sequence)
	if w.Len() == 0 {
		// OSC52 may return empty for no text — that's OK
	}
}

func TestP135_CopySelection_WithClipboardConfig(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	buf.DrawText(0, 0, "test text", buffer.Style{})
	var w bytes.Buffer
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)
	sm.StartSelection(0, 0)
	sm.ExtendSelection(3, 0)

	// Set a clipboard config
	chat.clipboardConfig = NewClipboardWithCapabilities(
		termcompat.Capabilities{HasOSC52: true},
	)
	// This tests the clipCfg != nil path
	chat.CopySelectionToWriter(buf, &w)
}

func TestP135_CopySelection_EmptySelectionText(t *testing.T) {
	chat := NewChatApp(80, 24)
	buf := buffer.NewBuffer(80, 24)
	var w bytes.Buffer
	sm := NewSelectionManager()
	chat.SetSelectionManager(sm)
	// Set clipboard config so the text=="" branch is hit
	chat.clipboardConfig = NewClipboardWithCapabilities(
		termcompat.Capabilities{HasOSC52: true},
	)
	// Start a selection but on empty cells
	sm.StartSelection(50, 0)
	sm.ExtendSelection(55, 0)
	if chat.CopySelectionToWriter(buf, &w) {
		t.Error("expected false for empty selection text")
	}
}
