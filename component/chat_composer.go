package component

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ChatComposer is a feature-rich chat input component that wraps TextArea
// with a prompt symbol, token count display, hint text, slash-command
// indicator, and auto-grow behavior.
//
// Features:
//   - Enter to submit, Shift+Enter for newline
//   - Auto-grow from 1 to maxRows lines
//   - Token count badge (input + output) in bottom-right
//   - Hint text below the input (e.g. "Enter to send, Shift+Enter for newline")
//   - Disabled state shows "Thinking..." with muted styling
//   - Placeholder when empty
//   - Slash command detection (shows indicator)
//   - Thread-safe
type ChatComposer struct {
	BaseComponent
	mu sync.Mutex

	textarea   *TextArea
	placeholder string
	hint        string
	onSubmit    func(string)
	maxRows     int
	disabled    bool
	tokenIn     int
	tokenOut    int
	slashMode   bool
}

// NewChatComposer creates a chat composer with default settings.
func NewChatComposer() *ChatComposer {
	return &ChatComposer{
		BaseComponent: BaseComponent{id: GenerateID("composer")},
		textarea:      NewTextArea(),
		placeholder:   "Type a message…",
		hint:          "Enter to send · Shift+Enter for newline",
		maxRows:       5,
	}
}

// Text returns the current input text.
func (c *ChatComposer) Text() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.textarea.Text()
}

// SetText replaces the input text.
func (c *ChatComposer) SetText(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.textarea.SetText(s)
	c.slashMode = strings.HasPrefix(s, "/")
}

// SetPlaceholder sets the placeholder shown when empty.
func (c *ChatComposer) SetPlaceholder(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.placeholder = s
}

// SetHint sets the hint text shown below the input.
func (c *ChatComposer) SetHint(s string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hint = s
}

// SetDisabled toggles the disabled (Thinking...) state.
func (c *ChatComposer) SetDisabled(v bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled = v
}

// IsDisabled returns whether the composer is disabled.
func (c *ChatComposer) IsDisabled() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.disabled
}

// SetOnSubmit sets the callback invoked when the user presses Enter.
func (c *ChatComposer) SetOnSubmit(fn func(string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSubmit = fn
}

// SetMaxRows sets the maximum visible rows before the input starts scrolling.
func (c *ChatComposer) SetMaxRows(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 1 {
		n = 1
	}
	c.maxRows = n
}

// SetTokenCount sets the displayed token counts.
func (c *ChatComposer) SetTokenCount(input, output int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenIn = input
	c.tokenOut = output
}

// Clear empties the input and resets the slash indicator.
func (c *ChatComposer) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.textarea.Clear()
	c.slashMode = false
}

// HandleKey processes keyboard input.
// Enter submits (if not disabled and not Shift+Enter).
// All other keys are forwarded to the wrapped TextArea.
func (c *ChatComposer) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	c.mu.Lock()
	disabled := c.disabled
	c.mu.Unlock()

	if disabled {
		return false
	}

	// Enter = submit (unless Shift is held)
	if ev.Key == term.KeyEnter {
		if ev.Modifiers&term.ModShift != 0 {
			// Shift+Enter → insert newline
			c.mu.Lock()
			c.textarea.HandleKey(ev)
			c.mu.Unlock()
			return true
		}
		// Plain Enter → submit
		c.mu.Lock()
		text := c.textarea.Text()
		c.textarea.Clear()
		c.slashMode = false
		fn := c.onSubmit
		c.mu.Unlock()
		if fn != nil {
			fn(text)
		}
		return true
	}

	// Forward to TextArea for all other keys
	c.mu.Lock()
	handled := c.textarea.HandleKey(ev)
	// Update slash mode
	c.slashMode = strings.HasPrefix(c.textarea.Text(), "/")
	c.mu.Unlock()
	return handled
}

// Measure returns the desired size, auto-growing with content.
func (c *ChatComposer) Measure(cs Constraints) Size {
	c.mu.Lock()
	defer c.mu.Unlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	h := c.measureHeightLocked()
	return Size{W: maxW, H: h}
}

// measureHeightLocked computes the height (caller must hold lock).
func (c *ChatComposer) measureHeightLocked() int {
	// Hint line (1) + input area
	h := 1

	if c.disabled {
		h += 1 // "Thinking..." line
		return h
	}

	lineCount := c.textarea.LineCount()
	if lineCount < 1 {
		lineCount = 1
	}
	if lineCount > c.maxRows {
		lineCount = c.maxRows
	}
	// +2 for top/bottom border
	h += lineCount + 2
	return h
}

// Paint renders the chat composer.
func (c *ChatComposer) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	if c.disabled {
		c.paintDisabled(buf, bounds)
		return
	}

	c.paintActive(buf, bounds)
}

// paintActive renders the input box with border, prompt, and content.
func (c *ChatComposer) paintActive(buf *buffer.Buffer, bounds Rect) {
	th := theme.Get()
	w := bounds.W
	inputH := bounds.H - 1 // minus hint line
	if inputH < 1 {
		inputH = 1
	}

	borderStyle := buffer.Style{Fg: th.Border}
	mutedStyle := buffer.Style{Fg: th.Muted}
	accentStyle := buffer.Style{Fg: th.Accent}

	// Draw top border
	for i := 0; i < w; i++ {
		buf.DrawText(bounds.X+i, bounds.Y, "─", borderStyle)
	}

	// Draw content lines
	contentW := w - 4 // "❯ " prefix + 1 right padding + scrollbar space
	if contentW < 1 {
		contentW = 1
	}

	lines := c.textarea.lines
	scrollY := 0
	if len(lines) > inputH {
		scrollY = len(lines) - inputH
	}

	for row := 0; row < inputH; row++ {
		y := bounds.Y + 1 + row
		if y >= bounds.Y+bounds.H {
			break
		}

		// Prompt on first line
		if row == 0 {
			buf.DrawText(bounds.X, y, "❯ ", accentStyle)
		} else {
			buf.DrawText(bounds.X, y, "  ", mutedStyle)
		}

		lineIdx := scrollY + row
		if lineIdx < len(lines) {
			line := string(lines[lineIdx])
			if utf8.RuneCountInString(line) > contentW {
				line = truncateStr(line, contentW)
			}
			buf.DrawText(bounds.X+2, y, line, buffer.Style{Fg: th.Fg})
		} else if row == 0 && c.textarea.Text() == "" && c.placeholder != "" {
			// Placeholder
			ph := c.placeholder
			if utf8.RuneCountInString(ph) > contentW {
				ph = truncateStr(ph, contentW)
			}
			buf.DrawText(bounds.X+2, y, ph, mutedStyle)
		}
	}

	// Bottom border
	botY := bounds.Y + inputH
	if botY < bounds.Y+bounds.H {
		for i := 0; i < w; i++ {
			buf.DrawText(bounds.X+i, botY, "─", borderStyle)
		}
	}

	// Token count (bottom-right of input area)
	if c.tokenIn > 0 || c.tokenOut > 0 {
		tokenStr := "↑" + formatTokenCount(c.tokenIn) + " ↓" + formatTokenCount(c.tokenOut)
		if utf8.RuneCountInString(tokenStr) < w-2 {
			buf.DrawText(bounds.X+w-utf8.RuneCountInString(tokenStr)-1, botY, tokenStr, mutedStyle)
		}
	}

	// Slash command indicator
	if c.slashMode {
		buf.DrawText(bounds.X, botY, "/", accentStyle)
	}

	// Hint line
	hintY := bounds.Y + bounds.H - 1
	if hintY > botY {
		hint := c.hint
		if utf8.RuneCountInString(hint) > w {
			hint = truncateStr(hint, w)
		}
		buf.DrawText(bounds.X, hintY, hint, mutedStyle)
	}
}

// paintDisabled renders the "Thinking..." state.
func (c *ChatComposer) paintDisabled(buf *buffer.Buffer, bounds Rect) {
	th := theme.Get()
	mutedStyle := buffer.Style{Fg: th.Muted}
	w := bounds.W

	// Top border
	for i := 0; i < w; i++ {
		buf.DrawText(bounds.X+i, bounds.Y, "─", buffer.Style{Fg: th.Border})
	}

	// "Thinking..." line
	buf.DrawText(bounds.X, bounds.Y+1, "⠋ Thinking…", mutedStyle)

	// Hint line
	hintY := bounds.Y + bounds.H - 1
	if hintY > bounds.Y+1 {
		buf.DrawText(bounds.X, hintY, c.hint, mutedStyle)
	}
}

// Placeholder returns the current placeholder text.
func (c *ChatComposer) Placeholder() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.placeholder
}

// Hint returns the current hint text.
func (c *ChatComposer) Hint() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.hint
}

// MaxRows returns the maximum visible rows.
func (c *ChatComposer) MaxRows() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.maxRows
}

// SlashMode returns whether the current input starts with "/".
func (c *ChatComposer) SlashMode() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.slashMode
}
