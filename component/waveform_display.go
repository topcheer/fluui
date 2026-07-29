package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── WaveformDisplay: Audio Waveform Visualization ───
//
// WaveformDisplay renders a vertical bar waveform from a series of
// amplitude samples (0-100). Each sample maps to a column of varying
// height. Useful for audio visualization and signal monitoring.
//
// Usage:
//
//	w := NewWaveformDisplay()
//	w.SetSamples([]int{30, 60, 80, 45, 20, 70, 90, 50})
//	w.Paint(buf)

// WaveformStyle holds styling.
type WaveformStyle struct {
	Peak   buffer.Style
	Normal buffer.Style
	Low    buffer.Style
	Center buffer.Style
}

// DefaultWaveformStyle returns defaults.
func DefaultWaveformStyle() WaveformStyle {
	return WaveformStyle{
		Peak:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Normal: buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Low:    buffer.Style{Fg: buffer.RGB(30, 58, 95)},
		Center: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

const waveformMaxSamples = 64

// WaveformDisplay renders an audio waveform.
type WaveformDisplay struct {
	BaseComponent
	mu sync.Mutex

	samples [waveformMaxSamples]int
	count   int
	height  int
	style   WaveformStyle
	// cached
	barHeights [waveformMaxSamples]int
}

// NewWaveformDisplay creates a WaveformDisplay.
func NewWaveformDisplay() *WaveformDisplay {
	w := &WaveformDisplay{height: 5, style: DefaultWaveformStyle()}
	w.SetID(GenerateID("waveform"))
	return w
}

// SetSamples sets amplitude samples (0-100 each).
func (w *WaveformDisplay) SetSamples(samples []int) *WaveformDisplay {
	w.mu.Lock()
	w.count = 0
	for _, s := range samples {
		if w.count >= waveformMaxSamples {
			break
		}
		if s < 0 {
			s = 0
		}
		if s > 100 {
			s = 100
		}
		w.samples[w.count] = s
		w.count++
	}
	w.recomputeLocked()
	w.mu.Unlock()
	return w
}

// SetHeight sets the waveform display height.
func (w *WaveformDisplay) SetHeight(h int) *WaveformDisplay {
	w.mu.Lock()
	if h < 2 {
		h = 2
	}
	w.height = h
	w.recomputeLocked()
	w.mu.Unlock()
	return w
}

func (w *WaveformDisplay) recomputeLocked() {
	maxH := w.height / 2
	if maxH < 1 {
		maxH = 1
	}
	for i := 0; i < w.count; i++ {
		h := w.samples[i] * maxH / 100
		if h < 1 && w.samples[i] > 0 {
			h = 1
		}
		w.barHeights[i] = h
	}
}

// SampleCount returns the number of samples.
func (w *WaveformDisplay) SampleCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.count
}

// SetStyle sets custom style.
func (w *WaveformDisplay) SetStyle(s WaveformStyle) *WaveformDisplay {
	w.mu.Lock()
	w.style = s
	w.mu.Unlock()
	return w
}

// Measure returns preferred size.
func (w *WaveformDisplay) Measure(cs Constraints) Size {
	w.mu.Lock()
	h := w.height
	w.mu.Unlock()
	if cs.MaxWidth > 0 && w.count > cs.MaxWidth {
		return Size{W: cs.MaxWidth, H: h}
	}
	return Size{W: w.count, H: h}
}

// Paint renders the waveform display.
func (w *WaveformDisplay) Paint(buf *buffer.Buffer) {
	w.mu.Lock()
	defer w.mu.Unlock()

	b := w.Bounds()
	x, y := b.X, b.Y

	peakStyle := w.style.Peak
	normalStyle := w.style.Normal
	lowStyle := w.style.Low
	centerStyle := w.style.Center

	midY := y + w.height/2

	for i := 0; i < w.count; i++ {
		cx := x + i
		if cx >= buf.Width {
			break
		}

		barH := w.barHeights[i]
		amp := w.samples[i]

		var st buffer.Style
		if amp >= 80 {
			st = peakStyle
		} else if amp >= 30 {
			st = normalStyle
		} else {
			st = lowStyle
		}

		// Draw bars centered around midY
		for row := 0; row < barH; row++ {
			// Upper half
			yy := midY - row
			if yy >= 0 && yy < buf.Height {
				buf.SetCell(cx, yy, buffer.Cell{Rune: '█', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
			}
			// Lower half (symmetric)
			if row > 0 {
				yy2 := midY + row
				if yy2 < buf.Height {
					buf.SetCell(cx, yy2, buffer.Cell{Rune: '█', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
				}
			}
		}
	}

	// Center line
	for i := 0; i < w.count; i++ {
		cx := x + i
		if cx >= buf.Width {
			break
		}
		if midY >= 0 && midY < buf.Height {
			// Only draw center line where there's no bar
			if w.barHeights[i] == 0 {
				buf.SetCell(cx, midY, buffer.Cell{Rune: '·', Fg: centerStyle.Fg, Bg: centerStyle.Bg, Flags: centerStyle.Flags, Width: 1})
			}
		}
	}
}

// Children returns nil.
func (w *WaveformDisplay) Children() []Component { return nil }
