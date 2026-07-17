package component

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestHeader_Paint_WithSubtitle_P274(t *testing.T) {
	h := NewHeader("Title")
	h.SetSubtitle("Sub")
	h.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	h.Paint(buf)
}

func TestHeader_Paint_SubtitleOverflow_P274(t *testing.T) {
	h := NewHeader("T")
	h.SetSubtitle(strings.Repeat("X", 50))
	h.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	h.Paint(buf)
}

func TestListView_RemoveItem_InvalidIdx_P274(t *testing.T) {
	lv := NewListView([]string{"a", "b"})
	lv.RemoveItem(-1)
	lv.RemoveItem(99)
	if lv.ItemCount() != 2 {
		t.Errorf("expected 2 items, got %d", lv.ItemCount())
	}
}

func TestListView_RemoveItem_CursorAdjust_P274(t *testing.T) {
	lv := NewListView([]string{"a", "b", "c"})
	lv.SetCursor(2)
	lv.RemoveItem(2)
	if lv.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", lv.Cursor())
	}
}

func TestListView_RemoveItem_EmptyList_P274(t *testing.T) {
	lv := NewListView([]string{})
	lv.RemoveItem(0)
	if lv.ItemCount() != 0 {
		t.Error("removing from empty list should stay empty")
	}
}

func TestListView_PageUp_WithBounds_P274(t *testing.T) {
	lv := NewListView([]string{})
	items := make([]ListItem, 20)
	for i := range items {
		items[i] = ListItem{Label: "item"}
	}
	lv.SetItems(items)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	lv.SetCursor(10)
	lv.PageUp()
}

func TestListView_PageDown_WithBounds_P274(t *testing.T) {
	lv := NewListView([]string{})
	items := make([]ListItem, 20)
	for i := range items {
		items[i] = ListItem{Label: "item"}
	}
	lv.SetItems(items)
	lv.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	lv.PageDown()
}

func TestListView_PageUp_ZeroBounds_P274(t *testing.T) {
	lv := NewListView([]string{"a", "b"})
	lv.SetCursor(1)
	lv.PageUp()
}

func TestTabBar_InsertTab_P274(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Tab A")
	tb.AddTab("c", "Tab C")
	tb.InsertTab(1, "b", "Tab B")
	if tb.TabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tb.TabCount())
	}
}

func TestTabBar_InsertTab_AtStart_P274(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Tab A")
	tb.InsertTab(0, "b", "Tab B")
	if tb.TabCount() != 2 {
		t.Errorf("expected 2 tabs, got %d", tb.TabCount())
	}
}

func TestTabBar_InsertTab_ClampIdx_P274(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Tab A")
	tb.InsertTab(-5, "b", "Tab B")
	tb.InsertTab(100, "c", "Tab C")
	if tb.TabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tb.TabCount())
	}
}

func TestTooltip_Paint_PlainText_P274(t *testing.T) {
	tt := NewTooltip("line1\nline2\nline3")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	tt.Paint(buf)
}

func TestTooltip_Paint_TruncateLines_P274(t *testing.T) {
	longText := strings.Repeat("line\n", 10)
	tt := NewTooltip(longText)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	tt.Paint(buf)
}

func TestTooltip_ComputePosition_P274(t *testing.T) {
	tt := NewTooltip("tip")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	x, y := tt.ComputePosition(50, 30)
	if x < 0 || y < 0 {
		t.Errorf("position should be non-negative, got (%d, %d)", x, y)
	}
}
