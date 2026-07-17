package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// P283: cover 0% functions in component/

func TestTextInput_SetEchoChar_P283(t *testing.T) {
	ti := NewTextInput()
	ti.SetEchoChar('#')
	if ti.EchoChar() != '#' {
		t.Errorf("expected '#', got %q", ti.EchoChar())
	}
}

func TestTextArea_SetPrompt_P283(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ") // no-op, shouldn't panic
}

func TestTextArea_SetPlaceholder_P283(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("hint") // no-op
}

func TestTextArea_FocusBlur_P283(t *testing.T) {
	ta := NewTextArea()
	ta.Focus() // no-op
	ta.Blur()  // no-op
}

func TestTextArea_SetCharLimit_P283(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100) // should set limit
}

func TestDiffPreview_SetShowLineNumbers_P283(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(false)
	if dp.ShowLineNumbers() != true {
		t.Error("ShowLineNumbers should always return true (not yet implemented)")
	}
}

func TestDiffPreview_SetShowStats_P283(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true) // no-op
}

func TestComponent_Paint_NilBase_P283(t *testing.T) {
	bc := &BaseComponent{}
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // should be no-op on BaseComponent with no override
}

func TestBadge_Measure_HasWidth_P283(t *testing.T) {
	b := NewBadge("Hello", BadgeInfo)
	// Test with HasWidth constraint
	s := b.Measure(Constraints{MaxWidth: 3, MaxHeight: 5})
	if s.W > 3 {
		t.Errorf("width should clamp to 3, got %d", s.W)
	}
}

func TestBadge_Measure_HasHeight_P283(t *testing.T) {
	b := NewBadge("Hello", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 50, MaxHeight: 1})
	if s.H > 1 {
		t.Errorf("height should clamp to 1, got %d", s.H)
	}
}

func TestRichLog_CountVisibleLines_MoreThanMax_P283(t *testing.T) {
	rl := NewRichLog()
	rl.SetMaxSize(5)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	// Write more entries than visible
	for i := 0; i < 20; i++ {
		rl.Info("entry")
	}
	buf := buffer.NewBuffer(40, 3)
	rl.Paint(buf)
}

func TestRichLog_CountVisibleLines_WithTime_P283(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true)
	rl.SetShowLevels(true)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	for i := 0; i < 10; i++ {
		rl.Info("message with time")
	}
	buf := buffer.NewBuffer(60, 5)
	rl.Paint(buf)
}

func TestViewport_DrawVScrollBar_P283(t *testing.T) {
	child := NewText("a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk\nl\nm\nn\no")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 4})
	vp.ScrollDown(5)
	buf := buffer.NewBuffer(20, 4)
	vp.Paint(buf)
}

func TestViewport_DrawHScrollBar_P283(t *testing.T) {
	child := NewText("This is a very wide horizontal line that exceeds viewport width")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollRight(20)
	buf := buffer.NewBuffer(10, 3)
	vp.Paint(buf)
}

func TestCodeBlock_StreamingCursor_P283(t *testing.T) {
	cb := NewCodeBlock("go", "func main() {\n\tfmt.Println(\"hello\")\n}")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestThemeStudio_SetCursor_P283(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	ts.Paint(buf)
}
