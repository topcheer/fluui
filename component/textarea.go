package component

import (
	"strings"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// TabWidth is the number of spaces inserted when Tab is pressed.
const TabWidth = 4

// TextArea is a multi-line text editor component.
// It supports full keyboard editing: character insertion, line splitting,
// deletion (char/word/line), cursor movement across lines, and scrolling.
type TextArea struct {
	BaseComponent

	lines    [][]rune // text content, one slice per line
	cursorX  int      // cursor column (rune index within current line)
	cursorY  int      // cursor line index
	scrollY  int      // vertical scroll offset (top visible line)
	defStyle buffer.Style
}

// NewTextArea creates a TextArea with a single empty line and default style.
func NewTextArea() *TextArea {
	return &TextArea{
		lines:    [][]rune{{}},
		defStyle: buffer.Style{},
	}
}

// SetStyle sets the default text style.
func (ta *TextArea) SetStyle(s buffer.Style) {
	ta.defStyle = s
}

// Text returns the full content as a string with newlines between lines.
func (ta *TextArea) Text() string {
	if len(ta.lines) == 0 {
		return ""
	}
	parts := make([]string, len(ta.lines))
	for i, line := range ta.lines {
		parts[i] = string(line)
	}
	return strings.Join(parts, "\n")
}

// SetText replaces the entire content. The text is split on newlines.
func (ta *TextArea) SetText(s string) {
	if s == "" {
		ta.lines = [][]rune{{}}
		ta.cursorX = 0
		ta.cursorY = 0
		return
	}
	parts := strings.Split(s, "\n")
	ta.lines = make([][]rune, len(parts))
	for i, p := range parts {
		ta.lines[i] = []rune(p)
	}
	ta.cursorX = 0
	ta.cursorY = 0
	ta.scrollY = 0
}

// Clear removes all content and resets cursor.
func (ta *TextArea) Clear() {
	ta.lines = [][]rune{{}}
	ta.cursorX = 0
	ta.cursorY = 0
	ta.scrollY = 0
}

// CursorPos returns the current cursor position (x, y).
func (ta *TextArea) CursorPos() (int, int) {
	return ta.cursorX, ta.cursorY
}

// LineCount returns the number of lines.
func (ta *TextArea) LineCount() int {
	return len(ta.lines)
}

// InsertText inserts a string at the cursor position.
// Newlines in the string split the current line.
func (ta *TextArea) InsertText(s string) {
	parts := strings.Split(s, "\n")
	if len(parts) == 1 {
		ta.insertRunes([]rune(s))
		return
	}
	// Multi-line insert: split current line and insert each part
	for i, part := range parts {
		if i > 0 {
			ta.splitLine()
		}
		if len(part) > 0 {
			ta.insertRunes([]rune(part))
		}
	}
}

// insertRunes inserts runes at cursor in the current line (no newlines).
func (ta *TextArea) insertRunes(runes []rune) {
	line := ta.lines[ta.cursorY]
	newLine := make([]rune, 0, len(line)+len(runes))
	newLine = append(newLine, line[:ta.cursorX]...)
	newLine = append(newLine, runes...)
	newLine = append(newLine, line[ta.cursorX:]...)
	ta.lines[ta.cursorY] = newLine
	ta.cursorX += len(runes)
}

// splitLine breaks the current line at the cursor position.
func (ta *TextArea) splitLine() {
	line := ta.lines[ta.cursorY]
	before := make([]rune, ta.cursorX)
	copy(before, line[:ta.cursorX])
	after := make([]rune, len(line)-ta.cursorX)
	copy(after, line[ta.cursorX:])

	// Insert new line after current
	newLines := make([][]rune, 0, len(ta.lines)+1)
	newLines = append(newLines, ta.lines[:ta.cursorY]...)
	newLines = append(newLines, before, after)
	if ta.cursorY+1 < len(ta.lines) {
		newLines = append(newLines, ta.lines[ta.cursorY+1:]...)
	}
	ta.lines = newLines
	ta.cursorY++
	ta.cursorX = 0
	ta.ensureVisible()
}

// HandleKey processes a key event. Returns true if the key was consumed.
//
// Supported keys:
//   - Printable chars: insert at cursor
//   - Enter: split line at cursor
//   - Backspace: delete char before cursor, or join with previous line
//   - Delete: delete char at cursor, or join with next line
//   - Tab: insert spaces
//   - Arrow keys: Up/Down/Left/Right
//   - Home/End: beginning/end of line
//   - Ctrl+A/E: Home/End (emacs style)
//   - Ctrl+K: delete to end of line
//   - Ctrl+U: delete to beginning of line
//   - Ctrl+W: delete previous word
//   - PageUp/PageDown: scroll
//   - Alt+Up/Down: move line up/down
func (ta *TextArea) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	// --- Alt modifiers ---
	if key.Modifiers&term.ModAlt != 0 {
		switch key.Key {
		case term.KeyUp:
			ta.moveLine(-1)
			return true
		case term.KeyDown:
			ta.moveLine(1)
			return true
		}
		// Alt+Backspace = delete word back (same as Ctrl+W on some terminals)
	}

	// --- Ctrl shortcuts ---
	if key.Modifiers&term.ModCtrl != 0 && key.Rune != 0 {
		switch key.Rune {
		case 'a':
			ta.cursorX = 0
			return true
		case 'e':
			ta.cursorX = len(ta.lines[ta.cursorY])
			return true
		case 'k':
			ta.deleteToEndOfLine()
			return true
		case 'u':
			ta.deleteToStartOfLine()
			return true
		case 'w':
			ta.deleteWordBack()
			return true
		}
	}

	// --- Special keys ---
	switch key.Key {
	case term.KeyEnter:
		ta.splitLine()
		return true

	case term.KeyTab:
		spaces := make([]rune, TabWidth)
		for i := range spaces {
			spaces[i] = ' '
		}
		ta.insertRunes(spaces)
		return true

	case term.KeyBackspace:
		if ta.cursorX > 0 {
			line := ta.lines[ta.cursorY]
			ta.lines[ta.cursorY] = append(line[:ta.cursorX-1], line[ta.cursorX:]...)
			ta.cursorX--
		} else if ta.cursorY > 0 {
			// Join with previous line
			prev := ta.lines[ta.cursorY-1]
			curr := ta.lines[ta.cursorY]
			ta.cursorX = len(prev)
			ta.lines[ta.cursorY-1] = append(prev, curr...)
			// Remove current line
			ta.lines = append(ta.lines[:ta.cursorY], ta.lines[ta.cursorY+1:]...)
			ta.cursorY--
			ta.ensureVisible()
		}
		return true

	case term.KeyDelete:
		line := ta.lines[ta.cursorY]
		if ta.cursorX < len(line) {
			ta.lines[ta.cursorY] = append(line[:ta.cursorX], line[ta.cursorX+1:]...)
		} else if ta.cursorY < len(ta.lines)-1 {
			// Join with next line
			next := ta.lines[ta.cursorY+1]
			ta.lines[ta.cursorY] = append(line, next...)
			// Remove next line
			ta.lines = append(ta.lines[:ta.cursorY+1], ta.lines[ta.cursorY+2:]...)
		}
		return true

	case term.KeyLeft:
		if ta.cursorX > 0 {
			ta.cursorX--
		} else if ta.cursorY > 0 {
			ta.cursorY--
			ta.cursorX = len(ta.lines[ta.cursorY])
			ta.ensureVisible()
		}
		return true

	case term.KeyRight:
		if ta.cursorX < len(ta.lines[ta.cursorY]) {
			ta.cursorX++
		} else if ta.cursorY < len(ta.lines)-1 {
			ta.cursorY++
			ta.cursorX = 0
			ta.ensureVisible()
		}
		return true

	case term.KeyUp:
		if ta.cursorY > 0 {
			ta.cursorY--
			ta.clampCursorX()
		}
		ta.ensureVisible()
		return true

	case term.KeyDown:
		if ta.cursorY < len(ta.lines)-1 {
			ta.cursorY++
			ta.clampCursorX()
		}
		ta.ensureVisible()
		return true

	case term.KeyHome:
		ta.cursorX = 0
		return true

	case term.KeyEnd:
		ta.cursorX = len(ta.lines[ta.cursorY])
		return true

	case term.KeyPageUp:
		bounds := ta.Bounds()
		scrollAmt := bounds.H
		if scrollAmt <= 0 {
			scrollAmt = 10
		}
		ta.cursorY -= scrollAmt
		if ta.cursorY < 0 {
			ta.cursorY = 0
		}
		ta.clampCursorX()
		ta.ensureVisible()
		return true

	case term.KeyPageDown:
		bounds := ta.Bounds()
		scrollAmt := bounds.H
		if scrollAmt <= 0 {
			scrollAmt = 10
		}
		ta.cursorY += scrollAmt
		if ta.cursorY >= len(ta.lines) {
			ta.cursorY = len(ta.lines) - 1
		}
		ta.clampCursorX()
		ta.ensureVisible()
		return true

	case term.KeySpace:
		ta.insertRunes([]rune{' '})
		return true
	}

	// --- Printable character ---
	if key.Rune != 0 && key.Key == 0 {
		ta.insertRunes([]rune{key.Rune})
		return true
	}

	return false
}

// --- Editing helpers ---

func (ta *TextArea) clampCursorX() {
	if ta.cursorX > len(ta.lines[ta.cursorY]) {
		ta.cursorX = len(ta.lines[ta.cursorY])
	}
	if ta.cursorX < 0 {
		ta.cursorX = 0
	}
}

func (ta *TextArea) ensureVisible() {
	bounds := ta.Bounds()
	h := bounds.H
	if h <= 0 {
		return
	}
	if ta.cursorY < ta.scrollY {
		ta.scrollY = ta.cursorY
	}
	if ta.cursorY >= ta.scrollY+h {
		ta.scrollY = ta.cursorY - h + 1
	}
	if ta.scrollY < 0 {
		ta.scrollY = 0
	}
}

func (ta *TextArea) deleteToEndOfLine() {
	ta.lines[ta.cursorY] = ta.lines[ta.cursorY][:ta.cursorX]
}

func (ta *TextArea) deleteToStartOfLine() {
	line := ta.lines[ta.cursorY]
	ta.lines[ta.cursorY] = line[ta.cursorX:]
	ta.cursorX = 0
}

func (ta *TextArea) deleteWordBack() {
	line := ta.lines[ta.cursorY]
	if ta.cursorX == 0 {
		return
	}
	// Skip trailing spaces
	end := ta.cursorX
	for end > 0 && line[end-1] == ' ' {
		end--
	}
	// Delete word chars
	for end > 0 && line[end-1] != ' ' {
		end--
	}
	ta.lines[ta.cursorY] = append(line[:end], line[ta.cursorX:]...)
	ta.cursorX = end
}

func (ta *TextArea) moveLine(delta int) {
	if len(ta.lines) <= 1 {
		return
	}
	newY := ta.cursorY + delta
	if newY < 0 || newY >= len(ta.lines) {
		return
	}

	// Swap current and target line
	from, to := ta.cursorY, newY
	ta.lines[from], ta.lines[to] = ta.lines[to], ta.lines[from]
	ta.cursorY = newY
	ta.ensureVisible()
}

// --- Component interface ---

// Measure returns the desired size: width = longest line, height = line count.
func (ta *TextArea) Measure(cs Constraints) Size {
	maxW := 0
	for _, line := range ta.lines {
		w := buffer.StringWidth(string(line))
		if w > maxW {
			maxW = w
		}
	}
	h := len(ta.lines)
	if cs.MaxWidth > 0 && maxW > cs.MaxWidth {
		maxW = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: maxW, H: h}
}

// Paint renders the text content within bounds, with a reverse-video cursor.
func (ta *TextArea) Paint(buf *buffer.Buffer) {
	bounds := ta.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	// Draw visible lines
	for row := 0; row < bounds.H; row++ {
		lineIdx := ta.scrollY + row
		if lineIdx >= len(ta.lines) {
			break
		}
		line := ta.lines[lineIdx]
		x := bounds.X

		for _, r := range line {
			if x >= bounds.X+bounds.W {
				break
			}
			cell := buffer.Cell{
				Rune:  r,
				Width: uint8(buffer.RuneWidth(r)),
				Fg:    ta.defStyle.Fg,
				Bg:    ta.defStyle.Bg,
				Flags: ta.defStyle.Flags,
			}
			buf.SetCell(x, bounds.Y+row, cell)
			x += int(cell.Width)
		}

		// Draw cursor if on this line
		if lineIdx == ta.cursorY {
			curX := bounds.X
			for col := 0; col < ta.cursorX; col++ {
				if col < len(line) {
					curX += buffer.RuneWidth(line[col])
				} else {
					curX++
				}
			}
			if curX < bounds.X+bounds.W {
				var cur buffer.Cell
				if ta.cursorX < len(line) {
					r := line[ta.cursorX]
					cur = buffer.Cell{
						Rune:  r,
						Width: uint8(buffer.RuneWidth(r)),
						Fg:    ta.defStyle.Bg,
						Bg:    ta.defStyle.Fg,
						Flags: ta.defStyle.Flags | buffer.Reverse,
					}
				} else {
					cur = buffer.Cell{
						Rune:  ' ',
						Width: 1,
						Bg:    ta.defStyle.Fg,
						Flags: buffer.Reverse,
					}
				}
				buf.SetCell(curX, bounds.Y+row, cur)
			}
		}
	}
}
