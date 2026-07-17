package overlay

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestModal_SetBounds_SmallContainer_P279(t *testing.T) {
	m := NewModal("id", "Title", nil, []string{"OK"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	m.Paint(buf)
}

func TestModal_SetBounds_LargeContainer_P279(t *testing.T) {
	m := NewModal("id", "Title", nil, []string{"OK"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	m.Paint(buf)
}

func TestModal_Paint_WithBodyComponent_P279(t *testing.T) {
	body := component.NewText("Custom body content")
	m := NewModal("id", "Custom", body, []string{"Close"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 12})
	buf := buffer.NewBuffer(40, 12)
	m.Paint(buf)
}

func TestModal_Paint_TooSmall_P279(t *testing.T) {
	m := NewModal("id", "Title", nil, []string{"OK"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	m.Paint(buf)
}

func TestModal_Paint_WithTitle_P279(t *testing.T) {
	m := NewModal("id", "Long Title Here", nil, []string{"OK", "Cancel"})
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	m.Paint(buf)
}
