package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewProgressBar_Defaults(t *testing.T) {
	p := NewProgressBar()
	if p.Progress() != 0 {
		t.Errorf("Progress = %v, want 0", p.Progress())
	}
	if p.Mode() != ProgressDeterminate {
		t.Errorf("Mode = %v, want ProgressDeterminate", p.Mode())
	}
	if p.Label() != "" {
		t.Errorf("Label = %q, want empty", p.Label())
	}
	if p.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestProgressBar_SetProgress(t *testing.T) {
	p := NewProgressBar()
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 0},
		{50, 50},
		{100, 100},
		{-10, 0},
		{150, 100},
	}
	for _, tt := range tests {
		p.SetProgress(tt.input)
		if got := p.Progress(); got != tt.want {
			t.Errorf("SetProgress(%v) → Progress() = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestProgressBar_SetLabel(t *testing.T) {
	p := NewProgressBar()
	p.SetLabel("Loading...")
	if p.Label() != "Loading..." {
		t.Errorf("Label = %q, want 'Loading...'", p.Label())
	}
	p.SetLabel("")
	if p.Label() != "" {
		t.Errorf("Label = %q, want empty", p.Label())
	}
}

func TestProgressBar_SetMode(t *testing.T) {
	p := NewProgressBar()
	if p.Mode() != ProgressDeterminate {
		t.Error("default mode should be Determinate")
	}
	p.SetMode(ProgressIndeterminate)
	if p.Mode() != ProgressIndeterminate {
		t.Error("mode should be Indeterminate after SetMode")
	}
	p.SetMode(ProgressDeterminate)
	if p.Mode() != ProgressDeterminate {
		t.Error("mode should be Determinate after SetMode")
	}
}

func TestProgressBar_Measure(t *testing.T) {
	p := NewProgressBar()
	s := p.Measure(Constraints{MaxWidth: 40})
	if s.W != 40 {
		t.Errorf("W = %d, want 40", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestProgressBar_Measure_DefaultWidth(t *testing.T) {
	p := NewProgressBar()
	s := p.Measure(Constraints{}) // no max width
	if s.W != 40 {
		t.Errorf("W = %d, want default 40", s.W)
	}
}

func TestProgressBar_Measure_WithLabel(t *testing.T) {
	p := NewProgressBar()
	p.SetLabel("Downloading")
	s := p.Measure(Constraints{MaxWidth: 40})
	if s.H != 2 {
		t.Errorf("H = %d, want 2 (label + bar)", s.H)
	}
}

func TestProgressBar_SetShowPercentage(t *testing.T) {
	p := NewProgressBar()
	p.SetShowPercentage(false)
	p.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	p.Paint(buf) // should not panic
}

func TestProgressBar_SetStyle(t *testing.T) {
	p := NewProgressBar()
	style := buffer.Style{Fg: buffer.RGB(255, 0, 0), Bg: buffer.RGB(0, 0, 0)}
	p.SetStyle(style)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
}

func TestProgressBar_Paint_Determinate(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(50)
	p.SetShowPercentage(false)
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	p.Paint(buf)

	// At 50% with width 20, first 10 cells should be filled (█).
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != '█' {
			t.Errorf("cell %d: expected '█', got %q", x, c.Rune)
		}
	}
	// Cells 10-19 should be empty (░).
	for x := 10; x < 20; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != '░' {
			t.Errorf("cell %d: expected '░', got %q", x, c.Rune)
		}
	}
}

func TestProgressBar_Paint_ZeroProgress(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(0)
	p.SetShowPercentage(false)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != '░' {
			t.Errorf("cell %d: expected '░' at 0%%, got %q", x, c.Rune)
		}
	}
}

func TestProgressBar_Paint_FullProgress(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(100)
	p.SetShowPercentage(false)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune != '█' {
			t.Errorf("cell %d: expected '█' at 100%%, got %q", x, c.Rune)
		}
	}
}

func TestProgressBar_Paint_Indeterminate(t *testing.T) {
	p := NewProgressBar()
	p.SetMode(ProgressIndeterminate)
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	p.Paint(buf)

	scanCount := 0
	for x := 0; x < 20; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '█' {
			scanCount++
		}
	}
	if scanCount == 0 {
		t.Error("expected scan segment cells in indeterminate mode")
	}
}

func TestProgressBar_Paint_WithLabel(t *testing.T) {
	p := NewProgressBar()
	p.SetLabel("Downloading")
	p.SetProgress(75)
	p.SetShowPercentage(false)
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(30, 2)
	p.Paint(buf)
	// Label should be on line 0.
	c := buf.GetCell(0, 0)
	if c.Rune != 'D' {
		t.Errorf("Label line: expected 'D', got %q", c.Rune)
	}
}

func TestProgressBar_Paint_ZeroWidth(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(50)
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf) // should not panic
}

func TestProgressBar_Tick_Indeterminate(t *testing.T) {
	p := NewProgressBar()
	p.SetMode(ProgressIndeterminate)
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	// Tick multiple times — should bounce.
	for i := 0; i < 30; i++ {
		p.Tick()
	}
}

func TestProgressBar_Tick_Determinate_NoOp(t *testing.T) {
	p := NewProgressBar()
	p.SetProgress(50)
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	p.Tick() // should be no-op in determinate mode
	if p.Progress() != 50 {
		t.Error("Tick should not affect determinate progress")
	}
}

func TestProgressBar_SetIndeterminateWidth(t *testing.T) {
	p := NewProgressBar()
	p.SetIndeterminateWidth(8)
	p.SetMode(ProgressIndeterminate)
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	p.Paint(buf)

	scanCount := 0
	for x := 0; x < 20; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '█' {
			scanCount++
		}
	}
	if scanCount != 8 {
		t.Errorf("scan cells = %d, want 8", scanCount)
	}
}

func TestProgressBar_ColorProgression(t *testing.T) {
	// At 0%, color should be red.
	c0 := progressColor(0)
	if c0.R() != 255 || c0.G() != 0 {
		t.Errorf("progressColor(0) = R:%d G:%d, want R:255 G:0", c0.R(), c0.G())
	}
	// At 50%, color should be yellow.
	c50 := progressColor(50)
	if c50.R() != 255 || c50.G() != 255 {
		t.Errorf("progressColor(50) = R:%d G:%d, want R:255 G:255", c50.R(), c50.G())
	}
	// At 100%, color should be green.
	c100 := progressColor(100)
	if c100.R() != 0 || c100.G() != 255 {
		t.Errorf("progressColor(100) = R:%d G:%d, want R:0 G:255", c100.R(), c100.G())
	}
}

func TestProgressBar_ConcurrentAccess(t *testing.T) {
	p := NewProgressBar()
	done := make(chan struct{})

	go func() {
		for i := 0; i < 100; i++ {
			p.SetProgress(float64(i % 101))
			p.SetLabel("test")
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		_ = p.Progress()
		_ = p.Mode()
		_ = p.Label()
	}

	<-done
}
