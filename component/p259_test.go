package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

func TestThemeStudio_SetCursor_Wrap_P259(t *testing.T) {
	ts := NewThemeStudio(theme.Get())
	ts.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 20})
	ts.SetCursor(-1)
	ts.SetCursor(999)
	ts.SetCursor(1)
	buf := buffer.NewBuffer(80, 20)
	ts.Paint(buf)
}

func TestHelpOverlay_EnsureSelectedVisible_ScrollUp_P259(t *testing.T) {
	groups := make([]HelpGroup, 0)
	for i := 0; i < 30; i++ {
		groups = append(groups, HelpGroup{
			Name: "group",
			Entries: []HelpEntry{
				{Keys: "Ctrl+A", Description: "action"},
			},
		})
	}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	h.ScrollDown(10)
	buf := buffer.NewBuffer(40, 8)
	h.Paint(buf)
}

func TestCodeBlock_PaintStreamingCursor_Empty_P259(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	cb.Paint(buf)
}

func TestCodeBlock_PaintStreamingCursor_WithTitle_P259(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetTitle("test.go")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	cb.Paint(buf)
}

func TestLoadingIndicator_StartDouble_P259(t *testing.T) {
	li := NewLoadingIndicator("Loading")
	li.Start()
	li.Start()
	li.Stop()
}

func TestLoadingIndicator_FrameProgression_P259(t *testing.T) {
	li := NewLoadingIndicator("Loading")
	li.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	li.Start()
	time.Sleep(200 * time.Millisecond)
	buf := buffer.NewBuffer(20, 1)
	li.Paint(buf)
	li.Stop()
}
