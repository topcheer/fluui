// Package main implements a real AI chat demo with tool calling support.
//
// Showcases all Fluui block types:
//   - UserMessage: user input
//   - ThinkingBlock: AI reasoning (if model supports reasoning_content)
//   - ToolCallBlock: function calls to mock tools
//   - ToolResultBlock: mock tool outputs
//   - AssistantTextBlock: final markdown response
//
// The demo defines simple mock tools (list_files, read_file, search, calc)
// that the LLM can call. Results are auto-generated and sent back so the
// LLM can produce a final response — giving a full agentic loop demo.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/overlay"
)

// --- Mock tool definitions ---

var toolDefs = []ai.ToolDef{
	{
		Type: "function",
		Function: ai.ToolFunction{
			Name:        "list_files",
			Description: "List files in the given directory",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"directory": map[string]any{
						"type":        "string",
						"description": "The directory to list",
					},
				},
				"required": []string{"directory"},
			},
		},
	},
	{
		Type: "function",
		Function: ai.ToolFunction{
			Name:        "read_file",
			Description: "Read the contents of a file",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The file path to read",
					},
				},
				"required": []string{"path"},
			},
		},
	},
	{
		Type: "function",
		Function: ai.ToolFunction{
			Name:        "calculate",
			Description: "Evaluate a simple math expression (e.g. '2+3*4')",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"expression": map[string]any{
						"type":        "string",
						"description": "The math expression to evaluate",
					},
				},
				"required": []string{"expression"},
			},
		},
	},
	{
		Type: "function",
		Function: ai.ToolFunction{
			Name:        "search_web",
			Description: "Search the web for information",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "The search query",
					},
				},
				"required": []string{"query"},
			},
		},
	},
}

// executeMockTool runs a tool and returns its result.
func executeMockTool(name, argsJSON string) string {
	var args map[string]any
	_ = json.Unmarshal([]byte(argsJSON), &args)

	switch name {
	case "list_files":
		dir, _ := args["directory"].(string)
		if dir == "" {
			dir = "."
		}
		return fmt.Sprintf("%s/\n  main.go\n  go.mod\n  go.sum\n  README.md\n  Makefile\n  cmd/\n  internal/\n  block/\n  ai/\n", dir)

	case "read_file":
		path, _ := args["path"].(string)
		return fmt.Sprintf("// %s\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n", path)

	case "calculate":
		expr, _ := args["expression"].(string)
		return fmt.Sprintf("Result: %s = (see eval result here)\n", expr)

	case "search_web":
		query, _ := args["query"].(string)
		return fmt.Sprintf("Search results for '%s':\n1. Relevant article about %s\n2. Documentation and examples\n3. Community discussion on %s\n", query, query, query)

	default:
		return fmt.Sprintf("Unknown tool: %s\n", name)
	}
}

// --- Input line ---

type inputLine struct {
	mu     sync.Mutex
	runes  []rune
	cursor int
}

func newInputLine() *inputLine { return &inputLine{} }

func (il *inputLine) handleKey(k *term.KeyEvent) bool {
	il.mu.Lock()
	defer il.mu.Unlock()

	switch {
	case k.Key == term.KeyEnter:
		return true
	case k.Key == term.KeyBackspace:
		if il.cursor > 0 {
			il.runes = append(il.runes[:il.cursor-1], il.runes[il.cursor:]...)
			il.cursor--
		}
	case k.Key == term.KeyLeft:
		if il.cursor > 0 {
			il.cursor--
		}
	case k.Key == term.KeyRight:
		if il.cursor < len(il.runes) {
			il.cursor++
		}
	case k.Key == term.KeyHome:
		il.cursor = 0
	case k.Key == term.KeyEnd:
		il.cursor = len(il.runes)
	case k.Rune != 0 && k.Modifiers == 0:
		il.runes = append(il.runes[:il.cursor], append([]rune{k.Rune}, il.runes[il.cursor:]...)...)
		il.cursor++
	case k.Rune == 'u' && k.Modifiers&term.ModCtrl != 0:
		il.runes = nil
		il.cursor = 0
	default:
		return false
	}
	return true
}

func (il *inputLine) text() string {
	il.mu.Lock()
	defer il.mu.Unlock()
	return string(il.runes)
}

func (il *inputLine) clear() {
	il.mu.Lock()
	defer il.mu.Unlock()
	il.runes = nil
	il.cursor = 0
}

func (il *inputLine) paint(buf *buffer.Buffer, x, y, w int) {
	il.mu.Lock()
	defer il.mu.Unlock()

	// Clear line
	for i := 0; i < w; i++ {
		buf.SetCell(x+i, y, buffer.Cell{Rune: ' ', Width: 1, Bg: buffer.RGB(40, 42, 54)})
	}

	// Prompt
	promptStyle := buffer.Style{Fg: buffer.RGB(139, 233, 253), Bg: buffer.RGB(40, 42, 54), Flags: buffer.Bold}
	buf.DrawText(x, y, "▶ ", promptStyle)

	textStart := x + 2
	availW := w - 2
	textStyle := buffer.Style{Fg: buffer.RGB(248, 248, 242), Bg: buffer.RGB(40, 42, 54)}

	visualX := textStart
	for _, r := range il.runes {
		rw := buffer.RuneWidth(r)
		if rw <= 0 {
			rw = 1
		}
		if visualX+rw > textStart+availW {
			break
		}
		buf.SetCell(visualX, y, buffer.Cell{Rune: r, Width: rw, Fg: textStyle.Fg, Bg: textStyle.Bg})
		if rw == 2 {
			buf.SetCell(visualX+1, y, buffer.Cell{Rune: 0, Width: 0, Fg: textStyle.Fg, Bg: textStyle.Bg})
		}
		visualX += rw
	}

	// Cursor
	cursorX := textStart
	for i := 0; i < il.cursor && i < len(il.runes); i++ {
		rw := buffer.RuneWidth(il.runes[i])
		if rw <= 0 {
			rw = 1
		}
		cursorX += rw
	}
	if cursorX >= textStart+availW {
		cursorX = textStart + availW - 1
	}
	c := buf.GetCell(cursorX, y)
	buf.SetCell(cursorX, y, buffer.Cell{Rune: c.Rune, Width: c.Width, Fg: buffer.RGB(40, 42, 54), Bg: buffer.RGB(248, 248, 242)})
}

// --- MultiText: simple multiline text component for overlays ---

type MultiText struct {
	component.BaseComponent
	lines []string
	style buffer.Style
}

func NewMultiText(lines []string) *MultiText {
	return &MultiText{
		lines: lines,
		style: buffer.Style{Fg: buffer.RGB(0xF8, 0xF8, 0xF2)},
	}
}

func (m *MultiText) Measure(cs component.Constraints) component.Size {
	w := 0
	for _, l := range m.lines {
		if lw := buffer.StringWidth(l); lw > w {
			w = lw
		}
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	h := len(m.lines)
	if h < 1 {
		h = 1
	}
	return component.Size{W: w, H: h}
}

func (m *MultiText) Paint(buf *buffer.Buffer) {
	bounds := m.Bounds()
	for i, line := range m.lines {
		y := bounds.Y + i
		if y >= bounds.Y+bounds.H {
			break
		}
		// Apply special styling for section headers (lines in uppercase)
		s := m.style
		if len(line) > 0 && line == strings.ToUpper(line) && strings.ContainsAny(line, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			s = buffer.Style{Fg: buffer.RGB(0xBD, 0x93, 0xF9), Flags: buffer.Bold} // purple bold
		}
		// Truncate to bounds width (rune-aware)
		r := []rune(line)
		if len(r) > bounds.W {
			r = r[:bounds.W]
		}
		buf.DrawText(bounds.X, y, string(r), s)
	}
}

// --- Main ---

func main() {
	cfg, err := ai.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	client := ai.NewClient(cfg)

	terminal, err := fluui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Terminal init error: %v\n", err)
		os.Exit(1)
	}
	defer terminal.Close()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		terminal.Quit()
	}()

	w, h := terminal.Size()
	chat := app.NewChatApp(w, h)
	chat.SetInputHeight(2)

	// Welcome message
	info := chat.AddAssistantText()
	info.AppendDelta(fmt.Sprintf(
		"✦ Fluui AI Chat — %s (model: %s)\n\n"+
			"Available tools: list_files, read_file, calculate, search_web\n"+
			"Try asking: \"list files in src/\" or \"calculate 2+3*4\"\n\n"+
			"Keys: Enter=send · ↑↓=scroll · Tab=toggle thinking · Ctrl+L=clear · Ctrl+C/Esc=quit\n",
		cfg.BaseURL, cfg.Model,
	))
	info.Complete()

	il := newInputLine()
	var conversation []ai.Message
	streaming := false
	var streamMu sync.Mutex

	systemPrompt := cfg.SystemPrompt + "\n\nYou have access to tools. When asked about files, calculations, or searching, use the appropriate tool."
	conversation = append(conversation, ai.Message{Role: ai.RoleSystem, Content: systemPrompt})

	// sendAndHandleAI runs the full AI streaming loop with tool calling.
	sendAndHandleAI := func(userText string) {
		streamMu.Lock()
		streaming = true
		streamMu.Unlock()

		go func() {
			defer func() {
				streamMu.Lock()
				streaming = false
				streamMu.Unlock()
				terminal.MarkDirty()
			}()

			for round := 0; round < 5; round++ { // max 5 tool-call rounds
				// Create blocks for this round
				var thinkingBlock *block.ThinkingBlock
				var textBlock *block.AssistantTextBlock

				// Track tool calls from this response
				var responseToolCalls []ai.ToolCall
				var finishReason string

				// Build callbacks
				cb := ai.StreamCallbacks{
					OnReasoning: func(text string) {
						if thinkingBlock == nil {
							thinkingBlock = chat.AddThinking()
						}
						thinkingBlock.AppendDelta(text)
						terminal.MarkDirty()
					},
					OnContent: func(text string) {
						if textBlock == nil {
							textBlock = chat.AddAssistantText()
						}
						textBlock.AppendDelta(text)
						terminal.MarkDirty()
					},
					OnToolCall: func(tc ai.ToolCall) {
						// Check if this is a new tool call
						found := false
						for i := range responseToolCalls {
							if responseToolCalls[i].Index == tc.Index {
								responseToolCalls[i] = tc
								found = true
								break
							}
						}
						if !found {
							responseToolCalls = append(responseToolCalls, tc)
						}
					},
					OnFinish: func(reason string) {
						finishReason = reason
					},
				}

				err := client.ChatStreamEx(conversation, toolDefs, cb)

				// Complete thinking block if created
				if thinkingBlock != nil {
					thinkingBlock.Complete()
				}

				if err != nil {
					if textBlock == nil {
						textBlock = chat.AddAssistantText()
					}
					textBlock.AppendDelta(fmt.Sprintf("\n❌ Error: %v", err))
					textBlock.Complete()
					terminal.MarkDirty()
					return
				}

				// Complete text block if created
				if textBlock != nil {
					textBlock.Complete()
				}
				terminal.MarkDirty()

				// If no tool calls, we're done
				if finishReason != "tool_calls" || len(responseToolCalls) == 0 {
					// Save assistant response to conversation
					if textBlock != nil {
						conversation = append(conversation, ai.Message{
							Role:    ai.RoleAssistant,
							Content: textBlock.Content(),
						})
					}
					return
				}

				// Build assistant message with tool_calls
				assistantContent := ""
				if textBlock != nil {
					assistantContent = textBlock.Content()
				}
				assistantMsg := ai.Message{
					Role:      ai.RoleAssistant,
					Content:   assistantContent,
					ToolCalls: responseToolCalls,
				}
				conversation = append(conversation, assistantMsg)

				// Execute each tool call and render results
				for _, tc := range responseToolCalls {
					// Tool call block
					tcBlock := chat.AddToolCall(tc.Function.Name, tc.Function.Arguments)
					tcBlock.Complete()
					terminal.MarkDirty()

					// Simulate processing time
					time.Sleep(200 * time.Millisecond)

					// Execute mock tool
					result := executeMockTool(tc.Function.Name, tc.Function.Arguments)

					// Tool result block
					trBlock := chat.AddToolResult()
					trBlock.AppendDelta(result)
					trBlock.Complete()
					terminal.MarkDirty()

					// Add tool result to conversation
					conversation = append(conversation, ai.Message{
						Role:       ai.RoleTool,
						Content:    result,
						ToolCallID: tc.ID,
					})
				}

				// Loop continues — LLM will get tool results and respond
			}
		}()
	}

	// --- Overlay manager ---
	mgr := overlay.NewOverlayManager()

	showHelpOverlay := func() {
		lines := []string{
			"Fluui AI Chat — Keyboard & Mouse Help",
			"",
			"INPUT",
			"  Type text + Enter       Send message",
			"  Ctrl+U                  Clear input line",
			"",
			"NAVIGATION",
			"  Up / Down               Scroll one line",
			"  PageUp / PageDown       Scroll one page",
			"  Home / End              Top / Bottom",
			"  Mouse Wheel             Scroll up/down",
			"",
			"BLOCKS",
			"  Tab                     Toggle thinking/result collapse",
			"  Click header            Toggle thinking/result block",
			"  Right-click block       View full content in popup",
			"",
			"OVERLAYS",
			"  ?                       Show this help",
			"  Esc / Ctrl+C            Quit (or close overlay)",
			"",
			"OTHER",
			"  Ctrl+L                  Clear all messages",
		}
		body := NewMultiText(lines)
		helpModal := overlay.NewModal("help", " Help ", body, []string{"Close"})
		mgr.Show(helpModal)
		terminal.MarkDirty()
	}

	showDetailPopup := func(title, content string) {
		lines := strings.Split(content, "\n")
		body := NewMultiText(lines)
		popup := overlay.NewPopup("detail", title, body)
		mgr.Show(popup)
		terminal.MarkDirty()
	}

	// --- Key handler ---
	terminal.OnKey(func(k *term.KeyEvent) {
		// If overlay is open, route key to it first
		if mgr.Top() != nil {
			mgr.HandleKey(k)
			// Enter on modal, Esc on any overlay → dismiss
			if k.Key == term.KeyEnter || k.Key == term.KeyEscape {
				mgr.HideAll()
			}
			terminal.MarkDirty()
			return
		}

		// '?' shows help overlay
		if k.Rune == '?' {
			showHelpOverlay()
			return
		}

		// Ctrl+L: clear
		if k.Rune == 'l' && k.Modifiers&term.ModCtrl != 0 {
			chat.Clear()
			conversation = nil
			il.clear()
			terminal.MarkDirty()
			return
		}

		streamMu.Lock()
		isStreaming := streaming
		streamMu.Unlock()

		if isStreaming {
			if k.Key == term.KeyEscape {
				terminal.Quit()
			}
			return
		}

		if k.Key == term.KeyEscape {
			terminal.Quit()
			return
		}

		// Scroll keys
		switch k.Key {
		case term.KeyUp:
			chat.ScrollUp()
			terminal.MarkDirty()
			return
		case term.KeyDown:
			chat.ScrollDown()
			terminal.MarkDirty()
			return
		case term.KeyPageUp:
			chat.ScrollView().ScrollUp(h)
			terminal.MarkDirty()
			return
		case term.KeyPageDown:
			chat.ScrollView().ScrollDown(h)
			terminal.MarkDirty()
			return
		case term.KeyHome:
			chat.ScrollView().ScrollTo(0)
			terminal.MarkDirty()
			return
		case term.KeyEnd:
			chat.ScrollToBottom()
			terminal.MarkDirty()
			return
		case term.KeyTab:
			for _, b := range chat.Container().Blocks() {
				if tb, ok := b.(*block.ThinkingBlock); ok {
					tb.Toggle()
				}
				if tr, ok := b.(*block.ToolResultBlock); ok {
					tr.Toggle()
				}
			}
			terminal.MarkDirty()
			return
		}

		// Input
		if il.handleKey(k) && k.Key == term.KeyEnter {
			text := il.text()
			if text == "" {
				return
			}
			il.clear()
			chat.AddUserMessage(text)
			conversation = append(conversation, ai.Message{Role: ai.RoleUser, Content: text})
			terminal.MarkDirty()
			sendAndHandleAI(text)
			return
		}

		terminal.MarkDirty()
	})

	// --- Mouse handler ---
	terminal.OnMouse(func(m *term.MouseEvent) {
		// If overlay is open, route mouse to it
		if mgr.Top() != nil {
			mgr.HandleMouse(m.X, m.Y)
			// Click outside modal dismisses it
			if m.Action == term.MouseDown {
				top := mgr.Top()
				bounds := top.Bounds()
				if m.X < bounds.X || m.X >= bounds.X+bounds.W ||
					m.Y < bounds.Y || m.Y >= bounds.Y+bounds.H {
					mgr.HideAll()
					terminal.MarkDirty()
				}
			}
			return
		}

		if m.Action == term.MouseWheel {
			switch m.Button {
			case term.MouseWheelUp:
				chat.ScrollUp()
			case term.MouseWheelDown:
				chat.ScrollDown()
			}
			terminal.MarkDirty()
			return
		}

		// Left-click: toggle blocks on header row
		if m.Action == term.MouseDown && m.Button == term.MouseLeft {
			for _, b := range chat.Container().Blocks() {
				bounds := b.Bounds()
				if m.Y >= bounds.Y && m.Y < bounds.Y+bounds.H {
					clickRow := m.Y - bounds.Y
					if clickRow <= 1 {
						switch blk := b.(type) {
						case *block.ThinkingBlock:
							blk.Toggle()
							terminal.MarkDirty()
						case *block.ToolResultBlock:
							blk.Toggle()
							terminal.MarkDirty()
						}
					}
					return
				}
			}
		}

		// Right-click: show detail popup for any block
		if m.Action == term.MouseDown && m.Button == term.MouseRight {
			for _, b := range chat.Container().Blocks() {
				bounds := b.Bounds()
				if m.Y >= bounds.Y && m.Y < bounds.Y+bounds.H {
					var title, content string
					switch blk := b.(type) {
					case *block.ThinkingBlock:
						title = "Thinking Detail"
						content = blk.Content()
					case *block.ToolResultBlock:
						title = "Tool Result Detail"
						content = blk.Output()
					case *block.ToolCallBlock:
						title = "Tool Call Detail"
						content = fmt.Sprintf("Tool: %s\nArgs: %s", blk.ToolName(), blk.RawArgs())
					case *block.AssistantTextBlock:
						title = "Assistant Message Detail"
						content = blk.Content()
					case *block.UserMessageBlock:
						title = "User Message Detail"
						content = blk.Content()
					}
					if content != "" {
						showDetailPopup(title, content)
					}
					return
				}
			}
		}
	})

	// --- Paint ---
	terminal.OnPaint(func(buf *buffer.Buffer) {
		w, h := terminal.Size()
		chat.SetSize(w, h)
		buf.Fill(buffer.Cell{Rune: ' ', Width: 1, Bg: buffer.RGB(40, 42, 54)})
		chat.Render(buf)

		// Separator + status
		sepY := h - 2
		inputY := h - 1
		sepFg := buffer.RGB(98, 114, 164)
		for x := 0; x < w; x++ {
			buf.SetCell(x, sepY, buffer.Cell{Rune: '─', Width: 1, Fg: sepFg, Bg: buffer.RGB(40, 42, 54)})
		}

		streamMu.Lock()
		isStreaming := streaming
		streamMu.Unlock()

		toolCount := 0
		thinkCount := 0
		for _, b := range chat.Container().Blocks() {
			switch b.Type() {
			case block.TypeToolCall:
				toolCount++
			case block.TypeThinking:
				thinkCount++
			}
		}

		status := fmt.Sprintf(" %d msgs · %d tools · %d thinking · %s ", len(conversation), toolCount, thinkCount, cfg.Model)
		if isStreaming {
			status = " ● streaming · " + status
		}
		statusStyle := buffer.Style{Fg: buffer.RGB(98, 114, 164), Bg: buffer.RGB(40, 42, 54)}
		buf.DrawText(w-len(status)-1, sepY, status, statusStyle)

		il.paint(buf, 0, inputY, w)

		// Render overlays on top of everything
		mgr.Measure(w, h)
		mgr.Paint(buf)
	})

	// Auto-scroll only while streaming — when idle, let user scroll freely
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			streamMu.Lock()
			isStreaming := streaming
			streamMu.Unlock()
			if isStreaming {
				chat.ScrollToBottom()
				terminal.MarkDirty()
			}
		}
	}()

	if err := terminal.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Run error: %v\n", err)
	}
}
