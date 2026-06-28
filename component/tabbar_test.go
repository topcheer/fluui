package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewTabBar(t *testing.T) {
	tb := NewTabBar()
	if tb.TabCount() != 0 {
		t.Errorf("expected 0 tabs, got %d", tb.TabCount())
	}
	if tb.ActiveIndex() != 0 {
		t.Errorf("expected active 0")
	}
}

func TestTabBar_AddTab(t *testing.T) {
	tb := NewTabBar()
	idx := tb.AddTab("a", "Tab A")
	if idx != 0 {
		t.Errorf("expected idx 0, got %d", idx)
	}
	if tb.TabCount() != 1 {
		t.Errorf("expected 1 tab, got %d", tb.TabCount())
	}
	tb.AddTab("b", "Tab B")
	tb.AddTab("c", "Tab C")
	if tb.TabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tb.TabCount())
	}
}

func TestTabBar_AddTabSetsActive(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	if tb.ActiveIndex() != 0 {
		t.Errorf("first tab should be active")
	}
}

func TestTabBar_TabAt(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tab := tb.TabAt(0)
	if tab == nil || tab.ID != "a" {
		t.Error("expected tab a at index 0")
	}
	if tb.TabAt(5) != nil {
		t.Error("expected nil for out-of-range")
	}
	if tb.TabAt(-1) != nil {
		t.Error("expected nil for negative")
	}
}

func TestTabBar_FindTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tab := tb.FindTab("b")
	if tab == nil || tab.Title != "B" {
		t.Error("expected to find tab b")
	}
	if tb.FindTab("nonexistent") != nil {
		t.Error("expected nil for unknown id")
	}
}

func TestTabBar_RemoveTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.RemoveTab("b")
	if tb.TabCount() != 2 {
		t.Errorf("expected 2 tabs, got %d", tb.TabCount())
	}
	if tb.FindTab("b") != nil {
		t.Error("tab b should be gone")
	}
}

func TestTabBar_RemoveTab_First(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.SetActive(1)
	tb.RemoveTab("a")
	if tb.ActiveIndex() != 0 {
		t.Errorf("expected active 0 after removing first, got %d", tb.ActiveIndex())
	}
}

func TestTabBar_RemoveTab_Last(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.SetActive(1)
	tb.RemoveTab("b")
	if tb.ActiveIndex() != 0 {
		t.Errorf("expected active 0 after removing last, got %d", tb.ActiveIndex())
	}
}

func TestTabBar_RemoveTab_NotFound(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.RemoveTab("nonexistent")
	if tb.TabCount() != 1 {
		t.Errorf("expected 1 tab")
	}
}

func TestTabBar_RemoveAll(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.RemoveTab("a")
	tb.RemoveTab("b")
	if tb.TabCount() != 0 {
		t.Errorf("expected 0 tabs")
	}
	if tb.ActiveIndex() != 0 {
		t.Errorf("expected active 0 with no tabs")
	}
}

func TestTabBar_CloseActive(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.SetActive(1)
	tb.CloseActive()
	if tb.TabCount() != 1 {
		t.Errorf("expected 1 tab, got %d", tb.TabCount())
	}
	if tb.ActiveIndex() != 0 {
		t.Errorf("expected active 0, got %d", tb.ActiveIndex())
	}
}

func TestTabBar_CloseActive_Empty(t *testing.T) {
	tb := NewTabBar()
	tb.CloseActive()
	if tb.TabCount() != 0 {
		t.Error("expected 0 tabs")
	}
}

func TestTabBar_SetActive(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.SetActive(2)
	if tb.ActiveIndex() != 2 {
		t.Errorf("expected active 2, got %d", tb.ActiveIndex())
	}
}

func TestTabBar_SetActive_OutOfRange(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.SetActive(5)
	if tb.ActiveIndex() != 0 {
		t.Errorf("active should stay 0 for OOB")
	}
	tb.SetActive(-1)
	if tb.ActiveIndex() != 0 {
		t.Errorf("active should stay 0 for negative")
	}
}

func TestTabBar_SetActiveByID(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	if !tb.SetActiveByID("b") {
		t.Error("expected true for existing id")
	}
	if tb.ActiveIndex() != 1 {
		t.Errorf("expected active 1")
	}
	if tb.SetActiveByID("nonexistent") {
		t.Error("expected false for unknown id")
	}
}

func TestTabBar_NextTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.NextTab()
	if tb.ActiveIndex() != 1 {
		t.Errorf("expected 1, got %d", tb.ActiveIndex())
	}
	tb.NextTab()
	if tb.ActiveIndex() != 2 {
		t.Errorf("expected 2, got %d", tb.ActiveIndex())
	}
	tb.NextTab()
	if tb.ActiveIndex() != 0 {
		t.Errorf("expected wrap to 0, got %d", tb.ActiveIndex())
	}
}

func TestTabBar_PrevTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.PrevTab()
	if tb.ActiveIndex() != 2 {
		t.Errorf("expected wrap to 2, got %d", tb.ActiveIndex())
	}
}

func TestTabBar_NextTab_Empty(t *testing.T) {
	tb := NewTabBar()
	tb.NextTab()
	if tb.ActiveIndex() != 0 {
		t.Error("should be no-op with empty")
	}
}

func TestTabBar_Tabs_ReturnsCopy(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tabs := tb.Tabs()
	tabs[0].Title = "modified"
	if tb.FindTab("a").Title != "A" {
		t.Error("Tabs() should return a copy")
	}
}

func TestTabBar_InsertTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("c", "C")
	tb.InsertTab(1, "b", "B")
	if tb.TabCount() != 3 {
		t.Errorf("expected 3 tabs, got %d", tb.TabCount())
	}
	if tb.TabAt(1).ID != "b" {
		t.Errorf("expected tab b at index 1, got %s", tb.TabAt(1).ID)
	}
}

func TestTabBar_InsertTab_AtStart(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.InsertTab(0, "x", "X")
	if tb.TabAt(0).ID != "x" {
		t.Error("expected x at index 0")
	}
}

func TestTabBar_SetStyle(t *testing.T) {
	tb := NewTabBar()
	s := DefaultTabBarStyle()
	s.Normal = buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	tb.SetStyle(s)
	if tb.Style().Normal.Fg != s.Normal.Fg {
		t.Error("style mismatch")
	}
}

func TestTabBar_DefaultTabBarStyle(t *testing.T) {
	s := DefaultTabBarStyle()
	if !s.Active.HasFlag(buffer.Bold) {
		t.Error("active style should have bold")
	}
}

func TestTabBar_ShowNewButton(t *testing.T) {
	tb := NewTabBar()
	if !tb.ShowNewButton() {
		t.Error("should show by default")
	}
	tb.SetShowNewButton(false)
	if tb.ShowNewButton() {
		t.Error("should be hidden")
	}
}

func TestTabBar_MaxTitleWidth(t *testing.T) {
	tb := NewTabBar()
	tb.SetMaxTitleWidth(10)
	if tb.MaxTitleWidth() != 10 {
		t.Error("expected 10")
	}
}

func TestTabBar_Measure(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Hello")
	tb.AddTab("b", "World")
	sz := tb.Measure(Unbounded())
	if sz.H != 1 {
		t.Errorf("expected height 1, got %d", sz.H)
	}
	if sz.W <= 0 {
		t.Error("expected non-zero width")
	}
}

func TestTabBar_Measure_Empty(t *testing.T) {
	tb := NewTabBar()
	tb.SetShowNewButton(false)
	sz := tb.Measure(Unbounded())
	if sz.W != 0 {
		t.Errorf("expected width 0 for empty, got %d", sz.W)
	}
}

func TestTabBar_Paint_NoPanic(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := newTestBuffer(40, 1)
	tb.Paint(buf)
}

func TestTabBar_Paint_ZeroBounds(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := newTestBuffer(10, 10)
	tb.Paint(buf)
}

func TestTabBar_Paint_RendersTitles(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Hello")
	tb.AddTab("b", "World")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	tb.SetShowNewButton(false)
	buf := newTestBuffer(40, 1)
	tb.Paint(buf)
	got := cellRunes(buf, 0, 0, 5)
	if got != "Hello" {
		t.Errorf("expected Hello, got %q", got)
	}
}

func TestTabBar_Paint_TruncatesLongTitle(t *testing.T) {
	tb := NewTabBar()
	tb.SetMaxTitleWidth(5)
	tb.AddTab("a", "This is a very long title")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := newTestBuffer(40, 1)
	tb.Paint(buf)
	got := cellRunes(buf, 0, 0, 5)
	if len([]rune(got)) > 5 {
		t.Errorf("title should be truncated to 5, got %d", len([]rune(got)))
	}
}

func TestTabBar_Children(t *testing.T) {
	tb := NewTabBar()
	if tb.Children() != nil {
		t.Error("TabBar should have no children")
	}
}

func TestTabBar_SetBounds(t *testing.T) {
	tb := NewTabBar()
	r := Rect{X: 1, Y: 2, W: 30, H: 1}
	tb.SetBounds(r)
	b := tb.Bounds()
	if b.X != 1 || b.Y != 2 || b.W != 30 || b.H != 1 {
		t.Errorf("bounds mismatch: %+v", b)
	}
}

func TestTabBar_HitTest(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Hello")
	tb.AddTab("b", "World")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	idx := tb.HitTest(0, 0)
	if idx != 0 {
		t.Errorf("expected tab 0 at x=0, got %d", idx)
	}
}

func TestTabBar_HitTest_OutsideY(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Hello")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	if tb.HitTest(0, 5) != -1 {
		t.Error("expected -1 for outside Y")
	}
}

func TestTabBar_IsCloseButton(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "Hello")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	// close button is at title_len (5) + space (1) = position 6
	idx, ok := tb.IsCloseButton(6, 0)
	if !ok || idx != 0 {
		t.Errorf("expected close button at (6,0) for tab 0, got idx=%d ok=%v", idx, ok)
	}
}

func TestTabBar_String(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	s := tb.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestTabBar_TruncateTitle(t *testing.T) {
	short := truncateTabTitle("Hi", 10)
	if short != "Hi" {
		t.Errorf("expected Hi, got %s", short)
	}
	long := truncateTabTitle("This is very long", 5)
	if len([]rune(long)) != 5 {
		t.Errorf("expected 5 runes, got %d", len([]rune(long)))
	}
}

func TestTabBar_Concurrent(t *testing.T) {
	tb := NewTabBar()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				tb.AddTab("tab", "T")
				tb.NextTab()
				_ = tb.Tabs()
				_ = tb.ActiveIndex()
				if j%10 == 0 {
					tb.PrevTab()
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestTabBar_ConcurrentPaint(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := newTestBuffer(40, 1)
				tb.Paint(buf)
			}
		}()
	}
	wg.Wait()
}
