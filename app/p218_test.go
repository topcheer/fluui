package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

type mockPanelP218 struct {
	BasePanel
	id    string
	title string
}

func (p *mockPanelP218) ID() string                         { return p.id }
func (p *mockPanelP218) Title() string                      { return p.title }
func (p *mockPanelP218) HandleKey(ev *term.KeyEvent) bool      { return false }
func (p *mockPanelP218) Paint(buf *buffer.Buffer, w, h int) {}

func TestP218_AppShell_DrawTextBold(t *testing.T) {
	root := &mockPanelP218{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.SetStatus("Running", "Edit", "Ctrl+Q quit")
	buf := buffer.NewBuffer(80, 24)
	shell.Paint(buf, 80, 24)
}

func TestP218_AppShell_DrawTextNarrowWidth(t *testing.T) {
	root := &mockPanelP218{id: "root", title: "Root"}
	shell := NewAppShell(root)
	shell.SetStatus("Very Long Status Text Here", "Edit Mode", "Ctrl+Q quit")
	buf := buffer.NewBuffer(20, 10)
	shell.Paint(buf, 20, 10)
}
