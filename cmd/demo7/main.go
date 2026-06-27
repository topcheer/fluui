// Package main implements a production-grade AI Agent demo showcasing
// all Fluui features from Phase 1 through Phase 9.
//
// This demo simulates a realistic AI agent workflow with:
//   - Animated startup splash screen with ASCII art logo
//   - Full AI conversation loop with streaming responses
//   - Multi-step tool chains (read_file → analyze → respond)
//   - Thinking, ToolCall, ToolResult, Error, and markdown blocks
//   - Theme cycling, search (Ctrl+F), clipboard (OSC52)
//   - Virtual scrolling with 100+ blocks
//   - Help overlay modal
//
// CLI flags:
//
//	--theme <name>    Set initial theme (dracula, nord, gruvbox, solarized, tokyo)
//	--blocks <n>      Generate N test blocks on startup
//	--no-animation    Skip startup animation
//
// Key bindings:
//
//	Enter       Send message
//	Up / Down   Input history
//	Ctrl+F      Search conversation
//	T           Cycle theme forward
//	Shift+T     Cycle theme backward
//	Y           Yank last block to clipboard (OSC52)
//	?           Toggle help overlay
//	V           Generate 100 virtual scroll blocks
//	Tab         Toggle thinking block collapse
//	Q / Esc     Quit
//	Ctrl+C      Force quit
//	Mouse       Wheel to scroll, click to collapse blocks
package main

import (
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/overlay"
	"github.com/topcheer/fluui/theme"
)

// ─── Constants ─────────────────────────────────────────────────────────────

const version = "v1.0.0"

// asciiLogo is the Fluui ASCII art logo rendered line-by-line during startup.
var asciiLogo = []string{
	"    ███████╗      ██╗██████╗ ██████╗ ██╗███╗   ██╗",
	"    ██╔════╝      ██║██╔══██╗██╔══██╗██║████╗  ██║",
	"    ██╗     █████╗██║██████╔╝██║  ██║██║██╔██╗ ██║",
	"    ██║     ╚════╝██║██╔══██╗██║  ██║██║██║╚██╗██║",
	"    ╚██████╗     ██║██║  ██║██████╔╝██║██║ ╚████║",
	"     ╚═════╝     ╚═╝╚═╝  ╚═╝╚═════╝ ╚═╝╚═╝  ╚═══╝",
}

// agentScenarios holds pre-built AI agent workflows that demonstrate
// multi-step tool chains with thinking, tool calls, and rich responses.
var agentScenarios = []agentScenario{
	{
		trigger: "analyze code",
		thinking: "The user wants me to analyze code. I'll need to:\n" +
			"1. Read the file to understand its structure\n" +
			"2. Identify potential issues and improvements\n" +
			"3. Provide actionable recommendations",
		steps: []toolStep{
			{name: "read_file", args: `{"path": "main.go"}`, result: `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}

// TODO: Add error handling
// FIXME: Hardcoded values should be configurable`},
			{name: "analyze", args: `{"type": "complexity", "target": "main.go"}`, result: `Analysis complete:
- Complexity: Low (score: 2)
- Lines of code: 10
- Issues found: 2
- TODOs: 1
- FIXMEs: 1
- Test coverage: 0%`},
		},
		response: "## Code Analysis Report\n\n" +
			"I've analyzed `main.go` and found the following:\n\n" +
			"### Summary\n\n" +
			"| Metric | Value |\n" +
			"|--------|-------|\n" +
			"| Complexity | Low (2) |\n" +
			"| Lines | 10 |\n" +
			"| Issues | 2 |\n" +
			"| Test Coverage | 0% |\n\n" +
			"### Findings\n\n" +
			"1. **TODO**: Add error handling — no `err` checks anywhere\n" +
			"2. **FIXME**: Hardcoded values should be configurable\n\n" +
			"### Recommendations\n\n" +
			"- Add `error` return checks after I/O operations\n" +
			"- Extract constants into a config file\n" +
			"- Write unit tests for `main()`\n\n" +
			"```go\n// Improved version:\nfunc main() {\n    if err := run(); err != nil {\n        log.Fatal(err)\n    }\n}\n```\n",
	},
	{
		trigger: "search web",
		thinking: "User wants to search the web. I'll perform the search and summarize results.",
		steps: []toolStep{
			{name: "search_web", args: `{"query": "best Go TUI libraries 2024"}`, result: `Top results:
1. Bubble Tea — 28k stars, Elm architecture
2. tview — 12k stars, widget-based
3. termui — 13k stars, dashboard layouts
4. Fluui — AI-native TUI library
5. gofpb — Low-level terminal I/O`},
		},
		response: "## Search Results: Best Go TUI Libraries\n\n" +
			"Based on my search, here are the top contenders:\n\n" +
			"| Library | Stars | Architecture | Best For |\n" +
			"|---------|-------|-------------|----------|\n" +
			"| Bubble Tea | 28k | Elm MVU | Structured apps |\n" +
			"| tview | 12k | Widget | Forms, tables |\n" +
			"| termui | 13k | Grid | Dashboards |\n" +
			"| Fluui | New | Block-based | AI chat |\n\n" +
			"**Key insight**: Fluui is the only library purpose-built for AI agent interfaces with streaming content blocks.\n",
	},
	{
		trigger: "calculate",
		thinking: "Simple arithmetic. Let me compute this step by step.",
		steps: []toolStep{
			{name: "calculate", args: `{"expression": "user_input"}`, result: "Result: computed"},
		},
		response: "The calculation is complete.\n\n" +
			"> Tip: For complex expressions, I follow standard PEMDAS order of operations.\n",
	},
}

// ─── Types ─────────────────────────────────────────────────────────────────

// agentScenario defines a complete AI agent workflow with tool chain.
type agentScenario struct {
	trigger  string     // substring to match against user input
	thinking string     // thinking block content
	steps    []toolStep // tool call chain
	response string     // final markdown response
}

type toolStep struct {
	name   string
	args   string
	result string
}

// ─── Main ──────────────────────────────────────────────────────────────────

func main() {
	// Parse CLI flags.
	themeFlag := flag.String("theme", "", "Initial theme name (dracula, nord, gruvbox, solarized, tokyo)")
	blocksFlag := flag.Int("blocks", 0, "Generate N test blocks on startup")
	noAnim := flag.Bool("no-animation", false, "Skip startup animation")
	flag.Parse()

	// Initialize the terminal application.
	a, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer a.Close()

	w, h := a.Size()

	// Apply initial theme if specified.
	if *themeFlag != "" {
		applyThemeByName(*themeFlag)
	}

	// --- ChatApp setup ---
	chat := app.NewChatApp(w, h)
	chat.SetInputHeight(3)

	// --- Search mode ---
	searchMode := app.NewSearchMode()

	// --- State ---
	var streamMu sync.Mutex
	streaming := false
	helpVisible := false
	startupPhase := true
	startupFrame := 0

	// --- Input line ---
	inputLine := app.NewInputLineWithHandler("> ", func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}

		// Built-in commands.
		switch text {
		case "/clear":
			chat.Clear()
			a.MarkDirty()
			return
		case "/help":
			helpVisible = !helpVisible
			a.MarkDirty()
			return
		case "/quit", "/exit":
			a.Quit()
			return
		}

		// Prevent concurrent streams.
		streamMu.Lock()
		if streaming {
			streamMu.Unlock()
			eb := block.NewErrorBlockWithMessage(
				fmt.Sprintf("err-%d", chat.Container().Len()),
				"Already processing a request. Please wait for the current response to complete.",
			)
			chat.Container().AddBlock(eb)
			a.MarkDirty()
			return
		}
		streaming = true
		streamMu.Unlock()

		go runAgentWorkflow(a, chat, text, &streamMu, &streaming)
	})
	chat.SetInputLine(inputLine)

	// Seed history with example prompts.
	for _, s := range agentScenarios {
		inputLine.AddHistory(s.trigger)
	}
	inputLine.AddHistory("Tell me about yourself")
	inputLine.AddHistory("What can you do?")

	// --- Help modal ---
	helpModal := createHelpModal()

	// --- Key handler ---
	a.OnKey(func(k *term.KeyEvent) {
		// Help modal takes priority.
		if helpVisible {
			if k.Key == term.KeyEscape || k.Key == term.KeyEnter {
				helpVisible = false
				a.MarkDirty()
				return
			}
			helpModal.HandleKey(k)
			a.MarkDirty()
			return
		}

		// Search mode.
		if searchMode.IsActive() {
			if searchMode.HandleKey(k) {
				searchMode.UpdateQuery(searchMode.Query(), chat.Container().Blocks())
				a.MarkDirty()
				return
			}
		}

		// Ctrl+F: toggle search.
		if k.Modifiers&term.ModCtrl != 0 && (k.Rune == 'f' || k.Rune == 'F') {
			if searchMode.IsActive() {
				searchMode.NextMatch()
			} else {
				searchMode.StartSearch()
			}
			a.MarkDirty()
			return
		}

		switch {
		// Quit.
		case k.Rune == 'Q' && k.Modifiers == 0:
			a.Quit()
			return
		case k.Key == term.KeyEscape:
			a.Quit()
			return

		// Theme cycling.
		case k.Rune == 'T' && k.Modifiers == 0:
			chat.CycleTheme()
			a.MarkDirty()
			return

		// Yank (OSC52).
		case k.Rune == 'Y' && k.Modifiers == 0:
			if text, ok := chat.LastBlockText(); ok && text != "" {
				a.Copy(text)
			}
			return

		// Help.
		case k.Rune == '?' && k.Modifiers == 0:
			helpVisible = !helpVisible
			a.MarkDirty()
			return

		// Virtual scroll generator.
		case k.Rune == 'V' && k.Modifiers == 0:
			go generateVirtualBlocks(a, chat, 100)
			return

		// Toggle thinking.
		case k.Key == term.KeyTab:
			toggleLastThinking(chat)
			a.MarkDirty()
			return

		// Scroll.
		case k.Key == term.KeyPageUp:
			chat.ScrollView().ScrollUp(h)
			a.MarkDirty()
			return
		case k.Key == term.KeyPageDown:
			chat.ScrollView().ScrollDown(h)
			a.MarkDirty()
			return
		case k.Key == term.KeyHome:
			chat.ScrollView().ScrollTo(0)
			a.MarkDirty()
			return
		case k.Key == term.KeyEnd:
			chat.ScrollToBottom()
			a.MarkDirty()
			return
		}

		// Remaining keys go to input.
		inputLine.HandleKey(k)
		a.MarkDirty()
	})

	// --- Mouse handler ---
	a.OnMouse(func(m *term.MouseEvent) {
		if m.Action == term.MouseWheel {
			switch m.Button {
			case term.MouseWheelUp:
				chat.ScrollUp()
				a.MarkDirty()
			case term.MouseWheelDown:
				chat.ScrollDown()
				a.MarkDirty()
			}
			return
		}
		if m.Action == term.MouseDown && m.Button == term.MouseLeft {
			handleBlockClick(chat, m.Y)
			a.MarkDirty()
		}
	})

	// --- Paint handler ---
	a.OnPaint(func(buf *buffer.Buffer) {
		bw, bh := buf.Width, buf.Height
		chat.SetSize(bw, bh)

		// Startup animation.
		if startupPhase {
			drawStartupScreen(buf, bw, bh, startupFrame)
			startupFrame++
			if !*noAnim && startupFrame > 25 {
				startupPhase = false
				showWelcomeMessage(chat, a)
				if *blocksFlag > 0 {
					generateVirtualBlocks(a, chat, *blocksFlag)
				}
			} else if *noAnim {
				startupPhase = false
				showWelcomeMessage(chat, a)
				if *blocksFlag > 0 {
					generateVirtualBlocks(a, chat, *blocksFlag)
				}
			}
			return
		}

		// Normal render.
		chat.Render(buf)
		drawTitleBar(buf, bw, chat, searchMode)
		drawStatusBar(buf, bw, bh, chat, searchMode, streaming)

		// Search bar.
		if searchMode.IsActive() {
			searchMode.RenderSearchBar(buf, bw, bh-2)
		}

		// Theme toast.
		if toast, ok := chat.ThemeToast(); ok {
			drawThemeToast(buf, bw, toast)
		}

		// Help modal.
		if helpVisible {
			helpModal.Measure(component.Bounded(bw, bh))
			helpModal.SetBounds(component.Rect{X: 0, Y: 0, W: bw, H: bh})
			helpModal.Paint(buf)
		}
	})

	if err := a.Run(); err != nil {
		panic(err)
	}
}

// ─── Startup Screen ────────────────────────────────────────────────────────

// drawStartupScreen renders the animated splash screen.
// The logo lines appear one at a time, followed by a loading indicator.
func drawStartupScreen(buf *buffer.Buffer, w, h int, frame int) {
	t := theme.Get()

	// Background fill.
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: t.Bg})
		}
	}

	// Calculate vertical centering.
	logoH := len(asciiLogo)
	startY := (h - logoH - 6) / 2
	if startY < 2 {
		startY = 2
	}

	// Draw logo lines up to the current frame.
	logoStyle := buffer.Style{
		Fg:    t.Accent,
		Bg:    t.Bg,
		Flags: buffer.Bold,
	}
	linesToShow := frame / 3
	if linesToShow > logoH {
		linesToShow = logoH
	}
	for i := 0; i < linesToShow; i++ {
		line := asciiLogo[i]
		lineW := buffer.StringWidth(line)
		x := (w - lineW) / 2
		if x < 0 {
			x = 0
		}
		buf.DrawTextClamped(x, startY+i, line, logoStyle)
	}

	// Version tag.
	if frame > logoH*3 {
		verStyle := buffer.Style{Fg: t.BorderMuted, Bg: t.Bg, Flags: buffer.Italic}
		verText := fmt.Sprintf("  %s  ", version)
		verW := buffer.StringWidth(verText)
		vx := (w - verW) / 2
		buf.DrawTextClamped(vx, startY+logoH+1, verText, verStyle)
	}

	// Loading dots animation.
	if frame > logoH*3+5 {
		dotCount := (frame / 5) % 4
		dots := strings.Repeat(".", dotCount) + strings.Repeat(" ", 3-dotCount)
		dotStyle := buffer.Style{Fg: t.Fg, Bg: t.Bg}
		label := "Loading" + dots
		lw := buffer.StringWidth(label)
		lx := (w - lw) / 2
		buf.DrawTextClamped(lx, startY+logoH+3, label, dotStyle)
	}

	// Tip text.
	if frame > logoH*3+10 {
		tipStyle := buffer.Style{Fg: t.BorderMuted, Bg: t.Bg}
		tip := "Press ? for help  |  Type a message and press Enter to start"
		tw := buffer.StringWidth(tip)
		tx := (w - tw) / 2
		if tx < 0 {
			tx = 0
		}
		buf.DrawTextClamped(tx, h-2, tip, tipStyle)
	}
}

// ─── Welcome Message ───────────────────────────────────────────────────────

// showWelcomeMessage populates the chat with an initial assistant greeting
// and an error block demo.
func showWelcomeMessage(chat *app.ChatApp, a *fluui.App) {
	welcome := chat.AddAssistantText()
	welcome.AppendDelta(
		"## Welcome to Fluui Demo7 — Production AI Agent\n\n" +
			"This is a **production-grade** AI agent demo showcasing all Phase 1-9 features.\n\n" +
			"### Try these commands:\n\n" +
			"1. `analyze code` — Watch a multi-step agent workflow with tool chains\n" +
			"2. `search web` — See tool call + result + markdown table response\n" +
			"3. `calculate` — Quick computation with thinking block\n\n" +
			"### Interactive Features:\n\n" +
			"- **Ctrl+F** — Search conversation\n" +
			"- **T** — Cycle theme (5 built-in themes)\n" +
			"- **Y** — Copy to clipboard (OSC52)\n" +
			"- **V** — Generate 100 blocks (virtual scroll demo)\n" +
			"- **?** — Help overlay\n" +
			"- **Tab** — Toggle thinking block\n" +
			"- **Mouse** — Click to collapse blocks, wheel to scroll\n\n" +
			"> Built with Fluui " + version + " — AI-native TUI library for Go\n",
	)
	welcome.Complete()

	// Error block demo.
	eb := block.NewErrorBlockWithMessage("err-init",
		"This is an ErrorBlock. In production, this would show API failures,\n"+
			"streaming errors, or tool execution failures.")
	chat.Container().AddBlock(eb)

	chat.ScrollToBottom()
	a.MarkDirty()
}

// ─── Agent Workflow ────────────────────────────────────────────────────────

// runAgentWorkflow simulates a complete AI agent response cycle.
// It demonstrates: user message → thinking → tool chain → markdown response.
func runAgentWorkflow(a *fluui.App, chat *app.ChatApp, userText string, mu *sync.Mutex, streaming *bool) {
	defer func() {
		mu.Lock()
		*streaming = false
		mu.Unlock()
		a.MarkDirty()
	}()

	// Add user message.
	chat.AddUserMessage(userText)
	a.MarkDirty()

	// Find matching scenario or create generic response.
	scenario := findScenario(userText)
	if scenario == nil {
		genericResponse(a, chat, userText)
		return
	}

	// Phase 1: Thinking block.
	time.Sleep(300 * time.Millisecond)
	thinking := chat.AddThinking()
	streamText(a, thinking, scenario.thinking, 15*time.Millisecond)
	thinking.Complete()

	// Phase 2: Tool chain.
	for _, step := range scenario.steps {
		time.Sleep(200 * time.Millisecond)
		tc := chat.AddToolCall(step.name, step.args)
		a.MarkDirty()
		time.Sleep(400 * time.Millisecond)
		tc.Complete()

		time.Sleep(150 * time.Millisecond)
		tr := chat.AddToolResult()
		streamText(a, tr, step.result, 10*time.Millisecond)
		tr.Complete()
	}

	// Phase 3: Markdown response.
	time.Sleep(200 * time.Millisecond)
	asst := chat.AddAssistantText()
	streamText(a, asst, scenario.response, 30*time.Millisecond)
	asst.Complete()

	chat.ScrollToBottom()
	a.MarkDirty()
}

// streamText appends text to a streaming block in small chunks.
func streamText(a *fluui.App, blk interface {
	AppendDelta(string)
}, text string, delay time.Duration) {
	words := strings.Fields(text)
	for i, word := range words {
		delta := word
		if i > 0 {
			delta = " " + word
		}
		// Preserve line breaks.
		if strings.Contains(word, "\n") {
			blk.AppendDelta(delta)
		} else {
			blk.AppendDelta(delta)
		}
		a.MarkDirty()
		time.Sleep(delay)
	}
	// Add trailing content if any.
	if strings.HasSuffix(text, "\n") {
		blk.AppendDelta("\n")
	}
}

// findScenario matches user input to a pre-built scenario.
func findScenario(text string) *agentScenario {
	lower := strings.ToLower(text)
	for i := range agentScenarios {
		s := &agentScenarios[i]
		if strings.Contains(lower, s.trigger) {
			return s
		}
	}
	return nil
}

// genericResponse handles unrecognized input with a markdown showcase.
func genericResponse(a *fluui.App, chat *app.ChatApp, userText string) {
	thinking := chat.AddThinking()
	thinking.AppendDelta(fmt.Sprintf("The user said: %q\nI'll provide a helpful markdown response.", userText))
	a.MarkDirty()
	time.Sleep(500 * time.Millisecond)
	thinking.Complete()

	time.Sleep(200 * time.Millisecond)
	asst := chat.AddAssistantText()
	response := fmt.Sprintf(
		"You said: **%s**\n\n"+
			"I'm a simulated AI agent. Here's what I can demonstrate:\n\n"+
			"### Markdown Features\n\n"+
			"- **Bold** and *italic* text\n"+
			"- `inline code`\n"+
			"- Code blocks with syntax highlighting:\n\n"+
			"```go\nfunc main() {\n    fmt.Println(\"Hello from Fluui!\")\n}\n```\n\n"+
			"- Blockquotes for tips\n"+
			"- Tables for structured data\n\n" +
			"> Try: `analyze code`, `search web`, or `calculate`\n",
		userText,
	)
	streamText(a, asst, response, 30*time.Millisecond)
	asst.Complete()

	chat.ScrollToBottom()
	a.MarkDirty()
}

// ─── UI Rendering ──────────────────────────────────────────────────────────

// drawTitleBar renders the top title bar with app name and theme indicator.
func drawTitleBar(buf *buffer.Buffer, w int, chat *app.ChatApp, search *app.SearchMode) {
	t := theme.Get()

	// Background.
	for x := 0; x < w; x++ {
		buf.SetCell(x, 0, buffer.Cell{Rune: ' ', Width: 1, Bg: t.BorderMuted})
	}

	// Title.
	titleStyle := buffer.Style{Fg: t.Accent, Bg: t.BorderMuted, Flags: buffer.Bold}
	title := fmt.Sprintf(" Fluui Agent %s — AI TUI Demo", version)
	buf.DrawTextClamped(1, 0, title, titleStyle)

	// Theme name on the right.
	themeLabel := fmt.Sprintf(" %s ", t.Name)
	ts := buffer.StringWidth(themeLabel)
	buf.DrawTextClamped(w-ts-1, 0, themeLabel, titleStyle)

	// Search active indicator.
	if search.IsActive() {
		srLabel := " SEARCH "
		srW := buffer.StringWidth(srLabel)
		srStyle := buffer.Style{Fg: t.SearchMatch, Bg: t.BorderMuted, Flags: buffer.Bold}
		buf.DrawTextClamped(w-ts-srW-3, 0, srLabel, srStyle)
	}
}

// drawStatusBar renders the bottom status bar with block stats and key hints.
func drawStatusBar(buf *buffer.Buffer, w, h int, chat *app.ChatApp, search *app.SearchMode, streaming bool) {
	t := theme.Get()

	statusY := h - 1
	for x := 0; x < w; x++ {
		buf.SetCell(x, statusY, buffer.Cell{Rune: ' ', Width: 1, Bg: t.BorderMuted})
	}

	style := buffer.Style{Fg: t.Fg, Bg: t.BorderMuted}

	streamCount := 0
	for _, b := range chat.Container().Blocks() {
		if b.State() == block.BlockStreaming {
			streamCount++
		}
	}

	status := fmt.Sprintf(" %d blocks | %d streaming | Ctrl+F:search  T:theme  Y:yank  ?:help  V:vscroll  Q:quit",
		chat.Container().Len(), streamCount)
	if streaming {
		status += " [BUSY]"
	}
	buf.DrawTextClamped(1, statusY, status, style)
}

// drawThemeToast renders the temporary theme name notification.
func drawThemeToast(buf *buffer.Buffer, w int, toast string) {
	t := theme.Get()
	toastStyle := buffer.Style{
		Fg:    t.Bg,
		Bg:    t.Accent,
		Flags: buffer.Bold,
	}
	text := fmt.Sprintf(" Theme: %s ", toast)
	tx := w - buffer.StringWidth(text) - 1
	if tx >= 0 {
		buf.DrawText(tx, 1, text, toastStyle)
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────

// createHelpModal builds the help overlay modal with all key bindings.
func createHelpModal() *overlay.Modal {
	lines := []string{
		"FLUUI AGENT — KEY BINDINGS",
		"",
		"  Enter         Send message to AI agent",
		"  Up / Down     Navigate input history",
		"  Ctrl+F        Search conversation (Enter/Shift+Enter to cycle)",
		"  PageUp/Down   Scroll content",
		"  Home / End    Scroll to top / bottom",
		"  Tab           Toggle thinking block collapse",
		"  T             Cycle theme forward",
		"  Shift+T       Cycle theme backward",
		"  Y             Yank to clipboard (OSC52)",
		"  V             Generate 100 virtual scroll blocks",
		"  ?             Toggle this help overlay",
		"  Q / Esc       Quit",
		"  Ctrl+C        Force quit",
		"",
		"  Mouse wheel   Scroll up / down",
		"  Mouse click   Toggle block collapse",
		"",
		"  Commands: /clear  /help  /quit",
		"",
		"Press Esc or Enter to close.",
	}
	body := component.NewText(strings.Join(lines, "\n"))
	return overlay.NewModal("help", "Help", body, []string{"Close"})
}

// applyThemeByName sets the active theme by name.
func applyThemeByName(name string) {
	themes := theme.Builtin()
	lower := strings.ToLower(name)
	for _, t := range themes {
		if strings.ToLower(t.Name) == lower {
			theme.SetActive(t)
			return
		}
	}
}

// toggleLastThinking collapses/expands the most recent ThinkingBlock.
func toggleLastThinking(chat *app.ChatApp) {
	blocks := chat.Container().Blocks()
	for i := len(blocks) - 1; i >= 0; i-- {
		if tb, ok := blocks[i].(*block.ThinkingBlock); ok {
			tb.Toggle()
			return
		}
	}
}

// handleBlockClick finds the block at mouseY and toggles it if collapsible.
func handleBlockClick(chat *app.ChatApp, mouseY int) {
	sv := chat.ScrollView()
	bounds := sv.Bounds()
	offset := sv.Offset()
	contentY := mouseY - bounds.Y + offset
	if contentY < 0 {
		return
	}

	container := chat.Container()
	positions := container.BlockPositions()
	blocks := container.Blocks()

	for i, pos := range positions {
		if contentY >= pos.Y && contentY < pos.Y+pos.H {
			if i < len(blocks) {
				if tb, ok := blocks[i].(*block.ThinkingBlock); ok {
					tb.Toggle()
				}
			}
			return
		}
	}
}

// generateVirtualBlocks adds n AssistantTextBlocks to demo virtual scrolling.
func generateVirtualBlocks(a *fluui.App, chat *app.ChatApp, n int) {
	for i := 0; i < n; i++ {
		blk := block.NewAssistantTextBlock(fmt.Sprintf("vs-%d", i))
		blk.AppendDelta(fmt.Sprintf(
			"Block #%d — Virtual scroll test (%d total).\n"+
				"PaintVisible uses O(log n) binary search to paint only visible blocks.",
			i+1, n,
		))
		blk.Complete()
		chat.Container().AddBlock(blk)
	}
	chat.ScrollToBottom()
	a.MarkDirty()
}
