// Package main implements an interactive Phase 4 demo.
//
// Showcases the full ChatApp API with streaming content blocks,
// scroll support, and modal/popup overlays.
//
// Controls:
//
//	Up/Down or mouse wheel — scroll
//	Tab                    — toggle thinking block
//	m                      — open/close modal dialog
//	p                      — open/close popup viewer
//	q or Esc               — quit
package main

import (
	"fmt"
	"time"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/overlay"
)

func main() {
	a, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer a.Close()

	w, h := a.Size()

	// Create ChatApp
	chat := app.NewChatApp(w, h)
	chat.SetInputHeight(1)

	// Create overlays
	modalBody := component.NewText("This is a modal dialog.\n\nPress Esc or Enter to close.")
	modal := overlay.NewModal("demo-modal", "About", modalBody, []string{"OK", "Cancel"})
	modal.Measure(component.Bounded(w, h))

	popupContent := component.NewText(
		"package main\n\n" +
			"import \"fmt\"\n\n" +
			"func main() {\n" +
			"    fmt.Println(\"Hello, World!\")\n" +
			"}\n",
	)
	popup := overlay.NewPopup("code-popup", "main.go", popupContent)
	popup.Measure(component.Bounded(w, h))

	// Start with a user message
	chat.AddUserMessage("Show me an example Go program.")

	// Simulate AI streaming in background
	streamDone := make(chan struct{})
	go func() {
		// Thinking phase
		time.Sleep(400 * time.Millisecond)
		chat.StreamDelta(block.StreamDelta{
			Type:    "thinking",
			Content: "The user wants a simple Go program.",
		})
		a.MarkDirty()
		time.Sleep(300 * time.Millisecond)
		chat.StreamDelta(block.StreamDelta{
			Type:    "thinking",
			Content: " I'll write a Hello World example.",
		})
		a.MarkDirty()
		time.Sleep(300 * time.Millisecond)

		// Tool call
		chat.StreamDelta(block.StreamDelta{
			Type:     "tool_call",
			ToolName: "write_file",
			ToolArgs: `{file: "main.go"}`,
		})
		a.MarkDirty()
		time.Sleep(400 * time.Millisecond)

		// Tool result
		chat.StreamDelta(block.StreamDelta{
			Type:    "tool_result",
			Content: "File written successfully (45 bytes)",
		})
		a.MarkDirty()
		time.Sleep(300 * time.Millisecond)

		// Assistant response
		responseParts := []string{
			"Here's a simple Go program ",
			"that prints \"Hello, World!\":\n\n",
			"The program imports `fmt` ",
			"and calls `fmt.Println` ",
			"in the `main` function.\n\n",
			"Press 'p' to view the full source code.",
		}
		for _, part := range responseParts {
			chat.StreamDelta(block.StreamDelta{
				Type:    "text",
				Content: part,
			})
			a.MarkDirty()
			time.Sleep(120 * time.Millisecond)
		}

		close(streamDone)
	}()

	// Keyboard handler
	a.OnKey(func(k *term.KeyEvent) {
		// If modal or popup is visible, let them handle keys
		if modal.Visible() {
			modal.HandleKey(k)
			a.MarkDirty()
			return
		}
		if popup.Visible() {
			popup.HandleKey(k)
			a.MarkDirty()
			return
		}

		switch {
		case k.Key == term.KeyEscape || (k.Rune == 'q' && k.Modifiers == 0):
			a.Quit()

		case k.Key == term.KeyUp:
			chat.ScrollUp()
			a.MarkDirty()

		case k.Key == term.KeyDown:
			chat.ScrollDown()
			a.MarkDirty()

		case k.Key == term.KeyHome:
			chat.ScrollView().ScrollTo(0)
			a.MarkDirty()

		case k.Key == term.KeyEnd:
			chat.ScrollToBottom()
			a.MarkDirty()

		case k.Key == term.KeyPageUp:
			chat.ScrollView().ScrollUp(h)
			a.MarkDirty()

		case k.Key == term.KeyPageDown:
			chat.ScrollView().ScrollDown(h)
			a.MarkDirty()

		case k.Key == term.KeyTab:
			// Toggle last thinking block
			for _, b := range chat.Container().Blocks() {
				if tb, ok := b.(*block.ThinkingBlock); ok {
					tb.Toggle()
				}
			}
			a.MarkDirty()

		case k.Rune == 'm' && k.Modifiers == 0:
			// Toggle modal
			if !modal.Visible() {
				modal.SetVisible(true)
				modal.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
			}
			a.MarkDirty()

		case k.Rune == 'p' && k.Modifiers == 0:
			// Toggle popup
			if !popup.Visible() {
				popup.SetVisible(true)
				popup.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
			}
			a.MarkDirty()
		}
	})

	// Mouse handler
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
		}
	})

	// Paint handler
	a.OnPaint(func(buf *buffer.Buffer) {
		bw, bh := buf.Width, buf.Height

		// Update chat size
		chat.SetSize(bw, bh)

		// Fill background
		bgCell := buffer.Cell{Rune: ' ', Width: 1, Bg: buffer.RGB(40, 42, 54)}
		buf.Fill(bgCell)

		// Render chat content
		chat.Render(buf)

		// Title bar
		titleStyle := buffer.Style{}.
			WithFg(buffer.RGB(255, 121, 198)).
			WithFlags(buffer.Bold)
		buf.DrawText(1, 0, " Fluui Phase 4 — ChatApp + Overlay Demo", titleStyle)

		// Status bar at bottom
		statusY := bh - 1
		for x := 0; x < bw; x++ {
			buf.SetCell(x, statusY, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    buffer.RGB(68, 71, 90),
			})
		}
		statusStyle := buffer.Style{}.
			WithFg(buffer.RGB(248, 248, 242)).
			WithBg(buffer.RGB(68, 71, 90))

		streaming := 0
		for _, b := range chat.Container().Blocks() {
			if b.State() == block.BlockStreaming {
				streaming++
			}
		}

		overlayInfo := ""
		if modal.Visible() {
			overlayInfo = " | [MODAL]"
		} else if popup.Visible() {
			overlayInfo = " | [POPUP]"
		}

		status := fmt.Sprintf(" %d blocks | %d streaming%s | ↑↓:scroll Tab:thinking m:modal p:popup q:quit",
			chat.Container().Len(), streaming, overlayInfo)
		buf.DrawTextClamped(1, statusY, status, statusStyle)

		// Render overlays on top
		if modal.Visible() {
			modal.Paint(buf)
		}
		if popup.Visible() {
			popup.Paint(buf)
		}
	})

	if err := a.Run(); err != nil {
		panic(err)
	}

	// Wait for stream goroutine to finish
	<-streamDone
}
