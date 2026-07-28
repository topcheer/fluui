package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestThinkingTrace_New_P462(t *testing.T) {
	tt := NewThinkingTrace()
	if tt.State() != ThinkingIdle {
		t.Errorf("State = %v, want Idle", tt.State())
	}
	if tt.Text() != "" {
		t.Errorf("Text = %q", tt.Text())
	}
}

func TestThinkingTrace_Start_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	if tt.State() != ThinkingActive {
		t.Errorf("State = %v, want Active", tt.State())
	}
}

func TestThinkingTrace_Append_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("Analyzing...")
	tt.Append(" Step 1.")
	if tt.Text() != "Analyzing... Step 1." {
		t.Errorf("Text = %q", tt.Text())
	}
}

func TestThinkingTrace_Complete_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("Reasoning")
	tt.Complete()
	if tt.State() != ThinkingDone {
		t.Errorf("State = %v, want Done", tt.State())
	}
	if !tt.IsCollapsed() {
		t.Error("should be collapsed after Complete")
	}
}

func TestThinkingTrace_Collapse_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Complete()
	tt.SetCollapsed(false)
	if tt.IsCollapsed() {
		t.Error("should be expanded")
	}
	tt.SetCollapsed(true)
	if !tt.IsCollapsed() {
		t.Error("should be collapsed")
	}
}

func TestThinkingTrace_TickSpinner_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.TickSpinner() // should not panic
}

func TestThinkingTrace_SetStyle_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.SetStyle(DefaultThinkingTraceStyle())
}

func TestThinkingTrace_Measure_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("line1\nline2")
	sz := tt.Measure(Constraints{})
	if sz.H < 3 {
		t.Errorf("H = %d, want >= 3", sz.H)
	}
}

func TestThinkingTrace_Paint_Active_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("Let me think about this...")
	tt.TickSpinner()
	tt.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	tt.Paint(buf)
}

func TestThinkingTrace_Paint_Done_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("I analyzed the problem.")
	tt.Complete()
	tt.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	tt.Paint(buf)
}

func TestThinkingTrace_Paint_DoneExpanded_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("Reasoning text\nMultiple lines")
	tt.Complete()
	tt.SetCollapsed(false)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	tt.Paint(buf)
}

func TestThinkingTrace_Paint_Idle_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	tt.Paint(buf) // idle = no-op
}

func TestThinkingTrace_Paint_ZeroBounds_P462(t *testing.T) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tt.Paint(buf)
}

func TestThinkingTrace_Children_P462(t *testing.T) {
	if NewThinkingTrace().Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestFormatThinkingDuration_P462(t *testing.T) {
	// Just verify no panic and non-empty
	s := formatThinkingDuration(0)
	if s == "" {
		t.Error("should not be empty")
	}
}

func BenchmarkThinkingTrace_Paint_P462(b *testing.B) {
	tt := NewThinkingTrace()
	tt.Start()
	tt.Append("Let me analyze this step by step.\nFirst, checking input.\nThen processing.\nFinally, output.")
	tt.TickSpinner()
	tt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tt.Paint(buf)
	}
}
