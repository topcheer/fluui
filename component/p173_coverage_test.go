package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Test StyleBuilder flag methods with explicit false to cover else branches.

func TestP173_StyleBuilder_UnderlineFalse(t *testing.T) {
	sb := NewStyle().Underline(true).Underline(false)
	if sb.Style().Flags&buffer.Underline != 0 {
		t.Error("expected underline cleared after false")
	}
}

func TestP173_StyleBuilder_DimFalse(t *testing.T) {
	sb := NewStyle().Dim(true).Dim(false)
	if sb.Style().Flags&buffer.Dim != 0 {
		t.Error("expected dim cleared after false")
	}
}

func TestP173_StyleBuilder_BlinkFalse(t *testing.T) {
	sb := NewStyle().Blink(true).Blink(false)
	if sb.Style().Flags&buffer.Blink != 0 {
		t.Error("expected blink cleared after false")
	}
}

func TestP173_StyleBuilder_ReverseFalse(t *testing.T) {
	sb := NewStyle().Reverse(true).Reverse(false)
	if sb.Style().Flags&buffer.Reverse != 0 {
		t.Error("expected reverse cleared after false")
	}
}

func TestP173_StyleBuilder_StrikethroughFalse(t *testing.T) {
	sb := NewStyle().Strikethrough(true).Strikethrough(false)
	if sb.Style().Flags&buffer.Strikethrough != 0 {
		t.Error("expected strikethrough cleared after false")
	}
}

func TestP173_StyleBuilder_Inherit(t *testing.T) {
	parent := NewStyle().Bold(true).Foreground(buffer.NamedColor(buffer.NamedRed))
	child := NewStyle().Inherit(parent)
	if child.Style().Flags&buffer.Bold == 0 {
		t.Error("expected inherited bold")
	}
	// Inherit with no-op parent
	empty := NewStyle().Inherit(NewStyle())
	if empty.Style().Flags != 0 {
		t.Error("expected no flags from empty parent")
	}
}

func TestP173_StyleBuilder_parseLipglossColor(t *testing.T) {
	// Test edge cases
	sb := NewStyle()
	// Valid hex
	sb.ForegroundHex("#ff0000")
	if sb.Style().Fg.Type != buffer.ColorTrue {
		t.Error("expected RGB color for #ff0000")
	}
	// Named color
	sb2 := NewStyle()
	sb2.ForegroundColor("red")
	if sb2.Style().Fg.Type != buffer.ColorNamed {
		_ = sb2
		t.Error("expected named color for 'red'")
	}
	// Invalid color string — should be no-op or default
	sb3 := NewStyle()
	sb3.ForegroundColor("notacolor")
	// Should not panic, color stays as-is (none)
}

func TestP173_AutoComplete_Paint_EdgeCases(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	// Paint with no items
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf)
	// Should not panic

	// Add items and paint
	ac.SetItems([]CompletionItem{
		{Label: "test1", Description: "desc1"},
		{Label: "test2", Description: "desc2", Category: "cat"},
	})
	ac.Paint(buf)
}

func TestP173_Badge_Measure_EdgeCases(t *testing.T) {
	b := NewBadge("Hi", BadgeInfo)
	b.SetIcon("!")
	s := b.Measure(Constraints{MaxWidth: 100, MaxHeight: 100})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", s.W, s.H)
	}

	// Very narrow
	b2 := NewBadge("Hello World", BadgeSuccess)
	s2 := b2.Measure(Constraints{MaxWidth: 3, MaxHeight: 100})
	if s2.W > 3 {
		t.Errorf("expected width <= 3, got %d", s2.W)
	}
}

func TestP173_CodeBlock_StreamingCursor_EdgeCases(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetTitle("test.go")
	cb.AppendSource("line1\nline2")
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)

	// Empty content with title and line numbers
	cb2 := NewCodeBlock("go", "")
	cb2.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	cb2.SetStreaming(true)
	cb2.SetShowLineNumbers(true)
	cb2.SetTitle("x.go")
	buf2 := buffer.NewBuffer(5, 3)
	cb2.Paint(buf2)
}

func TestP173_DiffPreview_Setters(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowStats(true)
	// Should not panic
	dp.SetDiff("line1\nline2")
	buf := buffer.NewBuffer(40, 10)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	dp.Paint(buf)
}

func TestP173_Sparkline_ValueToBar(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{1.0, 0.5, 0.0, -1.0})
	buf := buffer.NewBuffer(20, 5)
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	sl.Paint(buf)

	// All same values
	sl2 := NewSparkline()
	sl2.SetData([]float64{5.0, 5.0, 5.0})
	buf2 := buffer.NewBuffer(20, 5)
	sl2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	sl2.Paint(buf2)
}

func TestP173_ScrollView_ContentW(t *testing.T) {
	sv := NewScrollView(NewParagraph("test content here"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	sv.Paint(buf)

	// Very narrow
	sv2 := NewScrollView(NewParagraph("abc"))
	sv2.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	buf2 := buffer.NewBuffer(1, 5)
	sv2.Paint(buf2)
}

func TestP173_RichLog_CountVisibleLines(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("short message")
	rl.Info("a very long message that should wrap across multiple lines when the width is small")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	rl.Paint(buf)
}

func TestP173_DiffPreview_PaintBorder_Narrow(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("a\nb\nc\nd\ne\nf\ng\nh")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	dp.Paint(buf)
}
