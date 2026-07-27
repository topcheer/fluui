package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestFunnelChart_New_P446(t *testing.T) {
	fc := NewFunnelChart()
	if fc.StageCount() != 0 {
		t.Errorf("StageCount = %d, want 0", fc.StageCount())
	}
}

func TestFunnelChart_AddStage_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "A", Value: 100})
	fc.AddStage(FunnelStage{Label: "B", Value: 50})
	if fc.StageCount() != 2 {
		t.Errorf("StageCount = %d, want 2", fc.StageCount())
	}
}

func TestFunnelChart_SetStages_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.SetStages([]FunnelStage{
		{Label: "X", Value: 10},
		{Label: "Y", Value: 5},
		{Label: "Z", Value: 1},
	})
	if fc.StageCount() != 3 {
		t.Errorf("StageCount = %d, want 3", fc.StageCount())
	}
}

func TestFunnelChart_Stages_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "X", Value: 10})
	stages := fc.Stages()
	if len(stages) != 1 || stages[0].Label != "X" {
		t.Errorf("Stages mismatch: %v", stages)
	}
}

func TestFunnelChart_AutoColor_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "X", Value: 10})
	stages := fc.Stages()
	if stages[0].Color.Type == 0 {
		t.Error("color should be auto-assigned")
	}
}

func TestFunnelChart_Clear_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "X", Value: 10})
	fc.Clear()
	if fc.StageCount() != 0 {
		t.Error("should have 0 stages after Clear")
	}
}

func TestFunnelChart_Style_P446(t *testing.T) {
	fc := NewFunnelChart()
	st := DefaultFunnelChartStyle()
	fc.SetStyle(st)
	if fc.Style().Label.Fg != st.Label.Fg {
		t.Error("style mismatch")
	}
}

func TestFunnelChart_Measure_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "A", Value: 10})
	fc.AddStage(FunnelStage{Label: "B", Value: 5})
	sz := fc.Measure(Constraints{})
	if sz.H < 3 {
		t.Errorf("H = %d, want >= 3", sz.H)
	}
}

func TestFunnelChart_Paint_NoPanic_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "Visitors", Value: 10000})
	fc.AddStage(FunnelStage{Label: "Signups", Value: 3000})
	fc.AddStage(FunnelStage{Label: "Purchased", Value: 800})
	fc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 12})
	buf := buffer.NewBuffer(50, 12)
	fc.Paint(buf)
}

func TestFunnelChart_Paint_ZeroBounds_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	fc.Paint(buf)
}

func TestFunnelChart_Paint_Empty_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	fc.Paint(buf)
}

func TestFunnelChart_Paint_TinyBounds_P446(t *testing.T) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "X", Value: 10})
	fc.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 2})
	buf := buffer.NewBuffer(5, 2)
	fc.Paint(buf)
}

func TestFunnelChart_Children_P446(t *testing.T) {
	if NewFunnelChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkFunnelChart_Paint_P446(b *testing.B) {
	fc := NewFunnelChart()
	fc.AddStage(FunnelStage{Label: "Visitors", Value: 10000})
	fc.AddStage(FunnelStage{Label: "Signups", Value: 3000})
	fc.AddStage(FunnelStage{Label: "Trials", Value: 1500})
	fc.AddStage(FunnelStage{Label: "Purchased", Value: 800})
	fc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 12})
	buf := buffer.NewBuffer(50, 12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fc.Paint(buf)
	}
}
