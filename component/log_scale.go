package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── LogScale: Logarithmic Scale Gauge ───
//
// LogScale renders a logarithmic progress gauge where the fill represents
// a value on a log scale. Useful for displaying values that span orders
// of magnitude (e.g., latency in microseconds to seconds, file sizes).
//
// Usage:
//
//	ls := NewLogScale()
//	ls.SetValue(1000, 1, 1000000) // value=1000, min=1, max=1M
//	ls.Paint(buf)

// LogScaleStyle holds styling.
type LogScaleStyle struct {
	Fill   buffer.Style
	Empty  buffer.Style
	Label  buffer.Style
	Value  buffer.Style
	Marker buffer.Style
}

// DefaultLogScaleStyle returns defaults.
func DefaultLogScaleStyle() LogScaleStyle {
	return LogScaleStyle{
		Fill:   buffer.Style{Fg: buffer.RGB(168, 85, 247)},
		Empty:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:  buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Marker: buffer.Style{Fg: buffer.RGB(251, 191, 36), Flags: buffer.Bold},
	}
}

// LogScale renders a logarithmic scale gauge.
type LogScale struct {
	BaseComponent
	mu sync.Mutex

	value int
	minV  int
	maxV  int
	width int
	style LogScaleStyle
	// cached
	fillPct  int
	valueStr string
	scaleStr string
}

// NewLogScale creates a LogScale.
func NewLogScale() *LogScale {
	ls := &LogScale{minV: 1, maxV: 1000000, width: 24, style: DefaultLogScaleStyle()}
	ls.SetID(GenerateID("logscale"))
	ls.recomputeLocked()
	return ls
}

// SetValue sets the current value, min, and max on a linear scale.
// The gauge renders the log-position of value between min and max.
func (ls *LogScale) SetValue(value, minV, maxV int) *LogScale {
	ls.mu.Lock()
	if minV < 1 {
		minV = 1
	}
	if maxV <= minV {
		maxV = minV + 1
	}
	if value < minV {
		value = minV
	}
	if value > maxV {
		value = maxV
	}
	ls.value = value
	ls.minV = minV
	ls.maxV = maxV
	ls.recomputeLocked()
	ls.mu.Unlock()
	return ls
}

func (ls *LogScale) recomputeLocked() {
	// Log-scale percentage: log(value/min) / log(max/min)
	logVal := logApprox(ls.value) - logApprox(ls.minV)
	logRange := logApprox(ls.maxV) - logApprox(ls.minV)
	if logRange == 0 {
		logRange = 1
	}
	ls.fillPct = logVal * 100 / logRange
	if ls.fillPct < 0 {
		ls.fillPct = 0
	}
	if ls.fillPct > 100 {
		ls.fillPct = 100
	}

	ls.valueStr = itoa(ls.value)
	ls.scaleStr = itoa(ls.minV) + ".." + itoa(ls.maxV)
}

// logApprox returns an integer approximation of log10(n) * 1000.
// This avoids importing math and keeps it allocation-free.
func logApprox(n int) int {
	if n <= 0 {
		return 0
	}
	result := 0
	for n >= 10 {
		result += 1000
		n /= 10
	}
	// Linear interpolation within the decade
	result += n * 1000 / 10
	return result
}

// Value returns the current value.
func (ls *LogScale) Value() int {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.value
}

// SetWidth sets the gauge width.
func (ls *LogScale) SetWidth(w int) *LogScale {
	ls.mu.Lock()
	if w < 10 {
		w = 10
	}
	ls.width = w
	ls.mu.Unlock()
	return ls
}

// SetStyle sets custom style.
func (ls *LogScale) SetStyle(s LogScaleStyle) *LogScale {
	ls.mu.Lock()
	ls.style = s
	ls.mu.Unlock()
	return ls
}

// Measure returns preferred size.
func (ls *LogScale) Measure(cs Constraints) Size {
	w := ls.width + 12
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 2}
}

// Paint renders the log scale gauge.
func (ls *LogScale) Paint(buf *buffer.Buffer) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	b := ls.Bounds()
	x, y := b.X, b.Y

	fillStyle := ls.style.Fill
	emptyStyle := ls.style.Empty
	labelStyle := ls.style.Label
	valueStyle := ls.style.Value

	// Row 0: bar
	barW := 20
	fillW := ls.fillPct * barW / 100
	col := x
	for i := 0; i < fillW; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
		col++
	}
	for i := fillW; i < barW; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
		col++
	}

	// Row 1: value + scale
	col = x
	for _, r := range ls.valueStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y+1, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range ls.scaleStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (ls *LogScale) Children() []Component { return nil }
