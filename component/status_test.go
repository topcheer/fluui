package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewStatusIndicator_Defaults(t *testing.T) {
	s := NewStatusIndicator()
	if s.IsRunning() {
		t.Error("IsRunning should be false initially")
	}
	if s.Message() != "" {
		t.Errorf("Message = %q, want empty", s.Message())
	}
	if s.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestStatusIndicator_SetMessage(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("Loading data...")
	if s.Message() != "Loading data..." {
		t.Errorf("Message = %q, want 'Loading data...'", s.Message())
	}
	s.SetMessage("")
	if s.Message() != "" {
		t.Errorf("Message = %q, want empty", s.Message())
	}
}

func TestStatusIndicator_Start(t *testing.T) {
	s := NewStatusIndicator()
	s.Start()
	if !s.IsRunning() {
		t.Error("IsRunning = false after Start()")
	}
}

func TestStatusIndicator_Stop(t *testing.T) {
	s := NewStatusIndicator()
	s.Start()
	s.Stop()
	if s.IsRunning() {
		t.Error("IsRunning = true after Stop()")
	}
}

func TestStatusIndicator_StartStop(t *testing.T) {
	s := NewStatusIndicator()
	if s.IsRunning() {
		t.Error("should not be running initially")
	}
	s.Start()
	if !s.IsRunning() {
		t.Error("should be running after Start()")
	}
	s.Stop()
	if s.IsRunning() {
		t.Error("should not be running after Stop()")
	}
}

func TestStatusIndicator_CurrentFrame_WhenStopped(t *testing.T) {
	s := NewStatusIndicator()
	frame := s.CurrentFrame()
	if frame != " " {
		t.Errorf("CurrentFrame = %q, want ' ' when stopped", frame)
	}
}

func TestStatusIndicator_CurrentFrame_WhenRunning(t *testing.T) {
	s := NewStatusIndicator()
	s.Start()
	frame := s.CurrentFrame()
	if frame == " " || frame == "" {
		t.Errorf("CurrentFrame = %q, expected a spinner character when running", frame)
	}
}

func TestStatusIndicator_Update_WhenStopped(t *testing.T) {
	s := NewStatusIndicator()
	result := s.Update(100 * time.Millisecond)
	if result {
		t.Error("Update should return false when not running")
	}
}

func TestStatusIndicator_Update_WhenRunning(t *testing.T) {
	s := NewStatusIndicator()
	s.Start()
	result := s.Update(50 * time.Millisecond)
	if !result {
		t.Error("Update should return true when running")
	}
}

func TestStatusIndicator_Update_AdvancesSpinner(t *testing.T) {
	s := NewStatusIndicator()
	s.Start()
	frame0 := s.CurrentFrame()
	// Update with enough time to advance at least one frame (default 100ms interval)
	s.Update(150 * time.Millisecond)
	frame1 := s.CurrentFrame()
	if frame0 == frame1 {
		t.Errorf("spinner did not advance: frame0=%q frame1=%q", frame0, frame1)
	}
}

func TestStatusIndicator_Measure(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("Hello")
	size := s.Measure(Constraints{MaxWidth: 50})
	// Width = 2 (spinner + gap) + 5 ("Hello") = 7
	if size.W != 7 {
		t.Errorf("W = %d, want 7", size.W)
	}
	if size.H != 1 {
		t.Errorf("H = %d, want 1", size.H)
	}
}

func TestStatusIndicator_Measure_EmptyMessage(t *testing.T) {
	s := NewStatusIndicator()
	size := s.Measure(Constraints{MaxWidth: 50})
	if size.W != 2 {
		t.Errorf("W = %d, want 2", size.W)
	}
}

func TestStatusIndicator_Measure_ClampedWidth(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("A very long message that exceeds the width")
	size := s.Measure(Constraints{MaxWidth: 20})
	if size.W > 20 {
		t.Errorf("W = %d, should be clamped to 20", size.W)
	}
}

func TestStatusIndicator_Paint_Running(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("Processing...")
	s.Start()
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(40, 3)
	s.Paint(buf)

	// Spinner should be at position 0
	c := buf.GetCell(0, 0)
	if c.Rune == ' ' || c.Rune == 0 {
		t.Error("spinner cell should not be empty when running")
	}
	// Message should start at position 2
	msgCell := buf.GetCell(2, 0)
	if msgCell.Rune != 'P' {
		t.Errorf("message cell Rune = %q, want 'P'", msgCell.Rune)
	}
}

func TestStatusIndicator_Paint_Stopped(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("Idle")
	s.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(30, 3)
	s.Paint(buf)

	// When stopped, spinner shows space
	c := buf.GetCell(0, 0)
	if c.Rune != ' ' {
		t.Errorf("spinner cell Rune = %q, want ' ' when stopped", c.Rune)
	}
	// Message should still be drawn
	msgCell := buf.GetCell(2, 0)
	if msgCell.Rune != 'I' {
		t.Errorf("message cell Rune = %q, want 'I'", msgCell.Rune)
	}
}

func TestStatusIndicator_Paint_ZeroBounds(t *testing.T) {
	s := NewStatusIndicator()
	s.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 1)
	s.Paint(buf) // should not panic
}

func TestStatusIndicator_Paint_LongMessage(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("This is a very long message that exceeds the available width")
	s.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(20, 3)
	s.Paint(buf) // should not panic, message truncated
}

func TestStatusIndicator_SetSpinnerStyle(t *testing.T) {
	s := NewStatusIndicator()
	style := buffer.Style{Fg: buffer.RGB(255, 0, 0), Flags: buffer.Bold}
	s.SetSpinnerStyle(style)
	// No panic = pass
}

func TestStatusIndicator_SetStyle(t *testing.T) {
	s := NewStatusIndicator()
	style := buffer.Style{Fg: buffer.RGB(100, 200, 255)}
	s.SetStyle(style)
	// No panic = pass
}

func TestStatusIndicator_SetSpinnerStyleName(t *testing.T) {
	s := NewStatusIndicator()
	s.SetSpinnerStyleName("arc")
	s.Start()
	frame := s.CurrentFrame()
	if frame == " " || frame == "" {
		t.Error("expected arc spinner character")
	}
}

func TestStatusIndicator_SetSpinnerStyleName_Unknown(t *testing.T) {
	s := NewStatusIndicator()
	s.SetSpinnerStyleName("nonexistent")
	s.Start()
	frame := s.CurrentFrame()
	if frame == " " || frame == "" {
		t.Error("unknown style should fall back to dots")
	}
}

func TestStatusIndicator_ConcurrentAccess(t *testing.T) {
	s := NewStatusIndicator()
	s.SetMessage("test")
	s.Start()
	done := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			s.SetMessage("msg")
			s.Update(10 * time.Millisecond)
			_ = s.Message()
			_ = s.IsRunning()
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		_ = s.Message()
		_ = s.IsRunning()
		_ = s.CurrentFrame()
	}

	<-done
}
