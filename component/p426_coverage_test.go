package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Coverage tests for 0% functions across multiple components.

func TestFilteredList_Coverage_P426(t *testing.T) {
	fl := NewFilteredList([]string{"apple", "banana", "cherry", "date", "elderberry"})

	// SetSelected (valid + invalid) — test before filtering so all 5 items are visible
	fl.SetSelected(2)
	if fl.Selected() != 2 {
		t.Errorf("Selected = %d, want 2", fl.Selected())
	}
	fl.SetSelected(100) // should clamp to last item
	if fl.Selected() < 0 || fl.Selected() >= 5 {
		t.Errorf("Selected = %d, should be clamped to [0,4]", fl.Selected())
	}
	fl.SetSelected(-5) // should clamp to 0
	if fl.Selected() != 0 {
		t.Errorf("Selected = %d, want 0", fl.Selected())
	}

	// Query / SetQuery
	fl.SetQuery("an")
	if fl.Query() != "an" {
		t.Errorf("Query = %q, want %q", fl.Query(), "an")
	}

	// MaxVisible / SetMaxVisible
	fl.SetMaxVisible(3)
	if fl.MaxVisible() != 3 {
		t.Errorf("MaxVisible = %d, want 3", fl.MaxVisible())
	}
	fl.SetMaxVisible(0) // should clamp to 1
	if fl.MaxVisible() != 1 {
		t.Errorf("MaxVisible = %d, want 1 (clamped)", fl.MaxVisible())
	}
}

func TestTogglePill_Coverage_P426(t *testing.T) {
	tp := NewTogglePill(false)
	tp.SetOnText("YES")
	if tp.OnText() != "YES" {
		t.Errorf("OnText = %q, want %q", tp.OnText(), "YES")
	}
	tp.SetOffText("NOPE")
	if tp.OffText() != "NOPE" {
		t.Errorf("OffText = %q, want %q", tp.OffText(), "NOPE")
	}
}

func TestSearchBar_Coverage_P426(t *testing.T) {
	sb := NewSearchBar("Search...")
	sb.SetPlaceholder("Search files...")
	if sb.Placeholder() != "Search files..." {
		t.Errorf("Placeholder = %q, want %q", sb.Placeholder(), "Search files...")
	}
}

func TestTextArea_Coverage_P426(t *testing.T) {
	ta := NewTextArea()

	// SetPrompt (no-op but should not panic)
	ta.SetPrompt("> ")
	_ = ta.Prompt()

	// SetPlaceholder (no-op but should not panic)
	ta.SetPlaceholder("Type here...")
	_ = ta.Placeholder()

	// Focus / Blur
	ta.Focus()
	ta.Blur()

	// SetCharLimit (no-op but should not panic)
	ta.SetCharLimit(100)
	_ = ta.CharLimit()
}

func TestDiffPreview_Coverage_P426(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added line\n-removed line\n context line")

	// SetShowLineNumbers (no-op stub)
	dp.SetShowLineNumbers(true)
	dp.SetShowLineNumbers(false)

	// SetShowStats (no-op stub)
	dp.SetShowStats(true)
	dp.SetShowStats(false)
}

func TestSliderRange_Children_P426(t *testing.T) {
	sr := NewSliderRange()
	if sr.Children() != nil {
		t.Error("SliderRange.Children should return nil")
	}
}

func TestDetectLinksInto_Coverage_P426(t *testing.T) {
	text := "Visit https://example.com and http://test.org for more"
	var dst []LinkRange
	dst = DetectLinksInto(text, 0, 0, dst)
	if len(dst) != 2 {
		t.Errorf("expected 2 links, got %d", len(dst))
	}

	// No links
	dst = nil
	dst = DetectLinksInto("no links here", 0, 0, dst)
	if len(dst) != 0 {
		t.Errorf("expected 0 links, got %d", len(dst))
	}
}

func TestBaseComponent_Paint_Coverage_P426(t *testing.T) {
	// BaseComponent.Paint is a no-op; just verify it doesn't panic
	bc := BaseComponent{}
	bc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	bc.Paint(buf) // should be a no-op
}
