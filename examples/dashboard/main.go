// Package main implements a realtime system dashboard using Fluui.
//
// This example showcases:
//   - Gauge for CPU, Memory, and Disk usage with color thresholds
//   - Sparkline for network traffic history
//   - Table for process listing with zebra striping
//   - StatusBar with live clock and status indicators
//   - Tab-based panel switching
//
// Keys:
//
//	1/2/3     — switch panels (overview, processes, network)
//	q/Esc     — quit
//	r         — toggle auto-refresh
//	Up/Down   — scroll table
//	Ctrl+C    — quit
package main

import (
	"fmt"
	"math/rand"
	"time"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	// --- Components ---

	// Gauges
	cpuGauge := component.NewGauge()
	cpuGauge.SetRange(0, 100)
	cpuGauge.SetLabel("CPU")
	cpuGauge.SetShowValue(true)

	memGauge := component.NewGauge()
	memGauge.SetRange(0, 100)
	memGauge.SetLabel("Memory")
	memGauge.SetShowValue(true)

	diskGauge := component.NewGauge()
	diskGauge.SetRange(0, 100)
	diskGauge.SetLabel("Disk")
	diskGauge.SetShowValue(true)

	// Sparklines
	netIn := component.NewSparkline()
	netIn.SetLabel("Network In (KB/s)")
	netIn.SetColorMode(component.ColorGradient)
	netIn.SetAutoScale(true)

	netOut := component.NewSparkline()
	netOut.SetLabel("Network Out (KB/s)")
	netOut.SetColorMode(component.ColorGradient)
	netOut.SetAutoScale(true)

	// Process table
	processTable := component.NewTable(
		[]string{"PID", "Name", "CPU%", "Mem%", "Status"},
		generateProcesses(8)...,
	)
	processTable.SetZebra(true)

	// Tab bar
	tabBar := component.NewTabBar()
	tabBar.AddTab("overview", "1:Overview")
	tabBar.AddTab("processes", "2:Processes")
	tabBar.AddTab("network", "3:Network")
	tabBar.SetActive(0)

	// Status bar
	statusBar := component.NewStatusBar()
	statusBar.AddLeft("app", " Fluui Dashboard")
	statusBar.AddCenter("status", " MONITORING")
	statusBar.AddRight("time", "")
	statusBar.AddRight("hint", " [q]uit [r]efresh [1/2/3]panel ")

	// --- State ---
	activeTab := 0
	autoRefresh := true

	// --- Auto-refresh ticker ---
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if autoRefresh {
				cpuGauge.SetValue(float64(rand.Intn(40) + 20))
				memGauge.SetValue(float64(rand.Intn(30) + 40))
				diskGauge.SetValue(float64(rand.Intn(20) + 50))

				netIn.Push(float64(rand.Intn(500)+100), 40)
				netOut.Push(float64(rand.Intn(300)+50), 40)

				processTable.SetRows(generateProcesses(8))

				base.MarkDirty()
			}
		}
	}()

	// --- Key handling ---
	base.OnKey(func(k *term.KeyEvent) {
		switch {
		case k.Key == term.KeyEscape || k.Rune == 'q':
			base.Quit()

		case k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0:
			base.Quit()

		case k.Rune == '1':
			activeTab = 0
			tabBar.SetActive(0)

		case k.Rune == '2':
			activeTab = 1
			tabBar.SetActive(1)

		case k.Rune == '3':
			activeTab = 2
			tabBar.SetActive(2)

		case k.Rune == 'r':
			autoRefresh = !autoRefresh
			if autoRefresh {
				statusBar.SetItemText("status", " MONITORING")
			} else {
				statusBar.SetItemText("status", " PAUSED")
			}

		case k.Key == term.KeyUp:
			row := processTable.SelectedRow()
			if row > 0 {
				processTable.SetSelectedRow(row - 1)
			}

		case k.Key == term.KeyDown:
			row := processTable.SelectedRow()
			if row < processTable.RowCount()-1 {
				processTable.SetSelectedRow(row + 1)
			}
		}

		base.MarkDirty()
	})

	// --- Rendering ---
	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()

		// Layout: TabBar (1) + content (h-2) + StatusBar (1)
		tabBar.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: 1})
		tabBar.Paint(buf)

		statusBar.SetItemText("time", fmt.Sprintf(" %s ", time.Now().Format("15:04:05")))
		statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})
		statusBar.Paint(buf)

		// Content area
		contentY := 1
		contentH := h - 2

		sepStyle := buffer.Style{Fg: buffer.RGB(0x55, 0x55, 0x55)}

		switch activeTab {
		case 0: // Overview: gauges + sparklines
			gaugeW := w / 3
			if gaugeW < 20 {
				gaugeW = w
			}

			if w >= 60 {
				// Side-by-side gauges
				cpuGauge.SetBounds(component.Rect{X: 0, Y: contentY + 1, W: gaugeW, H: 3})
				memGauge.SetBounds(component.Rect{X: gaugeW, Y: contentY + 1, W: gaugeW, H: 3})
				diskGauge.SetBounds(component.Rect{X: gaugeW * 2, Y: contentY + 1, W: w - gaugeW*2, H: 3})
				cpuGauge.Paint(buf)
				memGauge.Paint(buf)
				diskGauge.Paint(buf)

				// Network sparklines below
				sparkStart := contentY + 5
				sparkH := contentH - 6
				if sparkH < 3 {
					sparkH = 3
				}
				halfH := sparkH / 2
				if halfH < 2 {
					halfH = 2
				}
				netIn.SetBounds(component.Rect{X: 0, Y: sparkStart, W: w, H: halfH})
				netOut.SetBounds(component.Rect{X: 0, Y: sparkStart + halfH, W: w, H: sparkH - halfH})
				netIn.Paint(buf)
				netOut.Paint(buf)
			} else {
				// Stacked gauges for narrow terminals
				cpuGauge.SetBounds(component.Rect{X: 0, Y: contentY + 1, W: w, H: 3})
				memGauge.SetBounds(component.Rect{X: 0, Y: contentY + 5, W: w, H: 3})
				diskGauge.SetBounds(component.Rect{X: 0, Y: contentY + 9, W: w, H: 3})
				cpuGauge.Paint(buf)
				memGauge.Paint(buf)
				diskGauge.Paint(buf)
			}

		case 1: // Processes: table
			processTable.SetBounds(component.Rect{X: 0, Y: contentY, W: w, H: contentH})
			processTable.Paint(buf)

		case 2: // Network: sparklines
			halfH := contentH / 2
			if halfH < 3 {
				halfH = 3
			}
			netIn.SetBounds(component.Rect{X: 0, Y: contentY, W: w, H: halfH})
			netOut.SetBounds(component.Rect{X: 0, Y: contentY + halfH, W: w, H: contentH - halfH})
			netIn.Paint(buf)
			netOut.Paint(buf)
		}

		// Draw separators
		for x := 0; x < w; x++ {
			buf.SetCell(x, 0, buffer.NewCell('─', sepStyle))
		}
		for x := 0; x < w; x++ {
			buf.SetCell(x, h-2, buffer.NewCell('─', sepStyle))
		}
	})

	base.Run()
}

func generateProcesses(n int) [][]string {
	names := []string{"fluui", "go-build", "chrome", "code", "docker", "node",
		"python3", "rust-analyzer", "postgres", "redis-server", "ssh", "tmux"}
	statuses := []string{"running", "sleeping", "running", "running", "sleeping", "running"}

	rows := make([][]string, n)
	for i := 0; i < n; i++ {
		pid := 1000 + i*7
		name := names[i%len(names)]
		cpu := fmt.Sprintf("%.1f", rand.Float64()*30)
		mem := fmt.Sprintf("%.1f", rand.Float64()*15)
		status := statuses[i%len(statuses)]
		rows[i] = []string{fmt.Sprintf("%d", pid), name, cpu, mem, status}
	}
	return rows
}
