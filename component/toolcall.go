package component

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// ToolCallStatus represents the execution state of a tool call.
type ToolCallStatus uint8

const (
	// ToolCallRunning indicates the tool is currently executing.
	ToolCallRunning ToolCallStatus = iota
	// ToolCallCompleted indicates the tool finished successfully.
	ToolCallCompleted
	// ToolCallErrored indicates the tool failed.
	ToolCallErrored
)

// ToolCallView is a standalone component for rendering AI tool/function call
// invocations and their results. Unlike the block.ToolCallBlock (which is
// designed for the chat container framework), ToolCallView is a self-contained
// component usable directly in any layout.
//
// Features:
//   - Tool name + status icon header
//   - Collapsible JSON pretty-printed arguments
//   - Status indicator: ⠋ running, ✓ completed, ✗ errored
//   - Result preview with truncation and "show more" toggle
//   - Duration display
//   - Streaming partial results via AppendResult
//   - Thread-safe
type ToolCallView struct {
	BaseComponent
	mu sync.Mutex

	toolName  string
	args      string
	prettyArg string // pretty-printed args (if valid JSON)
	result    string
	status    ToolCallStatus
	expanded  bool // show full args
	showFull  bool // show full result (vs truncated)
	startTime time.Time
	endTime   time.Time
	spinnerF  int

	// Config
	maxResultPreview int // max lines of result to show when collapsed
}

// spinnerFrames for the running animation.
var toolCallViewSpinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewToolCallView creates a new tool call view in running state.
func NewToolCallView(toolName, args string) *ToolCallView {
	tc := &ToolCallView{
		BaseComponent: BaseComponent{id: GenerateID("toolcall")},
		toolName:      toolName,
		args:          args,
		status:        ToolCallRunning,
		startTime:     time.Now(),
		maxResultPreview: 3,
	}
	tc.prettyPrintArgs()
	return tc
}

// prettyPrintArgs attempts to JSON pretty-print the args.
func (t *ToolCallView) prettyPrintArgs() {
	if t.args == "" {
		return
	}
	var parsed any
	if err := json.Unmarshal([]byte(t.args), &parsed); err == nil {
		if formatted, err := json.MarshalIndent(parsed, "", "  "); err == nil {
			t.prettyArg = string(formatted)
			return
		}
	}
	t.prettyArg = t.args
}

// ToolName returns the tool name.
func (t *ToolCallView) ToolName() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.toolName
}

// Args returns the raw arguments string.
func (t *ToolCallView) Args() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.args
}

// Result returns the result text.
func (t *ToolCallView) Result() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.result
}

// Status returns the current status.
func (t *ToolCallView) Status() ToolCallStatus {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.status
}

// Expanded returns whether args are shown expanded.
func (t *ToolCallView) Expanded() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.expanded
}

// SetExpanded controls arg expand/collapse.
func (t *ToolCallView) SetExpanded(v bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.expanded = v
}

// Toggle switches expanded/collapsed.
func (t *ToolCallView) Toggle() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.expanded = !t.expanded
}

// SetArgs updates the arguments.
func (t *ToolCallView) SetArgs(args string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.args = args
	t.prettyPrintArgs()
}

// AppendResult appends streaming result text.
func (t *ToolCallView) AppendResult(delta string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.result += delta
}

// SetResult replaces the result text.
func (t *ToolCallView) SetResult(result string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.result = result
}

// Complete marks the tool call as completed successfully.
func (t *ToolCallView) Complete() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = ToolCallCompleted
	t.endTime = time.Now()
}

// Error marks the tool call as errored.
func (t *ToolCallView) Error() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.status = ToolCallErrored
	t.endTime = time.Now()
}

// Duration returns the elapsed time.
func (t *ToolCallView) Duration() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.durationLocked()
}

// durationLocked returns duration without acquiring the lock (caller must hold mu).
func (t *ToolCallView) durationLocked() time.Duration {
	if !t.endTime.IsZero() {
		return t.endTime.Sub(t.startTime)
	}
	return time.Since(t.startTime)
}

// AdvanceSpinner increments the spinner frame.
func (t *ToolCallView) AdvanceSpinner() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.spinnerF++
}

// SetMaxResultPreview sets the max result lines shown when collapsed.
func (t *ToolCallView) SetMaxResultPreview(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n < 1 {
		n = 1
	}
	t.maxResultPreview = n
}

// ShowFull returns whether the full result is shown.
func (t *ToolCallView) ShowFull() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.showFull
}

// SetShowFull controls whether to show the full result.
func (t *ToolCallView) SetShowFull(v bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.showFull = v
}

// Measure returns the desired size.
func (t *ToolCallView) Measure(cs Constraints) Size {
	t.mu.Lock()
	defer t.mu.Unlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	// Header is always 1 line
	h := 1

	// Args section (when expanded)
	if t.expanded && t.args != "" {
		argText := t.prettyArg
		if argText == "" {
			argText = t.args
		}
		argLines := strings.Count(argText, "\n") + 1
		h += 1 + argLines + 1 // border top + lines + border bottom
	}

	// Result section (when result exists)
	if t.result != "" {
		resultLines := strings.Count(t.result, "\n") + 1
		if !t.showFull && resultLines > t.maxResultPreview {
			resultLines = t.maxResultPreview
			h += 1 // "show more" line
		}
		h += 1 + resultLines + 1 // border top + lines + border bottom
	}

	if h < 1 {
		h = 1
	}
	return Size{W: maxW, H: h}
}

// Paint renders the tool call view.
func (t *ToolCallView) Paint(buf *buffer.Buffer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	bounds := t.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	y := bounds.Y
	remaining := bounds.H

	// --- Header line ---
	style := t.headerStyleLocked()
	icon := t.statusIconLocked()
	dur := formatToolDuration(t.durationLocked())

	// Calculate total header width for truncation check
	iconW := utf8.RuneCountInString(icon)
	nameW := utf8.RuneCountInString(t.toolName)
	durW := utf8.RuneCountInString(dur)
	totalW := iconW + 1 + nameW + 2 + durW

	if totalW > bounds.W {
		// Truncation needed — build string (1 alloc) for rare case
		header := icon + " " + t.toolName + "  " + dur
		header = truncateRunes(header, bounds.W-1) + "…"
		buf.DrawText(bounds.X, y, header, style)
	} else {
		// Zero-alloc: draw each piece directly
		x := bounds.X
		buf.DrawText(x, y, icon, style)
		x += iconW
		buf.DrawText(x, y, " ", style)
		x += 1
		buf.DrawText(x, y, t.toolName, style)
		x += nameW
		buf.DrawText(x, y, "  ", style)
		x += 2
		buf.DrawText(x, y, dur, style)
	}
	y++
	remaining--
	if remaining <= 0 {
		return
	}

	// --- Args section ---
	if t.expanded && t.args != "" && remaining > 1 {
		argText := t.prettyArg
		if argText == "" {
			argText = t.args
		}
		used := t.drawBorderedText(buf, bounds.X, y, bounds.W, remaining, "args", argText, 0, t.argLineStyle())
		y += used
		remaining -= used
		if remaining <= 0 {
			return
		}
	}

	// --- Result section ---
	if t.result != "" && remaining > 1 {
		maxLines := 0 // 0 = show all
		if !t.showFull {
			maxLines = t.maxResultPreview
		}
		used := t.drawBorderedText(buf, bounds.X, y, bounds.W, remaining, "result", t.result, maxLines, t.resultLineStyle())
		y += used
		remaining -= used

		// "show more" hint
		if !t.showFull {
			totalLines := countLines(t.result)
			if totalLines > t.maxResultPreview && remaining > 0 {
				more := totalLines - t.maxResultPreview
				var hbuf [64]byte
				hb := hbuf[:0]
				hb = append(hb, "  \u2937 "...)
				hb = strconv.AppendInt(hb, int64(more), 10)
				hb = append(hb, " more lines\u2026 (toggle to expand)"...)
				hintStyle := buffer.Style{Fg: theme.Get().Muted}
				buf.DrawBytes(bounds.X, y, hb, hintStyle)
			}
		}
	}
}

func (t *ToolCallView) statusIconLocked() string {
	switch t.status {
	case ToolCallRunning:
		return toolCallViewSpinnerFrames[t.spinnerF%len(toolCallViewSpinnerFrames)]
	case ToolCallCompleted:
		return "✓"
	case ToolCallErrored:
		return "✗"
	}
	return "?"
}

func (t *ToolCallView) headerStyleLocked() buffer.Style {
	switch t.status {
	case ToolCallRunning:
		return buffer.Style{Fg: theme.Get().Muted}
	case ToolCallCompleted:
		return buffer.Style{Fg: theme.Get().Success}
	case ToolCallErrored:
		return buffer.Style{Fg: theme.Get().Error}
	}
	return buffer.DefaultStyle
}

func (t *ToolCallView) argLineStyle() buffer.Style {
	return buffer.Style{Fg: theme.Get().Muted}
}

func (t *ToolCallView) resultLineStyle() buffer.Style {
	return buffer.Style{Fg: theme.Get().Fg}
}

// drawBorderedSection draws a ╭─ label ─╮ ... ╰─╯ box with content lines.
// Returns the number of rows used.
// drawBorderedText draws a bordered section with text content.
// Splits text on newlines internally (zero-alloc, no strings.Split).
// maxLines <= 0 means show all lines.
func (t *ToolCallView) drawBorderedText(buf *buffer.Buffer, x, y, w, maxH int, label string, text string, maxLines int, contentStyle buffer.Style) int {
	borderStyle := buffer.Style{Fg: theme.Get().Border}

	contentMax := maxH - 2
	if contentMax < 1 {
		contentMax = 1
	}
	lineLimit := contentMax
	if maxLines > 0 && maxLines < lineLimit {
		lineLimit = maxLines
	}

	// Count actual lines we'll draw by scanning text
	linesDrawn := 0
	pos := 0
	for pos <= len(text) && linesDrawn < lineLimit {
		nlIdx := strings.IndexByte(text[pos:], '\n')
		if nlIdx < 0 {
			linesDrawn++
			pos = len(text) + 1
		} else {
			linesDrawn++
			pos += nlIdx + 1
		}
	}

	totalH := 2 + linesDrawn
	if totalH > maxH {
		totalH = maxH
	}

	// Top border: ╭─ label ──╮
	labelText := " " + label + " "
	labelW := utf8.RuneCountInString(labelText)
	dashesTotal := w - 2 - labelW
	if dashesTotal < 0 {
		dashesTotal = 0
	}
	leftDashes := dashesTotal / 2
	rightDashes := dashesTotal - leftDashes

	drawX := x
	buf.DrawText(drawX, y, "\u256d", borderStyle) // ╭
	drawX++
	for i := 0; i < leftDashes; i++ {
		buf.DrawText(drawX, y, "\u2500", borderStyle) // ─
		drawX++
	}
	buf.DrawText(drawX, y, labelText, borderStyle)
	drawX += labelW
	for i := 0; i < rightDashes; i++ {
		buf.DrawText(drawX, y, "\u2500", borderStyle)
		drawX++
	}
	if drawX < x+w {
		buf.DrawText(drawX, y, "\u256e", borderStyle) // ╮
	}

	// Content lines — iterate over text without strings.Split
	contentW := w - 2
	if contentW < 1 {
		contentW = 1
	}
	lineIdx := 0
	pos = 0
	for pos <= len(text) && lineIdx < lineLimit {
		if lineIdx >= contentMax {
			break
		}
		var line string
		nlIdx := strings.IndexByte(text[pos:], '\n')
		if nlIdx < 0 {
			line = text[pos:]
			pos = len(text) + 1
		} else {
			line = text[pos : pos+nlIdx]
			pos += nlIdx + 1
		}

		ly := y + 1 + lineIdx
		if ly >= y+maxH-1 {
			break
		}
		buf.DrawText(x, ly, "\u2502", borderStyle) // │
		lineW := utf8.RuneCountInString(line)
		if lineW > contentW-1 {
			line = truncateRunes(line, contentW-1)
		}
		buf.DrawText(x+1, ly, " ", contentStyle)
		buf.DrawText(x+2, ly, line, contentStyle)
		if x+w-1 > x {
			buf.DrawText(x+w-1, ly, "\u2502", borderStyle)
		}
		lineIdx++
	}

	// Bottom border: ╰──╯
	botY := y + 1 + lineIdx
	if botY < y+maxH {
		buf.DrawText(x, botY, "\u2570", borderStyle) // ╰
		for i := 1; i < w-1; i++ {
			buf.DrawText(x+i, botY, "\u2500", borderStyle)
		}
		if w > 1 {
			buf.DrawText(x+w-1, botY, "\u256f", borderStyle) // ╯
		}
	}

	return totalH
}

// countLines returns the number of newline-separated lines in text (zero alloc).
func countLines(text string) int {
	if text == "" {
		return 0
	}
	count := 1
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			count++
		}
	}
	return count
}

// drawStyledText draws text clamped to width.
func (t *ToolCallView) drawStyledText(buf *buffer.Buffer, x, y, maxW int, text string, style buffer.Style) {
	if utf8.RuneCountInString(text) <= maxW {
		buf.DrawText(x, y, text, style)
		return
	}
	buf.DrawText(x, y, truncateRunes(text, maxW), style)
}

// formatToolDuration formats a duration for display (e.g. "1.2s", "234ms").
func formatToolDuration(d time.Duration) string {
	if d < time.Millisecond {
		return "0ms"
	}
	var buf [24]byte
	b := buf[:0]
	if d < time.Second {
		b = strconv.AppendInt(b, d.Milliseconds(), 10)
		b = append(b, "ms"...)
	} else {
		b = strconv.AppendFloat(b, d.Seconds(), 'f', 1, 64)
		b = append(b, 's')
	}
	return string(b)
}
