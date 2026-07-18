package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

func TestThemeStudio_HandleKey_OpenBrowse_P289(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	ts.HandleKey(&term.KeyEvent{Rune: 'o'})
}

func TestThemeStudio_HandleKey_Quit_P289(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	ts.pickerOpen = true
	ts.HandleKey(&term.KeyEvent{Rune: 'q'})
}

func TestThemeStudio_HandleKey_UnknownKey_P289(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	ts.HandleKey(&term.KeyEvent{Rune: 'z'})
}

func TestThemeStudio_Paint_TinyBounds_P289(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	buf := buffer.NewBuffer(10, 2)
	ts.Paint(buf)
}

func TestThemeStudio_Paint_SmallHeight_P289(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	ts.Paint(buf)
}

func TestThemeStudio_PickerOnChange_P289(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	ts.Paint(buf)
	// Open picker then simulate color change
	ts.HandleKey(&term.KeyEvent{Rune: 'o'})
}

func TestViewport_DrawScrollBars_ZeroBarHeight_P289(t *testing.T) {
	// Content wider+ taller than viewport → both scrollbars active
	// but barH = H - hBarHeight() could be 0 when H=1
	child := NewText("xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx\nline2\nline3\nline4\nline5")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	vp.Paint(buf)
}

func TestViewport_DrawScrollBars_BothActive_P289(t *testing.T) {
	// Both V and H scrollbars active simultaneously
	child := NewText("very long line that overflows width\nline2\nline3\nline4\nline5\nline6\nline7")
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	vp.ScrollDown(1)
	vp.ScrollRight(5)
	buf := buffer.NewBuffer(5, 3)
	vp.Paint(buf)
}

func TestViewport_DrawScrollBars_ThumbClamp_P289(t *testing.T) {
	// Large content → thumb height clamps to 1 minimum
	child := NewText(makeLongString(500))
	vp := NewViewport(child)
	vp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	vp.ScrollDown(100)
	buf := buffer.NewBuffer(20, 5)
	vp.Paint(buf)
}

func makeLongString(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "line " + string(rune('A'+i%26)) + "\n"
	}
	return result
}
