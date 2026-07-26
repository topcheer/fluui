package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === Tooltip tests ===

func TestP397_NewHintLabel(t *testing.T) {
	tt := NewHintLabel("Press Ctrl+S to save")
	if tt.Text() != "Press Ctrl+S to save" {
		t.Errorf("Text = %q", tt.Text())
	}
	if tt.ID() == "" { t.Error("ID empty") }
}

func TestP397_HintLabel_SetText(t *testing.T) {
	tt := NewHintLabel("old")
	tt.SetText("new")
	if tt.Text() != "new" { t.Errorf("Text = %q", tt.Text()) }
}

func TestP397_HintLabel_SetColors(t *testing.T) {
	tt := NewHintLabel("test")
	tt.SetColors(buffer.RGB(1, 2, 3), buffer.RGB(4, 5, 6))
	tt.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	tt.Paint(buf)
	cell := buf.GetCell(1, 0)
	if cell.Fg != buffer.RGB(1, 2, 3) { t.Error("custom fg not applied") }
}

func TestP397_HintLabel_Measure(t *testing.T) {
	tt := NewHintLabel("Hello")
	s := tt.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// " Hello " = 7
	if s.W != 7 { t.Errorf("W = %d, want 7", s.W) }
	if s.H != 1 { t.Errorf("H = %d", s.H) }
}

func TestP397_HintLabel_Paint(t *testing.T) {
	tt := NewHintLabel("Tip")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 1})
	buf := buffer.NewBuffer(6, 1)
	tt.Paint(buf)
	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' { t.Errorf("cell[0] = %q, want ' '", string(c0.Rune)) }
	c1 := buf.GetCell(1, 0)
	if c1.Rune != 'T' { t.Errorf("cell[1] = %q, want 'T'", string(c1.Rune)) }
}

func TestP397_HintLabel_Paint_LongText(t *testing.T) {
	tt := NewHintLabel("This is a very long tooltip message")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 6, H: 1})
	buf := buffer.NewBuffer(6, 1)
	tt.Paint(buf) // should truncate with …
}

func TestP397_HintLabel_Paint_ZeroBounds(t *testing.T) {
	tt := NewHintLabel("test")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tt.Paint(buf)
}

func TestP397_HintLabel_Concurrent(t *testing.T) {
	tt := NewHintLabel("test")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { tt.SetText("concurrent") }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = tt.Text() }
	<-done
}

func TestP397_HintLabel_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Tooltip)(nil)
}

func BenchmarkP397_HintLabel_Paint(b *testing.B) {
	tt := NewHintLabel("Press Ctrl+S to save your work")
	tt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { tt.Paint(buf) }
}

// === StatCard tests ===

func TestP397_NewStatCard(t *testing.T) {
	sc := NewStatCard("Latency", "42ms")
	if sc.Label() != "Latency" { t.Errorf("Label = %q", sc.Label()) }
	if sc.Value() != "42ms" { t.Errorf("Value = %q", sc.Value()) }
	if sc.ID() == "" { t.Error("ID empty") }
}

func TestP397_StatCard_SetLabel(t *testing.T) {
	sc := NewStatCard("old", "val")
	sc.SetLabel("new")
	if sc.Label() != "new" { t.Errorf("Label = %q", sc.Label()) }
}

func TestP397_StatCard_SetValue(t *testing.T) {
	sc := NewStatCard("label", "old")
	sc.SetValue("new")
	if sc.Value() != "new" { t.Errorf("Value = %q", sc.Value()) }
}

func TestP397_StatCard_SetDelta(t *testing.T) {
	sc := NewStatCard("TPS", "1500")
	sc.SetDelta("+12%", true)
	d, pos := sc.Delta()
	if d != "+12%" { t.Errorf("Delta = %q", d) }
	if !pos { t.Error("should be positive") }

	sc.SetDelta("-5%", false)
	d, pos = sc.Delta()
	if d != "-5%" { t.Errorf("Delta = %q", d) }
	if pos { t.Error("should be negative") }
}

func TestP397_StatCard_ClearDelta(t *testing.T) {
	sc := NewStatCard("TPS", "100")
	sc.SetDelta("+1%", true)
	sc.ClearDelta()
	d, _ := sc.Delta()
	if d != "" { t.Errorf("Delta = %q, want empty", d) }
}

func TestP397_StatCard_Measure(t *testing.T) {
	sc := NewStatCard("Latency", "42ms")
	sc.SetDelta("+5%", true)
	s := sc.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H != 3 { t.Errorf("H = %d, want 3", s.H) }
	if s.W < 8 { t.Errorf("W = %d, too small", s.W) }
}

func TestP397_StatCard_Paint(t *testing.T) {
	sc := NewStatCard("Latency", "42ms")
	sc.SetDelta("+5%", true)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 3})
	buf := buffer.NewBuffer(12, 3)
	sc.Paint(buf)
	// Row 0: label at x=1
	c := buf.GetCell(1, 0)
	if c.Rune != 'L' { t.Errorf("label cell = %q, want 'L'", string(c.Rune)) }
	// Row 1: value
	c = buf.GetCell(1, 1)
	if c.Rune != '4' { t.Errorf("value cell = %q, want '4'", string(c.Rune)) }
	// Row 2: delta
	c = buf.GetCell(1, 2)
	if c.Rune != '+' { t.Errorf("delta cell = %q, want '+'", string(c.Rune)) }
}

func TestP397_StatCard_Paint_NoDelta(t *testing.T) {
	sc := NewStatCard("Count", "100")
	sc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	sc.Paint(buf) // row 2 should be empty
}

func TestP397_StatCard_Paint_NegativeDelta(t *testing.T) {
	sc := NewStatCard("Errors", "5")
	sc.SetDelta("-3%", false)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	sc.Paint(buf)
	c := buf.GetCell(1, 2)
	if c.Rune != '-' { t.Errorf("delta cell = %q, want '-'", string(c.Rune)) }
}

func TestP397_StatCard_Paint_ZeroBounds(t *testing.T) {
	sc := NewStatCard("X", "1")
	sc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	sc.Paint(buf)
}

func TestP397_StatCard_FromInt(t *testing.T) {
	sc := StatCardFromInt("Count", 42)
	if sc.Value() != "42" { t.Errorf("Value = %q", sc.Value()) }
}

func TestP397_StatCard_FromFloat(t *testing.T) {
	sc := StatCardFromFloat("Rate", 3.14)
	if sc.Value() != "3.14" { t.Errorf("Value = %q", sc.Value()) }
}

func TestP397_StatCard_Concurrent(t *testing.T) {
	sc := NewStatCard("X", "1")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			sc.SetValue("2")
			sc.SetDelta("+1%", true)
		}
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = sc.Value() }
	<-done
}

func TestP397_StatCard_SatisfiesComponent(t *testing.T) {
	var _ Component = (*StatCard)(nil)
}

func BenchmarkP397_StatCard_Paint(b *testing.B) {
	sc := NewStatCard("Tokens/sec", "1,234")
	sc.SetDelta("+12%", true)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { sc.Paint(buf) }
}
