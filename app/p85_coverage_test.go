package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/hit"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func buf85(w, h int) *buffer.Buffer {
	return buffer.NewBuffer(w, h)
}

// ─── MouseHandler.Handle scrollbar lifecycle ───

func TestP85_MouseHandle_ScrollbarDownDragUp(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	// Add enough content to enable scrollbar
	for i := 0; i < 30; i++ {
		b := block.NewAssistantTextBlock("b85-" + itoaP85(i))
		b.AppendDelta("Line " + itoaP85(i))
		app.container.AddBlock(b)
	}
	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	barX := app.scrollView.ScrollbarColumn()

	if barX >= 0 {
		mh.Handle(&term.MouseEvent{X: barX, Y: 5, Action: term.MouseDown})
		mh.Handle(&term.MouseEvent{X: barX, Y: 10, Action: term.MouseDrag})
		mh.Handle(&term.MouseEvent{X: barX, Y: 10, Action: term.MouseUp})
	}
}

func TestP85_MouseHandle_DragOutsideColumn(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	for i := 0; i < 30; i++ {
		b := block.NewAssistantTextBlock("b85d-" + itoaP85(i))
		b.AppendDelta("Line " + itoaP85(i))
		app.container.AddBlock(b)
	}
	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	barX := app.scrollView.ScrollbarColumn()

	if barX >= 0 {
		mh.Handle(&term.MouseEvent{X: barX, Y: 5, Action: term.MouseDown})
		mh.Handle(&term.MouseEvent{X: 10, Y: 15, Action: term.MouseDrag})
		mh.Handle(&term.MouseEvent{X: 10, Y: 15, Action: term.MouseUp})
	}
}

func TestP85_MouseHandle_WheelUpAndDown(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	consumed := mh.Handle(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: term.MouseWheelUp,
	})
	if !consumed {
		t.Error("WheelUp should be consumed")
	}

	consumed = mh.Handle(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: term.MouseWheelDown,
	})
	if !consumed {
		t.Error("WheelDown should be consumed")
	}
}

func TestP85_MouseHandle_UnknownWheelButton(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	consumed := mh.Handle(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: 99,
	})
	if consumed {
		t.Error("unknown wheel should not be consumed")
	}
}

func TestP85_MouseHandle_ClickNoRegion(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	consumed := mh.Handle(&term.MouseEvent{
		X: 5, Y: 5, Action: term.MouseDown,
	})
	if consumed {
		t.Error("click with no region should not be consumed")
	}
}

func TestP85_MouseHandle_ClickWithCustomAction(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	called := false
	mh.tree.Add(hit.Region{
		BlockID: "custom-block",
		Bounds:  hit.Rect{X: 5, Y: 5, W: 10, H: 1},
		Action: hit.Action{
			Type: hit.ActionCustom,
			Fn:   func() { called = true },
		},
	})

	tb := block.NewThinkingBlock("custom-block")
	tb.AppendDelta("content")
	app.container.AddBlock(tb)

	consumed := mh.Handle(&term.MouseEvent{
		X: 8, Y: 5, Action: term.MouseDown,
	})
	if !consumed {
		t.Error("custom action click should be consumed")
	}
	if !called {
		t.Error("custom action function was not called")
	}
}

func TestP85_MouseHandle_ClickWithCustomActionNilFn(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	mh.tree.Add(hit.Region{
		BlockID: "nil-fn-block",
		Bounds:  hit.Rect{X: 5, Y: 5, W: 10, H: 1},
		Action: hit.Action{
			Type: hit.ActionCustom,
			Fn:   nil,
		},
	})

	tb := block.NewThinkingBlock("nil-fn-block")
	tb.AppendDelta("content")
	app.container.AddBlock(tb)

	consumed := mh.Handle(&term.MouseEvent{
		X: 8, Y: 5, Action: term.MouseDown,
	})
	if !consumed {
		t.Error("click with nil Fn should still be consumed")
	}
}

// ─── HandleClick direct tests ───

func TestP85_HandleClick_NoTree(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	result := mh.HandleClick(10, 10)
	if result {
		t.Error("HandleClick with no regions should return false")
	}
}

func TestP85_HandleClick_RegionFound(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	tb := block.NewThinkingBlock("tb-click")
	tb.AppendDelta("thinking")
	app.container.AddBlock(tb)

	mh.tree.Add(hit.Region{
		BlockID: "tb-click",
		Bounds:  hit.Rect{X: 0, Y: 0, W: 20, H: 1},
		Action:  hit.Action{Type: hit.ActionToggle},
	})

	result := mh.HandleClick(5, 0)
	if !result {
		t.Error("HandleClick on region should return true")
	}
}

func TestP85_HandleClick_RegionOutsideBounds(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	mh.tree.Add(hit.Region{
		BlockID: "test",
		Bounds:  hit.Rect{X: 0, Y: 0, W: 10, H: 1},
		Action:  hit.Action{Type: hit.ActionToggle},
	})

	result := mh.HandleClick(50, 20)
	if result {
		t.Error("HandleClick outside region should return false")
	}
}

// ─── RebuildRegions ───

func TestP85_RebuildRegions_ThinkingBlock(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	tb := block.NewThinkingBlock("rb-think")
	tb.AppendDelta("thinking content")
	app.container.AddBlock(tb)

	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	mh.RebuildRegions()
}

func TestP85_RebuildRegions_ToolResultBlock(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)

	tr := block.NewToolResultBlock("rb-result")
	tr.SetOutput("result output")
	app.container.AddBlock(tr)

	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	mh.RebuildRegions()
}

func TestP85_RebuildRegions_NoBlocks(t *testing.T) {
	app := NewChatApp(80, 24)
	mh := NewMouseHandler(app)
	mh.RebuildRegions()
}

// ─── ChatApp.HandleKey coverage ───

func TestP85_HandleKey_Escape(t *testing.T) {
	app := NewChatApp(80, 24)
	app.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
}

func TestP85_HandleKey_CtrlC(t *testing.T) {
	app := NewChatApp(80, 24)
	app.HandleKey(&term.KeyEvent{Rune: 'c', Modifiers: term.ModCtrl})
}

func TestP85_HandleKey_ArrowKeys(t *testing.T) {
	app := NewChatApp(80, 24)
	app.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	app.HandleKey(&term.KeyEvent{Key: term.KeyDown})
}

func TestP85_HandleKey_PageUpDown(t *testing.T) {
	app := NewChatApp(80, 24)
	app.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	app.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
}

func TestP85_HandleKey_HomeEnd(t *testing.T) {
	app := NewChatApp(80, 24)
	app.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	app.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
}

// ─── ChatApp.Render coverage ───

func TestP85_Render_WithContent(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := block.NewAssistantTextBlock("render-test")
	tb.AppendDelta("# Title\n\nSome content")
	app.container.AddBlock(tb)

	app.SetSize(80, 24)
	app.Render(buf85(80, 24))
}

// ─── scrollToBottomLocked ───

func TestP85_ScrollToBottomLocked(t *testing.T) {
	app := NewChatApp(80, 24)
	for i := 0; i < 10; i++ {
		b := block.NewAssistantTextBlock("sb-" + itoaP85(i))
		b.AppendDelta("content " + itoaP85(i))
		app.container.AddBlock(b)
	}
	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	app.mu.Lock()
	app.scrollToBottomLocked()
	app.mu.Unlock()
}

// ─── Theme coverage ───

func TestP85_ThemeName(t *testing.T) {
	app := NewChatApp(80, 24)
	_ = app.ThemeName()
}

func TestP85_ThemeToast(t *testing.T) {
	app := NewChatApp(80, 24)
	app.ThemeToast()
}

// ─── Session coverage ───

func TestP85_ActiveSession(t *testing.T) {
	app := NewChatApp(80, 24)
	_ = app.ActiveSession()
}

func TestP85_SessionCount(t *testing.T) {
	app := NewChatApp(80, 24)
	count := app.SessionCount()
	if count < 0 {
		t.Errorf("SessionCount = %d, want >= 0", count)
	}
}

// ─── IsStreaming ───

func TestP85_IsStreaming_Empty(t *testing.T) {
	app := NewChatApp(80, 24)
	if app.IsStreaming() {
		t.Error("IsStreaming should be false for empty app")
	}
}

// ─── HandleMouseP16 coverage ───

func TestP85_HandleMouseP16_StatusBarClick(t *testing.T) {
	app := NewChatApp(80, 24)
	sb := component.NewStatusBar()
	sb.AddLeft("status", "Test")
	app.SetStatusBar(sb)
	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	// Click in status bar area (bottom of screen)
	app.HandleMouse(&term.MouseEvent{
		X: 5, Y: 23, Action: term.MouseDown,
	})
}

func TestP85_HandleMouseP16_InputAreaClick(t *testing.T) {
	app := NewChatApp(80, 24)
	app.SetSize(80, 24)
	app.Render(buf85(80, 24))

	// Click in input area (bottom rows)
	app.HandleMouse(&term.MouseEvent{
		X: 5, Y: 22, Action: term.MouseDown,
	})
}

// helper
func itoaP85(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
