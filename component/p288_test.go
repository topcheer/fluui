package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P288: codeblock streaming cursor + richlog countVisibleLines + badge overflow

func TestCodeBlock_StreamingCursor_EmptyLines_P288(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursor_WithTitle_P288(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetShowTitle(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursor_WithContent_P288(t *testing.T) {
	cb := NewCodeBlock("go", "func main() {\n\tfmt.Println(\"hello\")\n}")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursor_TinyBounds_P288(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingFinish_P288(t *testing.T) {
	cb := NewCodeBlock("go", "code")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
	cb.FinishStreaming()
	buf2 := buffer.NewBuffer(40, 5)
	cb.Paint(buf2)
}

func TestRichLog_CountVisibleLines_ZeroHeight_P288(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 0}) // H<=0 → return len(entries)
	rl.Info("line1")
	rl.Info("line2")
	buf := buffer.NewBuffer(40, 1)
	rl.Paint(buf)
}

func TestRichLog_CountVisibleLines_LongWrap_P288(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	rl.Info(strings.Repeat("x", 50)) // long line that wraps multiple times
	buf := buffer.NewBuffer(10, 10)
	rl.Paint(buf)
}

func TestRichLog_CountVisibleLines_MinLevelFilter_P288(t *testing.T) {
	rl := NewRichLog()
	rl.SetMinLevel(LogWarn) // filter out Info/Debug
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	rl.Debug("debug") // filtered
	rl.Info("info")   // filtered
	rl.Warn("warn")   // visible
	rl.Error("error") // visible
	buf := buffer.NewBuffer(40, 10)
	rl.Paint(buf)
}

func TestRichLog_CountVisibleLines_NarrowWidth_P288(t *testing.T) {
	rl := NewRichLog()
	rl.SetShowLevels(true) // increases hdrWidth → contentW shrinks
	rl.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 5})
	rl.Info("test message")
	buf := buffer.NewBuffer(5, 5)
	rl.Paint(buf)
}

func TestRichLog_Paint_TruncateText_P288(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	rl.Info("This is a very long message that should be truncated")
	buf := buffer.NewBuffer(8, 3)
	rl.Paint(buf)
}

func TestRichLog_AutoScroll_Follow_P288(t *testing.T) {
	rl := NewRichLog()
	rl.SetAutoScroll(true)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	for i := 0; i < 10; i++ {
		rl.Info("entry")
	}
	buf := buffer.NewBuffer(40, 3)
	rl.Paint(buf)
}

func TestBadge_Measure_ClampAll_P288(t *testing.T) {
	b := NewBadge("Hello World Title", BadgeInfo)
	// Both width and height clamped
	s := b.Measure(Constraints{MaxWidth: 3, MaxHeight: 1})
	if s.W > 3 || s.H > 1 {
		t.Errorf("should clamp to 3x1, got %dx%d", s.W, s.H)
	}
}
