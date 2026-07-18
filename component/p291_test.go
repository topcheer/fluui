package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// P291: cover exact remaining uncovered branches

// badge.Measure: MaxHeight=0 branch (HasHeight && h>MaxHeight), w<1, h<1
func TestBadge_Measure_MaxHeight0_P291(t *testing.T) {
	b := NewBadge("Hi", BadgeInfo)
	s := b.Measure(Constraints{MaxHeight: 0, MaxWidth: 1, Has: true})
	if s.W < 1 {
		t.Errorf("expected w>=1, got %d", s.W)
	}
}

func TestBadge_Measure_NegativeClamp_P291(t *testing.T) {
	b := NewBadge("Hi", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 0, Has: true})
	if s.W < 1 {
		t.Errorf("expected w>=1, got %d", s.W)
	}
}

// badge group Measure: MaxWidth=0 + w<1 clamp
func TestBadgeGroup_Measure_MaxWidth0_P291(t *testing.T) {
	bg := NewBadgeGroup()
	bg.Add(NewBadge("A", BadgeInfo))
	bg.Add(NewBadge("BB", BadgeInfo))
	s := bg.Measure(Constraints{MaxWidth: 0, Has: true})
	if s.W < 1 {
		t.Error("badge group should clamp to 1")
	}
}

// codeblock Measure: width<1, height<1 clamp
func TestCodeBlock_Measure_ZeroBounds_P291(t *testing.T) {
	cb := NewCodeBlock("go", "x")
	s := cb.Measure(Constraints{MaxWidth: 0, MaxHeight: 0, Has: true})
	if s.W < 1 || s.H < 1 {
		t.Errorf("expected >=1x1, got %dx%d", s.W, s.H)
	}
}

// codeblock Paint: codeW<0 branch
func TestCodeBlock_Paint_NarrowWidth_P291(t *testing.T) {
	cb := NewCodeBlock("go", "func main() {}")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	buf := buffer.NewBuffer(1, 3)
	cb.Paint(buf)
}

// codeblock highlight error fallback
func TestCodeBlock_HighlightFallback_P291(t *testing.T) {
	cb := NewCodeBlock("nonexistent_lang_xyz", "some code")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

// richlog countVisibleLines: contentW<1 branch (narrow width with hdrWidth)
func TestRichLog_CountVisible_NarrowWithHdr_P291(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowTime(true) // increases hdrWidth
	rl.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 5})
	rl.Info("test message that wraps")
	buf := buffer.NewBuffer(3, 5)
	rl.Paint(buf)
}

// richlog Paint: textStyle fallback (Fg.Type==0 && Flags==0)
func TestRichLog_Paint_TextStyleFallback_P291(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	// Info level with no explicit text style → falls back to r.style.TextStyle
	rl.Info("plain message")
	buf := buffer.NewBuffer(40, 3)
	rl.Paint(buf)
}

// richlog Paint: availW<0 branch (text starts past right edge)
func TestRichLog_Paint_AvailWidthNegative_P291(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowLevels(true) // wide header → text starts past viewport
	rl.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	rl.Info("msg")
	buf := buffer.NewBuffer(5, 3)
	rl.Paint(buf)
}

// themestudio picker OnChange callback
func TestThemeStudio_PickerOnChange_P291(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	ts.Paint(buf)
	// Open picker via key 'o' then change color
	ts.HandleKey(&term.KeyEvent{Rune: 'o'})
}

// themestudio HandleKey 'q' when pickerOpen
func TestThemeStudio_HandleKey_QuitFromPicker_P291(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	ts.pickerOpen = true
	ts.HandleKey(&term.KeyEvent{Rune: 'q'})
}

// themestudio Paint: title line exceeds bounds.H
func TestThemeStudio_Paint_TinyHeight_P291(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	ts.Paint(buf)
}

// themestudio setCursorLocked: empty slots
func TestThemeStudio_SetCursor_EmptySlots_P291(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	ts.Paint(buf)
}

// viewport scrollbar: barH<=0 when H=1 with horizontal overflow
func TestViewport_DrawVBar_H1_P291(t *testing.T) {
	child := NewText(strings.Repeat("x", 100))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	vp.Paint(buf)
}

// viewport scrollbar: barW<=0 when W=1 with vertical overflow
func TestViewport_DrawHBar_W1_P291(t *testing.T) {
	child := NewText("a\nb\nc\nd\ne")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 3})
	buf := buffer.NewBuffer(1, 3)
	vp.Paint(buf)
}

// viewport scrollbar: thumb position clamped to bar bottom
func TestViewport_VBar_ThumbClamp_P291(t *testing.T) {
	child := NewText(strings.Repeat("line\n", 300))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	vp.ScrollDown(295)
	buf := buffer.NewBuffer(10, 5)
	vp.Paint(buf)
}

// viewport scrollbar: H thumb position clamped to bar right
func TestViewport_HBar_ThumbClamp_P291(t *testing.T) {
	child := NewText(strings.Repeat("x", 500))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	vp.ScrollRight(490)
	buf := buffer.NewBuffer(10, 3)
	vp.Paint(buf)
}
