package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P29 coverage tests for component utility functions and getters.

// === Constraints ===

func TestP29_Constraints_Clamp(t *testing.T) {
	cs := Constraints{MinWidth: 100, MaxWidth: 50, MinHeight: 80, MaxHeight: 40}
	cs.Clamp()
	if cs.MinWidth != 50 {
		t.Errorf("expected MinWidth=50, got %d", cs.MinWidth)
	}
	if cs.MinHeight != 40 {
		t.Errorf("expected MinHeight=40, got %d", cs.MinHeight)
	}
}

func TestP29_Constraints_Clamp_NoOp(t *testing.T) {
	cs := Constraints{MinWidth: 10, MaxWidth: 50}
	cs.Clamp()
	if cs.MinWidth != 10 {
		t.Errorf("should not change valid constraints")
	}
}

func TestP29_Constraints_ClampWidth(t *testing.T) {
	cs := Constraints{MaxWidth: 50, MinWidth: 10}
	if cs.ClampWidth(100) != 50 {
		t.Error("should clamp to max")
	}
	if cs.ClampWidth(5) != 10 {
		t.Error("should clamp to min")
	}
	if cs.ClampWidth(30) != 30 {
		t.Error("should not change in-range value")
	}
}

func TestP29_Constraints_ClampHeight(t *testing.T) {
	cs := Constraints{MaxHeight: 40, MinHeight: 5}
	if cs.ClampHeight(100) != 40 {
		t.Error("should clamp to max")
	}
	if cs.ClampHeight(2) != 5 {
		t.Error("should clamp to min")
	}
	if cs.ClampHeight(20) != 20 {
		t.Error("should not change in-range value")
	}
}

func TestP29_Fixed(t *testing.T) {
	cs := Fixed(30, 10)
	if cs.MinWidth != 30 || cs.MaxWidth != 30 {
		t.Error("Fixed width should be exact")
	}
	if cs.MinHeight != 10 || cs.MaxHeight != 10 {
		t.Error("Fixed height should be exact")
	}
}

func TestP29_Unbounded(t *testing.T) {
	cs := Unbounded()
	if cs.HasWidth() {
		t.Error("unbounded should have no width")
	}
	if cs.HasHeight() {
		t.Error("unbounded should have no height")
	}
}

func TestP29_Bounded(t *testing.T) {
	cs := Bounded(80, 24)
	if cs.MaxWidth != 80 {
		t.Error("bounded max width should be 80")
	}
	if cs.MaxHeight != 24 {
		t.Error("bounded max height should be 24")
	}
}

// === AutoComplete Query getter ===

func TestP29_AutoComplete_Query(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetQuery("test")
	if ac.Query() != "test" {
		t.Errorf("expected 'test', got %q", ac.Query())
	}
}

// === CommandPalette extra methods ===

func TestP29_CommandPalette_SetCaseSensitive(t *testing.T) {
	cp := NewCommandPalette()
	cp.SetCaseSensitive(true)
}

func TestP29_CommandPalette_LastUpdate(t *testing.T) {
	cp := NewCommandPalette()
	_ = cp.LastUpdate()
}

func TestP29_max2(t *testing.T) {
	if max2(3, 5) != 5 {
		t.Error("max2(3,5) should be 5")
	}
	if max2(10, 2) != 10 {
		t.Error("max2(10,2) should be 10")
	}
}

func TestP29_min2(t *testing.T) {
	if min2(3, 5) != 3 {
		t.Error("min2(3,5) should be 3")
	}
	if min2(10, 2) != 2 {
		t.Error("min2(10,2) should be 2")
	}
}

// === Badge SizeName ===

func TestP29_Badge_SizeName(t *testing.T) {
	if SizeName(BadgeSizeSmall) != "small" {
		t.Errorf("expected 'small', got %q", SizeName(BadgeSizeSmall))
	}
	if SizeName(BadgeSizeLarge) != "large" {
		t.Errorf("expected 'large', got %q", SizeName(BadgeSizeLarge))
	}
}

// === GenerateID uniqueness ===

func TestP29_GenerateID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateID("test")
		if ids[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

// === TabBar methods ===

func TestP29_TabBar_Tabs(t *testing.T) {
	tb := NewTabBar()
	tb.AddTab("t1", "Tab 1")
	tb.AddTab("t2", "Tab 2")
	tabs := tb.Tabs()
	if len(tabs) != 2 {
		t.Errorf("expected 2 tabs, got %d", len(tabs))
	}
}

// === buffer helper ===

func TestP29_buffer_BlankCell(t *testing.T) {
	buf := buffer.NewBuffer(5, 5)
	buf.SetCell(0, 0, buffer.BlankCell)
	c := buf.GetCell(0, 0)
	if c.Rune != ' ' {
		t.Error("blank cell should be space")
	}
}
