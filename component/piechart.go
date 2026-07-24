package component

import (
	"math"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// PieSlice represents a single slice of the pie.
type PieSlice struct {
	Label string
	Value float64
}

// PieChart renders a pie or donut chart using ASCII block characters.
// Useful for visualizing AI token usage breakdown, model distribution,
// or any proportional data.
//
// The chart renders as a circular grid of half-block characters,
// with each slice using a different color from the theme palette.
// Labels with percentages are rendered alongside.
//
// Thread-safe.
type PieChart struct {
	BaseComponent
	mu sync.Mutex

	slices  []PieSlice
	donut   bool  // if true, renders a donut (hole in center)
	radius  int   // override radius (0 = auto from bounds)
}

// NewPieChart creates a pie chart with the given slices.
func NewPieChart(slices []PieSlice) *PieChart {
	return &PieChart{
		BaseComponent: BaseComponent{id: GenerateID("pie")},
		slices:        slices,
	}
}

// SetSlices replaces all slices.
func (p *PieChart) SetSlices(slices []PieSlice) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.slices = slices
}

// Slices returns a copy of the current slices.
func (p *PieChart) Slices() []PieSlice {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]PieSlice, len(p.slices))
	copy(out, p.slices)
	return out
}

// SetDonut toggles donut mode (hole in center).
func (p *PieChart) SetDonut(v bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.donut = v
}

// IsDonut returns whether donut mode is active.
func (p *PieChart) IsDonut() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.donut
}

// SetRadius overrides the auto-calculated radius.
func (p *PieChart) SetRadius(r int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.radius = r
}

// TotalValue returns the sum of all slice values.
func (p *PieChart) TotalValue() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	total := 0.0
	for _, s := range p.slices {
		total += s.Value
	}
	return total
}

// SliceCount returns the number of slices.
func (p *PieChart) SliceCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.slices)
}

// Measure returns the desired size.
func (p *PieChart) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 40
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 15
	}
	return Size{W: maxW, H: maxH}
}

// sliceColor returns a color for the given slice index using theme palette.
func sliceColor(idx int, th *theme.Theme) buffer.Color {
	colors := []buffer.Color{
		th.Accent,
		th.Success,
		th.Warning,
		th.Error,
		th.BorderMuted,
		th.Muted,
	}
	if idx < len(colors) {
		return colors[idx]
	}
	return colors[idx%len(colors)]
}

// Paint renders the pie chart using half-block characters.
func (p *PieChart) Paint(buf *buffer.Buffer) {
	p.mu.Lock()
	slices := p.slices
	donut := p.donut
	radiusOverride := p.radius
	p.mu.Unlock()

	if len(slices) == 0 {
		return
	}

	b := p.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()

	// Calculate total
	total := 0.0
	for _, s := range slices {
		total += s.Value
	}
	if total <= 0 {
		return
	}

	// Calculate cumulative angles (in radians, 0 = top)
	cumAngles := make([]float64, len(slices))
	cum := 0.0
	for i, s := range slices {
		cum += s.Value / total
		cumAngles[i] = cum * 2 * math.Pi
	}

	// Calculate radius from bounds
	cx := b.X + b.W/2
	cy := b.Y + b.H/2
	radius := radiusOverride
	if radius <= 0 {
		radius = b.H / 2
		if b.W/2 < radius {
			radius = b.W / 2
		}
	}
	if radius < 2 {
		radius = 2
	}
	innerRadius := 0
	if donut {
		innerRadius = radius * 3 / 5
		if innerRadius < 1 {
			innerRadius = 1
		}
	}

	// Render using half-block characters (each char = 0.5 row)
	for py := 0; py < b.H*2; py++ {
		y := b.Y + py/2
		if y >= b.Y+b.H {
			break
		}
		topHalf := py%2 == 0

		for px := 0; px < b.W; px++ {
			x := b.X + px
			if x >= b.X+b.W {
				break
			}

			// Distance from center
			dx := float64(px) - float64(cx-b.X)
			dy := float64(py) - float64((cy-b.Y)*2) - 0.5
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist > float64(radius) {
				continue // outside pie
			}
			if donut && dist < float64(innerRadius) {
				continue // inside donut hole
			}

			// Angle from top (0), clockwise
			angle := math.Atan2(dy, dx) + math.Pi/2
			if angle < 0 {
				angle += 2 * math.Pi
			}

			// Find which slice this angle belongs to
			sliceIdx := 0
			for i, ca := range cumAngles {
				if angle <= ca {
					sliceIdx = i
					break
				}
				if i == len(cumAngles)-1 {
					sliceIdx = i
				}
			}

			color := sliceColor(sliceIdx, th)

			// Use half-block: top half = ▀, bottom half = ▄
			// If cell already has content from the other half, merge
			existing := buf.GetCell(x, y)
			if topHalf {
				if existing.Rune == '\u2584' { // ▄ bottom half already
					buf.SetCell(x, y, buffer.Cell{
						Rune:  ' ', // full block (both halves filled)
						Width: 1,
						Fg:    existing.Fg,
						Bg:    color,
					})
				} else if existing.Rune == 0 || existing.Rune == ' ' {
					buf.SetCell(x, y, buffer.Cell{
						Rune:  '\u2580', // ▀ top half
						Width: 1,
						Fg:    color,
						Bg:    existing.Bg,
					})
				}
			} else {
				if existing.Rune == '\u2580' { // ▀ top half already
					buf.SetCell(x, y, buffer.Cell{
						Rune:  ' ',
						Width: 1,
						Fg:    existing.Fg,
						Bg:    color,
					})
				} else if existing.Rune == 0 || existing.Rune == ' ' {
					buf.SetCell(x, y, buffer.Cell{
						Rune:  '\u2584', // ▄ bottom half
						Width: 1,
						Fg:    color,
						Bg:    existing.Bg,
					})
				}
			}
		}
	}

	// Draw labels in the remaining space (right side)
	labelX := b.X + radius*2 + 2
	if labelX < b.X+b.W {
		labelY := b.Y
		for i, s := range slices {
			if labelY >= b.Y+b.H {
				break
			}
			pct := s.Value / total * 100

			color := sliceColor(i, th)
			// Draw color marker
			buf.SetCell(labelX, labelY, buffer.Cell{
				Rune:  '\u2588', // █ full block
				Width: 1,
				Fg:    color,
				Bg:    color,
			})

			// Draw label text
			label := s.Label
			maxLabelW := b.X + b.W - labelX - 10
			if maxLabelW < 4 {
				maxLabelW = 4
			}
			if utf8.RuneCountInString(label) > maxLabelW {
				label = truncateRunes(label, maxLabelW-1) + "\u2026"
			}
			buf.DrawText(labelX+2, labelY, label, buffer.Style{Fg: th.Fg})

			// Draw percentage
			var pbuf [8]byte
			pb := pbuf[:0]
			pb = appendFloat(pb, pct, 0)
			pb = append(pb, '%')
			pctW := utf8.RuneCountInString(string(pb))
			buf.DrawText(b.X+b.W-pctW-1, labelY, string(pb), buffer.Style{Fg: th.Muted})

			labelY++
		}
	}

	// Donut center text
	if donut && b.W > 4 && b.H > 3 {
		totalStr := pieFormatFloat(total, 1)
		totalW := utf8.RuneCountInString(totalStr)
		tx := cx - totalW/2
		ty := cy
		if tx >= b.X && tx+totalW < b.X+b.W && ty < b.Y+b.H {
			buf.DrawText(tx, ty, totalStr, buffer.Style{Fg: th.Accent})
		}
	}
}

// pieFormatFloat formats a float to a string with the given precision.
func pieFormatFloat(v float64, prec int) string {
	var buf [24]byte
	return string(appendFloat(buf[:0], v, prec))
}

// appendFloat appends a float64 to a byte slice.
func appendFloat(b []byte, v float64, prec int) []byte {
	// Handle integers without decimal point
	if v == math.Trunc(v) && prec == 0 {
		return appendFloatSimple(b, int64(v))
	}
	// Use math for decimal formatting
	mult := math.Pow10(prec)
	scaled := int64(v*mult + 0.5)
	whole := scaled / int64(mult)
	frac := scaled % int64(mult)
	if frac < 0 {
		frac = -frac
	}
	if v < 0 && whole == 0 {
		b = append(b, '-')
	}
	b = appendFloatSimple(b, whole)
	if prec > 0 {
		b = append(b, '.')
		for i := prec - 1; i > 0; i-- {
			pow := int64(math.Pow10(i))
			b = append(b, byte('0'+frac/pow%10))
		}
		b = append(b, byte('0'+frac%10))
	}
	return b
}

// appendFloatSimple appends an int64 to a byte slice.
func appendFloatSimple(b []byte, v int64) []byte {
	if v == 0 {
		return append(b, '0')
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var digits [20]byte
	n := 0
	for v > 0 {
		digits[n] = byte('0' + v%10)
		v /= 10
		n++
	}
	if neg {
		b = append(b, '-')
	}
	for i := n - 1; i >= 0; i-- {
		b = append(b, digits[i])
	}
	return b
}
