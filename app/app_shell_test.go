package app

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestAppShell_Basic(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	if shell.ID() != "app-shell" {
		t.Error("wrong ID")
	}
	if shell.PanelManager() == nil {
		t.Error("PanelManager should not be nil")
	}
	if shell.StatusBar() == nil {
		t.Error("StatusBar should not be nil")
	}
}

func TestAppShell_SidebarToggle(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	if shell.IsSidebarVisible() {
		t.Error("sidebar should start hidden")
	}

	shell.ShowSidebar()
	if !shell.IsSidebarVisible() {
		t.Error("sidebar should be visible after ShowSidebar")
	}

	shell.HideSidebar()
	if shell.IsSidebarVisible() {
		t.Error("sidebar should be hidden after HideSidebar")
	}

	shell.ToggleSidebar()
	if !shell.IsSidebarVisible() {
		t.Error("sidebar should be visible after ToggleSidebar")
	}

	shell.ToggleSidebar()
	if shell.IsSidebarVisible() {
		t.Error("sidebar should be hidden after second ToggleSidebar")
	}
}

func TestAppShell_SetStatus(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	shell.SetStatus("thinking", "read_file", "Ctrl+Q quit")

	// StatusBar should have items
	items := shell.StatusBar().LeftItems()
	if len(items) == 0 {
		t.Error("status bar should have left items after SetStatus")
	}
}

func TestAppShell_AddSidebarSection(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	shell.AddSidebarSection("Version", []string{"v1.0.0", "build 123"})
	shell.AddSidebarSection("Session", []string{"main"})

	// Just verify it doesn't crash — sections are internal
}

func TestAppShell_AddPanelItem(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	shell.AddPanelItem("chat", "Chat", ">")
	shell.AddPanelItem("files", "Files", "📁")

	// Just verify it doesn't crash
}

func TestAppShell_PushPopPanel(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	if shell.PanelDepth() != 1 {
		t.Error("should start with depth 1")
	}

	modal := newTestPanel("modal", "Modal")
	shell.Push(modal)

	if shell.PanelDepth() != 2 {
		t.Error("should have depth 2 after push")
	}
	if shell.ActivePanel().ID() != "modal" {
		t.Error("modal should be active")
	}

	shell.Pop()
	if shell.PanelDepth() != 1 {
		t.Error("should have depth 1 after pop")
	}
}

func TestAppShell_CloseAllPanels(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	shell.Push(newTestPanel("p1", "P1"))
	shell.Push(newTestPanel("p2", "P2"))
	shell.Push(newTestPanel("p3", "P3"))

	shell.CloseAllPanels()

	if shell.PanelDepth() != 1 {
		t.Error("should have depth 1 after CloseAllPanels")
	}
	if shell.ActivePanel().ID() != "root" {
		t.Error("root should be active after CloseAllPanels")
	}
}

func TestAppShell_PaintNoSidebar(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	buf := buffer.NewBuffer(80, 24)
	shell.Paint(buf, 80, 24)

	if !root.painted {
		t.Error("root panel should be painted")
	}
}

func TestAppShell_PaintWithSidebar(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)
	shell.ShowSidebar()
	shell.AddSidebarSection("Info", []string{"v1.0", "session-1"})
	shell.AddPanelItem("main", "Main", ">")

	buf := buffer.NewBuffer(80, 24)
	shell.Paint(buf, 80, 24)

	if !root.painted {
		t.Error("root panel should be painted even with sidebar")
	}
}

func TestAppShell_SetSidebarWidth(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	shell.SetSidebarWidth(30)
	// Should be accepted (between 5 and 60)
	shell.SetSidebarWidth(3)  // too small, rejected
	shell.SetSidebarWidth(100) // too large, rejected
}

func TestAppShell_ImplementsPanel(t *testing.T) {
	var _ Panel = (*AppShell)(nil)
}

func TestAppShell_HandleKey(t *testing.T) {
	root := newTestPanel("root", "Root")
	shell := NewAppShell(root)

	// 'x' is consumed by testPanel
	consumed := shell.HandleKey(&term.KeyEvent{Rune: 'x'})
	if !consumed {
		t.Error("HandleKey should route to root panel")
	}
}

func TestSplitLines(t *testing.T) {
	lines := SplitLines("hello world", 5)
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "hello" || lines[1] != " worl" || lines[2] != "d" {
		t.Errorf("unexpected split: %v", lines)
	}
}

func TestSplitLines_ShortLine(t *testing.T) {
	lines := SplitLines("hi", 10)
	if len(lines) != 1 || lines[0] != "hi" {
		t.Errorf("expected single line 'hi', got %v", lines)
	}
}

func TestSplitLines_MultiLine(t *testing.T) {
	lines := SplitLines("line1\nline2", 10)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}