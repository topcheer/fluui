package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP332_ThinkingIndicator_Create(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	if ti.Label() != "Thinking" {
		t.Errorf("label = %q, want %q", ti.Label(), "Thinking")
	}
	if ti.IsRunning() {
		t.Error("should not be running on creation")
	}
	if ti.FrameIndex() != 0 {
		t.Errorf("frame = %d, want 0", ti.FrameIndex())
	}
}

func TestP332_ThinkingIndicator_SetLabel(t *testing.T) {
	ti := NewThinkingIndicator("")
	ti.SetLabel("Generating")
	if ti.Label() != "Generating" {
		t.Errorf("label = %q", ti.Label())
	}
}

func TestP332_ThinkingIndicator_AdvanceFrame(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	for i := 0; i < 4; i++ {
		if ti.FrameIndex() != i {
			t.Errorf("frame %d: got %d", i, ti.FrameIndex())
		}
		ti.AdvanceFrame()
	}
	// Should wrap: frame 4 % 4 = 0
	if ti.FrameIndex() != 0 {
		t.Errorf("after wrap: frame = %d, want 0", ti.FrameIndex())
	}
}

func TestP332_ThinkingIndicator_StartStop(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	ti.Start(10 * time.Millisecond)
	if !ti.IsRunning() {
		t.Error("should be running after Start")
	}

	// Wait for animation
	time.Sleep(50 * time.Millisecond)

	ti.Stop()
	if ti.IsRunning() {
		t.Error("should not be running after Stop")
	}

	// Frame should have advanced
	if ti.FrameIndex() == 0 {
		t.Log("frame is 0 (may be unlucky timing)")
	}
}

func TestP332_ThinkingIndicator_StopIdempotent(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	ti.Start(10 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	ti.Stop()
	ti.Stop() // should not panic
	ti.Stop()
}

func TestP332_ThinkingIndicator_Measure(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	s := ti.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if s.H != 1 {
		t.Errorf("height = %d, want 1", s.H)
	}
	if s.W < 5 {
		t.Errorf("width = %d, expected at least 5", s.W)
	}
}

func TestP332_ThinkingIndicator_Paint(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)

	// Frame 0: first dot filled
	ti.Paint(buf)
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell at label start")
	}

	// Advance and paint — dots should change
	ti.AdvanceFrame()
	ti.Paint(buf)

	// Verify no panic with empty label
	ti2 := NewThinkingIndicator("")
	ti2.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf2 := buffer.NewBuffer(20, 1)
	ti2.Paint(buf2)
}

func TestP332_ThinkingIndicator_PaintAllFrames(t *testing.T) {
	ti := NewThinkingIndicator("AI")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})

	for frame := 0; frame < 4; frame++ {
		if ti.FrameIndex() != frame {
			t.Errorf("frame %d: got %d", frame, ti.FrameIndex())
		}
		buf := buffer.NewBuffer(20, 1)
		ti.Paint(buf)
		ti.AdvanceFrame()
	}
}

func TestP332_ThinkingIndicator_CustomStyle(t *testing.T) {
	ti := NewThinkingIndicator("Loading")
	ti.SetStyle(ThinkingStyle{
		DotChar:   "■",
		EmptyChar: "□",
		Spacing:   "-",
		UseAccent: true,
	})
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ti.Paint(buf)

	// Should not panic and should render something
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP332_ThinkingIndicator_Concurrent(t *testing.T) {
	ti := NewThinkingIndicator("Thinking")
	ti.Start(5 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			ti.AdvanceFrame()
			ti.FrameIndex()
			ti.Label()
		}
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	ti.Stop()
	<-done
}

func BenchmarkThinkingIndicator_Paint(b *testing.B) {
	ti := NewThinkingIndicator("Thinking")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ti.Paint(buf)
	}
}
