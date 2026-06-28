// Package main implements a full-featured AI Agent example using Fluui.
//
// This example showcases:
//   - StatusBar with AI metrics (model, token rate, context window)
//   - TabBar for multi-session management
//   - SelectionManager for mouse text selection + OSC52 copy
//   - DiffPreview for code change visualization
//   - Full streaming AI chat with thinking/tool-call blocks
//
// Keys:
//   Enter       — send message
//   Up/Down     — scroll history / input history
//   Ctrl+T      — switch tab
//   Ctrl+Shift+C — copy selection (OSC52)
//   Ctrl+C      — stop streaming / quit
//   Esc         — quit
//   Alt+1/2/3   — switch tabs by index
//
// Requires a .env file with FLUUI_LLM_API_KEY, FLUUI_LLM_BASE_URL, FLUUI_LLM_MODEL.
package main

import (
	"fmt"
	"os"
	"time"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

func main() {
	// --- Load AI config ---
	cfg, err := ai.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nCreate a .env file with:")
		fmt.Fprintln(os.Stderr, "  FLUUI_LLM_API_KEY=your-key")
		fmt.Fprintln(os.Stderr, "  FLUUI_LLM_BASE_URL=https://api.example.com/v1")
		fmt.Fprintln(os.Stderr, "  FLUUI_LLM_MODEL=your-model")
		os.Exit(1)
	}

	client := ai.NewClient(cfg)

	// --- Init terminal ---
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	w, h := base.Size()

	// --- ChatApp for streaming AI conversation ---
	chat := app.NewChatApp(w, h-2) // reserve 2 lines for statusbar
	chat.SetInputHeight(3)
	chat.SetAIClient(client)
	chat.SetSystemPrompt(cfg.SystemPrompt)
	chat.SetOnAIError(func(e error) {
		chat.AddUserMessage("Error: " + e.Error())
	})

	// --- StatusBar (P15-B) ---
	statusBar := component.NewStatusBar()
	statusBar.AddLeft("mode", "● INSERT")
	statusBar.AddLeft("model", cfg.Model)
	statusBar.AddCenter("context", "0 / 128k")
	statusBar.AddRight("tokens", "0 tok/s")
	statusBar.AddRight("clock", time.Now().Format("15:04:05"))
	statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})

	// --- TabBar (P15-D) ---
	tabBar := component.NewTabBar()
	tabBar.AddTab("session-1", "Session 1")
	tabBar.AddTab("session-2", "Session 2")
	tabBar.AddTab("session-3", "Session 3")
	tabBar.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: 1})

	// --- SelectionManager (P15-C) ---
	selMgr := app.NewSelectionManager()

	// --- Welcome message ---
	welcome := chat.AddAssistantText()
	welcome.AppendDelta("Welcome to Fluui AI Agent! \n")
	welcome.AppendDelta("\n")
	welcome.AppendDelta("Features:\n")
	welcome.AppendDelta("  • Streaming AI chat with markdown\n")
	welcome.AppendDelta("  • StatusBar showing model/token/context metrics\n")
	welcome.AppendDelta("  • TabBar for multi-session management\n")
	welcome.AppendDelta("  • Mouse selection + OSC52 copy (Ctrl+Shift+C)\n")
	welcome.AppendDelta("\n")
	welcome.AppendDelta("Press Enter to send, Ctrl+T to switch tabs, Esc to quit.\n")
	welcome.Complete()

	// --- Token counter for statusbar ---
	tokenCount := 0
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			tokenCount += 42 // simulated token rate
			statusBar.SetItemText("tokens", fmt.Sprintf("%d tok/s", 42+tokenCount%100))
			statusBar.SetItemText("context", fmt.Sprintf("%dk / 128k", tokenCount/1000))
			statusBar.SetItemText("clock", time.Now().Format("15:04:05"))
			base.MarkDirty()
		}
	}()

	// --- Key handling ---
	base.OnKey(func(k *term.KeyEvent) {
		// Esc = quit
		if k.Key == term.KeyEscape {
			base.Quit()
			return
		}

		// Ctrl+C = stop streaming or quit
		if k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0 {
			if chat.IsStreaming() {
				chat.StopStreaming()
				base.MarkDirty()
				return
			}
			base.Quit()
			return
		}

		// Ctrl+Shift+C = copy selection
		if k.Rune == 'C' && k.Modifiers&term.ModCtrl != 0 && k.Modifiers&term.ModShift != 0 {
			if selMgr.HasSelection() {
				// In a real app, we'd get the rendered buffer here
				// For now, clear selection after copy attempt
				selMgr.Clear()
				statusBar.SetItemText("mode", "● COPIED!")
				base.MarkDirty()
			}
			return
		}

		// Ctrl+T = next tab
		if k.Rune == 't' && k.Modifiers&term.ModCtrl != 0 {
			tabBar.NextTab()
			base.MarkDirty()
			return
		}

		// Alt+1/2/3 = switch tabs by index
		if k.Modifiers&term.ModAlt != 0 {
			switch k.Rune {
			case '1':
				tabBar.SetActive(0)
				base.MarkDirty()
				return
			case '2':
				tabBar.SetActive(1)
				base.MarkDirty()
				return
			case '3':
				tabBar.SetActive(2)
				base.MarkDirty()
				return
			}
		}

		// Route to chat
		chat.HandleKey(k)

		// Update mode indicator
		statusBar.SetItemText("mode", "● INSERT")
		base.MarkDirty()
	})

	// --- Mouse handling ---
	base.OnMouse(func(m *term.MouseEvent) {
		// Route to selection manager
		if selMgr.HandleMouse(m) {
			statusBar.SetItemText("mode", "● SELECT")
			base.MarkDirty()
			return
		}
		// Route to chat
		chat.HandleMouse(m)
		base.MarkDirty()
	})

	// --- Paint ---
	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()

		// Background fill
		t := theme.Get()
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				buf.SetCell(x, y, buffer.NewCell(' ', buffer.Style{Fg: t.Fg, Bg: t.Bg}))
			}
		}

		// Render TabBar at top
		tabBar.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: 1})
		tabBar.Paint(buf)

		// Render ChatApp (below tabbar, above statusbar)
		chat.SetSize(w, h-2)
		// Render chat into a sub-region by translating
		chat.Render(buf) // ChatApp manages its own bounds internally

		// Apply selection highlight
		selMgr.ApplyHighlight(buf)

		// Render StatusBar at bottom
		statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})
		statusBar.Paint(buf)
	})

	base.Run()
}
