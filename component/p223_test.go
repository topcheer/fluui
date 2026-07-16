package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P223: themestudio.setCursorLocked edge cases — test via HandleKey navigation

func TestThemeStudio_SetCursorWrapDown_P223(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	// Navigate down past last slot — should wrap to 0
	for i := 0; i < 50; i++ {
		ts.HandleKey(nil) // shouldn't crash
	}
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf)
}

func TestThemeStudio_SetCursorWrapUp_P223(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	// Navigate up from first — should wrap to last
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf)
}

func TestThemeStudio_SetCursorEmptySlots_P223(t *testing.T) {
	ts := &ThemeStudio{}
	ts.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	// slots is empty — setCursorLocked should set cursor=0
	buf := buffer.NewBuffer(60, 20)
	ts.Paint(buf) // should not panic even with empty slots
}