package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestResponseInspectorBasic(t *testing.T) {
	ri := NewResponseInspector()
	ri.SetModel("gpt-4o")
	ri.SetLatency(450 * time.Millisecond)
	ri.SetTokens(120, 350)
	ri.SetFinishReason(FinishStop)
	ri.SetTemperature(0.7)

	if ri.Model() != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", ri.Model())
	}
	if ri.Latency() != 450*time.Millisecond {
		t.Errorf("Latency = %v, want 450ms", ri.Latency())
	}
	if ri.InputTokens() != 120 {
		t.Errorf("InputTokens = %d, want 120", ri.InputTokens())
	}
	if ri.OutputTokens() != 350 {
		t.Errorf("OutputTokens = %d, want 350", ri.OutputTokens())
	}
	if ri.TotalTokens() != 470 {
		t.Errorf("TotalTokens = %d, want 470", ri.TotalTokens())
	}
	if ri.FinishReason() != FinishStop {
		t.Errorf("FinishReason = %q, want stop", ri.FinishReason())
	}
	if ri.Temperature() != 0.7 {
		t.Errorf("Temperature = %f, want 0.7", ri.Temperature())
	}
}

func TestResponseInspectorMeasure(t *testing.T) {
	ri := NewResponseInspector()
	s := ri.Measure(Constraints{})
	if s.W < 20 {
		t.Errorf("W = %d, want >= 20", s.W)
	}
	if s.H < 8 {
		t.Errorf("H = %d, want >= 8", s.H)
	}
}

func TestResponseInspectorPaint(t *testing.T) {
	ri := NewResponseInspector()
	ri.SetModel("claude-3.5")
	ri.SetLatency(120 * time.Millisecond)
	ri.SetTokens(50, 200)
	ri.SetFinishReason(FinishToolCalls)
	ri.SetTemperature(0.5)
	ri.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	buf := buffer.NewBuffer(40, 10)
	ri.Paint(buf)

	// Check top-left border corner
	cell := buf.GetCell(0, 0)
	if cell.Rune != '┌' {
		t.Errorf("top-left corner = %q, want ┌", cell.Rune)
	}
	// Check header exists
	foundHeader := false
	for x := 0; x < 40; x++ {
		c := buf.GetCell(x, 1)
		if c.Rune == 'R' {
			foundHeader = true
			break
		}
	}
	if !foundHeader {
		t.Error("header not found in painted buffer")
	}
}

func TestResponseInspectorChildren(t *testing.T) {
	ri := NewResponseInspector()
	if ri.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestResponseInspectorStyle(t *testing.T) {
	ri := NewResponseInspector()
	custom := ResponseInspectorStyle{
		Label:  buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Value:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Header: buffer.Style{Fg: buffer.RGB(0, 0, 255), Flags: buffer.Bold},
		Border: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	}
	ri.SetStyle(custom)
	// Just verify it doesn't panic
	ri.SetBounds(Rect{X: 0, Y: 0, W: 34, H: 8})
	buf := buffer.NewBuffer(34, 8)
	ri.Paint(buf)
}

func TestFormatInspectorDuration(t *testing.T) {
	tests := []struct {
		d   time.Duration
		got string
	}{
		{500 * time.Nanosecond, "500ns"},
		{500 * time.Microsecond, "500us"},
		{500 * time.Millisecond, "500ms"},
		{2 * time.Second, "2s"},
	}
	for _, tt := range tests {
		got := formatInspectorDuration(tt.d)
		if got != tt.got {
			t.Errorf("formatInspectorDuration(%v) = %q, want %q", tt.d, got, tt.got)
		}
	}
}
