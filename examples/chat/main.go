// Package main implements a basic AI chat example.
//
// Connects to an OpenAI-compatible LLM API and streams responses.
// Requires a .env file with FLUUI_LLM_API_KEY, FLUUI_LLM_BASE_URL, FLUUI_LLM_MODEL.
//
// Keys:
//   Enter     — send message
//   Up/Down   — scroll history / input history
//   Ctrl+C    — quit (or stop streaming, then quit)
package main

import (
	"fmt"
	"os"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/ai"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	cfg, err := ai.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
		os.Exit(1)
	}

	client := ai.NewClient(cfg)

	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	w, h := base.Size()
	chat := app.NewChatApp(w, h)
	chat.SetInputHeight(2)
	chat.SetAIClient(client)
	chat.SetSystemPrompt(cfg.SystemPrompt)
	chat.SetOnAIError(func(e error) {
		chat.AddUserMessage("Error: " + e.Error())
	})

	// Welcome message
	welcome := chat.AddAssistantText()
	welcome.AppendDelta("Welcome to Fluui Chat! Ask me anything.\n")
	welcome.AppendDelta("Press Enter to send, Up/Down to scroll, Ctrl+C to quit.\n")
	welcome.Complete()

	// Route events
	base.OnKey(func(k *term.KeyEvent) {
		if k.Key == term.KeyEscape {
			base.Quit()
			return
		}
		if k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0 {
			if chat.IsStreaming() {
				chat.StopStreaming()
				base.MarkDirty()
				return
			}
			base.Quit()
			return
		}
		chat.HandleKey(k)
		base.MarkDirty()
	})

	base.OnMouse(func(m *term.MouseEvent) {
		chat.HandleMouse(m)
		base.MarkDirty()
	})

	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()
		chat.SetSize(w, h)
		chat.Render(buf)
	})

	chat.OnQuit(func() { base.Quit() })
	base.Run()
}
