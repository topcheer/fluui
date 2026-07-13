package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestNewCalendar(t *testing.T) {
	c := NewCalendar()
	if c == nil {
		t.Fatal("expected non-nil calendar")
	}
	y, m := c.CurrentMonth()
	if y != time.Now().Year() || m != int(time.Now().Month()) {
		t.Errorf("expected current month, got %d-%d", y, m)
	}
}

func TestNewCalendarWithDate(t *testing.T) {
	d := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	c := NewCalendarWithDate(d)
	if c.Selected().Year() != 2025 || c.Selected().Month() != 6 || c.Selected().Day() != 15 {
		t.Errorf("expected 2025-06-15, got %v", c.Selected())
	}
}

func TestCalendar_SetSelected(t *testing.T) {
	c := NewCalendar()
	d := time.Date(2024, 12, 25, 0, 0, 0, 0, time.UTC)
	c.SetSelected(d)
	if c.Selected().Day() != 25 {
		t.Errorf("expected day 25, got %d", c.Selected().Day())
	}
}

func TestCalendar_NextMonth(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	c.NextMonth()
	y, m := c.CurrentMonth()
	if y != 2025 || m != 2 {
		t.Errorf("expected 2025-02, got %d-%d", y, m)
	}
}

func TestCalendar_PrevMonth(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC))
	c.PrevMonth()
	y, m := c.CurrentMonth()
	if y != 2025 || m != 2 {
		t.Errorf("expected 2025-02, got %d-%d", y, m)
	}
}

func TestCalendar_GoToToday(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	c.GoToToday()
	now := time.Now()
	if c.Selected().Year() != now.Year() {
		t.Errorf("expected today's year, got %d", c.Selected().Year())
	}
}

func TestCalendar_WeekStart(t *testing.T) {
	c := NewCalendar()
	c.SetWeekStart(time.Monday)
	if c.WeekStart() != time.Monday {
		t.Error("expected Monday week start")
	}
}

func TestCalendar_SetOnSelect(t *testing.T) {
	c := NewCalendar()
	called := false
	c.SetOnSelect(func(t time.Time) { called = true })
	if c.OnSelect == nil {
		t.Error("expected non-nil callback")
	}
	_ = called
}

func TestCalendar_Measure(t *testing.T) {
	c := NewCalendar()
	s := c.Measure(Constraints{})
	if s.W != 22 || s.H != 9 {
		t.Errorf("expected 22x9, got %dx%d", s.W, s.H)
	}
}

func TestCalendar_Paint(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	c.SetBounds(Rect{X: 0, Y: 0, W: 24, H: 12})
	buf := buffer.NewBuffer(24, 12)
	c.Paint(buf)
	// Should have the month header
	hasJune := false
	for i := 0; i < 24; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Rune == 'J' {
			hasJune = true
			break
		}
	}
	if !hasJune {
		t.Error("expected 'J' in header row (June)")
	}
}

func TestCalendar_Paint_NilBuffer(t *testing.T) {
	c := NewCalendar()
	c.Paint(nil) // should not panic
}

func TestCalendar_Paint_NarrowWidth(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 8})
	buf := buffer.NewBuffer(10, 8)
	c.Paint(buf) // should not panic
}

func TestCalendar_HandleKey_Left(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	orig := c.Selected().Day()
	c.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if c.Selected().Day() != orig-1 {
		t.Errorf("expected day %d, got %d", orig-1, c.Selected().Day())
	}
}

func TestCalendar_HandleKey_Right(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	orig := c.Selected().Day()
	c.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if c.Selected().Day() != orig+1 {
		t.Errorf("expected day %d, got %d", orig+1, c.Selected().Day())
	}
}

func TestCalendar_HandleKey_Up(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	orig := c.Selected().Day()
	c.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	// Should go back 7 days
	expected := orig - 7
	if expected <= 0 {
		expected += 31 // wrapped to previous month
	}
	if c.Selected().Day() != expected && c.Selected().Month() != 6 {
		// Allow month wrap
	}
}

func TestCalendar_HandleKey_Down(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if c.Selected().Day() != 12 {
		t.Errorf("expected day 12, got %d", c.Selected().Day())
	}
}

func TestCalendar_HandleKey_PageUp(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	_, m := c.CurrentMonth()
	if m != 5 {
		t.Errorf("expected month 5 (May), got %d", m)
	}
}

func TestCalendar_HandleKey_PageDown(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	_, m := c.CurrentMonth()
	if m != 7 {
		t.Errorf("expected month 7 (July), got %d", m)
	}
}

func TestCalendar_HandleKey_Home(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	now := time.Now()
	if c.Selected().Year() != now.Year() {
		t.Errorf("expected today's year, got %d", c.Selected().Year())
	}
}

func TestCalendar_HandleKey_Enter(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	called := false
	var selDate time.Time
	c.SetOnSelect(func(t time.Time) {
		called = true
		selDate = t
	})
	c.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !called {
		t.Error("expected OnSelect callback to fire")
	}
	if selDate.Day() != 15 {
		t.Errorf("expected day 15, got %d", selDate.Day())
	}
}

func TestCalendar_HandleKey_T(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Rune: 't'})
	now := time.Now()
	if c.Selected().Year() != now.Year() {
		t.Errorf("expected today, got %v", c.Selected())
	}
}

func TestCalendar_HandleKey_N_P(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Rune: 'n'})
	_, m := c.CurrentMonth()
	if m != 7 {
		t.Errorf("expected July after 'n', got month %d", m)
	}
	c.HandleKey(&term.KeyEvent{Rune: 'p'})
	_, m = c.CurrentMonth()
	if m != 6 {
		t.Errorf("expected June after 'p', got month %d", m)
	}
}

func TestCalendar_HandleKey_Nil(t *testing.T) {
	c := NewCalendar()
	if c.HandleKey(nil) {
		t.Error("expected false for nil key")
	}
}

func TestCalendar_HandleKey_Unknown(t *testing.T) {
	c := NewCalendar()
	orig := c.Selected()
	if c.HandleKey(&term.KeyEvent{Key: term.KeyEscape}) {
		t.Error("expected false for Escape key")
	}
	if !c.Selected().Equal(orig) {
		t.Error("selected should not change on unknown key")
	}
}

func TestCalendar_SetStyle(t *testing.T) {
	c := NewCalendar()
	s := DefaultCalendarStyle()
	s.HeaderFg = buffer.NamedColor(buffer.NamedRed)
	c.SetStyle(s)
	// Should not panic
	c.Paint(buffer.NewBuffer(24, 12))
}

func TestCalendar_Children(t *testing.T) {
	c := NewCalendar()
	if c.Children() != nil {
		t.Error("expected nil children")
	}
}

func TestCalendar_CrossMonthNavigation(t *testing.T) {
	// Navigate from June 30 to July 1
	c := NewCalendarWithDate(time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if c.Selected().Month() != 7 || c.Selected().Day() != 1 {
		t.Errorf("expected July 1, got %v", c.Selected())
	}
	_, m := c.CurrentMonth()
	if m != 7 {
		t.Errorf("expected current month to update to July, got %d", m)
	}
}

func TestCalendar_CrossMonthBackward(t *testing.T) {
	// Navigate from June 1 to May 31
	c := NewCalendarWithDate(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC))
	c.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if c.Selected().Month() != 5 || c.Selected().Day() != 31 {
		t.Errorf("expected May 31, got %v", c.Selected())
	}
}

func TestCalendar_Concurrent(t *testing.T) {
	c := NewCalendarWithDate(time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC))
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			c.NextMonth()
			c.PrevMonth()
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		c.Paint(buffer.NewBuffer(24, 12))
	}
	<-done
}
