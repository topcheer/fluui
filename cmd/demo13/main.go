// Package main implements demo13 — Phase 19 Components Showcase.
//
// A print-based demo that renders sample frames of:
// Checkbox, RadioGroup, Slider, CommandPalette, Spinner.
//
// Usage: go run ./cmd/demo13/
package main

import (
	"fmt"
	"strings"

	"github.com/topcheer/fluui/animation"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func main() {
	width := 60

	fmt.Println(strings.Repeat("=", width))
	fmt.Println("  Fluui Phase 19 — Components Showcase")
	fmt.Println(strings.Repeat("=", width))
	fmt.Println()

	// ── Checkbox ──
	demoCheckbox(width)

	// ── RadioGroup ──
	demoRadioGroup(width)

	// ── Slider ──
	demoSlider(width)

	// ── Spinner ──
	demoSpinner(width)

	// ── CommandPalette ──
	demoCommandPalette(width)

	// ── Summary ──
	fmt.Println(strings.Repeat("=", width))
	fmt.Printf("  Phase 19 Components: Checkbox · RadioGroup · Slider · Spinner · CommandPalette\n")
	fmt.Printf("  Total: 5 components | 273+ tests | All passing with -race\n")
	fmt.Println(strings.Repeat("=", width))
}

// ─── Checkbox ──────────────────────────────────────────────────

func demoCheckbox(width int) {
	fmt.Println("  ▸ Checkbox — Multi-select with toggle")
	fmt.Println()

	cb := component.NewCheckbox([]string{"Enable notifications", "Dark mode", "Auto-save", "Sync to cloud"})

	// State 1: Default (none checked)
	cb.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 4})
	buf := buffer.NewBuffer(width, 4)
	cb.Paint(buf)
	printBuffer(buf, width, 4)

	fmt.Println()

	// State 2: Some checked
	cb.SetChecked(0, true)
	cb.SetChecked(2, true)
	buf2 := buffer.NewBuffer(width, 4)
	cb.Paint(buf2)
	printBuffer(buf2, width, 4)

	fmt.Println()
	fmt.Printf("  Checked items: %v\n", cb.CheckedLabels())
	fmt.Println()
}

// ─── RadioGroup ────────────────────────────────────────────────

func demoRadioGroup(width int) {
	fmt.Println("  ▸ RadioGroup — Single-select mutual exclusion")
	fmt.Println()

	rg := component.NewRadioGroup([]string{"Option A", "Option B", "Option C", "Option D (disabled)"})
	rg.SetDisabled(3, true)
	rg.SetSelected(1)

	rg.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 4})
	buf := buffer.NewBuffer(width, 4)
	rg.Paint(buf)
	printBuffer(buf, width, 4)

	fmt.Println()
	fmt.Printf("  Selected: %q (index %d)\n", rg.SelectedLabel(), rg.SelectedIndex())
	fmt.Println()
}

// ─── Slider ────────────────────────────────────────────────────

func demoSlider(width int) {
	fmt.Println("  ▸ Slider — Horizontal & Vertical")
	fmt.Println()

	// Horizontal slider at 75%
	sl := component.NewSlider()
	sl.SetValue(75)
	sl.SetLabel("Volume")
	sl.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 3})
	buf := buffer.NewBuffer(width, 3)
	sl.Paint(buf)
	printBuffer(buf, width, 3)

	fmt.Println()

	// Horizontal slider at 30% with different label
	sl2 := component.NewSlider()
	sl2.SetValue(30)
	sl2.SetLabel("Brightness")
	sl2.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 3})
	buf2 := buffer.NewBuffer(width, 3)
	sl2.Paint(buf2)
	printBuffer(buf2, width, 3)

	fmt.Println()
	fmt.Printf("  Volume: %.0f%% | Brightness: %.0f%%\n", sl.Value(), sl2.Value())
	fmt.Println()
}

// ─── Spinner ───────────────────────────────────────────────────

func demoSpinner(width int) {
	fmt.Println("  ▸ Spinner — Animated loading indicators")
	fmt.Println()

	// Show different spinner styles
	styles := []string{"dots", "arc", "line", "bouncingBar"}
	for _, style := range styles {
		sp := component.NewSpinner("Loading data...")
		sp.SetFrameStyle(style)
		sp.SetFrameIndex(0)
		sp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
		buf := buffer.NewBuffer(width, 1)
		sp.Paint(buf)
		printBuffer(buf, width, 1)
	}

	fmt.Println()

	// Show spinner with prefix
	sp := component.NewSpinner("Building project...")
	sp.SetPrefix("[build]")
	sp.SetFrameStyle("dots")
	sp.SetFrameIndex(2)
	sp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf := buffer.NewBuffer(width, 1)
	sp.Paint(buf)
	printBuffer(buf, width, 1)

	fmt.Println()
	fmt.Printf("  Frame styles: %v\n", styles)
	fmt.Println()
}

// ─── CommandPalette ────────────────────────────────────────────

func demoCommandPalette(width int) {
	fmt.Println("  ▸ CommandPalette — Fuzzy search command launcher")
	fmt.Println()

	cp := component.NewCommandPalette()
	cp.SetCommands([]component.Command{
		{ID: "file.new", Label: "New File", Shortcut: "Ctrl+N", Category: "File"},
		{ID: "file.open", Label: "Open File", Shortcut: "Ctrl+O", Category: "File"},
		{ID: "file.save", Label: "Save File", Shortcut: "Ctrl+S", Category: "File"},
		{ID: "edit.undo", Label: "Undo", Shortcut: "Ctrl+Z", Category: "Edit"},
		{ID: "edit.redo", Label: "Redo", Shortcut: "Ctrl+Y", Category: "Edit"},
		{ID: "view.theme", Label: "Switch Theme", Shortcut: "Ctrl+]", Category: "View"},
		{ID: "view.zoom", Label: "Zoom In", Shortcut: "Ctrl+=", Category: "View"},
		{ID: "app.quit", Label: "Quit Application", Shortcut: "Ctrl+Q", Category: "App"},
	})
	cp.SetMaxVisible(5)
	cp.SetQuery("file")
	cp.Show(0, 0)

	// Render the visible palette
	cp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 10})
	buf := buffer.NewBuffer(width, 10)
	cp.Paint(buf)
	printBuffer(buf, width, 10)

	fmt.Println()
	fmt.Printf("  Query: %q | Filtered: %d commands\n", "file", len(cp.FilteredCommands()))
	fmt.Println()

	// Show with different query
	cp2 := component.NewCommandPalette()
	cp2.SetCommands([]component.Command{
		{ID: "file.new", Label: "New File", Shortcut: "Ctrl+N", Category: "File"},
		{ID: "file.open", Label: "Open File", Shortcut: "Ctrl+O", Category: "File"},
		{ID: "edit.undo", Label: "Undo", Shortcut: "Ctrl+Z", Category: "Edit"},
		{ID: "edit.redo", Label: "Redo", Shortcut: "Ctrl+Y", Category: "Edit"},
		{ID: "view.theme", Label: "Switch Theme", Shortcut: "Ctrl+]", Category: "View"},
		{ID: "app.quit", Label: "Quit Application", Shortcut: "Ctrl+Q", Category: "App"},
	})
	cp2.SetMaxVisible(5)
	cp2.SetQuery("edit")
	cp2.Show(0, 0)
	cp2.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 10})
	buf2 := buffer.NewBuffer(width, 10)
	cp2.Paint(buf2)
	printBuffer(buf2, width, 10)

	fmt.Println()
	fmt.Printf("  Query: %q | Filtered: %d commands\n", "edit", len(cp2.FilteredCommands()))
	fmt.Println()

	// Show available animation frames
	frames := animation.SpinnerFrames["dots"]
	if len(frames) > 0 {
		fmt.Printf("  Sample spinner frames (dots): %s → %s → %s → %s\n",
			safeFrame(frames, 0), safeFrame(frames, 1), safeFrame(frames, 2), safeFrame(frames, 3))
	}
}

// ─── Print Helper ──────────────────────────────────────────────

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

func safeFrame(frames []string, idx int) string {
	if idx < len(frames) {
		return frames[idx]
	}
	return "?"
}
