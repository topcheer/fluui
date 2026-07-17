package lipgloss

import (
	"strings"
	"testing"
)

func TestPlaceHorizontal_Left_P269(t *testing.T) {
	result := PlaceHorizontal(10, Left, "hello")
	if !strings.HasSuffix(result, strings.Repeat(" ", 5)) {
		t.Errorf("left align should pad right, got: %q", result)
	}
}

func TestPlaceHorizontal_Right_P269(t *testing.T) {
	result := PlaceHorizontal(10, Right, "hello")
	if !strings.HasPrefix(result, strings.Repeat(" ", 5)) {
		t.Errorf("right align should pad left, got: %q", result)
	}
}

func TestPlaceHorizontal_Center_P269(t *testing.T) {
	result := PlaceHorizontal(11, Center, "hello")
	expected := "   hello   "
	if result != expected {
		t.Errorf("center align, expected %q, got %q", expected, result)
	}
}

func TestPlaceHorizontal_TooWide_P269(t *testing.T) {
	result := PlaceHorizontal(3, Left, "hello world")
	if result != "hello world" {
		t.Error("content wider than width should return as-is")
	}
}

func TestPlaceVertical_Top_P269(t *testing.T) {
	result := PlaceVertical(3, Top, "hello")
	if !strings.HasSuffix(result, "\n\n") {
		t.Errorf("top align should pad bottom with newlines, got: %q", result)
	}
}

func TestPlaceVertical_Bottom_P269(t *testing.T) {
	result := PlaceVertical(3, Bottom, "hello")
	if !strings.HasPrefix(result, "\n\n") {
		t.Errorf("bottom align should pad top with newlines, got: %q", result)
	}
}

func TestPlaceVertical_Middle_P269(t *testing.T) {
	result := PlaceVertical(3, Middle, "hello")
	// should have 1 newline before and 1 after
	expected := "\nhello\n"
	if result != expected {
		t.Errorf("middle align, expected %q, got %q", expected, result)
	}
}

func TestPlaceVertical_TooTall_P269(t *testing.T) {
	result := PlaceVertical(1, Top, "line1\nline2\nline3")
	if result != "line1\nline2\nline3" {
		t.Error("content taller than height should return as-is")
	}
}
