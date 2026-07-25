package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P372: Coverage for 0% functions across multiple components

func TestP372_Breadcrumb_Items(t *testing.T) {
	b := NewBreadcrumb([]string{"home", "docs", "api"})
	items := b.Items()
	if len(items) != 3 {
		t.Fatalf("len(Items) = %d, want 3", len(items))
	}
	if items[0] != "home" || items[2] != "api" {
		t.Errorf("Items = %v", items)
	}
	// Verify it's a copy (mutating should not affect original)
	items[0] = "changed"
	original := b.Items()
	if original[0] != "home" {
		t.Error("Items() should return a copy")
	}
}

func TestP372_FunnelChart_Slices(t *testing.T) {
	fc := NewFunnelChart([]FunnelSlice{
		{Label: "Visit", Value: 1000},
		{Label: "Signup", Value: 500},
	})
	slices := fc.Slices()
	if len(slices) != 2 {
		t.Fatalf("len(Slices) = %d, want 2", len(slices))
	}
	if slices[0].Label != "Visit" {
		t.Errorf("Slices[0].Label = %q", slices[0].Label)
	}
	// Verify copy
	slices[0].Label = "modified"
	orig := fc.Slices()
	if orig[0].Label != "Visit" {
		t.Error("Slices() should return a copy")
	}
}

func TestP372_PieChart_Slices(t *testing.T) {
	pc := NewPieChart([]PieSlice{
		{Label: "A", Value: 30},
		{Label: "B", Value: 70},
	})
	slices := pc.Slices()
	if len(slices) != 2 {
		t.Fatalf("len(Slices) = %d, want 2", len(slices))
	}
	if slices[1].Value != 70 {
		t.Errorf("Slices[1].Value = %v", slices[1].Value)
	}
	// Verify copy
	slices[1].Value = 0
	orig := pc.Slices()
	if orig[1].Value != 70 {
		t.Error("Slices() should return a copy")
	}
}

func TestP372_RadarChart_Axes(t *testing.T) {
	rc := NewRadarChart([]RadarAxis{
		{Label: "Speed", Max: 100},
		{Label: "Power", Max: 100},
		{Label: "Range", Max: 100},
	})
	axes := rc.Axes()
	if len(axes) != 3 {
		t.Fatalf("len(Axes) = %d, want 3", len(axes))
	}
	if axes[0].Label != "Speed" {
		t.Errorf("Axes[0].Label = %q", axes[0].Label)
	}
	// Verify copy
	axes[0].Label = "modified"
	orig := rc.Axes()
	if orig[0].Label != "Speed" {
		t.Error("Axes() should return a copy")
	}
}

func TestP372_DiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(false)
	// Currently always returns true (stub)
	if !dp.ShowLineNumbers() {
		t.Error("ShowLineNumbers should always return true (stub)")
	}
	dp.SetShowLineNumbers(true)
	if !dp.ShowLineNumbers() {
		t.Error("ShowLineNumbers should always return true (stub)")
	}
}

func TestP372_DiffPreview_SetShowStats(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)  // should not panic
	dp.SetShowStats(false) // should not panic
}

func TestP372_BaseComponent_Paint(t *testing.T) {
	// BaseComponent.Paint is a no-op default
	bc := &BaseComponent{}
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // should not panic, should do nothing
}

func TestP372_BaseComponent_Measure(t *testing.T) {
	// BaseComponent.Measure returns zero size by default
	bc := &BaseComponent{}
	s := bc.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.W != 0 || s.H != 0 {
		t.Errorf("BaseComponent.Measure = %v, want {0,0}", s)
	}
}

func TestP372_BaseComponent_Children(t *testing.T) {
	bc := &BaseComponent{}
	if children := bc.Children(); children != nil {
		t.Errorf("BaseComponent.Children = %v, want nil", children)
	}
}
