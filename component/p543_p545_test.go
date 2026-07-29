package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StreamingCursor Tests ───

func TestStreamingCursorBasic(t *testing.T) {
	c := NewStreamingCursor()
	if c.IsActive() {
		t.Error("New cursor should be inactive")
	}
}

func TestStreamingCursorActive(t *testing.T) {
	c := NewStreamingCursor()
	c.SetActive(true)
	if !c.IsActive() {
		t.Error("Cursor should be active")
	}
}

func TestStreamingCursorFrame(t *testing.T) {
	c := NewStreamingCursor()
	c.SetFrame(5)
	// Just verify it doesn't panic on large frame values
	c.SetFrame(100)
}

func TestStreamingCursorLabel(t *testing.T) {
	c := NewStreamingCursor()
	c.SetLabel("thinking")
	if c.label != "thinking" {
		t.Errorf("label = %q, want 'thinking'", c.label)
	}
}

func TestStreamingCursorPaintActive(t *testing.T) {
	c := NewStreamingCursor()
	c.SetActive(true)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	c.Paint(buf)
	// First char should be '['
	if r := buf.GetCell(0, 0).Rune; r != '[' {
		t.Errorf("First rune = %q, want '['", r)
	}
}

func TestStreamingCursorPaintIdle(t *testing.T) {
	c := NewStreamingCursor()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	c.Paint(buf)
	// First char should be '['
	if r := buf.GetCell(0, 0).Rune; r != '[' {
		t.Errorf("First rune = %q, want '['", r)
	}
	// Second char should be '●' in idle mode
	if r := buf.GetCell(1, 0).Rune; r != '●' {
		t.Errorf("Second rune = %q, want '●'", r)
	}
}

func TestStreamingCursorChildren(t *testing.T) {
	c := NewStreamingCursor()
	if children := c.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestStreamingCursorStyle(t *testing.T) {
	c := NewStreamingCursor()
	custom := StreamingCursorStyle{
		Active:   buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Idle:     buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		BracketL: buffer.Style{Fg: buffer.RGB(50, 50, 50)},
		BracketR: buffer.Style{Fg: buffer.RGB(50, 50, 50)},
	}
	c.SetStyle(custom)
	c.SetActive(true)
	buf := buffer.NewBuffer(20, 1)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	c.Paint(buf)
}

// ─── AIMemoryBar Tests ───

func TestAIMemoryBarBasic(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(2000, 4000, 16000)
	pct := m.UsedPercent()
	// (2000+4000)/16000 = 37%
	if pct != 37 {
		t.Errorf("UsedPercent = %d, want 37", pct)
	}
}

func TestAIMemoryBarFull(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(8000, 8000, 16000)
	if pct := m.UsedPercent(); pct != 100 {
		t.Errorf("UsedPercent = %d, want 100", pct)
	}
}

func TestAIMemoryBarOverflow(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(20000, 20000, 16000)
	// Should cap at 100%
	if pct := m.UsedPercent(); pct != 100 {
		t.Errorf("UsedPercent = %d, want 100 (capped)", pct)
	}
}

func TestAIMemoryBarZero(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(0, 0, 16000)
	if pct := m.UsedPercent(); pct != 0 {
		t.Errorf("UsedPercent = %d, want 0", pct)
	}
}

func TestAIMemoryBarNegative(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(-10, -20, -5)
	// Should clamp to 0/0/1
	if pct := m.UsedPercent(); pct != 0 {
		t.Errorf("UsedPercent = %d, want 0", pct)
	}
}

func TestAIMemoryBarWarning(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(7000, 7000, 16000)
	// 14000/16000 = 87% → should be warning
	if !m.isWarning {
		t.Error("Expected warning state at 87% usage")
	}
}

func TestAIMemoryBarPaint(t *testing.T) {
	m := NewAIMemoryBar()
	m.SetUsage(2000, 6000, 16000)
	m.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	m.Paint(buf)
	hasContent := false
	for i := 0; i < 50; i++ {
		r := buf.GetCell(i, 0).Rune
		if r == '█' || r == '░' {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no bar content")
	}
}

func TestAIMemoryBarChildren(t *testing.T) {
	m := NewAIMemoryBar()
	if children := m.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestAIMemoryBarStyle(t *testing.T) {
	m := NewAIMemoryBar()
	custom := AIMemoryStyle{
		System:  buffer.Style{Fg: buffer.RGB(255, 0, 255)},
		History: buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Avail:   buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Warning: buffer.Style{Fg: buffer.RGB(255, 128, 0)},
	}
	m.SetStyle(custom)
	m.SetUsage(3000, 5000, 16000)
	buf := buffer.NewBuffer(50, 3)
	m.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	m.Paint(buf)
}

// ─── ResponseTimer Tests ───

func TestResponseTimerBasic(t *testing.T) {
	rt := NewResponseTimer()
	rt.SetDurations(120, 3500)
	if tfb := rt.TTFB(); tfb != 120 {
		t.Errorf("TTFB = %d, want 120", tfb)
	}
	if total := rt.TotalDuration(); total != 3500 {
		t.Errorf("TotalDuration = %d, want 3500", total)
	}
}

func TestResponseTimerZero(t *testing.T) {
	rt := NewResponseTimer()
	if tfb := rt.TTFB(); tfb != 0 {
		t.Errorf("TTFB = %d, want 0", tfb)
	}
}

func TestResponseTimerNegative(t *testing.T) {
	rt := NewResponseTimer()
	rt.SetDurations(-10, -50)
	if tfb := rt.TTFB(); tfb != 0 {
		t.Errorf("TTFB = %d, want 0 (clamped)", tfb)
	}
}

func TestResponseTimerPaint(t *testing.T) {
	rt := NewResponseTimer()
	rt.SetDurations(120, 3500)
	rt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	rt.Paint(buf)
	// Should start with ◷ icon
	if r := buf.GetCell(0, 0).Rune; r != '◷' {
		t.Errorf("First rune = %q, want '◷'", r)
	}
}

func TestResponseTimerChildren(t *testing.T) {
	rt := NewResponseTimer()
	if children := rt.Children(); children != nil {
		t.Errorf("Children = %v, want nil", children)
	}
}

func TestResponseTimerStyle(t *testing.T) {
	rt := NewResponseTimer()
	custom := ResponseTimerStyle{
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Unit:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Icon:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
	}
	rt.SetStyle(custom)
	rt.SetDurations(100, 2000)
	buf := buffer.NewBuffer(30, 1)
	rt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	rt.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintStreamingCursor(b *testing.B) {
	c := NewStreamingCursor()
	c.SetActive(true)
	c.SetFrame(3)
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintAIMemoryBar(b *testing.B) {
	m := NewAIMemoryBar()
	m.SetUsage(2000, 6000, 16000)
	m.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Paint(buf)
	}
}

func BenchmarkPaintResponseTimer(b *testing.B) {
	rt := NewResponseTimer()
	rt.SetDurations(120, 3500)
	rt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rt.Paint(buf)
	}
}
