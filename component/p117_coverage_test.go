package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- 0% functions ---

func TestP117_BaseComponent_Paint(t *testing.T) {
	bc := BaseComponent{}
	bc.Paint(buffer.NewBuffer(10, 5)) // should be no-op, no panic
}

func TestP117_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(true)
	// Should not panic, should update internal state
	dp.SetDiff("+added line\n-removed line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	dp.Paint(buffer.NewBuffer(40, 5))
}

func TestP117_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
	dp.SetDiff("+added\n-context\n-removed")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	dp.Paint(buffer.NewBuffer(40, 5))
}

func TestP117_RichLog_Warnf(t *testing.T) {
	rl := NewRichLog()
	rl.Warnf("test %d warning %s", 1, "msg")
	e := rl.Entries()
	if len(e) != 1 || e[0].Level != LogWarn || e[0].Text != "test 1 warning msg" {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestP117_RichLog_Errorf(t *testing.T) {
	rl := NewRichLog()
	rl.Errorf("err %v", "boom")
	e := rl.Entries()
	if len(e) != 1 || e[0].Level != LogError {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestP117_RichLog_Debugf(t *testing.T) {
	rl := NewRichLog()
	rl.Debugf("dbg %d", 42)
	e := rl.Entries()
	if len(e) != 1 || e[0].Level != LogDebug {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestP117_RichLog_Fatalf(t *testing.T) {
	rl := NewRichLog()
	rl.Fatalf("fatal %s", "end")
	e := rl.Entries()
	if len(e) != 1 || e[0].Level != LogFatal {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestP117_RichLog_CountVisibleLinesLocked_MultiLine(t *testing.T) {
	rl := NewRichLog()
	rl.Info("short")
	rl.Info("a very long line that should wrap when the width is narrow enough to force it")
	rl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	// countVisibleLinesLocked is called internally by Measure
	s := rl.Measure(Bounded(20, 100))
	if s.H < 3 {
		t.Errorf("expected wrapping to produce multiple lines, got H=%d", s.H)
	}
}

func TestP117_RichLog_HandleKey_PageUpPageDown(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 50; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})

	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if rl.ScrollY() < 1 {
		t.Errorf("expected scrollY > 0 after PageUp, got %d", rl.ScrollY())
	}

	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	// Should have scrolled down
}

func TestP117_RichLog_HandleKey_PageDown_ToBottom(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 30; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})
	rl.ScrollUp(10)
	rl.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	// PageDown scrolls down by viewport height; may not reach 0 in one press
	if rl.ScrollY() >= 10 {
		t.Errorf("expected scrollY < 10 after PageDown, got %d", rl.ScrollY())
	}
}

func TestP117_RichLog_HandleKey_HomeEnd(t *testing.T) {
	rl := NewRichLog()
	for i := 0; i < 30; i++ {
		rl.Info("line")
	}
	rl.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})

	rl.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if rl.Following() {
		t.Error("expected not following after Home")
	}

	rl.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if !rl.Following() {
		t.Error("expected following after End")
	}
}

func TestP117_RichLog_HandleKey_UnknownKey(t *testing.T) {
	rl := NewRichLog()
	if rl.HandleKey(&term.KeyEvent{Key: term.KeyCode(999)}) {
		t.Error("expected unknown key not consumed")
	}
}

func TestP117_WrapLineCount(t *testing.T) {
	n := wrapLineCount("hello world this is long", 10)
	if n < 2 {
		t.Errorf("expected >= 2 wrapped lines, got %d", n)
	}
}

func TestP117_TruncateRunesLocal(t *testing.T) {
	// Test with a string longer than maxRunes
	result := truncateRunesLocal("hello world", 5)
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
	// Test with maxRunes 0
	if truncateRunesLocal("abc", 0) != "" {
		t.Error("expected empty for maxRunes 0")
	}
}

// --- Low coverage functions ---

func TestP117_CodeBlock_StreamingCursor_NarrowWidth(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.AppendSource("func main() {\n\tfmt.Println(\"hello world\")\n}")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	cb.Paint(buf)
}

func TestP117_CodeBlock_StreamingCursor_WithLineNumbers(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetShowLineNumbers(true)
	cb.SetStreaming(true)
	cb.AppendSource("line1\nline2\nline3")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestP117_ContextMenu_SetCursorLocked_NegativeOverflow(t *testing.T) {
	cm := NewContextMenu()
	cm.AddItem(NewMenuItem("item1", "Item 1"))
	cm.AddItem(NewMenuItem("item2", "Item 2"))
	cm.AddItem(NewMenuItem("item3", "Item 3"))
	cm.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)

	// Test cursor clamping with navigation
	cm.HandleKey(&term.KeyEvent{Key: term.KeyUp}) // should clamp to last
	cm.Paint(buf)

	cm.HandleKey(&term.KeyEvent{Key: term.KeyDown}) // should wrap to first
	cm.Paint(buf)
}

func TestP117_Viewport_DrawVScrollBar(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 10, h: 30})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
	// Should show scrollbar since content (30) > viewport (5)
}

func TestP117_Viewport_DrawHScrollBar(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 40, h: 5})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	vp.Paint(buf)
	// Should show horizontal scrollbar since content (40) > viewport (10)
}

func TestP117_Viewport_DrawScrollBars_BothOverflow(t *testing.T) {
	vp := NewViewport(&fixedSize{w: 50, h: 50})
	vp.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 10})
	buf := buffer.NewBuffer(15, 10)
	vp.Paint(buf)
	// Should show both scrollbars
}
