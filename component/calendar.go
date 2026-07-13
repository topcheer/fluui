package component

import (
	"fmt"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// CalendarStyle holds visual styling for the Calendar component.
type CalendarStyle struct {
	HeaderFg    buffer.Color
	WeekdayFg   buffer.Color
	DayFg       buffer.Color
	TodayFg     buffer.Color
	TodayBg     buffer.Color
	SelectedFg  buffer.Color
	SelectedBg  buffer.Color
	OtherMonthFg buffer.Color
	BorderFg    buffer.Color
}

// DefaultCalendarStyle returns a Dracula-themed calendar style.
func DefaultCalendarStyle() CalendarStyle {
	return CalendarStyle{
		HeaderFg:     buffer.NamedColor(buffer.NamedCyan),
		WeekdayFg:    buffer.NamedColor(buffer.NamedBlue),
		DayFg:        buffer.NamedColor(buffer.NamedWhite),
		TodayFg:      buffer.NamedColor(buffer.NamedBlack),
		TodayBg:      buffer.NamedColor(buffer.NamedYellow),
		SelectedFg:   buffer.NamedColor(buffer.NamedBlack),
		SelectedBg:   buffer.NamedColor(buffer.NamedGreen),
		OtherMonthFg: buffer.NamedColor(buffer.NamedBrightBlack),
		BorderFg:     buffer.NamedColor(buffer.NamedBrightBlack),
	}
}

// Calendar is a monthly calendar widget with date selection.
type Calendar struct {
	BaseComponent

	current   time.Time
	selected  time.Time
	today     time.Time
	style     CalendarStyle
	weekStart time.Weekday // 0=Sunday, 1=Monday

	OnSelect func(t time.Time)

	mu sync.RWMutex
}

// NewCalendar creates a calendar showing the current month.
func NewCalendar() *Calendar {
	now := time.Now()
	return &Calendar{
		current:   now,
		selected:  now,
		today:     now,
		style:     DefaultCalendarStyle(),
		weekStart: time.Sunday,
	}
}

// NewCalendarWithDate creates a calendar initialized to a specific date.
func NewCalendarWithDate(t time.Time) *Calendar {
	return &Calendar{
		current:   t,
		selected:  t,
		today:     time.Now(),
		style:     DefaultCalendarStyle(),
		weekStart: time.Sunday,
	}
}

// SetSelected sets the selected date.
func (c *Calendar) SetSelected(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.selected = t
	c.current = t
}

// Selected returns the currently selected date.
func (c *Calendar) Selected() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.selected
}

// SetStyle sets the calendar visual style.
func (c *Calendar) SetStyle(s CalendarStyle) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.style = s
}

// SetWeekStart sets the first day of the week (Sunday or Monday).
func (c *Calendar) SetWeekStart(d time.Weekday) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.weekStart = d
}

// WeekStart returns the configured first day of week.
func (c *Calendar) WeekStart() time.Weekday {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.weekStart
}

// CurrentMonth returns the currently displayed year and month.
func (c *Calendar) CurrentMonth() (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.current.Year(), int(c.current.Month())
}

// NextMonth advances to the next month.
func (c *Calendar) NextMonth() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = c.current.AddDate(0, 1, 0)
}

// PrevMonth goes to the previous month.
func (c *Calendar) PrevMonth() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.current = c.current.AddDate(0, -1, 0)
}

// GoToToday sets the current month to today and selects today.
func (c *Calendar) GoToToday() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	c.today = now
	c.current = now
	c.selected = now
}

// SetOnSelect sets the callback for date selection.
func (c *Calendar) SetOnSelect(fn func(time.Time)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.OnSelect = fn
}

// Measure returns the desired size (7 columns x 8 rows minimum).
func (c *Calendar) Measure(cs Constraints) Size {
	return Size{W: 22, H: 9}
}

// Paint renders the calendar into the buffer.
func (c *Calendar) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := c.Bounds()
	x, y := bounds.X, bounds.Y
	w := bounds.W
	if w < 22 {
		w = 22
	}

	// Header: "Month YYYY"
	header := c.current.Format("January 2006")
	hx := x + (w-len(header))/2
	if hx < x {
		hx = x
	}
	for i, r := range header {
		if hx+i < x+w {
			buf.SetCell(hx+i, y, buffer.Cell{Rune: r, Width: 1, Fg: c.style.HeaderFg, Flags: buffer.Bold})
		}
	}

	// Navigation hints
	if x+w >= 22 {
		buf.SetCell(x, y, buffer.Cell{Rune: '<', Width: 1, Fg: c.style.BorderFg})
		buf.SetCell(x+w-1, y, buffer.Cell{Rune: '>', Width: 1, Fg: c.style.BorderFg})
	}

	// Weekday headers
	weekdays := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	if c.weekStart == time.Monday {
		weekdays = []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	}
	colW := w / 7
	for i, wd := range weekdays {
		cx := x + i*colW
		for j, r := range wd {
			if cx+j < x+w {
				buf.SetCell(cx+j, y+1, buffer.Cell{Rune: r, Width: 1, Fg: c.style.WeekdayFg})
			}
		}
	}

	// Calculate first day of month
	firstDay := time.Date(c.current.Year(), c.current.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, -1)

	// Find starting position
	startWeekday := firstDay.Weekday()
	if c.weekStart == time.Monday {
		startWeekday = (startWeekday + 6) % 7
	}

	// Draw days
	day := 1
	row := 2
	col := int(startWeekday)
	daysInMonth := lastDay.Day()

	// Reset for actual drawing
	col = int(startWeekday)
	day = 1

	for day <= daysInMonth {
		cx := x + col*colW
		dayStr := fmt.Sprintf("%2d", day)

		isToday := c.isTodayLocked(c.current.Year(), c.current.Month(), day)
		isSelected := c.isSelectedLocked(c.current.Year(), c.current.Month(), day)

		fg := c.style.DayFg
		bg := buffer.Color{}
		flags := buffer.StyleFlags(0)

		if isToday {
			fg = c.style.TodayFg
			bg = c.style.TodayBg
			flags = buffer.Bold
		}
		if isSelected {
			fg = c.style.SelectedFg
			bg = c.style.SelectedBg
			flags = buffer.Bold
		}

		for j, r := range dayStr {
			if cx+j < x+w {
				buf.SetCell(cx+j, y+row, buffer.Cell{Rune: r, Width: 1, Fg: fg, Bg: bg, Flags: flags})
			}
		}

		col++
		if col >= 7 {
			col = 0
			row++
		}
		day++
	}

	// Footer with selected date
	if y+row+1 < bounds.Y+bounds.H {
		footer := c.selected.Format("2006-01-02")
		fx := x + (w-len(footer))/2
		for i, r := range footer {
			if fx+i < x+w {
				buf.SetCell(fx+i, y+row+1, buffer.Cell{Rune: r, Width: 1, Fg: c.style.BorderFg})
			}
		}
	}
}

func (c *Calendar) isTodayLocked(y int, m time.Month, d int) bool {
	return c.today.Year() == y && c.today.Month() == m && c.today.Day() == d
}

func (c *Calendar) isSelectedLocked(y int, m time.Month, d int) bool {
	return c.selected.Year() == y && c.selected.Month() == m && c.selected.Day() == d
}

// HandleKey processes keyboard navigation.
func (c *Calendar) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}

	c.mu.Lock()
	var cb func(time.Time)
	changed := false

	switch k.Key {
	case term.KeyLeft:
		c.selected = c.selected.AddDate(0, 0, -1)
		if c.selected.Month() != c.current.Month() || c.selected.Year() != c.current.Year() {
			c.current = c.selected
		}
		changed = true
	case term.KeyRight:
		c.selected = c.selected.AddDate(0, 0, 1)
		if c.selected.Month() != c.current.Month() || c.selected.Year() != c.current.Year() {
			c.current = c.selected
		}
		changed = true
	case term.KeyUp:
		c.selected = c.selected.AddDate(0, 0, -7)
		if c.selected.Month() != c.current.Month() || c.selected.Year() != c.current.Year() {
			c.current = c.selected
		}
		changed = true
	case term.KeyDown:
		c.selected = c.selected.AddDate(0, 0, 7)
		if c.selected.Month() != c.current.Month() || c.selected.Year() != c.current.Year() {
			c.current = c.selected
		}
		changed = true
	case term.KeyPageUp:
		c.current = c.current.AddDate(0, -1, 0)
		c.mu.Unlock()
		return true
	case term.KeyPageDown:
		c.current = c.current.AddDate(0, 1, 0)
		c.mu.Unlock()
		return true
	case term.KeyHome:
		c.selected = c.today
		c.current = c.today
		changed = true
	case term.KeyEnter:
		cb = c.OnSelect
		changed = true
	default:
		// 't' key = go to today
		if k.Rune == 't' {
			c.selected = c.today
			c.current = c.today
			changed = true
		}
		// 'n' = next month, 'p' = prev month
		if k.Rune == 'n' {
			c.current = c.current.AddDate(0, 1, 0)
			c.mu.Unlock()
			return true
		}
		if k.Rune == 'p' {
			c.current = c.current.AddDate(0, -1, 0)
			c.mu.Unlock()
			return true
		}
	}

	sel := c.selected
	c.mu.Unlock()

	if changed && cb != nil {
		cb(sel)
	}
	return changed
}

// Children returns nil.
func (c *Calendar) Children() []Component {
	return nil
}
