package component

import (
	"strconv"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// SGRColor holds foreground and background color state.
type SGRColor struct {
	Fg buffer.Color
	Bg buffer.Color
}

// TerminalLine represents a single line of terminal output with style information.
type TerminalLine struct {
	Cells []buffer.Cell
}

// TerminalPanel embeds a scrollable terminal view with ANSI parsing.
type TerminalPanel struct {
	BaseComponent

	lines       []*TerminalLine
	maxLines    int
	scrollOff   int
	scrollMode  bool
	autoScroll  bool

	// ANSI parser state
	curFg buffer.Color
	curBg buffer.Color
	curFlags buffer.StyleFlags

	// Input buffer for shell mode
	inputBuf  []rune
	cursorPos int

	// Callbacks
	OnInput func(s string)
	OnKey   func(k *term.KeyEvent) bool

	mu sync.RWMutex
}

// NewTerminalPanel creates a new terminal panel with a max line buffer.
func NewTerminalPanel(maxLines int) *TerminalPanel {
	if maxLines <= 0 {
		maxLines = 1000
	}
	return &TerminalPanel{
		lines:      make([]*TerminalLine, 0, maxLines),
		maxLines:   maxLines,
		autoScroll: true,
	}
}

// Write parses ANSI escape sequences from raw bytes and appends styled cells.
func (tp *TerminalPanel) Write(data []byte) (int, error) {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	tp.writeLocked(data)
	return len(data), nil
}

// WriteString is a convenience wrapper for Write.
func (tp *TerminalPanel) WriteString(s string) {
	tp.Write([]byte(s))
}

func (tp *TerminalPanel) writeLocked(data []byte) {
	i := 0
	for i < len(data) {
		if data[i] == 0x1b { // ESC
			// Parse escape sequence
			consumed := tp.parseEscapeLocked(data[i:])
			i += consumed
		} else {
			tp.appendByteLocked(data[i])
			i++
		}
	}
	tp.trimLocked()
	if tp.autoScroll {
		tp.scrollOff = 0
	}
}

func (tp *TerminalPanel) appendByteLocked(b byte) {
	if b == '\n' || b == '\r' {
		// Start new line
		tp.lines = append(tp.lines, &TerminalLine{Cells: []buffer.Cell{}})
		return
	}
	if b == '\t' {
		// Expand tab to 4 spaces
		for i := 0; i < 4; i++ {
			tp.appendByteLocked(' ')
		}
		return
	}
	if b < 0x20 {
		return // ignore other control chars
	}

	// Append to current line
	var line *TerminalLine
	if len(tp.lines) == 0 {
		tp.lines = append(tp.lines, &TerminalLine{Cells: []buffer.Cell{}})
	}
	line = tp.lines[len(tp.lines)-1]

	cell := buffer.Cell{
		Rune:  rune(b),
		Width: 1,
		Fg:    tp.curFg,
		Bg:    tp.curBg,
		Flags: tp.curFlags,
	}
	line.Cells = append(line.Cells, cell)
}

// parseEscapeLocked parses ANSI/SGR escape sequences, returns bytes consumed.
func (tp *TerminalPanel) parseEscapeLocked(data []byte) int {
	if len(data) < 2 {
		return len(data)
	}

	// CSI sequence: ESC [ ... <final>
	if data[1] == '[' {
		return tp.parseCSILocked(data)
	}

	// OSC sequence: ESC ] ... BEL or ESC \
	if data[1] == ']' {
		return tp.parseOSCLocked(data)
	}

	// Unknown escape, consume 2 bytes
	return 2
}

// parseCSILocked parses CSI (Control Sequence Introducer) sequences.
func (tp *TerminalPanel) parseCSILocked(data []byte) int {
	i := 2 // skip ESC [
	for i < len(data) {
		c := data[i]
		if c >= 0x40 && c <= 0x7e {
			// Final byte
			if c == 'm' {
				tp.parseSGRLocked(data[2:i])
			}
			return i + 1
		}
		i++
	}
	return len(data)
}

// parseSGRLocked parses SGR (Select Graphic Rendition) parameters.
func (tp *TerminalPanel) parseSGRLocked(params []byte) {
	if len(params) == 0 {
		// Reset
		tp.curFg = buffer.Color{}
		tp.curBg = buffer.Color{}
		tp.curFlags = 0
		return
	}

	parts := strings.Split(string(params), ";")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			nums = append(nums, 0)
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			nums = append(nums, 0)
		} else {
			nums = append(nums, n)
		}
	}

	i := 0
	for i < len(nums) {
		n := nums[i]
		switch {
		case n == 0:
			tp.curFg = buffer.Color{}
			tp.curBg = buffer.Color{}
			tp.curFlags = 0
		case n == 1:
			tp.curFlags |= buffer.Bold
		case n == 2:
			tp.curFlags |= buffer.Dim
		case n == 3:
			tp.curFlags |= buffer.Italic
		case n == 4:
			tp.curFlags |= buffer.Underline
		case n == 5:
			tp.curFlags |= buffer.Blink
		case n == 7:
			tp.curFlags |= buffer.Reverse
		case n == 9:
			tp.curFlags |= buffer.Strikethrough
		case n == 22:
			tp.curFlags &^= buffer.Bold
			tp.curFlags &^= buffer.Dim
		case n == 23:
			tp.curFlags &^= buffer.Italic
		case n == 24:
			tp.curFlags &^= buffer.Underline
		case n == 25:
			tp.curFlags &^= buffer.Blink
		case n == 27:
			tp.curFlags &^= buffer.Reverse
		case n == 29:
			tp.curFlags &^= buffer.Strikethrough
		case n >= 30 && n <= 37:
			tp.curFg = buffer.NamedColor(n - 30)
		case n == 38:
			// Extended foreground color
			if i+1 < len(nums) {
				consumed, c := tp.parseExtendedColorLocked(nums[i+1:])
				tp.curFg = c
				i += consumed
			}
		case n == 39:
			tp.curFg = buffer.Color{}
		case n >= 40 && n <= 47:
			tp.curBg = buffer.NamedColor(n - 40)
		case n == 48:
			// Extended background color
			if i+1 < len(nums) {
				consumed, c := tp.parseExtendedColorLocked(nums[i+1:])
				tp.curBg = c
				i += consumed
			}
		case n == 49:
			tp.curBg = buffer.Color{}
		case n >= 90 && n <= 97:
			tp.curFg = buffer.NamedColor(n - 90 + 8)
		case n >= 100 && n <= 107:
			tp.curBg = buffer.NamedColor(n - 100 + 8)
		}
		i++
	}
}

// parseExtendedColorLocked parses 38;5;n (256-color) or 38;2;r;g;b (truecolor).
func (tp *TerminalPanel) parseExtendedColorLocked(nums []int) (int, buffer.Color) {
	if len(nums) == 0 {
		return 0, buffer.Color{}
	}
	switch nums[0] {
	case 5: // 256-color
		if len(nums) >= 2 {
			return 2, buffer.Color256Val(uint8(nums[1]))
		}
		return 1, buffer.Color{}
	case 2: // truecolor
		if len(nums) >= 4 {
			return 4, buffer.RGB(uint8(nums[1]), uint8(nums[2]), uint8(nums[3]))
		}
		return 1, buffer.Color{}
	default:
		return 1, buffer.Color{}
	}
}

// parseOSCLocked skips OSC sequences (title setting, etc).
func (tp *TerminalPanel) parseOSCLocked(data []byte) int {
	i := 2
	for i < len(data) {
		if data[i] == 0x07 { // BEL
			return i + 1
		}
		if data[i] == 0x1b && i+1 < len(data) && data[i+1] == '\\' { // ESC \
			return i + 2
		}
		i++
	}
	return len(data)
}

func (tp *TerminalPanel) trimLocked() {
	if len(tp.lines) > tp.maxLines {
		excess := len(tp.lines) - tp.maxLines
		tp.lines = tp.lines[excess:]
	}
}

// LineCount returns the number of stored lines.
func (tp *TerminalPanel) LineCount() int {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return len(tp.lines)
}

// ScrollUp scrolls up by n lines.
func (tp *TerminalPanel) ScrollUp(n int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.scrollOff += n
	maxOff := len(tp.lines) - 1
	if maxOff < 0 {
		maxOff = 0
	}
	if tp.scrollOff > maxOff {
		tp.scrollOff = maxOff
	}
	tp.autoScroll = false
}

// ScrollDown scrolls down by n lines.
func (tp *TerminalPanel) ScrollDown(n int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.scrollOff -= n
	if tp.scrollOff < 0 {
		tp.scrollOff = 0
	}
	if tp.scrollOff == 0 {
		tp.autoScroll = true
	}
}

// ScrollToBottom enables auto-scroll mode.
func (tp *TerminalPanel) ScrollToBottom() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.scrollOff = 0
	tp.autoScroll = true
}

// ScrollOffset returns the current scroll offset.
func (tp *TerminalPanel) ScrollOffset() int {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	return tp.scrollOff
}

// SetMaxLines sets the maximum line buffer size.
func (tp *TerminalPanel) SetMaxLines(n int) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	if n <= 0 {
		n = 1000
	}
	tp.maxLines = n
	tp.trimLocked()
}

// Clear removes all lines.
func (tp *TerminalPanel) Clear() {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.lines = tp.lines[:0]
	tp.scrollOff = 0
}

// Lines returns a defensive copy of the terminal lines.
func (tp *TerminalPanel) Lines() []*TerminalLine {
	tp.mu.RLock()
	defer tp.mu.RUnlock()
	out := make([]*TerminalLine, len(tp.lines))
	copy(out, tp.lines)
	return out
}

// SetOnInput sets the callback invoked when Enter is pressed in shell mode.
func (tp *TerminalPanel) SetOnInput(fn func(string)) {
	tp.mu.Lock()
	defer tp.mu.Unlock()
	tp.OnInput = fn
}

// Measure returns the bounds dimensions.
func (tp *TerminalPanel) Measure(cs Constraints) Size {
	w, h := 80, 24
	if cs.MaxWidth > 0 && cs.MaxWidth < w {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && cs.MaxHeight < h {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the visible lines into the buffer.
func (tp *TerminalPanel) Paint(buf *buffer.Buffer) {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	if buf == nil {
		return
	}

	bounds := tp.Bounds()
	offX, offY := bounds.X, bounds.Y
	maxW := bounds.W
	maxH := bounds.H

	if maxW <= 0 || maxH <= 0 {
		return
	}

	// Calculate visible line range
	end := len(tp.lines) - tp.scrollOff
	if end < 0 {
		end = 0
	}
	start := end - maxH
	if start < 0 {
		start = 0
	}

	// Paint lines
	row := 0
	for i := start; i < end && row < maxH; i++ {
		if i < 0 || i >= len(tp.lines) {
			continue
		}
		line := tp.lines[i]
		col := 0
		for _, cell := range line.Cells {
			if col >= maxW {
				break
			}
			x := offX + col
			y := offY + row
			if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
				buf.SetCell(x, y, cell)
			}
			col += int(cell.Width)
			if cell.Width == 0 {
				col++
			}
		}
		row++
	}

	// Paint input line if scroll mode is off (at bottom)
	if tp.scrollOff == 0 {
		y := offY + row
		if y < offY+maxH && y < buf.Height {
			x := offX
			for _, r := range string(tp.inputBuf) {
				if x >= offX+maxW || x >= buf.Width {
					break
				}
				buf.SetCell(x, y, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    tp.curFg,
					Bg:    tp.curBg,
					Flags: tp.curFlags,
				})
				x++
			}
			// Cursor
			cx := offX + tp.cursorPos
			if cx >= 0 && cx < buf.Width && cx < offX+maxW {
				c := buf.GetCell(cx, y)
				c.Flags |= buffer.Reverse
				buf.SetCell(cx, y, c)
			}
		}
	}
}

// HandleKey processes keyboard input for terminal navigation and text input.
func (tp *TerminalPanel) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}

	tp.mu.Lock()

	// Scroll mode keys
	switch k.Key {
	case term.KeyUp:
		tp.scrollOff++
		maxOff := len(tp.lines)
		if tp.scrollOff > maxOff {
			tp.scrollOff = maxOff
		}
		tp.autoScroll = false
		tp.mu.Unlock()
		return true
	case term.KeyDown:
		tp.scrollOff--
		if tp.scrollOff < 0 {
			tp.scrollOff = 0
		}
		if tp.scrollOff == 0 {
			tp.autoScroll = true
		}
		tp.mu.Unlock()
		return true
	case term.KeyPageUp:
		tp.scrollOff += 10
		maxOff := len(tp.lines)
		if tp.scrollOff > maxOff {
			tp.scrollOff = maxOff
		}
		tp.autoScroll = false
		tp.mu.Unlock()
		return true
	case term.KeyPageDown:
		tp.scrollOff -= 10
		if tp.scrollOff < 0 {
			tp.scrollOff = 0
		}
		if tp.scrollOff == 0 {
			tp.autoScroll = true
		}
		tp.mu.Unlock()
		return true
	case term.KeyHome:
		tp.scrollOff = len(tp.lines)
		tp.autoScroll = false
		tp.mu.Unlock()
		return true
	case term.KeyEnd:
		tp.scrollOff = 0
		tp.autoScroll = true
		tp.mu.Unlock()
		return true
	}

	// Text input (when at bottom / auto-scroll mode)
	if tp.scrollOff == 0 {
		if k.Key == term.KeyEnter {
			input := string(tp.inputBuf)
			tp.inputBuf = tp.inputBuf[:0]
			tp.cursorPos = 0
			cb := tp.OnInput
			tp.mu.Unlock()
			if cb != nil {
				cb(input)
			}
			return true
		}

		if k.Key == term.KeyBackspace {
			if tp.cursorPos > 0 && tp.cursorPos <= len(tp.inputBuf) {
				tp.inputBuf = append(tp.inputBuf[:tp.cursorPos-1], tp.inputBuf[tp.cursorPos:]...)
				tp.cursorPos--
			}
			tp.mu.Unlock()
			return true
		}

		if k.Key == term.KeyLeft {
			if tp.cursorPos > 0 {
				tp.cursorPos--
			}
			tp.mu.Unlock()
			return true
		}

		if k.Key == term.KeyRight {
			if tp.cursorPos < len(tp.inputBuf) {
				tp.cursorPos++
			}
			tp.mu.Unlock()
			return true
		}

		// Printable character
		if k.Rune >= 0x20 {
			tp.inputBuf = append(tp.inputBuf[:tp.cursorPos], append([]rune{k.Rune}, tp.inputBuf[tp.cursorPos:]...)...)
			tp.cursorPos++
			tp.mu.Unlock()
			return true
		}
	}

	cb := tp.OnKey
	tp.mu.Unlock()
	if cb != nil {
		return cb(k)
	}
	return false
}

// Children returns nil.
func (tp *TerminalPanel) Children() []Component {
	return nil
}
