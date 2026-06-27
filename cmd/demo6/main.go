// Package main implements a comprehensive interactive demo showcasing
// all major Fluui features in a single application.
//
// This demo demonstrates:
//   - All content blocks: Thinking, ToolCall, ToolResult, AssistantText,
//     UserMessage, ErrorBlock
//   - Markdown rendering with syntax-highlighted code blocks
//   - Theme cycling (T key) with toast notification
//   - Input history navigation (Up/Down arrows)
//   - Help overlay modal (? key)
//   - Mouse interaction: click blocks to collapse/expand
//   - Virtual scrolling with 100+ blocks (V key to generate)
//   - OSC52 clipboard yank (Y key)
//
// Key bindings:
//
//	Enter      — send message (simulated AI response)
//	Up/Down    — navigate input history
//	PageUp/Down — scroll content
//	Home/End   — scroll to top / bottom
//	Tab        — toggle thinking block collapse
//	T          — cycle theme forward
//	Y          — yank last block text to clipboard (OSC52)
//	?          — toggle help overlay
//	V          — generate 100 blocks (virtual scroll demo)
//	Q / Esc    — quit
//	Ctrl+C     — quit (graceful)
package main

import (
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

// responseTemplates are pre-built simulated AI responses.
// Each template demonstrates different block types and markdown features.
var responseTemplates = []struct {
	prompt    string
	thinking  string
	toolName  string
	toolArgs  string
	toolResult string
	response  string
}{
	{
		prompt:    "What is Fluui?",
		thinking:  "The user is asking about Fluui. Let me explain what it is — a terminal UI library for building AI chat interfaces in Go.",
		toolName:  "",
		toolArgs:  "",
		toolResult: "",
		response: "**Fluui** is an AI-native TUI library written in Go.\n\n" +
			"Key features:\n\n" +
			"1. **No framework dependency** — built from scratch (termios to render engine)\n" +
			"2. **Double-buffer diff rendering** for flicker-free display\n" +
			"3. **Full markdown** with syntax highlighting via Chroma\n" +
			"4. **Streaming content blocks** for real-time AI output\n\n" +
			"```go\napp, _ := fluui.New()\nchat := app.NewChatApp(w, h)\nchat.AddUserMessage(\"Hello!\")\napp.Run()\n```\n\n" +
			"It powers terminal-based AI assistants with rich, interactive content.",
	},
	{
		prompt:     "Search for Go TUI libraries",
		thinking:   "The user wants to find Go TUI libraries. I'll search and present the most popular options.",
		toolName:   "search_web",
		toolArgs:   `{"query": "Go TUI library terminal"}`,
		toolResult: "Found 5 results:\n1. Bubble Tea (charmbracelet) — Elm architecture\n2. tview (rivo) — Rich interactive widgets\n3. termui (gizak) — Dashboard-style layouts\n4. gofpb (gdamore) — Low-level terminal I/O\n5. Fluui (topcheer) — AI-native TUI",
		response: "Here are the top Go TUI libraries:\n\n" +
			"| Library | Style | Stars |\n" +
			"|---------|-------|-------|\n" +
			"| Bubble Tea | Elm MVU | 28k+ |\n" +
			"| tview | Widget-based | 12k+ |\n" +
			"| termui | Dashboard | 13k+ |\n" +
			"| Fluui | AI-native | New |\n\n" +
			"Each has different strengths:\n\n" +
			"- **Bubble Tea**: Great for structured apps with the Elm pattern\n" +
			"- **tview**: Best for form-heavy interactive applications\n" +
			"- **termui**: Ideal for monitoring dashboards\n" +
			"- **Fluui**: Purpose-built for AI chat interfaces\n\n" +
			"`Fluui` is unique in its focus on streaming AI content blocks.",
	},
	{
		prompt:    "Calculate 2+3*4",
		thinking:  "Need to evaluate the expression 2+3*4. Following order of operations: 3*4=12, then 2+12=14.",
		toolName:  "calculate",
		toolArgs:  `{"expression": "2+3*4"}`,
		toolResult: "Result: 2 + 3*4 = 14",
		response: "The result of `2 + 3 * 4` is **14**.\n\n" +
			"Following the standard order of operations (PEMDAS):\n\n" +
			"1. Multiplication first: `3 * 4 = 12`\n" +
			"2. Then addition: `2 + 12 = 14`\n\n" +
			"> Tip: You can override precedence with parentheses, e.g. `(2+3)*4 = 20`",
	},
}

// helpText returns the key bindings shown in the help overlay.
var helpText = []string{
	"KEY BINDINGS",
	"",
	"  Enter         Send message",
	"  Up / Down     Input history",
	"  PgUp / PgDn   Scroll content",
	"  Home / End    Scroll to top/bottom",
	"  Tab           Toggle thinking block",
	"  T             Cycle theme",
	"  Y             Yank to clipboard (OSC52)",
	"  ?             Toggle this help",
	"  V             Generate 100 blocks",
	"  Q / Esc       Quit",
	"  Ctrl+C        Force quit",
	"",
	"  Mouse wheel   Scroll",
	"  Mouse click   Toggle block collapse",
	"",
	"Press Esc or Enter to close.",
}

func main() {
	a, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer a.Close()

	w, h := a.Size()

	// --- ChatApp setup ---
	chat := app.NewChatApp(w, h)
	chat.SetInputHeight(2)

	// InputLine with submit handler.
	var streamMu sync.Mutex
	streaming := false

	inputLine := app.NewInputLineWithHandler("> ", func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}

		// Built-in commands
		switch text {
		case "/clear":
			chat.Clear()
			a.MarkDirty()
			return
		case "/help":
			toggleHelp()
			a.MarkDirty()
			return
		}

		// Prevent concurrent streams.
		streamMu.Lock()
		if streaming {
			streamMu.Unlock()
			// Show error block for busy state.
			errBlock := block.NewErrorBlockWithMessage(
				fmt.Sprintf("err-%d", chat.Container().Len()),
				"Already streaming a response. Please wait...",
			)
			chat.Container().AddBlock(errBlock)
			a.MarkDirty()
			return
		}
		streaming = true
		streamMu.Unlock()

		go simulateAIResponse(a, chat, text, &streamMu, &streaming)
	})
	chat.SetInputLine(inputLine)

	// Seed input history with example prompts.
	for _, tmpl := range responseTemplates {
		inputLine.AddHistory(tmpl.prompt)
	}

	// --- Help modal ---
	helpBody := component.NewText(strings.Join(helpText, "\n"))
	helpModal := overlay.NewModal("help-modal", "Help — Key Bindings", helpBody, []string{"Close"})
	helpModal.SetVisible(false)

	// helpVisible is toggled by '?' and managed via closure.
	helpVisible := false
	toggleHelp = func() {
		helpVisible = !helpVisible
		helpModal.SetVisible(helpVisible)
	}

	// --- Welcome message ---
	welcome := chat.AddAssistantText()
	welcome.AppendDelta(
		"**Welcome to Fluui Demo6** — Full Interactive Showcase\n\n" +
			"This demo demonstrates all major features:\n\n" +
			"- All content block types (thinking, tool call, tool result, error)\n" +
			"- Markdown rendering with code highlighting\n" +
			"- Theme cycling (press **T**)\n" +
			"- Input history (press **Up/Down**)\n" +
			"- Help overlay (press **?**)\n" +
			"- Virtual scrolling (press **V** for 100 blocks)\n" +
			"- Mouse: click blocks to collapse/expand\n\n" +
			"Type a message and press Enter, or try:\n" +
			"`What is Fluui?` · `Search for Go TUI libraries` · `Calculate 2+3*4`\n",
	)
	welcome.Complete()

	// --- Error block demo ---
	errBlock := block.NewErrorBlockWithMessage("err-welcome",
		"This is an ErrorBlock demo. Errors are always expanded by default.\n"+
			"In real usage, this would show API failures or streaming errors.")
	chat.Container().AddBlock(errBlock)

	// --- Key handler ---
	a.OnKey(func(k *term.KeyEvent) {
		// If help modal is visible, route keys to it first.
		if helpVisible {
			if k.Key == term.KeyEscape || k.Key == term.KeyEnter {
				toggleHelp()
				a.MarkDirty()
				return
			}
			helpModal.HandleKey(k)
			a.MarkDirty()
			return
		}

		switch {
		// Quit: Q or Esc
		case k.Rune == 'Q' && k.Modifiers == 0:
			a.Quit()
			return
		case k.Key == term.KeyEscape:
			a.Quit()
			return

		// Theme cycling: T (uppercase, so lowercase 't' can still type)
		case k.Rune == 'T' && k.Modifiers == 0:
			chat.CycleTheme()
			a.MarkDirty()
			return

		// Yank: Y (OSC52 clipboard copy of last block text)
		case k.Rune == 'Y' && k.Modifiers == 0:
			yankLastBlock(a, chat)
			return

		// Help overlay: ?
		case k.Rune == '?' && k.Modifiers == 0:
			toggleHelp()
			a.MarkDirty()
			return

		// Virtual scroll demo: V (generate 100 blocks)
		case k.Rune == 'V' && k.Modifiers == 0:
			go generateBlocks(a, chat, 100)
			return

		// Tab: toggle last thinking block
		case k.Key == term.KeyTab:
			toggleLastThinking(chat)
			a.MarkDirty()
			return

		// Scroll keys
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

		// Everything else goes to the input line.
		inputLine.HandleKey(k)
		a.MarkDirty()
	})

	// --- Mouse handler ---
	a.OnMouse(func(m *term.MouseEvent) {
		// Scroll wheel
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

		// Click: toggle block at position
		if m.Action == term.MouseDown && m.Button == term.MouseLeft {
			handleBlockClick(chat, m.Y)
			a.MarkDirty()
		}
	})

	// --- Paint handler ---
	a.OnPaint(func(buf *buffer.Buffer) {
		bw, bh := buf.Width, buf.Height

		// Sync chat size.
		chat.SetSize(bw, bh)

		// Render chat (background, scroll content, input line).
		chat.Render(buf)

		// Title bar.
		t := theme.Get()
		titleStyle := buffer.Style{
			Fg:    t.Accent,
			Bg:    t.Bg,
			Flags: buffer.Bold,
		}
		buf.DrawText(1, 0, " Fluui Demo6 — Full Interactive Showcase", titleStyle)

		// Theme indicator on the right side of the title bar.
		themeLabel := fmt.Sprintf("Theme: %s ", t.Name)
		buf.DrawTextClamped(bw-buffer.StringWidth(themeLabel)-1, 0, themeLabel, titleStyle)

		// Status bar.
		statusY := bh - 1
		statusBg := t.BorderMuted
		for x := 0; x < bw; x++ {
			buf.SetCell(x, statusY, buffer.Cell{Rune: ' ', Width: 1, Bg: statusBg})
		}
		statusStyle := buffer.Style{Fg: t.Fg, Bg: statusBg}

		streamCount := 0
		for _, b := range chat.Container().Blocks() {
			if b.State() == block.BlockStreaming {
				streamCount++
			}
		}
		status := fmt.Sprintf(" %d blocks | %d streaming | T:theme  Y:yank  ?:help  V:vscroll  Tab:thinking  Q:quit",
			chat.Container().Len(), streamCount)
		buf.DrawTextClamped(1, statusY, status, statusStyle)

		// Theme toast (auto-expires after 3s).
		if toast, ok := chat.ThemeToast(); ok {
			toastStyle := buffer.Style{
				Fg:    t.Bg,
				Bg:    t.Accent,
				Flags: buffer.Bold,
			}
			toastText := fmt.Sprintf(" Theme: %s ", toast)
			tx := bw - buffer.StringWidth(toastText) - 1
			ty := 1
			if tx >= 0 && ty < bh {
				buf.DrawText(tx, ty, toastText, toastStyle)
			}
		}

		// Help modal on top.
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

// toggleHelp is set in main() to control the help modal visibility.
var toggleHelp func()

// simulateAIResponse creates a realistic AI conversation with all block types.
// It picks the best matching template or generates a generic response.
func simulateAIResponse(a *fluui.App, chat *app.ChatApp, userText string, mu *sync.Mutex, streaming *bool) {
	defer func() {
		mu.Lock()
		*streaming = false
		mu.Unlock()
		a.MarkDirty()
	}()

	// Add user message.
	chat.AddUserMessage(userText)
	a.MarkDirty()

	// Find matching template, or use a generic one.
	tmpl := findTemplate(userText)
	if tmpl == nil {
		tmpl = &struct {
			prompt     string
			thinking   string
			toolName   string
			toolArgs   string
			toolResult string
			response   string
		}{
			thinking: fmt.Sprintf("The user said: %q. Let me provide a helpful response.", userText),
			response: fmt.Sprintf(
				"You said: **%s**\n\n"+
					"This is a simulated response demonstrating the `AssistantTextBlock` with markdown.\n\n"+
					"Features shown:\n"+
					"- **Bold** and *italic* text\n"+
					"- `Inline code` snippets\n"+
					"- Bullet lists\n"+
					"- Code blocks:\n\n"+
					"```go\nfmt.Println(\"Hello from Fluui!\")\n```\n\n"+
					"> Try pressing **T** to cycle themes, or **V** to generate 100 blocks for virtual scrolling.\n",
				userText,
			),
		}
	}

	// Phase 1: Thinking block.
	time.Sleep(300 * time.Millisecond)
	thinking := chat.AddThinking()
	thinking.AppendDelta(tmpl.thinking)
	a.MarkDirty()
	time.Sleep(400 * time.Millisecond)
	thinking.Complete()

	// Phase 2: Tool call + result (if applicable).
	if tmpl.toolName != "" {
		time.Sleep(200 * time.Millisecond)
		tc := chat.AddToolCall(tmpl.toolName, tmpl.toolArgs)
		a.MarkDirty()
		time.Sleep(300 * time.Millisecond)
		tc.Complete()

		time.Sleep(200 * time.Millisecond)
		tr := chat.AddToolResult()
		tr.AppendDelta(tmpl.toolResult)
		a.MarkDirty()
		time.Sleep(300 * time.Millisecond)
		tr.Complete()
	}

	// Phase 3: Assistant response with markdown.
	time.Sleep(200 * time.Millisecond)
	asst := chat.AddAssistantText()

	// Stream response in chunks for realistic effect.
	chunks := splitResponse(tmpl.response)
	for _, chunk := range chunks {
		asst.AppendDelta(chunk)
		a.MarkDirty()
		time.Sleep(80 * time.Millisecond)
	}
	asst.Complete()

	// Scroll to bottom to show the new content.
	chat.ScrollToBottom()
	a.MarkDirty()
}

// findTemplate finds a response template matching the user's input.
func findTemplate(text string) *struct {
	prompt     string
	thinking   string
	toolName   string
	toolArgs   string
	toolResult string
	response   string
} {
	lower := strings.ToLower(text)
	for i := range responseTemplates {
		t := &responseTemplates[i]
		if strings.Contains(lower, strings.ToLower(t.prompt)) ||
			strings.Contains(strings.ToLower(t.prompt), lower) {
			return t
		}
	}
	return nil
}

// splitResponse breaks a response string into stream-friendly chunks.
func splitResponse(s string) []string {
	parts := strings.SplitAfter(s, " ")
	if len(parts) <= 1 {
		return []string{s}
	}
	// Group 2-3 words per chunk for natural streaming.
	var chunks []string
	for i := 0; i < len(parts); i += 2 {
		end := i + 2
		if end > len(parts) {
			end = len(parts)
		}
		chunks = append(chunks, strings.Join(parts[i:end], ""))
	}
	return chunks
}

// toggleLastThinking finds the last ThinkingBlock and toggles its collapse state.
func toggleLastThinking(chat *app.ChatApp) {
	for _, b := range chat.Container().Blocks() {
		if tb, ok := b.(*block.ThinkingBlock); ok {
			tb.Toggle()
		}
	}
}

// yankLastBlock copies the text of the last meaningful block to clipboard.
func yankLastBlock(a *fluui.App, chat *app.ChatApp) {
	if text, ok := chat.LastBlockText(); ok && text != "" {
		a.Copy(text)
	}
}

// handleBlockClick determines which block was clicked and toggles it if collapsible.
func handleBlockClick(chat *app.ChatApp, mouseY int) {
	sv := chat.ScrollView()
	bounds := sv.Bounds()
	offset := sv.Offset()

	// Convert screen Y to content-local Y.
	contentY := mouseY - bounds.Y + offset
	if contentY < 0 {
		return
	}

	// Find the block at this position.
	container := chat.Container()
	positions := container.BlockPositions()
	blocks := container.Blocks()

	for i, pos := range positions {
		if contentY >= pos.Y && contentY < pos.Y+pos.H {
			if i < len(blocks) {
				// Toggle collapsible blocks.
				if tb, ok := blocks[i].(*block.ThinkingBlock); ok {
					tb.Toggle()
				}
			}
			return
		}
	}
}

// generateBlocks adds N blocks to demonstrate virtual scrolling.
// Uses lightweight blocks for instant creation.
func generateBlocks(a *fluui.App, chat *app.ChatApp, n int) {
	for i := 0; i < n; i++ {
		asst := block.NewAssistantTextBlock(fmt.Sprintf("vs-%d", i))
		asst.AppendDelta(fmt.Sprintf(
			"Block #%d — Virtual scrolling test item.\n"+
				"This block is part of a batch of %d generated blocks.\n"+
				"Only visible blocks are painted for O(log n) performance.",
			i+1, n,
		))
		asst.Complete()
		chat.Container().AddBlock(asst)
	}
	chat.ScrollToBottom()
	a.MarkDirty()
}
