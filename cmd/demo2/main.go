// Package main implements an interactive Phase 2 component system demo.
//
// Showcases: Border (with title), Text, Flex layout (row & column),
// and ScrollView with keyboard scrolling.
package main

import (
	"fmt"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/component/layout"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	app, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer app.Close()

	// --- Build scrollable content ---
	// 40 lines of styled text to demonstrate scrolling.
	scrollCol := layout.NewFlexGap(layout.FlexColumn, 0)
	pink := buffer.Style{}.WithFg(buffer.RGB(255, 121, 198))
	cyan := buffer.Style{}.WithFg(buffer.RGB(139, 233, 253))
	green := buffer.Style{}.WithFg(buffer.RGB(80, 250, 123))
	yellow := buffer.Style{}.WithFg(buffer.RGB(241, 250, 140))
	purple := buffer.Style{}.WithFg(buffer.RGB(189, 147, 249))
	orange := buffer.Style{}.WithFg(buffer.RGB(255, 184, 108))

	styles := []buffer.Style{pink, cyan, green, yellow, purple, orange}
	for i := 0; i < 40; i++ {
		line := fmt.Sprintf("  Line %2d: The quick brown fox jumps over the lazy dog", i+1)
		txt := component.NewText(line)
		txt.Style = styles[i%len(styles)]
		scrollCol.AddChild(txt)
	}
	scrollView := component.NewScrollView(scrollCol)

	// --- Welcome heading ---
	welcomeText := component.NewText("  Welcome to Fluui Phase 2!")
	welcomeText.Style = buffer.Style{}.
		WithFg(buffer.RGB(255, 121, 198)).
		WithFlags(buffer.Bold)

	subText := component.NewText("  A Go TUI framework with components, layout, and rendering.")
	subText.Style = buffer.Style{}.WithFg(buffer.RGB(98, 114, 164))

	// --- Components showcase row ---
	compRow := layout.NewFlexGap(layout.FlexRow, 3)

	t1 := component.NewText("[Text]")
	t1.Style = green.WithFlags(buffer.Bold)
	compRow.AddChild(t1)

	t2 := component.NewText("[Border]")
	t2.Style = yellow.WithFlags(buffer.Bold)
	compRow.AddChild(t2)

	t3 := component.NewText("[Flex]")
	t3.Style = purple.WithFlags(buffer.Bold)
	compRow.AddChild(t3)

	t4 := component.NewText("[ScrollView]")
	t4.Style = pink.WithFlags(buffer.Bold)
	compRow.AddChild(t4)

	compBorder := component.NewBorder(compRow)
	compBorder.Title = " Components "
	compBorder.Style = buffer.Style{}.WithFg(buffer.RGB(98, 114, 164))

	// --- Root border ---
	rootBorder := component.NewBorder(subText) // placeholder child, replaced each frame
	rootBorder.Title = " Fluui Phase 2 Demo "
	rootBorder.Style = buffer.Style{}.WithFg(buffer.RGB(189, 147, 249))

	// --- Keyboard handler ---
	app.OnKey(func(k *term.KeyEvent) {
		switch {
		case k.Key == term.KeyEscape || (k.Rune == 'q' && k.Modifiers == 0):
			app.Quit()
		case k.Key == term.KeyUp:
			scrollView.ScrollUp(1)
		case k.Key == term.KeyDown:
			scrollView.ScrollDown(1)
		case k.Key == term.KeyLeft:
			scrollView.ScrollUp(5)
		case k.Key == term.KeyRight:
			scrollView.ScrollDown(5)
		}
	})

	// --- Paint handler ---
	app.OnPaint(func(buf *buffer.Buffer) {
		w, h := buf.Width, buf.Height

		// Fill background (Dracula bg)
		bgCell := buffer.Cell{Rune: ' ', Width: 1, Bg: buffer.RGB(40, 42, 54)}
		buf.Fill(bgCell)

		// Root border fills the entire terminal
		rootBorder.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
		rootBorder.Paint(buf)

		// Inner area (inside the border frame)
		innerX := 1
		innerY := 1
		innerW := w - 2
		innerH := h - 2
		if innerW < 1 {
			innerW = 1
		}
		if innerH < 1 {
			innerH = 1
		}

		// Layout sections top-to-bottom inside the border:
		//   y+0: Welcome heading (1 line)
		//   y+1: Subtitle (1 line)
		//   y+2: (blank gap)
		//   y+3: Components border (3 lines: border top/mid/bottom)
		//   y+6: (blank gap)
		//   y+7: ScrollView (remaining space)
		//   bottom-1: Status bar (1 line)
		y := innerY

		// Welcome heading
		welcomeText.SetBounds(component.Rect{X: innerX, Y: y, W: innerW, H: 1})
		welcomeText.Paint(buf)
		y++

		// Subtitle
		subText.SetBounds(component.Rect{X: innerX, Y: y, W: innerW, H: 1})
		subText.Paint(buf)
		y += 2 // subtitle + gap

		// Components showcase border
		compBorder.SetBounds(component.Rect{X: innerX, Y: y, W: innerW, H: 3})
		compBorder.Paint(buf)
		y += 4 // border (3) + gap

		// ScrollView takes remaining space (minus status bar)
		statusH := 1
		scrollH := innerH - (y - innerY) - statusH
		if scrollH < 1 {
			scrollH = 1
		}
		scrollView.SetBounds(component.Rect{X: innerX, Y: y, W: innerW, H: scrollH})
		scrollView.Paint(buf)

		// Status bar
		statusY := innerY + innerH - statusH
		if statusY > y {
			// Clear status line
			for x := innerX; x < innerX+innerW; x++ {
				buf.SetCell(x, statusY, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Bg:    buffer.RGB(248, 248, 242),
				})
			}
			statusStyle := buffer.Style{}.
				WithFg(buffer.RGB(40, 42, 54)).
				WithBg(buffer.RGB(248, 248, 242))
			status := fmt.Sprintf("  Up/Down: scroll   Left/Right: page   q/Esc: quit   offset: %d/%d",
				scrollView.Offset(), scrollView.MaxOffset())
			buf.DrawTextClamped(innerX, statusY, status, statusStyle)
		}
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}
