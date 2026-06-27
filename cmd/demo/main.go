package main

import (
	"fmt"
	"strings"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// This is a minimal demo showing Phase 1 capabilities:
// - Terminal rendering with double-buffer diff
// - Keyboard input (type characters, arrow keys)
// - Mouse click tracking
// - Color and style
// - Resize handling

func main() {
	app, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	// Track state
	var (
		typedText strings.Builder
		cursor    struct{ x, y int }
		mouseLog  []string
		bgColor   = buffer.NoColor()
	)
	cursor.x, cursor.y = 2, 2

	// Keyboard handler
	app.OnKey(func(k *term.KeyEvent) {
		switch {
		case k.Key == term.KeyEscape:
			app.Quit()
		case k.Key == term.KeyBackspace:
			if typedText.Len() > 0 {
				str := typedText.String()
				typedText.Reset()
				typedText.WriteString(str[:len(str)-1])
			}
		case k.Key == term.KeyUp && cursor.y > 0:
			cursor.y--
		case k.Key == term.KeyDown:
			cursor.y++
		case k.Key == term.KeyLeft && cursor.x > 0:
			cursor.x--
		case k.Key == term.KeyRight:
			cursor.x++
		case k.Rune != 0 && k.Modifiers == 0:
			typedText.WriteRune(k.Rune)
		}
	})

	// Mouse handler
	app.OnMouse(func(m *term.MouseEvent) {
		if m.Action == term.MouseDown {
			cursor.x = m.X
			cursor.y = m.Y
			mouseLog = append(mouseLog, fmt.Sprintf("(%d,%d) btn=%d", m.X, m.Y, m.Button))
			if len(mouseLog) > 5 {
				mouseLog = mouseLog[1:]
			}
		}
	})

	// Paint handler
	app.OnPaint(func(buf *buffer.Buffer) {
		w, h := buf.Width, buf.Height

		// Fill background
		bgStyle := buffer.DefaultStyle
		if !bgColor.IsDefault() {
			bgStyle = bgStyle.WithBg(bgColor)
		}
		buf.Fill(buffer.Cell{Rune: ' ', Width: 1, Bg: bgStyle.Bg})

		// Title bar
		titleStyle := buffer.Style{}.
			WithFg(buffer.RGB(255, 255, 255)).
			WithBg(buffer.RGB(40, 42, 54)).
			WithFlags(buffer.Bold)
		title := fmt.Sprintf(" Fluui Phase 1 Demo — %dx%d ", w, h)
		for x := 0; x < w; x++ {
			buf.SetCell(x, 0, buffer.Cell{Rune: ' ', Width: 1, Fg: titleStyle.Fg, Bg: titleStyle.Bg})
		}
		buf.DrawText(0, 0, title, titleStyle)

		// Instructions
		instructions := []string{
			"Type to add characters · Backspace to delete",
			"Arrow keys move cursor · Click to move cursor",
			"Ctrl+C or Esc to quit",
			"",
			"Typed text:",
		}
		instStyle := buffer.Style{}.WithFg(buffer.RGB(180, 190, 210))
		for i, line := range instructions {
			buf.DrawText(2, 2+i, line, instStyle)
		}

		// Show typed text
		textStyle := buffer.Style{}.
			WithFg(buffer.RGB(139, 233, 253)).
			WithFlags(buffer.Bold)
		buf.DrawTextClamped(2, 7, typedText.String(), textStyle)

		// Show cursor position
		if cursor.x >= 0 && cursor.x < w && cursor.y >= 0 && cursor.y < h {
			cell := buf.GetCell(cursor.x, cursor.y)
			cell.Flags |= buffer.Reverse
			buf.SetCell(cursor.x, cursor.y, cell)
		}

		// Mouse log
		if len(mouseLog) > 0 {
			logStyle := buffer.Style{}.WithFg(buffer.RGB(255, 121, 198)).AddFlags(buffer.Dim)
			buf.DrawText(2, h-6, "Mouse clicks:", logStyle.WithFlags(0))
			for i, entry := range mouseLog {
				buf.DrawText(2, h-5+i, entry, logStyle)
			}
		}

		// Color palette demo
		colors := []struct {
			name string
			c    buffer.Color
		}{
			{"Red", buffer.RGB(255, 85, 85)},
			{"Green", buffer.RGB(80, 250, 123)},
			{"Yellow", buffer.RGB(241, 250, 140)},
			{"Blue", buffer.RGB(189, 147, 249)},
			{"Cyan", buffer.RGB(139, 233, 253)},
			{"Pink", buffer.RGB(255, 121, 198)},
		}
		paletteY := h - 10
		if paletteY < 10 {
			paletteY = 10
		}
		for i, cc := range colors {
			x := 2 + i*12
			boxStyle := buffer.Style{}.WithBg(cc.c).WithFg(buffer.RGB(0, 0, 0))
			buf.DrawText(x, paletteY, "          ", boxStyle)
			buf.DrawText(x, paletteY+1, " "+cc.name+" ", boxStyle)
			buf.DrawText(x, paletteY+2, "          ", boxStyle)
		}

		// Status bar
		statusStyle := buffer.Style{}.
			WithFg(buffer.RGB(40, 42, 54)).
			WithBg(buffer.RGB(180, 190, 210))
		status := fmt.Sprintf(" cursor: (%d, %d) | text len: %d ", cursor.x, cursor.y, typedText.Len())
		for x := 0; x < w; x++ {
			buf.SetCell(x, h-1, buffer.Cell{Rune: ' ', Width: 1, Fg: statusStyle.Fg, Bg: statusStyle.Bg})
		}
		buf.DrawText(0, h-1, status, statusStyle)
	})

	app.Run()
}
