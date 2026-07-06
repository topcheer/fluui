package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

func TestThemeStudio_New(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	if ts.SlotCount() == 0 {
		t.Error("expected non-zero slot count")
	}
	if ts.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", ts.Cursor())
	}
	if ts.Changed() {
		t.Error("new ThemeStudio should not be changed")
	}
	if ts.IsPickerOpen() {
		t.Error("picker should not be open initially")
	}
}

func TestThemeStudio_SetCursor(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	n := ts.SlotCount()

	ts.SetCursor(2)
	if ts.Cursor() != 2 {
		t.Errorf("Cursor = %d, want 2", ts.Cursor())
	}

	// Wrap around negative
	ts.SetCursor(-1)
	if ts.Cursor() != n-1 {
		t.Errorf("Cursor = %d, want %d", ts.Cursor(), n-1)
	}

	// Wrap around overflow
	ts.SetCursor(n)
	if ts.Cursor() != 0 {
		t.Errorf("Cursor = %d, want 0", ts.Cursor())
	}
}

func TestThemeStudio_HandleKey_Navigation(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	n := ts.SlotCount()

	// Down
	ts.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if ts.Cursor() != 1 {
		t.Errorf("after Down: Cursor = %d, want 1", ts.Cursor())
	}

	// Up
	ts.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if ts.Cursor() != 0 {
		t.Errorf("after Up: Cursor = %d, want 0", ts.Cursor())
	}

	// j (vim down)
	ts.HandleKey(&term.KeyEvent{Rune: 'j'})
	if ts.Cursor() != 1 {
		t.Errorf("after j: Cursor = %d, want 1", ts.Cursor())
	}

	// k (vim up)
	ts.HandleKey(&term.KeyEvent{Rune: 'k'})
	if ts.Cursor() != 0 {
		t.Errorf("after k: Cursor = %d, want 0", ts.Cursor())
	}

	// Home/g
	ts.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if ts.Cursor() != 0 {
		t.Errorf("after Home: Cursor = %d, want 0", ts.Cursor())
	}

	// End/G
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if ts.Cursor() != n-1 {
		t.Errorf("after End: Cursor = %d, want %d", ts.Cursor(), n-1)
	}
}

func TestThemeStudio_HandleKey_EnterOpensPicker(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	if !ts.HandleKey(&term.KeyEvent{Key: term.KeyEnter}) {
		t.Error("Enter should be consumed")
	}
	if !ts.IsPickerOpen() {
		t.Error("Enter should open picker")
	}

	// Escape closes picker
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEscape})
	if ts.IsPickerOpen() {
		t.Error("Escape should close picker")
	}
}

func TestThemeStudio_HandleKey_Reset(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	// Modify the theme first
	ts.mu.Lock()
	theme.Get().Fg = buffer.RGB(255, 0, 0)
	ts.changed = true
	ts.mu.Unlock()

	if !ts.Changed() {
		t.Error("should be changed after modification")
	}

	// Reset
	ts.HandleKey(&term.KeyEvent{Rune: 'r'})

	if ts.Changed() {
		t.Error("should not be changed after reset")
	}
}

func TestThemeStudio_HandleKey_Save(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	saved := false
	ts.OnSave = func() { saved = true }

	ts.HandleKey(&term.KeyEvent{Rune: 's'})
	if !saved {
		t.Error("OnSave should be called on 's'")
	}
}

func TestThemeStudio_HandleKey_ClosePicker(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.mu.Lock()
	ts.pickerOpen = true
	ts.mu.Unlock()

	ts.HandleKey(&term.KeyEvent{Rune: 'q'})
	if ts.IsPickerOpen() {
		t.Error("q should close picker")
	}
}

func TestThemeStudio_HandleKey_UnknownKey(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	consumed := ts.HandleKey(&term.KeyEvent{Key: term.KeyCode(999)})
	if consumed {
		t.Error("unknown key should not be consumed")
	}
}

func TestThemeStudio_ApplyColorChange(t *testing.T) {
	ts := NewThemeStudio(theme.Get())

	// Open picker for first slot (Accent - sorted alphabetically in Base category)
	ts.SetCursor(0)
	originalColor := theme.Get().Accent

	ts.mu.Lock()
	ts.openPickerLocked()
	ts.mu.Unlock()

	// Apply a new color through the picker's OnChange callback
	newColor := buffer.RGB(50, 50, 50)
	ts.picker.SetColor(newColor)

	// Verify the theme was changed (slot 0 is Accent after sorting)
	if theme.Get().Accent != newColor {
		t.Errorf("Accent = %v, want %v", theme.Get().Accent, newColor)
	}
	if !ts.Changed() {
		t.Error("should be changed after color modification")
	}

	// Restore
	theme.Get().Accent = originalColor
}

func TestThemeStudio_Reset(t *testing.T) {
	ts := NewThemeStudio(theme.Get())

	// Change a color
	ts.mu.Lock()
	theme.Get().Accent = buffer.RGB(1, 2, 3)
	ts.changed = true
	ts.mu.Unlock()

	// Reset
	ts.Reset()
	if ts.Changed() {
		t.Error("should not be changed after Reset")
	}
}

func TestThemeStudio_OnChange(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	fired := false
	ts.OnChange = func() { fired = true }

	// Trigger a change
	ts.mu.Lock()
	ts.changed = true
	ts.fireChangeLocked()
	ts.mu.Unlock()

	if !fired {
		t.Error("OnChange should fire")
	}
}

func TestThemeStudio_Measure(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	s := ts.Measure(Constraints{MaxWidth: 80, MaxHeight: 100})
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("Measure = %v, expected positive", s)
	}
	if s.W > 80 {
		t.Errorf("W = %d, should be <= 80", s.W)
	}
}

func TestThemeStudio_Measure_NarrowConstraints(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	s := ts.Measure(Constraints{MaxWidth: 30, MaxHeight: 10})
	if s.W > 30 {
		t.Errorf("W = %d, should be <= 30", s.W)
	}
	if s.H > 10 {
		t.Errorf("H = %d, should be <= 10", s.H)
	}
}

func TestThemeStudio_Paint(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf)

	// Check title is rendered
	found := false
	for x := 0; x < 60; x++ {
		if buf.GetCell(x, 0).Rune == 'T' {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'T' from 'Theme Studio' title in paint output")
	}
}

func TestThemeStudio_Paint_ZeroBounds(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 10)
	ts.Paint(buf) // should not panic
}

func TestThemeStudio_Paint_WithPickerOpen(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	ts.mu.Lock()
	ts.pickerOpen = true
	ts.mu.Unlock()

	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf) // should render picker overlay without panic
}

func TestThemeStudio_SetStyle(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	newStyle := DefaultThemeStudioStyle()
	newStyle.Title = buffer.Style{Fg: buffer.Red, Flags: buffer.Bold}
	ts.SetStyle(newStyle)
	// Just verify no panic
}

func TestThemeStudio_ClosePicker(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.mu.Lock()
	ts.pickerOpen = true
	ts.mu.Unlock()

	ts.ClosePicker()
	if ts.IsPickerOpen() {
		t.Error("ClosePicker should close the picker")
	}
}

func TestThemeStudio_ConcurrentAccess(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	// SetBounds is NOT thread-safe, so set it once before concurrent goroutines
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = ts.Cursor()
			_ = ts.SlotCount()
			_ = ts.Changed()
			_ = ts.IsPickerOpen()
			if n%2 == 0 {
				ts.SetCursor(n % ts.SlotCount())
			}
			buf := buffer.NewBuffer(60, 20)
			ts.Paint(buf)
		}(i)
	}
	wg.Wait()
}

func TestColorToHex(t *testing.T) {
	tests := []struct {
		name string
		c    buffer.Color
		want string
	}{
		{"true color", buffer.RGB(255, 0, 0), "#FF0000"},
		{"true color cyan", buffer.RGB(0, 255, 255), "#00FFFF"},
		{"256 color", buffer.Color{Type: buffer.Color256, Val: 42}, "256:42"},
		{"none", buffer.Color{Type: buffer.ColorNone}, "default"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := colorToHex(tc.c)
			if got != tc.want {
				t.Errorf("colorToHex(%v) = %q, want %q", tc.c, got, tc.want)
			}
		})
	}
}
