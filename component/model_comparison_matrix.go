package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ModelComparisonMatrix: Multi-Model Comparison Grid ───
//
// ModelComparisonMatrix renders a comparison table of AI models with metrics
// (speed, quality, cost, context) as colored cells with value badges.
//
// Usage:
//
//	mcm := NewModelComparisonMatrix()
//	mcm.SetMetrics([]string{"Speed", "Quality", "Cost", "Context"})
//	mcm.AddModel("GPT-4o", []float64{85, 92, 40, 128})
//	mcm.AddModel("Claude-3.5", []float64{78, 95, 35, 200})
//	mcm.Paint(buf)

// ModelEntry represents a single model row.
type ModelEntry struct {
	Name    string
	Values  []float64
	// cached
	Strs    []string
}

// ModelComparisonStyle holds styling.
type ModelComparisonStyle struct {
	Header   buffer.Style
	ModelName buffer.Style
	High     buffer.Style  // >=70
	Medium   buffer.Style  // 40-69
	Low      buffer.Style  // <40
	Border   buffer.Style
}

// DefaultModelComparisonStyle returns defaults.
func DefaultModelComparisonStyle() ModelComparisonStyle {
	hdr := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold}
	name := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	high := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	med := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	low := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return ModelComparisonStyle{Header: hdr, ModelName: name, High: high, Medium: med, Low: low, Border: border}
}

// ModelComparisonMatrix renders a multi-model comparison grid.
type ModelComparisonMatrix struct {
	BaseComponent
	mu sync.Mutex

	metrics []string
	models  []ModelEntry
	style   ModelComparisonStyle
}

// NewModelComparisonMatrix creates a ModelComparisonMatrix.
func NewModelComparisonMatrix() *ModelComparisonMatrix {
	mcm := &ModelComparisonMatrix{style: DefaultModelComparisonStyle()}
	mcm.SetID(GenerateID("modelcmp"))
	return mcm
}

// SetMetrics sets the column metric headers.
func (mcm *ModelComparisonMatrix) SetMetrics(m []string) *ModelComparisonMatrix {
	mcm.mu.Lock()
	mcm.metrics = m
	mcm.mu.Unlock()
	return mcm
}

// AddModel adds a model with metric values (caches display strings).
func (mcm *ModelComparisonMatrix) AddModel(name string, values []float64) *ModelComparisonMatrix {
	mcm.mu.Lock()
	entry := ModelEntry{Name: name, Values: values}
	entry.Strs = make([]string, len(values))
	for i, v := range values {
		entry.Strs[i] = itoa(int(v))
	}
	mcm.models = append(mcm.models, entry)
	mcm.mu.Unlock()
	return mcm
}

// ModelCount returns the number of models.
func (mcm *ModelComparisonMatrix) ModelCount() int {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()
	return len(mcm.models)
}

// Clear removes all models.
func (mcm *ModelComparisonMatrix) Clear() *ModelComparisonMatrix {
	mcm.mu.Lock()
	mcm.models = mcm.models[:0]
	mcm.mu.Unlock()
	return mcm
}

// SetStyle sets custom style.
func (mcm *ModelComparisonMatrix) SetStyle(s ModelComparisonStyle) *ModelComparisonMatrix {
	mcm.mu.Lock()
	mcm.style = s
	mcm.mu.Unlock()
	return mcm
}

// valueStyleLocked returns style based on value.
func (mcm *ModelComparisonMatrix) valueStyleLocked(v float64) buffer.Style {
	if v >= 70 { return mcm.style.High }
	if v >= 40 { return mcm.style.Medium }
	return mcm.style.Low
}

// Measure returns preferred size.
func (mcm *ModelComparisonMatrix) Measure(cs Constraints) Size {
	mcm.mu.Lock()
	mCount := len(mcm.metrics)
	rowCount := len(mcm.models)
	mcm.mu.Unlock()
	colW := 10
	w := 16 + mCount*colW
	if w < 30 { w = 30 }
	h := rowCount + 3
	if h < 4 { h = 4 }
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the comparison matrix into the buffer.
func (mcm *ModelComparisonMatrix) Paint(buf *buffer.Buffer) {
	mcm.mu.Lock()
	defer mcm.mu.Unlock()

	b := mcm.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 30 { w = 40 }
	if h < 4 { h = 6 }

	bs := mcm.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	hdrStyle := mcm.style.Header
	nameStyle := mcm.style.ModelName
	colW := 10
	nameW := 14
	headerY := y + 1

	// Column headers
	col := x + 1 + nameW
	for _, metric := range mcm.metrics {
		mLen := len(metric)
		offset := (colW - mLen) / 2
		if offset < 0 { offset = 0 }
		for i := 0; i < colW && col < x+w-1 && col < buf.Width; i++ {
			var ch rune = ' '
			if i >= offset && i < offset+mLen { ch = rune(metric[i-offset]) }
			buf.SetCell(col, headerY, buffer.Cell{Rune: ch, Fg: hdrStyle.Fg, Bg: hdrStyle.Bg, Flags: hdrStyle.Flags, Width: 1})
			col++
		}
	}

	// Model rows
	for rowIdx, model := range mcm.models {
		rowY := y + 2 + rowIdx
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		// Model name
		col := x + 1
		for _, r := range model.Name {
			if col >= x+nameW || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}
		for col < x+1+nameW && col < buf.Width {
			buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}

		// Metric values
		for vIdx, valStr := range model.Strs {
			style := mcm.valueStyleLocked(model.Values[vIdx])
			valLen := len(valStr)
			valStart := col + (colW-valLen)/2
			for c := col; c < col+colW && c < x+w-1 && c < buf.Width; c++ {
				buf.SetCell(c, rowY, buffer.Cell{Rune: ' ', Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			for i, r := range valStr {
				cx := valStart + i
				if cx >= x+w-1 || cx >= buf.Width { break }
				buf.SetCell(cx, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			col += colW
		}
	}
}

// Children returns nil.
func (mcm *ModelComparisonMatrix) Children() []Component { return nil }
