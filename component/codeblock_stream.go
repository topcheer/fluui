package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CodeBlockStream: Live Streaming Code Block ───
//
// CodeBlockStream renders code text as it streams in token-by-token from
// an AI model. It shows a typing cursor at the streaming position and
// supports basic syntax highlighting via simple token coloring.
//
// Usage:
//
//	cbs := NewCodeBlockStream("go")
//	cbs.Start()
//	cbs.Append("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}")
//	cbs.Complete()
//	cbs.Paint(buf)

// CodeBlockStreamStyle holds visual styles.
type CodeBlockStreamStyle struct {
	Keyword  buffer.Style // func, var, if, for, return, etc
	String   buffer.Style // "quoted strings"
	Comment  buffer.Style // // comments
	Number   buffer.Style // 123, 3.14
	Default  buffer.Style
	Cursor   buffer.Style
	Border   buffer.Style
	LineNum  buffer.Style
}

// DefaultCodeBlockStreamStyle returns sensible defaults.
func DefaultCodeBlockStreamStyle() CodeBlockStreamStyle {
	return CodeBlockStreamStyle{
		Keyword: buffer.Style{Fg: buffer.RGB(86, 156, 214)},  // VS Code blue
		String:  buffer.Style{Fg: buffer.RGB(206, 145, 120)},  // orange
		Comment: buffer.Style{Fg: buffer.RGB(106, 153, 85)},   // green
		Number:  buffer.Style{Fg: buffer.RGB(181, 206, 168)},  // light green
		Default: buffer.Style{Fg: buffer.RGB(212, 212, 212)},  // light gray
		Cursor:  buffer.Style{Fg: buffer.RGB(100, 149, 237), Flags: buffer.Bold},
		Border:  buffer.Style{Fg: buffer.RGB(60, 60, 60)},
		LineNum: buffer.Style{Fg: buffer.RGB(90, 90, 90)},
	}
}

// CodeBlockStream renders streaming code with basic syntax highlighting.
type CodeBlockStream struct {
	BaseComponent
	mu          sync.RWMutex
	language    string
	code        string
	streaming   bool
	cursorOn    bool
	completed   bool
	showLineNum bool
	style       CodeBlockStreamStyle
}

// NewCodeBlockStream creates a streaming code block.
func NewCodeBlockStream(language string) *CodeBlockStream {
	cbs := &CodeBlockStream{
		language:    language,
		streaming:   false,
		showLineNum: true,
		style:       DefaultCodeBlockStreamStyle(),
	}
	cbs.SetID(GenerateID("codestream"))
	return cbs
}

// Start begins streaming.
func (cbs *CodeBlockStream) Start() *CodeBlockStream {
	cbs.mu.Lock()
	cbs.streaming = true
	cbs.completed = false
	cbs.code = ""
	cbs.mu.Unlock()
	return cbs
}

// Append adds code text.
func (cbs *CodeBlockStream) Append(text string) *CodeBlockStream {
	cbs.mu.Lock()
	cbs.code += text
	cbs.mu.Unlock()
	return cbs
}

// SetCode replaces all code.
func (cbs *CodeBlockStream) SetCode(code string) *CodeBlockStream {
	cbs.mu.Lock()
	cbs.code = code
	cbs.mu.Unlock()
	return cbs
}

// Code returns the current code text.
func (cbs *CodeBlockStream) Code() string {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	return cbs.code
}

// Complete marks streaming as done.
func (cbs *CodeBlockStream) Complete() *CodeBlockStream {
	cbs.mu.Lock()
	cbs.streaming = false
	cbs.completed = true
	cbs.mu.Unlock()
	return cbs
}

// IsStreaming returns true if currently streaming.
func (cbs *CodeBlockStream) IsStreaming() bool {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	return cbs.streaming
}

// IsCompleted returns true if streaming is done.
func (cbs *CodeBlockStream) IsCompleted() bool {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	return cbs.completed
}

// Language returns the language identifier.
func (cbs *CodeBlockStream) Language() string {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	return cbs.language
}

// SetLanguage sets the language.
func (cbs *CodeBlockStream) SetLanguage(lang string) *CodeBlockStream {
	cbs.mu.Lock()
	cbs.language = lang
	cbs.mu.Unlock()
	return cbs
}

// SetCursor toggles cursor blink state.
func (cbs *CodeBlockStream) SetCursor(on bool) *CodeBlockStream {
	cbs.mu.Lock()
	cbs.cursorOn = on
	cbs.mu.Unlock()
	return cbs
}

// SetShowLineNumbers toggles line number display.
func (cbs *CodeBlockStream) SetShowLineNumbers(show bool) *CodeBlockStream {
	cbs.mu.Lock()
	cbs.showLineNum = show
	cbs.mu.Unlock()
	return cbs
}

// ShowLineNumbers returns whether line numbers are shown.
func (cbs *CodeBlockStream) ShowLineNumbers() bool {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	return cbs.showLineNum
}

// LineCount returns the number of lines.
func (cbs *CodeBlockStream) LineCount() int {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	return countLinesInStr(cbs.code)
}

// SetStyle sets the visual style.
func (cbs *CodeBlockStream) SetStyle(s CodeBlockStreamStyle) *CodeBlockStream {
	cbs.mu.Lock()
	cbs.style = s
	cbs.mu.Unlock()
	return cbs
}

// Measure computes the desired size.
func (cbs *CodeBlockStream) Measure(cs Constraints) Size {
	cbs.mu.RLock()
	defer cbs.mu.RUnlock()
	w := 60
	h := countLinesInStr(cbs.code) + 2 // border + lines
	if h < 3 {
		h = 3
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// countLinesInStr counts newlines + 1.
func countLinesInStr(s string) int {
	if s == "" {
		return 1
	}
	n := 1
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			n++
		}
	}
	return n
}

// simpleTokenStyle returns the style for a character position based on
// simple keyword/string/comment detection (zero-alloc, best-effort).
func (cbs *CodeBlockStream) charStyle(line string, col int, style CodeBlockStreamStyle) buffer.Style {
	// Check if inside string literal
	inString := false
	inComment := false
	for i := 0; i < col && i < len(line); i++ {
		if inComment {
			continue
		}
		if line[i] == '/' && i+1 < len(line) && line[i+1] == '/' {
			inComment = true
			continue
		}
		if line[i] == '"' {
			inString = !inString
		}
	}

	if inComment {
		return style.Comment
	}
	if inString {
		return style.String
	}
	if col < len(line) {
		ch := line[col]
		if ch >= '0' && ch <= '9' {
			return style.Number
		}
	}
	return style.Default
}

// Paint renders the streaming code block.
func (cbs *CodeBlockStream) Paint(buf *buffer.Buffer) {
	cbs.mu.Lock()
	defer cbs.mu.Unlock()

	b := cbs.bounds
	if b.W < 4 || b.H < 2 {
		return
	}

	codeW := b.W
	lineNumW := 0
	if cbs.showLineNum {
		lineNumW = 4 // "  1 │ "
		codeW -= lineNumW
	}
	if codeW < 1 {
		codeW = 1
	}

	// Top border with language label
	for x := 0; x < b.W; x++ {
		buf.SetCell(b.X+x, b.Y, buffer.Cell{Rune: '─', Fg: cbs.style.Border.Fg, Bg: cbs.style.Border.Bg, Width: 1})
	}
	// Language label on top-right
	if cbs.language != "" && len(cbs.language) < b.W-4 {
		labelX := b.X + b.W - len(cbs.language) - 2
		for i, r := range cbs.language {
			buf.SetCell(labelX+i, b.Y, buffer.Cell{Rune: r, Fg: cbs.style.Border.Fg, Bg: cbs.style.Border.Bg, Width: 1})
		}
	}

	// Render code lines
	lines := splitCodeLines(cbs.code)
	displayLines := b.H - 1 // minus top border
	startLine := 0
	if len(lines) > displayLines {
		startLine = len(lines) - displayLines // show last N lines (auto-scroll)
	}

	for li := startLine; li < len(lines); li++ {
		row := b.Y + 1 + (li - startLine)
		if row >= b.Y+b.H {
			break
		}

		line := lines[li]
		x := b.X

		// Line number
		if cbs.showLineNum {
			lnStr := codeItoa(li + 1)
			// Right-align in lineNumW-1 chars
			padX := lineNumW - 1 - len(lnStr)
			if padX < 0 {
				padX = 0
			}
			for i := 0; i < padX; i++ {
				buf.SetCell(x, row, buffer.Cell{Rune: ' ', Fg: cbs.style.LineNum.Fg, Bg: cbs.style.LineNum.Bg, Width: 1})
				x++
			}
			for _, r := range lnStr {
				if x >= b.X+lineNumW-1 {
					break
				}
				buf.SetCell(x, row, buffer.Cell{Rune: r, Fg: cbs.style.LineNum.Fg, Bg: cbs.style.LineNum.Bg, Width: 1})
				x++
			}
			// Separator
			buf.SetCell(x, row, buffer.Cell{Rune: '│', Fg: cbs.style.Border.Fg, Bg: cbs.style.Border.Bg, Width: 1})
			x++
		}

		// Code text
		for col, r := range line {
			if x >= b.X+b.W {
				break
			}
			charStyle := cbs.charStyle(line, col, cbs.style)
			buf.SetCell(x, row, buffer.Cell{Rune: r, Fg: charStyle.Fg, Bg: charStyle.Bg, Flags: charStyle.Flags, Width: 1})
			x++
		}

		// Streaming cursor at end of last line
		if cbs.streaming && cbs.cursorOn && li == len(lines)-1 && x < b.X+b.W {
			buf.SetCell(x, row, buffer.Cell{Rune: '▋', Fg: cbs.style.Cursor.Fg, Bg: cbs.style.Cursor.Bg, Flags: cbs.style.Cursor.Flags, Width: 1})
		}
	}
}

// splitCodeLines splits code into lines.
func splitCodeLines(code string) []string {
	if code == "" {
		return []string{""}
	}
	var lines []string
	start := 0
	for i := 0; i < len(code); i++ {
		if code[i] == '\n' {
			lines = append(lines, code[start:i])
			start = i + 1
		}
	}
	lines = append(lines, code[start:])
	return lines
}

// codeItoa converts int to string (local, avoids import).
func codeItoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [8]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// Children returns nil.
func (cbs *CodeBlockStream) Children() []Component { return nil }
