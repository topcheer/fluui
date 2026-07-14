package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P173: Tests for new lipgloss-compatible StyleBuilder methods

func TestStyleBuilder_String(t *testing.T) {
	s := NewStyle().Bold().Foreground(buffer.NamedColor(buffer.NamedRed))
	result := s.String()
	if result == "" {
		t.Error("String() should produce non-empty output")
	}
}

func TestStyleBuilder_SetWidth(t *testing.T) {
	s := NewStyle().SetWidth(20)
	if s == nil {
		t.Error("SetWidth should return builder")
	}
}

func TestStyleBuilder_SetHeight(t *testing.T) {
	s := NewStyle().SetHeight(10)
	if s == nil {
		t.Error("SetHeight should return builder")
	}
}

func TestStyleBuilder_WidthChain(t *testing.T) {
	s := NewStyle().Width(20).Height(10).Bold().Render("test")
	if s == "" {
		t.Error("Width/Height chain should work")
	}
}

func TestStyleBuilder_Padding(t *testing.T) {
	s := NewStyle().Padding(1)
	if s == nil {
		t.Error("Padding should return builder")
	}
	s2 := NewStyle().Padding(1, 2)
	if s2 == nil {
		t.Error("Padding(1,2) should return builder")
	}
}

func TestStyleBuilder_Border(t *testing.T) {
	s := NewStyle().Border(nil)
	if s == nil {
		t.Error("Border should return builder")
	}
}

func TestStyleBuilder_HeightChain(t *testing.T) {
	s := NewStyle().Height(5).Width(10).Render("test")
	if s == "" {
		t.Error("Height chain should work")
	}
}

func TestStyleBuilder_MeasureWidth(t *testing.T) {
	s := NewStyle()
	if s.MeasureWidth("hello") != 5 {
		t.Error("MeasureWidth should return display width")
	}
	if s.MeasureWidth("") != 0 {
		t.Error("MeasureWidth of empty string should be 0")
	}
	if s.MeasureWidth("日本語") != 6 {
		t.Error("MeasureWidth should count wide chars as 2")
	}
}