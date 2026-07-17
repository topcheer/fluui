package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

func TestThemeStudio_HandleKey_Q_P262(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	// Open picker first
	ts.SetCursor(0)
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	// Press 'q' to close
	ts.HandleKey(&term.KeyEvent{Rune: 'q'})
	if ts.IsPickerOpen() {
		t.Error("picker should be closed after q")
	}
}

func TestThemeStudio_HandleKey_Browse_P262(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	ts.SetCursor(0)
	ts.HandleKey(&term.KeyEvent{Rune: 'b'})
}

func TestThemeStudio_Paint_SmallBounds_P262(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 4}) // very small
	buf := buffer.NewBuffer(10, 4)
	ts.Paint(buf)
}

func TestThemeStudio_Paint_CategoryOverflow_P262(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 6}) // small height, categories overflow
	buf := buffer.NewBuffer(80, 6)
	ts.Paint(buf)
}

func TestThemeStudio_OpenPicker_InvalidCursor_P262(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	ts.SetCursor(-5) // wraps to last
	ts.SetCursor(999) // wraps to 0
	// This should work fine since cursor wraps
}

func TestThemeStudio_PickerSmallBounds_P262(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10}) // small for picker
	ts.SetCursor(0)
	ts.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	buf := buffer.NewBuffer(20, 10)
	ts.Paint(buf)
}
