package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── OutputFormatSelector Tests ───

func TestOutputFormatBasic(t *testing.T) {
	of := NewOutputFormatSelector()
	if active := of.Active(); active != FormatText {
		t.Errorf("Active = %d, want FormatText(%d)", active, FormatText)
	}
}

func TestOutputFormatSet(t *testing.T) {
	of := NewOutputFormatSelector()
	of.SetActive(FormatMarkdown)
	if active := of.Active(); active != FormatMarkdown {
		t.Errorf("Active = %d, want FormatMarkdown(%d)", active, FormatMarkdown)
	}
}

func TestOutputFormatCycle(t *testing.T) {
	of := NewOutputFormatSelector()
	of.CycleNext()
	if active := of.Active(); active != FormatJSON {
		t.Errorf("After CycleNext: Active = %d, want FormatJSON(%d)", active, FormatJSON)
	}
	// Cycle through all
	for i := 0; i < len(formatNames)-1; i++ {
		of.CycleNext()
	}
	// Should wrap back to FormatText
	if active := of.Active(); active != FormatText {
		t.Errorf("After full cycle: Active = %d, want FormatText(%d)", active, FormatText)
	}
}

func TestOutputFormatInvalid(t *testing.T) {
	of := NewOutputFormatSelector()
	of.SetActive(OutputFormat(99))
	if active := of.Active(); active != FormatText {
		t.Errorf("Active = %d, want FormatText (clamped)", active)
	}
}

func TestOutputFormatPaint(t *testing.T) {
	of := NewOutputFormatSelector()
	of.SetActive(FormatJSON)
	of.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	of.Paint(buf)
	// Should start with '['
	if r := buf.GetCell(0, 0).Rune; r != '[' {
		t.Errorf("First rune = %q, want '['", r)
	}
}

func TestOutputFormatChildren(t *testing.T) {
	of := NewOutputFormatSelector()
	if children := of.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestOutputFormatStyle(t *testing.T) {
	of := NewOutputFormatSelector()
	custom := OutputFormatStyle{
		Active:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Inactive:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Bracket:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Separator: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	of.SetStyle(custom)
	of.SetActive(FormatCode)
	buf := buffer.NewBuffer(40, 1)
	of.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	of.Paint(buf)
}

// ─── GradientBar Tests ───

func TestGradientBarBasic(t *testing.T) {
	gb := NewGradientBar()
	gb.SetValue(50, 100)
	if pct := gb.Percent(); pct != 50 {
		t.Errorf("Percent = %d, want 50", pct)
	}
}

func TestGradientBarFull(t *testing.T) {
	gb := NewGradientBar()
	gb.SetValue(100, 100)
	if pct := gb.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100", pct)
	}
}

func TestGradientBarZero(t *testing.T) {
	gb := NewGradientBar()
	gb.SetValue(0, 100)
	if pct := gb.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestGradientBarNegative(t *testing.T) {
	gb := NewGradientBar()
	gb.SetValue(-10, -5)
	if pct := gb.Percent(); pct != 0 {
		t.Errorf("Percent = %d, want 0", pct)
	}
}

func TestGradientBarOverflow(t *testing.T) {
	gb := NewGradientBar()
	gb.SetValue(200, 100)
	if pct := gb.Percent(); pct != 100 {
		t.Errorf("Percent = %d, want 100 (capped)", pct)
	}
}

func TestGradientBarWidth(t *testing.T) {
	gb := NewGradientBar()
	gb.SetWidth(30)
	if gb.width != 30 {
		t.Errorf("width = %d, want 30", gb.width)
	}
	gb.SetWidth(5)
	if gb.width != 10 {
		t.Errorf("width = %d, want 10 (clamped)", gb.width)
	}
}

func TestGradientBarPaint(t *testing.T) {
	gb := NewGradientBar()
	gb.SetValue(50, 100)
	gb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	gb.Paint(buf)
	// Should have filled cells (█)
	hasFilled := false
	for i := 0; i < 30; i++ {
		if buf.GetCell(i, 0).Rune == '█' {
			hasFilled = true
			break
		}
	}
	if !hasFilled {
		t.Error("Paint should have filled cells")
	}
}

func TestGradientBarColorGradient(t *testing.T) {
	// Test gradientBarColor function
	r, g, b := gradientBarColor(0)
	if r != 34 || g != 197 || b != 94 {
		t.Errorf("gradientBarColor(0) = (%d,%d,%d), want green (34,197,94)", r, g, b)
	}
	r, g, b = gradientBarColor(50)
	if r != 234 || g != 179 || b != 8 {
		t.Errorf("gradientBarColor(50) = (%d,%d,%d), want yellow (234,179,8)", r, g, b)
	}
	r, g, b = gradientBarColor(100)
	if r != 239 || g != 68 || b != 68 {
		t.Errorf("gradientBarColor(100) = (%d,%d,%d), want red (239,68,68)", r, g, b)
	}
}

func TestGradientBarChildren(t *testing.T) {
	gb := NewGradientBar()
	if children := gb.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestGradientBarStyle(t *testing.T) {
	gb := NewGradientBar()
	custom := GradientBarStyle{
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	}
	gb.SetStyle(custom)
	gb.SetValue(75, 100)
	buf := buffer.NewBuffer(30, 1)
	gb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	gb.Paint(buf)
}

// ─── ModelParameterBar Tests ───

func TestModelParameterBasic(t *testing.T) {
	mp := NewModelParameterBar()
	if temp := mp.Temperature(); temp != 70 {
		t.Errorf("Temperature = %d, want 70", temp)
	}
}

func TestModelParameterSet(t *testing.T) {
	mp := NewModelParameterBar()
	mp.SetParams(50, 80, 2048)
	if temp := mp.Temperature(); temp != 50 {
		t.Errorf("Temperature = %d, want 50", temp)
	}
}

func TestModelParameterClamp(t *testing.T) {
	mp := NewModelParameterBar()
	mp.SetParams(300, 150, -100)
	if temp := mp.Temperature(); temp != 200 {
		t.Errorf("Temperature = %d, want 200 (clamped)", temp)
	}
}

func TestModelParameterZero(t *testing.T) {
	mp := NewModelParameterBar()
	mp.SetParams(0, 0, 0)
	if temp := mp.Temperature(); temp != 0 {
		t.Errorf("Temperature = %d, want 0", temp)
	}
}

func TestModelParameterFormatting(t *testing.T) {
	mp := NewModelParameterBar()
	mp.SetParams(70, 90, 4096)
	// Check cached formatting strings
	if mp.tempStr != "0.70" {
		t.Errorf("tempStr = %q, want '0.70'", mp.tempStr)
	}
	if mp.topPStr != "0.90" {
		t.Errorf("topPStr = %q, want '0.90'", mp.topPStr)
	}
}

func TestModelParameterFormattingHighTemp(t *testing.T) {
	mp := NewModelParameterBar()
	mp.SetParams(150, 50, 8192)
	// 1.50
	if mp.tempStr != "1.50" {
		t.Errorf("tempStr = %q, want '1.50'", mp.tempStr)
	}
}

func TestModelParameterPaint(t *testing.T) {
	mp := NewModelParameterBar()
	mp.SetParams(70, 90, 4096)
	mp.SetBounds(Rect{X: 0, Y: 0, W: 34, H: 1})
	buf := buffer.NewBuffer(34, 1)
	mp.Paint(buf)
	// Should start with 't' from "temp:"
	if r := buf.GetCell(0, 0).Rune; r != 't' {
		t.Errorf("First rune = %q, want 't'", r)
	}
}

func TestModelParameterChildren(t *testing.T) {
	mp := NewModelParameterBar()
	if children := mp.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestModelParameterStyle(t *testing.T) {
	mp := NewModelParameterBar()
	custom := ModelParameterStyle{
		Label:     buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:     buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Separator: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	}
	mp.SetStyle(custom)
	mp.SetParams(80, 95, 6000)
	buf := buffer.NewBuffer(34, 1)
	mp.SetBounds(Rect{X: 0, Y: 0, W: 34, H: 1})
	mp.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintOutputFormatSelector(b *testing.B) {
	of := NewOutputFormatSelector()
	of.SetActive(FormatMarkdown)
	of.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		of.Paint(buf)
	}
}

func BenchmarkPaintGradientBar(b *testing.B) {
	gb := NewGradientBar()
	gb.SetValue(65, 100)
	gb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gb.Paint(buf)
	}
}

func BenchmarkPaintModelParameterBar(b *testing.B) {
	mp := NewModelParameterBar()
	mp.SetParams(70, 90, 4096)
	mp.SetBounds(Rect{X: 0, Y: 0, W: 34, H: 1})
	buf := buffer.NewBuffer(34, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mp.Paint(buf)
	}
}
