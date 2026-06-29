package app

import (
	"io"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- P28-A: Clipboard key bindings and bridge ---

// handleClipboardKey processes clipboard-related key events.
// It handles:
//   - Ctrl+Shift+C: copy the current selection to the system clipboard via OSC52
//   - Ctrl+Shift+V: request clipboard paste from the terminal via OSC52 query
//
// buf is the current rendered buffer for extracting selection text.
// w is the terminal writer for outputting OSC52 sequences.
// Returns true if the key was consumed.
func (a *ChatApp) handleClipboardKey(key *term.KeyEvent, buf *buffer.Buffer, w io.Writer) bool {
	if key == nil {
		return false
	}

	// Ctrl+Shift+C = copy selection
	if key.Modifiers&term.ModCtrl != 0 && key.Modifiers&term.ModShift != 0 && key.Rune == 'c' {
		return a.copySelectionToWriter(buf, w)
	}

	// Ctrl+Shift+V = paste from clipboard
	if key.Modifiers&term.ModCtrl != 0 && key.Modifiers&term.ModShift != 0 && key.Rune == 'v' {
		return a.requestPaste(w)
	}

	return false
}

// copySelectionToWriter writes the OSC52 copy sequence for the current
// selection to the given writer (typically the terminal).
// Returns false if there is no active selection or no text to copy.
func (a *ChatApp) copySelectionToWriter(buf *buffer.Buffer, w io.Writer) bool {
	a.mu.Lock()
	sm := a.selectionMgr
	clipCfg := a.clipboardConfig
	a.mu.Unlock()

	if sm == nil || buf == nil || w == nil {
		return false
	}

	if !sm.HasSelection() {
		return false
	}

	// Use ClipboardConfig.CopyToWriter if available, otherwise raw OSC52.
	if clipCfg != nil {
		text := sm.GetSelectedText(buf)
		if text == "" {
			return false
		}
		if err := clipCfg.CopyToWriter(w, text); err != nil {
			return false
		}
		return true
	}

	// Fallback: direct OSC52.
	seq := sm.CopySelection(buf)
	if seq == "" {
		return false
	}
	_, _ = io.WriteString(w, seq)
	return true
}

// requestPaste sends an OSC52 paste query to the terminal.
// The terminal's response will arrive later as a paste event in the input stream.
// Returns false if the writer is nil.
func (a *ChatApp) requestPaste(w io.Writer) bool {
	if w == nil {
		return false
	}
	_, _ = io.WriteString(w, term.PasteQuery())
	return true
}

// CopySelectionToWriter is the public API for copying the current selection.
// It writes the OSC52 escape sequence to the given writer.
func (a *ChatApp) CopySelectionToWriter(buf *buffer.Buffer, w io.Writer) bool {
	return a.copySelectionToWriter(buf, w)
}

// PasteFromClipboard requests a paste from the system clipboard.
// The response arrives as a paste event in the input stream.
func (a *ChatApp) PasteFromClipboard(w io.Writer) bool {
	return a.requestPaste(w)
}
