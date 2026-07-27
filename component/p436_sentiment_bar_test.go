package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestSentimentBar_New_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	if sb.Value() != 0.5 {
		t.Errorf("Value = %v, want 0.5", sb.Value())
	}
	if sb.Sentiment() != SentimentPositive {
		t.Error("expected positive")
	}
}

func TestSentimentBar_Clamp_P436(t *testing.T) {
	sb := NewSentimentBar(5.0)
	if sb.Value() != 1.0 {
		t.Errorf("Value = %v, want 1.0 (clamped)", sb.Value())
	}
	sb.SetValue(-10)
	if sb.Value() != -1.0 {
		t.Errorf("Value = %v, want -1.0", sb.Value())
	}
}

func TestSentimentBar_SentimentCategories_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	if sb.Sentiment() != SentimentPositive {
		t.Error("0.5 should be positive")
	}
	sb.SetValue(-0.5)
	if sb.Sentiment() != SentimentNegative {
		t.Error("-0.5 should be negative")
	}
	sb.SetValue(0.0)
	if sb.Sentiment() != SentimentNeutral {
		t.Error("0.0 should be neutral")
	}
	sb.SetValue(0.1)
	if sb.Sentiment() != SentimentNeutral {
		t.Error("0.1 should be neutral (below 0.15 threshold)")
	}
}

func TestSentimentBar_Label_P436(t *testing.T) {
	sb := NewSentimentBar(0.7)
	if sb.Label() != "positive" {
		t.Errorf("Label = %q", sb.Label())
	}
	sb.SetValue(-0.3)
	if sb.Label() != "negative" {
		t.Errorf("Label = %q", sb.Label())
	}
	sb.SetLabel("custom")
	if sb.Label() != "custom" {
		t.Errorf("Label = %q, want custom", sb.Label())
	}
}

func TestSentimentBar_Confidence_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	sb.SetConfidence(0.9)
	if sb.Confidence() != 0.9 {
		t.Errorf("Confidence = %v", sb.Confidence())
	}
	sb.SetConfidence(5.0)
	if sb.Confidence() != 1.0 {
		t.Error("should clamp to 1.0")
	}
}

func TestSentimentBar_Style_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	st := DefaultSentimentBarStyle()
	sb.SetStyle(st)
	if sb.Style().Positive.Fg != st.Positive.Fg {
		t.Error("style mismatch")
	}
}

func TestSentimentBar_Measure_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	sz := sb.Measure(Constraints{})
	if sz.H != 1 {
		t.Errorf("H = %v, want 1", sz.H)
	}
	if sz.W < 20 {
		t.Errorf("W = %v, want >= 20", sz.W)
	}
}

func TestSentimentBar_Paint_NoPanic_P436(t *testing.T) {
	sb := NewSentimentBar(0.72)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	sb.Paint(buf)
}

func TestSentimentBar_Paint_Negative_P436(t *testing.T) {
	sb := NewSentimentBar(-0.8)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	sb.Paint(buf)
}

func TestSentimentBar_Paint_ZeroBounds_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	sb.Paint(buf)
}

func TestSentimentBar_Fluent_P436(t *testing.T) {
	sb := NewSentimentBar(0.5).SetConfidence(0.8).SetLabel("good").SetShowPct(false)
	if sb.Confidence() != 0.8 || sb.Label() != "good" || sb.ShowPct() {
		t.Error("fluent chain failed")
	}
}

func TestSentimentBar_Children_P436(t *testing.T) {
	if NewSentimentBar(0).Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestSentimentBar_ShowPct_P436(t *testing.T) {
	sb := NewSentimentBar(0.5)
	if !sb.ShowPct() {
		t.Error("default should show pct")
	}
	sb.SetShowPct(false)
	if sb.ShowPct() {
		t.Error("should be false after SetShowPct(false)")
	}
}

func BenchmarkSentimentBar_Paint_P436(b *testing.B) {
	sb := NewSentimentBar(0.72)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb.Paint(buf)
	}
}

func BenchmarkSentimentBar_Measure_P436(b *testing.B) {
	sb := NewSentimentBar(0.5)
	cs := Constraints{MaxWidth: 40, MaxHeight: 1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb.Measure(cs)
	}
}
