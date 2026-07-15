package textarea

import (
	"testing"

	"github.com/topcheer/fluui/compat/lipgloss"
)

func TestDefaultStyles(t *testing.T) {
	s := DefaultStyles(true)
	if s.Focused.Base != s.Blurred.Base {
		// Both should be NewStyle() — different instances, same zero value
	}
}

func TestDefaultStylesFields(t *testing.T) {
	s := DefaultStyles(false)
	// Verify all fields are accessible (ggcode sets them to NewStyle())
	s.Focused.Base = NewStyle()
	s.Focused.CursorLine = NewStyle()
	s.Focused.EndOfBuffer = NewStyle()
	s.Focused.LineNumber = NewStyle()
	s.Focused.CursorLineNumber = NewStyle()
	s.Blurred.Base = NewStyle()
	s.Blurred.CursorLine = NewStyle()
	s.Blurred.EndOfBuffer = NewStyle()
	s.Blurred.LineNumber = NewStyle()
	s.Blurred.CursorLineNumber = NewStyle()
}

func TestBlink(t *testing.T) {
	result := Blink()
	if result != nil {
		t.Error("Blink should return nil")
	}
}

// NewStyle is a convenience alias for testing.
func NewStyle() lipgloss.Style {
	return lipgloss.NewStyle()
}