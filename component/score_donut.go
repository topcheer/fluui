package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ScoreDonut: Donut/Ring Score Display ───
//
// ScoreDonut renders a circular donut chart showing a percentage score.
// Uses block characters for the ring fill. The center shows the numeric value.
// Useful for dashboards showing completion rates, health scores, etc.
//
// Usage:
//
//	sd := NewScoreDonut()
//	sd.SetValue(72) // 72%
//	sd.SetLabel("Health")
//	sd.Paint(buf)

type ScoreDonutStyle struct {
	Filled buffer.Style
	Empty  buffer.Style
	Center buffer.Style
	Label  buffer.Style
}

func DefaultScoreDonutStyle() ScoreDonutStyle {
	return ScoreDonutStyle{
		Filled: buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Empty:  buffer.Style{Fg: buffer.RGB(51, 65, 85)},
		Center: buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// 12-segment ring positions
var donutSegments = [13][12]rune{
	{}, // placeholder for 0%
	{'▏', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	{'▎', '▏', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	{'▍', '▎', '▏', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	{'▌', '▍', '▎', '▏', ' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	{'▋', '▌', '▍', '▎', '▏', ' ', ' ', ' ', ' ', ' ', ' ', ' '},
	{'▊', '▋', '▌', '▍', '▎', '▏', ' ', ' ', ' ', ' ', ' ', ' '},
	{'▉', '▊', '▋', '▌', '▍', '▎', '▏', ' ', ' ', ' ', ' ', ' '},
	{'█', '▉', '▊', '▋', '▌', '▍', '▎', '▏', ' ', ' ', ' ', ' '},
	{'█', '█', '▉', '▊', '▋', '▌', '▍', '▎', '▏', ' ', ' ', ' '},
	{'█', '█', '█', '▉', '▊', '▋', '▌', '▍', '▎', '▏', ' ', ' '},
	{'█', '█', '█', '█', '▉', '▊', '▋', '▌', '▍', '▎', '▏', ' '},
	{'█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█', '█'},
}

// ScoreDonut renders a donut score chart.
type ScoreDonut struct {
	BaseComponent
	mu sync.Mutex

	value int // 0-100
	label string
	style ScoreDonutStyle
	// cached
	pctStr string
	segIdx int
	ring   [12]rune
}

func NewScoreDonut() *ScoreDonut {
	sd := &ScoreDonut{style: DefaultScoreDonutStyle()}
	sd.SetID(GenerateID("donut"))
	sd.recomputeLocked()
	return sd
}

func (sd *ScoreDonut) SetValue(v int) *ScoreDonut {
	sd.mu.Lock()
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	sd.value = v
	sd.recomputeLocked()
	sd.mu.Unlock()
	return sd
}

func (sd *ScoreDonut) SetLabel(l string) *ScoreDonut {
	sd.mu.Lock()
	sd.label = l
	sd.mu.Unlock()
	return sd
}

func (sd *ScoreDonut) recomputeLocked() {
	sd.pctStr = itoa(sd.value) + "%"
	sd.segIdx = sd.value * 12 / 100
	if sd.segIdx > 12 {
		sd.segIdx = 12
	}

	if sd.segIdx == 0 {
		for i := range sd.ring {
			sd.ring[i] = ' '
		}
	} else {
		seg := donutSegments[sd.segIdx]
		copy(sd.ring[:], seg[:])
	}
}

func (sd *ScoreDonut) Value() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.value
}

func (sd *ScoreDonut) SetStyle(s ScoreDonutStyle) *ScoreDonut {
	sd.mu.Lock()
	sd.style = s
	sd.mu.Unlock()
	return sd
}

func (sd *ScoreDonut) Measure(cs Constraints) Size {
	return Size{W: 8, H: 3}
}

func (sd *ScoreDonut) Paint(buf *buffer.Buffer) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	b := sd.Bounds()
	x, y := b.X, b.Y

	filledStyle := sd.style.Filled
	emptyStyle := sd.style.Empty
	centerStyle := sd.style.Center
	labelStyle := sd.style.Label

	// Row 0: ring top (3-6 segments)
	for i := 3; i <= 6 && i < 12; i++ {
		cx := x + i - 3
		if cx >= buf.Width {
			break
		}
		r := sd.ring[i%12]
		var st buffer.Style
		if r == ' ' || r == 0 {
			st = emptyStyle
		} else {
			st = filledStyle
		}
		if r == 0 {
			r = ' '
		}
		buf.SetCell(cx, y, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
	}

	// Row 1: sides + center value
	// Left side: segments 2,1
	for _, idx := range []int{2, 1} {
		if idx >= 0 && idx < 12 {
			cx := x + (2 - idx)
			if cx < buf.Width && cx >= 0 {
				r := sd.ring[idx]
				var st buffer.Style
				if r == ' ' || r == 0 {
					st = emptyStyle
				} else {
					st = filledStyle
				}
				if r == 0 {
					r = ' '
				}
				buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
			}
		}
	}
	// Center value
	col := x + 3
	for _, r := range sd.pctStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: centerStyle.Fg, Bg: centerStyle.Bg, Flags: centerStyle.Flags, Width: 1})
		col++
	}
	// Right side: segments 7,8
	for _, idx := range []int{7, 8} {
		if idx >= 0 && idx < 12 {
			cx := x + 6 + (idx - 7)
			if cx < buf.Width {
				r := sd.ring[idx]
				var st buffer.Style
				if r == ' ' || r == 0 {
					st = emptyStyle
				} else {
					st = filledStyle
				}
				if r == 0 {
					r = ' '
				}
				buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
			}
		}
	}

	// Row 2: ring bottom (9-0) + label
	for i := 9; i <= 11; i++ {
		cx := x + i - 9
		if cx >= buf.Width {
			break
		}
		r := sd.ring[i]
		var st buffer.Style
		if r == ' ' || r == 0 {
			st = emptyStyle
		} else {
			st = filledStyle
		}
		if r == 0 {
			r = ' '
		}
		buf.SetCell(cx, y+2, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
	}
	// Segment 0 at bottom-right
	cx := x + 3
	if cx < buf.Width {
		r := sd.ring[0]
		var st buffer.Style
		if r == ' ' || r == 0 {
			st = emptyStyle
		} else {
			st = filledStyle
		}
		if r == 0 {
			r = ' '
		}
		buf.SetCell(cx, y+2, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
	}

	// Label below or beside
	if sd.label != "" {
		for i, r := range sd.label {
			if x+i >= buf.Width {
				break
			}
			_ = i
			_ = r
			_ = labelStyle
		}
	}
}

func (sd *ScoreDonut) Children() []Component { return nil }
