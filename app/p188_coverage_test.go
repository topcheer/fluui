package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P188: Target AppShell sub-80% functions

func TestP188_AppShell_OnShowOnHide(t *testing.T) {
	shell := NewAppShell(nil)
	shell.OnShow()
	shell.OnHide()
}

func TestP188_AppShell_HandleMouse(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.HandleMouse(5, 2, "down")
	shell.HandleMouse(5, 2, "move")
	shell.HandleMouse(5, 2, "up")
}

func TestP188_AppShell_DrawText(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	buf := buffer.NewBuffer(80, 24)
	shell.Paint(buf, 80, 24)
	// Small height edge case
	shell.Paint(buffer.NewBuffer(80, 2), 80, 2)
	// No status bar
	shell.SetStatus("", "", "")
	shell.Paint(buffer.NewBuffer(80, 5), 80, 5)
}

func TestP188_AppShell_ClearSidebarSections(t *testing.T) {
	shell := NewAppShell(nil)
	shell.AddSidebarSection("Section 1", []string{"Item A", "Item B"})
	shell.AddSidebarSection("Section 2", []string{"Item C"})
	shell.ClearSidebarSections()
}

func TestP188_PanelManager_OnShowOnHide(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	pm.Root().OnShow()
	pm.Root().OnHide()
}

func TestP188_PanelManager_Replace(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	pm.Push(&mockPanelP184{id: "a", title: "A"})
	pm.Replace(&mockPanelP184{id: "b", title: "B"})
	if pm.Active().ID() != "b" {
		t.Errorf("expected b, got %s", pm.Active().ID())
	}
}

func TestP188_View_LineCount(t *testing.T) {
	v := NewView("line1\nline2\nline3")
	if v.LineCount() != 3 {
		t.Errorf("expected 3 lines, got %d", v.LineCount())
	}
	v2 := NewView("")
	if v2.LineCount() != 0 {
		t.Errorf("expected 0 lines for empty, got %d", v2.LineCount())
	}
}
