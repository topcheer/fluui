package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CodeEditor: Syntax-Highlighted Code Display ───
//
// CodeEditor renders code with line numbers, basic syntax highlighting for
// keywords, and a cursor position indicator. It is a read-only display
// component optimized for AI-generated code previews and code review UIs.
//
// Usage:
//
//	ce := NewCodeEditor()
//	ce.SetLanguage("go")
//	ce.SetCode("package main\n\nfunc main() {\n    println(\"hello\")\n}\n")
//	ce.SetCursorLine(3)
//	ce.Paint(buf)

// CodeEditorStyle holds styling for CodeEditor.
type CodeEditorStyle struct {
	Normal     buffer.Style
	Keyword    buffer.Style
	String     buffer.Style
	Comment    buffer.Style
	Number     buffer.Style
	LineNumber buffer.Style
	Cursor     buffer.Style
	Border     buffer.Style
}

// DefaultCodeEditorStyle returns sensible defaults.
func DefaultCodeEditorStyle() CodeEditorStyle {
	normal := buffer.Style{Fg: buffer.RGB(226, 232, 240)}    // slate-200
	keyword := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold} // violet-400 bold
	str := buffer.Style{Fg: buffer.RGB(134, 239, 172)}       // green-300
	comment := buffer.Style{Fg: buffer.RGB(100, 116, 139)}   // slate-500
	number := buffer.Style{Fg: buffer.RGB(251, 146, 60)}    // orange-400
	lineNum := buffer.Style{Fg: buffer.RGB(71, 85, 105)}     // slate-600
	cursor := buffer.Style{Fg: buffer.RGB(250, 204, 21), Flags: buffer.Underline} // yellow-400 underline
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}      // slate-600
	return CodeEditorStyle{
		Normal:     normal,
		Keyword:    keyword,
		String:     str,
		Comment:    comment,
		Number:     number,
		LineNumber: lineNum,
		Cursor:     cursor,
		Border:     border,
	}
}

// keywordSets holds language-specific keyword lookup maps.
var codeKeywordSets = map[string]map[string]bool{
	"go": {
		"break": true, "case": true, "chan": true, "const": true, "continue": true,
		"default": true, "defer": true, "else": true, "fallthrough": true, "for": true,
		"func": true, "go": true, "goto": true, "if": true, "import": true,
		"interface": true, "map": true, "package": true, "range": true, "return": true,
		"select": true, "struct": true, "switch": true, "type": true, "var": true,
	},
	"python": {
		"def": true, "class": true, "if": true, "elif": true, "else": true,
		"for": true, "while": true, "try": true, "except": true, "finally": true,
		"return": true, "import": true, "from": true, "as": true, "with": true,
		"lambda": true, "yield": true, "global": true, "nonlocal": true, "pass": true,
		"break": true, "continue": true, "raise": true, "assert": true, "del": true,
		"in": true, "is": true, "not": true, "and": true, "or": true, "None": true,
		"True": true, "False": true, "self": true,
	},
	"javascript": {
		"var": true, "let": true, "const": true, "function": true, "return": true,
		"if": true, "else": true, "for": true, "while": true, "do": true,
		"switch": true, "case": true, "break": true, "continue": true,
		"class": true, "extends": true, "super": true, "new": true, "this": true,
		"import": true, "export": true, "from": true, "default": true,
		"try": true, "catch": true, "finally": true, "throw": true,
		"typeof": true, "instanceof": true, "void": true, "delete": true,
		"async": true, "await": true, "yield": true,
	},
}

// CodeEditor renders code with line numbers and basic syntax highlighting.
type CodeEditor struct {
	BaseComponent
	mu sync.Mutex

	language       string
	code           string
	cachedLines    []string
	showLineNums   bool
	cursorLine     int

	style CodeEditorStyle
}

// NewCodeEditor creates a CodeEditor with defaults.
func NewCodeEditor() *CodeEditor {
	ce := &CodeEditor{
		language:     "go",
		showLineNums: true,
		cursorLine:   -1,
		style:        DefaultCodeEditorStyle(),
	}
	ce.SetID(GenerateID("editor"))
	return ce
}

// SetLanguage sets the syntax highlighting language ("go", "python", "javascript").
func (ce *CodeEditor) SetLanguage(lang string) *CodeEditor {
	ce.mu.Lock()
	ce.language = lang
	ce.mu.Unlock()
	return ce
}

// Language returns the current language.
func (ce *CodeEditor) Language() string {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	return ce.language
}

// SetCode sets the code content. Lines are cached via strings.Split.
func (ce *CodeEditor) SetCode(text string) *CodeEditor {
	ce.mu.Lock()
	ce.code = text
	// Cache lines — trim trailing newline to avoid empty last line
	trimmed := strings.TrimRight(text, "\n")
	if trimmed == "" {
		ce.cachedLines = nil
	} else {
		ce.cachedLines = strings.Split(trimmed, "\n")
	}
	ce.mu.Unlock()
	return ce
}

// Code returns the current code content.
func (ce *CodeEditor) Code() string {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	return ce.code
}

// LineCount returns the number of code lines.
func (ce *CodeEditor) LineCount() int {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	return len(ce.cachedLines)
}

// SetShowLineNumbers toggles line number display.
func (ce *CodeEditor) SetShowLineNumbers(v bool) *CodeEditor {
	ce.mu.Lock()
	ce.showLineNums = v
	ce.mu.Unlock()
	return ce
}

// ShowLineNumbers returns whether line numbers are shown.
func (ce *CodeEditor) ShowLineNumbers() bool {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	return ce.showLineNums
}

// SetCursorLine sets the cursor line (0-based). Use -1 to hide.
func (ce *CodeEditor) SetCursorLine(line int) *CodeEditor {
	ce.mu.Lock()
	ce.cursorLine = line
	ce.mu.Unlock()
	return ce
}

// CursorLine returns the current cursor line.
func (ce *CodeEditor) CursorLine() int {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	return ce.cursorLine
}

// SetStyle sets the custom style.
func (ce *CodeEditor) SetStyle(s CodeEditorStyle) *CodeEditor {
	ce.mu.Lock()
	ce.style = s
	ce.mu.Unlock()
	return ce
}

// Measure returns the preferred size.
func (ce *CodeEditor) Measure(cs Constraints) Size {
	ce.mu.Lock()
	lineCount := len(ce.cachedLines)
	ce.mu.Unlock()

	w := 60
	h := lineCount + 2 // lines + borders
	if h < 5 {
		h = 5
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// classifyTokenLocked determines the style for a word token.
func (ce *CodeEditor) classifyTokenLocked(word string) buffer.Style {
	// Check if it's a keyword
	if kwSet, ok := codeKeywordSets[ce.language]; ok {
		if kwSet[word] {
			return ce.style.Keyword
		}
	}
	// Check if numeric
	isNum := len(word) > 0
	for _, r := range word {
		if r < '0' || r > '9' {
			if r == '.' {
				continue
			}
			isNum = false
			break
		}
	}
	if isNum {
		return ce.style.Number
	}
	return ce.style.Normal
}

// Paint renders the code editor into the buffer.
func (ce *CodeEditor) Paint(buf *buffer.Buffer) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	b := ce.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 20 {
		w = 60
	}
	if h < 5 {
		h = 5
	}

	// Draw border
	bs := ce.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Calculate line number column width
	lineNumW := 0
	if ce.showLineNums {
		lineCount := len(ce.cachedLines)
		if lineCount < 10 {
			lineNumW = 3 // " N "
		} else if lineCount < 100 {
			lineNumW = 4 // " NN "
		} else {
			lineNumW = 5 // " NNN "
		}
	}

	codeStartX := x + 1 + lineNumW
	codeEndX := x + w - 2

	// Draw each line
	for idx, line := range ce.cachedLines {
		rowY := y + 1 + idx
		if rowY >= y+h-1 || rowY >= buf.Height {
			break
		}

		// Line number
		if ce.showLineNums {
			numStr := itoa(idx + 1)
			// Right-align line number in the column
			numStart := x + 1 + lineNumW - len(numStr) - 1
			lnStyle := ce.style.LineNumber
			if idx == ce.cursorLine {
				lnStyle = ce.style.Cursor
			}
			for i, r := range numStr {
				cx := numStart + i
				if cx < buf.Width && cx >= x+1 {
					buf.SetCell(cx, rowY, buffer.Cell{Rune: r, Fg: lnStyle.Fg, Bg: lnStyle.Bg, Flags: lnStyle.Flags, Width: 1})
				}
			}
		}

		// Code content with basic syntax highlighting
		isComment := strings.HasPrefix(strings.TrimSpace(line), "//") ||
			strings.HasPrefix(strings.TrimSpace(line), "#")
		col := codeStartX

		if isComment {
			// Entire line is a comment
			commentStyle := ce.style.Comment
			for _, r := range line {
				if col >= codeEndX || col >= buf.Width {
					break
				}
				buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: commentStyle.Fg, Bg: commentStyle.Bg, Flags: commentStyle.Flags, Width: 1})
				col++
			}
		} else {
				// Tokenize by scanning runes directly — avoids strings.Fields and []rune allocations
				normalStyle := ce.style.Normal
				strStyle := ce.style.String
				inString := false
				wordStart := -1

				for i, r := range line {
					if col >= codeEndX || col >= buf.Width {
						break
					}

					// Track string state
					if r == '"' {
						inString = !inString
						if wordStart < 0 {
							wordStart = i
						}
						// Draw the quote char as string style
						buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: strStyle.Fg, Bg: strStyle.Bg, Flags: strStyle.Flags, Width: 1})
						col++
						continue
					}

					if inString {
						// Inside string — all chars get string style
						buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: strStyle.Fg, Bg: strStyle.Bg, Flags: strStyle.Flags, Width: 1})
						col++
						continue
					}

					// Check for word boundary (space, tab, punctuation that ends words)
					isWordChar := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
						(r >= '0' && r <= '9') || r == '_' || r == '.'

					if isWordChar {
						if wordStart < 0 {
							wordStart = i
						}
						// Don't draw yet — accumulate word to check if keyword
					} else {
						// Word boundary — flush accumulated word if any
						if wordStart >= 0 {
							// Extract the word from line using indices
							word := line[wordStart:i]
							tokenStyle := ce.classifyTokenLocked(word)
							for _, wr := range word {
								if col >= codeEndX || col >= buf.Width {
									break
								}
								buf.SetCell(col, rowY, buffer.Cell{Rune: wr, Fg: tokenStyle.Fg, Bg: tokenStyle.Bg, Flags: tokenStyle.Flags, Width: 1})
								col++
							}
							wordStart = -1
						}
						// Draw the non-word character in normal style
						buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: normalStyle.Fg, Bg: normalStyle.Bg, Flags: normalStyle.Flags, Width: 1})
						col++
					}
				}
				// Flush trailing word
				if wordStart >= 0 && col < codeEndX && col < buf.Width {
					word := line[wordStart:]
					tokenStyle := ce.classifyTokenLocked(word)
					for _, wr := range word {
						if col >= codeEndX || col >= buf.Width {
							break
						}
						buf.SetCell(col, rowY, buffer.Cell{Rune: wr, Fg: tokenStyle.Fg, Bg: tokenStyle.Bg, Flags: tokenStyle.Flags, Width: 1})
						col++
					}
				}
			}

		// Cursor indicator — underline the entire line
		if idx == ce.cursorLine {
			cursorStyle := ce.style.Cursor
			for cx := codeStartX; cx <= codeEndX && cx < buf.Width; cx++ {
				existing := buf.GetCell(cx, rowY)
				buf.SetCell(cx, rowY, buffer.Cell{
					Rune:  existing.Rune,
					Fg:    cursorStyle.Fg,
					Bg:    cursorStyle.Bg,
					Flags: cursorStyle.Flags,
					Width: existing.Width,
				})
			}
		}
	}
}

// Children returns nil.
func (ce *CodeEditor) Children() []Component { return nil }
