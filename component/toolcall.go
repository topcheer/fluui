package component

import (
	"encoding/json"
	"fmt"
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
	header := t.buildHeaderLocked(bounds.W)
	t.drawStyledText(buf, bounds.X, y, bounds.W, header, t.headerStyleLocked())
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
		argLines := strings.Split(argText, "\n")
		used := t.drawBorderedSection(buf, bounds.X, y, bounds.W, remaining, "args", argLines, t.argLineStyle())
		y += used
		remaining -= used
		if remaining <= 0 {
			return
		}
	}

	// --- Result section ---
	if t.result != "" && remaining > 1 {
		resultLines := strings.Split(t.result, "\n")
		previewLines := resultLines
		if !t.showFull && len(resultLines) > t.maxResultPreview {
			previewLines = resultLines[:t.maxResultPreview]
		}
		used := t.drawBorderedSection(buf, bounds.X, y, bounds.W, remaining, "result", previewLines, t.resultLineStyle())
		y += used
		remaining -= used

		// "show more" hint
		if !t.showFull && len(resultLines) > t.maxResultPreview && remaining > 0 {
			more := len(resultLines) - t.maxResultPreview
			hint := fmt.Sprintf("  ⤷ %d more lines… (toggle to expand)", more)
			hintStyle := buffer.Style{Fg: theme.Get().Muted}
			t.drawStyledText(buf, bounds.X, y, bounds.W, hint, hintStyle)
		}
	}
}

// buildHeaderLocked constructs the header text: "icon toolName  duration"
func (t *ToolCallView) buildHeaderLocked(maxW int) string {
	icon := t.statusIconLocked()
	dur := formatToolDuration(t.durationLocked())
	header := icon + " " + t.toolName + "  " + dur
	// Use utf8.RuneCountInString to avoid []rune allocation
	if utf8.RuneCountInString(header) > maxW {
		// Truncate by runes
		return truncateRunes(header, maxW-1) + "…"
	}
	return header
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
func (t *ToolCallView) drawBorderedSection(buf *buffer.Buffer, x, y, w, maxH int, label string, lines []string, contentStyle buffer.Style) int {
	borderStyle := buffer.Style{Fg: theme.Get().Border}

	// Available height: 1 (top border) + len(lines) + 1 (bottom border)
	// But cap to maxH
	contentMax := maxH - 2 // minus two borders
	if contentMax < 1 {
		contentMax = 1
	}
	linesToShow := lines
	if len(linesToShow) > contentMax {
		linesToShow = linesToShow[:contentMax]
	}

	totalH := 2 + len(linesToShow)
	if totalH > maxH {
		totalH = maxH
	}

	// Top border: ╭─ label ──╮
	labelRunes := []rune(" " + label + " ")
	labelW := len(labelRunes)
	dashesTotal := w - 2 - labelW
	if dashesTotal < 0 {
		dashesTotal = 0
	}
	leftDashes := dashesTotal / 2
	rightDashes := dashesTotal - leftDashes

	drawX := x
	buf.DrawText(drawX, y, "╭", borderStyle)
	drawX++
	for i := 0; i < leftDashes; i++ {
		buf.DrawText(drawX, y, "─", borderStyle)
		drawX++
	}
	for _, r := range labelRunes {
		buf.DrawText(drawX, y, string(r), borderStyle)
		drawX++
	}
	for i := 0; i < rightDashes; i++ {
		buf.DrawText(drawX, y, "─", borderStyle)
		drawX++
	}
	if drawX < x+w {
		buf.DrawText(drawX, y, "╮", borderStyle)
	}

	// Content lines
	contentW := w - 2 // space inside │ borders
	if contentW < 1 {
		contentW = 1
	}
	for i, line := range linesToShow {
		ly := y + 1 + i
		if ly >= y+maxH-1 {
			break
		}
		buf.DrawText(x, ly, "│", borderStyle)
		lineRunes := []rune(line)
		if len(lineRunes) > contentW-1 {
			lineRunes = lineRunes[:contentW-1]
		}
		buf.DrawText(x+1, ly, " "+string(lineRunes), contentStyle)
		if x+w-1 > x {
			buf.DrawText(x+w-1, ly, "│", borderStyle)
		}
	}

	// Bottom border: ╰──╯
	botY := y + 1 + len(linesToShow)
	if botY < y+maxH {
		buf.DrawText(x, botY, "╰", borderStyle)
		for i := 1; i < w-1; i++ {
			buf.DrawText(x+i, botY, "─", borderStyle)
		}
		if w > 1 {
			buf.DrawText(x+w-1, botY, "╯", borderStyle)
		}
	}

	return totalH
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
	if d < time.Second {
		return strconv.FormatInt(d.Milliseconds(), 10) + "ms"
	}
	var buf [16]byte
	b := strconv.AppendFloat(buf[:0], d.Seconds(), 'f', 1, 64)
	return string(b) + "s"
}
