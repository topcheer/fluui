// demo17 — ColorPicker component showcase.
//
// Interactive color picker with three modes:
//   1. Palette — 256-color grid
//   2. RGB     — Three sliders (Red, Green, Blue)
//   3. Hex     — Direct hex code input
//
// Keys: Tab to switch modes, arrows to navigate, Enter to print selected color.
package main

import (
	"fmt"
	"os"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

type colorPickerDemo struct {
	app  *fluui.App
	cp   *component.ColorPicker
	info string
}

func (d *colorPickerDemo) handleKey(k *term.KeyEvent) {
	// Tab to switch mode
	if k.Key == term.KeyTab && k.Modifiers&term.ModShift == 0 {
		d.cp.NextMode()
		mode := d.cp.Mode()
		switch mode {
		case component.PickerPalette:
			d.info = "Mode: Palette (256 colors)"
		case component.PickerRGB:
			d.info = "Mode: RGB Sliders"
		case component.PickerHex:
			d.info = "Mode: Hex Input"
		}
		d.app.MarkDirty()
		return
	}

	// Shift+Tab to go backwards
	if k.Modifiers&term.ModShift != 0 && k.Key == term.KeyTab {
		d.cp.PrevMode()
		d.app.MarkDirty()
		return
	}

	// Number keys to jump to mode
	if k.Rune >= '1' && k.Rune <= '3' {
		switch k.Rune {
		case '1':
			d.cp.SetMode(component.PickerPalette)
		case '2':
			d.cp.SetMode(component.PickerRGB)
		case '3':
			d.cp.SetMode(component.PickerHex)
		}
		d.app.MarkDirty()
		return
	}

	// Enter prints the selected color
	if k.Key == term.KeyEnter {
		c := d.cp.Color()
		fmt.Fprintf(os.Stderr, "\nSelected: %s (%s)\n", c.String(), component.ColorName(c))
	}

	// Delegate to color picker for all other keys
	if d.cp.HandleKey(k) {
		d.app.MarkDirty()
	}
}

func main() {
	app, err := fluui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	cp := component.NewColorPicker()

	demo := &colorPickerDemo{
		app:  app,
		cp:   cp,
		info: "Mode: Palette (256 colors)",
	}

	// Track color changes
	cp.OnChange = func(c buffer.Color) {
		demo.info = "Color: " + component.ColorName(c)
	}

	app.OnKey(demo.handleKey)

	app.OnPaint(func(buf *buffer.Buffer) {
		buf.Fill(buffer.BlankCell)

		w, h := buf.Width, buf.Height

		// Title
		title := " ColorPicker — Interactive Color Selection "
		for i, r := range title {
			if i >= w {
				break
			}
			buf.SetCell(i, 0, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    buffer.NamedColor(buffer.NamedCyan),
				Flags: buffer.Bold,
			})
		}

		// Info bar
		for i, r := range demo.info {
			if i >= w {
				break
			}
			buf.SetCell(i, 1, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    buffer.NamedColor(buffer.NamedYellow),
			})
		}

		// Color picker bounds
		cpW := 50
		cpH := h - 4
		if cpH < 16 {
			cpH = 16
		}
		cpX := (w - cpW) / 2
		cpY := 3
		if cpX < 0 {
			cpX = 0
		}

		cp.SetBounds(component.Rect{X: cpX, Y: cpY, W: cpW, H: cpH})
		cp.Paint(buf)

		// Bottom help
		helpY := h - 1
		help := " Tab: switch mode  1/2/3: jump to mode  Enter: print color  Esc: quit"
		for i, r := range help {
			if cpX+i >= w {
				break
			}
			buf.SetCell(cpX+i, helpY, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    buffer.NamedColor(buffer.NamedWhite),
				Flags: buffer.Dim,
			})
		}
	})

	app.OnQuit(func() {})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
