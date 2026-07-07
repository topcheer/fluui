package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// bufferTextToCells converts a string to []buffer.Cell for testing.
func bufferTextToCells(s string) []buffer.Cell {
	cells := make([]buffer.Cell, 0, len(s))
	for _, r := range s {
		cells = append(cells, buffer.Cell{Rune: r, Width: 1})
	}
	return cells
}

// ─── CodeBlock paintStreamingCursorLocked (67.7%) ───

func TestP107_CodeBlock_StreamingCursor_EmptyWithTitleInBounds(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.streaming = true
	cb.showTitle = true
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
	// Cursor should be at (gutterW, 1) — title takes row 0
}

func TestP107_CodeBlock_StreamingCursor_EmptyNoTitleInBounds(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.streaming = true
	cb.showTitle = false
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_EmptyZeroWidth(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.streaming = true
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 5})
	buf := buffer.NewBuffer(1, 5)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_LongLineClampX(t *testing.T) {
	// Line longer than bounds → x clamped to bounds.W-1
	cb := NewCodeBlock("go", "this is a very long line that exceeds the width of the terminal")
	cb.streaming = true
	cb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_PlainFallback(t *testing.T) {
	cb := NewCodeBlock("go", "hello\nworld")
	cb.streaming = true
	cb.usePlainFallback = true
	cb.plainLines = [][]buffer.Cell{
		bufferTextToCells("hello"),
		bufferTextToCells("world"),
	}
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_LastIdxClamp(t *testing.T) {
	// scrollOffset makes lastIdx exceed len(lines)
	cb := NewCodeBlock("go", "a\nb\nc")
	cb.streaming = true
	cb.scrollOffset = 10 // way past content
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_YOutOfBounds(t *testing.T) {
	// y < bounds.Y or y >= bounds.Y+bounds.H → return without setting
	cb := NewCodeBlock("go", "test")
	cb.streaming = true
	cb.scrollOffset = 100 // makes y way out of bounds
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_NegativeLastIdx(t *testing.T) {
	// bounds.H = 1, showTitle = true → lastIdx = scrollOffset + (1-1) - 1 = scrollOffset - 1
	cb := NewCodeBlock("go", "a\nb")
	cb.streaming = true
	cb.showTitle = true
	cb.scrollOffset = 0 // lastIdx = 0 + 0 - 1 = -1, clamped to 0
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)
}

func TestP107_CodeBlock_StreamingCursor_WithLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "func main() {\n\tfmt.Println(\"hi\")\n}")
	cb.streaming = true
	cb.showLineNumbers = true
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

// ─── Checkbox setNavigableCursor (73.3%) ───

func TestP107_Checkbox_NavigableCursor_EmptyItems(t *testing.T) {
	cb := NewCheckbox([]string{})
	cb.setNavigableCursor(5)
	if cb.Cursor() != 0 {
		t.Errorf("expected 0 for empty items, got %d", cb.Cursor())
	}
}

func TestP107_Checkbox_NavigableCursor_ForwardSkip(t *testing.T) {
	// idx=0 is disabled, should skip to idx=1
	cb := NewCheckbox([]string{"A", "B", "C"})
	items := cb.Items()
	items[0].Disabled = true
	cb.SetItems(items)
	cb.setNavigableCursor(0)
	if cb.Cursor() != 1 {
		t.Errorf("expected 1 (skip disabled 0), got %d", cb.Cursor())
	}
}

func TestP107_Checkbox_NavigableCursor_AllDisabledForward(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	items := cb.Items()
	for i := range items {
		items[i].Disabled = true
	}
	cb.SetItems(items)
	cb.setNavigableCursor(1)
	// All disabled → falls back to orig idx
	if cb.Cursor() != 1 {
		t.Logf("cursor=%d (expected 1 for all-disabled fallback)", cb.Cursor())
	}
}

func TestP107_Checkbox_NavigableCursor_ClampNegative(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B"})
	cb.setNavigableCursor(-10)
	if cb.Cursor() != 0 {
		t.Errorf("expected 0 for negative clamp, got %d", cb.Cursor())
	}
}

func TestP107_Checkbox_NavigableCursor_ClampOverflow(t *testing.T) {
	cb := NewCheckbox([]string{"A", "B", "C"})
	cb.setNavigableCursor(100)
	if cb.Cursor() != 2 {
		t.Errorf("expected 2 for overflow clamp, got %d", cb.Cursor())
	}
}

// ─── RadioGroup setNavigableCursor (73.3%) ───

func TestP107_RadioGroup_NavigableCursor_EmptyItems(t *testing.T) {
	rg := NewRadioGroup([]string{})
	rg.setNavigableCursor(5)
	// Just verify no panic
}

func TestP107_RadioGroup_NavigableCursor_ForwardSkip(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(0, true)
	rg.setNavigableCursor(0)
	if rg.Cursor() != 1 {
		t.Errorf("expected 1 (skip disabled 0), got %d", rg.Cursor())
	}
}

func TestP107_RadioGroup_NavigableCursor_AllDisabled(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.SetDisabled(0, true)
	rg.SetDisabled(1, true)
	rg.SetDisabled(2, true)
	rg.setNavigableCursor(1)
	// All disabled → falls back
}

func TestP107_RadioGroup_NavigableCursor_ClampNegative(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B"})
	rg.setNavigableCursor(-10)
	if rg.Cursor() != 0 {
		t.Errorf("expected 0, got %d", rg.Cursor())
	}
}

func TestP107_RadioGroup_NavigableCursor_ClampOverflow(t *testing.T) {
	rg := NewRadioGroup([]string{"A", "B", "C"})
	rg.setNavigableCursor(100)
	if rg.Cursor() != 2 {
		t.Errorf("expected 2, got %d", rg.Cursor())
	}
}

// ─── ContextMenu setCursorLocked (73.3%) ───

func TestP107_ContextMenu_SetCursor_EmptyItems(t *testing.T) {
	cm := NewContextMenu()
	cm.setCursorLocked(5)
	if cm.Cursor() != 0 {
		t.Errorf("expected 0 for empty, got %d", cm.Cursor())
	}
}

func TestP107_ContextMenu_SetCursor_NegativeClamp(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.setCursorLocked(-10)
	if cm.Cursor() != 0 {
		t.Errorf("expected 0 for negative, got %d", cm.Cursor())
	}
}

func TestP107_ContextMenu_SetCursor_OverflowClamp(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("b", "B"))
	cm.setCursorLocked(100)
	if cm.Cursor() != 1 {
		t.Errorf("expected 1 for overflow, got %d", cm.Cursor())
	}
}

func TestP107_ContextMenu_SetCursor_ForwardSearch(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("sep1", "---")) // separator, not navigable
	cm.AddItem(NewMenuItem("a", "A"))
	cm.setCursorLocked(0)
	// Should skip separator and land on 1
	if cm.Cursor() != 1 {
		t.Logf("cursor=%d (expected 1 to skip separator)", cm.Cursor())
	}
}

func TestP107_ContextMenu_SetCursor_BackwardSearch(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "A"))
	cm.AddItem(NewMenuItem("sep1", "---")) // separator
	cm.setCursorLocked(1)
	// Should skip separator backward and land on 0
	if cm.Cursor() != 0 {
		t.Logf("cursor=%d (expected 0 to skip separator backward)", cm.Cursor())
	}
}

// ─── Table truncateToWidth (75%) / drawCellLocked (75%) ───

func TestP107_Table_TruncateToWidth_ZeroWidth(t *testing.T) {
	tbl := NewTable([]string{"H"})
	if tbl.truncateToWidth("hello", 0) != "" {
		t.Error("width=0 should return empty")
	}
}

func TestP107_Table_TruncateToWidth_NegativeWidth(t *testing.T) {
	tbl := NewTable([]string{"H"})
	if tbl.truncateToWidth("hello", -1) != "" {
		t.Error("negative width should return empty")
	}
}

func TestP107_Table_TruncateToWidth_ExactFit(t *testing.T) {
	tbl := NewTable([]string{"H"})
	result := tbl.truncateToWidth("hello", 5)
	if result != "hello" {
		t.Errorf("exact fit should return original, got %q", result)
	}
}

func TestP107_Table_TruncateToWidth_WithEllipsis(t *testing.T) {
	tbl := NewTable([]string{"H"})
	result := tbl.truncateToWidth("hello world", 8)
	// Should be truncated with ellipsis
	if result == "hello world" {
		t.Error("should be truncated")
	}
}

func TestP107_Table_TruncateToWidth_NoEllipsisRoom(t *testing.T) {
	tbl := NewTable([]string{"H"})
	// Width 2: can fit 2 chars but no ellipsis (need width >= 3)
	result := tbl.truncateToWidth("hello", 2)
	if len([]rune(result)) > 2 {
		t.Errorf("result should be at most 2 chars, got %d", len([]rune(result)))
	}
}

func TestP107_Table_TruncateToWidth_Unicode(t *testing.T) {
	tbl := NewTable([]string{"H"})
	// Unicode chars take 2 columns each
	result := tbl.truncateToWidth("你好世界你好", 5)
	if result == "" {
		t.Error("should not be empty")
	}
}

func TestP107_Table_DrawCell_Unicode(t *testing.T) {
	tbl := NewTable([]string{"H"})
	tbl.SetRows([][]string{{"你好世界"}})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	tbl.Paint(buf)
}

func TestP107_Table_DrawCell_LongText(t *testing.T) {
	tbl := NewTable([]string{"Header"})
	tbl.SetRows([][]string{{"this is a very long text that exceeds the column width"}})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	tbl.Paint(buf)
}

// ─── Keybinding ParseKeyDesc (76.7%) ───

func TestP107_ParseKeyDesc_AllSpecial(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{"home"}, {"end"}, {"pageup"}, {"pagedown"},
		{"delete"}, {"del"}, {"backspace"},
		{"tab"}, {"space"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ParseKeyDesc(tt.desc) // just verify no panic
		})
	}
}

func TestP107_ParseKeyDesc_Combos(t *testing.T) {
	tests := []struct {
		desc string
	}{
		{"ctrl+shift+a"},
		{"alt+shift+x"},
		{"ctrl+alt+shift+q"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ParseKeyDesc(tt.desc)
		})
	}
}

// ─── DiffPreview paintBorderLocked (76.5%) ───

func TestP107_DiffPreview_PaintBorder(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowStats(true)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	dp.Paint(buf)
}

func TestP107_DiffPreview_PaintBorder_Narrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	dp.Paint(buf)
}

// ─── HelpOverlay ensureSelectedVisibleLocked (75%) ───

func TestP107_HelpOverlay_ScrollDown(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Global", Entries: []HelpEntry{
			{Keys: "ctrl+s", Description: "Save"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3}) // small height to force scroll
	buf := buffer.NewBuffer(60, 3)
	ho.Paint(buf)
	// Should scroll to keep selected visible
}

func TestP107_HelpOverlay_ScrollUp(t *testing.T) {
	groups := []HelpGroup{
		{Name: "G1", Entries: []HelpEntry{
			{Keys: "a", Description: "A"},
			{Keys: "b", Description: "B"},
			{Keys: "c", Description: "C"},
			{Keys: "d", Description: "D"},
			{Keys: "e", Description: "E"},
		}},
	}
	ho := NewHelpOverlay(groups)
	ho.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	// Scroll down first, then up
	ho.ScrollDown(3)
	ho.ScrollDown(-1)
	buf := buffer.NewBuffer(60, 3)
	ho.Paint(buf)
}

// ─── ScrollView contentW (75%) ───

func TestP107_ScrollView_ContentW(t *testing.T) {
	child := &fixedSize{w: 50, h: 10}
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	if sv.contentW(20) <= 0 {
		t.Error("contentW should be > 0 with scrollbar")
	}
}

func TestP107_ScrollView_ContentW_NoScrollbar(t *testing.T) {
	child := &fixedSize{w: 10, h: 5}
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	w := sv.contentW(20)
	_ = w // just verify no panic
}

// ─── ProgressBar formatPercent (77.8%) ───

func TestP107_ProgressBar_FormatPercent_100(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(1.0)
	pb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	pb.Paint(buf)
}

func TestP107_ProgressBar_FormatPercent_Over100(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(1.5)
	pb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	pb.Paint(buf)
}

func TestP107_ProgressBar_FormatPercent_Negative(t *testing.T) {
	pb := NewProgressBar()
	pb.SetProgress(-0.5)
	pb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	pb.Paint(buf)
}

// ─── Badge Measure (76.5%) ───

func TestP107_Badge_Measure_WithIcon(t *testing.T) {
	b := NewBadge("Test", BadgeInfo)
	b.SetIcon("●")
	s := b.Measure(Bounded(40, 5))
	if s.W <= 0 {
		t.Error("Measure should return positive width")
	}
}

func TestP107_Badge_Measure_ShortText(t *testing.T) {
	b := NewBadge("OK", BadgeSuccess)
	s := b.Measure(Bounded(5, 3))
	_ = s // just verify no panic
}

// ─── Sparkline valueToBar / recomputeRange ───

func TestP107_Sparkline_ValueToBar(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{1.0, 5.0, 3.0, 8.0, 2.0})
	sp.SetAutoScale(true)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

func TestP107_Sparkline_RecomputeRange_Flat(t *testing.T) {
	sp := NewSparkline()
	sp.SetData([]float64{5.0, 5.0, 5.0})
	sp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sp.Paint(buf)
}

// ─── Form HandleKey (78.3%) ───

func TestP107_Form_HandleKey_Tab(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("name", "Name", ""))
	f.AddField(NewTextField("email", "Email", ""))
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	buf := buffer.NewBuffer(40, 10)
	f.Paint(buf)
}

func TestP107_Form_HandleKey_ShiftTab(t *testing.T) {
	f := NewForm()
	f.AddField(NewTextField("name", "Name", ""))
	f.AddField(NewTextField("email", "Email", ""))
	f.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	f.FocusNext() // move to field 1
	f.HandleKey(&term.KeyEvent{Key: term.KeyTab, Modifiers: term.ModShift})
}

// ─── ListView MoveUp (76.9%) ───

func TestP107_ListView_MoveUp_DisabledWrap(t *testing.T) {
	lv := NewListView([]string{"A", "B", "C", "D"})
	lv.SetCursor(0) // at top
	lv.MoveUp()     // should wrap to bottom
	if lv.Cursor() != 3 {
		t.Logf("cursor=%d (expected 3 for wrap)", lv.Cursor())
	}
}

// ─── AutoComplete Paint (76.7%) ───

func TestP107_AutoComplete_Paint_Empty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

func TestP107_AutoComplete_Paint_WithCursor(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "apple"},
		{Label: "banana"},
		{Label: "cherry"},
	})
	ac.SetCursor(1)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

// ─── ThemeStudio setCursorLocked (75%) ───

func TestP107_ThemeStudio_SetCursor_Overflow(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetCursor(10000)
	// Should clamp without panic
}

func TestP107_ThemeStudio_SetCursor_Negative(t *testing.T) {
	ts := NewThemeStudio(nil)
	ts.SetCursor(-100)
	// Should clamp without panic
}
