// Package main implements demo23 — AI Chat Application.
//
// An interactive TUI that demonstrates multi-component integration:
// - ConversationView with streaming message simulation
// - ChatComposer with Enter-to-send and slash commands
// - ThinkingIndicator animation
// - TokenUsageWidget for cost tracking
// - SegmentedControl for mode switching (Chat/Code/Settings)
// - Breadcrumb for navigation context
// - StatusBar for hints and status
// - PieChart for token distribution
//
// This validates that all AI-native components work together in a
// realistic application scenario.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	app, err := fluui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	app.SetTitle("Fluui AI Chat — Integration Demo")

	// ── Components ──────────────────────────────────────────

	// Conversation view (main content)
	conv := component.NewConversationView()
	conv.AddAssistantMessage(
		"Welcome to the Fluui AI Chat demo! This app demonstrates multi-component integration: ConversationView, ChatComposer, ThinkingIndicator, TokenUsageWidget, SegmentedControl, Breadcrumb, and StatusBar all rendering together in real-time.\n\nType a message and press Enter to chat. Use /help for commands.",
		"fluui-demo",
	)

	// Chat composer (input)
	composer := component.NewChatComposer()
	composer.SetPlaceholder("Type a message... (/help for commands)")

	// Thinking indicator
	thinking := component.NewThinkingIndicator("Thinking")
	thinking.SetLabel("AI is generating response")

	// Token usage
	tokens := component.NewTokenUsageWidget("gpt-4")
	tokens.AddTokens(150, 80)

	// Mode switcher
	modes := component.NewSegmentedControl([]string{"Chat", "Code", "Settings"})

	// Breadcrumb navigation
	bc := component.NewBreadcrumb([]string{"AI Assistant", "Demo Session"})
	bc.SetActive(1)

	// Status bar
	statusBar := component.NewStatusBar()
	statusBar.AddLeft("status", "● Ready")
	statusBar.AddRight("hint", "Enter: send  /help: commands  Tab: switch mode  q: quit")
	statusBar.AddCenter("msgs", "Messages: 1")

	// Token pie chart
	pie := component.NewPieChart([]component.PieSlice{
		{Label: "Input", Value: 150},
		{Label: "Output", Value: 80},
	})
	pie.SetDonut(true)

	// ── State ───────────────────────────────────────────────

	msgCount := 1
	focusMode := 0 // 0=composer, 1=modes
	inputTokens := 150
	outputTokens := 80
	streamingActive := false

	// ── Composer submit handler ─────────────────────────────

	composer.SetOnSubmit(func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}

		// Handle slash commands
		if strings.HasPrefix(text, "/") {
			handleCommand(text, conv, app)
			statusBar.SetItemText("msgs", fmt.Sprintf("Messages: %d", conv.MessageCount()))
			app.MarkDirty()
			return
		}

		// Add user message
		conv.AddUserMessage(text)
		msgCount++
		inputTokens += estimateTokens(text)

		// Start streaming simulation
		streamingActive = true
		thinking.Start(300 * time.Millisecond)
		statusBar.SetItemText("status", "● Thinking...")
		app.MarkDirty()

		// Simulate AI response after delay
		go func() {
			time.Sleep(800 + time.Duration(rand.Intn(800))*time.Millisecond)

			response := generateResponse(text)
			conv.AddAssistantMessage(response, "gpt-4")
			msgCount++
			outputTokens += estimateTokens(response)

			thinking.Stop()
			streamingActive = false
			tokens.AddTokens(estimateTokens(text), estimateTokens(response))

			// Update pie chart
			pie.SetSlices([]component.PieSlice{
				{Label: "Input", Value: float64(inputTokens)},
				{Label: "Output", Value: float64(outputTokens)},
			})

			statusBar.SetItemText("status", "● Ready")
			statusBar.SetItemText("msgs", fmt.Sprintf("Messages: %d", conv.MessageCount()))
			app.MarkDirty()
		}()

		app.MarkDirty()
	})

	// ── Key handling ────────────────────────────────────────

	app.OnKey(func(k *term.KeyEvent) {
		// Global: quit
		if k.Rune == 'q' && focusMode == 1 {
			app.Quit()
			return
		}
		if k.Key == term.KeyEscape {
			app.Quit()
			return
		}

		// Tab: toggle focus between composer and mode switcher
		if k.Key == term.KeyTab {
			focusMode = 1 - focusMode
			app.MarkDirty()
			return
		}

		// Route to focused component
		if focusMode == 0 {
			if composer.HandleKey(k) {
				app.MarkDirty()
			}
		} else {
			switch k.Key {
			case term.KeyLeft:
				modes.SelectPrev()
				app.MarkDirty()
			case term.KeyRight:
				modes.SelectNext()
				app.MarkDirty()
			}
			if k.Rune == 'q' {
				app.Quit()
			}
		}
	})

	// ── Layout and Paint ────────────────────────────────────

	app.OnResize(func(w, h int) {
		app.MarkDirty()
	})

	app.OnPaint(func(buf *buffer.Buffer) {
		w, h := app.Size()

		// Layout:
		// Line 0: Breadcrumb + Mode switcher
		// Line 1: Conversation (flexible height)
		// Line h-6: Thinking indicator (if active)
		// Line h-5: Token usage + Pie chart
		// Line h-3: Chat composer (2 lines)
		// Line h-1: Status bar

		// Breadcrumb (top-left) + Mode switcher (top-right)
		bc.SetBounds(component.Rect{X: 0, Y: 0, W: w / 2, H: 1})
		bc.Paint(buf)

		modes.SetBounds(component.Rect{X: w / 2, Y: 0, W: w / 2, H: 1})
		modes.Paint(buf)

		// Separator
		buf.DrawText(0, 1, strings.Repeat("─", w), buffer.Style{Fg: buffer.NamedColor(8)})

		// Conversation view
		convTop := 2
		convBot := h - 7
		if streamingActive {
			convBot = h - 8
		}
		convH := convBot - convTop
		if convH < 3 {
			convH = 3
		}
		conv.SetBounds(component.Rect{X: 1, Y: convTop, W: w - 2, H: convH})
		conv.Paint(buf)

		// Thinking indicator (if active)
		if streamingActive {
			thinking.SetBounds(component.Rect{X: 1, Y: h - 8, W: 30, H: 1})
			thinking.Paint(buf)
		}

		// Token usage + pie chart side by side
		pieW := 24
		tokens.SetBounds(component.Rect{X: 1, Y: h - 5, W: w - pieW - 4, H: 1})
		tokens.Paint(buf)

		pie.SetBounds(component.Rect{X: w - pieW - 2, Y: h - 6, W: pieW, H: 3})
		pie.Paint(buf)

		// Separator before composer
		buf.DrawText(0, h-3, strings.Repeat("─", w), buffer.Style{Fg: buffer.NamedColor(8)})

		// Chat composer
		composer.SetBounds(component.Rect{X: 1, Y: h - 2, W: w - 2, H: 1})
		composer.Paint(buf)

		// Status bar
		statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})
		statusBar.Paint(buf)
	})

	// ── Run ─────────────────────────────────────────────────

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// handleCommand processes slash commands.
func handleCommand(cmd string, conv *component.ConversationView, app *fluui.App) {
	switch cmd {
	case "/help":
		conv.AddAssistantMessage(
			"Available commands:\n"+
				"  /help    — Show this help message\n"+
				"  /clear   — Clear conversation\n"+
				"  /stats   — Show token statistics\n"+
				"  /about   — About this demo\n"+
				"\nType any message to chat with the AI.",
			"system",
		)
	case "/clear":
		conv.Clear()
		conv.AddAssistantMessage("Conversation cleared.", "system")
	case "/about":
		conv.AddAssistantMessage(
			"Fluui AI Chat Demo — Integration Showcase\n\n"+
				"This demo validates multi-component integration:\n"+
				"• ConversationView — scrollable chat history\n"+
				"• ChatComposer — input with slash commands\n"+
				"• ThinkingIndicator — animated loading state\n"+
				"• TokenUsageWidget — cost tracking\n"+
				"• SegmentedControl — mode switching\n"+
				"• Breadcrumb — navigation context\n"+
				"• PieChart — token distribution\n"+
				"• StatusBar — hints and status\n\n"+
				"All components render with zero heap allocations on the hot path.",
			"system",
		)
	case "/stats":
		conv.AddAssistantMessage(
			"Statistics:\n"+
				"  Messages: "+fmt.Sprintf("%d", conv.MessageCount())+"\n"+
				"  Components: 8 active\n"+
				"  Zero-alloc Paint: 23/26 components",
			"system",
		)
	default:
		conv.AddAssistantMessage("Unknown command: "+cmd+". Type /help for available commands.", "system")
	}
}

// generateResponse simulates an AI response.
func generateResponse(input string) string {
	responses := []string{
		"That's a great question! In the context of TUI development, this relates to how terminal rendering pipelines handle buffer diffing efficiently. The key insight is that double-buffering with cell-level comparison eliminates full-screen redraws, which is critical for smooth 60fps animations.\n\nWould you like me to elaborate on the rendering architecture?",
		"I understand what you're asking about. Let me break it down:\n\n1. The terminal sends input events (keyboard, mouse) via escape sequences\n2. The event loop dispatches these to the focused component\n3. The component updates its state and marks itself dirty\n4. The renderer diffs the back buffer against the front buffer\n5. Only changed cells are written to the terminal\n\nThis pipeline ensures minimal I/O and smooth rendering.",
		"Interesting! Here's my analysis:\n\nThe zero-allocation rendering approach is what sets Fluui apart from other Go TUI libraries. By using stack buffers and strconv.Append* functions instead of fmt.Sprintf, we can render an entire 10-message conversation with literally zero heap allocations.\n\nThis matters for long-running TUI apps that must avoid GC pauses.",
		"Based on your input \"" + truncateForDisplay(input, 40) + "\", here are my thoughts:\n\nThe component architecture follows a strict Paint pattern: each component receives a buffer and writes directly to it. This avoids intermediate string allocations entirely. The trade-off is slightly more complex code, but the performance gains are substantial — typically 10-100x fewer allocations per frame.",
	}

	idx := rand.Intn(len(responses))
	return responses[idx]
}

// estimateTokens roughly estimates token count for a string.
func estimateTokens(s string) int {
	// Rough heuristic: ~4 chars per token
	return len(s) / 4
}

// truncateForDisplay truncates a string for inline display.
func truncateForDisplay(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
