package fluui

import (
	"testing"
)

func TestP47_SetTitle(t *testing.T) {
	// App requires a terminal; test the setter/getter directly
	a := &App{}
	a.SetTitle("My App")
	if a.Title() != "My App" {
		t.Errorf("expected 'My App', got %q", a.Title())
	}
}

func TestP47_SetTitle_Empty(t *testing.T) {
	a := &App{}
	if a.Title() != "" {
		t.Error("expected empty title by default")
	}
}

func TestP47_SetTitle_Multiple(t *testing.T) {
	a := &App{}
	a.SetTitle("First")
	a.SetTitle("Second")
	if a.Title() != "Second" {
		t.Errorf("expected 'Second', got %q", a.Title())
	}
}
