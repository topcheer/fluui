// Package main implements demo14 — Phase 25 Full Chat Showcase.
//
// A print-based demo that simulates a complete AI chat session with:
// - Streaming markdown rendering (headings, lists, code blocks, bold/italic)
// - Command palette (Ctrl+P) with searchable commands
// - Spinner animation during "AI thinking"
// - Theme switching (Dracula, GitHub, Monokai)
// - Undo/Redo on input line
// - User messages and assistant responses
//
// Usage: go run ./cmd/demo14/
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func main() {
	width := 70

	printHeader(width)

	// ── Simulate chat session ──
	chat := app.NewChatApp(width, 24)

	// Set up command palette
	palette := component.NewCommandPalette()
	chat.SetCommandPalette(palette)
	registerCommands(chat, palette)

	// Set up spinner
	spinner := component.NewSpinner("Thinking...")
	chat.SetSpinner(spinner)

	// Welcome message
	welcome := chat.AddAssistantText()
	welcome.AppendDelta("# Fluui Chat Demo\n\n")
	welcome.AppendDelta("This demo showcases all **P19-P20** features:\n")
	welcome.AppendDelta("- **Streaming markdown** with syntax highlighting\n")
	welcome.AppendDelta("- **Command Palette** (Ctrl+P) for quick actions\n")
	welcome.AppendDelta("- **Spinner** during AI responses\n")
	welcome.AppendDelta("- **Theme switching** — try Ctrl+] and Ctrl+\\\n")
	welcome.AppendDelta("- **Undo/Redo** — Ctrl+Z / Ctrl+Y\n\n")
	welcome.AppendDelta("```go\nfmt.Println(\"Hello, Fluui!\")\n```\n")
	welcome.Complete()

	// Simulate user message
	fmt.Println("\n┌─ User Message ──────────────────────────────────────────────────┐")
	fmt.Println("│ Explain the benefits of using Fluui for TUI development.        │")
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")

	// Simulate spinner
	demoSpinner(spinner, width)

	// Simulate streaming AI response
	response := chat.AddAssistantText()
	streamResponse(response)

	// Render the chat
	fmt.Println("\n┌─ Chat Render ───────────────────────────────────────────────────┐")
	buf := buffer.NewBuffer(width, 24)
	chat.Render(buf)
	printBuffer(buf, width)
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")

	// ── Command Palette ──
	demoCommandPalette(palette, width)

	// ── Theme Switching ──
	demoThemeSwitching(chat, width)

	// ── Undo/Redo ──
	demoUndoRedo(chat, width)

	// ── Summary ──
	printSummary(width)
}

// ─── Streaming simulation ──────────────────────────────────

func streamResponse(b *block.AssistantTextBlock) {
	chunks := []string{
		"## Benefits of Fluui\n\n",
		"Fluui is a **production-grade** AI-native TUI library for Go:\n\n",
		"### Key Features\n\n",
		"1. **Zero dependencies** — pure Go, no C bindings\n",
		"2. **Streaming markdown** — real-time AI response rendering\n",
		"3. **33+ components** — Gauge, Table, Tree, Sparkline, more\n",
		"4. **Thread-safe** — built for concurrent AI streaming\n",
		"5. **Themeable** — Dracula, GitHub, Monokai built-in\n\n",
		"### Performance\n\n",
		"- Render pipeline: **44% faster** (P24-A)\n",
		"- Link detection: **90% fewer allocs** (P24-B)\n",
		"- Markdown cache: **99.85% fewer allocs** (P24-D)\n\n",
		"```go\n// Create a chat app in 5 lines\napp := app.NewChatApp(80, 24)\napp.SetAIClient(client)\napp.AddAssistantText()\nbase.Run(app)\n```\n",
	}

	for _, chunk := range chunks {
		b.AppendDelta(chunk)
	}
	b.Complete()
}

// ─── Command Palette demo ─────────────────────────────────

func demoCommandPalette(palette *component.CommandPalette, width int) {
	fmt.Println("\n┌─ Command Palette (Ctrl+P) ──────────────────────────────────────┐")
	fmt.Println()

	// Show all commands
	palette.SetQuery("")
	commands := palette.Commands()
	fmt.Printf("  All commands (%d):\n", len(commands))
	for _, c := range commands {
		fmt.Printf("    • %-24s [%s]\n", c.Label, c.Category)
	}

	// Simulate search
	fmt.Println()
	palette.SetQuery("theme")
	filtered := palette.FilteredCommands()
	fmt.Printf("  Search \"theme\" → %d results:\n", len(filtered))
	for _, c := range filtered {
		fmt.Printf("    • %-24s [%s]\n", c.Label, c.Category)
	}

	fmt.Println("\n└──────────────────────────────────────────────────────────────────┘")
}

// ─── Theme switching demo ─────────────────────────────────

func demoThemeSwitching(chat *app.ChatApp, width int) {
	fmt.Println("\n┌─ Theme Switching (Ctrl+] / Ctrl+\\) ─────────────────────────────┐")
	fmt.Println()

	themes := chat.ThemeList()
	fmt.Printf("  Available themes (%d):\n", len(themes))
	for i, name := range themes {
		marker := ""
		if i == chat.ThemeIndex() {
			marker = " ← current"
		}
		fmt.Printf("    %d. %s%s\n", i, name, marker)
	}

	// Cycle through themes
	fmt.Println()
	for i := 0; i < len(themes); i++ {
		chat.SetThemeByIndex(i)
		fmt.Printf("  → Switched to: %s\n", chat.ThemeName())
	}

	// Reset to first
	chat.SetThemeByIndex(0)
	fmt.Printf("  → Reset to: %s\n", chat.ThemeName())

	fmt.Println("\n└──────────────────────────────────────────────────────────────────┘")
}

// ─── Undo/Redo demo ───────────────────────────────────────

func demoUndoRedo(chat *app.ChatApp, width int) {
	fmt.Println("\n┌─ Undo/Redo (Ctrl+Z / Ctrl+Y) ───────────────────────────────────┐")
	fmt.Println()

	// Need to init InputLine
	chat.OnSubmit(func(string) {})
	il := chat.InputLine()
	if il == nil {
		fmt.Println("  (InputLine not available)")
		fmt.Println("\n└──────────────────────────────────────────────────────────────────┘")
		return
	}

	// Simulate typing
	il.SetText("Hello")
	fmt.Printf("  Typed:  %q\n", il.Text())

	il.InsertText(" World")
	fmt.Printf("  Typed:  %q\n", il.Text())

	il.InsertText("!")
	fmt.Printf("  Typed:  %q\n", il.Text())

	fmt.Println("\n  (Undo/Redo via Ctrl+Z / Ctrl+Y in interactive mode)")

	fmt.Println("\n└──────────────────────────────────────────────────────────────────┘")
}

// ─── Spinner demo ─────────────────────────────────────────

func demoSpinner(spinner *component.Spinner, width int) {
	fmt.Println("\n┌─ AI Thinking ───────────────────────────────────────────────────┐")
	spinner.Start()

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	for i := 0; i < 6; i++ {
		fmt.Printf("\r  %s Thinking...", frames[i%len(frames)])
		time.Sleep(150 * time.Millisecond)
	}

	spinner.Stop()
	fmt.Printf("\r  ✓ Response ready!                    \n")
	fmt.Println("└──────────────────────────────────────────────────────────────────┘")
}

// ─── Commands ─────────────────────────────────────────────

func registerCommands(chat *app.ChatApp, palette *component.CommandPalette) {
	type cmd struct {
		id, label, category string
	}
	commands := []cmd{
		{"clear", "Clear Chat", "edit"},
		{"export", "Export Conversation", "file"},
		{"theme-dark", "Switch to Dark Theme", "appearance"},
		{"theme-light", "Switch to Light Theme", "appearance"},
		{"undo", "Undo Typing", "edit"},
		{"redo", "Redo Typing", "edit"},
		{"settings", "Open Settings", "app"},
		{"help", "Show Help", "app"},
	}
	for _, c := range commands {
		chat.AddCommand(c.id, c.label, c.category, func() {})
	}
}

// ─── Print helpers ────────────────────────────────────────

func printHeader(width int) {
	bar := strings.Repeat("═", width)
	fmt.Println("╔" + bar + "╗")
	fmt.Println("║" + center("Fluui Demo 14 — Full Chat Showcase", width) + "║")
	fmt.Println("╠" + bar + "╣")
	fmt.Println("║" + center("Streaming Markdown · Command Palette · Spinner", width) + "║")
	fmt.Println("║" + center("Theme Switching · Undo/Redo · 2800+ tests", width) + "║")
	fmt.Println("╚" + bar + "╝")
}

func printSummary(width int) {
	bar := strings.Repeat("─", width)
	fmt.Println()
	fmt.Println("┌" + bar + "┐")
	fmt.Printf("│ %s%s│\n", "Phase 25 Demo14 — Showcases all P19-P24 features in one demo", pad(width-62))
	fmt.Printf("│ %s%s│\n", "Features: Streaming MD · CommandPalette · Spinner · Themes · Undo", pad(width-66))
	fmt.Printf("│ %s%s│\n", "Project: 2800 tests | 33 components | 21 packages | 0 deps", pad(width-59))
	fmt.Println("└" + bar + "┘")
}

func printBuffer(buf *buffer.Buffer, width int) {
	h := buf.Height
	if h > 16 {
		h = 16
	}
	for y := 0; y < h; y++ {
		var line strings.Builder
		for x := 0; x < width; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != 0 {
				line.WriteRune(cell.Rune)
			} else {
				line.WriteRune(' ')
			}
		}
		fmt.Printf("│%s│\n", line.String())
	}
}

func center(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	pad := (width - len(s)) / 2
	return strings.Repeat(" ", pad) + s + strings.Repeat(" ", width-len(s)-pad)
}

func pad(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}
