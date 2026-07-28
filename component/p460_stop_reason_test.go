package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestStopReasonBadge_New_P460(t *testing.T) {
	sr := NewStopReasonBadge(StopReasonStop)
	if sr.Reason() != StopReasonStop {
		t.Errorf("Reason = %v", sr.Reason())
	}
	if sr.ReasonText() != "stop" {
		t.Errorf("ReasonText = %q", sr.ReasonText())
	}
}

func TestStopReasonBadge_AllReasons_P460(t *testing.T) {
	tests := []struct {
		reason StopReason
		text   string
		icon   rune
	}{
		{StopReasonStop, "stop", '✓'},
		{StopReasonLength, "max tokens", '⋯'},
		{StopReasonToolCall, "tool call", '⚡'},
		{StopReasonContentFilter, "filtered", '⊘'},
		{StopReasonError, "error", '✗'},
		{StopReasonNone, "streaming", '●'},
	}
	for _, tc := range tests {
		sr := NewStopReasonBadge(tc.reason)
		if sr.ReasonText() != tc.text {
			t.Errorf("Reason(%v) text = %q, want %q", tc.reason, sr.ReasonText(), tc.text)
		}
		if sr.ReasonIcon() != tc.icon {
			t.Errorf("Reason(%v) icon = %q, want %q", tc.reason, sr.ReasonIcon(), tc.icon)
		}
	}
}

func TestStopReasonBadge_SetReason_P460(t *testing.T) {
	sr := NewStopReasonBadge(StopReasonNone)
	sr.SetReason(StopReasonError)
	if sr.Reason() != StopReasonError {
		t.Errorf("Reason = %v", sr.Reason())
	}
}

func TestStopReasonBadge_Style_P460(t *testing.T) {
	sr := NewStopReasonBadge(StopReasonStop)
	sr.SetStyle(DefaultStopReasonStyle())
}

func TestStopReasonBadge_Measure_P460(t *testing.T) {
	sr := NewStopReasonBadge(StopReasonStop)
	sz := sr.Measure(Constraints{})
	if sz.H != 1 || sz.W < 10 {
		t.Errorf("size = %v", sz)
	}
}

func TestStopReasonBadge_Paint_All_P460(t *testing.T) {
	for _, reason := range []StopReason{StopReasonNone, StopReasonStop, StopReasonLength, StopReasonToolCall, StopReasonContentFilter, StopReasonError} {
		sr := NewStopReasonBadge(reason)
		sr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
		buf := buffer.NewBuffer(20, 1)
		sr.Paint(buf)
	}
}

func TestStopReasonBadge_Paint_ZeroBounds_P460(t *testing.T) {
	sr := NewStopReasonBadge(StopReasonStop)
	sr.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	sr.Paint(buf)
}

func TestStopReasonBadge_Children_P460(t *testing.T) {
	if NewStopReasonBadge(StopReasonStop).Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkStopReasonBadge_Paint_P460(b *testing.B) {
	sr := NewStopReasonBadge(StopReasonStop)
	sr.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sr.Paint(buf)
	}
}
