package app

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── AppShell: Standard Application Layout ───
//
// AppShell provides the standard TUI app layout that most apps need:
//
//	┌─────────────┬──────────────────────────────┐
//	│  Sidebar    │  Panel Area (PanelManager)    │
//	│             │                              │
//	│  - Panel    │  Active panel renders here    │
//	│    list     │                              │
//	│  - Info     │                              │
//	│    section  │                              │
//	├─────────────┴──────────────────────────────┤
//	│  Status Bar (activity, tool, hints)        │
//	└────────────────────────────────────────────┘
//
// AppShell implements Panel so it can be used as the root panel.
// It owns the sidebar visibility, status bar, and panel rendering area.

// SidebarSection is a titled section in the sidebar.
type SidebarSection struct {
	Title string
	Lines []string
}

// AppShell is a root Panel that provides a standard app layout
// with sidebar, panel area, and status bar.
type AppShell struct {
	mu sync.RWMutex

	// Layout
	sidebarVisible bool
	sidebarWidth   int

	// Components
	panels   *PanelManager
	statusBar *component.StatusBar

	// Sidebar content
	sections []SidebarSection

	// Panel list (shown in sidebar)
	panelItems []panelListItem

	// Styling
	style AppShellStyle

	// Status
	statusActivity string
	statusTool     string
	statusHint     string
}

type panelListItem struct {
	id    string
	label string
	icon  string
}

// AppShellStyle holds colors for the shell.
type AppShellStyle struct {
	SidebarBg    buffer.Color
	SidebarTitle buffer.Color
	SidebarItem  buffer.Color
	Separator    buffer.Color
	StatusBg     buffer.Color
	StatusFg     buffer.Color
}

func defaultShellStyle() AppShellStyle {
	return AppShellStyle{
		SidebarBg:    buffer.RGB(40, 42, 54),   // Dracula bg
		SidebarTitle: buffer.RGB(189, 147, 249), // purple
		SidebarItem:  buffer.RGB(248, 248, 242), // white
		Separator:    buffer.RGB(98, 114, 164),  // comment
		StatusBg:     buffer.RGB(40, 42, 54),
		StatusFg:     buffer.RGB(248, 248, 242),
	}
}

// NewAppShell creates an AppShell with the given root panel as the
// initial content. The sidebar starts hidden (use ShowSidebar to show it).
func NewAppShell(rootPanel Panel) *AppShell {
	sb := component.NewStatusBar()
	sb.SetSeparator(" │ ")

	return &AppShell{
		panels:        NewPanelManager(rootPanel),
		statusBar:     sb,
		sidebarVisible: false,
		sidebarWidth:  20,
		style:         defaultShellStyle(),
	}
}

// ─── Accessors ───

func (s *AppShell) ID() string   { return "app-shell" }
func (s *AppShell) Title() string { return "App" }

// PanelManager returns the underlying panel manager.
func (s *AppShell) PanelManager() *PanelManager { return s.panels }

// StatusBar returns the status bar component.
func (s *AppShell) StatusBar() *component.StatusBar { return s.statusBar }

// IsSidebarVisible returns whether the sidebar is shown.
func (s *AppShell) IsSidebarVisible() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sidebarVisible
}

// ShowSidebar shows the sidebar.
func (s *AppShell) ShowSidebar() {
	s.mu.Lock()
	s.sidebarVisible = true
	s.mu.Unlock()
}

// HideSidebar hides the sidebar.
func (s *AppShell) HideSidebar() {
	s.mu.Lock()
	s.sidebarVisible = false
	s.mu.Unlock()
}

// ToggleSidebar toggles sidebar visibility.
func (s *AppShell) ToggleSidebar() {
	s.mu.Lock()
	s.sidebarVisible = !s.sidebarVisible
	s.mu.Unlock()
}

// SetSidebarWidth sets the sidebar column width.
func (s *AppShell) SetSidebarWidth(w int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if w > 5 && w < 60 {
		s.sidebarWidth = w
	}
}

// SetStatus sets the status bar content.
func (s *AppShell) SetStatus(activity, tool, hint string) {
	s.mu.Lock()
	s.statusActivity = activity
	s.statusTool = tool
	s.statusHint = hint
	s.mu.Unlock()

	s.statusBar.Clear()
	s.statusBar.AddLeft("activity", activity)
	s.statusBar.AddCenter("tool", tool)
	s.statusBar.AddRight("hint", hint)
}

// AddSidebarSection adds a titled info section to the sidebar.
func (s *AppShell) AddSidebarSection(title string, lines []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sections = append(s.sections, SidebarSection{Title: title, Lines: lines})
}

// ClearSidebarSections removes all sidebar sections.
func (s *AppShell) ClearSidebarSections() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sections = nil
}

// AddPanelItem adds a panel entry to the sidebar list.
func (s *AppShell) AddPanelItem(id, label, icon string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.panelItems = append(s.panelItems, panelListItem{id, label, icon})
}

// ─── Panel interface ───

func (s *AppShell) OnShow() {}
func (s *AppShell) OnHide() {}

func (s *AppShell) HandleKey(ev *term.KeyEvent) bool {
	return s.panels.HandleKey(ev)
}

func (s *AppShell) HandleMouse(x, y int, action string) bool {
	return s.panels.HandleMouse(x, y, action)
}

func (s *AppShell) Paint(buf *buffer.Buffer, w, h int) {
	s.mu.RLock()
	sidebarVisible := s.sidebarVisible
	sidebarW := s.sidebarWidth
	s.mu.RUnlock()

	// Layout: [sidebar | separator | panel area] + [status bar at bottom]
	statusH := 1
	panelH := h - statusH

	if panelH < 0 {
		panelH = 0
	}

	panelX := 0
	panelW := w

	if sidebarVisible {
		panelX = sidebarW + 1 // +1 for separator
		panelW = w - sidebarW - 1
		if panelW < 0 {
			panelW = 0
		}
	}

	// Paint sidebar
	if sidebarVisible {
		s.paintSidebar(buf, 0, 0, sidebarW, panelH)
		// Paint separator
		for y := 0; y < panelH; y++ {
			if sidebarW < w {
				buf.SetCell(sidebarW, y, buffer.Cell{
					Rune:   '│',
					Width:  1,
					Fg:     s.style.Separator,
					Bg:     s.style.SidebarBg,
				})
			}
		}
	}

	// Paint panel area (active panel)
	if panelW > 0 && panelH > 0 {
		// Create a sub-buffer view for the panel
		subBuf := buffer.NewBuffer(panelW, panelH)
		s.panels.Paint(subBuf, panelW, panelH)

		// Blit sub-buffer into main buffer at (panelX, 0)
		for y := 0; y < panelH; y++ {
			for x := 0; x < panelW; x++ {
				if panelX+x < w {
					cell := subBuf.GetCell(x, y)
					buf.SetCell(panelX+x, y, cell)
				}
			}
		}
	}

	// Paint status bar at the bottom
	if statusH > 0 && h > 0 {
		s.paintStatusBar(buf, 0, h-1, w)
	}
}

func (s *AppShell) paintSidebar(buf *buffer.Buffer, x, y, w, h int) {
	// Fill background
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			buf.SetCell(x+col, y+row, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    s.style.SidebarBg,
			})
		}
	}

	row := 1
	col := 1

	// Panel list
	s.mu.RLock()
	items := s.panelItems
	sections := s.sections
	s.mu.RUnlock()

	if len(items) > 0 {
		s.drawText(buf, x+1, y+row, "Panels", w-2, s.style.SidebarTitle, s.style.SidebarBg, true)
		row++
		for _, item := range items {
			if row >= h-1 {
				break
			}
			label := item.label
			if item.icon != "" {
				label = item.icon + " " + label
			}
			s.drawText(buf, x+1, y+row, label, w-2, s.style.SidebarItem, s.style.SidebarBg, false)
			row++
		}
		row++ // gap
	}

	// Info sections
	for _, sec := range sections {
		if row >= h-1 {
			break
		}
		s.drawText(buf, x+col, y+row, sec.Title, w-2, s.style.SidebarTitle, s.style.SidebarBg, true)
		row++
		for _, line := range sec.Lines {
			if row >= h-1 {
				break
			}
			s.drawText(buf, x+1, y+row, line, w-2, s.style.SidebarItem, s.style.SidebarBg, false)
			row++
		}
		row++ // gap
	}
}

func (s *AppShell) paintStatusBar(buf *buffer.Buffer, x, y, w int) {
	// Fill background
	for col := 0; col < w; col++ {
		buf.SetCell(x+col, y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Bg:    s.style.StatusBg,
			Fg:    s.style.StatusFg,
		})
	}

	s.mu.RLock()
	activity := s.statusActivity
	tool := s.statusTool
	hint := s.statusHint
	s.mu.RUnlock()

	// Left: activity
	s.drawText(buf, x+1, y, activity, w/3, s.style.StatusFg, s.style.StatusBg, false)

	// Center: tool
	if tool != "" {
		centerX := x + w/2 - len(tool)/2
		if centerX > x+len(activity)+2 {
			s.drawText(buf, centerX, y, tool, len(tool), s.style.StatusFg, s.style.StatusBg, false)
		}
	}

	// Right: hint
	if hint != "" {
		hintX := x + w - 1 - len(hint)
		if hintX > x+1 {
			s.drawText(buf, hintX, y, hint, len(hint), s.style.StatusFg, s.style.StatusBg, false)
		}
	}
}

func (s *AppShell) drawText(buf *buffer.Buffer, x, y int, text string, maxW int, fg, bg buffer.Color, bold bool) {
	if maxW <= 0 {
		return
	}
	if len(text) > maxW {
		text = text[:maxW]
	}
	flags := buffer.StyleFlags(0)
	if bold {
		flags |= buffer.Bold
	}
	for i, r := range text {
		if x+i >= buf.Width {
			break
		}
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:   r,
			Width:  1,
			Fg:     fg,
			Bg:     bg,
			Flags:  flags,
		})
	}
}

// Ensure AppShell satisfies Panel interface
var _ Panel = (*AppShell)(nil)

// Re-export key methods for convenience

// Push opens a new panel on top of the stack.
func (s *AppShell) Push(p Panel) { s.panels.Push(p) }

// Pop closes the topmost panel.
func (s *AppShell) Pop() Panel { return s.panels.Pop() }

// ActivePanel returns the active (topmost) panel.
func (s *AppShell) ActivePanel() Panel { return s.panels.Active() }

// CloseAllPanels pops all panels except root.
func (s *AppShell) CloseAllPanels() { s.panels.CloseAll() }

// PanelDepth returns the number of panels in the stack.
func (s *AppShell) PanelDepth() int { return s.panels.Depth() }

// HelpText returns the keybinding help text from the EventRouter.
// (AppShell doesn't own the router — caller calls this with the router.)
func ShellHelpText(router *EventRouter) string {
	return router.HelpText()
}

// SplitLines splits text into lines for sidebar display.
func SplitLines(text string, maxW int) []string {
	if maxW <= 0 {
		return []string{text}
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		for len(line) > maxW {
			lines = append(lines, line[:maxW])
			line = line[maxW:]
		}
		lines = append(lines, line)
	}
	return lines
}