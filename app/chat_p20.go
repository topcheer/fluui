package app

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

// P20: CommandPalette + Spinner integration into ChatApp.
// CommandPalette is toggled with Ctrl+P, providing fuzzy command search.
// Spinner shows during AI streaming or loading operations.

// SetCommandPalette attaches a CommandPalette to the ChatApp.
// When attached, Ctrl+P toggles the palette overlay.
func (a *ChatApp) SetCommandPalette(cp *component.CommandPalette) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.commandPalette = cp
}

// CommandPalette returns the attached command palette, or nil.
func (a *ChatApp) CommandPalette() *component.CommandPalette {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.commandPalette
}

// SetSpinner attaches a Spinner to the ChatApp for loading indicators.
func (a *ChatApp) SetSpinner(s *component.Spinner) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.spinner = s
}

// Spinner returns the attached spinner, or nil.
func (a *ChatApp) Spinner() *component.Spinner {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.spinner
}

// IsCommandPaletteVisible reports whether the palette overlay is shown.
func (a *ChatApp) IsCommandPaletteVisible() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.commandPalette == nil {
		return false
	}
	return a.commandPalette.Visible()
}

// ToggleCommandPalette shows or hides the command palette.
// Returns true if the key event was consumed.
func (a *ChatApp) ToggleCommandPalette() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.commandPalette == nil {
		return false
	}
	if a.commandPalette.Visible() {
		a.commandPalette.Hide()
	} else {
		a.commandPalette.Show(a.width/4, 3)
	}
	return true
}

// handleP20Key processes P20-specific key events (Ctrl+P for palette).
// Returns true if the event was consumed.
func (a *ChatApp) handleP20Key(k *term.KeyEvent) bool {
	// Ctrl+P = toggle command palette
	if k.Rune == 'p' && k.Modifiers&term.ModCtrl != 0 {
		return a.ToggleCommandPalette()
	}

	// If palette is visible, route keys to it
	if a.commandPalette != nil && a.commandPalette.Visible() {
		return a.commandPalette.HandleKey(k)
	}

	return false
}

// StartSpinner begins the loading spinner animation with an optional label.
func (a *ChatApp) StartSpinner(label string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.spinner == nil {
		return
	}
	a.spinner.SetLabel(label)
	a.spinner.Start()
}

// StopSpinner stops the loading spinner.
func (a *ChatApp) StopSpinner() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.spinner == nil {
		return
	}
	a.spinner.Stop()
}

// IsSpinnerActive reports whether the spinner is currently animating.
func (a *ChatApp) IsSpinnerActive() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.spinner == nil {
		return false
	}
	return a.spinner.Running()
}

// AddCommand adds a command to the attached command palette.
// Returns false if no palette is attached.
func (a *ChatApp) AddCommand(id, label, category string, fn func()) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.commandPalette == nil {
		return false
	}
	a.commandPalette.AddCommand(component.Command{
		ID:       id,
		Label:    label,
		Category: category,
		Action:   fn,
	})
	return true
}
