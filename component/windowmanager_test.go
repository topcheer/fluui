package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Helper: create a simple test component with fixed size.
type testPane struct {
	BaseComponent
	label string
}

func newTestPane(label string) *testPane {
	tp := &testPane{label: label}
	tp.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	return tp
}

func (tp *testPane) Measure(cs Constraints) Size {
	return Size{W: 80, H: 24}
}

func (tp *testPane) Paint(buf *buffer.Buffer) {
	// No-op paint
}

// --- WindowManager tests ---

func TestWindowManager_New(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	if wm.PaneCount() != 1 {
		t.Errorf("expected 1 pane, got %d", wm.PaneCount())
	}
	if wm.Root() == nil {
		t.Error("root should not be nil")
	}
}

func TestWindowManager_Focus(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	wm.SplitRight(newTestPane("right"), "right")

	if wm.FocusedIndex() != 0 {
		t.Errorf("expected focus 0, got %d", wm.FocusedIndex())
	}

	wm.FocusNext()
	if wm.FocusedIndex() != 1 {
		t.Errorf("expected focus 1 after next, got %d", wm.FocusedIndex())
	}

	wm.FocusNext() // wraps around
	if wm.FocusedIndex() != 0 {
		t.Errorf("expected focus 0 after wrap, got %d", wm.FocusedIndex())
	}

	wm.FocusPrev() // wraps to last
	if wm.FocusedIndex() != 1 {
		t.Errorf("expected focus 1 after prev wrap, got %d", wm.FocusedIndex())
	}
}

func TestWindowManager_SplitRight(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	wm.SplitRight(newTestPane("p2"), "p2")

	if wm.PaneCount() != 2 {
		t.Errorf("expected 2 panes, got %d", wm.PaneCount())
	}
	if _, ok := wm.Root().(*SplitPane); !ok {
		t.Error("root should be SplitPane with 2 panes")
	}
}

func TestWindowManager_SplitDown(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	wm.SplitDown(newTestPane("p2"), "p2")
	wm.SplitDown(newTestPane("p3"), "p3")

	if wm.PaneCount() != 3 {
		t.Errorf("expected 3 panes, got %d", wm.PaneCount())
	}
}

func TestWindowManager_ClosePane(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	wm.SplitRight(newTestPane("p2"), "p2")

	closed := wm.ClosePane()
	if !closed {
		t.Error("should close focused pane")
	}
	if wm.PaneCount() != 1 {
		t.Errorf("expected 1 pane after close, got %d", wm.PaneCount())
	}
}

func TestWindowManager_CloseLastPane(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	closed := wm.ClosePane()
	if closed {
		t.Error("should not close the last pane")
	}
}

func TestWindowManager_FocusIndex(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")
	wm.SplitRight(newTestPane("p3"), "p3")

	wm.FocusIndex(2)
	if wm.FocusedIndex() != 2 {
		t.Errorf("expected focus 2, got %d", wm.FocusedIndex())
	}

	wm.FocusIndex(-1) // invalid, should not change
	if wm.FocusedIndex() != 2 {
		t.Errorf("expected focus still 2, got %d", wm.FocusedIndex())
	}

	wm.FocusIndex(99) // out of range
	if wm.FocusedIndex() != 2 {
		t.Errorf("expected focus still 2, got %d", wm.FocusedIndex())
	}
}

func TestWindowManager_Focused(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")

	f := wm.Focused()
	if f == nil {
		t.Fatal("Focused() should not be nil")
	}
	if f.Label != "main" {
		t.Errorf("expected 'main', got %q", f.Label)
	}
}

func TestWindowManager_Panes(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")

	panes := wm.Panes()
	if len(panes) != 2 {
		t.Errorf("expected 2 panes, got %d", len(panes))
	}

	// Verify it's a copy (modifying shouldn't affect internal state)
	panes[0] = nil
	if wm.Panes()[0] == nil {
		t.Error("Panes() should return a copy")
	}
}

func TestWindowManager_Measure(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	size := wm.Measure(Unbounded())
	if size.W == 0 {
		t.Error("Measure should return non-zero width")
	}
}

func TestWindowManager_Paint(t *testing.T) {
	wm := NewWindowManager(newTestPane("main"))
	wm.SplitRight(newTestPane("p2"), "p2")
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})

	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
	// Should not panic
}

func TestWindowManager_Equalize(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")

	// Should not panic
	wm.Equalize()
}

func TestWindowManager_SetShowHandle(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")

	// Should not panic
	wm.SetShowHandle(true)
	wm.SetShowHandle(false)
}

func TestWindowManager_SetBounds(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	bounds := Rect{X: 0, Y: 0, W: 100, H: 30}
	wm.SetBounds(bounds)
	b := wm.Bounds()
	if b.W != 100 || b.H != 30 {
		t.Errorf("expected 100x30, got %dx%d", b.W, b.H)
	}
}

func TestWindowManager_MultipleSplits(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")
	wm.SplitRight(newTestPane("p3"), "p3")
	wm.SplitRight(newTestPane("p4"), "p4")

	if wm.PaneCount() != 4 {
		t.Errorf("expected 4 panes, got %d", wm.PaneCount())
	}

	// Navigate through all panes
	for i := 0; i < 4; i++ {
		wm.FocusIndex(i)
		if wm.FocusedIndex() != i {
			t.Errorf("expected focus %d, got %d", i, wm.FocusedIndex())
		}
	}
}

func TestWindowManager_CloseMiddle(t *testing.T) {
	wm := NewWindowManager(newTestPane("p1"))
	wm.SplitRight(newTestPane("p2"), "p2")
	wm.SplitRight(newTestPane("p3"), "p3")

	wm.FocusIndex(1) // focus p2
	wm.ClosePane()

	if wm.PaneCount() != 2 {
		t.Errorf("expected 2 panes after close, got %d", wm.PaneCount())
	}
}

func TestWindowManager_FocusNextSingle(t *testing.T) {
	wm := NewWindowManager(newTestPane("only"))
	wm.FocusNext()
	if wm.FocusedIndex() != 0 {
		t.Error("single pane focus should stay at 0")
	}
}
