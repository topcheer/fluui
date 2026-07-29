package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ConfidenceIntervalChart: Confidence Interval Band Chart ───
//
// ConfidenceIntervalChart renders a horizontal band showing a confidence
// interval (lower bound, mean, upper bound) with gradient coloring.
// Useful for ML predictions, survey results, and statistical estimates.
//
// Usage:
//
//	cic := NewConfidenceIntervalChart()
//	cic.SetLabel("Accuracy")
//	cic.SetBounds(0.72, 0.85, 0.92) // lower, mean, upper
//	cic.SetRange(0, 1)
//	cic.Paint(buf)

// ConfidenceIntervalStyle holds styling.
type ConfidenceIntervalStyle struct {
	Label  buffer.Style
	Mean   buffer.Style
	Range  buffer.Style // the confidence band
	Border buffer.Style
}

// DefaultConfidenceIntervalStyle returns defaults.
func DefaultConfidenceIntervalStyle() ConfidenceIntervalStyle {
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	mean := buffer.Style{Fg: buffer.RGB(96, 165, 250), Flags: buffer.Bold}
	rng := buffer.Style{Fg: buffer.RGB(129, 140, 248), Flags: buffer.Dim}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return ConfidenceIntervalStyle{Label: label, Mean: mean, Range: rng, Border: border}
}

// ConfidenceIntervalChart renders a confidence interval band.
type ConfidenceIntervalChart struct {
	BaseComponent
	mu sync.Mutex

	label     string
	lower     float64
	mean      float64
	upper     float64
	rangeMin  float64
	rangeMax  float64
	style     ConfidenceIntervalStyle
	// cached
	meanStr string
	lowStr  string
	upStr   string
}

// NewConfidenceIntervalChart creates a ConfidenceIntervalChart.
func NewConfidenceIntervalChart() *ConfidenceIntervalChart {
	cic := &ConfidenceIntervalChart{
		rangeMin: 0, rangeMax: 1,
		style: DefaultConfidenceIntervalStyle(),
	}
	cic.SetID(GenerateID("confinterval"))
	cic.cacheStringsLocked()
	return cic
}

func (cic *ConfidenceIntervalChart) cacheStringsLocked() {
	cic.meanStr = formatFloatMC(cic.mean*100) + "%"
	cic.lowStr = formatFloatMC(cic.lower*100) + "%"
	cic.upStr = formatFloatMC(cic.upper*100) + "%"
}

// SetLabel sets the display label.
func (cic *ConfidenceIntervalChart) SetLabel(l string) *ConfidenceIntervalChart {
	cic.mu.Lock()
	cic.label = l
	cic.mu.Unlock()
	return cic
}

// Label returns the label.
func (cic *ConfidenceIntervalChart) Label() string {
	cic.mu.Lock()
	defer cic.mu.Unlock()
	return cic.label
}

// SetBounds sets lower, mean, upper bounds (caches display strings).
func (cic *ConfidenceIntervalChart) SetBounds(lower, mean, upper float64) *ConfidenceIntervalChart {
	cic.mu.Lock()
	cic.lower = lower
	cic.mean = mean
	cic.upper = upper
	cic.cacheStringsLocked()
	cic.mu.Unlock()
	return cic
}

// SetRange sets the min/max display range.
func (cic *ConfidenceIntervalChart) SetRange(min, max float64) *ConfidenceIntervalChart {
	cic.mu.Lock()
	cic.rangeMin = min
	cic.rangeMax = max
	cic.mu.Unlock()
	return cic
}

// SetStyleS sets custom style.

func (cic *ConfidenceIntervalChart) SetStyleS(s ConfidenceIntervalStyle) *ConfidenceIntervalChart {
	cic.mu.Lock()
	cic.style = s
	cic.mu.Unlock()
	return cic
}

// Measure returns preferred size.
func (cic *ConfidenceIntervalChart) Measure(cs Constraints) Size {
	w := 50
	h := 4
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the confidence interval chart into the buffer.
func (cic *ConfidenceIntervalChart) Paint(buf *buffer.Buffer) {
	cic.mu.Lock()
	defer cic.mu.Unlock()

	b := cic.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 { w = 50 }
	if h < 3 { h = 4 }

	bs := cic.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	labelStyle := cic.style.Label
	meanStyle := cic.style.Mean
	rangeStyle := cic.style.Range

	// Label on row 1
	col := x + 1
	for _, r := range cic.label {
		if col >= x+w-1 || col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Band on row 2
	barStart := x + 1
	barW := w - 2
	rngSpan := cic.rangeMax - cic.rangeMin
	if rngSpan <= 0 { rngSpan = 1 }

	lowPos := barStart + int((cic.lower-cic.rangeMin)/rngSpan*float64(barW))
	meanPos := barStart + int((cic.mean-cic.rangeMin)/rngSpan*float64(barW))
	upPos := barStart + int((cic.upper-cic.rangeMin)/rngSpan*float64(barW))

	if lowPos < barStart { lowPos = barStart }
	if upPos > barStart+barW-1 { upPos = barStart + barW - 1 }
	if meanPos < lowPos { meanPos = lowPos }
	if meanPos > upPos { meanPos = upPos }

	// Draw range background (lighter)
	for c := lowPos; c <= upPos && c < buf.Width; c++ {
		buf.SetCell(c, y+2, buffer.Cell{Rune: '▒', Fg: rangeStyle.Fg, Bg: rangeStyle.Bg, Flags: rangeStyle.Flags, Width: 1})
	}
	// Draw mean marker
	if meanPos >= barStart && meanPos < buf.Width {
		buf.SetCell(meanPos, y+2, buffer.Cell{Rune: '█', Fg: meanStyle.Fg, Bg: meanStyle.Bg, Flags: meanStyle.Flags, Width: 1})
	}
	// Lower/upper markers
	if lowPos >= barStart && lowPos < buf.Width {
		buf.SetCell(lowPos, y+2, buffer.Cell{Rune: '│', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	}
	if upPos < buf.Width {
		buf.SetCell(upPos, y+2, buffer.Cell{Rune: '│', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	}

	// Values on row 3 (if height allows)
	if h > 3 && y+3 < buf.Height {
		col = x + 1
		for _, r := range cic.lowStr {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, y+3, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		// Mean (centered)
		mLen := len(cic.meanStr)
		mStart := x + (w-mLen)/2
		for i, r := range cic.meanStr {
			cx := mStart + i
			if cx >= x+w-1 || cx >= buf.Width { break }
			buf.SetCell(cx, y+3, buffer.Cell{Rune: r, Fg: meanStyle.Fg, Bg: meanStyle.Bg, Flags: meanStyle.Flags, Width: 1})
		}
		// Upper (right)
		upLen := len(cic.upStr)
		upStart := x + w - 1 - upLen
		for i, r := range cic.upStr {
			cx := upStart + i
			if cx >= x+w-1 || cx >= buf.Width { break }
			buf.SetCell(cx, y+3, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (cic *ConfidenceIntervalChart) Children() []Component { return nil }

// SetBounds_ sets the component bounds (disambiguates from SetBounds CI method).
func (cic *ConfidenceIntervalChart) SetBounds_(r Rect) {
	cic.BaseComponent.SetBounds(r)
}
