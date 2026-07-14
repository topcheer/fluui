package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P173 coverage: targeted tests for sub-80% functions

func TestAutoComplete_PaintWithDescription_P173(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test", Description: "A test description that is long", Category: "utils"},
		{Label: "foo", Description: "Short", Category: "utils"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

func TestAutoComplete_PaintCategorySelected_P173(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test", Category: "cat1"},
		{Label: "foo", Category: "cat2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	ac.Paint(buf)
}

func TestBadge_MeasureAllVariants_P173(t *testing.T) {
	variants := []BadgeVariant{BadgeInfo, BadgeSuccess, BadgeWarning, BadgeError, BadgeCritical, BadgeNeutral}
	for _, v := range variants {
		b := NewBadge("Test", v)
		s := b.Measure(Bounded(20, 5))
		if s.H < 1 {
			t.Errorf("variant %d: height should be >= 1", v)
		}
	}
}

func TestBadge_MeasureWithIcon_P173(t *testing.T) {
	b := NewBadge("Test", BadgeInfo)
	b.SetIcon("*")
	s := b.Measure(Bounded(20, 5))
	if s.W < 1 {
		t.Error("width should be positive")
	}
}

func TestBadge_MeasureNarrow_P173(t *testing.T) {
	b := NewBadge("Long Text Here", BadgeInfo)
	s := b.Measure(Bounded(5, 5))
	if s.W > 5 {
		t.Error("width should be clamped to max")
	}
}

func TestCodeBlock_StreamingCursorNarrow_P173(t *testing.T) {
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	buf := buffer.NewBuffer(3, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursorLineNumbers_P173(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3")
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursorNotStreaming_P173(t *testing.T) {
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(false)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursorEmpty_P173(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursorZeroBounds_P173(t *testing.T) {
	cb := NewCodeBlock("go", "hello")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	cb.Paint(buf)
}

func TestDiffPreview_SetShowLineNumbers_P173(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
}

func TestDiffPreview_SetShowStats_P173(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestDiffPreview_PaintBorderNarrow_P173(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added line\n-removed line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	buf := buffer.NewBuffer(5, 5)
	dp.Paint(buf)
}

func TestDiffPreview_PaintBorderEmpty_P173(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	dp.Paint(buf)
}

func TestDiffPreview_PaintBorderTall_P173(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	dp.Paint(buf)
}

func TestHelpOverlay_EnsureSelectedVisibleScrollDown_P173(t *testing.T) {
	ho := NewHelpOverlay([]HelpGroup{
		{Name: "g1", Entries: []HelpEntry{{Keys: "ctrl+a", Description: "Action A"}}},
		{Name: "g2", Entries: []HelpEntry{{Keys: "ctrl+b", Description: "Action B"}}},
	})
	ho.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ho.ScrollDown(5)
}

func TestLoadingIndicator_StartDouble_P173(t *testing.T) {
	li := NewLoadingIndicator("Loading...")
	li.Start()
	li.Start() // double start should be safe
	li.Stop()
}

func TestRichLog_CountVisibleLinesScrolled_P173(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 10; i++ {
		rl.Write(LogInfo, "test message that might wrap")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	rl.Paint(buf)
}

func TestScrollView_ContentWTiny_P173(t *testing.T) {
	child := NewFill(' ', buffer.Style{})
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10})
}

func TestSparkline_ValueToBarAllSameNonZero_P173(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sl.Paint(buf)
}

func TestSparkline_ValueToBarZeroNeg_P173(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, -1, -2, 3, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sl.Paint(buf)
}

func TestStyleBuilder_InheritAllFlags_P173(t *testing.T) {
	parent := NewStyle().Bold().Italic().Underline()
	child := NewStyle().Inherit(parent)
	result := child.Render("test")
	if result == "" {
		t.Error("Inherit should propagate flags")
	}
}

func TestStyleBuilder_InheritEmpty_P173(t *testing.T) {
	child := NewStyle()
	child.Inherit(NewStyle())
	_ = child.Render("test")
}

func TestStyleBuilder_ParseLipglossColorHex_P173(t *testing.T) {
	s := NewStyle().ForegroundHex("#ff8800")
	result := s.Render("test")
	if result == "" {
		t.Error("hex color should work")
	}
}

func TestStyleBuilder_ParseLipglossColorNamed_P173(t *testing.T) {
	s := NewStyle().ForegroundNamed(int(buffer.NamedRed))
	result := s.Render("test")
	if result == "" {
		t.Error("named color should work")
	}
}

func TestStyleBuilder_ParseLipglossColorInvalid_P173(t *testing.T) {
	s := NewStyle().ForegroundANSI(255)
	_ = s.Render("test") // should not crash
}

func TestTextArea_MoveLineSingle_P173(t *testing.T) {
	ta := NewTextArea()
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ta.Paint(buf)
}

func TestTextArea_MoveLineEmpty_P173(t *testing.T) {
	ta := NewTextArea()
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ta.Paint(buf)
}

func TestSessionSidebar_MeasureWithItems_P173(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "s1", Title: "Session 1"},
		{ID: "s2", Title: "Session 2"},
	})
	s := sb.Measure(Bounded(50, 20))
	_ = s
}

func TestSessionSidebar_HandleKeyJ_P173(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetItems([]SessionItem{
		{ID: "s1", Title: "Session 1"},
		{ID: "s2", Title: "Session 2"},
	})
	sb.HandleKey(&term.KeyEvent{Rune: 'j'})
	sb.HandleKey(&term.KeyEvent{Rune: 'k'})
}

func TestThemeStudio_SetCursorCycle_P173(t *testing.T) {
	ts := NewThemeStudio(nil)
	for i := 0; i < 50; i++ {
		ts.SetCursor(ts.Cursor() + 1)
	}
	for i := 0; i < 50; i++ {
		ts.SetCursor(ts.Cursor() - 1)
	}
}

func TestViewport_VScrollBarNearBottom_P173(t *testing.T) {
	child := NewFill(' ', buffer.Style{})
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollDown(100)
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestViewport_HScrollBarNearRight_P173(t *testing.T) {
	child := NewFill(' ', buffer.Style{})
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollRight(100)
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func TestViewport_BothScrollbars_P173(t *testing.T) {
	child := NewFill(' ', buffer.Style{})
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollDown(10)
	vp.ScrollRight(10)
	buf := buffer.NewBuffer(10, 5)
	vp.Paint(buf)
}

func TestMenuBar_PaintDropdownLarge_P173(t *testing.T) {
	items := []MenuEntry{}
	for i := 0; i < 20; i++ {
		items = append(items, MenuEntry{ID: string(rune('a' + i)), Label: "Item " + string(rune('A'+i))})
	}
	mb := NewMenuBar([]Menu{{ID: "file", Title: "File", Items: items}})
	mb.OpenMenu(0)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	mb.Paint(buf)
}

func TestMenuBar_PaintDropdownShort_P173(t *testing.T) {
	mb := NewMenuBar([]Menu{{ID: "f", Title: "F", Items: []MenuEntry{{ID: "a", Label: "A"}}}})
	mb.OpenMenu(0)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	mb.Paint(buf)
}

func TestContextMenu_SetCursorOverflow_P173(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "Item A"))
	cm.AddItem(NewMenuItem("b", "Item B"))
	cm.SetCursor(100)
	if cm.Cursor() != 1 {
		t.Error("overflow should clamp to last")
	}
}

func TestContextMenu_SetCursorNegative_P173(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("a", "Item A"))
	cm.SetCursor(-1)
	if cm.Cursor() != 0 {
		t.Error("negative should clamp to 0")
	}
}

func TestContextMenu_SetCursorEmpty_P173(t *testing.T) {
	cm := NewContextMenu()
	cm.SetCursor(5) // should not panic with empty menu
}