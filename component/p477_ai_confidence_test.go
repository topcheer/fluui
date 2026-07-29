package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestAIConfidenceBarBasic(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetConfidence(85.5)
	if cb.Confidence() != 85.5 {
		t.Errorf("Confidence = %f, want 85.5", cb.Confidence())
	}
}

func TestAIConfidenceBarClamp(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetConfidence(150)
	if cb.Confidence() != 100 {
		t.Errorf("Clamped = %f, want 100", cb.Confidence())
	}
	cb.SetConfidence(-10)
	if cb.Confidence() != 0 {
		t.Errorf("Clamped = %f, want 0", cb.Confidence())
	}
}

func TestAIConfidenceBarLabel(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetLabel("Prediction")
	if cb.Label() != "Prediction" {
		t.Errorf("Label = %q, want Prediction", cb.Label())
	}
}

func TestAIConfidenceBarLevels(t *testing.T) {
	cb := NewAIConfidenceBar()

	cb.SetConfidence(20)
	if level := cb.confidenceLevelLocked(); level != 0 {
		t.Errorf("Level at 20%% = %d, want 0 (low)", level)
	}
	cb.SetConfidence(50)
	if level := cb.confidenceLevelLocked(); level != 1 {
		t.Errorf("Level at 50%% = %d, want 1 (medium)", level)
	}
	cb.SetConfidence(80)
	if level := cb.confidenceLevelLocked(); level != 2 {
		t.Errorf("Level at 80%% = %d, want 2 (high)", level)
	}
}

func TestAIConfidenceBarMeasure(t *testing.T) {
	cb := NewAIConfidenceBar()
	s := cb.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestAIConfidenceBarPaint(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetConfidence(75)
	cb.SetBarWidth(10)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)

	// Should have filled bar cells
	filledCount := 0
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '█' {
			filledCount++
		}
	}
	// 75% of 10 = 7 (approximately)
	if filledCount < 6 || filledCount > 8 {
		t.Errorf("filled cells = %d, want ~7", filledCount)
	}
}

func TestAIConfidenceBarPaintColors(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetBarWidth(10)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})

	// Low confidence → red
	cb.SetConfidence(20)
	buf1 := buffer.NewBuffer(40, 1)
	cb.Paint(buf1)
	// Find first filled cell
	lowColor := buffer.RGB(0, 0, 0)
	for x := 0; x < 40; x++ {
		if buf1.GetCell(x, 0).Rune == '█' {
			lowColor = buf1.GetCell(x, 0).Fg
			break
		}
	}

	// High confidence → green
	cb.SetConfidence(90)
	buf2 := buffer.NewBuffer(40, 1)
	cb.Paint(buf2)
	highColor := buffer.RGB(0, 0, 0)
	for x := 0; x < 40; x++ {
		if buf2.GetCell(x, 0).Rune == '█' {
			highColor = buf2.GetCell(x, 0).Fg
			break
		}
	}

	if lowColor.Equal(highColor) {
		t.Error("expected different colors for low vs high confidence")
	}
}

func TestAIConfidenceBarPaintZero(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetConfidence(0)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf) // should not panic
}

func TestAIConfidenceBarChildren(t *testing.T) {
	cb := NewAIConfidenceBar()
	if cb.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestAIConfidenceBarStyle(t *testing.T) {
	cb := NewAIConfidenceBar()
	cb.SetStyle(AIConfidenceStyle{
		Low:    buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Medium: buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		High:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Label:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		BarBg:  buffer.Style{Fg: buffer.RGB(50, 50, 50)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	cb.SetConfidence(50)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)
}

func BenchmarkPaintAIConfidenceBar(b *testing.B) {
	cb := NewAIConfidenceBar()
	cb.SetConfidence(85.5)
	cb.SetBarWidth(20)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.Paint(buf)
	}
}
