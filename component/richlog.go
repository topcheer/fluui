package component

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// LogLevel describes the severity of a log entry.
type LogLevel uint8

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
	LogFatal
)

// LogEntry represents a single log line.
type LogEntry struct {
	Timestamp time.Time
	Level     LogLevel
	Text      string
}

// RichLogStyle controls the visual appearance of the RichLog.
type RichLogStyle struct {
	// Style per log level
	DebugStyle buffer.Style
	InfoStyle  buffer.Style
	WarnStyle  buffer.Style
	ErrorStyle buffer.Style
	FatalStyle buffer.Style
	// Timestamp style
	TimestampStyle buffer.Style
	// Default text style
	TextStyle buffer.Style
}

// DefaultRichLogStyle returns a dark-terminal-friendly style set.
func DefaultRichLogStyle() RichLogStyle {
	gray := buffer.NamedColor(buffer.NamedBrightBlack)
	cyan := buffer.NamedColor(buffer.NamedCyan)
	yellow := buffer.NamedColor(buffer.NamedYellow)
	red := buffer.NamedColor(buffer.NamedRed)
	brightRed := buffer.NamedColor(buffer.NamedBrightRed)
	white := buffer.NamedColor(buffer.NamedWhite)
	return RichLogStyle{
		DebugStyle:     buffer.Style{Fg: gray, Flags: buffer.Dim},
		InfoStyle:      buffer.Style{Fg: cyan},
		WarnStyle:      buffer.Style{Fg: yellow, Flags: buffer.Bold},
		ErrorStyle:     buffer.Style{Fg: red, Flags: buffer.Bold},
		FatalStyle:     buffer.Style{Fg: brightRed, Bg: red, Flags: buffer.Bold | buffer.Reverse},
		TimestampStyle: buffer.Style{Fg: gray, Flags: buffer.Dim},
		TextStyle:      buffer.Style{Fg: white},
	}
}

// LevelName returns a 5-char uppercase level name.
func LevelName(level LogLevel) string {
	switch level {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return " INFO"
	case LogWarn:
		return " WARN"
	case LogError:
		return "ERROR"
	case LogFatal:
		return "FATAL"
	default:
		return "?????"
	}
}

// LevelColor returns the color used for a given log level.
func LevelColor(level LogLevel, style RichLogStyle) buffer.Style {
	switch level {
	case LogDebug:
		return style.DebugStyle
	case LogInfo:
		return style.InfoStyle
	case LogWarn:
		return style.WarnStyle
	case LogError:
		return style.ErrorStyle
	case LogFatal:
		return style.FatalStyle
	default:
		return style.TextStyle
	}
}

// RichLog is a scrollable log viewer that accepts structured log entries
// with levels, optional timestamps, and auto-scroll behavior.
// It is safe for concurrent use.
type RichLog struct {
	BaseComponent
	mu sync.RWMutex

	entries []LogEntry
	maxSize int // maximum number of entries (0 = unlimited)

	// display config
	style      RichLogStyle
	showLevels bool
	showTime   bool
	autoScroll bool // follow new entries (tail mode)

	// scrolling
	scrollY    int
	following  bool // true when auto-scrolling (at bottom)
	hdrWidth   int  // cached width of "[HH:MM:SS] LEVEL " prefix

	// filter
	minLevel LogLevel // entries below this level are hidden
}

// NewRichLog creates a RichLog with sensible defaults:
// max 10000 entries, show levels, show timestamps, auto-scroll on.
func NewRichLog() *RichLog {
	style := DefaultRichLogStyle()
	return &RichLog{
		maxSize:   10000,
		style:     style,
		showLevels: true,
		showTime:   true,
		autoScroll: true,
		following:  true,
		minLevel:   LogDebug,
		hdrWidth:   15, // "[HH:MM:SS] LEVEL " = 15 chars
	}
}

// ─── Configuration ───

// SetStyle sets the visual style.
func (r *RichLog) SetStyle(s RichLogStyle) {
	r.mu.Lock()
	r.style = s
	r.mu.Unlock()
}

// Style returns the current style.
func (r *RichLog) Style() RichLogStyle {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.style
}

// SetMaxSize sets the maximum number of stored entries (0 = unlimited).
// Excess entries are trimmed from the front.
func (r *RichLog) SetMaxSize(n int) {
	r.mu.Lock()
	r.maxSize = n
	r.trimLocked()
	r.mu.Unlock()
}

// MaxSize returns the configured maximum.
func (r *RichLog) MaxSize() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.maxSize
}

// SetShowLevels toggles level prefix display.
func (r *RichLog) SetShowLevels(b bool) {
	r.mu.Lock()
	r.showLevels = b
	r.recomputeHdrWidthLocked()
	r.mu.Unlock()
}

// ShowLevels returns whether levels are shown.
func (r *RichLog) ShowLevels() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.showLevels
}

// SetShowTime toggles timestamp prefix display.
func (r *RichLog) SetShowTime(b bool) {
	r.mu.Lock()
	r.showTime = b
	r.recomputeHdrWidthLocked()
	r.mu.Unlock()
}

// ShowTime returns whether timestamps are shown.
func (r *RichLog) ShowTime() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.showTime
}

// SetAutoScroll toggles tail-follow behavior.
// When true, the view scrolls to show the latest entry on each write/paint.
func (r *RichLog) SetAutoScroll(b bool) {
	r.mu.Lock()
	r.autoScroll = b
	r.following = b
	if b {
		r.scrollY = 0
	}
	r.mu.Unlock()
}

// AutoScroll returns the auto-scroll setting.
func (r *RichLog) AutoScroll() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.autoScroll
}

// SetMinLevel sets the minimum visible log level.
func (r *RichLog) SetMinLevel(level LogLevel) {
	r.mu.Lock()
	r.minLevel = level
	r.scrollY = 0
	r.mu.Unlock()
}

// MinLevel returns the minimum visible level.
func (r *RichLog) MinLevel() LogLevel {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.minLevel
}

// recomputeHdrWidthLocked calculates the width of the header prefix.
// Must be called with lock held.
func (r *RichLog) recomputeHdrWidthLocked() {
	w := 0
	if r.showTime {
		w += 10 // "[HH:MM:SS] "
	}
	if r.showLevels {
		w += 6 // "LEVEL "
	}
	r.hdrWidth = w
}

// ─── Writing ───

// Write adds a log entry with the given level and text.
func (r *RichLog) Write(level LogLevel, text string) {
	r.mu.Lock()
	e := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Text:      text,
	}
	r.entries = append(r.entries, e)
	r.trimLocked()
	if r.autoScroll {
		r.following = true
		r.scrollY = 0
	}
	r.mu.Unlock()
}

// Writef adds a formatted log entry.
func (r *RichLog) Writef(level LogLevel, format string, args ...interface{}) {
	r.Write(level, sprintf(format, args...))
}

// Info is shorthand for Write(LogInfo, text).
func (r *RichLog) Info(text string)         { r.Write(LogInfo, text) }
func (r *RichLog) Infof(format string, a ...interface{}) {
	r.Writef(LogInfo, format, a...)
}

// Warn is shorthand for Write(LogWarn, text).
func (r *RichLog) Warn(text string) { r.Write(LogWarn, text) }
func (r *RichLog) Warnf(format string, a ...interface{}) {
	r.Writef(LogWarn, format, a...)
}

// Error is shorthand for Write(LogError, text).
func (r *RichLog) Error(text string) { r.Write(LogError, text) }
func (r *RichLog) Errorf(format string, a ...interface{}) {
	r.Writef(LogError, format, a...)
}

// Debug is shorthand for Write(LogDebug, text).
func (r *RichLog) Debug(text string) { r.Write(LogDebug, text) }
func (r *RichLog) Debugf(format string, a ...interface{}) {
	r.Writef(LogDebug, format, a...)
}

// Fatal is shorthand for Write(LogFatal, text).
func (r *RichLog) Fatal(text string) { r.Write(LogFatal, text) }
func (r *RichLog) Fatalf(format string, a ...interface{}) {
	r.Writef(LogFatal, format, a...)
}

// WriteLine adds an info-level entry (Textual-compatible alias).
func (r *RichLog) WriteLine(text string) { r.Write(LogInfo, text) }

// trimLocked removes excess entries from the front.
// Must be called with lock held.
func (r *RichLog) trimLocked() {
	if r.maxSize > 0 && len(r.entries) > r.maxSize {
		excess := len(r.entries) - r.maxSize
		r.entries = r.entries[excess:]
	}
}

// Clear removes all entries.
func (r *RichLog) Clear() {
	r.mu.Lock()
	r.entries = nil
	r.scrollY = 0
	r.mu.Unlock()
}

// ─── Reading ───

// EntryCount returns the total number of stored entries.
func (r *RichLog) EntryCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}

// Entries returns a defensive copy of all entries.
func (r *RichLog) Entries() []LogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]LogEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

// ScrollY returns the current scroll offset from the bottom (0 = at bottom).
func (r *RichLog) ScrollY() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.scrollY
}

// Following returns whether the view is in follow (tail) mode.
func (r *RichLog) Following() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.following
}

// ─── Scrolling ───

// ScrollUp scrolls up by n lines (toward older entries).
func (r *RichLog) ScrollUp(n int) {
	r.mu.Lock()
	r.scrollY += n
	r.following = false
	r.mu.Unlock()
}

// ScrollDown scrolls down by n lines (toward newer entries).
func (r *RichLog) ScrollDown(n int) {
	r.mu.Lock()
	r.scrollY -= n
	if r.scrollY <= 0 {
		r.scrollY = 0
		r.following = r.autoScroll
	}
	r.mu.Unlock()
}

// ScrollToBottom jumps to the latest entries and re-enables follow mode.
func (r *RichLog) ScrollToBottom() {
	r.mu.Lock()
	r.scrollY = 0
	r.following = r.autoScroll
	r.mu.Unlock()
}

// ScrollToTop jumps to the oldest entries.
func (r *RichLog) ScrollToTop() {
	r.mu.Lock()
	r.following = false
	// scrollY is offset from bottom; set to visibleCount - totalVisibleLines
	// We'll cap it during paint
	r.scrollY = max(0, len(r.entries))
	r.mu.Unlock()
}

// ─── Layout ───

// Measure returns the preferred size.
func (r *RichLog) Measure(cs Constraints) Size {
	r.mu.RLock()
	lines := r.countVisibleLinesLocked()
	r.mu.RUnlock()
	w := cs.MaxWidth
	if w <= 0 {
		w = 80
	}
	if lines < 1 {
		lines = 1
	}
	return Size{W: w, H: lines}
}

// countVisibleLinesLocked returns the number of display lines
// for visible entries. Must be called with lock held.
func (r *RichLog) countVisibleLinesLocked() int {
	if r.bounds.H <= 0 {
		return len(r.entries)
	}
	w := r.bounds.W
	if w <= 0 {
		w = 80
	}
	contentW := w - r.hdrWidth
	if contentW < 1 {
		contentW = 1
	}
	total := 0
	for _, e := range r.entries {
		if e.Level < r.minLevel {
			continue
		}
		total += wrapLineCount(e.Text, contentW)
	}
	return total
}

// Paint renders the visible entries into the buffer.
func (r *RichLog) Paint(buf *buffer.Buffer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	w := r.bounds.W
	h := r.bounds.H
	if w <= 0 || h <= 0 {
		return
	}

	contentW := w - r.hdrWidth
	if contentW < 1 {
		contentW = 1
	}

	// Build flat list of display lines for visible entries
	type dispLine struct {
		text   string
		level  LogLevel
		ts     time.Time
	}
	var allLines []dispLine
	for _, e := range r.entries {
		if e.Level < r.minLevel {
			continue
		}
		wrapped := wrapText(e.Text, contentW)
		for _, wl := range wrapped {
			allLines = append(allLines, dispLine{
				text:  wl,
				level: e.Level,
				ts:    e.Timestamp,
			})
		}
	}

	// Apply scroll
	totalLines := len(allLines)
	startIdx := totalLines - h - r.scrollY
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + h
	if endIdx > totalLines {
		endIdx = totalLines
	}

	// Render
	x0 := r.bounds.X
	y0 := r.bounds.Y
	row := 0
	for i := startIdx; i < endIdx; i++ {
		dl := allLines[i]
		x := x0
		y := y0 + row

		// Timestamp
		if r.showTime && dl.ts.After(time.Time{}) {
			tsStr := dl.ts.Format("15:04:05")
			tsCell := "[" + tsStr + "] "
			buf.DrawText(x, y, tsCell, r.style.TimestampStyle)
			x += len(tsCell)
		}

		// Level
		if r.showLevels {
			lvlStr := LevelName(dl.level) + " "
			lvlStyle := LevelColor(dl.level, r.style)
			buf.DrawText(x, y, lvlStr, lvlStyle)
			x += len(lvlStr)
		}

		// Text
		if x < x0+w {
			textStyle := LevelColor(dl.level, r.style)
			if textStyle.Fg.Type == 0 && textStyle.Flags == 0 {
				textStyle = r.style.TextStyle
			}
			// Truncate to fit
			text := dl.text
			availW := x0 + w - x
			if availW < 0 {
				availW = 0
			}
			if visibleWidth(text) > availW {
				text = truncateRunesLocal(text, availW)
			}
			buf.DrawText(x, y, text, textStyle)
		}

		row++
	}

	// Auto-scroll: if following, ensure we're at the bottom
	if r.autoScroll && r.following {
		r.scrollY = 0
	}
}

// HandleKey processes keyboard navigation.
func (r *RichLog) HandleKey(k *term.KeyEvent) bool {
	switch k.Key {
	case term.KeyUp, term.KeyDown:
		if k.Key == term.KeyUp {
			r.ScrollUp(1)
		} else {
			r.ScrollDown(1)
		}
		return true
	case term.KeyPageUp:
		r.mu.RLock()
		n := r.bounds.H
		r.mu.RUnlock()
		if n <= 0 {
			n = 10
		}
		r.ScrollUp(n)
		return true
	case term.KeyPageDown:
		r.mu.RLock()
		n := r.bounds.H
		r.mu.RUnlock()
		if n <= 0 {
			n = 10
		}
		r.ScrollDown(n)
		return true
	case term.KeyHome:
		r.ScrollToTop()
		return true
	case term.KeyEnd:
		r.ScrollToBottom()
		return true
	}

	// Vim keys
	if k.Rune != 0 {
		switch k.Rune {
		case 'j':
			r.ScrollDown(1)
			return true
		case 'k':
			r.ScrollUp(1)
			return true
		case 'g':
			r.ScrollToTop()
			return true
		case 'G':
			r.ScrollToBottom()
			return true
		}
	}

	return false
}

// ─── Helpers ───

// wrapText splits text into lines that fit within width.
// Each returned string is guaranteed to be at most `width` visible columns.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	if text == "" {
		return []string{""}
	}
	lines := strings.Split(text, "\n")
	var result []string
	for _, line := range lines {
		if visibleWidth(line) <= width {
			result = append(result, line)
			continue
		}
		// Word-wrap
		words := strings.Fields(line)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		current := ""
		for _, word := range words {
			if current == "" {
				current = word
			} else if visibleWidth(current)+1+visibleWidth(word) <= width {
				current += " " + word
			} else {
				result = append(result, current)
				current = word
			}
		}
		if current != "" {
			result = append(result, current)
		}
	}
	if len(result) == 0 {
		result = []string{""}
	}
	return result
}

// wrapLineCount returns the number of wrapped lines for given text and width.
func wrapLineCount(text string, width int) int {
	return len(wrapText(text, width))
}

// visibleWidth returns the display width of a string (rune count).
func visibleWidth(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// truncateRunesLocal truncates a string to at most maxRunes visible runes.
func truncateRunesLocal(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

// sprintf wraps fmt.Sprintf for formatted log entries.
func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
