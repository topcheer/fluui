package component

import (
	"math"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── SunburstChart: Hierarchical Radial Chart ───
//
// SunburstChart renders hierarchical data as concentric rings, where each
// ring segment's arc is proportional to its value. Common in dashboards
// for showing category breakdowns (e.g., budget allocation, storage usage).
//
// Usage:
//
//	sc := NewSunburstChart()
//	sc.AddSegment(SunburstSegment{Label: "Engineering", Value: 40, Children: nil})
//	sc.AddSegment(SunburstSegment{Label: "Sales", Value: 30})
//	sc.AddSegment(SunburstSegment{Label: "Ops", Value: 20})
//	sc.SetBounds(Rect{X:0, Y:0, W:30, H:20})
//	sc.Paint(buf)

// SunburstSegment represents a node in the hierarchy.
type SunburstSegment struct {
	Label    string
	Value    float64
	Color    buffer.Color
	Children []SunburstSegment
}

// SunburstChartStyle holds visual styles.
type SunburstChartStyle struct {
	Label  buffer.Style
	Border buffer.Style
}

// DefaultSunburstChartStyle returns sensible defaults.
func DefaultSunburstChartStyle() SunburstChartStyle {
	return SunburstChartStyle{
		Label:  buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Border: buffer.Style{Fg: buffer.RGB(60, 60, 60)},
	}
}

// sunburst palette: distinct colors for auto-assignment
var sunburstPalette = [...]buffer.Color{
	buffer.RGB(100, 149, 237),  // cornflower
	buffer.RGB(16, 163, 127),   // green
	buffer.RGB(220, 80, 80),    // red
	buffer.RGB(255, 175, 64),   // orange
	buffer.RGB(147, 112, 219),  // purple
	buffer.RGB(64, 224, 208),   // turquoise
	buffer.RGB(255, 192, 203),  // pink
	buffer.RGB(255, 215, 0),    // gold
}

// SunburstChart renders a hierarchical radial chart.
type SunburstChart struct {
	BaseComponent
	mu       sync.RWMutex
	segments []SunburstSegment
	style    SunburstChartStyle
}

// NewSunburstChart creates an empty sunburst chart.
func NewSunburstChart() *SunburstChart {
	sc := &SunburstChart{
		style: DefaultSunburstChartStyle(),
	}
	sc.SetID(GenerateID("sunburst"))
	return sc
}

// AddSegment adds a top-level segment.
func (sc *SunburstChart) AddSegment(seg SunburstSegment) *SunburstChart {
	sc.mu.Lock()
	if seg.Color.Type == 0 {
		seg.Color = sunburstPalette[len(sc.segments)%len(sunburstPalette)]
	}
	sc.segments = append(sc.segments, seg)
	sc.mu.Unlock()
	return sc
}

// SetSegments replaces all segments.
func (sc *SunburstChart) SetSegments(segs []SunburstSegment) *SunburstChart {
	sc.mu.Lock()
	sc.segments = segs
	sc.mu.Unlock()
	return sc
}

// Segments returns the current segments.
func (sc *SunburstChart) Segments() []SunburstSegment {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.segments
}

// SegmentCount returns the number of top-level segments.
func (sc *SunburstChart) SegmentCount() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return len(sc.segments)
}

// TotalValue returns the sum of all top-level segment values.
func (sc *SunburstChart) TotalValue() float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	total := 0.0
	for _, s := range sc.segments {
		total += s.Value
	}
	return total
}

// Clear removes all segments.
func (sc *SunburstChart) Clear() *SunburstChart {
	sc.mu.Lock()
	sc.segments = sc.segments[:0]
	sc.mu.Unlock()
	return sc
}

// SetStyle sets the visual style.
func (sc *SunburstChart) SetStyle(s SunburstChartStyle) *SunburstChart {
	sc.mu.Lock()
	sc.style = s
	sc.mu.Unlock()
	return sc
}

// Style returns the current style.
func (sc *SunburstChart) Style() SunburstChartStyle {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.style
}

// Measure computes the desired size.
func (sc *SunburstChart) Measure(cs Constraints) Size {
	w, h := 30, 20
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the sunburst chart.
func (sc *SunburstChart) Paint(buf *buffer.Buffer) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	b := sc.bounds
	if b.W < 6 || b.H < 6 || len(sc.segments) == 0 {
		return
	}

	total := 0.0
	for _, s := range sc.segments {
		total += s.Value
	}
	if total <= 0 {
		return
	}

	cx := b.X + b.W/2
	cy := b.Y + b.H/2
	maxR := b.W
	if b.H < maxR {
		maxR = b.H
	}
	maxR /= 2
	if maxR < 2 {
		maxR = 2
	}

	// Draw ring segments using polar-to-cartesian pixel mapping
	angle := 0.0
	for _, seg := range sc.segments {
		arc := (seg.Value / total) * 2 * math.Pi
		endAngle := angle + arc

		// Draw filled sector
		for y := -maxR; y <= maxR; y++ {
			for x := -maxR; x <= maxR; x++ {
				dist := math.Sqrt(float64(x*x + y*y))
				if dist > float64(maxR) {
					continue
				}
				// Calculate angle of this pixel
				pxAngle := math.Atan2(float64(y), float64(x))
				if pxAngle < 0 {
					pxAngle += 2 * math.Pi
				}

				// Check if pixel angle falls within segment arc
				a1 := angle
				a2 := endAngle
				if a2 > 2*math.Pi {
					a2 -= 2 * math.Pi
					if pxAngle >= a1 || pxAngle <= a2 {
						buf.SetCell(cx+x, cy+y, buffer.Cell{Rune: '●', Fg: seg.Color, Bg: sc.style.Border.Bg, Width: 1})
					}
				} else {
					if pxAngle >= a1 && pxAngle <= a2 {
						buf.SetCell(cx+x, cy+y, buffer.Cell{Rune: '●', Fg: seg.Color, Bg: sc.style.Border.Bg, Width: 1})
					}
				}
			}
		}

		angle = endAngle
		if angle >= 2*math.Pi {
			angle -= 2 * math.Pi
		}
	}
}

// Children returns nil.
func (sc *SunburstChart) Children() []Component { return nil }
