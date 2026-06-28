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
	"time"

	"github.com/topcheer/fluui/animation"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func main() {
	width := 60

	fmt.Println("╔" + strings.Repeat("═", width) + "╗")
	fmt.Println("║" + center("Fluui Demo 13 — Phase 19 Components Showcase", width) + "║")
	fmt.Println("╚" + strings.Repeat("═", width) + "╝")
	fmt.Println()

	// ── P19-A: Checkbox ──
	demoCheckbox(width)

	// ── P19-A: RadioGroup ──
	demoRadioGroup(width)

	// ── P19-A: Slider ──
	demoSlider(width)

	// ── P19-B: CommandPalette ──
	demoCommandPalette(width)

	// ── P19-B: Spinner ──
	demoSpinner(width)

	// ── Summary ──
	fmt.Println("┌" + strings.Repeat("─", width) + "┐")
	fmt.Println("│ " + fmt.Sprintf("Components: Checkbox | RadioGroup | Slider | CommandPalette | Spinner") + pad(27))
	fmt.Println("│ " + fmt.Sprintf("Phase 19 — 5 new widgets, 2504 tests, 40 packages, ~82K LOC") + pad(38))
	fmt.Println("└" + strings.Repeat("─", width) + "┘")
}

// ─── Checkbox ──────────────────────────────────────────────────

func demoCheckbox(width int) {
	fmt.Println("━━━ Checkbox ━━━")
	fmt.Println()

	cb := component.NewCheckbox([]string{"Enable notifications", "Auto-save on exit", "Show line numbers"})
	cb.SetChecked(0, true)
	cb.SetChecked(2, true)

	cb.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 5})
	buf := buffer.NewBuffer(width, 5)
	cb.Paint(buf)
	printBuffer(buf, width, 5)

	fmt.Println("  State: checked items =", cb.CheckedLabels())
	fmt.Println()

	// Checkbox with disabled item
	fmt.Println("  With disabled item:")
	cb2 := component.NewCheckbox([]string{"Option A", "Option B", "Option C"})
	cb2.SetItems([]component.CheckboxItem{
		{Label: "Option A", Checked: true},
		{Label: "Option B (disabled)", Disabled: true},
		{Label: "Option C"},
	})

	cb2.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 5})
	buf2 := buffer.NewBuffer(width, 5)
	cb2.Paint(buf2)
	printBuffer(buf2, width, 5)
	fmt.Println()
}

// ─── RadioGroup ────────────────────────────────────────────────

func demoRadioGroup(width int) {
	fmt.Println("━━━ RadioGroup ━━━")
	fmt.Println()

	rg := component.NewRadioGroup([]string{"Light theme", "Dark theme", "Nord theme", "Gruvbox"})
	rg.SetSelected(1) // Dark theme selected

	rg.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 6})
	buf := buffer.NewBuffer(width, 6)
	rg.Paint(buf)
	printBuffer(buf, width, 6)

	fmt.Printf("  Selected: index=%d, label=%q\n\n", rg.SelectedIndex(), rg.SelectedLabel())

	// RadioGroup with disabled item
	fmt.Println("  With disabled option:")
	rg2 := component.NewRadioGroup([]string{"Git", "Mercurial", "SVN", "Perforce"})
	rg2.SetSelected(0)
	rg2.SetDisabled(3, true)

	rg2.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 6})
	buf2 := buffer.NewBuffer(width, 6)
	rg2.Paint(buf2)
	printBuffer(buf2, width, 6)
	fmt.Println()
}

// ─── Slider ────────────────────────────────────────────────────

func demoSlider(width int) {
	fmt.Println("━━━ Slider ━━━")
	fmt.Println()

	// Horizontal slider at various values
	values := []float64{0, 25, 50, 75, 100}
	for _, v := range values {
		s := component.NewSliderWithRange(0, 100, v, 1)
		s.SetLabel("Volume")
		s.SetShowValue(true)
		s.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 3})
		buf := buffer.NewBuffer(width, 3)
		s.Paint(buf)
		printBuffer(buf, width, 3)
	}
	fmt.Println()

	// Vertical slider
	fmt.Println("  Vertical slider:")
	sv := component.NewSliderWithRange(0, 100, 60, 5)
	sv.SetOrientation(component.SliderVertical)
	sv.SetLabel("Brightness")
	sv.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)
	printBuffer(buf, 20, 10)
	fmt.Println()
}

// ─── CommandPalette ───────────────────────────────────────────

func demoCommandPalette(width int) {
	fmt.Println("━━━ CommandPalette ━━━")
	fmt.Println()

	cp := component.NewCommandPalette()
	cp.SetCommands([]component.Command{
		{ID: "file.new", Label: "New File", Shortcut: "Ctrl+N", Category: "File"},
		{ID: "file.open", Label: "Open File", Shortcut: "Ctrl+O", Category: "File"},
		{ID: "file.save", Label: "Save File", Shortcut: "Ctrl+S", Category: "File"},
		{ID: "edit.find", Label: "Find", Shortcut: "Ctrl+F", Category: "Edit"},
		{ID: "edit.replace", Label: "Replace", Shortcut: "Ctrl+H", Category: "Edit"},
		{ID: "view.theme", Label: "Switch Theme", Shortcut: "Ctrl+T", Category: "View"},
		{ID: "view.zoom", Label: "Zoom In", Shortcut: "Ctrl++", Category: "View"},
		{ID: "git.commit", Label: "Git Commit", Shortcut: "", Category: "Git"},
		{ID: "git.push", Label: "Git Push", Shortcut: "", Category: "Git"},
		{ID: "term.new", Label: "New Terminal", Shortcut: "Ctrl+Shift+T", Category: "Terminal"},
	})

	// Show palette with no filter
	cp.Show(0, 0)
	cp.SetMaxVisible(6)
	cp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 12})
	buf := buffer.NewBuffer(width, 12)
	cp.Paint(buf)
	fmt.Println("  All commands:")
	printBuffer(buf, width, 12)
	fmt.Println()

	// Show filtered results
	cp.SetQuery("git")
	cp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 8})
	buf2 := buffer.NewBuffer(width, 8)
	cp.Paint(buf2)
	fmt.Println("  Filtered by 'git':")
	printBuffer(buf2, width, 8)
	fmt.Println()

	// Show filtered results for 'file'
	cp.SetQuery("file")
	cp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 8})
	buf3 := buffer.NewBuffer(width, 8)
	cp.Paint(buf3)
	fmt.Println("  Filtered by 'file':")
	printBuffer(buf3, width, 8)
	fmt.Println()
}

// ─── Spinner ──────────────────────────────────────────────────

func demoSpinner(width int) {
	fmt.Println("━━━ Spinner ━━━")
	fmt.Println()

	// Show multiple spinner frame styles
	styles := []string{"dots", "arc", "line", "bouncingBar", "dotsScrolling"}

	for _, styleName := range styles {
		sp := component.NewSpinner("Loading...")
		sp.SetPrefix("[" + styleName + "]")
		sp.SetFrameStyle(styleName)
		sp.SetFrameIndex(0)
		sp.Start()

		sp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
		buf := buffer.NewBuffer(width, 1)
		sp.Paint(buf)
		printBuffer(buf, width, 1)
	}
	fmt.Println()

	// Show animation frames for the "dots" style
	fmt.Println("  Animation frames (dots style):")
	frames := animation.SpinnerFrames["dots"]
	if len(frames) > 0 {
		// Show 8 consecutive frames
		for i := 0; i < 8 && i < len(frames); i++ {
			sp := component.NewSpinner("Working...")
			sp.SetPrefix(fmt.Sprintf("[frame %d]", i))
			sp.SetFrameStyle("dots")
			sp.SetFrameIndex(i)

			sp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
			buf := buffer.NewBuffer(width, 1)
			sp.Paint(buf)
			printBuffer(buf, width, 1)
		}
	}
	fmt.Println()

	// Spinner with different labels
	fmt.Println("  Various states:")
	states := []struct {
		label  string
		prefix string
	}{
		{"Fetching data...", "[http]"},
		{"Building project...", "[build]"},
		{"Running tests...", "[test]"},
		{"Deploying to production...", "[deploy]"},
	}

	for _, st := range states {
		sp := component.NewSpinner(st.label)
		sp.SetPrefix(st.prefix)
		sp.SetFrameStyle("arc")
		sp.SetFrameIndex(int(time.Now().UnixNano()) % 6)

		sp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
		buf := buffer.NewBuffer(width, 1)
		sp.Paint(buf)
		printBuffer(buf, width, 1)
	}
	fmt.Println()
}

// ─── Helpers ──────────────────────────────────────────────────

func center(s string, w int) string {
	if len(s) >= w {
		return s[:w]
	}
	pad := (w - len(s)) / 2
	return strings.Repeat(" ", pad) + s + strings.Repeat(" ", w-len(s)-pad)
}

func pad(extra int) string {
	if extra < 0 {
		return ""
	}
	return strings.Repeat(" ", extra)
}

func printBuffer(buf *buffer.Buffer, w, h int) {
	fmt.Println("  ┌" + strings.Repeat("─", w) + "┐")
	for y := 0; y < h; y++ {
		var sb strings.Builder
		sb.WriteString("  │")
		for x := 0; x < w; x++ {
			cell := buf.GetCell(x, y)
			if cell.Width > 0 {
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
