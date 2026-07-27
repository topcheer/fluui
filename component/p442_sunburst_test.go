package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestSunburstChart_New_P442(t *testing.T) {
	sc := NewSunburstChart()
	if sc.SegmentCount() != 0 {
		t.Errorf("SegmentCount = %d, want 0", sc.SegmentCount())
	}
	if sc.TotalValue() != 0 {
		t.Errorf("TotalValue = %v, want 0", sc.TotalValue())
	}
}

func TestSunburstChart_AddSegment_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.AddSegment(SunburstSegment{Label: "A", Value: 40})
	sc.AddSegment(SunburstSegment{Label: "B", Value: 60})
	if sc.SegmentCount() != 2 {
		t.Errorf("SegmentCount = %d, want 2", sc.SegmentCount())
	}
	if sc.TotalValue() != 100 {
		t.Errorf("TotalValue = %v, want 100", sc.TotalValue())
	}
}

func TestSunburstChart_AutoColor_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.AddSegment(SunburstSegment{Label: "A", Value: 10})
	segs := sc.Segments()
	if segs[0].Color.Type == 0 {
		t.Error("color should be auto-assigned")
	}
}

func TestSunburstChart_SetSegments_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.SetSegments([]SunburstSegment{
		{Label: "X", Value: 20},
		{Label: "Y", Value: 30},
		{Label: "Z", Value: 50},
	})
	if sc.SegmentCount() != 3 {
		t.Errorf("SegmentCount = %d, want 3", sc.SegmentCount())
	}
}

func TestSunburstChart_Clear_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.AddSegment(SunburstSegment{Label: "A", Value: 10})
	sc.Clear()
	if sc.SegmentCount() != 0 {
		t.Error("should have 0 segments after Clear")
	}
}

func TestSunburstChart_Style_P442(t *testing.T) {
	sc := NewSunburstChart()
	st := DefaultSunburstChartStyle()
	sc.SetStyle(st)
	if sc.Style().Label.Fg != st.Label.Fg {
		t.Error("style mismatch")
	}
}

func TestSunburstChart_Measure_P442(t *testing.T) {
	sc := NewSunburstChart()
	sz := sc.Measure(Constraints{})
	if sz.W < 10 || sz.H < 10 {
		t.Errorf("size too small: %v", sz)
	}
}

func TestSunburstChart_Paint_NoPanic_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.AddSegment(SunburstSegment{Label: "A", Value: 40})
	sc.AddSegment(SunburstSegment{Label: "B", Value: 30})
	sc.AddSegment(SunburstSegment{Label: "C", Value: 30})
	sc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 15})
	buf := buffer.NewBuffer(20, 15)
	sc.Paint(buf)
}

func TestSunburstChart_Paint_ZeroBounds_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	sc.Paint(buf)
}

func TestSunburstChart_Paint_Empty_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 15})
	buf := buffer.NewBuffer(20, 15)
	sc.Paint(buf) // no segments, no-op
}

func TestSunburstChart_Children_P442(t *testing.T) {
	if NewSunburstChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestSunburstChart_Segments_P442(t *testing.T) {
	sc := NewSunburstChart()
	sc.AddSegment(SunburstSegment{Label: "X", Value: 10})
	segs := sc.Segments()
	if len(segs) != 1 || segs[0].Label != "X" {
		t.Errorf("Segments mismatch: %v", segs)
	}
}

func BenchmarkSunburstChart_Paint_P442(b *testing.B) {
	sc := NewSunburstChart()
	sc.AddSegment(SunburstSegment{Label: "A", Value: 40})
	sc.AddSegment(SunburstSegment{Label: "B", Value: 30})
	sc.AddSegment(SunburstSegment{Label: "C", Value: 20})
	sc.AddSegment(SunburstSegment{Label: "D", Value: 10})
	sc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 15})
	buf := buffer.NewBuffer(20, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sc.Paint(buf)
	}
}
