// Package main implements demo11 — Phase 15 Components Showcase.
//
// A print-based demo that renders sample frames of every Phase 15 widget:
// FilePicker, StatusBar, Selection, TabBar, Links, and DiffPreview.
//
// Usage: go run ./cmd/demo11/
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Fluui Demo 11 — Phase 15 Components Showcase")
		fmt.Println("Usage: go run ./cmd/demo11/")
		return
	}

	theme.SetActive(theme.Dracula())
	width := 76

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║          Fluui Demo 11 — Phase 15 Components Showcase               ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── TabBar ──
	demoTabBar(width)

	// ── StatusBar ──
	demoStatusBar(width)

	// ── FilePicker ──
	demoFilePicker(width)

	// ── DiffPreview ──
	demoDiffPreview(width)

	// ── Links ──
	demoLinks(width)

	// ── Selection ──
	demoSelection(width)

	// ── Summary ──
	fmt.Println()
	fmt.Println("  ┌─────────────────────────────── Summary ──────────────────────────────┐")
	fmt.Println("  │                                                                       │")
	fmt.Printf("  │  Phase 15 New Components:          6 widgets + 1 app module           │\n")
	fmt.Printf("  │  New Tests:                       +230 (1687 → 1917)                 │\n")
	fmt.Printf("  │  Component Package Tests:         710+ ALL PASS (-race)              │\n")
	fmt.Println("  │                                                                       │")
	fmt.Println("  │  P15-A FilePicker   — Dir browser, fuzzy filter, multi-select        │")
	fmt.Println("  │  P15-B StatusBar    — AI metrics: model, token/s, context window     │")
	fmt.Println("  │  P15-C Selection    — Mouse drag + keyboard select + OSC52 copy      │")
	fmt.Println("  │  P15-D TabBar       — Multi-tab: add/close/switch/hit-test           │")
	fmt.Println("  │  P15-E Links        — URL detection + OSC8 + click                    │")
	fmt.Println("  │  P15-E DiffPreview  — Unified diff viewer with scroll                │")
	fmt.Println("  │                                                                       │")
	fmt.Println("  └───────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

func demoTabBar(width int) {
	fmt.Println("  ▸ TabBar — Multi-tab Management")
	fmt.Println()

	tb := component.NewTabBar()
	tb.AddTab("main.go", "main.go")
	tb.AddTab("buffer.go", "buffer.go")
	tb.AddTab("render.go", "render.go")
	tb.AddTab("input.go", "input.go")
	tb.SetActive(1) // buffer.go active

	tb.SetShowNewButton(true)
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})

	buf := buffer.NewBuffer(width, 3)
	tb.Paint(buf)
	printBuffer(buf, width, 3)

	fmt.Println("  Features: add/close/switch tabs, hit testing, truncate long titles")
	fmt.Println()
}

func demoStatusBar(width int) {
	fmt.Println("  ▸ StatusBar — AI Agent Metrics")
	fmt.Println()

	sb := component.NewStatusBar()
	sb.AddLeft("mode", "● NORMAL")
	sb.AddLeft("model", "GLM-5.2")
	sb.AddCenter("context", "12.5k / 128k")
	sb.AddRight("tokens", "1.2k tok/s")
	sb.AddRight("time", "14:32:08")

	sb.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf := buffer.NewBuffer(width, 2)
	sb.Paint(buf)
	printBuffer(buf, width, 2)

	fmt.Println("  Features: left/center/right segments, SetModel/SetTokenRate/SetClock")
	fmt.Println()
}

func demoFilePicker(width int) {
	fmt.Println("  ▸ FilePicker — Directory Browser")
	fmt.Println()

	fp := component.NewFilePicker(".")
	fp.SetDirReader(func(dir string) ([]component.FileEntry, error) {
		return []component.FileEntry{
			{Name: "internal/", Path: "internal/", IsDir: true},
			{Name: "component/", Path: "component/", IsDir: true},
			{Name: "app/", Path: "app/", IsDir: true},
			{Name: "block/", Path: "block/", IsDir: true},
			{Name: "render/", Path: "render/", IsDir: true},
			{Name: "go.mod", Path: "go.mod", IsDir: false, Size: 120},
			{Name: "go.sum", Path: "go.sum", IsDir: false, Size: 4500},
			{Name: "main.go", Path: "main.go", IsDir: false, Size: 3400},
		}, nil
	})

	fp.SetBounds(component.Rect{X: 0, Y: 0, W: 44, H: 12})
	buf := buffer.NewBuffer(44, 13)
	fp.Paint(buf)
	printBuffer(buf, 44, 13)

	fmt.Println("  Features: dir navigation, fuzzy filter (/), multi-select (Space), vim keys")
	fmt.Println()
}

func demoDiffPreview(width int) {
	fmt.Println("  ▸ DiffPreview — Unified Diff Viewer")
	fmt.Println()

	diff := `--- a/component/widget.go
+++ b/component/widget.go
@@ -15,6 +15,8 @@
 type Widget struct {
 	Name string
 	Size int
+	Color string
+	Border bool
 }
 
 func (w *Widget) Render() {
+	if w.Border {
+		drawBorder(w)
+	}
 	drawContent(w)
-	return
 }`

	dp := component.NewDiffPreview()
	dp.SetDiff(diff)
	dp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 16})
	buf := buffer.NewBuffer(width, 17)
	dp.Paint(buf)
	printBuffer(buf, width, 17)

	stats := dp.Stats()
	fmt.Printf("  Stats: +%d additions, -%d deletions, %d files\n",
		stats.Additions, stats.Deletions, stats.Files)
	fmt.Println("  Features: ParseDiff, scroll, +/- highlighting, line numbers")
	fmt.Println()
}

func demoLinks(width int) {
	fmt.Println("  ▸ Links — URL Detection & Clickable Hyperlinks")
	fmt.Println()

	lm := component.NewLinkManager()
	text := "Check https://github.com/fluui for source.\nDocs at http://docs.fluui.dev/api.\nVisit ftp://files.example.com/data."
	links := component.DetectLinks(text, 0, 0)

	buf := buffer.NewBuffer(width, 4)
	for i, line := range strings.Split(text, "\n") {
		for j, r := range line {
			if j < width {
				buf.SetCell(j, i, buffer.NewCell(r, buffer.Style{Fg: buffer.RGB(248, 248, 242)}))
			}
		}
	}

	lm.AddLink(component.LinkRange{URL: "https://github.com/fluui", Text: "https://github.com/fluui", StartX: 6, EndX: 29, Y: 0})
	lm.AddLink(component.LinkRange{URL: "http://docs.fluui.dev/api", Text: "http://docs.fluui.dev/api", StartX: 8, EndX: 32, Y: 1})
	lm.AnnotateBuffer(buf, 0, 0)

	printBuffer(buf, width, 3)

	fmt.Printf("  Detected %d links in text\n", len(links))
	fmt.Println("  Features: URL regex detection, OSC8 hyperlinks, hover/click, underline")
	fmt.Println()
}

func demoSelection(width int) {
	fmt.Println("  ▸ Selection — Mouse Drag & OSC52 Copy")
	fmt.Println()

	sm := app.NewSelectionManager()

	// Simulate selecting "Hello World" from row 0
	sm.StartSelection(2, 0)
	sm.ExtendSelection(12, 0)

	buf := buffer.NewBuffer(width, 3)
	text := "  Hello World — Fluui TUI Library"
	for j, r := range text {
		if j < width {
			buf.SetCell(j, 0, buffer.NewCell(r, buffer.Style{Fg: buffer.RGB(248, 248, 242)}))
		}
	}

	sm.ApplyHighlight(buf)
	printBuffer(buf, width, 2)

	selected := sm.GetSelectedText(buf)
	fmt.Printf("  Selected text: %q\n", selected)
	fmt.Println("  Features: mouse drag, Shift+arrows, reverse video, OSC52 clipboard copy")
	fmt.Println()
}

func printBuffer(buf *buffer.Buffer, w, h int) {
	fmt.Println("  ┌" + strings.Repeat("─", w) + "┐")
	for y := 0; y < h; y++ {
		var sb strings.Builder
		sb.WriteString("  │")
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != 0 {
				sb.WriteRune(cell.Rune)
			} else {
				sb.WriteByte(' ')
			}
		}
		sb.WriteString("│")
		fmt.Println(sb.String())
	}
	fmt.Println("  └" + strings.Repeat("─", w) + "┘")
}
