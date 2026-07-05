package fluui

import (
	"testing"

	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/render"
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

func TestP51_OnFocus(t *testing.T) {
	a := &App{}
	called := false
	a.OnFocus(func(focused bool) {
		called = true
	})
	if called {
		t.Error("handler should not be called yet")
	}
}

func TestP51_SetSyncOutput(t *testing.T) {
	// Create a minimal renderer for testing
	bw := &nopWriter{}
	tw := term.NewWriter(bw, term.ProfileTrue)
	r := render.New(tw, 10, 5)
	a := &App{renderer: r}

	a.SetSyncOutput(true)
	// Verify it was set on the renderer
	if !r.SyncOutput() {
		t.Error("expected sync output enabled")
	}

	a.SetSyncOutput(false)
	if r.SyncOutput() {
		t.Error("expected sync output disabled")
	}
}

type nopWriter struct{}

func (nw *nopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
