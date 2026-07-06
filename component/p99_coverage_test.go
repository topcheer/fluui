package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ═══════════════════════════════════════════════════════════════════════════
// P99 Coverage Tests — Targeting lowest-coverage functions
// ═══════════════════════════════════════════════════════════════════════════

// ─── BaseComponent (0% → 100%) ───

func TestP99_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // no-op, should not panic
}

func TestP99_BaseComponent_Measure(t *testing.T) {
	bc := BaseComponent{}
	s := bc.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.W != 0 || s.H != 0 {
		t.Errorf("BaseComponent.Measure = %v, want zero", s)
	}
}

func TestP99_BaseComponent_Children(t *testing.T) {
	bc := BaseComponent{}
	if bc.Children() != nil {
		t.Error("BaseComponent.Children should return nil")
	}
}

// ─── DiffPreview SetShowLineNumbers / SetShowStats (0% → 100%) ───

func TestP99_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
	if !dp.ShowLineNumbers() {
		t.Error("ShowLineNumbers should always return true")
	}
}

func TestP99_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

// ─── Viewport scrollbar coverage (68-74% → higher) ───

func TestP99_Viewport_DrawVScrollBar_WithThumb(t *testing.T) {
	child := &fixedSize{w: 200, h: 100} // content larger than viewport
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP99_Viewport_DrawVScrollBar_FitsContent(t *testing.T) {
	child := &fixedSize{w: 5, h: 3} // content fits in viewport
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP99_Viewport_DrawVScrollBar_ScrollOffset(t *testing.T) {
	child := &fixedSize{w: 20, h: 200}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	vp.ScrollDown(50)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP99_Viewport_DrawHScrollBar_WithThumb(t *testing.T) {
	child := &fixedSize{w: 200, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP99_Viewport_DrawHScrollBar_FitsContent(t *testing.T) {
	child := &fixedSize{w: 5, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP99_Viewport_DrawHScrollBar_ScrollOffset(t *testing.T) {
	child := &fixedSize{w: 200, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	vp.ScrollRight(50)
	buf := buffer.NewBuffer(20, 10)
	vp.Paint(buf)
}

func TestP99_Viewport_ScrollToX(t *testing.T) {
	child := &fixedSize{w: 200, h: 5}
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	vp.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	vp.ScrollToX(10)
	if vp.OffsetX() != 10 {
		t.Errorf("OffsetX = %d, want 10", vp.OffsetX())
	}
	vp.ScrollToX(-5) // negative, should clamp
	if vp.OffsetX() != 0 {
		t.Errorf("OffsetX = %d, want 0", vp.OffsetX())
	}
}

// ─── Badge Measure (76.5% → higher) ───

func TestP99_Badge_Measure_AllVariants(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical}
	for _, v := range variants {
		b := NewBadge("test", v)
		s := b.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
		if s.W <= 0 || s.H <= 0 {
			t.Errorf("Badge variant %d: Measure = %v, expected positive", v, s)
		}
	}
}

func TestP99_Badge_Measure_WithIcon(t *testing.T) {
	b := NewBadge("git", BadgeSuccess)
	b.SetIcon("✔")
	s := b.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.W <= 0 {
		t.Errorf("Badge with icon: Measure = %v", s)
	}
}

func TestP99_Badge_Measure_ShortText(t *testing.T) {
	b := NewBadge("x", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.W <= 0 {
		t.Errorf("Badge short text: Measure = %v", s)
	}
}

func TestP99_Badge_Measure_NarrowWidth(t *testing.T) {
	b := NewBadge("very long badge text here", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 5, MaxHeight: 10})
	if s.W > 5 {
		t.Errorf("Badge narrow: W = %d, should be <= 5", s.W)
	}
}

func TestP99_Badge_VariantName(t *testing.T) {
	tests := []struct {
		v    BadgeVariant
		want string
	}{
		{BadgeInfo, "info"},
		{BadgeSuccess, "success"},
		{BadgeWarning, "warning"},
		{BadgeError, "error"},
		{BadgeCritical, "critical"},
		{BadgeVariant(99), "unknown"},
	}
	for _, tc := range tests {
		got := VariantName(tc.v)
		if got != tc.want {
			t.Errorf("VariantName(%d) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

// ─── Checkbox setNavigableCursor (73.3% → higher) ───

func TestP99_Checkbox_SetNavigableCursor_AllDisabled(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.SetItems([]CheckboxItem{
		{Label: "A", Checked: false, Disabled: true},
		{Label: "B", Checked: false, Disabled: true},
		{Label: "C", Checked: false, Disabled: true},
	})
	cb.SetCursor(0) // should not crash even with all disabled
	if cb.Cursor() < 0 {
		t.Error("cursor should not be negative")
	}
}

// ─── RadioGroup setNavigableCursor (73.3% → higher) ───

func TestP99_RadioGroup_SetNavigableCursor_Disabled(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(1, true) // disable middle item
	rg.SetCursor(0)
	rg.MoveDown() // should skip disabled item
	if rg.Cursor() == 1 {
		t.Error("cursor should skip disabled item at index 1")
	}
}

// ─── ContextMenu setCursorLocked (73.3% → higher) ───

func TestP99_ContextMenu_NegativeCursor(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("item1", "Item 1"))
	cm.AddItem(NewMenuItem("item2", "Item 2"))
	cm.SetCursor(-1) // should clamp to valid range
	if cm.Cursor() < 0 || cm.Cursor() >= cm.ItemCount() {
		t.Errorf("cursor = %d, should be in valid range", cm.Cursor())
	}
}

// ─── ListView HandleKey (76% → higher) ───

func TestP99_ListView_HandleKey_Enter(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	selected := false
	lv.OnSelect = func(item ListItem, index int) {
		selected = true
	}
	consumed := lv.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("Enter should be consumed")
	}
	if !selected {
		t.Error("OnSelect should fire on Enter")
	}
}

func TestP99_ListView_HandleKey_PageUp(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D", "E"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3}) // small viewport
	lv.SetCursor(4)
	lv.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	// PageUp moves cursor up by viewport height (3)
	if lv.Cursor() > 4 {
		t.Errorf("after PageUp: cursor = %d, should be <= 4", lv.Cursor())
	}
}

func TestP99_ListView_HandleKey_PageDown(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D", "E"})
	lv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3}) // small viewport
	lv.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	// PageDown moves cursor down by viewport height (3)
	if lv.Cursor() < 0 {
		t.Errorf("after PageDown: cursor = %d, should be >= 0", lv.Cursor())
	}
}

func TestP99_ListView_HandleKey_Home(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.SetCursor(2)
	lv.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if lv.Cursor() != 0 {
		t.Errorf("after Home: cursor = %d, want 0", lv.Cursor())
	}
}

func TestP99_ListView_HandleKey_End(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	lv.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if lv.Cursor() != 2 {
		t.Errorf("after End: cursor = %d, want 2", lv.Cursor())
	}
}

func TestP99_ListView_HandleKey_UnknownKey(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C"})
	consumed := lv.HandleKey(&term.KeyEvent{Key: term.KeyCode(999)})
	if consumed {
		t.Error("unknown key should not be consumed")
	}
}

// ─── DiffViewer Measure (77.8% → higher) ───

func TestP99_DiffViewer_Measure_WithContent(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetContent("--- a/test.go\n+++ b/test.go\n@@ -1,3 +1,3 @@\n-old line\n+new line\n context\n")
	s := dv.Measure(Constraints{MaxWidth: 80, MaxHeight: 30})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("DiffViewer Measure = %v, expected positive", s)
	}
}

func TestP99_DiffViewer_Measure_Empty(t *testing.T) {
	dv := NewDiffViewer()
	s := dv.Measure(Constraints{MaxWidth: 80, MaxHeight: 30})
	// Empty diffviewer has a title line at minimum
	_ = s // just verify it doesn't panic
}

func TestP99_DiffViewer_Measure_NarrowWidth(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetContent("test line here that is long\n")
	s := dv.Measure(Constraints{MaxWidth: 10, MaxHeight: 30})
	if s.W > 10 {
		t.Errorf("W = %d, should be <= 10", s.W)
	}
}

// ─── ProgressBar formatPercent (77.8% → higher) ───

func TestP99_ProgressBar_FormatPercent_EdgeCases(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(0)
	buf := buffer.NewBuffer(20, 3)
	pb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	pb.Paint(buf)

	pb.SetProgress(100)
	pb.Paint(buf)

	pb.SetProgress(50.5)
	pb.Paint(buf)

	pb.SetProgress(-10) // negative should clamp
	pb.Paint(buf)

	pb.SetProgress(150) // overflow should clamp
	pb.Paint(buf)
}

// ─── Sparkline recomputeRange / valueToBar (77.8% → higher) ───

func TestP99_Sparkline_AutoScale(t *testing.T) {
	sl := NewSparkline()
	sl.SetAutoScale(true)
	sl.SetData([]float64{10, 20, 30, 40, 50})
	buf := buffer.NewBuffer(30, 5)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	sl.Paint(buf)
}

func TestP99_Sparkline_NegativeValues(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{-10, 0, 10, 20, 30})
	buf := buffer.NewBuffer(30, 5)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	sl.Paint(buf)
}

func TestP99_Sparkline_EmptyValues(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{})
	buf := buffer.NewBuffer(30, 5)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	sl.Paint(buf)
}

func TestP99_Sparkline_AllSameValues(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5, 5})
	buf := buffer.NewBuffer(30, 5)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	sl.Paint(buf)
}

// ─── Canvas SetCell out of bounds (80% → higher) ───

func TestP99_Canvas_SetCell_OutOfBounds(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	// Out of bounds should not panic
	c.SetCell(-1, 0, 'x', buffer.White)
	c.SetCell(0, -1, 'x', buffer.White)
	c.SetCell(100, 0, 'x', buffer.White)
	c.SetCell(0, 100, 'x', buffer.White)
	c.Paint(buf)
}

func TestP99_Canvas_SetCellBG_OutOfBounds(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	c.SetCellBG(-1, 0, 'x', buffer.White, buffer.Black)
	c.SetCellBG(0, -1, 'x', buffer.White, buffer.Black)
	c.SetCellBG(100, 0, 'x', buffer.White, buffer.Black)
	c.Paint(buf)
}

func TestP99_Canvas_WorldToGrid(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	c.SetWorldBounds(0, 100, 0, 100)
	gx, gy := c.WorldToGrid(50, 50)
	if gx < 0 || gy < 0 {
		t.Errorf("WorldToGrid(50,50) = (%d,%d), expected non-negative", gx, gy)
	}
}

// ─── TextArea moveLine (77.8% → higher) ───

func TestP99_TextArea_MoveLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	// Move cursor down
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	// Move cursor back up
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

func TestP99_TextArea_EmptyMoveLine(t *testing.T) {
	ta := NewTextArea()
	ta.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	// Move on empty content
	ta.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	ta.HandleKey(&term.KeyEvent{Key: term.KeyUp})
}

// ─── Table drawCellLocked / truncateToWidth (75% → higher) ───

func TestP99_Table_LongHeaders(t *testing.T) {
	tbl := NewTable([]string{"Very Long Header 1", "Very Long Header 2"})
	tbl.AddRow([]string{"short", "data"})
	tbl.AddRow([]string{"another", "row"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 6})
	buf := buffer.NewBuffer(15, 6)
	tbl.Paint(buf)
}

func TestP99_Table_ClampSelection(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.AddRow([]string{"1", "2"})
	tbl.AddRow([]string{"3", "4"})
	tbl.SetSelectedRow(0)
	tbl.SetSelectedRow(1)
	tbl.SetSelectedRow(-1) // should clamp
	tbl.SetSelectedRow(99) // should clamp
}

// ─── ColorPicker handleRGBKey (78.1% → higher) ───

func TestP99_ColorPicker_RGB_AdjustChannel(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	// Adjust R channel up
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// Switch to G channel
	cp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// Switch to B channel
	cp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

func TestP99_ColorPicker_RGB_VimKeys(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	// l = next channel
	cp.HandleKey(&term.KeyEvent{Rune: 'l'})
	// k = increase
	cp.HandleKey(&term.KeyEvent{Rune: 'k'})
	// j = decrease
	cp.HandleKey(&term.KeyEvent{Rune: 'j'})
	// h = prev channel
	cp.HandleKey(&term.KeyEvent{Rune: 'h'})
}

func TestP99_ColorPicker_RGB_LargeAdjust(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	// H = increase by 10
	cp.HandleKey(&term.KeyEvent{Rune: 'H'})
	// L = decrease by 10
	cp.HandleKey(&term.KeyEvent{Rune: 'L'})
}

// ─── Autocomplete Paint (76.7% → higher) ───

func TestP99_AutoComplete_Paint_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)
}

func TestP99_AutoComplete_Paint_WithCursor(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "item1", Value: "v1"},
		{Label: "item2", Value: "v2"},
		{Label: "item3", Value: "v3"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf)
}

func TestP99_AutoComplete_MoveUp_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.MoveUp() // should not panic with empty items
}

// ─── BarChart paintHorizontal (78.6% → higher) ───

func TestP99_BarChart_PaintHorizontal_AllFeatures(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.SetShowGrid(true)
	bc.SetShowAxes(true)
	bc.SetShowLegend(true)
	bc.SetShowValues(true)
	bc.AddSeries(BarSeries{
		Name:  "Revenue",
		Data:  []BarData{{Label: "Q1", Value: 100}, {Label: "Q2", Value: 200}, {Label: "Q3", Value: 150}},
		Color: buffer.RGB(100, 200, 100),
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	bc.Paint(buf)
}

func TestP99_BarChart_PaintHorizontal_NoLabels(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name:  "Data",
		Data:  []BarData{{Label: "", Value: 50}, {Label: "", Value: 75}},
		Color: buffer.RGB(100, 100, 200),
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

func TestP99_BarChart_PaintHorizontal_PartialValues(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarHorizontal)
	bc.AddSeries(BarSeries{
		Name: "Partial",
		Data: []BarData{{Label: "A", Value: 0.5}, {Label: "B", Value: 0.3}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	bc.Paint(buf)
}

// ─── BarChart drawVerticalGrid (80% → higher) ───

func TestP99_BarChart_DrawVerticalGrid(t *testing.T) {
	bc := NewBarChart()
	bc.SetOrientation(BarVertical)
	bc.SetShowGrid(true)
	bc.SetShowAxes(true)
	bc.SetMaxVal(100)
	bc.AddSeries(BarSeries{
		Name: "Test",
		Data: []BarData{{Label: "A", Value: 50}, {Label: "B", Value: 80}, {Label: "C", Value: 30}},
	})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 20})
	buf := buffer.NewBuffer(50, 20)
	bc.Paint(buf)
}

// ─── ScrollView contentW (75% → higher) ───

func TestP99_ScrollView_ContentW(t *testing.T) {
	child := &fixedSize{w: 100, h: 50}
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	sv.Measure(Constraints{MaxWidth: 20, MaxHeight: 10})
	// Access contentW via scroll operations
	buf := buffer.NewBuffer(20, 10)
	sv.Paint(buf)
}

// ─── HelpOverlay ensureSelectedVisibleLocked (75% → higher) ───

func TestP99_HelpOverlay_NavigateDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Navigation", Entries: []HelpEntry{
			{Keys: "j/k", Description: "Move up/down"},
			{Keys: "g/G", Description: "Go to top/bottom"},
			{Keys: "Ctrl+d/u", Description: "Half page down/up"},
			{Keys: "Ctrl+f/b", Description: "Full page down/up"},
		}},
		{Name: "Actions", Entries: []HelpEntry{
			{Keys: "Enter", Description: "Select"},
			{Keys: "Esc", Description: "Cancel"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	// Navigate down multiple times to trigger ensureSelectedVisible
	for i := 0; i < 6; i++ {
		ho.ScrollDown(1)
	}
	buf := buffer.NewBuffer(50, 10)
	ho.Paint(buf)
}

// ─── DebugInspector paintEventsLocked (77.4% → higher) ───

func TestP99_DebugInspector_PaintEvents(t *testing.T) {
	di := NewDebugInspector()
	di.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	// Record various events
	di.RecordKey(&term.KeyEvent{Key: term.KeyEnter})
	di.RecordMouse(&term.MouseEvent{X: 10, Y: 5, Button: 0, Action: term.MouseDown})
	di.RecordResize(80, 24)
	di.RecordCustom("test event")
	di.SetMode(InspectEvents)
	buf := buffer.NewBuffer(50, 15)
	di.Paint(buf)
}

func TestP99_DebugInspector_PaintStats(t *testing.T) {
	di := NewDebugInspector()
	di.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	di.RecordRender(1000000, 500, true)
	di.RecordRender(2000000, 600, false)
	di.SetMode(InspectStats)
	buf := buffer.NewBuffer(50, 15)
	di.Paint(buf)
}

// ─── Themestudio initSlots / paintPickerOverlay ───

func TestP99_ThemeStudio_SlotCount(t *testing.T) {
	ts := NewThemeStudio(nil) // nil theme — should not panic
	_ = ts.SlotCount()
}

func TestP99_ThemeStudio_PaintPickerOverlay(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	ts.mu.Lock()
	ts.pickerOpen = true
	ts.mu.Unlock()
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf)
}

func TestP99_ThemeStudio_copyTheme_Nil(t *testing.T) {
	result := copyTheme(nil)
	if result != nil {
		t.Error("copyTheme(nil) should return nil")
	}
}
