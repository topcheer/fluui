// Package main implements demo8 — Phase 12 Rich Widgets Showcase.
//
// A print-based demo that renders sample frames of every Phase 12 widget.
// Usage: go run ./cmd/demo8/
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/topcheer/fluui/animation"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Fluui Demo 8 — Phase 12 Rich Widgets Showcase")
		fmt.Println("Usage: go run ./cmd/demo8/")
		return
	}

	theme.SetActive(theme.Dracula())
	width := 70

	// ─── ProgressBar Demo ─────────────────────────────
	printBanner(width, "ProgressBar")

	for _, pct := range []float64{0, 25, 50, 75, 100} {
		bar := component.NewProgressBar()
		bar.SetProgress(pct)
		bar.SetShowPercentage(true)
		bar.SetBounds(component.Rect{X: 0, Y: 0, W: width - 10, H: 1})
		buf := buffer.NewBuffer(width, 1)
		bar.Paint(buf)
		fmt.Println("  " + renderLine(buf, width))
	}

	// Indeterminate bar
	fmt.Println()
	indBar := component.NewProgressBar()
	indBar.SetMode(component.ProgressIndeterminate)
	indBar.SetBounds(component.Rect{X: 0, Y: 0, W: width - 10, H: 1})
	for i := 0; i < 15; i++ {
		indBar.Tick()
	}
	buf := buffer.NewBuffer(width, 1)
	indBar.Paint(buf)
	fmt.Println("  " + renderLine(buf, width))

	fmt.Println()
	fmt.Println("  Color: red (0%) -> yellow (50%) -> green (100%)")
	fmt.Println()

	// ─── StatusIndicator Demo ─────────────────────────
	printBanner(width, "StatusIndicator")

	spinners := []string{"dots", "arc", "arrow", "bouncing", "moon"}
	for _, name := range spinners {
		sp := animation.NewSpinner(name)
		sp.Update(150 * time.Millisecond)
		fmt.Printf("  %-12s %s  Working on it...\n", name, sp.Current())
	}

	fmt.Println()
	fmt.Println("  (running)  spinner + message")
	fmt.Println("  (stopped)  space + message")
	fmt.Println()

	// ─── Table Demo ───────────────────────────────────
	printBanner(width, "Table")

	data := [][]string{
		{"ID", "Name", "Lang", "Stars", "Status"},
		{"001", "fluui", "Go", "1.2k", "Active"},
		{"002", "bubbletea", "Go", "28k", "Active"},
		{"003", "tview", "Go", "8.1k", "Active"},
		{"004", "lipgloss", "Go", "8.5k", "Active"},
		{"005", "termui", "Go", "13k", "Maint"},
		{"006", "gocui", "Go", "9.8k", "Archived"},
		{"007", "tcolor", "Go", "567", "Active"},
	}

	colWidths := []int{6, 16, 8, 10, 10}
	for r, row := range data {
		line := ""
		for c, cell := range row {
			if c < len(colWidths) {
				line += fmt.Sprintf("%-*s", colWidths[c], cell)
			}
		}
		if r == 0 {
			fmt.Printf("  \033[1m%s\033[0m\n", strings.TrimSpace(line))
			fmt.Println("  " + strings.Repeat("-", width-4))
		} else {
			prefix := " "
			if r%2 == 0 {
				prefix = " "
			}
			fmt.Printf(" %s%s\n", prefix, strings.TrimSpace(line))
		}
	}

	fmt.Println()
	fmt.Printf("  %d rows x %d columns -- sortable, scrollable, zebra\n", len(data)-1, len(data[0]))
	fmt.Println()

	// ─── Tree Demo ────────────────────────────────────
	printBanner(width, "Tree")

	treeLines := []struct {
		depth int
		icon  string
		label string
	}{
		{0, "v", "fluui/"},
		{1, "v", "component/"},
		{2, "", "table.go"},
		{2, "", "tree.go"},
		{2, "", "form.go"},
		{2, "", "progress.go"},
		{2, "", "status.go"},
		{1, ">", "block/"},
		{1, ">", "render/"},
		{1, ">", "theme/"},
		{1, ">", "animation/"},
		{0, ">", "cmd/"},
		{0, "", "go.mod"},
	}

	for _, n := range treeLines {
		indent := strings.Repeat("  ", n.depth)
		marker := " "
		if n.depth == 0 || n.depth == 1 {
			marker = n.icon
		}
		fmt.Printf("  %s%s %s\n", indent, marker, n.label)
	}

	fmt.Println()
	fmt.Println("  > collapsed  v expanded  leaf nodes show filename")
	fmt.Println()

	// ─── Form Demo ────────────────────────────────────
	printBanner(width, "Form")

	forms := []struct {
		typ     string
		label   string
		value   string
		checked bool
	}{
		{"text", "Username", "fluui_dev", false},
		{"text", "API Key", "sk-xxxx...xxxx", false},
		{"checkbox", "Enable Streaming", "", true},
		{"checkbox", "Verbose Logging", "", false},
		{"select", "Model", "GLM-4-Plus", false},
		{"select", "Theme", "Dracula", false},
	}

	for _, f := range forms {
		switch f.typ {
		case "text":
			fmt.Printf("  %-18s [%s]\n", f.label+":", f.value)
		case "checkbox":
			box := "[ ]"
			if f.checked {
				box = "[x]"
			}
			fmt.Printf("  %s %s\n", box, f.label)
		case "select":
			fmt.Printf("  %-18s < %s >\n", f.label+":", f.value)
		}
	}

	fmt.Println()
	fmt.Println("  TextField + CheckboxField + SelectField")
	fmt.Println("  Tab/Shift+Tab nav + Enter submit + Escape cancel")
	fmt.Println()

	// ─── Workflow Demo ────────────────────────────────
	printBanner(width, "Workflow")

	steps := []struct {
		name   string
		status string
		tm     string
	}{
		{"1. Parse Input", "done", "0.3s"},
		{"2. Plan", "done", "1.2s"},
		{"3. Read Files", "running", "..."},
		{"4. Analyze", "pending", ""},
		{"5. Generate", "pending", ""},
		{"6. Validate", "pending", ""},
	}

	icons := map[string]string{"done": "v", "running": "~", "pending": "o"}
	for i, s := range steps {
		icon := icons[s.status]
		fmt.Printf("  %s %-20s", icon, s.name)
		if s.tm != "" {
			fmt.Printf("  %s", s.tm)
		}
		fmt.Println()
		if i < len(steps)-1 {
			fmt.Println("  |")
		}
	}

	fmt.Println()
	fmt.Println("  Progress: 2/6 (33%)")
	fmt.Println("  o pending  ~ running  v done  x failed  / skipped")
	fmt.Println()

	// ─── Summary ──────────────────────────────────────
	printBanner(width, "Phase 12 Complete")
	fmt.Println("  ProgressBar       -- determinate + indeterminate (20 tests)")
	fmt.Println("  StatusIndicator   -- spinner + message (22 tests)")
	fmt.Println("  Table             -- sortable grid (43 tests, by gg_dev1)")
	fmt.Println("  Tree              -- expandable hierarchy (49 tests, by gg_dev2)")
	fmt.Println("  Form              -- text/checkbox/select (58 tests, by gg_dev3)")
	fmt.Println("  WorkflowBlock     -- agent visualization (41 tests, by gg_arch)")
	fmt.Println()
	fmt.Println("  Total Phase 12: 233 new tests across 6 widgets")
	fmt.Println()
}

func printBanner(width int, title string) {
	pad := width - len(title) - 4
	if pad < 0 {
		pad = 0
	}
	left := pad / 2
	right := pad - left
	fmt.Printf("  +%s %s %s+\n", strings.Repeat("=", left), title, strings.Repeat("=", right))
}

func renderLine(buf *buffer.Buffer, width int) string {
	var sb strings.Builder
	for x := 0; x < width && x < buf.Width; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != 0 {
			sb.WriteRune(c.Rune)
		} else {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
