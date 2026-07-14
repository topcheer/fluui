package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// P189: Coverage tests for sub-80% app functions
// (testPanel and newTestPanel are defined in panel_manager_test.go)

func TestAppShell_OnShow_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	shell.OnShow()
}

func TestAppShell_OnHide_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	shell.OnHide()
}

func TestAppShell_DrawText_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	buf := buffer.NewBuffer(40, 10)
	shell.SetStatus("activity", "tool", "hint")
	shell.Paint(buf, 40, 10)
}

func TestAppShell_ClearSidebarSections_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	shell.AddSidebarSection("group1", []string{"item1", "item2"})
	shell.ClearSidebarSections()
}

func TestAppShell_PaintSmallHeight_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	buf := buffer.NewBuffer(40, 1)
	shell.Paint(buf, 40, 1)
}

func TestAppShell_HandleMouse_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	shell.HandleMouse(5, 3, "click")
}

func TestAppShell_Title_P189(t *testing.T) {
	shell := NewAppShell(newTestPanel("test", "Test"))
	if shell.Title() != "App" {
		t.Errorf("expected 'App', got '%s'", shell.Title())
	}
}

func TestEventRouter_ActiveContext_P189(t *testing.T) {
	pm := NewPanelManager(newTestPanel("root", "Root"))
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	if router.ActiveContext() != "global" {
		t.Error("default context should be 'global'")
	}
}

func TestEventRouter_PopPanel_P189(t *testing.T) {
	pm := NewPanelManager(newTestPanel("root", "Root"))
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	router.PushPanel(newTestPanel("overlay", "Overlay"))
	router.PopPanel()
}

func TestEventRouter_HandleMouse_P189(t *testing.T) {
	pm := NewPanelManager(newTestPanel("root", "Root"))
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	router.HandleMouse(5, 3, "click")
}

func TestPanelManager_Root_P189(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)
	if pm.Root() == nil {
		t.Error("Root should return non-nil")
	}
}

func TestPanelManager_FindByID_P189(t *testing.T) {
	pm := NewPanelManager(newTestPanel("root", "Root"))
	overlay := newTestPanel("overlay", "Overlay")
	pm.Push(overlay)
	found := pm.FindByID("overlay")
	if found == nil {
		t.Error("FindByID should find pushed panel")
	}
}