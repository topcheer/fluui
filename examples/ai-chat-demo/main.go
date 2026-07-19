// Package main implements a minimal AI chat demo showing Fluui's
// complete AI-native chat framework: ConversationView, ChatComposer,
// TokenUsageWidget, ToolCallView, and CitationsBlock.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	// --- Create the conversation ---
	conv := component.NewConversationView()

	// System message
	conv.AddSystemMessage("You are chatting with a Fluui-powered AI assistant.")

	// User greeting
	conv.AddUserMessage("Hi! Can you list files in /tmp?")

	// Simulated tool call
	tc := component.NewToolCallView("list_files", `{"dir":"/tmp"}`)
	conv.AddToolCall(tc)

	// Simulated assistant response (streaming)
	conv.AddAssistantMessage(
		"I found 3 files in /tmp:\n\n1. `fluui-test.log`\n2. `config.yaml`\n3. `cache.db`",
		"fluui-demo",
	)

	// Citations
	conv.AddCitations(component.NewCitationsBlock([]component.Citation{
		{Index: 1, Title: "Fluui Docs", URL: "https://github.com/topcheer/fluui", Snippet: "The strongest TUI library for Go"},
		{Index: 2, Title: "Go Standard Library", URL: "https://pkg.go.dev", Snippet: "Official Go documentation"},
	}))

	// --- Create the composer ---
	composer := component.NewChatComposer()
	composer.SetPlaceholder("Ask anything…")
	composer.SetHint("Enter to send · Shift+Enter for newline · Ctrl+C to quit")
	composer.SetTokenCount(842, 156)

	// --- Create the token widget ---
	tokens := component.NewTokenUsageWidget("fluui-demo")
	tokens.AddTokens(842, 156)
	tokens.SetContextUsage(5000, 128000)

	// --- Render a single frame to show everything works ---
	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║     Fluui AI Chat Framework — Demo Snapshot     ║")
	fmt.Println("╚══════════════════════════════════════════════════╝")
	fmt.Println()

	// Measure conversation
	width := 54
	convHeight := 18
	composerHeight := 5

	// Paint conversation
	conv.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: convHeight})
	convBuf := buffer.NewBuffer(width, convHeight)
	conv.Paint(convBuf)
	fmt.Println("=== ConversationView ===")
	fmt.Println(renderBuffer(convBuf))

	// Complete the tool call
	tc.Complete()

	// Paint composer
	composer.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: composerHeight})
	composerBuf := buffer.NewBuffer(width, composerHeight)
	composer.Paint(composerBuf)
	fmt.Println("=== ChatComposer ===")
	fmt.Println(renderBuffer(composerBuf))

	// Paint token widget
	tokens.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	tokenBuf := buffer.NewBuffer(width, 1)
	tokens.Paint(tokenBuf)
	fmt.Println("=== TokenUsageWidget ===")
	fmt.Println(renderBuffer(tokenBuf))

	// --- Simulate streaming response ---
	fmt.Println()
	fmt.Println("=== Simulating streaming response ===")
	composer.SetDisabled(true)
	conv.AddMessage(component.ConversationMessage{
		Role:      component.RoleAssistant,
		Content:   "Let me analyze those files",
		ModelName: "fluui-demo",
		Streaming:  true,
	})
	conv.SetStreaming(true)

	// Simulate streaming tokens
	go func() {
		tokens := []string{"Let", " me", " analyze", " those", " files", " for", " you", "."}
		for _, tk := range tokens {
			time.Sleep(50 * time.Millisecond)
			_ = tk // In a real app, you'd append to the message content
		}
	}()

	time.Sleep(200 * time.Millisecond)
	composer.SetDisabled(false)

	// --- Handle key input (Enter to send) ---
	composer.SetText("What about the config file?")
	composer.SetOnSubmit(func(text string) {
		fmt.Printf("\n[Submitted]: %q\n", text)
		conv.AddUserMessage(text)
	})
	composer.HandleKey(&term.KeyEvent{Key: term.KeyEnter})

	fmt.Println("\n=== After submit ===")
	fmt.Printf("Messages: %d\n", conv.MessageCount())
	fmt.Printf("Composer text after submit: %q\n", composer.Text())

	// Expand tool call
	tc.SetExpanded(true)
	tc.SetResult("config.yaml (2.1KB)\n  mode: production\n  port: 8080")
	tc.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 10})
	tcBuf := buffer.NewBuffer(width, 10)
	tc.Paint(tcBuf)
	fmt.Println("\n=== ToolCallView (expanded with result) ===")
	fmt.Println(renderBuffer(tcBuf))

	fmt.Println("\n✓ All Fluui AI-native components work end-to-end!")
}

// renderBuffer converts a buffer to a printable string with a border.
func renderBuffer(buf *buffer.Buffer) string {
	var sb strings.Builder
	w, h := buf.Width, buf.Height
	sb.WriteString("┌")
	sb.WriteString(strings.Repeat("─", w))
	sb.WriteString("┐\n")
	for y := 0; y < h; y++ {
		sb.WriteString("│")
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Width > 0 {
				sb.WriteRune(cell.Rune)
			} else {
				sb.WriteRune(' ')
			}
		}
		sb.WriteString("│\n")
	}
	sb.WriteString("└")
	sb.WriteString(strings.Repeat("─", w))
	sb.WriteString("┘")
	return sb.String()
}
