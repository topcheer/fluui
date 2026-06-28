// Package main implements demo12 — Phase 17-18 Components Showcase.
//
// A print-based demo that renders sample frames of:
// VirtualScroller (10k items), Pagination, Dialog, AutoComplete, Wizard.
//
// Usage: go run ./cmd/demo12/
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("Fluui Demo 12 — Phase 17-18 Components Showcase")
		fmt.Println("Usage: go run ./cmd/demo12/")
		return
	}

	theme.SetActive(theme.Dracula())
	width := 70

	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("  ║        Fluui Demo 12 — Phase 17-18 Components Showcase          ║")
	fmt.Println("  ╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ── P17: VirtualScroller ──
	demoVirtualScroller(width)

	// ── P17: Pagination ──
	demoPagination(width)

	// ── P18: Dialog ──
	demoDialog(width)

	// ── P18: AutoComplete ──
	demoAutoComplete(width)

	// ── P18: Wizard ──
	demoWizard(width)

	// ── Summary ──
	fmt.Println()
	fmt.Println("  ┌─────────────────────────── Summary ───────────────────────────────┐")
	fmt.Println("  │                                                                     │")
	fmt.Println("  │  Phase 17:                                                           │")
	fmt.Println("  │    VirtualScroller — O(1) rendering for 10k+ items                  │")
	fmt.Println("  │    Pagination      — Page navigation with ellipsis                  │")
	fmt.Println("  │                                                                     │")
	fmt.Println("  │  Phase 18:                                                           │")
	fmt.Println("  │    Dialog          — Modal confirm/input/info dialogs               │")
	fmt.Println("  │    AutoComplete    — Fuzzy input completion popup                   │")
	fmt.Println("  │    Wizard          — Multi-step wizard with progress indicator      │")
	fmt.Println("  │                                                                     │")
	fmt.Println("  └─────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}

func demoVirtualScroller(width int) {
	fmt.Println("  ▸ VirtualScroller — 10,000 Items, O(1) Rendering")
	fmt.Println()

	vs := component.NewVirtualScroller()
	vs.SetHeader("Employee Directory (10,000 entries)")

	items := make([]component.VirtualItem, 10000)
	for i := range items {
		items[i] = component.VirtualItem{
			ID:   fmt.Sprintf("emp-%04d", i),
			Text: fmt.Sprintf("Employee #%04d — Dept %d", i, i%20),
		}
	}
	vs.SetItems(items)
	vs.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 14})
	vs.SetCursor(50)
	vs.ScrollTo(45)

	buf := buffer.NewBuffer(width, 15)
	vs.Paint(buf)
	printBuffer(buf, width, 15)

	start, end := vs.VisibleRange()
	fmt.Printf("  Total: 10,000 items | Visible: [%d, %d) | Cursor: %d\n\n",
		start, end, vs.Cursor())
}

func demoPagination(width int) {
	fmt.Println("  ▸ Pagination — Page Navigation")
	fmt.Println()

	p := component.NewPagination()
	p.SetTotalItems(500)
	p.SetItemsPerPage(25)
	p.SetPage(5)
	p.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 1})
	buf := buffer.NewBuffer(width, 1)
	p.Paint(buf)
	printBuffer(buf, width, 1)
	fmt.Printf("  Page %d/%d (items %d-%d)\n\n",
		p.CurrentPage()+1, p.TotalPages(), p.PageStartIndex()+1, p.PageEndIndex())
}

func demoDialog(width int) {
	fmt.Println("  ▸ Dialog — Confirm / Prompt / Info")
	fmt.Println()

	// Confirm dialog
	d := component.NewConfirmDialog("Exit Application", "Are you sure you want to quit?")
	d.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 7})
	buf := buffer.NewBuffer(width, 7)
	d.Paint(buf)
	printBuffer(buf, width, 7)

	// Prompt dialog
	d2 := component.NewPromptDialog("New Project", "Enter project name:", "my-app")
	d2.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 9})
	buf2 := buffer.NewBuffer(width, 9)
	d2.Paint(buf2)
	printBuffer(buf2, width, 9)
	fmt.Println()
}

func demoAutoComplete(width int) {
	fmt.Println("  ▸ AutoComplete — Fuzzy Match Popup")
	fmt.Println()

	ac := component.NewAutoComplete()
	ac.SetItems([]component.CompletionItem{
		{Label: "github.com/topcheer/fluui", Description: "TUI library", Category: "go"},
		{Label: "github.com/topcheer/k8ops", Description: "K8s operator", Category: "go"},
		{Label: "go build", Description: "Compile packages", Category: "cmd"},
		{Label: "go test -race", Description: "Race detector tests", Category: "cmd"},
		{Label: "go vet", Description: "Static analysis", Category: "cmd"},
	})
	ac.SetQuery("go")
	ac.Show(0, 0)
	ac.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 8})
	buf := buffer.NewBuffer(width, 8)
	ac.Paint(buf)
	printBuffer(buf, width, 8)
	fmt.Printf("  Query: %q | Filtered: %d items\n\n", ac.Query(), ac.FilteredCount())
}

func demoWizard(width int) {
	fmt.Println("  ▸ Wizard — Multi-Step Setup")
	fmt.Println()

	w := component.NewWizard([]*component.WizardStep{
		component.NewWizardStep("lang", "Choose Language").
			SetDescription("Select your preferred programming language"),
		component.NewWizardStep("theme", "Choose Theme").
			SetDescription("Pick a color theme for your workspace"),
		component.NewWizardStep("ready", "Ready to Go").
			SetDescription("Configuration complete! Click Finish to start."),
	})
	w.SetCurrentStep(1) // Show middle step with Back + Next
	w.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: 10})
	buf := buffer.NewBuffer(width, 10)
	w.Paint(buf)
	printBuffer(buf, width, 10)
	fmt.Printf("  Step %d/%d | Buttons: %v\n\n",
		w.CurrentStepIndex()+1, w.StepCount(), w.ButtonOrder())
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
