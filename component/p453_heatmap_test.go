package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestHeatmapGrid_New_P453(t *testing.T) {
	hg := NewHeatmapGrid(7, 20)
	if hg.Rows() != 7 || hg.Cols() != 20 {
		t.Errorf("dims = %dx%d, want 7x20", hg.Rows(), hg.Cols())
	}
	if hg.CellCount() != 140 {
		t.Errorf("CellCount = %d, want 140", hg.CellCount())
	}
}

func TestHeatmapGrid_SetGet_P453(t *testing.T) {
	hg := NewHeatmapGrid(5, 5)
	hg.Set(0, 0, 10)
	hg.Set(1, 1, 20)
	if hg.Get(0, 0) != 10 {
		t.Errorf("Get(0,0) = %v, want 10", hg.Get(0, 0))
	}
	if hg.Get(1, 1) != 20 {
		t.Errorf("Get(1,1) = %v, want 20", hg.Get(1, 1))
	}
	if hg.Get(2, 2) != 0 {
		t.Errorf("Get(2,2) = %v, want 0 (unset)", hg.Get(2, 2))
	}
}

func TestHeatmapGrid_SetOutOfBounds_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.Set(10, 10, 5) // out of bounds, should be ignored
	if hg.Get(10, 10) != 0 {
		t.Error("out-of-bounds Set should be ignored")
	}
}

func TestHeatmapGrid_SetMaxValue_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.SetMaxValue(50)
	if hg.MaxValue() != 50 {
		t.Errorf("MaxValue = %v", hg.MaxValue())
	}
}

func TestHeatmapGrid_AutoMaxValue_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.SetMaxValue(10)
	hg.Set(0, 0, 25) // should auto-expand max
	if hg.MaxValue() < 25 {
		t.Errorf("MaxValue = %v, should be >= 25", hg.MaxValue())
	}
}

func TestHeatmapGrid_FilledCount_P453(t *testing.T) {
	hg := NewHeatmapGrid(5, 5)
	hg.Set(0, 0, 5)
	hg.Set(1, 1, 10)
	hg.Set(2, 2, 0)
	if hg.FilledCount() != 2 {
		t.Errorf("FilledCount = %d, want 2", hg.FilledCount())
	}
}

func TestHeatmapGrid_TotalValue_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.Set(0, 0, 5)
	hg.Set(0, 1, 10)
	hg.Set(0, 2, 15)
	if hg.TotalValue() != 30 {
		t.Errorf("TotalValue = %v, want 30", hg.TotalValue())
	}
}

func TestHeatmapGrid_Clear_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.Set(0, 0, 5)
	hg.Clear()
	if hg.FilledCount() != 0 {
		t.Error("should have 0 filled after Clear")
	}
}

func TestHeatmapGrid_CellSize_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.SetCellSize(3, 2)
	sz := hg.Measure(Constraints{})
	if sz.W < 10 {
		t.Errorf("W = %d, too small", sz.W)
	}
}

func TestHeatmapGrid_Style_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	st := DefaultHeatmapGridStyle()
	hg.SetStyle(st)
}

func TestHeatmapGrid_Measure_P453(t *testing.T) {
	hg := NewHeatmapGrid(7, 20)
	sz := hg.Measure(Constraints{})
	if sz.H < 5 {
		t.Errorf("H = %d, too small", sz.H)
	}
}

func TestHeatmapGrid_Paint_NoPanic_P453(t *testing.T) {
	hg := NewHeatmapGrid(7, 20)
	for i := 0; i < 7; i++ {
		for j := 0; j < 20; j++ {
			if (i+j)%3 == 0 {
				hg.Set(i, j, float64(i*j+1))
			}
		}
	}
	hg.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	hg.Paint(buf)
}

func TestHeatmapGrid_Paint_ZeroBounds_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	hg.Paint(buf)
}

func TestHeatmapGrid_Paint_Empty_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	hg.Paint(buf) // no data set, should render level 0 cells
}

func TestHeatmapGrid_Legend_P453(t *testing.T) {
	hg := NewHeatmapGrid(3, 3)
	hg.SetMaxValue(40)
	legend := hg.Legend()
	if legend == "" {
		t.Error("Legend should not be empty")
	}
}

func TestHeatmapGrid_Children_P453(t *testing.T) {
	if NewHeatmapGrid(1, 1).Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestLevelForValue_P453(t *testing.T) {
	tests := []struct {
		value, max float64
		want       int
	}{
		{0, 10, 0},
		{1, 10, 1},
		{3, 10, 2},
		{5, 10, 2},
		{8, 10, 4},
		{10, 10, 4},
		{-1, 10, 0},
		{5, 0, 0},
	}
	for _, tc := range tests {
		got := levelForValue(tc.value, tc.max)
		if got != tc.want {
			t.Errorf("levelForValue(%v, %v) = %d, want %d", tc.value, tc.max, got, tc.want)
		}
	}
}

func BenchmarkHeatmapGrid_Paint_P453(b *testing.B) {
	hg := NewHeatmapGrid(7, 20)
	for i := 0; i < 7; i++ {
		for j := 0; j < 20; j++ {
			hg.Set(i, j, float64(i*j%30+1))
		}
	}
	hg.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hg.Paint(buf)
	}
}
