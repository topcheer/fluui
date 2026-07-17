package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

type p266Panel struct {
	id     string
	title  string
	hidden bool
}

func (p *p266Panel) ID() string                              { return p.id }
func (p *p266Panel) Title() string                           { return p.title }
func (p *p266Panel) HandleKey(ev *term.KeyEvent) bool         { return false }
func (p *p266Panel) HandleMouse(x, y int, action string) bool { return false }
func (p *p266Panel) Paint(buf *buffer.Buffer, w, h int)       {}
func (p *p266Panel) OnShow()                                  {}
func (p *p266Panel) OnHide()                                  { p.hidden = true }

func TestPanelManager_Pop_SinglePanel_P266(t *testing.T) {
	pm := NewPanelManager(&p266Panel{id: "root", title: "root"})
	p := pm.Pop()
	if p != nil {
		t.Error("popping with single panel should return nil")
	}
}

func TestPanelManager_Pop_Multiple_P266(t *testing.T) {
	root := &p266Panel{id: "root", title: "root"}
	pm := NewPanelManager(root)
	child := &p266Panel{id: "child", title: "child"}
	pm.Push(child)
	p := pm.Pop()
	if p == nil {
		t.Error("popping should return the child panel")
	}
}

func TestPanelManager_Active_Empty_P266(t *testing.T) {
	pm := NewPanelManager(&p266Panel{id: "root", title: "root"})
	pm.panels = nil
	a := pm.Active()
	if a != nil {
		t.Error("active should be nil when no panels")
	}
}

func TestPanelManager_CloseAll_Single_P266(t *testing.T) {
	pm := NewPanelManager(&p266Panel{id: "root", title: "root"})
	pm.CloseAll()
	a := pm.Active()
	if a == nil {
		t.Error("root should still be active after CloseAll")
	}
}

func TestPanelManager_CloseAll_Multiple_P266(t *testing.T) {
	root := &p266Panel{id: "root", title: "root"}
	child := &p266Panel{id: "child", title: "child"}
	pm := NewPanelManager(root)
	pm.Push(child)
	pm.CloseAll()
	a := pm.Active()
	if a == nil || a.Title() != "root" {
		t.Errorf("expected root active, got %v", a)
	}
	if !child.hidden {
		t.Error("child should have been hidden")
	}
}

func TestSelection_InactiveRange_P266(t *testing.T) {
	sm := NewSelectionManager()
	_, _, ok := sm.SelectionRange()
	if ok {
		t.Error("should return false when not active")
	}
}

func TestSelection_ExtendInactive_P266(t *testing.T) {
	sm := NewSelectionManager()
	sm.ExtendKeyboardSelection(5, 5, 100, 100)
	_, _, ok := sm.SelectionRange()
	if ok {
		t.Error("should not be active after extend on inactive selection")
	}
}

func TestSelection_GetSelectedText_Inactive_P266(t *testing.T) {
	sm := NewSelectionManager()
	buf := buffer.NewBuffer(20, 10)
	txt := sm.GetSelectedText(buf)
	if txt != "" {
		t.Error("should return empty string when not active")
	}
}
