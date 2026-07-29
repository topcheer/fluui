package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestContextWindowBarBasic(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetContextLimit(128000)
	cwb.SetUsed(95000)

	if cwb.ContextLimit() != 128000 {
		t.Errorf("ContextLimit = %d, want 128000", cwb.ContextLimit())
	}
	if cwb.Used() != 95000 {
		t.Errorf("Used = %d, want 95000", cwb.Used())
	}
	pct := cwb.UsagePercent()
	// 95000/128000 = ~74.2%
	if pct < 74.0 || pct > 75.0 {
		t.Errorf("UsagePercent = %f, want ~74.2", pct)
	}
}

func TestContextWindowBarPercentBounds(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetContextLimit(1000)
	cwb.SetUsed(2000) // over limit
	pct := cwb.UsagePercent()
	if pct != 100 {
		t.Errorf("UsagePercent over limit = %f, want 100", pct)
	}
}

func TestContextWindowBarZeroLimit(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetContextLimit(0)
	cwb.SetUsed(100)
	if cwb.UsagePercent() != 0 {
		t.Errorf("UsagePercent with zero limit = %f, want 0", cwb.UsagePercent())
	}
}

func TestContextWindowBarMeasure(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetBarWidth(30)
	s := cwb.Measure(Constraints{})
	if s.W < 30 {
		t.Errorf("W = %d, want >= 30", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestContextWindowBarPaint(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetContextLimit(10000)
	cwb.SetUsed(5000)
	cwb.SetBarWidth(20)
	cwb.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})

	buf := buffer.NewBuffer(50, 1)
	cwb.Paint(buf)

	// Should have some filled cells and some empty
	filledCount := 0
	for x := 4; x < 24; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '█' {
			filledCount++
		}
	}
	if filledCount != 10 {
		t.Errorf("filled cells = %d, want 10 (50%% of 20)", filledCount)
	}
}

func TestContextWindowBarThresholdColors(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetContextLimit(1000)
	cwb.SetBarWidth(10)
	cwb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})

	// Normal (< 60%)
	cwb.SetUsed(100)
	buf := buffer.NewBuffer(40, 1)
	cwb.Paint(buf)
	normalColor := buf.GetCell(4, 0).Fg // first filled cell

	// Critical (>= 85%)
	cwb.SetUsed(900)
	buf2 := buffer.NewBuffer(40, 1)
	cwb.Paint(buf2)
	criticalColor := buf2.GetCell(4, 0).Fg

	if normalColor.Equal(criticalColor) {
		t.Error("expected different colors for normal vs critical usage")
	}
}

func TestContextWindowBarChildren(t *testing.T) {
	cwb := NewContextWindowBar()
	if cwb.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		n   int
		got string
	}{
		{500, "500"},
		{8000, "8K"},
		{128000, "128K"},
		{2000000, "2M"},
	}
	for _, tt := range tests {
		got := formatTokenCount(tt.n)
		if got != tt.got {
			t.Errorf("formatTokenCount(%d) = %q, want %q", tt.n, got, tt.got)
		}
	}
}
