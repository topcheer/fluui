package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// StatCard renders a compact metric card with label, value, and optional delta.
// Ideal for AI dashboards showing model stats (tokens/sec, latency, accuracy).
//
// Thread-safe.
type StatCard struct {
	BaseComponent
	mu         sync.RWMutex
	label      string
	value      string
	delta      string // e.g. "+12%" or "-5%"
	deltaPos   bool   // true = green, false = red
	customFg   buffer.Color
}

// NewStatCard creates a stat card with label and value.
func NewStatCard(label, value string) *StatCard {
	return &StatCard{
		BaseComponent: BaseComponent{id: GenerateID("statcard")},
		label:         label,
		value:         value,
	}
}

// Label returns the stat label.
func (s *StatCard) Label() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.label
}

// SetLabel updates the label.
func (s *StatCard) SetLabel(l string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.label = l
}

// Value returns the stat value.
func (s *StatCard) Value() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.value
}

// SetValue updates the value.
func (s *StatCard) SetValue(v string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = v
}

// SetDelta sets the delta indicator (e.g. "+12%", "-3%").
// positive=true renders green, false renders red.
func (s *StatCard) SetDelta(delta string, positive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delta = delta
	s.deltaPos = positive
}

// Delta returns the delta text and whether it's positive.
func (s *StatCard) Delta() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.delta, s.deltaPos
}

// ClearDelta removes the delta indicator.
func (s *StatCard) ClearDelta() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.delta = ""
}

// Measure returns preferred size.
func (s *StatCard) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w := buffer.StringWidth(s.label)
	if vw := buffer.StringWidth(s.value); vw > w {
		w = vw
	}
	if s.delta != "" {
		if dw := buffer.StringWidth(s.delta); dw > w {
			w = dw
		}
	}
	w += 4 // padding
	h := 3 // label + value + delta (or blank)
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if w < 4 { w = 4 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

// Paint draws the stat card. Zero allocations.
func (s *StatCard) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	tt := theme.Get()
	muted := buffer.Style{Fg: tt.Muted}
	valueStyle := buffer.Style{Fg: tt.Accent, Flags: buffer.Bold}

	// Row 0: label
	if bounds.Y < bounds.Y+bounds.H {
		buf.DrawText(bounds.X+1, bounds.Y, s.label, muted)
	}

	// Row 1: value
	if bounds.Y+1 < bounds.Y+bounds.H {
		buf.DrawText(bounds.X+1, bounds.Y+1, s.value, valueStyle)
	}

	// Row 2: delta
	if bounds.Y+2 < bounds.Y+bounds.H && s.delta != "" {
		deltaColor := tt.Error
		if s.deltaPos {
			deltaColor = tt.Success
		}
		deltaStyle := buffer.Style{Fg: deltaColor, Flags: buffer.Bold}
		buf.DrawText(bounds.X+1, bounds.Y+2, s.delta, deltaStyle)
	}
}

// StatCardFromInt creates a stat card from an integer value (zero-alloc value formatting).
func StatCardFromInt(label string, value int64) *StatCard {
	var buf [20]byte
	b := strconv.AppendInt(buf[:0], value, 10)
	return NewStatCard(label, string(b))
}

// StatCardFromFloat creates a stat card from a float value.
func StatCardFromFloat(label string, value float64) *StatCard {
	var buf [32]byte
	b := strconv.AppendFloat(buf[:0], value, 'f', 2, 64)
	return NewStatCard(label, string(b))
}
