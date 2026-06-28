package app

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

func TestP20_SetCommandPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)
	if a.CommandPalette() != cp {
		t.Error("CommandPalette() should return the set palette")
	}
}

func TestP20_SetCommandPalette_Nil(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetCommandPalette(nil)
	if a.CommandPalette() != nil {
		t.Error("CommandPalette() should be nil")
	}
}

func TestP20_SetSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	a.SetSpinner(s)
	if a.Spinner() != s {
		t.Error("Spinner() should return the set spinner")
	}
}

func TestP20_SetSpinner_Nil(t *testing.T) {
	a := NewChatApp(80, 24)
	a.SetSpinner(nil)
	if a.Spinner() != nil {
		t.Error("Spinner() should be nil")
	}
}

func TestP20_ToggleCommandPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	// Initially hidden
	if a.IsCommandPaletteVisible() {
		t.Error("palette should be hidden initially")
	}

	// Toggle on
	if !a.ToggleCommandPalette() {
		t.Error("ToggleCommandPalette should return true when palette exists")
	}
	if !a.IsCommandPaletteVisible() {
		t.Error("palette should be visible after toggle")
	}

	// Toggle off
	a.ToggleCommandPalette()
	if a.IsCommandPaletteVisible() {
		t.Error("palette should be hidden after second toggle")
	}
}

func TestP20_ToggleCommandPalette_NoPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.ToggleCommandPalette() {
		t.Error("ToggleCommandPalette should return false without palette")
	}
}

func TestP20_IsCommandPaletteVisible_NoPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.IsCommandPaletteVisible() {
		t.Error("should return false when no palette attached")
	}
}

func TestP20_HandleP20Key_CtrlP(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	// Ctrl+P should toggle palette
	k := &term.KeyEvent{Rune: 'p', Modifiers: term.ModCtrl}
	if !a.handleP20Key(k) {
		t.Error("Ctrl+P should be consumed when palette exists")
	}
	if !a.IsCommandPaletteVisible() {
		t.Error("palette should be visible after Ctrl+P")
	}
}

func TestP20_HandleP20Key_CtrlP_NoPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	k := &term.KeyEvent{Rune: 'p', Modifiers: term.ModCtrl}
	if a.handleP20Key(k) {
		t.Error("Ctrl+P should not be consumed without palette")
	}
}

func TestP20_HandleP20Key_OtherKey(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	k := &term.KeyEvent{Rune: 'x'}
	if a.handleP20Key(k) {
		t.Error("non-Ctrl+P key should not be consumed when palette hidden")
	}
}

func TestP20_StartSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	a.SetSpinner(s)

	a.StartSpinner("Loading...")
	if !a.IsSpinnerActive() {
		t.Error("spinner should be active after StartSpinner")
	}
	if s.Label() != "Loading..." {
		t.Errorf("label = %q, want %q", s.Label(), "Loading...")
	}
}

func TestP20_StopSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	s := component.NewSpinner("")
	a.SetSpinner(s)

	a.StartSpinner("Working")
	a.StopSpinner()
	if a.IsSpinnerActive() {
		t.Error("spinner should be inactive after StopSpinner")
	}
}

func TestP20_StartSpinner_NoSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	a.StartSpinner("test") // should not panic
	if a.IsSpinnerActive() {
		t.Error("should be false without spinner")
	}
}

func TestP20_StopSpinner_NoSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	a.StopSpinner() // should not panic
}

func TestP20_IsSpinnerActive_NoSpinner(t *testing.T) {
	a := NewChatApp(80, 24)
	if a.IsSpinnerActive() {
		t.Error("should be false without spinner")
	}
}

func TestP20_AddCommand(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)

	called := false
	ok := a.AddCommand("test", "Test Command", "General", func() {
		called = true
	})
	if !ok {
		t.Error("AddCommand should return true when palette exists")
	}

	cmds := cp.Commands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].ID != "test" || cmds[0].Label != "Test Command" {
		t.Errorf("unexpected command: %+v", cmds[0])
	}

	// Execute the action
	cmds[0].Action()
	if !called {
		t.Error("command action should have been called")
	}
}

func TestP20_AddCommand_NoPalette(t *testing.T) {
	a := NewChatApp(80, 24)
	ok := a.AddCommand("test", "Test", "Cat", func() {})
	if ok {
		t.Error("AddCommand should return false without palette")
	}
}

func TestP20_ConcurrentAccess(t *testing.T) {
	a := NewChatApp(80, 24)
	cp := component.NewCommandPalette()
	a.SetCommandPalette(cp)
	s := component.NewSpinner("")
	a.SetSpinner(s)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				a.ToggleCommandPalette()
				a.IsCommandPaletteVisible()
				a.StartSpinner("test")
				a.StopSpinner()
				a.AddCommand("cmd", "Cmd", "Cat", func() {})
			}
		}(i)
	}
	wg.Wait()
}
