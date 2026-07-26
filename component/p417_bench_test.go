package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P417: Paint benchmarks for 20 previously-unbenched components (Direction D)

func benchPaint(b *testing.B, c Component, w, h int) {
	c.SetBounds(Rect{X: 0, Y: 0, W: w, H: h})
	buf := buffer.NewBuffer(w, h)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkP417_Accordion_Paint(b *testing.B) {
	a := NewAccordion([]AccordionItem{{Title: "A", Content: "text1"}, {Title: "B", Content: "text2"}})
	benchPaint(b, a, 30, 5)
}

func BenchmarkP417_AIProgress_Paint(b *testing.B) {
	a := NewAIProgress()
	a.SetPhase(AIPhaseGenerating)
	a.SetProgress(0.5)
	benchPaint(b, a, 20, 1)
}

func BenchmarkP417_AutoComplete_Paint(b *testing.B) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{{Label: "fmt"}, {Label: "os"}, {Label: "io"}})
	ac.SetQuery("")
	benchPaint(b, ac, 30, 5)
}

func BenchmarkP417_Avatar_Paint(b *testing.B) {
	a := NewAvatar("Alice")
	benchPaint(b, a, 3, 1)
}

func BenchmarkP417_Badge_Paint(b *testing.B) {
	bd := NewBadge("NEW", BadgeInfo)
	benchPaint(b, bd, 10, 1)
}

func BenchmarkP417_Banner_Paint(b *testing.B) {
	bn := NewBanner(BannerInfo, "Welcome to Fluui")
	benchPaint(b, bn, 40, 3)
}

func BenchmarkP417_BarChart_Paint(b *testing.B) {
	bc := NewBarChart()
	benchPaint(b, bc, 40, 10)
}

func BenchmarkP417_Border_Paint(b *testing.B) {
	bd := NewBorder(NewText("content"))
	benchPaint(b, bd, 20, 5)
}


func BenchmarkP417_Breadcrumb_PaintBench(b *testing.B) {
	br := NewBreadcrumb([]string{"Home", "Settings", "Audio"})
	benchPaint(b, br, 40, 1)
}

func BenchmarkP417_Button_Paint(b *testing.B) {
	btn := NewButton("Save")
	benchPaint(b, btn, 10, 1)
}

func BenchmarkP417_Calendar_Paint(b *testing.B) {
	c := NewCalendar()
	benchPaint(b, c, 30, 10)
}

func BenchmarkP417_Canvas_Paint(b *testing.B) {
	c := NewCanvas()
	benchPaint(b, c, 20, 5)
}

func BenchmarkP417_Checkbox_Paint(b *testing.B) {
	c := NewCheckbox([]string{"Option A", "Option B"})
	benchPaint(b, c, 20, 3)
}

func BenchmarkP417_Chip_Paint(b *testing.B) {
	c := NewChip("gpt-4")
	c.SetIcon("🤖")
	benchPaint(b, c, 10, 1)
}

func BenchmarkP417_Citations_Paint(b *testing.B) {
	c := NewCitationsBlock(nil)
	benchPaint(b, c, 40, 3)
}

func BenchmarkP417_CodeBlock_Paint(b *testing.B) {
	cb := NewCodeBlock("go", "fmt.Println(\"Hello\")")
	benchPaint(b, cb, 40, 5)
}

func BenchmarkP417_Collapsible_Paint(b *testing.B) {
	c := NewCollapsible("Details", NewText("content"))
	c.SetExpanded(true)
	benchPaint(b, c, 30, 3)
}

func BenchmarkP417_ColorPicker_Paint(b *testing.B) {
	cp := NewColorPicker()
	benchPaint(b, cp, 30, 5)
}

func BenchmarkP417_CommandPalette_Paint(b *testing.B) {
	cp := NewCommandPalette()
	cp.SetCommands([]Command{{Label: "Save"}, {Label: "Open"}})
	benchPaint(b, cp, 40, 5)
}

func BenchmarkP417_ContextMenu_Paint(b *testing.B) {
	cm := NewContextMenu()
	benchPaint(b, cm, 20, 5)
}
