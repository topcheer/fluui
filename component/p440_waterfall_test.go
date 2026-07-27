package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestWaterfallChart_New_P440(t *testing.T) {
	wc := NewWaterfallChart()
	if wc.BarCount() != 0 {
		t.Errorf("BarCount = %d, want 0", wc.BarCount())
	}
}

func TestWaterfallChart_AddBar_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "Start", Value: 100, Type: WaterfallStart})
	wc.AddBar(WaterfallBar{Label: "Inc", Value: 30, Type: WaterfallPositive})
	wc.AddBar(WaterfallBar{Label: "Dec", Value: -20, Type: WaterfallNegative})
	if wc.BarCount() != 3 {
		t.Errorf("BarCount = %d, want 3", wc.BarCount())
	}
}

func TestWaterfallChart_SetBars_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.SetBars([]WaterfallBar{
		{Label: "A", Value: 50, Type: WaterfallStart},
		{Label: "B", Value: 10, Type: WaterfallPositive},
	})
	if wc.BarCount() != 2 {
		t.Errorf("BarCount = %d, want 2", wc.BarCount())
	}
}

func TestWaterfallChart_Bars_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "X", Value: 10, Type: WaterfallStart})
	bars := wc.Bars()
	if len(bars) != 1 || bars[0].Label != "X" {
		t.Errorf("Bars mismatch: %v", bars)
	}
}

func TestWaterfallChart_Clear_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "X", Value: 10, Type: WaterfallStart})
	wc.Clear()
	if wc.BarCount() != 0 {
		t.Error("should have 0 bars after Clear")
	}
}

func TestWaterfallChart_Style_P440(t *testing.T) {
	wc := NewWaterfallChart()
	st := DefaultWaterfallChartStyle()
	wc.SetStyle(st)
	if wc.Style().Positive.Fg != st.Positive.Fg {
		t.Error("style mismatch")
	}
}

func TestWaterfallChart_ComputeTotals_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "Start", Value: 100, Type: WaterfallStart})
	wc.AddBar(WaterfallBar{Label: "Inc", Value: 30, Type: WaterfallPositive})
	wc.AddBar(WaterfallBar{Label: "Dec", Value: -20, Type: WaterfallNegative})
	wc.AddBar(WaterfallBar{Label: "End", Value: 0, Type: WaterfallEnd})

	starts, running, maxAbs := wc.computeRunningTotals()
	if len(starts) != 4 || len(running) != 4 {
		t.Fatalf("expected 4 bars, got %d/%d", len(starts), len(running))
	}
	// Start bar: 0 → 100
	if starts[0] != 0 || running[0] != 100 {
		t.Errorf("bar 0: start=%v running=%v, want 0/100", starts[0], running[0])
	}
	// After Inc: 100 → 130
	if running[1] != 130 {
		t.Errorf("bar 1 running=%v, want 130", running[1])
	}
	// After Dec (value=-20): cumulative=110, bar top=130 (old), bottom=110 (new)
	if running[2] != 130 {
		t.Errorf("bar 2 running(top)=%v, want 130 (old cumulative)", running[2])
	}
	if maxAbs < 100 {
		t.Errorf("maxAbs=%v, want >= 100", maxAbs)
	}
}

func TestWaterfallChart_Measure_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "S", Value: 10, Type: WaterfallStart})
	sz := wc.Measure(Constraints{})
	if sz.H < 10 {
		t.Errorf("H = %d, want >= 10", sz.H)
	}
}

func TestWaterfallChart_Paint_NoPanic_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "Start", Value: 100, Type: WaterfallStart})
	wc.AddBar(WaterfallBar{Label: "Rev", Value: 40, Type: WaterfallPositive, })
	wc.AddBar(WaterfallBar{Label: "Cost", Value: -25, Type: WaterfallNegative})
	wc.AddBar(WaterfallBar{Label: "Tax", Value: -10, Type: WaterfallNegative})
	wc.AddBar(WaterfallBar{Label: "End", Value: 0, Type: WaterfallEnd})
	wc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	wc.Paint(buf)
}

func TestWaterfallChart_Paint_ZeroBounds_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	wc.Paint(buf)
}

func TestWaterfallChart_Paint_Empty_P440(t *testing.T) {
	wc := NewWaterfallChart()
	wc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	wc.Paint(buf) // no bars, should be no-op
}

func TestWaterfallChart_Children_P440(t *testing.T) {
	if NewWaterfallChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkWaterfallChart_Paint_P440(b *testing.B) {
	wc := NewWaterfallChart()
	wc.AddBar(WaterfallBar{Label: "Start", Value: 100, Type: WaterfallStart})
	wc.AddBar(WaterfallBar{Label: "Rev", Value: 40, Type: WaterfallPositive})
	wc.AddBar(WaterfallBar{Label: "Cost", Value: -25, Type: WaterfallNegative})
	wc.AddBar(WaterfallBar{Label: "End", Value: 0, Type: WaterfallEnd})
	wc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wc.Paint(buf)
	}
}
