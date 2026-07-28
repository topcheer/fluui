package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestTokenMeter_New_P459(t *testing.T) {
	tm := NewTokenMeter(128000)
	if tm.Max() != 128000 {
		t.Errorf("Max = %d", tm.Max())
	}
	if tm.Used() != 0 {
		t.Errorf("Used = %d", tm.Used())
	}
	if tm.Remaining() != 128000 {
		t.Errorf("Remaining = %d", tm.Remaining())
	}
}

func TestTokenMeter_SetUsed_P459(t *testing.T) {
	tm := NewTokenMeter(100000)
	tm.SetUsed(35000)
	if tm.Used() != 35000 {
		t.Errorf("Used = %d", tm.Used())
	}
	if tm.Remaining() != 65000 {
		t.Errorf("Remaining = %d", tm.Remaining())
	}
}

func TestTokenMeter_Percent_P459(t *testing.T) {
	tm := NewTokenMeter(100000)
	tm.SetUsed(50000)
	if tm.Percent() < 49.9 || tm.Percent() > 50.1 {
		t.Errorf("Percent = %v, want ~50", tm.Percent())
	}
}

func TestTokenMeter_Levels_P459(t *testing.T) {
	tm := NewTokenMeter(100000)
	tm.SetUsed(20000)
	if tm.IsWarning() || tm.IsCritical() {
		t.Error("20% should be safe")
	}
	tm.SetUsed(60000)
	if !tm.IsWarning() {
		t.Error("60% should warn")
	}
	if tm.IsCritical() {
		t.Error("60% should not be critical")
	}
	tm.SetUsed(80000)
	if !tm.IsCritical() {
		t.Error("80% should be critical")
	}
}

func TestTokenMeter_RemainingOverflow_P459(t *testing.T) {
	tm := NewTokenMeter(1000)
	tm.SetUsed(1500)
	if tm.Remaining() != 0 {
		t.Errorf("Remaining = %d, want 0", tm.Remaining())
	}
}

func TestTokenMeter_SetMax_P459(t *testing.T) {
	tm := NewTokenMeter(1000)
	tm.SetMax(200000)
	if tm.Max() != 200000 {
		t.Errorf("Max = %d", tm.Max())
	}
}

func TestTokenMeter_ZeroMax_P459(t *testing.T) {
	tm := NewTokenMeter(0)
	if tm.Percent() != 0 {
		t.Errorf("Percent = %v, want 0", tm.Percent())
	}
}

func TestTokenMeter_Measure_P459(t *testing.T) {
	tm := NewTokenMeter(1000)
	sz := tm.Measure(Constraints{})
	if sz.H != 1 {
		t.Errorf("H = %d", sz.H)
	}
}

func TestTokenMeter_Paint_NoPanic_P459(t *testing.T) {
	tm := NewTokenMeter(128000)
	tm.SetUsed(45000)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	tm.Paint(buf)
}

func TestTokenMeter_Paint_Critical_P459(t *testing.T) {
	tm := NewTokenMeter(100000)
	tm.SetUsed(90000)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	tm.Paint(buf)
}

func TestTokenMeter_Paint_NoPct_P459(t *testing.T) {
	tm := NewTokenMeter(1000)
	tm.SetUsed(500)
	tm.SetShowPct(false)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	tm.Paint(buf)
}

func TestTokenMeter_Paint_ZeroBounds_P459(t *testing.T) {
	tm := NewTokenMeter(1000)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tm.Paint(buf)
}

func TestTokenMeter_Style_P459(t *testing.T) {
	tm := NewTokenMeter(1000)
	tm.SetStyle(DefaultTokenMeterStyle())
}

func TestTokenMeter_Children_P459(t *testing.T) {
	if NewTokenMeter(100).Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestFormatTokenK_P459(t *testing.T) {
	if formatTokenK(500) != "500" {
		t.Errorf("formatTokenK(500) = %q", formatTokenK(500))
	}
	if formatTokenK(45000) != "45K" {
		t.Errorf("formatTokenK(45000) = %q", formatTokenK(45000))
	}
	if formatTokenK(1500000) != "1.5M" {
		t.Errorf("formatTokenK(1500000) = %q", formatTokenK(1500000))
	}
}

func BenchmarkTokenMeter_Paint_P459(b *testing.B) {
	tm := NewTokenMeter(128000)
	tm.SetUsed(45000)
	tm.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.Paint(buf)
	}
}
