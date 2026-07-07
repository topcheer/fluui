// demo20 — RichLog component showcase
// A structured log viewer with level filtering, auto-scroll, and keyboard navigation.
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/topcheer/fluui"
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

	rl := component.NewRichLog()
	rl.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 22})
	rl.SetShowLevels(true)
	rl.SetShowTime(true)
	rl.SetAutoScroll(true)

	// StatusBar
	status := component.NewStatusBar()
	status.AddLeft("mode", "LIVE")
	status.AddRight("hint", "q=quit r=toggle j/k=scroll 1-5=level g/G=top/bot")

	// Generate some initial log entries
	rl.Info("RichLog demo started")
	rl.Info("This component displays structured log entries")
	rl.Warn("You can filter by minimum level (keys 1-5)")
	rl.Error("Errors are highlighted in red")
	rl.Debug("Debug messages are dimmed gray")

	// Background goroutine generating random log entries
	stop := make(chan struct{})
	go func() {
		msgs := []string{
			"Connection established",
			"Data transfer complete",
			"Cache invalidated",
			"Request processed",
			"Background sync running",
			"Health check passed",
			"Metrics collected",
			"Worker pool resized",
		}
		levels := []component.LogLevel{
			component.LogDebug,
			component.LogInfo,
			component.LogInfo,
			component.LogInfo,
			component.LogWarn,
			component.LogError,
		}
		ticker := time.NewTicker(800 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				msg := msgs[rand.Intn(len(msgs))]
				level := levels[rand.Intn(len(levels))]
				rl.Write(level, fmt.Sprintf("[pid:%d] %s", rand.Intn(9999), msg))
				app.MarkDirty()
			}
		}
	}()
	defer close(stop)

	app.OnKey(func(k *term.KeyEvent) {
		switch k.Rune {
		case 'q':
			app.Quit()
		case 'r':
			rl.SetAutoScroll(!rl.AutoScroll())
			if rl.AutoScroll() {
				status.SetItemText("mode", "LIVE")
			} else {
				status.SetItemText("mode", "PAUSED")
			}
		case '1':
			rl.SetMinLevel(component.LogDebug)
			status.SetItemText("mode", "DEBUG+")
		case '2':
			rl.SetMinLevel(component.LogInfo)
			status.SetItemText("mode", "INFO+")
		case '3':
			rl.SetMinLevel(component.LogWarn)
			status.SetItemText("mode", "WARN+")
		case '4':
			rl.SetMinLevel(component.LogError)
			status.SetItemText("mode", "ERROR+")
		case '5':
			rl.SetMinLevel(component.LogFatal)
			status.SetItemText("mode", "FATAL+")
		}

		// Ctrl+C quits
		if k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0 {
			app.Quit()
		}
	})

	app.OnPaint(func(buf *buffer.Buffer) {
		buf.Fill(buffer.BlankCell)
		rl.Paint(buf)
		status.SetBounds(component.Rect{X: 0, Y: 23, W: 80, H: 1})
		status.Paint(buf)
	})

	app.Run()
}
