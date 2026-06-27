package app

import (
	"testing"

	"github.com/topcheer/fluui/theme"
)

func TestCycleTheme(t *testing.T) {
	theme.SetByIndex(0)
	app := NewChatApp(80, 24)
	app.SetTheme(theme.Dracula())
	if app.Theme().Name != "Dracula" {
		t.Errorf("expected Dracula, got %s", app.Theme().Name)
	}

	// Cycle forward: Dracula → Nord
	next := app.CycleTheme()
	if next.Name != "Nord" {
		t.Errorf("expected Nord, got %s", next.Name)
	}

	// Cycle forward: Nord → Gruvbox
	next = app.CycleTheme()
	if next.Name != "Gruvbox" {
		t.Errorf("expected Gruvbox, got %s", next.Name)
	}
}

func TestCycleThemeWrap(t *testing.T) {
	theme.SetByIndex(0)
	app := NewChatApp(80, 24)
	app.SetTheme(theme.Dracula())
	// Cycle through all 5 themes: Dracula→Nord→Gruvbox→SolarizedDark→TokyoNight
	for _, expected := range []string{"Nord", "Gruvbox", "SolarizedDark", "TokyoNight", "Dracula"} {
		next := app.CycleTheme()
		if next.Name != expected {
			t.Errorf("expected %s, got %s", expected, next.Name)
		}
	}
	// Should be back to Dracula
	if app.Theme().Name != "Dracula" {
		t.Errorf("expected to wrap to Dracula, got %s", app.Theme().Name)
	}
}

func TestCycleThemeBack(t *testing.T) {
	theme.SetByIndex(0)
	app := NewChatApp(80, 24)
	app.SetTheme(theme.Dracula())
	// Start at Dracula → backward wraps to TokyoNight
	prev := app.CycleThemeBack()
	if prev.Name != "TokyoNight" {
		t.Errorf("expected TokyoNight, got %s", prev.Name)
	}
}

func TestThemeToast(t *testing.T) {
	theme.SetByIndex(0)
	app := NewChatApp(80, 24)
	app.SetTheme(theme.Dracula())

	// No toast initially
	_, ok := app.ThemeToast()
	if ok {
		t.Error("expected no toast initially")
	}

	// Cycle should show toast
	app.CycleTheme()
	toast, ok := app.ThemeToast()
	if !ok {
		t.Error("expected toast after cycle")
	}
	if toast != "Nord" {
		t.Errorf("expected toast 'Nord', got %q", toast)
	}
}

func TestSetThemeUpdatesGlobal(t *testing.T) {
	app := NewChatApp(80, 24)
	nord := theme.Nord()
	app.SetTheme(nord)

	// Verify global theme was updated
	if theme.Get().Name != "Nord" {
		t.Errorf("expected global theme Nord, got %s", theme.Get().Name)
	}

	// Reset for other tests
	theme.SetByIndex(0)
}

func TestThemeCyclesAllBuiltin(t *testing.T) {
	all := theme.Builtin()
	theme.SetByIndex(0)
	app := NewChatApp(80, 24)
	app.SetTheme(theme.Dracula())

	// Visit each theme by cycling
	visited := map[string]bool{}
	for i := 0; i < len(all); i++ {
		t2 := app.CycleTheme()
		visited[t2.Name] = true
	}

	for _, b := range all {
		if !visited[b.Name] {
			t.Errorf("theme %s was never visited during cycle", b.Name)
		}
	}

	// Reset
	theme.SetByIndex(0)
}
