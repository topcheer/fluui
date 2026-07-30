package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StepDots Tests ───

func TestStepDotsBasic(t *testing.T) {
	sd := NewStepDots()
	sd.SetTotal(5)
	sd.SetCurrent(3)
	if c := sd.Current(); c != 3 {
		t.Errorf("Current = %d, want 3", c)
	}
}

func TestStepDotsZero(t *testing.T) {
	sd := NewStepDots()
	sd.SetCurrent(0)
	if c := sd.Current(); c != 0 {
		t.Errorf("Current = %d, want 0", c)
	}
}

func TestStepDotsComplete(t *testing.T) {
	sd := NewStepDots()
	sd.SetTotal(3)
	sd.SetCurrent(3)
	if c := sd.Current(); c != 3 {
		t.Errorf("Current = %d, want 3", c)
	}
}

func TestStepDotsNegative(t *testing.T) {
	sd := NewStepDots()
	sd.SetCurrent(-5)
	if c := sd.Current(); c != 0 {
		t.Errorf("Current = %d, want 0 (clamped)", c)
	}
}

func TestStepDotsTotalClamp(t *testing.T) {
	sd := NewStepDots()
	sd.SetTotal(0)
	if sd.total != 1 {
		t.Errorf("total = %d, want 1 (clamped)", sd.total)
	}
	sd.SetTotal(stepDotsMax + 10)
	if sd.total != stepDotsMax {
		t.Errorf("total = %d, want %d (clamped)", sd.total, stepDotsMax)
	}
}

func TestStepDotsConnected(t *testing.T) {
	sd := NewStepDots()
	sd.SetConnected(true)
	if !sd.connected {
		t.Error("Expected connected=true")
	}
}

func TestStepDotsPaint(t *testing.T) {
	sd := NewStepDots()
	sd.SetTotal(5)
	sd.SetCurrent(2)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	sd.Paint(buf)
	// Should have dot chars
	if r := buf.GetCell(0, 0).Rune; r != '●' {
		t.Errorf("First rune = %q, want '●' (done)", r)
	}
}

func TestStepDotsPaintConnected(t *testing.T) {
	sd := NewStepDots()
	sd.SetTotal(4)
	sd.SetCurrent(2)
	sd.SetConnected(true)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	sd.Paint(buf)
	// After first dot should have connector
	if r := buf.GetCell(1, 0).Rune; r != '━' {
		t.Errorf("Connector = %q, want '━'", r)
	}
}

func TestStepDotsChildren(t *testing.T) {
	sd := NewStepDots()
	if c := sd.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestStepDotsStyle(t *testing.T) {
	sd := NewStepDots()
	sd.SetStyle(StepDotsStyle{
		Done:      buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Current:   buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Pending:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Connector: buffer.Style{Fg: buffer.RGB(32, 32, 32)},
	})
	sd.SetTotal(5).SetCurrent(3)
	buf := buffer.NewBuffer(10, 1)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	sd.Paint(buf)
}

// ─── MiniCalendar Tests ───

func TestMiniCalendarBasic(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetMonth(2024, 3)
	mc.SetToday(15)
	if mc.month != 3 || mc.year != 2024 {
		t.Errorf("Expected March 2024, got %d/%d", mc.month, mc.year)
	}
}

func TestMiniCalendarClamp(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetMonth(2024, 0)
	if mc.month != 1 {
		t.Errorf("month = %d, want 1 (clamped)", mc.month)
	}
	mc.SetMonth(2024, 13)
	if mc.month != 12 {
		t.Errorf("month = %d, want 12 (clamped)", mc.month)
	}
}

func TestMiniCalendarFebLeap(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetMonth(2024, 2)
	if mc.daysInMonth != 29 {
		t.Errorf("Feb 2024 days = %d, want 29 (leap)", mc.daysInMonth)
	}
}

func TestMiniCalendarFebNonLeap(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetMonth(2023, 2)
	if mc.daysInMonth != 28 {
		t.Errorf("Feb 2023 days = %d, want 28", mc.daysInMonth)
	}
}

func TestMiniCalendarToday(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetMonth(2024, 6)
	mc.SetToday(0)
	if mc.today != 0 {
		t.Errorf("today = %d, want 0", mc.today)
	}
}

func TestMiniCalendarPaint(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetMonth(2024, 3)
	mc.SetToday(15)
	mc.SetBounds(Rect{X: 0, Y: 0, W: 21, H: 7})
	buf := buffer.NewBuffer(21, 7)
	mc.Paint(buf)
	// Header row should have day labels
	if r := buf.GetCell(0, 0).Rune; r == 0 || r == ' ' {
		t.Error("Header should have day labels")
	}
}

func TestMiniCalendarWeekStartMon(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetWeekStartMonday(true)
	if !mc.weekStartMon {
		t.Error("Expected weekStartMon=true")
	}
}

func TestMiniCalendarChildren(t *testing.T) {
	mc := NewMiniCalendar()
	if c := mc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMiniCalendarStyle(t *testing.T) {
	mc := NewMiniCalendar()
	mc.SetStyle(MiniCalendarStyle{
		Header:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Day:     buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Today:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Weekend: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Empty:   buffer.Style{Fg: buffer.RGB(30, 30, 30)},
	})
	mc.SetMonth(2024, 1).SetToday(1)
	buf := buffer.NewBuffer(21, 7)
	mc.SetBounds(Rect{X: 0, Y: 0, W: 21, H: 7})
	mc.Paint(buf)
}

// ─── BatteryGauge Tests ───

func TestBatteryGaugeBasic(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetLevel(75)
	if l := bg.Level(); l != 75 {
		t.Errorf("Level = %d, want 75", l)
	}
}

func TestBatteryGaugeZero(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetLevel(0)
	if l := bg.Level(); l != 0 {
		t.Errorf("Level = %d, want 0", l)
	}
}

func TestBatteryGaugeClamp(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetLevel(-10)
	if l := bg.Level(); l != 0 {
		t.Errorf("Level = %d, want 0 (clamped)", l)
	}
	bg.SetLevel(200)
	if l := bg.Level(); l != 100 {
		t.Errorf("Level = %d, want 100 (clamped)", l)
	}
}

func TestBatteryGaugeCharging(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetCharging(true)
	if !bg.charging {
		t.Error("Expected charging=true")
	}
}

func TestBatteryGaugeColorLevels(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetLevel(80)
	if bg.curStyle.Fg != bg.style.High.Fg {
		t.Error("Expected High style for >= 50%")
	}
	bg.SetLevel(30)
	if bg.curStyle.Fg != bg.style.Medium.Fg {
		t.Error("Expected Medium style for 20-49%")
	}
	bg.SetLevel(10)
	if bg.curStyle.Fg != bg.style.Low.Fg {
		t.Error("Expected Low style for < 20%")
	}
}

func TestBatteryGaugePaint(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetLevel(60)
	bg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	bg.Paint(buf)
	// Should start with '['
	if r := buf.GetCell(0, 0).Rune; r != '[' {
		t.Errorf("First rune = %q, want '['", r)
	}
}

func TestBatteryGaugeChargingPaint(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetLevel(50)
	bg.SetCharging(true)
	bg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	bg.Paint(buf)
	// Should have ⚡ icon
	hasBolt := false
	for i := 0; i < 20; i++ {
		if buf.GetCell(i, 0).Rune == '⚡' {
			hasBolt = true
			break
		}
	}
	if !hasBolt {
		t.Error("Paint should show ⚡ when charging")
	}
}

func TestBatteryGaugeChildren(t *testing.T) {
	bg := NewBatteryGauge()
	if c := bg.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestBatteryGaugeStyle(t *testing.T) {
	bg := NewBatteryGauge()
	bg.SetStyle(BatteryGaugeStyle{
		High:     buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Medium:   buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Low:      buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Charging: buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Shell:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	bg.SetLevel(45).SetCharging(true)
	buf := buffer.NewBuffer(20, 1)
	bg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	bg.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintStepDots(b *testing.B) {
	sd := NewStepDots()
	sd.SetTotal(8).SetCurrent(5).SetConnected(true)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd.Paint(buf)
	}
}

func BenchmarkPaintMiniCalendar(b *testing.B) {
	mc := NewMiniCalendar()
	mc.SetMonth(2024, 6).SetToday(15)
	mc.SetBounds(Rect{X: 0, Y: 0, W: 21, H: 7})
	buf := buffer.NewBuffer(21, 7)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Paint(buf)
	}
}

func BenchmarkPaintBatteryGauge(b *testing.B) {
	bg := NewBatteryGauge()
	bg.SetLevel(75)
	bg.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bg.Paint(buf)
	}
}
