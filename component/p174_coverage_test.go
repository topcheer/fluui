package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP174_TextArea_Value(t *testing.T) {
	ta := NewTextArea()
	ta.SetValue("hello")
	if ta.Value() != "hello" {
		t.Errorf("expected 'hello', got %q", ta.Value())
	}
}

func TestP174_TextArea_Prompt(t *testing.T) {
	ta := NewTextArea()
	if ta.Prompt() != "" {
		t.Errorf("expected empty prompt, got %q", ta.Prompt())
	}
	ta.SetPrompt("> ") // no-op, should not panic
}

func TestP174_TextArea_Placeholder(t *testing.T) {
	ta := NewTextArea()
	if ta.Placeholder() != "" {
		t.Errorf("expected empty placeholder, got %q", ta.Placeholder())
	}
	ta.SetPlaceholder("hint") // no-op, should not panic
}

func TestP174_TextArea_FocusBlurBlink(t *testing.T) {
	ta := NewTextArea()
	ta.Focus() // no-op, should not panic
	ta.Blur()  // no-op, should not panic
	if ta.Blink() {
		t.Error("expected Blink to return false")
	}
}

func TestP174_TextArea_SetHeight(t *testing.T) {
	ta := NewTextArea()
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.SetHeight(10)
	if ta.Bounds().H != 10 {
		t.Errorf("expected height 10, got %d", ta.Bounds().H)
	}
}

func TestP174_TextArea_SetWidth(t *testing.T) {
	ta := NewTextArea()
	ta.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	ta.SetWidth(80)
	if ta.Bounds().W != 80 {
		t.Errorf("expected width 80, got %d", ta.Bounds().W)
	}
}

func TestP174_TextArea_LineColumn(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2")
	if ta.Line() != 0 || ta.Column() != 0 {
		t.Errorf("expected 0,0 got %d,%d", ta.Line(), ta.Column())
	}
}

func TestP174_TextArea_Reset(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello\nworld")
	ta.Reset()
	if ta.Value() != "" {
		t.Errorf("expected empty after reset, got %q", ta.Value())
	}
}

func TestP174_TextArea_CharLimit(t *testing.T) {
	ta := NewTextArea()
	if ta.CharLimit() != 0 {
		t.Errorf("expected 0, got %d", ta.CharLimit())
	}
	ta.SetCharLimit(100) // no-op
}

func TestP174_TextArea_CursorDownUp(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.CursorDown()
	if ta.Line() != 1 {
		t.Errorf("expected line 1, got %d", ta.Line())
	}
	ta.CursorUp()
	if ta.Line() != 0 {
		t.Errorf("expected line 0, got %d", ta.Line())
	}
}

// StyleBuilder 0% methods

func TestP174_StyleBuilder_BorderForeground(t *testing.T) {
	sb := NewStyle()
	sb.BorderForeground(buffer.NamedColor(buffer.NamedRed))
	// no-op method, should not panic
}

func TestP174_StyleBuilder_BorderBackground(t *testing.T) {
	sb := NewStyle()
	sb.BorderBackground(buffer.NamedColor(buffer.NamedBlue))
}

func TestP174_StyleBuilder_MaxWidth(t *testing.T) {
	sb := NewStyle()
	sb.MaxWidth(80)
}

func TestP174_StyleBuilder_MaxHeight(t *testing.T) {
	sb := NewStyle()
	sb.MaxHeight(24)
}

func TestP174_StyleBuilder_Margins(t *testing.T) {
	sb := NewStyle()
	sb.MarginLeft(2)
	sb.MarginRight(2)
	sb.MarginTop(1)
	sb.MarginBottom(1)
}

func TestP174_StyleBuilder_Paddings(t *testing.T) {
	sb := NewStyle()
	sb.PaddingLeft(1)
	sb.PaddingRight(1)
	sb.PaddingTop(0)
	sb.PaddingBottom(0)
}

func TestP174_StyleBuilder_Align(t *testing.T) {
	sb := NewStyle()
	sb.Align(Left)
	sb.Align(Center)
	sb.Align(Right)
}

func TestP174_StyleBuilder_TabWidth(t *testing.T) {
	sb := NewStyle()
	sb.TabWidth(4)
}

func TestP174_StyleBuilder_UnderlineSpaces(t *testing.T) {
	sb := NewStyle()
	sb.UnderlineSpaces(true)
	sb.UnderlineSpaces(false)
}

func TestP174_StyleBuilder_StrikethroughSpaces(t *testing.T) {
	sb := NewStyle()
	sb.StrikethroughSpaces(true)
	sb.StrikethroughSpaces(false)
}

func TestP174_StyleBuilder_hexToByte(t *testing.T) {
	// Test valid hex
	v := hexToByte("ff")
	if v != 255 {
		t.Errorf("expected 255, got %d", v)
	}
	v = hexToByte("00")
	if v != 0 {
		t.Errorf("expected 0, got %d", v)
	}
	v = hexToByte("a")
	if v != -1 {
		t.Errorf("expected -1 for single char, got %d", v)
	}
	// Invalid
	v = hexToByte("gg")
	if v != -1 {
		t.Error("expected -1 for invalid hex 'gg'")
	}
	// Empty
	v = hexToByte("")
	if v != -1 {
		t.Error("expected -1 for empty string")
	}
}

func TestP174_StyleBuilder_hexCharToVal(t *testing.T) {
	cases := map[byte]int{
		'0': 0, '9': 9, 'a': 10, 'f': 15, 'A': 10, 'F': 15,
	}
	for ch, expected := range cases {
		v := hexCharToVal(ch)
		if v != expected {
			t.Errorf("hexCharToVal('%c'): expected %d, got %d", ch, expected, v)
		}
	}
	// Invalid
	v := hexCharToVal('g')
	if v != -1 {
		t.Error("expected -1 for 'g'")
	}
}

func TestP174_StyleBuilder_parseLipglossColor(t *testing.T) {
	sb := NewStyle()
	// Named color
	sb.ForegroundColor("blue")
	// Hex color
	sb.ForegroundHex("#abcdef")
	// Invalid hex (too short)
	sb.ForegroundHex("#abc")
	// RGB shorthand
	sb.ForegroundHex("#f00")
	// Unknown named color
	sb.ForegroundColor("nonexistent")
	// All should not panic
}

// Digits and LoadingIndicator 0% methods

func TestP174_Digits_SetStyle(t *testing.T) {
	d := NewDigits("1")
	d.SetStyle(DefaultDigitsStyle())
	// Should not panic
	d.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	d.Paint(buffer.NewBuffer(5, 5))
}

func TestP174_Digits_HandleKey(t *testing.T) {
	d := NewDigits("1")
	// HandleKey takes interface{} — test it doesn't panic
	d.HandleKey(nil)
}

func TestP174_LoadingIndicator_SetStyle(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.SetStyle(DefaultLoadingIndicatorStyle())
}

func TestP174_LoadingIndicator_StartDoubleStop(t *testing.T) {
	l := NewLoadingIndicator("test")
	l.Start()
	l.Start() // double start should be no-op
	l.Stop()
}

func TestP174_LoadingIndicator_MeasureNarrow(t *testing.T) {
	l := NewLoadingIndicator("test")
	s := l.Measure(Constraints{MaxWidth: 5})
	if s.W > 5 {
		t.Errorf("expected width <= 5, got %d", s.W)
	}
}

// Remaining sub-80% functions

func TestP174_AutoComplete_PaintWithItems(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetItems([]CompletionItem{
		{Label: "test1", Description: "desc1", Category: "cat1"},
		{Label: "test2", Description: "desc2"},
	})
	ac.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	ac.SetQuery("test")
	buf := buffer.NewBuffer(30, 10)
	ac.Paint(buf)
}

func TestP174_Badge_MeasureWithIcon(t *testing.T) {
	b := NewBadge("OK", BadgeSuccess)
	b.SetIcon("✓")
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 50})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected non-zero size, got %dx%d", s.W, s.H)
	}
	// Narrow
	b2 := NewBadge("Hello", BadgeInfo)
	s2 := b2.Measure(Constraints{MaxWidth: 2, MaxHeight: 50})
	if s2.W > 2 {
		t.Errorf("expected width <= 2, got %d", s2.W)
	}
}

func TestP174_CodeBlock_StreamingCursor(t *testing.T) {
	cb := NewCodeBlock("go", "x := 1\ny := 2")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	cb.SetStreaming(true)
	cb.SetShowLineNumbers(true)
	cb.SetTitle("test.go")
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)

	// Not streaming
	cb2 := NewCodeBlock("go", "z := 3")
	cb2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	cb2.SetStreaming(false)
	buf2 := buffer.NewBuffer(20, 3)
	cb2.Paint(buf2)
}

func TestP174_DiffPreview_Setters(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	dp.SetShowStats(true)
	dp.SetDiff("a\nb\nc")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	dp.Paint(buffer.NewBuffer(30, 10))
}

func TestP174_Sparkline_AllSame(t *testing.T) {
	sl := NewSparkline()
	sl.SetData([]float64{5, 5, 5, 5})
	sl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	sl.Paint(buffer.NewBuffer(20, 3))
}

func TestP174_ScrollView_NarrowW(t *testing.T) {
	sv := NewScrollView(NewParagraph("test"))
	sv.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	sv.Paint(buffer.NewBuffer(1, 3))
}

func TestP174_RichLog_VisibleLinesWrapped(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.Info("short")
	rl.Info("a very long message that wraps when width is small")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 10})
	rl.Paint(buffer.NewBuffer(15, 10))
}

func TestP174_DiffPreview_PaintBorderTall(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("a\nb\nc\nd\ne\nf\ng\nh\ni\nj")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	dp.Paint(buffer.NewBuffer(30, 3))
}
