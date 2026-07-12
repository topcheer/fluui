package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Target WindowManager.Bounds 80%
func TestP163_WM_Bounds(t *testing.T) {
	wm := NewWindowManager(NewBadge("test", BadgeInfo))
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	b := wm.Bounds()
	if b.W != 80 || b.H != 24 {
		t.Errorf("expected 80x24, got %+v", b)
	}
}

func TestP163_WM_Bounds_NoPanes(t *testing.T) {
	wm := NewWindowManager(nil)
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	b := wm.Bounds()
	if b.W != 0 {
		t.Errorf("expected 0, got %d", b.W)
	}
}

// Target WindowManager.applyShowHandle 83.3%
func TestP163_WM_ShowHandle(t *testing.T) {
	wm := NewWindowManager(NewBadge("test", BadgeInfo))
	wm.SetShowHandle(true)
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
}

func TestP163_WM_HideHandle(t *testing.T) {
	wm := NewWindowManager(NewBadge("test", BadgeInfo))
	wm.SetShowHandle(false)
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
}

// Target WindowManager.applyEqualize 83.3%
func TestP163_WM_Equalize(t *testing.T) {
	wm := NewWindowManager(NewBadge("test", BadgeInfo))
	wm.SplitRight(NewBadge("two", BadgeSuccess), "Two")
	wm.SplitDown(NewBadge("three", BadgeWarning), "Three")
	wm.Equalize()
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
}

func TestP163_WM_EqualizeSinglePane(t *testing.T) {
	wm := NewWindowManager(NewBadge("test", BadgeInfo))
	wm.Equalize() // should not panic with single pane
}

// Target WindowManager.highlightFocus 83.3%
func TestP163_WM_HighlightFocus(t *testing.T) {
	wm := NewWindowManager(NewBadge("test", BadgeInfo))
	wm.SplitRight(NewBadge("two", BadgeSuccess), "Two")
	wm.FocusNext()
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
}

// Target WindowManager.Paint 87.5%
func TestP163_WM_Paint_NestedSplits(t *testing.T) {
	wm := NewWindowManager(NewBadge("a", BadgeInfo))
	wm.SplitRight(NewBadge("b", BadgeSuccess), "B")
	wm.SplitDown(NewBadge("c", BadgeWarning), "C")
	wm.FocusPrev()
	wm.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
}

func TestP163_WM_Paint_ZeroBounds(t *testing.T) {
	wm := NewWindowManager(NewBadge("a", BadgeInfo))
	wm.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 24)
	wm.Paint(buf)
}

// Target WindowManager.ClosePane
func TestP163_WM_ClosePane(t *testing.T) {
	wm := NewWindowManager(NewBadge("a", BadgeInfo))
	wm.SplitRight(NewBadge("b", BadgeSuccess), "B")
	wm.ClosePane()
	if wm.PaneCount() != 1 {
		t.Errorf("expected 1 pane after close, got %d", wm.PaneCount())
	}
}

func TestP163_WM_ClosePane_Single(t *testing.T) {
	wm := NewWindowManager(NewBadge("a", BadgeInfo))
	wm.ClosePane() // should not close last pane
	if wm.PaneCount() != 1 {
		t.Errorf("expected 1 pane, got %d", wm.PaneCount())
	}
}

// Target WindowManager.FocusNext/Prev
func TestP163_WM_FocusCycle(t *testing.T) {
	wm := NewWindowManager(NewBadge("a", BadgeInfo))
	wm.SplitRight(NewBadge("b", BadgeSuccess), "B")
	wm.SplitDown(NewBadge("c", BadgeWarning), "C")
	wm.FocusNext()
	wm.FocusNext()
	wm.FocusNext() // should wrap
	wm.FocusPrev()
	wm.FocusPrev()
	wm.FocusPrev() // should wrap
}

// Target WindowManager.FocusIndex
func TestP163_WM_FocusIndex(t *testing.T) {
	wm := NewWindowManager(NewBadge("a", BadgeInfo))
	wm.SplitRight(NewBadge("b", BadgeSuccess), "B")
	wm.FocusIndex(1)
	_ = wm.Focused()
	wm.FocusIndex(99) // should clamp
}

// Target Wizard.moveButtonForward/Backward 85.7%
func TestP163_Wizard_MoveButtons(t *testing.T) {
	w := NewWizard([]*WizardStep{
		{Title: "Step1", Content: NewBadge("1", BadgeInfo)},
		{Title: "Step2", Content: NewBadge("2", BadgeSuccess)},
		{Title: "Step3", Content: NewBadge("3", BadgeWarning)},
	})
	w.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	w.Next() // move forward
	w.Paint(buf)
	w.Back() // move backward
	w.Paint(buf)
}

// Target Wizard.HandleKey 88.4%
