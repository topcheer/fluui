package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP361_Funnel_Create(t *testing.T) {
	f := NewFunnelChart()
	f.SetStages([]FunnelStage{
		{Label: "Visits", Value: 1000},
		{Label: "Sign-ups", Value: 500},
		{Label: "Active", Value: 200},
		{Label: "Paid", Value: 50},
	})
	if f.StageCount() != 4 {
		t.Errorf("count = %d", f.StageCount())
	}
}

func TestP361_Funnel_SetSlices(t *testing.T) {
	f := NewFunnelChart()
	f.SetStages([]FunnelStage{{Label: "A", Value: 10}})
	if f.StageCount() != 1 {
		t.Errorf("count = %d", f.StageCount())
	}
}

func TestP361_Funnel_Empty(t *testing.T) {
	f := NewFunnelChart()
	if f.StageCount() != 0 {
		t.Error("empty should have 0 stages")
	}
}

func TestP361_Funnel_Measure(t *testing.T) {
	f := NewFunnelChart()
	f.AddStage(FunnelStage{Label: "A", Value: 10})
	f.AddStage(FunnelStage{Label: "B", Value: 5})
	sz := f.Measure(Constraints{MaxWidth: 30, MaxHeight: 10})
	if sz.W > 30 {
		t.Errorf("W = %d, should be <= 30", sz.W)
	}
}

func TestP361_Funnel_Paint(t *testing.T) {
	f := NewFunnelChart()
	f.SetStages([]FunnelStage{
		{Label: "A", Value: 10},
		{Label: "B", Value: 5},
		{Label: "C", Value: 1},
	})
	f.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	f.Paint(buf)
}

func TestP361_Funnel_Style(t *testing.T) {
	f := NewFunnelChart()
	st := DefaultFunnelChartStyle()
	f.SetStyle(st)
	if f.Style().Label.Fg != st.Label.Fg {
		t.Error("style mismatch")
	}
}
