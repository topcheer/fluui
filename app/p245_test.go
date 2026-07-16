package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

func TestActiveSessionName_NilTabBar_P245(t *testing.T) {
	a := NewChatApp(80, 24)
	if name := a.ActiveSessionName(); name != "" {
		t.Error("nil tabBar should return empty")
	}
}

func TestActiveSessionName_NilTab_P245(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetTabBar(component.NewTabBar())
	if name := a.ActiveSessionName(); name != "" {
		t.Error("nil tab should return empty")
	}
}

func TestSessions_NilTabBar_P245(t *testing.T) {
	a := NewChatApp(80, 24)
	if s := a.Sessions(); s != nil {
		t.Error("nil tabBar should return nil")
	}
}

func TestHandleP20Key_PaletteVisible_P245(t *testing.T) {
	a := NewChatApp(80, 24)
	a.ToggleCommandPalette()
	a.handleP20Key(&term.KeyEvent{Key: term.KeyDown})
}

func TestHandleP20Key_PaletteHidden_P245(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.handleP20Key(&term.KeyEvent{Key: term.KeyDown}) {
		t.Error("should return false")
	}
}

func TestHandleP20Key_CtrlP_P245(t *testing.T) {
	a := NewChatApp(80, 24)
	a.handleP20Key(&term.KeyEvent{Rune: 'p', Modifiers: term.ModCtrl})
}
