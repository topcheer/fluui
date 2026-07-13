package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// P172: Coverage tests for sub-80% functions

func TestParseColorString_NamedColors(t *testing.T) {
	// Test parseColorString with various named colors
	cases := []string{"red", "blue", "green", "yellow", "white", "black", "cyan", "magenta"}
	for _, c := range cases {
		ac := AdaptiveColor{Light: c, Dark: c}
		_ = ac.Resolve() // should not panic
	}
}

func TestParseColorString_Invalid(t *testing.T) {
	ac := AdaptiveColor{Light: "notacolor", Dark: "alsonotacolor"}
	_ = ac.Resolve() // should handle invalid gracefully
}

func TestAutoComplete_PaintEmpty(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf) // should handle empty items
}

func TestAutoComplete_PaintScrollDown(t *testing.T) {
	ac := NewAutoComplete()
	items := []CompletionItem{}
	for i := 0; i < 20; i++ {
		items = append(items, CompletionItem{Label: string(rune('a' + i))})
	}
	ac.SetItems(items)
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ac.Paint(buf) // should handle scroll
}

func TestBadge_MeasureShortText(t *testing.T) {
	b := NewBadge("X", BadgeInfo)
	s := b.Measure(Bounded(3, 5))
	if s.W > 3 {
		t.Error("width should be clamped to 3")
	}
}

func TestBadge_MeasureWithIconNarrow(t *testing.T) {
	b := NewBadge("Test", BadgeSuccess)
	b.SetIcon("★")
	s := b.Measure(Bounded(2, 5))
	if s.W > 2 {
		t.Error("should clamp to 2")
	}
}

func TestCodeBlock_StreamingCursorTitleLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetStreaming(true)
	cb.SetTitle("main.go")
	cb.SetShowLineNumbers(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf) // should handle title + lineNumbers + streaming
}

func TestCodeBlock_StreamingCursorLongLine(t *testing.T) {
	cb := NewCodeBlock("go", "func main() { fmt.Println(\"hello world this is a long line\") }")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	cb.Paint(buf) // should handle long line with narrow width
}

func TestDiffPreview_SetShowLineNumbers_P172(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)
}

func TestDiffPreview_SetShowStats_P172(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestDiffPreview_PaintBorderNarrow_P172(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added line\n-removed line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	buf := buffer.NewBuffer(3, 5)
	dp.Paint(buf) // should handle narrow border
}

func TestGrid_AddItemZeroSpan(t *testing.T) {
	g := NewGrid()
	g.SetRows(5)
	g.SetColumns(5)
	g.AddItem(NewFill('A', buffer.Style{}), 0, 0, 0, 0) // zero span = 1
	if g.ItemCount() != 1 {
		t.Error("should add with default span 1")
	}
}

func TestPages_CurrentComponent(t *testing.T) {
	p := NewPages()
	f := NewFill('A', buffer.Style{})
	p.AddPage("a", f)
	if p.CurrentComponent() != f {
		t.Error("CurrentComponent should return active page")
	}
}

func TestPages_NextPageNotFound(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	p.current = "nonexistent"
	result := p.NextPage()
	// Should fall through and return first page
	if result != "a" {
		t.Errorf("expected 'a', got '%s'", result)
	}
}

func TestPages_PrevPageNotFound(t *testing.T) {
	p := NewPages()
	p.AddPage("a", NewFill('A', buffer.Style{}))
	p.current = "nonexistent"
	result := p.PrevPage()
	if result != "a" {
		t.Errorf("expected 'a', got '%s'", result)
	}
}

func TestRichLog_CountVisibleLinesShowTime(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(100)
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	for i := 0; i < 5; i++ {
		rl.Write(LogInfo, "test message that might wrap")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)
	rl.Paint(buf) // countVisibleLinesLocked should handle time+levels
}

func TestScrollView_ContentWTiny_P172(t *testing.T) {
	child := NewFill(' ', buffer.Style{})
	sv := NewScrollView(child)
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 10})
	// contentW with tiny width should not panic
}

func TestSelectionList_MoveCursorEmpty(t *testing.T) {
	sl := NewSelectionList([]string{})
	sl.moveCursor(1) // should not panic with empty list
	sl.moveCursor(-1)
}

func TestSelectionList_TruncateLabelShort(t *testing.T) {
	result := truncateLabel("Hi", 10)
	if result != "Hi" {
		t.Error("short label should not be truncated")
	}
}

func TestSelectionList_TruncateLabelExact(t *testing.T) {
	result := truncateLabel("ABCDE", 5)
	if result != "ABCDE" {
		t.Error("exact length should not be truncated")
	}
}

func TestSelectionList_TruncateLabelTooShort(t *testing.T) {
	result := truncateLabel("ABC", 2)
	if result != ".." {
		t.Errorf("expected '..', got '%s'", result)
	}
}

func TestSelectionList_TruncateLabelZero(t *testing.T) {
	result := truncateLabel("ABC", 0)
	if result != "" {
		t.Error("zero maxLen should return empty")
	}
}

func TestSessionSidebar_SetWidth(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetWidth(30)
}

func TestSessionSidebar_Measure(t *testing.T) {
	sb := NewSessionSidebar()
	s := sb.Measure(Bounded(50, 20))
	_ = s // just verify it doesn't crash
}

func TestSessionSidebar_SelectedItem(t *testing.T) {
	sb := NewSessionSidebar()
	// With no items, SelectedItem should handle gracefully
	_ = sb.SelectedItem()
}

func TestSessionSidebar_HandleKey(t *testing.T) {
	sb := NewSessionSidebar()
	sb.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	sb.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	sb.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	sb.HandleKey(&term.KeyEvent{Rune: 'j'})
	sb.HandleKey(&term.KeyEvent{Rune: 'k'})
}

func TestSessionSidebar_HandleMouse(t *testing.T) {
	sb := NewSessionSidebar()
	sb.HandleMouse(&term.MouseEvent{X: 5, Y: 10, Action: term.MouseDown})
}

func TestSessionSidebar_SetGroupExpanded(t *testing.T) {
	sb := NewSessionSidebar()
	sb.SetGroupExpanded("test-group", true)
	sb.SetGroupExpanded("test-group", false)
}

func TestSparkline_ValueToBarAllSame(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sl.Paint(buf) // all same values should not crash
}

func TestSparkline_ValueToBarZeroNegative(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{0, -1, -2, 3, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	sl.Paint(buf) // zero and negative should not crash
}

func TestStyleBuilder_Underline(t *testing.T) {
	s := NewStyle().Underline()
	result := s.Render("test")
	if result == "" {
		t.Error("Underline should produce output")
	}
}

func TestStyleBuilder_UnderlineTwice(t *testing.T) {
	s := NewStyle().Underline().Underline()
	_ = s.Render("test") // calling twice should not crash
}

func TestComponent_Paint(t *testing.T) {
	c := BaseComponent{}
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	c.Paint(buf) // should be no-op
}