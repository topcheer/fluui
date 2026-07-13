package app

import "strings"

// ─── tea.View / tea.NewView compatibility (P165) ───
//
// bubbletea v2's View() returns tea.View (a struct with string content)
// instead of raw string. This adapter provides a compatible type.
//
// Usage:
//
//	func (m *Model) View() View {
//	    return NewView(renderedString)
//	}

// View is a tea.View-compatible wrapper for rendered string output.
type View struct {
	content string
}

// NewView creates a View from a string (tea.NewView compatible).
func NewView(s string) View {
	return View{content: s}
}

// String returns the rendered content.
func (v View) String() string {
	return v.content
}

// Len returns the length of the rendered content.
func (v View) Len() int {
	return len(v.content)
}

// IsEmpty returns true if the view has no content.
func (v View) IsEmpty() bool {
	return len(v.content) == 0
}

// Lines returns the content split by newlines.
func (v View) Lines() []string {
	return strings.Split(v.content, "\n")
}

// LineCount returns the number of lines in the content.
func (v View) LineCount() int {
	if v.content == "" {
		return 0
	}
	return strings.Count(v.content, "\n") + 1
}