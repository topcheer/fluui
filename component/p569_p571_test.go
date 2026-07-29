package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Marquee Tests ───

func TestMarqueeBasic(t *testing.T) {
	m := NewMarquee()
	m.SetText("Hello World")
	if m.Offset() != 0 {
		t.Errorf("Offset = %d, want 0", m.Offset())
	}
}

func TestMarqueeShort(t *testing.T) {
	m := NewMarquee()
	m.SetText("Hi")
	m.SetWidth(10)
	m.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	m.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r != 'H' {
		t.Errorf("First rune = %q, want 'H'", r)
	}
}

func TestMarqueeScrolling(t *testing.T) {
	m := NewMarquee()
	m.SetText("This is a long scrolling text")
	m.SetWidth(10)
	m.SetOffset(5)
	m.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	m.Paint(buf)
	// At offset 5, first char should be the 6th rune 'i'
	if r := buf.GetCell(0, 0).Rune; r != 'i' {
		t.Errorf("Offset 5 rune = %q, want 'i'", r)
	}
}

func TestMarqueeBounce(t *testing.T) {
	m := NewMarquee()
	m.SetBounce(true)
	if !m.bounce {
		t.Error("Bounce should be true")
	}
}

func TestMarqueeNegativeOffset(t *testing.T) {
	m := NewMarquee()
	m.SetOffset(-5)
	if o := m.Offset(); o != 0 {
		t.Errorf("Offset = %d, want 0 (clamped)", o)
	}
}

func TestMarqueeWidthClamp(t *testing.T) {
	m := NewMarquee()
	m.SetWidth(2)
	if m.width != 5 {
		t.Errorf("width = %d, want 5 (clamped)", m.width)
	}
}

func TestMarqueeEmpty(t *testing.T) {
	m := NewMarquee()
	m.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	m.Paint(buf) // should not panic
}

func TestMarqueeChildren(t *testing.T) {
	m := NewMarquee()
	if c := m.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMarqueeStyle(t *testing.T) {
	m := NewMarquee()
	m.SetStyle(MarqueeStyle{
		Text:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Indicator: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
	})
	m.SetText("test")
	buf := buffer.NewBuffer(10, 1)
	m.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	m.Paint(buf)
}

// ─── DigitalClock Tests ───

func TestDigitalClockBasic(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetTime(14, 30, 45)
	if h := dc.Hour(); h != 14 {
		t.Errorf("Hour = %d, want 14", h)
	}
}

func TestDigitalClockFormat24h(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetTime(9, 5, 3)
	dc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	dc.Paint(buf)
	// Should show "09:05:03"
	if r := buf.GetCell(0, 0).Rune; r != '0' {
		t.Errorf("First rune = %q, want '0'", r)
	}
}

func TestDigitalClockFormat12h(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetFormat24h(false)
	dc.SetTime(14, 30, 0)
	if dc.suffixStr != "PM" {
		t.Errorf("suffix = %q, want 'PM'", dc.suffixStr)
	}
}

func TestDigitalClockMidnight12h(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetFormat24h(false)
	dc.SetTime(0, 0, 0)
	if dc.suffixStr != "AM" {
		t.Errorf("suffix = %q, want 'AM'", dc.suffixStr)
	}
	// 0 hour in 12h = 12 AM
	if dc.timeStr[:2] != "12" {
		t.Errorf("timeStr = %q, want '12:...'", dc.timeStr)
	}
}

func TestDigitalClockNoon12h(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetFormat24h(false)
	dc.SetTime(12, 0, 0)
	if dc.suffixStr != "PM" {
		t.Errorf("suffix = %q, want 'PM'", dc.suffixStr)
	}
}

func TestDigitalClockClamp(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetTime(25, 70, 99)
	if h := dc.Hour(); h != 23 {
		t.Errorf("Hour = %d, want 23 (clamped)", h)
	}
}

func TestDigitalClockPaint(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetTime(10, 20, 30)
	dc.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(8, 1)
	dc.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r != '1' {
		t.Errorf("First rune = %q, want '1'", r)
	}
}

func TestDigitalClockChildren(t *testing.T) {
	dc := NewDigitalClock()
	if c := dc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestDigitalClockStyle(t *testing.T) {
	dc := NewDigitalClock()
	dc.SetStyle(DigitalClockStyle{
		Digit:  buffer.Style{Fg: buffer.RGB(0, 255, 255)},
		Colon:  buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Suffix: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	dc.SetTime(12, 0, 0)
	buf := buffer.NewBuffer(10, 1)
	dc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	dc.Paint(buf)
}

// ─── VolumeMeter Tests ───

func TestVolumeMeterBasic(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetLevel(75)
	if l := vm.Level(); l != 75 {
		t.Errorf("Level = %d, want 75", l)
	}
}

func TestVolumeMeterZero(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetLevel(0)
	if l := vm.Level(); l != 0 {
		t.Errorf("Level = %d, want 0", l)
	}
}

func TestVolumeMeterClamp(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetLevel(-10)
	if l := vm.Level(); l != 0 {
		t.Errorf("Level = %d, want 0 (clamped)", l)
	}
	vm.SetLevel(200)
	if l := vm.Level(); l != 100 {
		t.Errorf("Level = %d, want 100 (clamped)", l)
	}
}

func TestVolumeMeterMuted(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetLevel(50)
	vm.SetMuted(true)
	if vm.displayStr != "MUTED" {
		t.Errorf("displayStr = %q, want 'MUTED'", vm.displayStr)
	}
}

func TestVolumeMeterColorLevels(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetLevel(30)
	if vm.curStyle.Fg != vm.style.Low.Fg {
		t.Error("Expected Low style for level < 50")
	}
	vm.SetLevel(60)
	if vm.curStyle.Fg != vm.style.Medium.Fg {
		t.Error("Expected Medium style for level 50-79")
	}
	vm.SetLevel(90)
	if vm.curStyle.Fg != vm.style.High.Fg {
		t.Error("Expected High style for level >= 80")
	}
}

func TestVolumeMeterPaint(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetLevel(50)
	vm.SetBounds(Rect{X: 0, Y: 0, W: 28, H: 1})
	buf := buffer.NewBuffer(28, 1)
	vm.Paint(buf)
	// Should have filled bar cells
	hasFilled := false
	for i := 0; i < 28; i++ {
		if buf.GetCell(i, 0).Rune == '█' {
			hasFilled = true
			break
		}
	}
	if !hasFilled {
		t.Error("Paint should have filled bar cells")
	}
}

func TestVolumeMeterWidthClamp(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetWidth(5)
	if vm.width != 10 {
		t.Errorf("width = %d, want 10 (clamped)", vm.width)
	}
}

func TestVolumeMeterChildren(t *testing.T) {
	vm := NewVolumeMeter()
	if c := vm.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestVolumeMeterStyle(t *testing.T) {
	vm := NewVolumeMeter()
	vm.SetStyle(VolumeMeterStyle{
		Low:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Medium: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		High:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Muted:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	vm.SetLevel(70)
	buf := buffer.NewBuffer(28, 1)
	vm.SetBounds(Rect{X: 0, Y: 0, W: 28, H: 1})
	vm.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintMarquee(b *testing.B) {
	m := NewMarquee()
	m.SetText("This is a scrolling marquee text that is longer than the window")
	m.SetWidth(30)
	m.SetOffset(10)
	m.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Paint(buf)
	}
}

func BenchmarkPaintDigitalClock(b *testing.B) {
	dc := NewDigitalClock()
	dc.SetTime(14, 30, 45)
	dc.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(8, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dc.Paint(buf)
	}
}

func BenchmarkPaintVolumeMeter(b *testing.B) {
	vm := NewVolumeMeter()
	vm.SetLevel(75)
	vm.SetBounds(Rect{X: 0, Y: 0, W: 28, H: 1})
	buf := buffer.NewBuffer(28, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.Paint(buf)
	}
}
