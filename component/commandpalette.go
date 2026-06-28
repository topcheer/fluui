package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/fuzzy"
	"github.com/topcheer/fluui/internal/term"
)

// Command represents a single command entry in the palette.
type Command struct {
	ID       string          // unique identifier
	Label    string          // display label shown to user
	Shortcut string          // optional keyboard shortcut (e.g. "Ctrl+S")
	Category string          // optional grouping category
	Action   func()          // executed when the command is selected
}

// CommandPaletteStyle holds the visual styling for the command palette.
type CommandPaletteStyle struct {
	Border       buffer.Style
	Prompt       buffer.Style
	Input        buffer.Style
	Cursor       buffer.Style
	Normal       buffer.Style
	Matched      buffer.Style
	Shortcut     buffer.Style
	Category     buffer.Style
	HelpText     buffer.Style
}

// DefaultCommandPaletteStyle returns a visually pleasing default style.
func DefaultCommandPaletteStyle() CommandPaletteStyle {
	return CommandPaletteStyle{
		Border:   buffer.Style{Fg: buffer.Color256Val(244)},
		Prompt:   buffer.Style{Fg: buffer.Color256Val(39)},
		Input:    buffer.Style{Fg: buffer.Color256Val(255)},
		Cursor:   buffer.Style{Fg: buffer.Color256Val(39), Flags: buffer.Reverse},
		Normal:   buffer.Style{Fg: buffer.Color256Val(250)},
		Matched:  buffer.Style{Fg: buffer.Color256Val(39), Flags: buffer.Bold},
		Shortcut: buffer.Style{Fg: buffer.Color256Val(241)},
		Category: buffer.Style{Fg: buffer.Color256Val(242)},
		HelpText: buffer.Style{Fg: buffer.Color256Val(238)},
	}
}

// filteredCommand holds a command plus its pre-computed highlight segments.
type filteredCommand struct {
	cmd      Command
	segments []fuzzy.Segment
}

// CommandPalette is a VS Code–style command palette popup.
// It provides fuzzy search over a list of commands with keyboard navigation.
type CommandPalette struct {
	BaseComponent
	mu sync.RWMutex

	commands []Command
	query    string
	filtered []filteredCommand
	cursor   int
	scrollY  int

	visible   bool
	x, y      int
	maxVisible int

	matcher *fuzzy.Matcher
	style   CommandPaletteStyle

	OnExecute   func(cmd Command)
	OnDismiss   func()
}

// NewCommandPalette creates a new command palette with no commands.
func NewCommandPalette() *CommandPalette {
	cp := &CommandPalette{
		commands:   make([]Command, 0),
		maxVisible: 10,
		matcher:    fuzzy.NewMatcher(),
		style:      DefaultCommandPaletteStyle(),
	}
	cp.SetID(GenerateID("commandpalette"))
	cp.recomputeLocked()
	return cp
}

// ─── Commands ───────────────────────────────────────────────────

// SetCommands replaces the command list and re-runs filtering.
func (cp *CommandPalette) SetCommands(cmds []Command) {
	cp.mu.Lock()
	cp.commands = make([]Command, len(cmds))
	copy(cp.commands, cmds)
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// AddCommand appends a single command.
func (cp *CommandPalette) AddCommand(cmd Command) {
	cp.mu.Lock()
	cp.commands = append(cp.commands, cmd)
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// Commands returns a copy of the command list.
func (cp *CommandPalette) Commands() []Command {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	result := make([]Command, len(cp.commands))
	copy(result, cp.commands)
	return result
}

// CommandCount returns the total number of registered commands.
func (cp *CommandPalette) CommandCount() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.commands)
}

// ─── Query & filtering ──────────────────────────────────────────

// SetQuery updates the search query and re-filters.
func (cp *CommandPalette) SetQuery(q string) {
	cp.mu.Lock()
	cp.query = q
	cp.cursor = 0
	cp.scrollY = 0
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// Query returns the current search query.
func (cp *CommandPalette) Query() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.query
}

// FilteredCount returns the number of commands matching the current query.
func (cp *CommandPalette) FilteredCount() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.filtered)
}

// HasResults returns true if there are any matching commands.
func (cp *CommandPalette) HasResults() bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return len(cp.filtered) > 0
}

// FilteredCommands returns the filtered commands (copy).
func (cp *CommandPalette) FilteredCommands() []Command {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	result := make([]Command, len(cp.filtered))
	for i, fc := range cp.filtered {
		result[i] = fc.cmd
	}
	return result
}

// recomputeLocked runs fuzzy ranking and stores highlight segments.
// Caller must hold cp.mu.
func (cp *CommandPalette) recomputeLocked() {
	if len(cp.commands) == 0 {
		cp.filtered = nil
		return
	}

	candidates := make([]string, len(cp.commands))
	for i, c := range cp.commands {
		candidates[i] = c.Label
	}

	results := cp.matcher.Rank(cp.query, candidates)
	cp.filtered = make([]filteredCommand, 0, len(results))
	for _, r := range results {
		idx := r.OriginalIndex
		if idx < 0 || idx >= len(cp.commands) {
			continue
		}
		segments := r.Highlight()
		cp.filtered = append(cp.filtered, filteredCommand{
			cmd:      cp.commands[idx],
			segments: segments,
		})
	}

	// Clamp cursor
	if cp.cursor >= len(cp.filtered) {
		if len(cp.filtered)-1 > 0 {
			cp.cursor = len(cp.filtered) - 1
		} else {
			cp.cursor = 0
		}
	}
	cp.clampScrollLocked()
}

// ─── Cursor navigation ──────────────────────────────────────────

// Cursor returns the current cursor index.
func (cp *CommandPalette) Cursor() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.cursor
}

// SetCursor sets the cursor index. Negative values wrap to bottom;
// values >= len wrap to top.
func (cp *CommandPalette) SetCursor(idx int) {
	cp.mu.Lock()
	n := len(cp.filtered)
	if n == 0 {
		cp.cursor = 0
	} else if idx < 0 {
		cp.cursor = n - 1 // wrap to bottom
	} else if idx >= n {
		cp.cursor = 0 // wrap to top
	} else {
		cp.cursor = idx
	}
	cp.clampScrollLocked()
	cp.mu.Unlock()
}

// MoveDown moves cursor down by one, wrapping to top.
func (cp *CommandPalette) MoveDown() {
	cp.mu.Lock()
	if len(cp.filtered) > 0 {
		cp.cursor++
		if cp.cursor >= len(cp.filtered) {
			cp.cursor = 0
		}
		cp.clampScrollLocked()
	}
	cp.mu.Unlock()
}

// MoveUp moves cursor up by one, wrapping to bottom.
func (cp *CommandPalette) MoveUp() {
	cp.mu.Lock()
	if len(cp.filtered) > 0 {
		cp.cursor--
		if cp.cursor < 0 {
			cp.cursor = len(cp.filtered) - 1
		}
		cp.clampScrollLocked()
	}
	cp.mu.Unlock()
}

// CurrentCommand returns the command at the cursor, or nil if empty.
func (cp *CommandPalette) CurrentCommand() *Command {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	if cp.cursor < 0 || cp.cursor >= len(cp.filtered) {
		return nil
	}
	cmd := cp.filtered[cp.cursor].cmd
	return &cmd
}

// ─── Scroll ─────────────────────────────────────────────────────

// ScrollY returns the current vertical scroll offset.
func (cp *CommandPalette) ScrollY() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.scrollY
}

func (cp *CommandPalette) clampScrollLocked() {
	if cp.cursor < cp.scrollY {
		cp.scrollY = cp.cursor
	}
	if cp.cursor >= cp.scrollY+cp.maxVisible {
		cp.scrollY = cp.cursor - cp.maxVisible + 1
	}
	if cp.scrollY < 0 {
		cp.scrollY = 0
	}
}

// ─── Visibility & position ──────────────────────────────────────

// Show makes the palette visible at the given position.
func (cp *CommandPalette) Show(x, y int) {
	cp.mu.Lock()
	cp.visible = true
	cp.x = x
	cp.y = y
	cp.query = ""
	cp.cursor = 0
	cp.scrollY = 0
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// Hide dismisses the palette.
func (cp *CommandPalette) Hide() {
	cp.mu.Lock()
	wasVisible := cp.visible
	cp.visible = false
	callback := cp.OnDismiss
	cp.mu.Unlock()

	if wasVisible && callback != nil {
		callback()
	}
}

// Visible returns whether the palette is currently shown.
func (cp *CommandPalette) Visible() bool {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.visible
}

// Position returns the (x, y) screen position.
func (cp *CommandPalette) Position() (int, int) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.x, cp.y
}

// SetPosition updates the screen position without toggling visibility.
func (cp *CommandPalette) SetPosition(x, y int) {
	cp.mu.Lock()
	cp.x = x
	cp.y = y
	cp.mu.Unlock()
}

// ─── Selection ──────────────────────────────────────────────────

// Select executes the command at the cursor and dismisses the palette.
func (cp *CommandPalette) Select() {
	cp.mu.Lock()
	if cp.cursor < 0 || cp.cursor >= len(cp.filtered) {
		cp.mu.Unlock()
		return
	}
	cmd := cp.filtered[cp.cursor].cmd
	onExecute := cp.OnExecute
	cp.visible = false
	cp.mu.Unlock()

	if cmd.Action != nil {
		cmd.Action()
	}
	if onExecute != nil {
		onExecute(cmd)
	}
}

// ─── Configuration ──────────────────────────────────────────────

// SetMaxVisible sets the maximum visible result rows.
func (cp *CommandPalette) SetMaxVisible(n int) {
	cp.mu.Lock()
	if n < 1 {
		n = 1
	}
	cp.maxVisible = n
	cp.clampScrollLocked()
	cp.mu.Unlock()
}

// MaxVisible returns the maximum visible result rows.
func (cp *CommandPalette) MaxVisible() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.maxVisible
}

// SetStyle updates the visual style.
func (cp *CommandPalette) SetStyle(s CommandPaletteStyle) {
	cp.mu.Lock()
	cp.style = s
	cp.mu.Unlock()
}

// Style returns the current visual style.
func (cp *CommandPalette) Style() CommandPaletteStyle {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.style
}

// SetCaseSensitive toggles case-sensitive matching.
func (cp *CommandPalette) SetCaseSensitive(b bool) {
	cp.mu.Lock()
	cp.matcher.SetCaseSensitive(b)
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// ─── Input ──────────────────────────────────────────────────────

// InsertRune appends a rune to the query.
func (cp *CommandPalette) InsertRune(r rune) {
	cp.mu.Lock()
	cp.query += string(r)
	cp.cursor = 0
	cp.scrollY = 0
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// Backspace removes the last rune from the query.
func (cp *CommandPalette) Backspace() {
	cp.mu.Lock()
	if len(cp.query) > 0 {
		runes := []rune(cp.query)
		cp.query = string(runes[:len(runes)-1])
		cp.cursor = 0
		cp.scrollY = 0
		cp.recomputeLocked()
	}
	cp.mu.Unlock()
}

// ─── Keyboard ───────────────────────────────────────────────────

// HandleKey processes a keyboard event. Returns true if the key was handled.
func (cp *CommandPalette) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}

	switch ev.Key {
	case term.KeyEscape:
		cp.Hide()
		return true
	case term.KeyEnter:
		cp.Select()
		return true
	case term.KeyUp:
		cp.MoveUp()
		return true
	case term.KeyDown:
		cp.MoveDown()
		return true
	case term.KeyTab:
		cp.MoveDown()
		return true
	case term.KeyBackspace:
		cp.Backspace()
		return true
	case term.KeyLeft, term.KeyRight:
		return false // unhandled
	default:
		// Printable character?
		if ev.Key == term.KeyUnknown && ev.Rune != 0 {
			cp.InsertRune(ev.Rune)
			return true
		}
		return false
	}
}

// ─── Component interface ────────────────────────────────────────

// Measure calculates the preferred size of the palette.
func (cp *CommandPalette) Measure(cs Constraints) Size {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	width := 40
	for _, c := range cp.commands {
		labelLen := len([]rune(c.Label))
		if c.Shortcut != "" {
			labelLen += 2 + len([]rune(c.Shortcut))
		}
		if c.Category != "" {
			labelLen += 2 + len([]rune(c.Category))
		}
		if labelLen+4 > width {
			width = labelLen + 4
		}
	}
	if cs.MaxWidth > 0 && width > cs.MaxWidth {
		width = cs.MaxWidth
	}
	if width < 20 {
		width = 20
	}

	height := 3 // border + input + help
	items := cp.maxVisible
	if len(cp.filtered) < items {
		items = len(cp.filtered)
	}
	height += items
	if cs.MaxHeight > 0 && height > cs.MaxHeight {
		height = cs.MaxHeight
	}

	return Size{W: width, H: height}
}

// Paint renders the palette into the buffer.
func (cp *CommandPalette) Paint(buf *buffer.Buffer) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	if !cp.visible {
		return
	}

	b := cp.bounds
	if b.W < 3 || b.H < 3 {
		return
	}

	style := cp.style

	// ─── Border ───
	borderY := b.Y
	buf.SetCell(b.X, borderY, buffer.NewCell('┌', style.Border))
	buf.SetCell(b.X+b.W-1, borderY, buffer.NewCell('┐', style.Border))
	for x := b.X + 1; x < b.X+b.W-1; x++ {
		buf.SetCell(x, borderY, buffer.NewCell('─', style.Border))
	}

	bottomY := b.Y + b.H - 1
	buf.SetCell(b.X, bottomY, buffer.NewCell('└', style.Border))
	buf.SetCell(b.X+b.W-1, bottomY, buffer.NewCell('┘', style.Border))
	for x := b.X + 1; x < b.X+b.W-1; x++ {
		buf.SetCell(x, bottomY, buffer.NewCell('─', style.Border))
	}
	for y := b.Y + 1; y < bottomY; y++ {
		buf.SetCell(b.X, y, buffer.NewCell('│', style.Border))
		buf.SetCell(b.X+b.W-1, y, buffer.NewCell('│', style.Border))
	}

	// ─── Input row (y = b.Y + 1) ───
	inputY := b.Y + 1
	buf.SetCell(b.X+1, inputY, buffer.NewCell('>', style.Prompt))
	queryRunes := []rune(cp.query)
	for i, r := range queryRunes {
		x := b.X + 3 + i
		if x >= b.X+b.W-2 {
			break
		}
		buf.SetCell(x, inputY, buffer.NewCell(r, style.Input))
	}
	// Cursor indicator
	cursorX := b.X + 3 + len(queryRunes)
	if cursorX < b.X+b.W-1 {
		buf.SetCell(cursorX, inputY, buffer.NewCell(' ', style.Cursor))
	}

	// Separator after input
	sepY := inputY + 1
	if sepY < bottomY {
		for x := b.X + 1; x < b.X+b.W-1; x++ {
			buf.SetCell(x, sepY, buffer.NewCell('─', style.Border))
		}
	}

	// ─── Results ───
	if len(cp.filtered) == 0 {
		msg := "No matching commands"
		msgRunes := []rune(msg)
		for i, r := range msgRunes {
			x := b.X + 2 + i
			if x >= b.X+b.W-2 {
				break
			}
			buf.SetCell(x, sepY+1, buffer.NewCell(r, style.HelpText))
		}
		return
	}

	itemsStart := sepY + 1
	itemsAvail := bottomY - itemsStart
	if itemsAvail <= 0 {
		return
	}

	maxItems := itemsAvail
	if maxItems > cp.maxVisible {
		maxItems = cp.maxVisible
	}

	end := cp.scrollY + maxItems
	if end > len(cp.filtered) {
		end = len(cp.filtered)
	}

	for i := cp.scrollY; i < end; i++ {
		fc := cp.filtered[i]
		rowY := itemsStart + (i - cp.scrollY)
		isSelected := i == cp.cursor
		rowStyle := style.Normal
		if isSelected {
			rowStyle = style.Cursor
		}

		// Draw background fill for selected row
		if isSelected {
			for x := b.X + 1; x < b.X+b.W-1; x++ {
				buf.SetCell(x, rowY, buffer.NewCell(' ', rowStyle))
			}
		}

		// Draw label with highlight segments
		x := b.X + 2
		maxX := b.X + b.W - 2
		for _, seg := range fc.segments {
			segStyle := style.Normal
			if seg.Matched {
				segStyle = style.Matched
			}
			if isSelected {
				segStyle = style.Cursor
			}
			for _, r := range seg.Text {
				if x >= maxX {
					break
				}
				buf.SetCell(x, rowY, buffer.NewCell(r, segStyle))
				x++
			}
		}

		// Draw shortcut on the right side
		if fc.cmd.Shortcut != "" {
			shortRunes := []rune(fc.cmd.Shortcut)
			shortX := b.X + b.W - 3 - len(shortRunes)
			if shortX > x+1 {
				shortStyle := style.Shortcut
				if isSelected {
					shortStyle = style.Cursor
				}
				for i, r := range shortRunes {
					buf.SetCell(shortX+i, rowY, buffer.NewCell(r, shortStyle))
				}
			}
		}
	}
}

// Children returns nil — the palette has no child components.
func (cp *CommandPalette) Children() []Component { return nil }

// String returns a debug description.
func (cp *CommandPalette) String() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return "CommandPalette(commands=" + itoa(len(cp.commands)) +
		",filtered=" + itoa(len(cp.filtered)) + ")"
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min2(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SetOnExecute sets the callback invoked when a command is executed.
func (cp *CommandPalette) SetOnExecute(fn func(cmd Command)) {
	cp.mu.Lock()
	cp.OnExecute = fn
	cp.mu.Unlock()
}

// SetOnDismiss sets the callback invoked when the palette is dismissed.
func (cp *CommandPalette) SetOnDismiss(fn func()) {
	cp.mu.Lock()
	cp.OnDismiss = fn
	cp.mu.Unlock()
}

// Reset clears the query and resets cursor.
func (cp *CommandPalette) Reset() {
	cp.mu.Lock()
	cp.query = ""
	cp.cursor = 0
	cp.scrollY = 0
	cp.recomputeLocked()
	cp.mu.Unlock()
}

// LastUpdate returns the current time (used for spinner integration demos).
func (cp *CommandPalette) LastUpdate() time.Time {
	return time.Now()
}