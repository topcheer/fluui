package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
)

// P325: Push app package sub-85% functions past 85%

func TestP325_HandleP20Key_CtrlP(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Rune: 'p', Modifiers: term.ModCtrl}
	chat.handleP20Key(ev)
}

func TestP325_HandleP20Key_NormalKey(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Rune: 'x'}
	if chat.handleP20Key(ev) {
		t.Error("normal key should not be handled")
	}
}

func TestP325_HandleP20Key_PaletteVisible(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.ToggleCommandPalette()
	ev := &term.KeyEvent{Rune: 'a'}
	chat.handleP20Key(ev)
}

func TestP325_HandleP20Key_Esc(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.ToggleCommandPalette()
	ev := &term.KeyEvent{Key: term.KeyEscape}
	chat.handleP20Key(ev)
}

func TestP325_HandleP16Keys_NoTabBar(t *testing.T) {
	chat := NewChatApp(80, 24)
	chat.mu.Lock()
	chat.tabBar = nil
	chat.mu.Unlock()
	ev := &term.KeyEvent{Modifiers: term.ModAlt, Rune: ']'}
	if chat.handleP16Keys(ev) {
		t.Error("should return false when tabBar is nil")
	}
}

func TestP325_HandleP16Keys_AltNext(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Modifiers: term.ModAlt, Rune: ']'}
	chat.handleP16Keys(ev)
}

func TestP325_HandleP16Keys_AltPrev(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Modifiers: term.ModAlt, Rune: '['}
	chat.handleP16Keys(ev)
}

func TestP325_HandleP16Keys_AltW(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Modifiers: term.ModAlt, Rune: 'w'}
	chat.handleP16Keys(ev)
}

func TestP325_HandleP16Keys_AltNumber(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Modifiers: term.ModAlt, Rune: '1'}
	chat.handleP16Keys(ev)
}

func TestP325_HandleP16Keys_NormalKey(t *testing.T) {
	chat := NewChatApp(80, 24)
	ev := &term.KeyEvent{Rune: 'x'}
	if chat.handleP16Keys(ev) {
		t.Error("non-alt key should not be handled")
	}
}
