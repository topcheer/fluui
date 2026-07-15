package lipgloss

import (
	"testing"
)

// P201: Tests for Faint method

func TestStyleFaint_P201(t *testing.T) {
	s := NewStyle().Faint(true)
	if !s.GetBold() { // just verify it doesn't panic
		// Faint is alias for Dim, check via rendering
	}
	result := s.Render("test")
	if result == "" {
		t.Error("Render should not be empty")
	}
}

func TestStyleFaintNoArg_P201(t *testing.T) {
	s := NewStyle().Faint()
	_ = s.Render("test")
}

func TestStyleFaintFalse_P201(t *testing.T) {
	s := NewStyle().Faint(false)
	_ = s.Render("test")
}

func TestStyleDimChain_P201(t *testing.T) {
	// Verify Dim and Faint both produce the same effect
	s1 := NewStyle().Dim(true)
	s2 := NewStyle().Faint(true)
	r1 := s1.Render("x")
	r2 := s2.Render("x")
	if r1 != r2 {
		t.Error("Dim and Faint should produce identical output")
	}
}