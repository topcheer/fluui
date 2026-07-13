// Demo21: Component-Declarative Architecture (CDA) Showcase
//
// This demo shows how fluui's CDA framework eliminates the monolithic
// Model/Update/View pattern of bubbletea. Instead of a 236-field Model
// and 90-case Update switch, each panel owns its own state and the
// EventRouter + PanelManager + StateBus handle all routing.
package main

import (
	"fmt"
	"strings"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Chat Panel ───

type chatPanel struct {
	app.BasePanel
	bus     *app.StateBus
	messages []string
	cursor  int
}

func newChatPanel(bus *app.StateBus) *chatPanel {
	cp := &chatPanel{
		bus:      bus,
		messages: []string{"Welcome to CDA Demo!", "Press Tab to switch panels.", "Press Ctrl+R to toggle sidebar."},
	}
	// Subscribe to cross-panel events
	app.Subscribe[string](bus, "chat.add", func(msg string) {
		cp.messages = append(cp.messages, msg)
	})
	return cp
}

func (c *chatPanel) ID() string    { return "chat" }
func (c *chatPanel) Title() string { return "Chat" }

func (c *chatPanel) HandleKey(ev *term.KeyEvent) bool {
	switch ev.Rune {
	case 'j', 'k':
		// Vim-style scroll
		if ev.Rune == 'j' && c.cursor < len(c.messages)-1 {
			c.cursor++
		}
		if ev.Rune == 'k' && c.cursor > 0 {
			c.cursor--
		}
		return true
	case 'n':
		c.messages = append(c.messages, fmt.Sprintf("Message #%d", len(c.messages)+1))
		c.cursor = len(c.messages) - 1
		// Publish event for other panels
		app.Publish(c.bus, "chat.add", fmt.Sprintf("new message #%d", len(c.messages)))
		return true
	}
	return false
}

func (c *chatPanel) Paint(buf *buffer.Buffer, w, h int) {
	fg := buffer.NamedColor(buffer.NamedWhite)
	bg := buffer.RGB(30, 30, 40)
 accent := buffer.NamedColor(buffer.NamedCyan)

	// Fill background
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
		}
	}

	// Title bar
	title := " CDA Chat — j/k scroll, n=new, Tab=switch "
	drawText(buf, 0, 0, title, accent, bg, true)

	// Messages
	start := 0
	if len(c.messages) > h-2 {
		start = len(c.messages) - (h - 2)
	}
	for i, msg := range c.messages[start:] {
		y := 1 + i
		if y >= h {
			break
		}
		style := fg
		if i+start == c.cursor {
			style = accent
		}
		drawText(buf, 1, y, "> "+msg, style, bg, false)
	}
}

// ─── Files Panel ───

type filesPanel struct {
	app.BasePanel
	files  []string
	cursor int
}

func newFilesPanel() *filesPanel {
	return &filesPanel{
		files: []string{
			"main.go", "app.go", "config.yaml", "README.md",
			"test.go", "utils.go", "types.go",
		},
	}
}

func (f *filesPanel) ID() string    { return "files" }
func (f *filesPanel) Title() string { return "Files" }

func (f *filesPanel) HandleKey(ev *term.KeyEvent) bool {
	switch ev.Key {
	case term.KeyUp:
		if f.cursor > 0 {
			f.cursor--
		}
		return true
	case term.KeyDown:
		if f.cursor < len(f.files)-1 {
			f.cursor++
		}
		return true
	}
	if ev.Rune == 'j' {
		if f.cursor < len(f.files)-1 {
			f.cursor++
		}
		return true
	}
	if ev.Rune == 'k' {
		if f.cursor > 0 {
			f.cursor--
		}
		return true
	}
	return false
}

func (f *filesPanel) Paint(buf *buffer.Buffer, w, h int) {
	fg := buffer.NamedColor(buffer.NamedWhite)
	bg := buffer.RGB(30, 30, 40)
	accent := buffer.NamedColor(buffer.NamedGreen)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
		}
	}

	drawText(buf, 0, 0, " Files — j/k or arrows, Enter=select ", accent, bg, true)

	for i, file := range f.files {
		y := 1 + i
		if y >= h {
			break
		}
		marker := "  "
		style := fg
		if i == f.cursor {
			marker = "→ "
			style = accent
		}
		drawText(buf, 1, y, marker+file, style, bg, false)
	}
}

// ─── Terminal Panel ───

type terminalPanel struct {
	app.BasePanel
	lines  []string
	count  int
}

func newTerminalPanel() *terminalPanel {
	return &terminalPanel{
		lines: []string{
			"$ go build ./... ✓",
			"$ go test -race ./... ✓",
			"$ go vet ./... ✓",
		},
	}
}

func (t *terminalPanel) ID() string    { return "terminal" }
func (t *terminalPanel) Title() string { return "Terminal" }

func (t *terminalPanel) HandleKey(ev *term.KeyEvent) bool {
	if ev.Rune == 'r' {
		t.count++
		t.lines = append(t.lines, fmt.Sprintf("$ echo run #%d → ok", t.count))
		return true
	}
	return false
}

func (t *terminalPanel) Paint(buf *buffer.Buffer, w, h int) {
	fg := buffer.NamedColor(buffer.NamedWhite)
	bg := buffer.RGB(30, 30, 40)
	accent := buffer.NamedColor(buffer.NamedYellow)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
		}
	}

	drawText(buf, 0, 0, " Terminal — r=run command ", accent, bg, true)

	start := 0
	if len(t.lines) > h-2 {
		start = len(t.lines) - (h - 2)
	}
	for i, line := range t.lines[start:] {
		y := 1 + i
		if y >= h {
			break
		}
		drawText(buf, 1, y, line, fg, bg, false)
	}
}

// ─── Tab Switcher Panel (wraps the above 3 panels) ───

type tabPanel struct {
	app.BasePanel
	tabs    []app.Panel
	active  int
}

func (t *tabPanel) ID() string    { return "tabs" }
func (t *tabPanel) Title() string { return "Tabs" }

func (t *tabPanel) HandleKey(ev *term.KeyEvent) bool {
	// Tab switches between panels
	if ev.Key == term.KeyTab {
		t.active = (t.active + 1) % len(t.tabs)
		return true
	}
	// Delegate to active sub-panel
	return t.tabs[t.active].HandleKey(ev)
}

func (t *tabPanel) Paint(buf *buffer.Buffer, w, h int) {
	// Draw tab bar at top
	bg := buffer.RGB(40, 42, 54)
	activeBg := buffer.RGB(60, 62, 74)
	fg := buffer.NamedColor(buffer.NamedWhite)
	activeFg := buffer.NamedColor(buffer.NamedCyan)

	for x := 0; x < w; x++ {
		buf.SetCell(x, 0, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
	}

	tabW := w / len(t.tabs)
	for i, tab := range t.tabs {
		title := " " + tab.Title() + " "
		bgColor := bg
		fgColor := fg
		if i == t.active {
			bgColor = activeBg
			fgColor = activeFg
		}
		for j, r := range title {
			x := i*tabW + j
			if x < w {
				buf.SetCell(x, 0, buffer.Cell{Rune: r, Width: 1, Fg: fgColor, Bg: bgColor})
			}
		}
	}

	// Delegate painting to active panel
	t.tabs[t.active].Paint(buf, w, h)
}

// ─── Helper ───

func drawText(buf *buffer.Buffer, x, y int, text string, fg, bg buffer.Color, bold bool) {
	flags := buffer.StyleFlags(0)
	if bold {
		flags |= buffer.Bold
	}
	for i, r := range text {
		if x+i >= buf.Width {
			break
		}
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    fg,
			Bg:    bg,
			Flags: flags,
		})
	}
}

// ─── Main ───

func main() {
	a, err := fluui.New()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer a.Close()

	a.SetTitle("Fluui CDA Demo — Component-Declarative Architecture")

	// Create StateBus for cross-panel communication
	bus := app.NewStateBus()

	// Create panels
	chat := newChatPanel(bus)
	files := newFilesPanel()
	terminal := newTerminalPanel()

	// Tab panel wraps the 3 content panels
	tabs := &tabPanel{
		tabs:   []app.Panel{chat, files, terminal},
		active: 0,
	}

	// AppShell provides sidebar + panel area + status bar
	shell := app.NewAppShell(tabs)
	shell.ShowSidebar()
	shell.AddPanelItem("chat", "Chat", ">")
	shell.AddPanelItem("files", "Files", "📁")
	shell.AddPanelItem("terminal", "Terminal", "$")
	shell.AddSidebarSection("Version", []string{"fluui CDA", "demo21"})
	shell.AddSidebarSection("Keys", []string{
		"Tab  switch panel",
		"Ctrl+R sidebar",
		"Ctrl+Q quit",
		"j/k   navigate",
		"n     new msg",
		"r     run cmd",
	})

	// Create EventRouter with KeybindingManager
	km := component.NewKeybindingManager()
	router := app.NewEventRouter(shell.PanelManager(), km)

	// Global keybindings (replaces 50+ if-branches in handleKeyPress)
	router.RegisterGlobal("quit", "ctrl+q", "Quit", func() bool {
		a.Quit()
		return true
	})
	router.RegisterGlobal("toggle-sidebar", "ctrl+r", "Toggle sidebar", func() bool {
		shell.ToggleSidebar()
		return true
	})

	// Wire router
	a.OnKey(func(ev *term.KeyEvent) {
		router.HandleKey(ev)
	})

	// Update status bar on resize
	a.OnResize(func(w, h int) {
		shell.SetStatus("ready", fmt.Sprintf("%d panels", shell.PanelDepth()), "Ctrl+Q quit | Tab switch")
	})
	shell.SetStatus("ready", "3 panels", "Ctrl+Q quit | Tab switch")

	// Paint
	a.OnPaint(func(buf *buffer.Buffer) {
		w, h := a.Size()
		shell.Paint(buf, w, h)
	})

	// Run
	if err := a.Run(); err != nil {
		fmt.Println("Error:", err)
	}

	_ = strings.Builder{} // keep import
}