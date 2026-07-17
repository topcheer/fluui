package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P256: 0% no-op setters + sub-80% functions

func TestTextArea_SetPrompt_P256(t *testing.T) {
	ta := NewTextArea()
	ta.SetPrompt("> ")
}

func TestTextArea_SetPlaceholder_P256(t *testing.T) {
	ta := NewTextArea()
	ta.SetPlaceholder("hint")
	if ta.Placeholder() != "" {
		t.Error("placeholder should be empty")
	}
}

func TestTextArea_FocusBlur_P256(t *testing.T) {
	ta := NewTextArea()
	ta.Focus()
	ta.Blur()
}

func TestTextArea_SetCharLimit_P256(t *testing.T) {
	ta := NewTextArea()
	ta.SetCharLimit(100)
	if ta.CharLimit() != 0 {
		t.Error("char limit should be 0")
	}
}

func TestDiffPreview_SetShowLineNumbers_P256(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(false)
	if !dp.ShowLineNumbers() {
		t.Error("should always be true")
	}
}

func TestDiffPreview_SetShowStats_P256(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowStats(true)
}

func TestBadge_Measure_P256(t *testing.T) {
	b := NewBadge("test", BadgeInfo)
	s := b.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if s.W <= 0 {
		t.Error("badge width should be > 0")
	}
}

func TestLoadingIndicator_Start_P256(t *testing.T) {
	li := NewLoadingIndicator("Loading")
	li.Start()
	li.Stop()
}

func TestCodeBlock_PaintStreamingCursor_P256(t *testing.T) {
	cb := NewCodeBlock("go", "package main")
	cb.SetStreaming(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	cb.Paint(buf)
}

func TestRichLog_CountVisibleLines_P256(t *testing.T) {
	rl := NewRichLog()
	rl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	for i := 0; i < 20; i++ {
		rl.Info("line")
	}
	buf := buffer.NewBuffer(40, 10)
	rl.Paint(buf)
}

func TestHelp_EnsureSelectedVisible_P256(t *testing.T) {
	groups := []HelpGroup{{Name: "test", Entries: []HelpEntry{{Keys: "Ctrl+S", Description: "Save"}}}}
	h := NewHelpOverlay(groups)
	h.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	h.ScrollDown(100)
	buf := buffer.NewBuffer(40, 5)
	h.Paint(buf)
}
