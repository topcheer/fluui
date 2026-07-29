package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Stopwatch Tests ───

func TestStopwatchBasic(t *testing.T) {
	sw := NewStopwatch()
	sw.SetElapsed(3661500) // 1h 1m 1.5s
	if ms := sw.ElapsedMs(); ms != 3661500 {
		t.Errorf("ElapsedMs = %d, want 3661500", ms)
	}
}

func TestStopwatchZero(t *testing.T) {
	sw := NewStopwatch()
	if ms := sw.ElapsedMs(); ms != 0 {
		t.Errorf("ElapsedMs = %d, want 0", ms)
	}
}

func TestStopwatchNegative(t *testing.T) {
	sw := NewStopwatch()
	sw.SetElapsed(-100)
	if ms := sw.ElapsedMs(); ms != 0 {
		t.Errorf("ElapsedMs = %d, want 0 (clamped)", ms)
	}
}

func TestStopwatchRunning(t *testing.T) {
	sw := NewStopwatch()
	sw.SetRunning(true)
	if !sw.running {
		t.Error("Should be running")
	}
}

func TestStopwatchPaint(t *testing.T) {
	sw := NewStopwatch()
	sw.SetElapsed(65000) // 1:05.00
	sw.SetRunning(true)
	sw.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	buf := buffer.NewBuffer(14, 1)
	sw.Paint(buf)
	// Should have ▶ running indicator
	if r := buf.GetCell(0, 0).Rune; r != '▶' {
		t.Errorf("First rune = %q, want '▶'", r)
	}
}

func TestStopwatchChildren(t *testing.T) {
	sw := NewStopwatch()
	if c := sw.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestStopwatchStyle(t *testing.T) {
	sw := NewStopwatch()
	sw.SetStyle(StopwatchStyle{
		Digits: buffer.Style{Fg: buffer.RGB(0, 255, 255)},
		Dot:    buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	sw.SetElapsed(50000)
	buf := buffer.NewBuffer(14, 1)
	sw.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	sw.Paint(buf)
}

// ─── CountdownTimer Tests ───

func TestCountdownBasic(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetRemaining(65000)
	if ms := ct.RemainingMs(); ms != 65000 {
		t.Errorf("RemainingMs = %d, want 65000", ms)
	}
}

func TestCountdownZero(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetRemaining(0)
	if ms := ct.RemainingMs(); ms != 0 {
		t.Errorf("RemainingMs = %d, want 0", ms)
	}
}

func TestCountdownNegative(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetRemaining(-100)
	if ms := ct.RemainingMs(); ms != 0 {
		t.Errorf("RemainingMs = %d, want 0 (clamped)", ms)
	}
}

func TestCountdownUrgency(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetUrgencyThreshold(30000)
	ct.SetRemaining(20000)
	if ct.curStyle.Fg != ct.style.Urgent.Fg {
		t.Error("Expected Urgent style for remaining < threshold")
	}
}

func TestCountdownExpired(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetRemaining(0)
	if ct.curStyle.Fg != ct.style.Expired.Fg {
		t.Error("Expected Expired style for 0 remaining")
	}
}

func TestCountdownNormal(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetUrgencyThreshold(10000)
	ct.SetRemaining(60000)
	if ct.curStyle.Fg != ct.style.Normal.Fg {
		t.Error("Expected Normal style for remaining > threshold")
	}
}

func TestCountdownPaint(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetRemaining(65000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	ct.Paint(buf)
	// Should start with ⏱ icon
	if r := buf.GetCell(0, 0).Rune; r != '⏱' {
		t.Errorf("First rune = %q, want '⏱'", r)
	}
}

func TestCountdownChildren(t *testing.T) {
	ct := NewCountdownTimer()
	if c := ct.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestCountdownStyle(t *testing.T) {
	ct := NewCountdownTimer()
	ct.SetStyle(CountdownTimerStyle{
		Normal:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Urgent:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Expired: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Dot:     buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	ct.SetRemaining(5000)
	buf := buffer.NewBuffer(10, 1)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	ct.Paint(buf)
}

// ─── WaveformDisplay Tests ───

func TestWaveformBasic(t *testing.T) {
	w := NewWaveformDisplay()
	w.SetSamples([]int{30, 60, 80, 45, 20})
	if n := w.SampleCount(); n != 5 {
		t.Errorf("SampleCount = %d, want 5", n)
	}
}

func TestWaveformEmpty(t *testing.T) {
	w := NewWaveformDisplay()
	if n := w.SampleCount(); n != 0 {
		t.Errorf("SampleCount = %d, want 0", n)
	}
}

func TestWaveformClamp(t *testing.T) {
	w := NewWaveformDisplay()
	w.SetSamples([]int{-10, 50, 200})
	if w.samples[0] != 0 || w.samples[2] != 100 {
		t.Errorf("Samples not clamped: %d, %d", w.samples[0], w.samples[2])
	}
}

func TestWaveformOverflow(t *testing.T) {
	w := NewWaveformDisplay()
	samples := make([]int, waveformMaxSamples+10)
	for i := range samples {
		samples[i] = 50
	}
	w.SetSamples(samples)
	if n := w.SampleCount(); n != waveformMaxSamples {
		t.Errorf("SampleCount = %d, want %d (capped)", n, waveformMaxSamples)
	}
}

func TestWaveformHeight(t *testing.T) {
	w := NewWaveformDisplay()
	w.SetHeight(10)
	if w.height != 10 {
		t.Errorf("height = %d, want 10", w.height)
	}
	w.SetHeight(1)
	if w.height != 2 {
		t.Errorf("height = %d, want 2 (clamped)", w.height)
	}
}

func TestWaveformPaint(t *testing.T) {
	w := NewWaveformDisplay()
	w.SetSamples([]int{50, 80, 30, 60, 90})
	w.SetHeight(6)
	w.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 6})
	buf := buffer.NewBuffer(10, 6)
	w.Paint(buf)
	// Should have bar content
	hasBar := false
	for row := 0; row < 6; row++ {
		for col := 0; col < 5; col++ {
			if buf.GetCell(col, row).Rune == '█' {
				hasBar = true
				break
			}
		}
		if hasBar {
			break
		}
	}
	if !hasBar {
		t.Error("Paint should have bar content")
	}
}

func TestWaveformChildren(t *testing.T) {
	w := NewWaveformDisplay()
	if c := w.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestWaveformStyle(t *testing.T) {
	w := NewWaveformDisplay()
	w.SetStyle(WaveformStyle{
		Peak:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Normal: buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Low:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Center: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	w.SetSamples([]int{90, 50, 20})
	w.SetHeight(6)
	buf := buffer.NewBuffer(10, 6)
	w.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 6})
	w.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintStopwatch(b *testing.B) {
	sw := NewStopwatch()
	sw.SetElapsed(3661500)
	sw.SetRunning(true)
	sw.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 1})
	buf := buffer.NewBuffer(14, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sw.Paint(buf)
	}
}

func BenchmarkPaintCountdownTimer(b *testing.B) {
	ct := NewCountdownTimer()
	ct.SetRemaining(65000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ct.Paint(buf)
	}
}

func BenchmarkPaintWaveformDisplay(b *testing.B) {
	w := NewWaveformDisplay()
	w.SetSamples([]int{50, 80, 30, 60, 90, 40, 70, 20, 85, 55})
	w.SetHeight(6)
	w.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 6})
	buf := buffer.NewBuffer(10, 6)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Paint(buf)
	}
}
