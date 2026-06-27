package app

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// InputLine is the bottom user input line component.
// It handles text entry, cursor movement, editing shortcuts,
// submits on Enter, and supports input history navigation via Up/Down arrows.
type InputLine struct {
	component.BaseComponent

	buf      []rune // input content
	cursor   int    // cursor position (rune index into buf)
	prompt   string // prefix prompt, e.g. "> "
	onSubmit func(text string)

	// history
	history      []string // submitted message history
	historyIdx   int      // current position in history (-1 = not browsing)
	draft        []rune   // unsaved text saved when navigating history
	maxHistory   int      // max entries (0 = unlimited)

	// styles
	promptStyle buffer.Style
	textStyle   buffer.Style

	// tab completion (nil if not configured)
	completion *CompletionManager
}

// NewInputLine creates an InputLine with the given prompt and no submit handler.
func NewInputLine(prompt string) *InputLine {
	return NewInputLineWithHandler(prompt, nil)
}

// NewInputLineWithHandler creates an InputLine whose onSubmit callback is
// invoked when the user presses Enter. After submission the buffer is cleared.
func NewInputLineWithHandler(prompt string, onSubmit func(string)) *InputLine {
	return &InputLine{
		prompt:     prompt,
		onSubmit:   onSubmit,
		maxHistory: 100,
		promptStyle: buffer.Style{
			Fg:    theme.Get().PromptFg,
			Flags: buffer.Bold,
		},
		textStyle: buffer.Style{
			Fg: theme.Get().Fg,
		},
	}
}

// --- Component interface ---

// Measure returns the desired size: height is always 1, width is the
// display width of prompt + content.
func (i *InputLine) Measure(cs component.Constraints) component.Size {
	w := buffer.StringWidth(i.prompt) + len(i.buf)
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return component.Size{W: w, H: 1}
}

// Paint draws the prompt, text content, and a reverse-video cursor.
func (i *InputLine) Paint(buf *buffer.Buffer) {
	bounds := i.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	x := bounds.X
	y := bounds.Y

	// Draw prompt.
	x = buf.DrawText(x, y, i.prompt, i.promptStyle)

	// Draw text content.
	for _, r := range i.buf {
		if x >= bounds.X+bounds.W {
			break
		}
		cell := buffer.Cell{
			Rune:  r,
			Width: buffer.RuneWidth(r),
			Fg:    i.textStyle.Fg,
			Bg:    i.textStyle.Bg,
			Flags: i.textStyle.Flags,
		}
		buf.SetCell(x, y, cell)
		x += cell.Width
	}

	// Draw cursor: reverse-video block at cursor position.
	cursorX := bounds.X + buffer.StringWidth(i.prompt) + i.runeWidthBefore(i.cursor)
	if cursorX < bounds.X+bounds.W {
		var cur buffer.Cell
		if i.cursor < len(i.buf) {
			// Cursor on existing character.
			cur = buffer.Cell{
				Rune:  i.buf[i.cursor],
				Width: buffer.RuneWidth(i.buf[i.cursor]),
				Fg:    i.textStyle.Bg, // swap fg/bg for reverse
				Bg:    i.textStyle.Fg,
				Flags: i.textStyle.Flags | buffer.Reverse,
			}
		} else {
			// Cursor at end — empty block.
			cur = buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    i.textStyle.Fg,
				Flags: buffer.Reverse,
			}
		}
		buf.SetCell(cursorX, y, cur)
	}
}

// --- Input handling ---

// HandleKey processes a key event. Returns true if the key was consumed.
//
// Supported keys:
//   - Printable characters: insert at cursor
//   - Backspace: delete character before cursor
//   - Left / Right: move cursor
//   - Up / Down: navigate input history
//   - Home / End: cursor to start / end
//   - Ctrl+A: cursor to start
//   - Ctrl+E: cursor to end
//   - Ctrl+U: clear all
//   - Ctrl+W: delete previous word
//   - Enter: call onSubmit, add to history, then clear
func (i *InputLine) HandleKey(key *term.KeyEvent) bool {
	// --- Ctrl shortcuts ---
	if key.Modifiers == term.ModCtrl && key.Rune != 0 {
		switch key.Rune {
		case 'a':
			i.cursor = 0
			return true
		case 'e':
			i.cursor = len(i.buf)
			return true
		case 'u':
			i.buf = i.buf[:0]
			i.cursor = 0
			i.historyIdx = -1
			return true
		case 'w':
			i.deleteWordBack()
			return true
		}
	}

	// --- Special keys ---
	switch key.Key {
	case term.KeyEnter:
		// If completion popup is active, accept the selection instead of submitting.
		if i.isCompletionActive() {
			i.completion.Cancel()
			// Fall through to normal Enter handling (submit).
		}

		text := string(i.buf)
		if i.onSubmit != nil {
			i.onSubmit(text)
		}
		// Add to history if non-empty.
		if len(text) > 0 {
			i.AddHistory(text)
		}
		i.Clear()
		return true

	case term.KeyBackspace:
		// Cancel completion on backspace.
		if i.isCompletionActive() {
			i.completion.Cancel()
		}
		if i.cursor > 0 {
			i.buf = append(i.buf[:i.cursor-1], i.buf[i.cursor:]...)
			i.cursor--
		}
		return true

	case term.KeyLeft:
		if i.isCompletionActive() {
			i.completion.Cancel()
		}
		if i.cursor > 0 {
			i.cursor--
		}
		return true

	case term.KeyRight:
		if i.isCompletionActive() {
			i.completion.Cancel()
		}
		if i.cursor < len(i.buf) {
			i.cursor++
		}
		return true

	case term.KeyUp:
		if i.isCompletionActive() {
			i.completion.Cancel()
		}
		i.navigateHistory(-1)
		return true

	case term.KeyDown:
		if i.isCompletionActive() {
			i.completion.Cancel()
		}
		i.navigateHistory(1)
		return true

	case term.KeyHome:
		i.cursor = 0
		return true

	case term.KeyEnd:
		i.cursor = len(i.buf)
		return true

	case term.KeyTab:
		return i.handleTab()

	case term.KeyBacktab:
		return i.handleShiftTab()

	case term.KeyEscape:
		return i.handleEscape()
	}

	// --- Printable character ---
	if key.Rune != 0 && key.Key == term.KeyUnknown && key.Modifiers == 0 {
		// Any typed character exits history browsing.
		i.historyIdx = -1
		i.insertRune(key.Rune)
		return true
	}

	return false
}

// --- Public API ---

// Text returns the current input text.
func (i *InputLine) Text() string { return string(i.buf) }

// SetText replaces the input text and places cursor at the end.
func (i *InputLine) SetText(s string) {
	i.buf = []rune(s)
	i.cursor = len(i.buf)
}

// Clear empties the input buffer and resets cursor to 0.
// Also exits history browsing mode.
func (i *InputLine) Clear() {
	i.buf = i.buf[:0]
	i.cursor = 0
	i.historyIdx = -1
}

// Cursor returns the current cursor position (rune index).
func (i *InputLine) Cursor() int { return i.cursor }

// Len returns the number of runes in the input buffer.
func (i *InputLine) Len() int { return len(i.buf) }

// Empty reports whether the input buffer is empty.
func (i *InputLine) Empty() bool { return len(i.buf) == 0 }

// InsertText inserts a string at the current cursor position.
// Used for paste operations (bracketed paste, OSC52 clipboard paste).
func (i *InputLine) InsertText(s string) {
	for _, r := range s {
		i.insertRune(r)
	}
	i.historyIdx = -1
}

// --- Tab Completion ---

// SetCompletionManager attaches a completion manager for Tab/Shift+Tab handling.
func (i *InputLine) SetCompletionManager(cm *CompletionManager) {
	i.completion = cm
}

// CompletionManager returns the attached completion manager, or nil.
func (i *InputLine) CompletionManager() *CompletionManager {
	return i.completion
}

// handleTab processes a Tab key event.
// If completion popup is not active, it triggers completion from the current word.
// If active, it cycles to the next candidate and applies it as a preview.
// Returns true if the key was consumed.
func (i *InputLine) handleTab() bool {
	if i.completion == nil {
		return false
	}

	if !i.completion.Active() {
		// Start completion from current cursor word.
		prefix := ExtractCompletionPrefix(i.Text(), i.cursor)
		if prefix == "" {
			return false
		}
		return i.completion.Start(prefix)
	}

	// Cycle to next candidate.
	item, ok := i.completion.CycleNext()
	if !ok {
		return false
	}

	// Apply the candidate as a preview: replace prefix with insert text.
	i.applyCompletion(item)
	return true
}

// handleShiftTab processes a Shift+Tab (BackTab) key event.
// Cycles to the previous completion candidate.
func (i *InputLine) handleShiftTab() bool {
	if i.completion == nil || !i.completion.Active() {
		return false
	}

	item, ok := i.completion.CyclePrev()
	if !ok {
		return false
	}

	i.applyCompletion(item)
	return true
}

// applyCompletion replaces the current prefix in the buffer with the item's Insert text.
// This is used for live preview while cycling through candidates.
func (i *InputLine) applyCompletion(item CompletionItem) {
	// Find the word boundary at cursor.
	text := i.Text()
	cursor := i.cursor

	// Find start of current word.
	start := cursor - 1
	for start >= 0 && text[start] != ' ' {
		start--
	}
	start++

	// Replace from start to cursor with insert text.
	insertRunes := []rune(item.Insert)
	restRunes := i.buf[cursor:]

	i.buf = append(i.buf[:start], append(insertRunes, restRunes...)...)
	i.cursor = start + len(insertRunes)
}

// handleEscape dismisses the completion popup if active.
func (i *InputLine) handleEscape() bool {
	if i.completion != nil && i.completion.Active() {
		i.completion.Cancel()
		return true
	}
	return false
}

// isCompletionActive reports whether the completion popup is visible.
func (i *InputLine) isCompletionActive() bool {
	return i.completion != nil && i.completion.Active()
}

// --- History ---

// AddHistory appends a message to the input history.
// Duplicate consecutive entries are ignored.
// If MaxHistorySize is exceeded, the oldest entry is removed.
func (i *InputLine) AddHistory(text string) {
	if text == "" {
		return
	}
	if len(i.history) > 0 && i.history[len(i.history)-1] == text {
		return // skip duplicate consecutive
	}
	i.history = append(i.history, text)
	// Trim oldest entries if over the limit.
	if i.maxHistory > 0 && len(i.history) > i.maxHistory {
		i.history = i.history[len(i.history)-i.maxHistory:]
	}
}

// SetHistory replaces the entire input history.
func (i *InputLine) SetHistory(h []string) {
	i.history = make([]string, len(h))
	copy(i.history, h)
	i.historyIdx = -1
}

// History returns a copy of the input history.
func (i *InputLine) History() []string {
	result := make([]string, len(i.history))
	copy(result, i.history)
	return result
}

// navigateHistory moves through the input history by delta (-1 = older, +1 = newer).
// When entering history mode, the current input is saved as draft.
// When navigating past the newest entry, the draft is restored.
func (i *InputLine) navigateHistory(delta int) {
	n := len(i.history)
	if n == 0 {
		return
	}

	// Entering history mode — save current input as draft.
	if i.historyIdx == -1 {
		if delta < 0 {
			// Going to older — save draft.
			i.draft = make([]rune, len(i.buf))
			copy(i.draft, i.buf)
			i.historyIdx = n // point past the end
		} else {
			// Going newer from draft — nothing to do.
			return
		}
	}

	i.historyIdx += delta

	switch {
	case i.historyIdx < 0:
		// Clamped to oldest.
		i.historyIdx = 0
		i.loadHistoryEntry(0)

	case i.historyIdx >= n:
		// Past newest — restore draft.
		i.historyIdx = -1
		i.buf = make([]rune, len(i.draft))
		copy(i.buf, i.draft)
		i.cursor = len(i.buf)

	default:
		i.loadHistoryEntry(i.historyIdx)
	}
}

// loadHistoryEntry loads the history entry at idx into the buffer.
func (i *InputLine) loadHistoryEntry(idx int) {
	if idx < 0 || idx >= len(i.history) {
		return
	}
	i.buf = []rune(i.history[idx])
	i.cursor = len(i.buf) // cursor at end
}

// --- internal helpers ---

// insertRune inserts r at the cursor position and advances the cursor.
func (i *InputLine) insertRune(r rune) {
	i.buf = append(i.buf[:i.cursor], append([]rune{r}, i.buf[i.cursor:]...)...)
	i.cursor++
}

// deleteWordBack deletes the word before the cursor.
// A word is a run of non-space characters; leading spaces before the word
// are also removed.
func (i *InputLine) deleteWordBack() {
	pos := i.cursor
	// Skip trailing spaces.
	for pos > 0 && i.buf[pos-1] == ' ' {
		pos--
	}
	// Skip word characters.
	for pos > 0 && i.buf[pos-1] != ' ' {
		pos--
	}
	i.buf = append(i.buf[:pos], i.buf[i.cursor:]...)
	i.cursor = pos
}

// runeWidthBefore returns the total display width of buf[0:idx].
func (i *InputLine) runeWidthBefore(idx int) int {
	w := 0
	for j := 0; j < idx && j < len(i.buf); j++ {
		w += buffer.RuneWidth(i.buf[j])
	}
	return w
}
