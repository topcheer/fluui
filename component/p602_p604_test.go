package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── NetworkLatencyMap Tests ───

func TestNetworkLatencyMapBasic(t *testing.T) {
	nlm := NewNetworkLatencyMap()
	nlm.SetRegions("US", "EU", "AS")
	nlm.SetLatency(0, 1, 85)
	if n := nlm.RegionCount(); n != 3 {
		t.Errorf("RegionCount = %d, want 3", n)
	}
}

func TestNetworkLatencyMapOverflow(t *testing.T) {
	nlm := NewNetworkLatencyMap()
	nlm.SetRegions("A", "B", "C", "D", "E", "F", "G", "H", "I", "J")
	if n := nlm.RegionCount(); n != latencyMapMaxRegions {
		t.Errorf("RegionCount = %d, want %d (capped)", n, latencyMapMaxRegions)
	}
}

func TestNetworkLatencyMapPaint(t *testing.T) {
	nlm := NewNetworkLatencyMap()
	nlm.SetRegions("US", "EU")
	nlm.SetLatency(0, 1, 85)
	nlm.SetLatency(1, 0, 90)
	nlm.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	nlm.Paint(buf)
	// Header should have "EU"
	hasContent := false
	for i := 0; i < 30; i++ {
		if buf.GetCell(i, 0).Rune == 'E' {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint should show region headers")
	}
}

func TestNetworkLatencyMapChildren(t *testing.T) {
	nlm := NewNetworkLatencyMap()
	if c := nlm.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestNetworkLatencyMapStyle(t *testing.T) {
	nlm := NewNetworkLatencyMap()
	nlm.SetStyle(LatencyCellStyle{
		Fast:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Medium: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Slow:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Local:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	nlm.SetRegions("A", "B")
	nlm.SetLatency(0, 1, 200)
	buf := buffer.NewBuffer(30, 4)
	nlm.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	nlm.Paint(buf)
}

// ─── AIConfidenceGauge Tests ───

func TestAIConfidenceGaugeBasic(t *testing.T) {
	cg := NewAIConfidenceGauge()
	cg.SetValue(85)
	if v := cg.Value(); v != 85 {
		t.Errorf("Value = %d, want 85", v)
	}
}

func TestAIConfidenceGaugeZero(t *testing.T) {
	cg := NewAIConfidenceGauge()
	cg.SetValue(0)
	if v := cg.Value(); v != 0 {
		t.Errorf("Value = %d, want 0", v)
	}
}

func TestAIConfidenceGaugeClamp(t *testing.T) {
	cg := NewAIConfidenceGauge()
	cg.SetValue(-10)
	if v := cg.Value(); v != 0 {
		t.Errorf("Value = %d, want 0 (clamped)", v)
	}
	cg.SetValue(200)
	if v := cg.Value(); v != 100 {
		t.Errorf("Value = %d, want 100 (clamped)", v)
	}
}

func TestAIConfidenceGaugeColorLevels(t *testing.T) {
	cg := NewAIConfidenceGauge()
	cg.SetValue(80)
	if cg.curStyle.Fg != cg.style.High.Fg {
		t.Error("Expected High style for >= 70")
	}
	cg.SetValue(50)
	if cg.curStyle.Fg != cg.style.Medium.Fg {
		t.Error("Expected Medium style for 40-69")
	}
	cg.SetValue(20)
	if cg.curStyle.Fg != cg.style.Low.Fg {
		t.Error("Expected Low style for < 40")
	}
}

func TestAIConfidenceGaugePaint(t *testing.T) {
	cg := NewAIConfidenceGauge()
	cg.SetValue(75)
	cg.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 2})
	buf := buffer.NewBuffer(12, 2)
	cg.Paint(buf)
	hasContent := false
	for i := 0; i < 12; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestAIConfidenceGaugeChildren(t *testing.T) {
	cg := NewAIConfidenceGauge()
	if c := cg.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIConfidenceGaugeStyle(t *testing.T) {
	cg := NewAIConfidenceGauge()
	cg.SetStyle(AIConfidenceGaugeStyle{
		High:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Medium: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Low:    buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Arc:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	cg.SetValue(45)
	buf := buffer.NewBuffer(12, 2)
	cg.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 2})
	cg.Paint(buf)
}

// ─── CompactTimeline Tests ───

func TestCompactTimelineBasic(t *testing.T) {
	ct := NewCompactTimeline()
	ct.SetRange(0, 100)
	ct.AddEvent(25, 0)
	ct.AddEvent(50, 1)
	if n := ct.EventCount(); n != 2 {
		t.Errorf("EventCount = %d, want 2", n)
	}
}

func TestCompactTimelineOverflow(t *testing.T) {
	ct := NewCompactTimeline()
	for i := 0; i < compactTimelineMaxEvents+5; i++ {
		ct.AddEvent(i, 0)
	}
	if n := ct.EventCount(); n != compactTimelineMaxEvents {
		t.Errorf("EventCount = %d, want %d (capped)", n, compactTimelineMaxEvents)
	}
}

func TestCompactTimelineClear(t *testing.T) {
	ct := NewCompactTimeline()
	ct.AddEvent(10, 0)
	ct.Clear()
	if n := ct.EventCount(); n != 0 {
		t.Errorf("EventCount after Clear = %d, want 0", n)
	}
}

func TestCompactTimelineRange(t *testing.T) {
	ct := NewCompactTimeline()
	ct.SetRange(50, 50) // min==max
	if ct.rangeMax != 51 {
		t.Errorf("rangeMax = %d, want 51 (clamped)", ct.rangeMax)
	}
}

func TestCompactTimelinePaint(t *testing.T) {
	ct := NewCompactTimeline()
	ct.SetRange(0, 100)
	ct.AddEvent(25, 0)
	ct.AddEvent(75, 2)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	ct.Paint(buf)
	// Should have line characters
	hasLine := false
	for i := 0; i < 40; i++ {
		if buf.GetCell(i, 0).Rune == '─' {
			hasLine = true
			break
		}
	}
	if !hasLine {
		t.Error("Paint should show timeline line")
	}
}

func TestCompactTimelineChildren(t *testing.T) {
	ct := NewCompactTimeline()
	if c := ct.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestCompactTimelineStyle(t *testing.T) {
	ct := NewCompactTimeline()
	ct.SetStyle(CompactTimelineStyle{
		Info:  buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Warn:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Error: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Line:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	ct.AddEvent(50, 1)
	buf := buffer.NewBuffer(40, 1)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	ct.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintNetworkLatencyMap(b *testing.B) {
	nlm := NewNetworkLatencyMap()
	nlm.SetRegions("US", "EU", "AS")
	nlm.SetLatency(0, 1, 85)
	nlm.SetLatency(0, 2, 200)
	nlm.SetLatency(1, 2, 150)
	nlm.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nlm.Paint(buf)
	}
}

func BenchmarkPaintAIConfidenceGauge(b *testing.B) {
	cg := NewAIConfidenceGauge()
	cg.SetValue(85)
	cg.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 2})
	buf := buffer.NewBuffer(12, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cg.Paint(buf)
	}
}

func BenchmarkPaintCompactTimeline(b *testing.B) {
	ct := NewCompactTimeline()
	ct.SetRange(0, 100)
	ct.AddEvent(10, 0)
	ct.AddEvent(30, 0)
	ct.AddEvent(60, 1)
	ct.AddEvent(90, 2)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ct.Paint(buf)
	}
}
