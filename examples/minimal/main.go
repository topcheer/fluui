// Package main implements a minimal Fluui example.
//
// Displays colored text on the terminal screen. Press Esc or Ctrl+C to exit.
package main

import (
	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	app, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	app.OnPaint(func(buf *buffer.Buffer) {
		w, h := buf.Width, buf.Height
		y := h/2 - 1

		buf.DrawText(w/2-6, y, "Hello, Fluui!", buffer.Style{
			Fg:    buffer.RGB(0xFF, 0x79, 0xC6),
			Flags: buffer.Bold,
		})
		buf.DrawText(w/2-10, y+2, "Press Esc or Ctrl+C to exit", buffer.Style{
			Fg: buffer.RGB(0x62, 0x72, 0xA4),
		})
	})

	app.OnKey(func(k *term.KeyEvent) {
		if k.Key == term.KeyEscape {
			app.Quit()
		}
	})

	app.Run()
}
