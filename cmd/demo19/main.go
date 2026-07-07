package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// demo19 showcases the KeybindingManager — declarative keyboard shortcuts
// with chord sequences, context scoping, and automatic help generation.
//
// Controls:
//   ctrl+s          Save (global)
//   ctrl+q          Quit
//   ctrl+h          Toggle help overlay
//   ctrl+x ctrl+s   Save all files (chord)
//   ctrl+x ctrl+c   Quit all (chord)
//   ctrl+e          Enter editor mode (changes available bindings)
//   ctrl+m          Enter modal mode
//   esc             Exit current mode / cancel chord
//   ?               Show conflicts

func main() {
	app, err := fluui.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	app.SetTitle("Fluui Demo 19 — Keybinding Manager")

	_, height := app.Size()

	// Create KeybindingManager
	km := component.NewKeybindingManager()

	// UI state
	statusLine := "Ready. Press ctrl+h for help."
	var helpVisible bool
	var outputLines []string

	addOutput := func(s string) {
		outputLines = append(outputLines, s)
		if len(outputLines) > height-6 {
			outputLines = outputLines[1:]
		}
	}

	// Register global bindings
	_ = km.Register("quit", "ctrl+q", "Quit application", func() bool {
		app.Quit()
		return true
	})
	_ = km.Register("help", "ctrl+h", "Toggle help", func() bool {
		helpVisible = !helpVisible
		if helpVisible {
			statusLine = "Help visible. ctrl+h to hide."
		} else {
			statusLine = "Help hidden."
		}
		return true
	})
	_ = km.Register("save", "ctrl+s", "Save current file", func() bool {
		addOutput("[saved] file.txt")
		statusLine = "Saved file.txt"
		return true
	})

	// Chord bindings (Emacs-style)
	_ = km.Register("save-all", "ctrl+x ctrl+s", "Save all files", func() bool {
		addOutput("[saved all] 3 files saved")
		statusLine = "Saved all files"
		return true
	})
	_ = km.Register("quit-all", "ctrl+x ctrl+c", "Quit (chord)", func() bool {
		app.Quit()
		return true
	})

	// Mode switching
	_ = km.Register("enter-editor", "ctrl+e", "Enter editor mode", func() bool {
		km.PushContext("editor")
		addOutput("[mode] entered editor context")
		statusLine = "Editor mode active. ctrl+f=format, esc=exit"
		return true
	})
	_ = km.Register("enter-modal", "ctrl+m", "Enter modal mode", func() bool {
		km.PushContext("modal")
		addLine(outputLines, "[mode] entered modal context")
		statusLine = "Modal mode active. ctrl+d=delete, esc=exit"
		return true
	})

	// Editor context bindings
	_ = km.RegisterIn("editor", "format", "ctrl+f", "Format code", func() bool {
		addOutput("[formatted] code formatted with gofmt")
		statusLine = "Code formatted"
		return true
	})
	_ = km.RegisterIn("editor", "rename", "f2", "Rename symbol", func() bool {
		addOutput("[rename] symbol renamed")
		statusLine = "Symbol renamed"
		return true
	})

	// Modal context bindings
	_ = km.RegisterIn("modal", "delete", "ctrl+d", "Delete item", func() bool {
		addOutput("[deleted] item removed")
		statusLine = "Item deleted"
		return true
	})

	// Escape: pop context
	_ = km.Register("exit-mode", "esc", "Exit current mode", func() bool {
		ctx := km.ActiveContext()
		if ctx != "global" {
			km.PopContext()
			addOutput(fmt.Sprintf("[mode] exited %s context", ctx))
			statusLine = "Back to global context"
			return true
		}
		// Cancel any pending chord
		if km.IsChordActive() {
			km.CancelChord()
			statusLine = "Chord cancelled"
			return true
		}
		return false
	})

	// Show conflicts
	_ = km.Register("conflicts", "?", "Show conflicts", func() bool {
		conflicts := km.CheckConflicts()
		if len(conflicts) == 0 {
			addOutput("[conflicts] none detected")
		} else {
			for _, c := range conflicts {
				addOutput("[conflict] " + c)
			}
		}
		statusLine = fmt.Sprintf("%d conflicts", len(conflicts))
		return true
	})

	app.OnKey(func(k *term.KeyEvent) {
		// Let KeybindingManager handle the key
		cmd, handled := km.Match(k)

		if km.IsChordActive() {
			statusLine = fmt.Sprintf("Chord: %s _ (waiting for next key)", km.ChordPrefix())
		}

		if !handled && cmd == "" && !km.IsChordActive() {
			// Show unhandled key info
			statusLine = fmt.Sprintf("Key: rune=%q key=%d mods=%d", string(k.Rune), k.Key, k.Modifiers)
		}

		app.MarkDirty()
	})

	app.OnPaint(func(buf *buffer.Buffer) {
		buf.Fill(buffer.BlankCell)

		// Title
		buf.DrawText(0, 0, "╔══════════════════════════════════════════════════════════════╗", buffer.Style{})
		titleText := fmt.Sprintf("║  Fluui Demo 19 — Keybinding Manager  [ctx: %-8s] [%2d bindings]  ║",
			km.ActiveContext(), km.BindingCount())
		// Truncate/pad title to fit
		if len(titleText) > 66 {
			titleText = titleText[:63] + "...║"
		}
		buf.DrawText(0, 1, titleText, buffer.Style{Flags: buffer.Bold})
		buf.DrawText(0, 2, "╚══════════════════════════════════════════════════════════════╝", buffer.Style{})

		row := 4

		// Output area
		buf.DrawText(0, row, "── Output ──", buffer.Style{Fg: buffer.NamedColor(buffer.NamedCyan)})
		row++
		for _, line := range outputLines {
			if row >= height-4 {
				break
			}
			buf.DrawText(0, row, line, buffer.Style{})
			row++
		}

		// Help overlay
		if helpVisible {
			helpText := km.HelpText()
			helpLines := strings.Split(strings.TrimSpace(helpText), "\n")
			row = height - 5 - len(helpLines)
			if row < 4 {
				row = 4
			}
			buf.DrawText(0, row, "╭── Help ──────────────────────────────────────────────────────╮", buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow)})
			row++
			for _, line := range helpLines {
				if row >= height-4 {
					break
				}
				displayLine := line
				if len(displayLine) > 61 {
					displayLine = displayLine[:61]
				}
				padded := fmt.Sprintf("│ %-61s │", displayLine)
				buf.DrawText(0, row, padded, buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow)})
				row++
			}
			buf.DrawText(0, row, "╰──────────────────────────────────────────────────────────────╯", buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow)})
		}

		// Status bar
		statusStyle := buffer.Style{Fg: buffer.NamedColor(buffer.NamedGreen)}
		if km.IsChordActive() {
			statusStyle = buffer.Style{Fg: buffer.NamedColor(buffer.NamedYellow), Flags: buffer.Bold}
		}
		buf.DrawText(0, height-2, statusLine, statusStyle)

		// Footer hints
		hints := "ctrl+s=save  ctrl+q=quit  ctrl+h=help  ctrl+e=editor  ctrl+m=modal  ?=conflicts"
		buf.DrawText(0, height-1, hints, buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack)})
	})

	app.OnResize(func(w, h int) {
		height = h
		app.MarkDirty()
	})

	addOutput("[ready] KeybindingManager initialized")
	addOutput(fmt.Sprintf("[info] %d bindings registered across 3 contexts", km.BindingCount()))

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Run error: %v\n", err)
		os.Exit(1)
	}
}

func addLine(lines []string, s string) []string {
	return append(lines, s)
}
