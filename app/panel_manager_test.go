package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Test Panels ───

type testPanel struct {
	BasePanel
	id       string
	title    string
	keyCalls int
	painted  bool
	shown    bool
	hidden   bool
}

func newTestPanel(id, title string) *testPanel {
	return &testPanel{id: id, title: title}
}

func (p *testPanel) ID() string       { return p.id }
func (p *testPanel) Title() string    { return p.title }
func (p *testPanel) HandleKey(ev *term.KeyEvent) bool {
	p.keyCalls++
	if ev.Rune == 'x' {
		return true // consume 'x'
	}
	return false
}
func (p *testPanel) Paint(buf *buffer.Buffer, w, h int) { p.painted = true }
func (p *testPanel) OnShow()                            { p.shown = true }
func (p *testPanel) OnHide()                            { p.hidden = true }

// ─── PanelManager Tests ───

func TestPanelManager_RootPanel(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	if pm.Active().ID() != "root" {
		t.Errorf("expected root active, got %s", pm.Active().ID())
	}
	if pm.Depth() != 1 {
		t.Errorf("expected depth 1, got %d", pm.Depth())
	}
	if !pm.IsRoot() {
		t.Error("expected IsRoot=true")
	}
}

func TestPanelManager_PushPop(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	modal := newTestPanel("modal", "Modal")
	pm.Push(modal)

	if pm.Depth() != 2 {
		t.Errorf("expected depth 2, got %d", pm.Depth())
	}
	if pm.Active().ID() != "modal" {
		t.Error("modal should be active")
	}
	if !modal.shown {
		t.Error("OnShow not called on modal")
	}

	popped := pm.Pop()
	if popped == nil || popped.ID() != "modal" {
		t.Error("should pop modal")
	}
	if !popped.(*testPanel).hidden {
		t.Error("OnHide not called on popped panel")
	}
	if pm.Active().ID() != "root" {
		t.Error("root should be active after pop")
	}
}

func TestPanelManager_EscPopsPanel(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	modal := newTestPanel("modal", "Modal")
	pm.Push(modal)

	// Esc should pop the modal
	consumed := pm.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if !consumed {
		t.Error("Esc should be consumed")
	}
	if pm.Depth() != 1 {
		t.Error("modal should be popped by Esc")
	}
}

func TestPanelManager_EscDoesNotPopRoot(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	consumed := pm.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if consumed {
		t.Error("Esc on root should not be consumed (no pop)")
	}
	if pm.Depth() != 1 {
		t.Error("root should not be popped")
	}
}

func TestPanelManager_ActivePanelGetsKey(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	modal := newTestPanel("modal", "Modal")
	pm.Push(modal)

	pm.HandleKey(&term.KeyEvent{Rune: 'a'})

	if root.keyCalls > 0 {
		t.Error("root should not receive keys when modal is active")
	}
	if modal.keyCalls != 1 {
		t.Error("modal should receive key")
	}
}

func TestPanelManager_Replace(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	first := newTestPanel("first", "First")
	pm.Push(first)

	second := newTestPanel("second", "Second")
	old := pm.Replace(second)

	if old == nil || old.ID() != "first" {
		t.Error("should return replaced panel")
	}
	if pm.Active().ID() != "second" {
		t.Error("second should be active")
	}
}

func TestPanelManager_CloseAll(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	pm.Push(newTestPanel("p1", "P1"))
	pm.Push(newTestPanel("p2", "P2"))
	pm.Push(newTestPanel("p3", "P3"))

	if pm.Depth() != 4 {
		t.Error("expected 4 panels")
	}

	pm.CloseAll()

	if pm.Depth() != 1 {
		t.Errorf("expected 1 after closeAll, got %d", pm.Depth())
	}
	if pm.Active().ID() != "root" {
		t.Error("root should be active after closeAll")
	}
}

func TestPanelManager_FindByID(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	pm.Push(newTestPanel("alpha", "Alpha"))
	pm.Push(newTestPanel("beta", "Beta"))

	p := pm.FindByID("alpha")
	if p == nil || p.ID() != "alpha" {
		t.Error("should find alpha")
	}

	if pm.FindByID("nonexistent") != nil {
		t.Error("nonexistent should return nil")
	}
}

func TestPanelManager_OnChangeCallback(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	var changes int
	pm.SetOnChange(func() { changes++ })

	pm.Push(newTestPanel("p1", "P1"))
	pm.Pop()

	if changes != 2 {
		t.Errorf("expected 2 change callbacks, got %d", changes)
	}
}

func TestPanelManager_Panels(t *testing.T) {
	root := newTestPanel("root", "Root")
	pm := NewPanelManager(root)

	pm.Push(newTestPanel("p1", "P1"))
	pm.Push(newTestPanel("p2", "P2"))

	panels := pm.Panels()
	if len(panels) != 3 {
		t.Errorf("expected 3 panels, got %d", len(panels))
	}
	if panels[0].ID() != "root" || panels[2].ID() != "p2" {
		t.Error("panel order should be bottom-to-top")
	}
}

func TestPanelManager_HandleMouse(t *testing.T) {
	root := &mouseTestPanel{id: "root"}
	pm := NewPanelManager(root)

	modal := &mouseTestPanel{id: "modal"}
	pm.Push(modal)

	pm.HandleMouse(5, 10, "click")

	if !modal.mouseReceived {
		t.Error("modal should receive mouse")
	}
	if root.mouseReceived {
		t.Error("root should not receive mouse when modal active")
	}
}

type mouseTestPanel struct {
	BasePanel
	id            string
	mouseReceived bool
}

func (m *mouseTestPanel) ID() string                                       { return m.id }
func (m *mouseTestPanel) Title() string                                    { return m.id }
func (m *mouseTestPanel) HandleKey(ev *term.KeyEvent) bool                 { return false }
func (m *mouseTestPanel) Paint(buf *buffer.Buffer, w, h int)               {}
func (m *mouseTestPanel) HandleMouse(x, y int, action string) bool {
	m.mouseReceived = true
	return true
}