package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestBubbleChart_New_P445(t *testing.T) {
	bc := NewBubbleChart()
	if bc.BubbleCount() != 0 {
		t.Errorf("BubbleCount = %d, want 0", bc.BubbleCount())
	}
}

func TestBubbleChart_AddBubble_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.AddBubble(BubbleData{X: 10, Y: 20, Size: 5, Label: "A"})
	bc.AddBubble(BubbleData{X: 30, Y: 50, Size: 10, Label: "B"})
	if bc.BubbleCount() != 2 {
		t.Errorf("BubbleCount = %d, want 2", bc.BubbleCount())
	}
}

func TestBubbleChart_SetBubbles_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.SetBubbles([]BubbleData{
		{X: 1, Y: 2, Size: 3},
		{X: 4, Y: 5, Size: 6},
		{X: 7, Y: 8, Size: 9},
	})
	if bc.BubbleCount() != 3 {
		t.Errorf("BubbleCount = %d, want 3", bc.BubbleCount())
	}
}

func TestBubbleChart_Bubbles_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.AddBubble(BubbleData{X: 1, Y: 2, Size: 3, Label: "X"})
	bubs := bc.Bubbles()
	if len(bubs) != 1 || bubs[0].Label != "X" {
		t.Errorf("Bubbles mismatch: %v", bubs)
	}
}

func TestBubbleChart_AutoColor_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.AddBubble(BubbleData{X: 1, Y: 2, Size: 3})
	bubs := bc.Bubbles()
	if bubs[0].Color.Type == 0 {
		t.Error("color should be auto-assigned")
	}
}

func TestBubbleChart_Clear_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.AddBubble(BubbleData{X: 1, Y: 2, Size: 3})
	bc.Clear()
	if bc.BubbleCount() != 0 {
		t.Error("should have 0 bubbles after Clear")
	}
}

func TestBubbleChart_Style_P445(t *testing.T) {
	bc := NewBubbleChart()
	st := DefaultBubbleChartStyle()
	bc.SetStyle(st)
	if bc.Style().Bubble.Fg != st.Bubble.Fg {
		t.Error("style mismatch")
	}
}

func TestBubbleChart_Measure_P445(t *testing.T) {
	bc := NewBubbleChart()
	sz := bc.Measure(Constraints{})
	if sz.W < 10 || sz.H < 10 {
		t.Errorf("size too small: %v", sz)
	}
}

func TestBubbleChart_Paint_NoPanic_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.AddBubble(BubbleData{X: 10, Y: 20, Size: 5, Label: "A"})
	bc.AddBubble(BubbleData{X: 30, Y: 50, Size: 10, Label: "B"})
	bc.AddBubble(BubbleData{X: 50, Y: 80, Size: 15, Label: "C"})
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	bc.Paint(buf)
}

func TestBubbleChart_Paint_ZeroBounds_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	bc.Paint(buf)
}

func TestBubbleChart_Paint_Empty_P445(t *testing.T) {
	bc := NewBubbleChart()
	bc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	bc.Paint(buf)
}

func TestBubbleChart_Children_P445(t *testing.T) {
	if NewBubbleChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkBubbleChart_Paint_P445(b *testing.B) {
	bc := NewBubbleChart()
	for i := 0; i < 10; i++ {
		bc.AddBubble(BubbleData{X: float64(i * 5), Y: float64(i * 3), Size: float64(i + 1), Label: "x"})
	}
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bc.Paint(buf)
	}
}
