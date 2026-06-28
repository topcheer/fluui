package app

import (
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// --- Theme management API ---

// ThemeCount returns the number of built-in themes.
func (a *ChatApp) ThemeCount() int {
	return len(theme.Builtin())
}

// ThemeList returns the names of all built-in themes.
func (a *ChatApp) ThemeList() []string {
	builtins := theme.Builtin()
	names := make([]string, len(builtins))
	for i, t := range builtins {
		names[i] = t.Name
	}
	return names
}

// ThemeIndex returns the current theme's index in the built-in list.
func (a *ChatApp) ThemeIndex() int {
	return theme.CurrentIndex()
}

// ThemeName returns the name of the currently active theme.
func (a *ChatApp) ThemeName() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.theme == nil {
		return ""
	}
	return a.theme.Name
}

// SetThemeByIndex sets the active theme by its index in the built-in list.
// Out-of-range indices are ignored.
func (a *ChatApp) SetThemeByIndex(idx int) {
	builtins := theme.Builtin()
	if idx < 0 || idx >= len(builtins) {
		return
	}
	a.SetTheme(builtins[idx])
}

// SetThemeByName sets the active theme by name (case-sensitive).
// Returns true if a matching theme was found.
func (a *ChatApp) SetThemeByName(name string) bool {
	for _, t := range theme.Builtin() {
		if t.Name == name {
			a.SetTheme(t)
			return true
		}
	}
	return false
}

// handleThemeKey checks for Ctrl+] (forward cycle) and Ctrl+\ (backward cycle).
// Returns true if the key was consumed.
func (a *ChatApp) handleThemeKey(key *term.KeyEvent) bool {
	if key.Modifiers&term.ModCtrl == 0 || key.Rune == 0 {
		return false
	}

	// Ctrl+] = cycle theme forward
	if key.Rune == ']' {
		a.CycleTheme()
		return true
	}

	// Ctrl+\ = cycle theme backward
	if key.Rune == '\\' {
		a.CycleThemeBack()
		return true
	}

	return false
}
