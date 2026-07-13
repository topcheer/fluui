package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Helper to create AppShell with a dummy root panel
func newTestAppShell() *AppShell {
	root := &p178TestPanel{}
	return NewAppShell(root)
}

type p178TestPanel struct{}

func (t *p178TestPanel) ID() string              { return "test" }
func (t *p178TestPanel) Title() string            { return "Test" }
func (t *p178TestPanel) HandleKey(*term.KeyEvent) bool { return false }
func (t *p178TestPanel) HandleMouse(x, y int, action string) bool { return false }
func (t *p178TestPanel) Paint(buf *buffer.Buffer, w, h int)  {}
func (t *p178TestPanel) OnShow()                  {}
func (t *p178TestPanel) OnHide()                  {}

// === AppShell coverage ===

func TestP178_AppShell_Title(t *testing.T) {
	s := newTestAppShell()
	if s.Title() != "App" {
		t.Errorf("expected 'App', got %q", s.Title())
	}
}

func TestP178_AppShell_ID(t *testing.T) {
	s := newTestAppShell()
	if s.ID() != "app-shell" {
		t.Errorf("expected 'app-shell', got %q", s.ID())
	}
}

func TestP178_AppShell_OnShow(t *testing.T) {
	s := newTestAppShell()
	s.OnShow()
}

func TestP178_AppShell_OnHide(t *testing.T) {
	s := newTestAppShell()
	s.OnHide()
}

func TestP178_AppShell_HandleMouse(t *testing.T) {
	s := newTestAppShell()
	// Should not panic
	s.HandleMouse(5, 5, "click")
}

func TestP178_AppShell_ClearSidebarSections(t *testing.T) {
	s := newTestAppShell()
	s.AddSidebarSection("Info", []string{"line1", "line2"})
	s.ClearSidebarSections()
	s.AddSidebarSection("New", []string{"new line"})
	s.Paint(buffer.NewBuffer(80, 24), 80, 24)
}

func TestP178_AppShell_ShellHelpText(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	router.RegisterGlobal("quit", "ctrl+q", "Quit", func() bool { return true })
	text := ShellHelpText(router)
	if text == "" {
		t.Error("expected non-empty help text")
	}
}

func TestP178_AppShell_PaintWithSidebar(t *testing.T) {
	s := newTestAppShell()
	s.AddPanelItem("chat", "Chat", "C")
	s.AddPanelItem("files", "Files", "F")
	s.AddSidebarSection("Status", []string{"Running", "v1.0"})
	s.ShowSidebar()
	s.SetStatus("active", "editor", "Ctrl+Q Quit")
	s.Paint(buffer.NewBuffer(80, 24), 80, 24)
}

func TestP178_AppShell_PaintWithoutSidebar(t *testing.T) {
	s := newTestAppShell()
	s.HideSidebar()
	s.Paint(buffer.NewBuffer(80, 24), 80, 24)
}

func TestP178_AppShell_PaintSmallHeight(t *testing.T) {
	s := newTestAppShell()
	s.Paint(buffer.NewBuffer(80, 1), 80, 1)
}

func TestP178_AppShell_PaintStatusBar(t *testing.T) {
	s := newTestAppShell()
	s.SetStatus("compiling", "go", "F1 Help")
	s.ShowSidebar()
	s.Paint(buffer.NewBuffer(60, 24), 60, 24)
}

func TestP178_AppShell_PushPop(t *testing.T) {
	root := newTestAppShell()
	s := newTestAppShell()
	s.Push(root)
	s.Push(root)
	if s.PanelDepth() != 3 {
		t.Errorf("expected depth 3, got %d", s.PanelDepth())
	}
	s.Pop()
	if s.PanelDepth() != 2 {
		t.Errorf("expected depth 2, got %d", s.PanelDepth())
	}
	s.CloseAllPanels()
	if s.PanelDepth() != 1 {
		t.Errorf("expected depth 1, got %d", s.PanelDepth())
	}
}

// === EventRouter coverage ===

func TestP178_EventRouter_ActiveContext(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	_ = router.ActiveContext()
	// Default context is "global"
	router.PushContext("editor")
	if router.ActiveContext() != "editor" {
		t.Errorf("expected 'editor', got %q", router.ActiveContext())
	}
	router.PopContext()
}

func TestP178_EventRouter_HandleMouse(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	router.HandleMouse(5, 5, "click")
}

func TestP178_EventRouter_PopPanel(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	router.PushPanel(newTestAppShell())
	popped := router.PopPanel()
	if popped == nil {
		t.Error("expected non-nil popped panel")
	}
}

func TestP178_EventRouter_HandleKeyEsc(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	router.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
}

func TestP178_EventRouter_HandleKeyBinding(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	called := false
	router.RegisterGlobal("quit", "ctrl+q", "Quit", func() bool { called = true; return true })
	router.HandleKey(&term.KeyEvent{Rune: 'q', Modifiers: term.ModCtrl})
	if !called {
		t.Error("expected binding called")
	}
}

func TestP178_EventRouter_HandleKeyFallback(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	km := component.NewKeybindingManager()
	router := NewEventRouter(pm, km)
	called := false
	router.SetFallback(func(ev *term.KeyEvent) bool {
		called = true
		return true
	})
	router.HandleKey(&term.KeyEvent{Rune: 'x'})
	if !called {
		t.Error("expected fallback called")
	}
}

// === PanelManager coverage ===

func TestP178_PanelManager_Root(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	if pm.Root() != root {
		t.Error("expected root panel")
	}
}

func TestP178_PanelManager_Replace(t *testing.T) {
	root := newTestAppShell()
	pm := NewPanelManager(root)
	// Replace when only root → pushes
	old := pm.Replace(newTestAppShell())
	if old != nil {
		t.Error("expected nil when only root")
	}
	// Replace with 2 panels → swaps top
	old = pm.Replace(newTestAppShell())
	if old == nil {
		t.Error("expected non-nil old panel")
	}
}

// === View LineCount ===

func TestP178_View_LineCount(t *testing.T) {
	v := View{content: "line1\nline2\nline3"}
	if v.LineCount() != 3 {
		t.Errorf("expected 3, got %d", v.LineCount())
	}
	v2 := View{content: "single line"}
	if v2.LineCount() != 1 {
		t.Errorf("expected 1, got %d", v2.LineCount())
	}
	v3 := View{content: ""}
	if v3.LineCount() != 0 {
		t.Errorf("expected 0 for empty, got %d", v3.LineCount())
	}
}
