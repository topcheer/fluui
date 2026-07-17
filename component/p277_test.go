package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestSplitPane_ComputeDividerPos_ZeroAvail_P277(t *testing.T) {
	sp := NewSplitPane(&BaseComponent{}, &BaseComponent{})
	sp.SetDirection(SplitHorizontal)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 5})
	sp.SetRatio(0.5)
	buf := buffer.NewBuffer(1, 5)
	sp.Paint(buf)
}

func TestSplitPane_ComputeDividerPos_Vertical_P277(t *testing.T) {
	sp := NewSplitPane(&BaseComponent{}, &BaseComponent{})
	sp.SetDirection(SplitVertical)
	sp.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	sp.SetRatio(0.5)
	buf := buffer.NewBuffer(10, 1)
	sp.Paint(buf)
}

func TestStatusBar_SetItemText_InvalidID_P277(t *testing.T) {
	sb := NewStatusBar()
	sb.AddItem(StatusItem{ID: "left", Text: "hello", Align: StatusAlignLeft})
	sb.SetItemText("nonexistent", "updated")
}

func TestStatusBar_SetItemText_LeftAlign_P277(t *testing.T) {
	sb := NewStatusBar()
	sb.AddItem(StatusItem{ID: "left", Text: "hello", Align: StatusAlignLeft})
	sb.SetItemText("left", "updated")
}

func TestStatusBar_SetItemText_RightAlign_P277(t *testing.T) {
	sb := NewStatusBar()
	sb.AddItem(StatusItem{ID: "right", Text: "val", Align: StatusAlignRight})
	sb.SetItemText("right", "updated2")
}

func TestStatusBar_Measure_NoItems_P277(t *testing.T) {
	sb := NewStatusBar()
	s := sb.Measure(Constraints{MaxWidth: 100, MaxHeight: 50})
	if s.H < 1 {
		t.Error("statusbar should have positive height")
	}
}

func TestTabBar_CloseActive_P277(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.SetActive(1)
	tb.CloseActive()
	if tb.TabCount() != 2 {
		t.Errorf("expected 2 tabs, got %d", tb.TabCount())
	}
}

func TestTabBar_CloseActive_LastTab_P277(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.CloseActive()
	if tb.TabCount() != 0 {
		t.Errorf("expected 0 tabs, got %d", tb.TabCount())
	}
}

func TestTabBar_PrevTab_P277(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.SetActive(2)
	tb.PrevTab()
	if tb.ActiveIndex() != 1 {
		t.Errorf("expected active=1 after prev, got %d", tb.ActiveIndex())
	}
	tb.SetActive(0)
	tb.PrevTab()
	if tb.ActiveIndex() != 2 {
		t.Errorf("expected active=2 after wrap, got %d", tb.ActiveIndex())
	}
}

func TestTextArea_Text_Empty_P277(t *testing.T) {
	ta := NewTextArea()
	if ta.Text() != "" {
		t.Error("empty textarea should return empty string")
	}
}

func TestTextArea_Text_WithContent_P277(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("hello\nworld")
	if ta.Text() != "hello\nworld" {
		t.Errorf("expected 'hello\\nworld', got %q", ta.Text())
	}
}

func TestTextArea_Paint_WithContent_P277(t *testing.T) {
	ta := NewTextArea()
	ta.SetText("line1\nline2\nline3")
	ta.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	ta.Paint(buf)
}
