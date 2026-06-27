// Package main demonstrates the search feature (Ctrl+F) with Fluui.
//
// Populates a chat with content, then shows how to trigger search,
// navigate results, and highlight matches.
//
// Keys:
//   Ctrl+F — open search
//   Enter  — next match (in search mode)
//   Esc    — close search
//   Esc / Ctrl+C — quit
package main

import (
	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	w, h := base.Size()
	chat := app.NewChatApp(w, h)
	chat.SetInputHeight(2)

	// Populate with searchable content
	content := []string{
		"Go is a statically typed, compiled programming language designed at Google.",
		"It is syntactically similar to C, but with memory safety, garbage collection,",
		"structural typing, and CSP-style concurrency.",
		"",
		"Go was designed by Robert Griesemer, Rob Pike, and Ken Thompson in 2007.",
		"It was announced in November 2009 and released in March 2012.",
		"",
		"Key features of Go include:",
		"- Minimal syntax and keywords",
		"- Fast compilation",
		"- Built-in concurrency (goroutines and channels)",
		"- Static linking (single binary output)",
		"- Cross-platform compilation",
	}

	md := chat.AddAssistantText()
	for _, line := range content {
		md.AppendDelta(line + "\n")
	}
	md.Complete()

	// Add instructions
	instr := chat.AddAssistantText()
	instr.AppendDelta("\nPress Ctrl+F to search this content.\n")
	instr.Complete()

	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()
		chat.SetSize(w, h)
		chat.Render(buf)
	})

	base.OnKey(func(k *term.KeyEvent) {
		if k.Key == term.KeyEscape {
			base.Quit()
			return
		}
		if k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0 {
			base.Quit()
			return
		}
		// ChatApp.HandleKey handles Ctrl+F (search), Up/Down (scroll), etc.
		chat.HandleKey(k)
		base.MarkDirty()
	})

	base.OnMouse(func(m *term.MouseEvent) {
		chat.HandleMouse(m)
		base.MarkDirty()
	})

	chat.OnQuit(func() { base.Quit() })

	base.Run()
}
