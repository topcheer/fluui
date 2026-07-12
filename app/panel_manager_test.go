package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// mockPanel implements Panel for testing.
type mockPanel struct {
	component.BaseComponent
	id      string
	title   string
	icon    string
	active  int
	deact   int
}

func (m *mockPanel) ID() string     { return m.id }
func (m *mockPanel) Title() string  { return m.title }
func (m *mockPanel) Icon() string   { return m.icon }
func (m *mockPanel) OnActivate()    { m.active++ }
func (m *mockPanel) OnDeactivate()  { m.deact++ }
func (m *mockPanel) Paint(buf *buffer.Buffer) {}
func (m *mockPanel) Children() []component.Component { return nil }

func TestPanelManager_RegisterAndSwitch(t *testing.T) {
	pm := NewPanelManager()
	p1 := &mockPanel{id: "chat", title: "Chat", icon: "\U0001F4AC"}
	p2 := &mockPanel{id: "tg", title: "Telegram", icon: "TG"}

	pm.RegisterPanel(p1, -1)
	pm.RegisterPanel(p2, -1)

	if pm.PanelCount() != 2 {
		t.Fatalf("expected 2 panels, got %d", pm.PanelCount())
	}
	if pm.ActiveID() != "chat" {
		t.Fatalf("expected active=chat, got %s", pm.ActiveID())
	}
	if p1.active != 1 {
		t.Fatal("first panel should be activated on registration")
	}

	// Switch to panel 2
	pm.SwitchTo("tg")
	if pm.ActiveID() != "tg" {
		t.Fatalf("expected active=tg, got %s", pm.ActiveID())
	}
	if p1.deact != 1 {
		t.Fatal("chat panel should be deactivated")
	}
	if p2.active != 1 {
		t.Fatal("tg panel should be activated")
	}
}

func TestPanelManager_NextPrev(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "a", title: "A"}, -1)
	pm.RegisterPanel(&mockPanel{id: "b", title: "B"}, -1)
	pm.RegisterPanel(&mockPanel{id: "c", title: "C"}, -1)

	pm.SwitchTo("b")
	pm.Next()
	if pm.ActiveID() != "c" {
		t.Fatalf("expected c after next, got %s", pm.ActiveID())
	}

	// Wrap around
	pm.Next()
	if pm.ActiveID() != "a" {
		t.Fatalf("expected a after wrap next, got %s", pm.ActiveID())
	}

	pm.Prev()
	if pm.ActiveID() != "c" {
		t.Fatalf("expected c after wrap prev, got %s", pm.ActiveID())
	}
}

func TestPanelManager_SwitchByIndex(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "a", title: "A"}, -1)
	pm.RegisterPanel(&mockPanel{id: "b", title: "B"}, -1)

	if !pm.SwitchByIndex(2) {
		t.Fatal("SwitchByIndex(2) should succeed")
	}
	if pm.ActiveID() != "b" {
		t.Fatalf("expected b, got %s", pm.ActiveID())
	}

	if pm.SwitchByIndex(5) {
		t.Fatal("SwitchByIndex(5) should fail")
	}
}

func TestPanelManager_Unregister(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "a", title: "A"}, -1)
	pm.RegisterPanel(&mockPanel{id: "b", title: "B"}, -1)

	pm.SwitchTo("b")
	pm.UnregisterPanel("b")
	if pm.ActiveID() != "a" {
		t.Fatalf("expected fallback to a, got %s", pm.ActiveID())
	}
	if pm.PanelCount() != 1 {
		t.Fatalf("expected 1 panel, got %d", pm.PanelCount())
	}
}

func TestPanelManager_HasPanel(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "x", title: "X"}, -1)

	if !pm.HasPanel("x") {
		t.Fatal("should have panel x")
	}
	if pm.HasPanel("y") {
		t.Fatal("should not have panel y")
	}
}

func TestPanelManager_Unread(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "chat", title: "Chat"}, -1)
	pm.SetUnread("chat", 5)
	// Switching to it clears unread
	pm.SwitchTo("chat")
	// Unread cleared on switch — just verify no panic
}

func TestPanelManager_HandleKey_AltNum(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "a", title: "A"}, -1)
	pm.RegisterPanel(&mockPanel{id: "b", title: "B"}, -1)

	// Alt+2 should switch to panel 2
	k := &term.KeyEvent{Key: term.KeyUnknown, Rune: '2', Modifiers: term.ModAlt}
	pm.HandleKey(k)
	if pm.ActiveID() != "b" {
		t.Fatalf("expected b after Alt+2, got %s", pm.ActiveID())
	}
}

func TestPanelManager_HandleKey_CtrlTab(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "a", title: "A"}, -1)
	pm.RegisterPanel(&mockPanel{id: "b", title: "B"}, -1)

	k := &term.KeyEvent{Key: term.KeyTab, Modifiers: term.ModCtrl}
	pm.HandleKey(k)
	if pm.ActiveID() != "b" {
		t.Fatalf("expected b after Ctrl+Tab, got %s", pm.ActiveID())
	}
}

func TestPanelManager_Paint(t *testing.T) {
	pm := NewPanelManager()
	pm.RegisterPanel(&mockPanel{id: "a", title: "Alpha"}, -1)
	pm.RegisterPanel(&mockPanel{id: "b", title: "Beta"}, -1)
	pm.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	pm.Paint(buf)

	// Tab bar should be at bottom (y=23)
	// Check that "[1]" appears somewhere in the last row
	found := false
	for x := 0; x < 80; x++ {
		c := buf.GetCell(x, 23)
		if c.Rune == '[' {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("tab bar label not found in bottom row")
	}
}
