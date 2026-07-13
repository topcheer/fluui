package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestEventRouter_PanelGetsKeyFirst(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	// 'x' is consumed by testPanel
	consumed := router.HandleKey(&term.KeyEvent{Rune: 'x'})
	if !consumed {
		t.Error("panel should consume 'x'")
	}
}

func TestEventRouter_KeybindingFallback(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	var called bool
	router.RegisterGlobal("test", "ctrl+s", "Test", func() bool {
		called = true
		return true
	})

	// Ctrl+S should match keybinding (panel doesn't consume it)
	router.HandleKey(&term.KeyEvent{Rune: 's', Modifiers: term.ModCtrl})

	if !called {
		t.Error("global keybinding should fire")
	}
}

func TestEventRouter_PanelPriorityOverKeybinding(t *testing.T) {
	// Panel that consumes everything
	root := &greedyPanel{id: "root"}
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	var keybindingCalled bool
	router.RegisterGlobal("test", "ctrl+s", "Test", func() bool {
		keybindingCalled = true
		return true
	})

	router.HandleKey(&term.KeyEvent{Rune: 's', Modifiers: term.ModCtrl})

	if keybindingCalled {
		t.Error("keybinding should not fire when panel consumes the key")
	}
}

type greedyPanel struct {
	BasePanel
	id string
}

func (g *greedyPanel) ID() string                       { return g.id }
func (g *greedyPanel) Title() string                    { return g.id }
func (g *greedyPanel) HandleKey(ev *term.KeyEvent) bool { return true } // consume all
func (g *greedyPanel) Paint(buf *buffer.Buffer, w, h int) {}
func (g *greedyPanel) OnShow()                          {}
func (g *greedyPanel) OnHide()                          {}

func TestEventRouter_FallbackHandler(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	var fallbackCalled bool
	router.SetFallback(func(ev *term.KeyEvent) bool {
		fallbackCalled = true
		return true
	})

	// Unmatched key goes to fallback
	router.HandleKey(&term.KeyEvent{Rune: 'z'})

	if !fallbackCalled {
		t.Error("fallback should be called for unmatched keys")
	}
}

func TestEventRouter_ContextSwitching(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	var editorCalled, globalCalled bool

	router.RegisterContext("editor", "save", "ctrl+s", "Save", func() bool {
		editorCalled = true
		return true
	})
	router.RegisterGlobal("search", "ctrl+f", "Search", func() bool {
		globalCalled = true
		return true
	})

	// Without editor context: ctrl+s should not match
	router.HandleKey(&term.KeyEvent{Rune: 's', Modifiers: term.ModCtrl})
	if editorCalled {
		t.Error("editor binding should not fire without context")
	}

	// Push editor context
	router.PushContext("editor")
	router.HandleKey(&term.KeyEvent{Rune: 's', Modifiers: term.ModCtrl})
	if !editorCalled {
		t.Error("editor binding should fire with context")
	}

	// Global still works
	router.HandleKey(&term.KeyEvent{Rune: 'f', Modifiers: term.ModCtrl})
	if !globalCalled {
		t.Error("global binding should always work")
	}

	// Pop context
	router.PopContext()
	editorCalled = false
	router.HandleKey(&term.KeyEvent{Rune: 's', Modifiers: term.ModCtrl})
	if editorCalled {
		t.Error("editor binding should not fire after context popped")
	}
}

func TestEventRouter_EscClosesPanel(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	modal := newTestPanel("modal", "Modal")
	router.PushPanel(modal)

	if pm.Depth() != 2 {
		t.Error("should have 2 panels")
	}

	// Esc closes modal
	router.HandleKey(&term.KeyEvent{Key: term.KeyEscape})

	if pm.Depth() != 1 {
		t.Error("Esc should close modal panel")
	}
}

func TestEventRouter_HelpText(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)

	router.RegisterGlobal("quit", "ctrl+q", "Quit app", func() bool { return true })

	help := router.HelpText()
	if help == "" {
		t.Error("help text should not be empty")
	}
}