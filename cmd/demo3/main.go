// Package main implements an interactive Phase 3 content block demo.
//
// Simulates an AI chat conversation with streaming content blocks:
// ThinkingBlock, ToolCallBlock, ToolResultBlock, AssistantTextBlock.
package main

import (
	"fmt"
	"time"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	app, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	// Create block container
	container := block.NewBlockContainer()

	// Simulated conversation blocks
	var allBlocks []block.Block

	// User message
	userMsg := block.NewUserMessageBlock("user-1", "Can you search for Go TUI libraries?")
	allBlocks = append(allBlocks, userMsg)

	// Thinking block
	thinking := block.NewThinkingBlock("think-1")
	allBlocks = append(allBlocks, thinking)

	// Tool call
	toolCall := block.NewToolCallBlock("tc-1", "search", `query: "Go TUI library"`)
	allBlocks = append(allBlocks, toolCall)

	// Tool result
	toolResult := block.NewToolResultBlock("tr-1")
	allBlocks = append(allBlocks, toolResult)

	// Final assistant response
	assistant := block.NewAssistantTextBlock("asst-1")
	allBlocks = append(allBlocks, assistant)

	// Add all to container
	for _, b := range allBlocks {
		container.AddBlock(b)
	}

	// Simulate streaming in background
	streamDone := make(chan struct{})
	go func() {
		// Phase 1: Thinking (streaming)
		time.Sleep(300 * time.Millisecond)
		thinkingTexts := []string{
			"The user wants Go TUI libraries. ",
			"Let me search for popular options. ",
			"Bubble Tea, tview, and termui are the most common ones.",
		}
		for _, t := range thinkingTexts {
			thinking.AppendDelta(t)
			app.MarkDirty()
			time.Sleep(200 * time.Millisecond)
		}
		thinking.Complete()

		// Phase 2: Tool call
		time.Sleep(200 * time.Millisecond)
		toolCall.Complete()

		// Phase 3: Tool result
		time.Sleep(300 * time.Millisecond)
		toolResult.AppendDelta("Found 5 results:\n1. Bubble Tea (charmbracelet)\n2. tview (rivo)\n3. termui (gizak)\n4. gofpb (gdamore)\n5. Fluui (topcheer)")
		toolResult.Complete()

		// Phase 4: Assistant response
		time.Sleep(200 * time.Millisecond)
		responseParts := []string{
			"Based on my search, ",
			"here are the top Go TUI libraries:\n\n",
			"1. Bubble Tea - Elm-style, very popular\n",
			"2. tview - Rich interactive widgets\n",
			"3. termui - Dashboard-style layouts\n\n",
			"Each has different strengths for terminal UI development.",
		}
		for _, part := range responseParts {
			assistant.AppendDelta(part)
			app.MarkDirty()
			time.Sleep(150 * time.Millisecond)
		}
		assistant.Complete()

		close(streamDone)
	}()

	// Keyboard handler
	app.OnKey(func(k *term.KeyEvent) {
		switch {
		case k.Key == term.KeyEscape || (k.Rune == 'q' && k.Modifiers == 0):
			app.Quit()
		case k.Key == term.KeyTab:
			// Toggle thinking block
			thinking.Toggle()
			app.MarkDirty()
		}
	})

	// Paint handler
	app.OnPaint(func(buf *buffer.Buffer) {
		w, h := buf.Width, buf.Height

		// Fill background (Dracula bg)
		bgCell := buffer.Cell{Rune: ' ', Width: 1, Bg: buffer.RGB(40, 42, 54)}
		buf.Fill(bgCell)

		// Title bar
		titleStyle := buffer.Style{}.
			WithFg(buffer.RGB(255, 121, 198)).
			WithFlags(buffer.Bold)
		buf.DrawText(1, 0, " Fluui Phase 3 — AI Chat Demo", titleStyle)

		// Layout blocks in the available space
		container.SetSpacing(1)
		container.SetBounds(component.Rect{X: 1, Y: 1, W: w - 2, H: h - 3})
		container.Paint(buf)

		// Status bar
		statusY := h - 1
		for x := 0; x < w; x++ {
			buf.SetCell(x, statusY, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    buffer.RGB(248, 248, 242),
			})
		}
		statusStyle := buffer.Style{}.
			WithFg(buffer.RGB(40, 42, 54)).
			WithBg(buffer.RGB(248, 248, 242))

		streaming := 0
		for _, b := range container.Blocks() {
			if b.State() == block.BlockStreaming {
				streaming++
			}
		}
		status := fmt.Sprintf("  %d blocks | %d streaming | Tab: toggle thinking | q/Esc: quit",
			container.Len(), streaming)
		buf.DrawTextClamped(1, statusY, status, statusStyle)
	})

	if err := app.Run(); err != nil {
		panic(err)
	}

	// Wait for stream goroutine to finish
	<-streamDone
}
