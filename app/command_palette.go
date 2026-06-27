package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// Command represents a single action in the command palette.
type Command struct {
	ID       string   // unique identifier
	Title    string   // display name, e.g. "Change Theme: Nord"
	Category string   // grouping: "Theme", "Edit", "View", etc.
	Hint     string   // optional shortcut hint, e.g. "Ctrl+T"
	Action   func()   // callback when selected
}

// scoredCommand pairs a command with its fuzzy match score.
type scoredCommand struct {
	cmd   Command
	score int
}

// CommandPalette manages the Ctrl+P fuzzy search UI.
// It maintains a list of registered commands and filters them
// based on user input using subsequence fuzzy matching.
type CommandPalette struct {
	active   bool
	query    string
	commands []Command
	filtered []scoredCommand
	selected int
}

// NewCommandPalette creates an empty, inactive CommandPalette.
func NewCommandPalette() *CommandPalette {
	return &CommandPalette{
		selected: 0,
	}
}

// Register adds a single command to the palette.
func (cp *CommandPalette) Register(cmd Command) {
	cp.commands = append(cp.commands, cmd)
}

// RegisterMany adds multiple commands at once.
func (cp *CommandPalette) RegisterMany(cmds []Command) {
	cp.commands = append(cp.commands, cmds...)
}

// Commands returns all registered commands (for testing/inspection).
func (cp *CommandPalette) Commands() []Command {
	return cp.commands
}

// IsActive reports whether the palette is currently visible.
func (cp *CommandPalette) IsActive() bool {
	return cp.active
}

// Query returns the current search query.
func (cp *CommandPalette) Query() string {
	return cp.query
}

// Open shows the palette and resets the query.
func (cp *CommandPalette) Open() {
	cp.active = true
	cp.query = ""
	cp.selected = 0
	cp.Filter()
}

// Close hides the palette and resets state.
func (cp *CommandPalette) Close() {
	cp.active = false
	cp.query = ""
	cp.filtered = nil
	cp.selected = 0
}

// FilteredCount returns the number of commands matching the current query.
func (cp *CommandPalette) FilteredCount() int {
	return len(cp.filtered)
}

// SelectedIndex returns the current selection index (0-based).
func (cp *CommandPalette) SelectedIndex() int {
	return cp.selected
}

// FilteredCommands returns the currently filtered commands in display order.
func (cp *CommandPalette) FilteredCommands() []Command {
	result := make([]Command, len(cp.filtered))
	for i, sc := range cp.filtered {
		result[i] = sc.cmd
	}
	return result
}

// Filter recomputes filtered commands based on the current query.
// Uses fuzzy subsequence matching with scoring:
//   - Consecutive match bonus (+5 per consecutive char)
//   - Start-of-word bonus (+10 for matching first char of a word)
//   - Exact substring bonus (+20)
func (cp *CommandPalette) Filter() {
	if cp.query == "" {
		cp.filtered = make([]scoredCommand, len(cp.commands))
		for i, cmd := range cp.commands {
			cp.filtered[i] = scoredCommand{cmd: cmd, score: 0}
		}
		cp.selected = 0
		return
	}

	lowerQuery := strings.ToLower(cp.query)
	cp.filtered = cp.filtered[:0]

	for _, cmd := range cp.commands {
		score := fuzzyScore(strings.ToLower(cmd.Title), lowerQuery)
		if score >= 0 {
			cp.filtered = append(cp.filtered, scoredCommand{cmd: cmd, score: score})
		}
	}

	// Sort by score descending, then alphabetically by title
	sort.SliceStable(cp.filtered, func(i, j int) bool {
		if cp.filtered[i].score != cp.filtered[j].score {
			return cp.filtered[i].score > cp.filtered[j].score
		}
		return cp.filtered[i].cmd.Title < cp.filtered[j].cmd.Title
	})

	if cp.selected >= len(cp.filtered) {
		cp.selected = 0
	}
}

// fuzzyScore returns a match score for query against target using
// subsequence matching. Returns -1 if query is not a subsequence of target.
//
// Scoring:
//   - Base: 1 point per matched character
//   - Consecutive match: +5 bonus per consecutive character
//   - Start-of-word match: +10 bonus (after space or at position 0)
//   - Exact substring: +20 bonus
func fuzzyScore(target, query string) int {
	if len(query) == 0 {
		return 0
	}

	// Check for exact substring — big bonus
	substringIdx := strings.Index(target, query)
	if substringIdx >= 0 {
		score := 20 + len(query)
		if substringIdx == 0 {
			score += 10 // match at start
		}
		return score
	}

	// Subsequence matching
	score := 0
	qi := 0
	consecutive := 0

	for ti := 0; ti < len(target) && qi < len(query); ti++ {
		if target[ti] == query[qi] {
			// Base point
			score++

			// Consecutive bonus
			if ti > 0 && qi > 0 && target[ti-1] == query[qi-1] {
				consecutive++
				score += consecutive * 5
			} else {
				consecutive = 0
			}

			// Start-of-word bonus
			if ti == 0 || target[ti-1] == ' ' {
				score += 10
			}

			qi++
		} else {
			consecutive = 0
		}
	}

	if qi < len(query) {
		return -1 // not all query chars matched
	}

	return score
}

// HandleKey processes keyboard input while the palette is active.
// Returns true if the key was consumed.
func (cp *CommandPalette) HandleKey(key *term.KeyEvent) bool {
	if !cp.active {
		return false
	}

	// Escape: close palette
	if key.Key == term.KeyEscape {
		cp.Close()
		return true
	}

	// Enter: execute selected command
	if key.Key == term.KeyEnter {
		if cp.selected >= 0 && cp.selected < len(cp.filtered) {
			cmd := cp.filtered[cp.selected].cmd
			cp.Close()
			if cmd.Action != nil {
				cmd.Action()
			}
		}
		return true
	}

	// Arrow Up: previous command
	if key.Key == term.KeyUp {
		if len(cp.filtered) > 0 {
			cp.selected--
			if cp.selected < 0 {
				cp.selected = len(cp.filtered) - 1
			}
		}
		return true
	}

	// Arrow Down: next command
	if key.Key == term.KeyDown {
		if len(cp.filtered) > 0 {
			cp.selected++
			if cp.selected >= len(cp.filtered) {
				cp.selected = 0
			}
		}
		return true
	}

	// Ctrl+P again: also navigate down (convenience)
	if key.Modifiers&term.ModCtrl != 0 && (key.Rune == 'p' || key.Rune == 'P') {
		if key.Modifiers&term.ModShift != 0 {
			// Ctrl+Shift+P: go up
			if len(cp.filtered) > 0 {
				cp.selected--
				if cp.selected < 0 {
					cp.selected = len(cp.filtered) - 1
				}
			}
		} else {
			if len(cp.filtered) > 0 {
				cp.selected++
				if cp.selected >= len(cp.filtered) {
					cp.selected = 0
				}
			}
		}
		return true
	}

	// Backspace: remove last char from query
	if key.Key == term.KeyBackspace {
		if len(cp.query) > 0 {
			runes := []rune(cp.query)
			cp.query = string(runes[:len(runes)-1])
			cp.Filter()
		}
		return true
	}

	// Printable character: append to query and refilter
	if key.Rune != 0 && key.Rune >= 0x20 && key.Modifiers&term.ModCtrl == 0 {
		cp.query += string(key.Rune)
		cp.Filter()
		return true
	}

	return false
}

// Paint draws the command palette at the top of the screen.
// It shows the query input, filtered commands with selection highlight,
// and match count.
func (cp *CommandPalette) Paint(buf *buffer.Buffer, w, h int) {
	if !cp.active || w <= 0 || h <= 0 {
		return
	}

	t := theme.Get()

	// Palette occupies up to 10 lines at the top
	paletteH := 10
	if paletteH > h {
		paletteH = h
	}

	// Draw background
	bgCell := buffer.Cell{Rune: ' ', Width: 1, Bg: t.SearchBarBg}
	for y := 0; y < paletteH; y++ {
		for x := 0; x < w; x++ {
			buf.SetCell(x, y, bgCell)
		}
	}

	// Line 0: query prompt "> query"
	promptText := "> " + cp.query
	promptStyle := buffer.Style{Fg: t.SearchBarFg, Bg: t.SearchBarBg, Flags: buffer.Bold}
	x := 0
	for _, r := range promptText {
		if x >= w-1 {
			break
		}
		buf.SetCell(x, 0, buffer.Cell{Rune: r, Width: 1, Fg: promptStyle.Fg, Bg: promptStyle.Bg, Flags: promptStyle.Flags})
		x++
	}
	// Draw cursor
	if x < w {
		buf.SetCell(x, 0, buffer.Cell{Rune: '_', Width: 1, Fg: t.Accent, Bg: t.SearchBarBg, Flags: buffer.Bold})
	}

	// Separator line (line 1)
	sepStyle := buffer.Style{Fg: t.Border, Bg: t.SearchBarBg}
	for x := 0; x < w; x++ {
		buf.SetCell(x, 1, buffer.Cell{Rune: '-', Width: 1, Fg: sepStyle.Fg, Bg: sepStyle.Bg})
	}

	// Command list (lines 2..paletteH-1)
	maxItems := paletteH - 2
	for i := 0; i < len(cp.filtered) && i < maxItems; i++ {
		y := i + 2
		sc := cp.filtered[i]
		cmd := sc.cmd

		isSelected := i == cp.selected
		var fg, bg buffer.Color
		var flags buffer.StyleFlags

		if isSelected {
			fg = t.SearchBarBg  // inverted
			bg = t.Accent       // highlighted background
			flags = buffer.Bold
		} else {
			fg = t.SearchBarFg
			bg = t.SearchBarBg
		}

		// Category prefix
		catText := fmt.Sprintf("[%s] ", cmd.Category)
		catStyle := buffer.Style{Fg: t.Accent, Bg: bg, Flags: flags}
		if isSelected {
			catStyle.Fg = t.SearchBarBg
		}

		x := 1 // left padding
		for _, r := range catText {
			if x >= w-1 {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: catStyle.Fg, Bg: bg, Flags: catStyle.Flags})
			x++
		}

		// Title
		for _, r := range cmd.Title {
			if x >= w-1 {
				break
			}
			buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: fg, Bg: bg, Flags: flags})
			x++
		}

		// Hint (right-aligned)
		if cmd.Hint != "" {
			hintW := buffer.StringWidth(cmd.Hint)
			hintX := w - hintX_padding - hintW
			if hintX > x+2 {
				for _, r := range cmd.Hint {
					if hintX >= w-1 {
						break
					}
					hintStyle := buffer.Style{Fg: t.Muted, Bg: bg, Flags: 0}
					if isSelected {
						hintStyle.Fg = t.SearchBarBg
					}
					buf.SetCell(hintX, y, buffer.Cell{Rune: r, Width: 1, Fg: hintStyle.Fg, Bg: bg})
					hintX++
				}
			}
		}
	}

	// Footer: match count
	footerY := paletteH - 1
	if footerY > 0 {
		footer := fmt.Sprintf("  %d commands", len(cp.filtered))
		if cp.query != "" {
			footer = fmt.Sprintf("  %d matches for '%s'", len(cp.filtered), cp.query)
		}
		footerStyle := buffer.Style{Fg: t.Muted, Bg: t.SearchBarBg}
		x := 0
		for _, r := range footer {
			if x >= w-1 {
				break
			}
			buf.SetCell(x, footerY, buffer.Cell{Rune: r, Width: 1, Fg: footerStyle.Fg, Bg: footerStyle.Bg})
			x++
		}
	}
}

const hintX_padding = 2

// DefaultCommands returns a set of standard commands for a ChatApp.
// The action callbacks are nil — the caller should set them after registration
// or use this as a template.
func DefaultCommands() []Command {
	commands := []Command{
		{ID: "theme.dracula", Title: "Change Theme: Dracula", Category: "Theme", Hint: "Ctrl+T"},
		{ID: "theme.nord", Title: "Change Theme: Nord", Category: "Theme", Hint: "Ctrl+T"},
		{ID: "theme.gruvbox", Title: "Change Theme: Gruvbox", Category: "Theme", Hint: "Ctrl+T"},
		{ID: "theme.solarized", Title: "Change Theme: Solarized Dark", Category: "Theme", Hint: "Ctrl+T"},
		{ID: "theme.tokyo", Title: "Change Theme: Tokyo Night", Category: "Theme", Hint: "Ctrl+T"},
		{ID: "search.toggle", Title: "Toggle Search", Category: "Edit", Hint: "Ctrl+F"},
		{ID: "conv.clear", Title: "Clear Conversation", Category: "Edit"},
		{ID: "yank.last", Title: "Yank Last Block", Category: "Edit"},
		{ID: "help", Title: "Help", Category: "View", Hint: "?"},
		{ID: "quit", Title: "Quit", Category: "App", Hint: "Ctrl+C"},
	}
	return commands
}
