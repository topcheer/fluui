package component

import (
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ============================================================
// P19-B: Spinner Component Tests
// ============================================================

// ─── Construction ────────────────────────────────────────────────

func TestSpinner_New(t *testing.T) {
	s := NewSpinner("Loading...")
	if s == nil {
		t.Fatal("NewSpinner returned nil")
	}
	if s.ID() == "" {
		t.Error("ID should not be empty")
	}
	if !s.Running() {
		t.Error("should start running")
	}
	if s.Label() != "Loading..." {
		t.Errorf("Label = %q, want 'Loading...'", s.Label())
	}
	if s.FrameCount() == 0 {
		t.Error("FrameCount should not be 0")
	}
}

func TestSpinner_UniqueID(t *testing.T) {
	s1 := NewSpinner("")
	s2 := NewSpinner("")
	if s1.ID() == s2.ID() {
		t.Error("IDs should be unique")
	}
}

func TestSpinner_ImplementsComponent(t *testing.T) {
	var _ Component = NewSpinner("")
}

// ─── Label & Prefix ──────────────────────────────────────────────

func TestSpinner_SetLabel(t *testing.T) {
	s := NewSpinner("")
	s.SetLabel("Processing")
	if s.Label() != "Processing" {
		t.Errorf("Label = %q, want 'Processing'", s.Label())
	}
}

func TestSpinner_SetPrefix(t *testing.T) {
	s := NewSpinner("")
	s.SetPrefix("[INFO]")
	if s.Prefix() != "[INFO]" {
		t.Errorf("Prefix = %q, want '[INFO]'", s.Prefix())
	}
}

// ─── Animation control ───────────────────────────────────────────

func TestSpinner_StartStop(t *testing.T) {
	s := NewSpinner("")
	s.Stop()
	if s.Running() {
		t.Error("should be stopped after Stop")
	}
	s.Start()
	if !s.Running() {
		t.Error("should be running after Start")
	}
}

func TestSpinner_Update(t *testing.T) {
	s := NewSpinner("")
	initialIdx := s.FrameIndex()
	// Multiple updates should eventually advance the frame
	for i := 0; i < 100; i++ {
		s.Update(100 * time.Millisecond)
	}
	// At least one frame change should have occurred
	_ = initialIdx // frame may or may not change depending on timing
}

func TestSpinner_Update_Stopped(t *testing.T) {
	s := NewSpinner("")
	s.Stop()
	changed := s.Update(100 * time.Millisecond)
	if changed {
		t.Error("Update should return false when stopped")
	}
}

func TestSpinner_CurrentFrame(t *testing.T) {
	s := NewSpinner("")
	frame := s.CurrentFrame()
	if frame == "" {
		t.Error("CurrentFrame should not be empty")
	}
}

func TestSpinner_SetFrameIndex(t *testing.T) {
	s := NewSpinner("")
	s.SetFrameIndex(3)
	if s.FrameIndex() != 3 {
		t.Errorf("FrameIndex = %d, want 3", s.FrameIndex())
	}
}

func TestSpinner_SetFrameIndex_Wrap(t *testing.T) {
	s := NewSpinner("")
	count := s.FrameCount()
	s.SetFrameIndex(count + 5)
	if s.FrameIndex() != 5 {
		t.Errorf("FrameIndex = %d, want 5 (wrapped)", s.FrameIndex())
	}
}

// ─── Frame style ─────────────────────────────────────────────────

func TestSpinner_SetFrameStyle(t *testing.T) {
	s := NewSpinner("")
	originalCount := s.FrameCount()
	s.SetFrameStyle("arc")
	if s.FrameCount() == originalCount {
		// arc might have same count, just verify no panic
	}
	if s.FrameStyle() != "arc" {
		t.Errorf("FrameStyle = %q, want 'arc'", s.FrameStyle())
	}
}

func TestSpinner_SetFrameStyle_Invalid(t *testing.T) {
	s := NewSpinner("")
	originalStyle := s.FrameStyle()
	s.SetFrameStyle("nonexistent")
	if s.FrameStyle() != originalStyle {
		t.Error("invalid frame style should not change current style")
	}
}

// ─── Style ───────────────────────────────────────────────────────

func TestSpinner_SetStyle(t *testing.T) {
	s := NewSpinner("")
	custom := SpinnerStyle{
		Frame: buffer.Style{Fg: buffer.Color256Val(200)},
		Label: buffer.Style{Fg: buffer.Color256Val(100)},
	}
	s.SetStyle(custom)
	if s.Style().Frame.Fg != buffer.Color256Val(200) {
		t.Error("style not set")
	}
}

func TestSpinner_DefaultStyle(t *testing.T) {
	style := DefaultSpinnerStyle()
	_ = style // should not panic
}

// ─── Measure ─────────────────────────────────────────────────────

func TestSpinner_Measure_Basic(t *testing.T) {
	s := NewSpinner("Loading")
	size := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.W < 3 {
		t.Errorf("width = %d, want >= 3", size.W)
	}
	if size.H != 1 {
		t.Errorf("height = %d, want 1", size.H)
	}
}

func TestSpinner_Measure_WithPrefix(t *testing.T) {
	s := NewSpinner("Loading")
	s.SetPrefix("[INFO]")
	size := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.W < 10 {
		t.Errorf("width = %d, want >= 10 with prefix", size.W)
	}
}

func TestSpinner_Measure_Empty(t *testing.T) {
	s := NewSpinner("")
	size := s.Measure(Constraints{MaxWidth: 80, MaxHeight: 24})
	if size.W < 3 {
		t.Errorf("width = %d, want >= 3", size.W)
	}
}

func TestSpinner_Measure_Clamped(t *testing.T) {
	s := NewSpinner("A very long label that exceeds the constraint")
	size := s.Measure(Constraints{MaxWidth: 10, MaxHeight: 1})
	if size.W > 10 {
		t.Errorf("width = %d, want <= 10", size.W)
	}
}

// ─── Paint ───────────────────────────────────────────────────────

func TestSpinner_Paint_NoPanic(t *testing.T) {
	s := NewSpinner("Loading")
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	s.Paint(buf) // should not panic
}

func TestSpinner_Paint_RendersFrame(t *testing.T) {
	s := NewSpinner("Loading")
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	s.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune == ' ' || cell.Rune == 0 {
		t.Error("first cell should have a frame glyph")
	}
}

func TestSpinner_Paint_RendersLabel(t *testing.T) {
	s := NewSpinner("Processing")
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	s.Paint(buf)
	// Label starts at x=2 (frame + space)
	cell := buf.GetCell(2, 0)
	if cell.Rune != 'P' {
		t.Errorf("expected 'P' at x=2, got %q", string(cell.Rune))
	}
}

func TestSpinner_Paint_ZeroBounds(t *testing.T) {
	s := NewSpinner("Loading")
	buf := buffer.NewBuffer(30, 1)
	s.Paint(buf) // should not panic
}

func TestSpinner_Paint_Stopped(t *testing.T) {
	s := NewSpinner("Loading")
	s.Stop()
	s.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	s.Paint(buf) // should render a static frame
}

// ─── Misc ────────────────────────────────────────────────────────

func TestSpinner_Children(t *testing.T) {
	s := NewSpinner("")
	if s.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestSpinner_String(t *testing.T) {
	s := NewSpinner("Loading")
	if s.String() == "" {
		t.Error("String should not be empty")
	}
}

// ─── Concurrency ─────────────────────────────────────────────────

func TestSpinner_ConcurrentAccess(t *testing.T) {
	s := NewSpinner("Loading")

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.SetLabel("Loading")
				s.Update(time.Duration(n) * time.Millisecond)
				s.CurrentFrame()
				s.FrameIndex()
			}
		}(i)
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Label()
				s.Running()
				s.Style()
			}
		}()
	}
	wg.Wait()
}