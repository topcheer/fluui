package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MiniCalendar: Compact Month Calendar ───
//
// MiniCalendar renders a compact month view showing day numbers in a
// grid. The current day is highlighted. Useful for date pickers and
// scheduling widgets in dashboard layouts.
//
// Usage:
//
//	mc := NewMiniCalendar()
//	mc.SetMonth(2024, 3) // March 2024
//	mc.SetToday(15)      // 15th is today
//	mc.Paint(buf)

// MiniCalendarStyle holds styling.
type MiniCalendarStyle struct {
	Header  buffer.Style
	Day     buffer.Style
	Today   buffer.Style
	Weekend buffer.Style
	Empty   buffer.Style
}

// DefaultMiniCalendarStyle returns defaults.
func DefaultMiniCalendarStyle() MiniCalendarStyle {
	return MiniCalendarStyle{
		Header:  buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Day:     buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Today:   buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Weekend: buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Empty:   buffer.Style{Fg: buffer.RGB(30, 41, 59)},
	}
}

// Days per month (non-leap year, index 1-12)
var monthDays = [13]int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
var dayLabels = [7]string{"S", "M", "T", "W", "T", "F", "S"}

// MiniCalendar renders a compact month calendar.
type MiniCalendar struct {
	BaseComponent
	mu sync.Mutex

	year         int
	month        int  // 1-12
	today        int  // 0=none
	weekStartMon bool // Monday-first if true
	style        MiniCalendarStyle
	// cached
	daysInMonth int
	startOffset int         // 0=Sunday for day 1
	dayRunes    [31][2]rune // pre-computed 2-rune day representations
}

// NewMiniCalendar creates a MiniCalendar.
func NewMiniCalendar() *MiniCalendar {
	mc := &MiniCalendar{year: 2024, month: 1, style: DefaultMiniCalendarStyle()}
	mc.SetID(GenerateID("minical"))
	mc.recomputeLocked()
	return mc
}

// SetMonth sets the year and month (1-12).
func (mc *MiniCalendar) SetMonth(year, month int) *MiniCalendar {
	mc.mu.Lock()
	if month < 1 {
		month = 1
	}
	if month > 12 {
		month = 12
	}
	mc.year = year
	mc.month = month
	mc.recomputeLocked()
	mc.mu.Unlock()
	return mc
}

// SetToday sets the highlighted day (0 = none).
func (mc *MiniCalendar) SetToday(day int) *MiniCalendar {
	mc.mu.Lock()
	mc.today = day
	mc.mu.Unlock()
	return mc
}

// SetWeekStartMonday sets Monday as the first day of week.
func (mc *MiniCalendar) SetWeekStartMonday(v bool) *MiniCalendar {
	mc.mu.Lock()
	mc.weekStartMon = v
	mc.recomputeLocked()
	mc.mu.Unlock()
	return mc
}

func (mc *MiniCalendar) recomputeLocked() {
	days := monthDays[mc.month]
	if mc.month == 2 && mc.year%4 == 0 && (mc.year%100 != 0 || mc.year%400 == 0) {
		days = 29
	}
	mc.daysInMonth = days
	// Simplified: Jan 1 of year Y is day-of-week = (Y + (Y-1)/4 - (Y-1)/100 + (Y-1)/400) % 7
	// We compute the first day of the month using a known reference (Jan 1, 2024 = Monday)
	totalDays := 0
	for y := 2024; y < mc.year; y++ {
		totalDays += 365
		if y%4 == 0 && (y%100 != 0 || y%400 == 0) {
			totalDays++
		}
	}
	for m := 1; m < mc.month; m++ {
		d := monthDays[m]
		if m == 2 && mc.year%4 == 0 && (mc.year%100 != 0 || mc.year%400 == 0) {
			d = 29
		}
		totalDays += d
	}
	// Jan 1, 2024 was Monday = 1 (0=Sun)
	mc.startOffset = (1 + totalDays) % 7
	if mc.startOffset < 0 {
		mc.startOffset += 7
	}

	// Pre-compute 2-rune representations for days 1-31 (zero-alloc in Paint)
	for d := 1; d <= 31; d++ {
		if d < 10 {
			mc.dayRunes[d-1] = [2]rune{'0', rune('0' + d)}
		} else {
			mc.dayRunes[d-1] = [2]rune{rune('0' + d/10), rune('0' + d%10)}
		}
	}
}

// SetStyle sets custom style.
func (mc *MiniCalendar) SetStyle(s MiniCalendarStyle) *MiniCalendar {
	mc.mu.Lock()
	mc.style = s
	mc.mu.Unlock()
	return mc
}

// Measure returns preferred size.
func (mc *MiniCalendar) Measure(cs Constraints) Size {
	w := 21 // 3 chars per day * 7 days
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 7} // header + 6 weeks max
}

// Paint renders the mini calendar.
func (mc *MiniCalendar) Paint(buf *buffer.Buffer) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	b := mc.Bounds()
	x, y := b.X, b.Y

	headerStyle := mc.style.Header
	dayStyle := mc.style.Day
	todayStyle := mc.style.Today
	weekendStyle := mc.style.Weekend
	emptyStyle := mc.style.Empty

	// Row 0: day labels
	col := x
	for i := 0; i < 7; i++ {
		idx := i
		if mc.weekStartMon {
			idx = (i + 1) % 7
		}
		label := dayLabels[idx]
		for _, r := range label {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
			col++
		}
	}

	// Days grid
	startOff := mc.startOffset
	if mc.weekStartMon {
		startOff = (startOff + 6) % 7
	}

	day := 1
	for row := 1; row <= 6 && day <= mc.daysInMonth; row++ {
		col = x
		for dow := 0; dow < 7; dow++ {
			if col >= buf.Width {
				break
			}
			var st buffer.Style

			cellIdx := (row-1)*7 + dow
			if cellIdx < startOff || day > mc.daysInMonth {
				// Empty cell
				if col < buf.Width {
					buf.SetCell(col, y+row, buffer.Cell{Rune: ' ', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
					col++
				}
				if col < buf.Width {
					buf.SetCell(col, y+row, buffer.Cell{Rune: ' ', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
					col++
				}
			} else {
				isWeekend := dow == 0 || dow == 6
				if mc.weekStartMon {
					isWeekend = dow == 5 || dow == 6
				}
				var st buffer.Style
				if day == mc.today {
					st = todayStyle
				} else if isWeekend {
					st = weekendStyle
				} else {
					st = dayStyle
				}
				runes := mc.dayRunes[day-1]
				day++
				if col < buf.Width {
					buf.SetCell(col, y+row, buffer.Cell{Rune: runes[0], Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
					col++
				}
				if col < buf.Width {
					buf.SetCell(col, y+row, buffer.Cell{Rune: runes[1], Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
					col++
				}
			}
			if col < buf.Width {
				buf.SetCell(col, y+row, buffer.Cell{Rune: ' ', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
				col++
			}
		}
	}
}

// Children returns nil.
func (mc *MiniCalendar) Children() []Component { return nil }
