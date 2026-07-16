package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P236: cover codeblock.paintStreamingCursorLocked + richlog + help

func TestCodeBlock_StreamCursorEmpty_P236(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestCodeBlock_StreamCursorEmptyWithTitle_P236(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetShowTitle(true)
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestCodeBlock_StreamCursorWithLines_P236(t *testing.T) {
	cb := NewCodeBlock("go", "package main\nfunc foo() {}")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestCodeBlock_StreamCursorWithScroll_P236(t *testing.T) {
	cb := NewCodeBlock("go", "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	cb.Paint(buf)
}

func TestCodeBlock_StreamCursorLongLine_P236(t *testing.T) {
	cb := NewCodeBlock("go", "package main // very long comment that exceeds width")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
}

func TestCodeBlock_StreamCursorNarrowHeight_P236(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	cb.Paint(buf)
}

func TestCodeBlock_StreamAppendDebounce_P236(t *testing.T) {
	cb := NewCodeBlock("go", "")
	cb.SetStreaming(true)
	cb.SetStreamDebounce(3)
	cb.AppendSource("package main\n") // count=1, 1%3!=0 → plain
	cb.AppendSource("func foo() {}\n") // count=2, 2%3!=0 → plain
	cb.AppendSource("func bar() {}\n") // count=3, 3%3==0 → full re-highlight
	cb.AppendSource("func baz() {}\n") // count=4, 4%3!=0 → plain fallback TRUE
	// Now usePlainFallback=true → Paint uses plainLines path
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestRichLog_CountVisibleWithMinLevel_P236(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	rl.SetMinLevel(LogWarn)
	rl.Info("info entry")
	rl.Warn("warn entry")
	rl.Error("error entry")
	buf := buffer.NewBuffer(40, 10)
	rl.Paint(buf)
}

func TestRichLog_CountVisibleWithLongEntry_P236(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	rl.Info("this is a very long log entry that should wrap multiple lines")
	buf := buffer.NewBuffer(10, 10)
	rl.Paint(buf)
}

func TestHelpOverlay_ScrollUp_P236(t *testing.T) {
	groups := []HelpGroup{
		{Name: "Nav", Entries: []HelpEntry{
			{Keys: "ctrl+a", Description: "Action A"},
			{Keys: "ctrl+b", Description: "Action B"},
			{Keys: "ctrl+c", Description: "Action C"},
			{Keys: "ctrl+d", Description: "Action D"},
			{Keys: "ctrl+e", Description: "Action E"},
			{Keys: "ctrl+f", Description: "Action F"},
			{Keys: "ctrl+g", Description: "Action G"},
			{Keys: "ctrl+h", Description: "Action H"},
		}},
	}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	h.Paint(buf)
	h.ScrollDown(1)
	h.Paint(buf)
	h.ScrollUp(1)
	h.Paint(buf)
}
