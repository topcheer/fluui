package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// Skeleton renders animated placeholder blocks (░░░░░░) used during
// loading states. Similar to Material Design skeleton loaders but
// using Unicode shading characters.
//
// Thread-safe.
type Skeleton struct {
	BaseComponent
	mu      sync.RWMutex
	width   int
	height  int
	animate bool
	frame   int // animation frame counter
}

// NewSkeleton creates a skeleton with the given width and height.
func NewSkeleton(w, h int) *Skeleton {
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return &Skeleton{
		BaseComponent: BaseComponent{id: GenerateID("skeleton")},
		width:         w,
		height:        h,
		animate:       true,
	}
}

func (s *Skeleton) Width() int { s.mu.RLock(); defer s.mu.RUnlock(); return s.width }
func (s *Skeleton) SetWidth(w int) { s.mu.Lock(); defer s.mu.Unlock(); if w < 1 { w = 1 }; s.width = w }

func (s *Skeleton) Height() int { s.mu.RLock(); defer s.mu.RUnlock(); return s.height }
func (s *Skeleton) SetHeight(h int) { s.mu.Lock(); defer s.mu.Unlock(); if h < 1 { h = 1 }; s.height = h }

func (s *Skeleton) Animate() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.animate }
func (s *Skeleton) SetAnimate(b bool) { s.mu.Lock(); defer s.mu.Unlock(); s.animate = b }

func (s *Skeleton) Tick() { s.mu.Lock(); defer s.mu.Unlock(); s.frame++ }

// Measure returns the preferred size.
func (s *Skeleton) Measure(cs Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w := s.width
	h := s.height
	if cs.HasWidth() && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.HasHeight() && h > cs.MaxHeight { h = cs.MaxHeight }
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return Size{W: w, H: h}
}

// Paint draws the skeleton blocks. Zero allocations.
func (s *Skeleton) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.bounds
	if bounds.W <= 0 || bounds.H <= 0 { return }

	tt := theme.Get()
	muted := tt.Muted

	// Animation: cycle through block chars
	var chars []rune
	if s.animate {
		chars = []rune{'\u2591', '\u2592', '\u2593'} // ░▒▓
	} else {
		chars = []rune{'\u2591'} // ░ static
	}

	for row := 0; row < bounds.H; row++ {
		y := bounds.Y + row
		for col := 0; col < bounds.W; col++ {
			x := bounds.X + col
			r := chars[0]
			if s.animate && len(chars) > 1 {
				r = chars[(s.frame+col+row)%len(chars)]
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: muted})
		}
	}
}

// FormatPercent converts a 0.0–1.0 progress value to a percentage string.
func FormatPercent(v float64) string {
	var buf [8]byte
	b := buf[:0]
	b = strconv.AppendInt(b, int64(v*100), 10)
	b = append(b, '%')
	return string(b)
}
