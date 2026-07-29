package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StreamingDiff: Live Text Diff Streaming Display ───
//
// StreamingDiff renders a side-by-side or inline diff that updates in
// real-time as new text streams in. Shows added lines in green, removed
// lines in red, and unchanged lines in dim. Useful for watching AI
// edit/code-change output stream in real-time.
//
// Usage:
//
//	sd := NewStreamingDiff()
//	sd.SetOldText("hello world")
//	sd.SetNewText("hello universe")
//	sd.Paint(buf)

// SDiffLineType represents the type of a diff line.
type SDiffLineType int

const (
	SDiffSame    SDiffLineType = 0
	SDiffAdded   SDiffLineType = 1
	SDiffRemoved SDiffLineType = 2
)

var diffLinePrefix = [3]rune{' ', '+', '-'}

// StreamingDiffStyle holds styling.
type StreamingDiffStyle struct {
	Added   buffer.Style
	Removed buffer.Style
	Same    buffer.Style
	Prefix  buffer.Style
}

// DefaultStreamingDiffStyle returns defaults.
func DefaultStreamingDiffStyle() StreamingDiffStyle {
	return StreamingDiffStyle{
		Added:   buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Removed: buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Same:    buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Prefix:  buffer.Style{Fg: buffer.RGB(71, 85, 105), Flags: buffer.Bold},
	}
}

const streamingDiffMaxLines = 40

// diffLine holds a single rendered diff line.
type diffLine struct {
	text     string
	lineType SDiffLineType
}

// StreamingDiff renders a live streaming diff.
type StreamingDiff struct {
	BaseComponent
	mu sync.Mutex

	oldText string
	newText string
	style   StreamingDiffStyle
	// cached
	lines    [streamingDiffMaxLines]diffLine
	count    int
	addCount int
	delCount int
}

// NewStreamingDiff creates a StreamingDiff.
func NewStreamingDiff() *StreamingDiff {
	sd := &StreamingDiff{style: DefaultStreamingDiffStyle()}
	sd.SetID(GenerateID("sdiff"))
	return sd
}

// SetOldText sets the original text.
func (sd *StreamingDiff) SetOldText(s string) *StreamingDiff {
	sd.mu.Lock()
	sd.oldText = s
	sd.recomputeLocked()
	sd.mu.Unlock()
	return sd
}

// SetNewText sets the new text.
func (sd *StreamingDiff) SetNewText(s string) *StreamingDiff {
	sd.mu.Lock()
	sd.newText = s
	sd.recomputeLocked()
	sd.mu.Unlock()
	return sd
}

func (sd *StreamingDiff) recomputeLocked() {
	sd.count = 0
	sd.addCount = 0
	sd.delCount = 0

	oldLines := splitLinesSimple(sd.oldText)
	newLines := splitLinesSimple(sd.newText)

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	for i := 0; i < maxLen && sd.count < streamingDiffMaxLines; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			sd.lines[sd.count] = diffLine{text: oldLine, lineType: SDiffSame}
		} else {
			if oldLine != "" {
				sd.lines[sd.count] = diffLine{text: oldLine, lineType: SDiffRemoved}
				sd.count++
				sd.delCount++
			}
			if newLine != "" && sd.count < streamingDiffMaxLines {
				sd.lines[sd.count] = diffLine{text: newLine, lineType: SDiffAdded}
				sd.addCount++
			}
		}
		if sd.count >= streamingDiffMaxLines {
			break
		}
		sd.count++
	}
}

func splitLinesSimple(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i, r := range s {
		if r == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// AddCount returns the number of added lines.
func (sd *StreamingDiff) AddCount() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.addCount
}

// DelCount returns the number of removed lines.
func (sd *StreamingDiff) DelCount() int {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.delCount
}

// SetStyle sets custom style.
func (sd *StreamingDiff) SetStyle(s StreamingDiffStyle) *StreamingDiff {
	sd.mu.Lock()
	sd.style = s
	sd.mu.Unlock()
	return sd
}

// Measure returns preferred size.
func (sd *StreamingDiff) Measure(cs Constraints) Size {
	sd.mu.Lock()
	h := sd.count
	sd.mu.Unlock()
	if h < 1 {
		h = 1
	}
	w := 40
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the streaming diff.
func (sd *StreamingDiff) Paint(buf *buffer.Buffer) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	b := sd.Bounds()
	x, y := b.X, b.Y

	for i := 0; i < sd.count; i++ {
		yy := y + i
		if yy >= buf.Height {
			break
		}
		col := x

		line := sd.lines[i]
		var st buffer.Style
		switch line.lineType {
		case SDiffAdded:
			st = sd.style.Added
		case SDiffRemoved:
			st = sd.style.Removed
		default:
			st = sd.style.Same
		}

		prefixStyle := sd.style.Prefix

		// Prefix char
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: diffLinePrefix[line.lineType], Fg: prefixStyle.Fg, Bg: prefixStyle.Bg, Flags: prefixStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
			col++
		}

		// Text
		for _, r := range line.text {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (sd *StreamingDiff) Children() []Component { return nil }
