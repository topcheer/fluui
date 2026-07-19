package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP302_SetOnCloseTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("tab1", "Tab 1")
	tb.AddTab("tab2", "Tab 2")

	called := ""
	tb.SetOnCloseTab(func(id string) { called = id })
	if tb.OnCloseTab() == nil {
		t.Error("OnCloseTab should not be nil")
	}

	tb.CloseActive() // closes tab1 (active)
	if called != "tab1" {
		t.Errorf("callback got %q, want tab1", called)
	}
	if tb.TabCount() != 1 {
		t.Errorf("count = %d, want 1", tb.TabCount())
	}
}

func TestP302_OnCloseTab_Nil(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "T1")
	// CloseActive without callback should not panic
	tb.CloseActive()
	if tb.TabCount() != 0 {
		t.Errorf("count = %d, want 0", tb.TabCount())
	}
}

func TestP302_SetTabClosable(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "T1")
	// All tabs start closable
	tab := tb.Tabs()[0]
	if !tab.Closable {
		t.Error("tab should be closable by default")
	}
	// Disable close
	if !tb.SetTabClosable("t1", false) {
		t.Error("SetTabClosable should return true for existing tab")
	}
	tab2 := tb.Tabs()[0]
	if tab2.Closable {
		t.Error("tab should NOT be closable after SetTabClosable(false)")
	}
	// Re-enable
	tb.SetTabClosable("t1", true)
	if !tb.Tabs()[0].Closable {
		t.Error("tab should be closable after re-enable")
	}
}

func TestP302_SetTabClosable_NotFound(t *testing.T) {
	tb := NewTabBar()
	if tb.SetTabClosable("nonexistent", true) {
		t.Error("should return false for nonexistent tab")
	}
}

func TestP302_HandleCloseClick(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab1")
	tb.AddTab("t2", "Tab2")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	tb.Paint(buffer.NewBuffer(40, 1))

	called := ""
	tb.SetOnCloseTab(func(id string) { called = id })

	// Find the close button position for tab 0
	// Tab1 is 4 chars + 2 close (space + ✕) = 6, close ✕ is at X=5
	hit := tb.HandleCloseClick(5, 0)
	if !hit {
		t.Error("should hit close button at x=6")
	}
	if called != "t1" {
		t.Errorf("callback got %q, want t1", called)
	}
	if tb.TabCount() != 1 {
		t.Errorf("count = %d, want 1 after close", tb.TabCount())
	}
}

func TestP302_HandleCloseClick_Miss(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab1")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	// Click on the title area (not close button)
	if tb.HandleCloseClick(0, 0) {
		t.Error("should not hit close button on title")
	}
}

func TestP302_HandleCloseClick_NonClosable(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab1")
	tb.SetTabClosable("t1", false)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	// No close button exists — any click should miss
	if tb.HandleCloseClick(10, 0) {
		t.Error("should not hit close on non-closable tab")
	}
}

func TestP302_Paint_CloseButton(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab1")
	tb.AddTab("t2", "Tab2")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	tb.Paint(buf) // should render ✕ characters, not panic
}

func TestP302_Paint_NonClosableTab(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab1")
	tb.SetTabClosable("t1", false)
	tb.AddTab("t2", "Tab2")
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	tb.Paint(buf) // t1 has no close button, t2 does
}

func TestP302_IsCloseButton(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab1") // 4+2=6 wide, close at X=5
	tb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	idx, ok := tb.IsCloseButton(5, 0)
	if !ok || idx != 0 {
		t.Errorf("IsCloseButton(5,0) = (%d,%v), want (0,true)", idx, ok)
	}
	// Not on close button
	_, ok2 := tb.IsCloseButton(0, 0)
	if ok2 {
		t.Error("should not detect close button on title")
	}
}

func TestP302_CloseActive_CallbackOrder(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("a", "A")
	tb.AddTab("b", "B")
	tb.AddTab("c", "C")
	tb.SetActive(1) // active = b

	order := []string{}
	tb.SetOnCloseTab(func(id string) { order = append(order, id) })
	tb.CloseActive() // close "b"
	if len(order) != 1 || order[0] != "b" {
		t.Errorf("callback order = %v, want [b]", order)
	}
	if tb.ActiveIndex() != 1 {
		t.Errorf("active = %d, want 1", tb.ActiveIndex())
	}
}

func TestP302_CloseActive_Empty(t *testing.T) {
	tb := NewTabBar()
	called := false
	tb.SetOnCloseTab(func(id string) { called = true })
	tb.CloseActive() // should not panic
	if called {
		t.Error("callback should not fire on empty tabbar")
	}
}

func TestP302_ConcurrentCloseAndAccess(t *testing.T) {
	tb := NewTabBar()
	for i := 0; i < 10; i++ {
		tb.AddTab("tab", "T")
	}
	done := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ {
			tb.CloseActive()
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		_ = tb.TabCount()
		_ = tb.Tabs()
	}
	<-done
}
