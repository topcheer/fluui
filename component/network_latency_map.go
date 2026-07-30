package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── NetworkLatencyMap: Regional Latency Grid Display ───
//
// NetworkLatencyMap renders a compact grid of region-to-region latency
// values using color-coded cells. Each cell shows milliseconds with
// green (fast) / amber (medium) / red (slow) coloring.
//
// Usage:
//
//	nlm := NewNetworkLatencyMap()
//	nlm.SetRegions("US", "EU", "AS")
//	nlm.SetLatency(0, 1, 85) // US→EU = 85ms
//	nlm.SetLatency(1, 2, 200) // EU→AS = 200ms
//	nlm.Paint(buf)

// LatencyCellStyle holds styling.
type LatencyCellStyle struct {
	Fast   buffer.Style // < 50ms
	Medium buffer.Style // 50-150ms
	Slow   buffer.Style // > 150ms
	Local  buffer.Style // self
	Label  buffer.Style
}

// DefaultLatencyCellStyle returns defaults.
func DefaultLatencyCellStyle() LatencyCellStyle {
	return LatencyCellStyle{
		Fast:   buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Medium: buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Slow:   buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Local:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

const latencyMapMaxRegions = 8

// NetworkLatencyMap renders a regional latency grid.
type NetworkLatencyMap struct {
	BaseComponent
	mu sync.Mutex

	regions   [latencyMapMaxRegions]string
	latencies [latencyMapMaxRegions * latencyMapMaxRegions]int
	count     int
	style     LatencyCellStyle
	// cached
	cellWidths [latencyMapMaxRegions * latencyMapMaxRegions]string
}

// NewNetworkLatencyMap creates a NetworkLatencyMap.
func NewNetworkLatencyMap() *NetworkLatencyMap {
	nlm := &NetworkLatencyMap{style: DefaultLatencyCellStyle()}
	nlm.SetID(GenerateID("latmap"))
	return nlm
}

// SetRegions sets the region labels (max 8).
func (nlm *NetworkLatencyMap) SetRegions(regions ...string) *NetworkLatencyMap {
	nlm.mu.Lock()
	nlm.count = 0
	for _, r := range regions {
		if nlm.count >= latencyMapMaxRegions {
			break
		}
		nlm.regions[nlm.count] = r
		nlm.count++
	}
	nlm.mu.Unlock()
	return nlm
}

// SetLatency sets latency in ms from region `from` to region `to`.
func (nlm *NetworkLatencyMap) SetLatency(from, to, ms int) *NetworkLatencyMap {
	nlm.mu.Lock()
	if from >= 0 && from < latencyMapMaxRegions && to >= 0 && to < latencyMapMaxRegions {
		idx := from*latencyMapMaxRegions + to
		nlm.latencies[idx] = ms
		nlm.cellWidths[idx] = itoa(ms)
	}
	nlm.mu.Unlock()
	return nlm
}

// RegionCount returns the number of regions.
func (nlm *NetworkLatencyMap) RegionCount() int {
	nlm.mu.Lock()
	defer nlm.mu.Unlock()
	return nlm.count
}

// SetStyle sets custom style.
func (nlm *NetworkLatencyMap) SetStyle(s LatencyCellStyle) *NetworkLatencyMap {
	nlm.mu.Lock()
	nlm.style = s
	nlm.mu.Unlock()
	return nlm
}

// Measure returns preferred size.
func (nlm *NetworkLatencyMap) Measure(cs Constraints) Size {
	nlm.mu.Lock()
	w := nlm.count*7 + 4
	h := nlm.count + 2
	nlm.mu.Unlock()
	if w < 4 {
		w = 4
	}
	if h < 2 {
		h = 2
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the latency map.
func (nlm *NetworkLatencyMap) Paint(buf *buffer.Buffer) {
	nlm.mu.Lock()
	defer nlm.mu.Unlock()

	b := nlm.Bounds()
	x, y := b.X, b.Y

	labelStyle := nlm.style.Label
	fastStyle := nlm.style.Fast
	mediumStyle := nlm.style.Medium
	slowStyle := nlm.style.Slow
	localStyle := nlm.style.Local

	// Header row: region names
	col := x + 4
	for i := 0; i < nlm.count; i++ {
		if col >= buf.Width {
			break
		}
		for _, r := range nlm.regions[i] {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		col += 7 - len([]rune(nlm.regions[i]))
		if col > buf.Width {
			col = buf.Width
		}
	}

	// Data rows
	for from := 0; from < nlm.count; from++ {
		yy := y + 1 + from
		if yy >= buf.Height {
			break
		}
		col = x

		// Row label
		for _, r := range nlm.regions[from] {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}
		for col < x+4 && col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
			col++
		}

		// Cells
		for to := 0; to < nlm.count; to++ {
			if col >= buf.Width {
				break
			}
			idx := from*latencyMapMaxRegions + to
			ms := nlm.latencies[idx]

			var st buffer.Style
			var cellRune rune
			if from == to {
				st = localStyle
				cellRune = '-'
			} else {
				cellRune = ' '
				if ms < 50 {
					st = fastStyle
				} else if ms <= 150 {
					st = mediumStyle
				} else {
					st = slowStyle
				}
			}

			// Latency value
			if from != to {
				for _, r := range nlm.cellWidths[idx] {
					if col >= buf.Width {
						break
					}
					buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
					col++
				}
			} else {
				if col < buf.Width {
					buf.SetCell(col, yy, buffer.Cell{Rune: cellRune, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
					col++
				}
			}
			if col < buf.Width {
				buf.SetCell(col, yy, buffer.Cell{Rune: 'm', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
				col++
			}
			if col < buf.Width {
				buf.SetCell(col, yy, buffer.Cell{Rune: 's', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
				col++
			}
			for col < x+4+(to+1)*7 && col < buf.Width {
				buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (nlm *NetworkLatencyMap) Children() []Component { return nil }
