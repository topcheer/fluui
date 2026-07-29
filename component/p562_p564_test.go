package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── RadialGauge Tests ───

func TestRadialGaugeBasic(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetValue(50)
	if v := rg.Value(); v != 50 {
		t.Errorf("Value = %d, want 50", v)
	}
}

func TestRadialGaugeZero(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetValue(0)
	if v := rg.Value(); v != 0 {
		t.Errorf("Value = %d, want 0", v)
	}
}

func TestRadialGaugeFull(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetValue(100)
	if v := rg.Value(); v != 100 {
		t.Errorf("Value = %d, want 100", v)
	}
}

func TestRadialGaugeClamp(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetValue(-10)
	if v := rg.Value(); v != 0 {
		t.Errorf("Value = %d, want 0 (clamped)", v)
	}
	rg.SetValue(200)
	if v := rg.Value(); v != 100 {
		t.Errorf("Value = %d, want 100 (clamped)", v)
	}
}

func TestRadialGaugeLabel(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetLabel("CPU")
	if rg.label != "CPU" {
		t.Errorf("label = %q, want 'CPU'", rg.label)
	}
}

func TestRadialGaugePaint(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetValue(75)
	rg.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 4})
	buf := buffer.NewBuffer(16, 4)
	rg.Paint(buf)
	// Center should have '%' character
	hasContent := false
	for i := 0; i < 16; i++ {
		if buf.GetCell(i, 1).Rune == '%' {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint should show percentage in center")
	}
}

func TestRadialGaugeChildren(t *testing.T) {
	rg := NewRadialGauge()
	if c := rg.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestRadialGaugeStyle(t *testing.T) {
	rg := NewRadialGauge()
	rg.SetStyle(RadialGaugeStyle{
		Filled: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Empty:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Center: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	rg.SetValue(50)
	buf := buffer.NewBuffer(16, 4)
	rg.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 4})
	rg.Paint(buf)
}

// ─── SignalBars Tests ───

func TestSignalBarsBasic(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(4)
	if l := sb.Level(); l != 4 {
		t.Errorf("Level = %d, want 4", l)
	}
}

func TestSignalBarsZero(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(0)
	if l := sb.Level(); l != 0 {
		t.Errorf("Level = %d, want 0", l)
	}
}

func TestSignalBarsMax(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(5)
	if l := sb.Level(); l != 5 {
		t.Errorf("Level = %d, want 5", l)
	}
}

func TestSignalBarsClamp(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(-5)
	if l := sb.Level(); l != 0 {
		t.Errorf("Level = %d, want 0 (clamped)", l)
	}
	sb.SetLevel(99)
	if l := sb.Level(); l != 5 {
		t.Errorf("Level = %d, want 5 (clamped)", l)
	}
}

func TestSignalBarsDbm(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(3)
	sb.SetDbm(-67)
	if sb.dbm != -67 {
		t.Errorf("dbm = %d, want -67", sb.dbm)
	}
}

func TestSignalBarsColorLevels(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(5)
	if sb.curStyle.Fg != sb.style.Strong.Fg {
		t.Error("Expected Strong style for level >= 4")
	}
	sb.SetLevel(2)
	if sb.curStyle.Fg != sb.style.Medium.Fg {
		t.Error("Expected Medium style for level 2-3")
	}
	sb.SetLevel(0)
	if sb.curStyle.Fg != sb.style.Weak.Fg {
		t.Error("Expected Weak style for level < 2")
	}
}

func TestSignalBarsPaint(t *testing.T) {
	sb := NewSignalBars()
	sb.SetLevel(3)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 5})
	buf := buffer.NewBuffer(14, 5)
	sb.Paint(buf)
	// Should have filled bars
	hasFilled := false
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			if buf.GetCell(col, row).Rune == '█' {
				hasFilled = true
				break
			}
		}
		if hasFilled { break }
	}
	if !hasFilled {
		t.Error("Paint should have filled bars")
	}
}

func TestSignalBarsChildren(t *testing.T) {
	sb := NewSignalBars()
	if c := sb.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestSignalBarsStyle(t *testing.T) {
	sb := NewSignalBars()
	sb.SetStyle(SignalBarsStyle{
		Strong: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Medium: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Weak:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	sb.SetLevel(4)
	buf := buffer.NewBuffer(14, 5)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 5})
	sb.Paint(buf)
}

// ─── Thermometer Tests ───

func TestThermometerBasic(t *testing.T) {
	th := NewThermometer()
	th.SetTemperature(25, 0, 100)
	if t2 := th.Temperature(); t2 != 25 {
		t.Errorf("Temperature = %d, want 25", t2)
	}
}

func TestThermometerZero(t *testing.T) {
	th := NewThermometer()
	th.SetTemperature(0, 0, 100)
	if v := th.Temperature(); v != 0 {
		t.Errorf("Temperature = %d, want 0", v)
	}
}

func TestThermometerClamp(t *testing.T) {
	th := NewThermometer()
	th.SetTemperature(-20, 0, 100)
	if v := th.Temperature(); v != 0 {
		t.Errorf("Temperature = %d, want 0 (clamped)", v)
	}
	th.SetTemperature(200, 0, 100)
	if v := th.Temperature(); v != 100 {
		t.Errorf("Temperature = %d, want 100 (clamped)", v)
	}
}

func TestThermometerInvalidRange(t *testing.T) {
	th := NewThermometer()
	th.SetTemperature(50, 50, 50) // min == max
	if v := th.Temperature(); v != 50 {
		t.Errorf("Temperature = %d, want 50", v)
	}
}

func TestThermometerColorLevels(t *testing.T) {
	th := NewThermometer()
	th.SetTemperature(90, 0, 100)
	if th.curStyle.Fg != th.style.Hot.Fg {
		t.Error("Expected Hot style for >= 70%")
	}
	th.SetTemperature(50, 0, 100)
	if th.curStyle.Fg != th.style.Warm.Fg {
		t.Error("Expected Warm style for 40-69%")
	}
	th.SetTemperature(20, 0, 100)
	if th.curStyle.Fg != th.style.Cold.Fg {
		t.Error("Expected Cold style for < 40%")
	}
}

func TestThermometerHeight(t *testing.T) {
	th := NewThermometer()
	th.SetHeight(12)
	if th.height != 12 {
		t.Errorf("height = %d, want 12", th.height)
	}
	th.SetHeight(2)
	if th.height != 4 {
		t.Errorf("height = %d, want 4 (clamped)", th.height)
	}
}

func TestThermometerPaint(t *testing.T) {
	th := NewThermometer()
	th.SetTemperature(75, 0, 100)
	th.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 9})
	buf := buffer.NewBuffer(8, 9)
	th.Paint(buf)
	// Should have tube walls
	if r := buf.GetCell(0, 0).Rune; r != '│' {
		t.Errorf("First rune = %q, want '│'", r)
	}
	// Should have bulb at bottom
	if r := buf.GetCell(1, 8).Rune; r != '◯' {
		t.Errorf("Bulb rune = %q, want '◯'", r)
	}
}

func TestThermometerChildren(t *testing.T) {
	th := NewThermometer()
	if c := th.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestThermometerStyle(t *testing.T) {
	th := NewThermometer()
	th.SetStyle(ThermometerStyle{
		Cold:  buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Warm:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Hot:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Bulb:  buffer.Style{Fg: buffer.RGB(255, 0, 0)},
	})
	th.SetTemperature(80, 0, 100)
	buf := buffer.NewBuffer(8, 9)
	th.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 9})
	th.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintRadialGauge(b *testing.B) {
	rg := NewRadialGauge()
	rg.SetValue(65)
	rg.SetBounds(Rect{X: 0, Y: 0, W: 16, H: 4})
	buf := buffer.NewBuffer(16, 4)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rg.Paint(buf)
	}
}

func BenchmarkPaintSignalBars(b *testing.B) {
	sb := NewSignalBars()
	sb.SetLevel(3)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 5})
	buf := buffer.NewBuffer(14, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb.Paint(buf)
	}
}

func BenchmarkPaintThermometer(b *testing.B) {
	th := NewThermometer()
	th.SetTemperature(75, 0, 100)
	th.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 9})
	buf := buffer.NewBuffer(8, 9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		th.Paint(buf)
	}
}
