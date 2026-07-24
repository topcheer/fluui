package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// SegmentedControl is an iOS-style segmented control for switching between
// mutually exclusive options. Common uses include mode switching
// (Chat | Code | Settings), view toggles (List | Grid), and filter
// selection (All | Active | Done).
//
// Features:
//   - Horizontal segments with highlighted active segment
//   - Inverse video for active segment
//   - Optional icons/short labels
//   - Thread-safe
//   - Zero-alloc Paint
type SegmentedControl struct {
	BaseComponent
	mu sync.Mutex

	segments []string
	active   int
}

// NewSegmentedControl creates a segmented control with the given labels.
// The first segment is active by default.
func NewSegmentedControl(segments []string) *SegmentedControl {
	return &SegmentedControl{
		BaseComponent: BaseComponent{id: GenerateID("segmented")},
		segments:      segments,
		active:        0,
	}
}

// ActiveIndex returns the index of the currently active segment.
func (s *SegmentedControl) ActiveIndex() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// SetActive sets the active segment by index. Out-of-range values are ignored.
func (s *SegmentedControl) SetActive(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx >= 0 && idx < len(s.segments) {
		s.active = idx
	}
}

// ActiveLabel returns the label of the active segment, or "" if empty.
func (s *SegmentedControl) ActiveLabel() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active < len(s.segments) {
		return s.segments[s.active]
	}
	return ""
}

// SetSegments replaces all segments.
func (s *SegmentedControl) SetSegments(segs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.segments = segs
	if s.active >= len(segs) {
		s.active = len(segs) - 1
	}
	if s.active < 0 {
		s.active = 0
	}
}

// SegmentCount returns the number of segments.
func (s *SegmentedControl) SegmentCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.segments)
}

// SelectNext moves to the next segment (wraps around).
func (s *SegmentedControl) SelectNext() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.segments) > 0 {
		s.active = (s.active + 1) % len(s.segments)
	}
}

// SelectPrev moves to the previous segment (wraps around).
func (s *SegmentedControl) SelectPrev() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.segments) > 0 {
		s.active = (s.active - 1 + len(s.segments)) % len(s.segments)
	}
}

// Measure computes the desired size: 1 row tall, width = sum of segments.
func (s *SegmentedControl) Measure(cs Constraints) Size {
	s.mu.Lock()
	defer s.mu.Unlock()

	w := 0
	for _, seg := range s.segments {
		w += utf8.RuneCountInString(seg) + 4 // " <label> "
	}
	if w < 1 {
		w = 1
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 1
	}
	return Size{W: w, H: 1}
}

// Paint renders the segmented control.
func (s *SegmentedControl) Paint(buf *buffer.Buffer) {
	s.mu.Lock()
	segments := s.segments
	active := s.active
	s.mu.Unlock()

	if len(segments) == 0 {
		return
	}

	b := s.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	normalStyle := buffer.Style{Fg: th.Muted}
	activeStyle := buffer.Style{Fg: th.Bg, Bg: th.Accent}
	borderStyle := buffer.Style{Fg: th.Border}

	x := b.X
	for i, seg := range segments {
		segW := utf8.RuneCountInString(seg) + 4 // padding
		style := normalStyle
		if i == active {
			style = activeStyle
		}

		// Draw left border (except first)
		if i > 0 {
			buf.DrawText(x, b.Y, "│", borderStyle)
			x++
		}

		// Draw " label " with brackets for active
		if i == active {
			buf.DrawText(x, b.Y, " ◆ ", style)
		} else {
			buf.DrawText(x, b.Y, "   ", style)
		}
		x += 3

		// Draw label text
		drawn := buf.DrawText(x, b.Y, seg, style)
		x += drawn
		buf.DrawText(x, b.Y, " ", style)
		x++

		_ = segW
	}
}
