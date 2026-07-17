package component

import (
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── TextField: bubbles.textinput-compatible standalone component ───
//
// This is a standalone text input component matching the bubbles.textinput API,
// with Focus/Blur, Blink, Placeholder, CharLimit, EchoPassword, Prompt, etc.
// It is designed for direct migration from bubbles.textinput.

// TextField is a standalone text input with cursor, placeholder, and echo modes.
type TextInput struct {
	mu sync.RWMutex

	// Content
	value      []rune
	cursor     int
	width      int
	maxWidth   int
	charLimit  int

	// Display
	prompt       string
	placeholder  string
	echoMode     EchoMode
	echoChar     rune
	focused      bool
	blinkEnabled bool
	blinking     bool
	lastBlink    time.Time

	// Style
	style       buffer.Style
	promptStyle buffer.Style
	placeholderStyle buffer.Style
	focusedStyle buffer.Style

	// History
	history    []string
	historyIdx int

	// Callbacks
	onChange func(string)
	onSubmit func(string)
}

// EchoMode controls how typed characters are displayed.
type EchoMode int

const (
	EchoNormal   EchoMode = iota // Display typed characters (default)
	EchoPassword                // Display '*' for each character
	EchoNone                    // Display nothing
)

// NewTextField creates a new TextField (bubbles.textinput.New compatible).
func NewTextInput() *TextInput {
	return &TextInput{
		value:        []rune{},
		cursor:       0,
		width:        20,
		maxWidth:     80,
		charLimit:    0, // 0 = no limit
		prompt:       "",
		placeholder:  "",
		echoMode:     EchoNormal,
		echoChar:     defaultEchoChar,
		focused:      false,
		blinkEnabled: true,
		style:        buffer.Style{Fg: buffer.NamedColor(buffer.NamedWhite)},
		promptStyle:  buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlue)},
		placeholderStyle: buffer.Style{Fg: buffer.NamedColor(buffer.NamedBrightBlack)},
		focusedStyle: buffer.Style{Fg: buffer.NamedColor(buffer.NamedWhite), Flags: buffer.Bold},
	}
}

// ─── Content API ───

// Value returns the current text content.
func (t *TextInput) Value() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return string(t.value)
}

// SetValue sets the text content and moves cursor to end.
func (t *TextInput) SetValue(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	runes := []rune(s)
	if t.charLimit > 0 && len(runes) > t.charLimit {
		runes = runes[:t.charLimit]
	}
	t.value = runes
	t.cursor = len(t.value)
	t.notifyChange()
}

// InsertText inserts text at cursor position.
func (t *TextInput) InsertText(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	runes := []rune(s)
	if t.charLimit > 0 && len(t.value)+len(runes) > t.charLimit {
		remaining := t.charLimit - len(t.value)
		if remaining <= 0 {
			return
		}
		runes = runes[:remaining]
	}
	t.value = append(t.value[:t.cursor], append(runes, t.value[t.cursor:]...)...)
	t.cursor += len(runes)
	t.notifyChange()
}

// Clear clears the text content.
func (t *TextInput) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.value = []rune{}
	t.cursor = 0
	t.notifyChange()
}

// Empty returns true if the text content is empty.
func (t *TextInput) Empty() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.value) == 0
}

// Len returns the number of runes in the text content.
func (t *TextInput) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.value)
}

// Cursor returns the cursor position (rune index).
func (t *TextInput) Cursor() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cursor
}

// SetCursor sets the cursor position, clamped to valid range.
func (t *TextInput) SetCursor(pos int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if pos < 0 {
		pos = 0
	}
	if pos > len(t.value) {
		pos = len(t.value)
	}
	t.cursor = pos
}

// CursorEnd moves cursor to end of text.
func (t *TextInput) CursorEnd() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursor = len(t.value)
}

// CursorStart moves cursor to start of text.
func (t *TextInput) CursorStart() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cursor = 0
}

// Position returns the cursor position as (line, col) — always (0, cursor) for single-line.
func (t *TextInput) Position() (int, int) {
	return 0, t.Cursor()
}

// ─── Display API ───

// Prompt returns the prompt string.
func (t *TextInput) Prompt() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.prompt
}

// SetPrompt sets the prompt string.
func (t *TextInput) SetPrompt(p string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.prompt = p
}

// Placeholder returns the placeholder string.
func (t *TextInput) Placeholder() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.placeholder
}

// SetPlaceholder sets the placeholder shown when empty.
func (t *TextInput) SetPlaceholder(p string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.placeholder = p
}

// EchoMode returns the current echo mode.
func (t *TextInput) EchoMode() EchoMode {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.echoMode
}

// SetEchoMode sets the echo mode (normal, password, none).
func (t *TextInput) SetEchoMode(m EchoMode) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.echoMode = m
}

// echoChar is the character shown for password echo (default '*').
var defaultEchoChar rune = '*'

// SetEchoChar sets the character displayed in password mode.
func (t *TextInput) SetEchoChar(r rune) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.echoChar = r
}

// EchoChar returns the character used for password echo.
func (t *TextInput) EchoChar() rune {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.echoChar
}

// EchoPassword sets echo mode to password (shorthand).
func (t *TextInput) EchoPassword() {
	t.SetEchoMode(EchoPassword)
}

// CharLimit returns the character limit (0 = no limit).
func (t *TextInput) CharLimit() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.charLimit
}

// SetCharLimit sets the max number of characters (0 = no limit).
func (t *TextInput) SetCharLimit(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.charLimit = n
	if n > 0 && len(t.value) > n {
		t.value = t.value[:n]
		if t.cursor > n {
			t.cursor = n
		}
	}
}

// Width returns the display width.
func (t *TextInput) Width() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.width
}

// SetWidth sets the display width.
func (t *TextInput) SetWidth(w int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if w > t.maxWidth {
		w = t.maxWidth
	}
	if w < 1 {
		w = 1
	}
	t.width = w
}

// Focused returns whether the text field is focused.
func (t *TextInput) Focused() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.focused
}

// Focus sets focus and starts blinking.
func (t *TextInput) Focus() {
	t.mu.Lock()
	t.focused = true
	t.lastBlink = time.Now()
	t.mu.Unlock()
}

// Blur removes focus and stops blinking.
func (t *TextInput) Blur() {
	t.mu.Lock()
	t.focused = false
	t.blinking = false
	t.mu.Unlock()
}

// Blink returns whether the cursor is currently blinking (visible state).
func (t *TextInput) Blink() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if !t.focused || !t.blinkEnabled {
		return false
	}
	// Toggle blink every 500ms
	elapsed := time.Since(t.lastBlink)
	return (elapsed / (500 * time.Millisecond)) % 2 == 0
}

// SetBlink enables/disables cursor blinking.
func (t *TextInput) SetBlink(b bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.blinkEnabled = b
}

// ─── Style API ───

// SetStyle sets the text style.
func (t *TextInput) SetStyle(s buffer.Style) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.style = s
}

// SetPromptStyle sets the prompt style.
func (t *TextInput) SetPromptStyle(s buffer.Style) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.promptStyle = s
}

// SetPlaceholderStyle sets the placeholder style.
func (t *TextInput) SetPlaceholderStyle(s buffer.Style) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.placeholderStyle = s
}

// SetFocusedStyle sets the style when focused.
func (t *TextInput) SetFocusedStyle(s buffer.Style) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.focusedStyle = s
}

// ─── Callbacks ───

// SetOnChange sets a callback called when text changes.
func (t *TextInput) SetOnChange(fn func(string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onChange = fn
}

// SetOnSubmit sets a callback called on Enter.
func (t *TextInput) SetOnSubmit(fn func(string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onSubmit = fn
}

// ─── History API ───

// SetHistory sets the command history.
func (t *TextInput) SetHistory(h []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.history = h
	t.historyIdx = len(h)
}

// History returns the command history.
func (t *TextInput) History() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.history
}

// AddHistory adds an entry to history.
func (t *TextInput) AddHistory(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.history = append(t.history, s)
	t.historyIdx = len(t.history)
}

// navigateHistory moves through history by delta.
func (t *TextInput) navigateHistory(delta int) {
	if len(t.history) == 0 {
		return
	}
	t.historyIdx += delta
	if t.historyIdx < 0 {
		t.historyIdx = 0
	}
	if t.historyIdx > len(t.history) {
		t.historyIdx = len(t.history)
	}
	if t.historyIdx < len(t.history) {
		t.value = []rune(t.history[t.historyIdx])
	} else {
		t.value = []rune{}
	}
	t.cursor = len(t.value)
}

// ─── Key Handling ───

// HandleKey processes a key event. Returns true if handled.
func (t *TextInput) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	// Printable character
	if key.Rune != 0 && key.Key == term.KeyUnknown {
		t.InsertText(string(key.Rune))
		return true
	}

	switch key.Key {
	case term.KeyBackspace:
		t.mu.Lock()
		if t.cursor > 0 {
			t.value = append(t.value[:t.cursor-1], t.value[t.cursor:]...)
			t.cursor--
			t.notifyChange()
		}
		t.mu.Unlock()
		return true

	case term.KeyLeft:
		if key.Modifiers&term.ModCtrl != 0 {
			// Ctrl+Left: move to start of previous word
			t.mu.Lock()
			for t.cursor > 0 && t.value[t.cursor-1] == ' ' {
				t.cursor--
			}
			for t.cursor > 0 && t.value[t.cursor-1] != ' ' {
				t.cursor--
			}
			t.mu.Unlock()
		} else {
			t.SetCursor(t.Cursor() - 1)
		}
		return true

	case term.KeyRight:
		if key.Modifiers&term.ModCtrl != 0 {
			// Ctrl+Right: move to start of next word
			t.mu.Lock()
			for t.cursor < len(t.value) && t.value[t.cursor] == ' ' {
				t.cursor++
			}
			for t.cursor < len(t.value) && t.value[t.cursor] != ' ' {
				t.cursor++
			}
			t.mu.Unlock()
		} else {
			t.SetCursor(t.Cursor() + 1)
		}
		return true

	case term.KeyHome:
		t.CursorStart()
		return true

	case term.KeyEnd:
		t.CursorEnd()
		return true

	case term.KeyEnter:
		t.mu.RLock()
		fn := t.onSubmit
		val := string(t.value)
		t.mu.RUnlock()
		if fn != nil {
			fn(val)
		}
		return true

	case term.KeyDelete:
		t.mu.Lock()
		if t.cursor < len(t.value) {
			t.value = append(t.value[:t.cursor], t.value[t.cursor+1:]...)
			t.notifyChange()
		}
		t.mu.Unlock()
		return true

	case term.KeyUp:
		t.navigateHistory(-1)
		return true

	case term.KeyDown:
		t.navigateHistory(1)
		return true
	}

	return false
}

// ─── Component Interface ───

// Measure returns the preferred size.
func (t *TextInput) Measure(cs Constraints) Size {
	t.mu.RLock()
	defer t.mu.RUnlock()
	w := t.width
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the text field into the buffer.
func (t *TextInput) Paint(buf *buffer.Buffer) {
	if buf == nil {
		return
	}
	t.mu.RLock()
	defer t.mu.RUnlock()

	x := 0
	y := 0

	// Draw prompt
	if t.prompt != "" {
		for _, r := range t.prompt {
			if x < buf.Width {
				buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: t.promptStyle.Fg, Bg: t.promptStyle.Bg, Flags: t.promptStyle.Flags})
				x++
			}
		}
	}

	// Draw content or placeholder
	if len(t.value) == 0 && t.placeholder != "" {
		for _, r := range []rune(t.placeholder) {
			if x < buf.Width {
				buf.SetCell(x, y, buffer.Cell{Rune: r, Width: 1, Fg: t.placeholderStyle.Fg, Bg: t.placeholderStyle.Bg, Flags: t.placeholderStyle.Flags})
				x++
			}
		}
	} else {
		style := t.style
		if t.focused {
			style = t.focusedStyle
		}
		for i, r := range t.value {
			if x >= buf.Width {
				break
			}
			displayRune := r
			if t.echoMode == EchoPassword {
				displayRune = '*'
			} else if t.echoMode == EchoNone {
				displayRune = ' '
			}
			// Cursor blink
			if t.focused && i == t.cursor && t.Blink() {
				buf.SetCell(x, y, buffer.Cell{Rune: displayRune, Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags | buffer.Reverse})
			} else {
				buf.SetCell(x, y, buffer.Cell{Rune: displayRune, Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
			}
			x++
		}
	}

	// Draw cursor at end if focused and cursor is at end
	if t.focused && t.cursor >= len(t.value) && t.Blink() && x < buf.Width {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Flags: buffer.Reverse})
	}
}

// SetBounds is a no-op for single-line text field (uses Measure size).
func (t *TextInput) SetBounds(r Rect) {
	// Use width from bounds if provided
	t.mu.Lock()
	if r.W > 0 {
		t.width = r.W
	}
	t.mu.Unlock()
}

// Children returns nil (no children).
func (t *TextInput) Children() []Component {
	return nil
}

// ─── Internal ───

func (t *TextInput) notifyChange() {
	if t.onChange != nil {
		t.onChange(string(t.value))
	}
}

// String returns the text content (Stringer for convenience).
func (t *TextInput) String() string {
	return t.Value()
}

// displayValue returns the visible representation (for testing).
func (t *TextInput) displayValue() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.echoMode == EchoPassword {
		return strings.Repeat("*", len(t.value))
	}
	if t.echoMode == EchoNone {
		return ""
	}
	return string(t.value)
}