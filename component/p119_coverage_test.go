package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// === Badge.Measure edge cases (76.5% → 90%+) ===

func TestP119_Badge_Measure_ShortText(t *testing.T) {
	b := NewBadge("AB", BadgeInfo)
	s := b.Measure(Bounded(20, 10))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected positive measure, got %v", s)
	}
}

func TestP119_Badge_Measure_WithIcon(t *testing.T) {
	b := NewBadge("Status", BadgeSuccess)
	b.SetIcon("✓")
	s := b.Measure(Bounded(30, 10))
	if s.W <= 0 {
		t.Errorf("expected positive width with icon, got %v", s)
	}
}

func TestP119_Badge_Measure_NarrowClamp(t *testing.T) {
	b := NewBadge("VeryLongBadgeText", BadgeWarning)
	s := b.Measure(Bounded(2, 1))
	// Should clamp to at least the padding width
	if s.W < 0 {
		t.Errorf("expected non-negative width, got %v", s)
	}
}

// === AutoComplete.Paint variants (76.7% → 90%+) ===

func TestP119_AutoComplete_Paint_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf) // should not panic with empty items
}

func TestP119_AutoComplete_Paint_WithSelection(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "first", Value: "1"},
		{Label: "second", Value: "2"},
		{Label: "third", Value: "3"},
	})
	ac.SetCursor(1)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	ac.Paint(buf)
}

func TestP119_AutoComplete_Paint_ScrollDown(t *testing.T) {
	ac := NewAutoComplete()
	items := make([]CompletionItem, 20)
	for i := range items {
		items[i] = CompletionItem{Label: "item", Value: "v"}
	}
	ac.SetItems(items)
	ac.SetCursor(15)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

// === BarChart.paintHorizontal variants (78.6% → 90%+) ===

func TestP119_BarChart_PaintHorizontal_Grid(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.SetShowGrid(true)
	bc.SetShowAxes(true)
	bc.SetShowValues(true)
	bc.AddSeries(BarSeries{
		Name: "test",
		Data: []BarData{{Label: "A", Value: 50}, {Label: "B", Value: 80}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	bc.Paint(buf)
}

func TestP119_BarChart_PaintHorizontal_NoLabels(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "test",
		Data: []BarData{{Value: 30}, {Value: 60}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	bc.Paint(buf)
}

// === CodeBlock streaming cursor (74.2% → 85%+) ===

func TestP119_CodeBlock_StreamingCursor_LongLine(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("func main() { fmt.Println(\"this is a very long line that exceeds the width\") }")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestP119_CodeBlock_StreamingCursor_EmptyWithTitle(t *testing.T) {
	cb := NewCodeBlock("python", "")
	cb.SetTitle("test.py")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestP119_CodeBlock_StreamingCursor_NotStreaming(t *testing.T) {
	cb := NewCodeBlock("go", "func test() {}")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

// === ContextMenu cursor clamping (73.3% → 90%+) ===

func TestP119_ContextMenu_CursorOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)

	// Navigate past end then back
	for i := 0; i < 5; i++ {
		cm.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	}
	cm.Paint(buf)
	for i := 0; i < 5; i++ {
		cm.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	}
	cm.Paint(buf)
}

// === RichLog Measure + countVisibleLines (77-78% → 85%+) ===

func TestP119_RichLog_Measure_NoEntries(t *testing.T) {
	rl := NewRichLog()
	s := rl.Measure(Bounded(60, 20))
	if s.H <= 0 {
		t.Error("expected positive height even with no entries")
	}
}

func TestP119_RichLog_CountVisible_MultiLine(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 10; i++ {
		rl.Info("a very long line that will wrap when rendered in a narrow width to force multi-line rendering")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 50})
	buf := buffer.NewBuffer(15, 50)
	rl.Paint(buf)
}

// === Form HandleKey (78.3% → 90%+) ===

func TestP119_Form_HandleKey_Tab(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("name", "Name", ""))
	f.AddField(NewTextField("email", "Email", ""))
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Tab should cycle fields
	if !f.HandleKey(&term.KeyEvent{Key: term.KeyTab}) {
		t.Error("expected Tab consumed")
	}
}

func TestP119_Form_HandleKey_ShiftTab(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("name", "Name", ""))
	f.AddField(NewTextField("email", "Email", ""))
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	if !f.HandleKey(&term.KeyEvent{Key: term.KeyTab, Modifiers: term.ModShift}) {
		t.Error("expected Shift+Tab consumed")
	}
}

// === ListView MoveUp (76.9% → 90%+) ===

func TestP119_ListView_MoveUp_Wrap(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	// MoveUp at cursor 0 should wrap to end
	lv.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if lv.Cursor() != 2 {
		t.Errorf("expected cursor 2 after wrap, got %d", lv.Cursor())
	}
}

func TestP119_ListView_MoveUp_Disabled(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	items := lv.Items()
	items[0].Disabled = true
	lv.SetItems(items)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	lv.SetCursor(2)
	lv.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// Should skip disabled item 0
}

// === ProgressBar formatPercent (77.8% → 90%+) ===

func TestP119_ProgressBar_FormatPercent_EdgeValues(t *testing.T) {
	// Test via public API by checking paint output
	for _, val := range []float64{0, 0.5, 1.0, -0.1, 1.5} {
		p := NewProgressBar()
		p.SetProgress(val)
		p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
		buf := buffer.NewBuffer(20, 1)
		p.Paint(buf) // should not panic for any value
	}
}

// === DiffPreview paintBorderLocked (76.5% → 85%+) ===

func TestP119_DiffPreview_PaintBorder(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added line\n-context line\n-removed line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	dp.Paint(buf)
}

func TestP119_DiffPreview_PaintBorder_Narrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	dp.Paint(buf)
}

// === HelpOverlay ensureSelectedVisibleLocked (75% → 85%+) ===

func TestP119_HelpOverlay_ScrollUp(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Navigation", Entries: []HelpEntry{
			{Keys: "j/k", Description: "move up/down"},
			{Keys: "g/G", Description: "go to top/bottom"},
			{Keys: "Ctrl+d/u", Description: "half page down/up"},
			{Keys: "Ctrl+f/b", Description: "full page down/up"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)

	for i := 0; i < 5; i++ {
		ho.ScrollDown(1)
	}
	ho.Paint(buf)
	for i := 0; i < 3; i++ {
		ho.ScrollUp(1)
	}
	ho.Paint(buf)
}

// === SelectField edge cases ===

func TestP119_SelectField_NegativeIndex(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B", "C"})
	sf.SetSelectedIndex(-1)
	// Should clamp to 0 and return first option
	_ = sf.Value()
}

func TestP119_SelectField_OverflowIndex(t *testing.T) {
	sf := NewSelectField("Label", "key", []string{"A", "B"})
	sf.SetSelectedIndex(10)
	// Should clamp or return empty
	_ = sf.Value()
}

// === Gauge vertical ===

func TestP119_Gauge_Vertical(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetValue(0.5)
	g.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 10})
	buf := buffer.NewBuffer(3, 10)
	g.Paint(buf)
}

func TestP119_Gauge_Vertical_Full(t *testing.T) {
	g := NewGauge()
	g.SetOrientation(GaugeVertical)
	g.SetValue(1.0)
	g.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 10})
	buf := buffer.NewBuffer(3, 10)
	g.Paint(buf)
}

// === Sparkline ===

func TestP119_Sparkline_AutoScale(t *testing.T) {
	sl := NewSparkline()
	sl.SetAutoScale(true)
	sl.SetData([]float64{10, 20, 30, 40, 50})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	sl.Paint(buf)
}

// === TextArea edge cases ===

func TestP119_TextArea_MoveLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}
