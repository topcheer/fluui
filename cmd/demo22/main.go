package main

import (
	"fmt"
	"os"
	"time"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/component/layout"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	app, err := fluui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cal := component.NewCalendarWithDate(time.Date(2025, 10, 15, 0, 0, 0, 0, time.UTC))

	statusBar := component.NewStatusBar()
	statusBar.AddLeft("mode", "Calendar Demo")
	statusBar.AddRight("hint", "Arrows: nav  n/p: month  t: today  q: quit")
	statusBar.AddCenter("selected", "Selected: "+cal.Selected().Format("2006-01-02"))

	cal.SetOnSelect(func(t time.Time) {
		statusBar.SetItemText("selected", "Selected: "+t.Format("2006-01-02"))
	})

	flex := layout.NewFlex(layout.FlexColumn)
	flex.AddChild(cal)
	flex.AddChild(statusBar)

	app.OnKey(func(k *term.KeyEvent) {
		if k.Rune == 'q' || k.Key == term.KeyEscape {
			app.Quit()
			return
		}
		if cal.HandleKey(k) {
			statusBar.SetItemText("selected", "Selected: "+cal.Selected().Format("2006-01-02"))
			app.MarkDirty()
		}
	})

	app.OnResize(func(w, h int) {
		flex.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: h})
		app.MarkDirty()
	})

	app.OnPaint(func(buf *buffer.Buffer) {
		flex.Paint(buf)
	})

	app.Run()
}
