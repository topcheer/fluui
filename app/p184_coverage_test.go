package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP184_AppShell_DrawText(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	buf := buffer.NewBuffer(40, 10)
	shell.Paint(buf, 40, 10)
}

func TestP184_AppShell_PaintStatusBar(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.SetStatus("Running", "Edit", "Ctrl+Q quit")
	buf := buffer.NewBuffer(80, 24)
	shell.Paint(buf, 80, 24)
}

func TestP184_AppShell_PaintSmallHeight(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	shell := NewAppShell(root)
	buf := buffer.NewBuffer(80, 3)
	shell.Paint(buf, 80, 3)
}

func TestP184_PanelManager_HandleMouse(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	result := pm.HandleMouse(5, 5, "down")
	_ = result
}

func TestP184_PanelManager_BasePanel(t *testing.T) {
	bp := BasePanel{}
	bp.HandleMouse(1, 1, "down")
	bp.OnShow()
	bp.OnHide()
}

func TestP184_PanelManager_Replace(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	pm.Push(&mockPanelP184{id: "a", title: "A"})
	old := pm.Replace(&mockPanelP184{id: "b", title: "B"})
	if old == nil || old.ID() != "a" {
		t.Fatal("Replace should return old panel")
	}
	if pm.Active().ID() != "b" {
		t.Fatal("Active should be b after replace")
	}
}

func TestP184_PanelManager_FindByID(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	pm.Push(&mockPanelP184{id: "x", title: "X"})
	found := pm.FindByID("x")
	if found == nil || found.ID() != "x" {
		t.Fatal("should find panel x")
	}
	missing := pm.FindByID("y")
	if missing != nil {
		t.Fatal("should not find panel y")
	}
}

func TestP184_PanelManager_CloseAll(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	pm.Push(&mockPanelP184{id: "a", title: "A"})
	pm.Push(&mockPanelP184{id: "b", title: "B"})
	pm.CloseAll()
	if pm.Depth() != 1 {
		t.Fatalf("expected depth 1 after CloseAll, got %d", pm.Depth())
	}
}

func TestP184_PanelManager_Pop(t *testing.T) {
	root := &mockPanelP184{id: "root", title: "Root"}
	pm := NewPanelManager(root)
	pm.Push(&mockPanelP184{id: "a", title: "A"})
	popped := pm.Pop()
	if popped == nil || popped.ID() != "a" {
		t.Fatal("Pop should return panel a")
	}
	if pm.Depth() != 1 {
		t.Fatal("should be back to root only")
	}
}

// mockPanelP184 implements Panel for testing.
type mockPanelP184 struct {
	component.BaseComponent
	id    string
	title string
}

func (m *mockPanelP184) ID() string                               { return m.id }
func (m *mockPanelP184) Title() string                            { return m.title }
func (m *mockPanelP184) OnShow()                                    {}
func (m *mockPanelP184) OnHide()                                   {}
func (m *mockPanelP184) Paint(buf *buffer.Buffer, w, h int)        {}
func (m *mockPanelP184) Children() []component.Component           { return nil }
func (m *mockPanelP184) HandleKey(ev *term.KeyEvent) bool         { return false }
func (m *mockPanelP184) HandleMouse(x, y int, action string) bool  { return false }