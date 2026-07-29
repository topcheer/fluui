package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── LogViewer: Scrollable Log Entry Viewer ───
//
// LogViewer renders a compact log viewer showing recent log entries with
// severity-based color coding. Supports INFO, WARN, ERROR, and DEBUG levels.
// Designed for embedding in terminal dashboards and debug panels.
//
// Usage:
//
//	lv := NewLogViewer()
//	lv.AddEntry(LogInfo, "Server started on :8080")
//	lv.AddEntry(LogWarn, "High memory usage detected")
//	lv.AddEntry(LogError, "Connection refused")
//	lv.Paint(buf)

// LVLevel represents log severity for LogViewer.
type LVLevel int

const (
	LVDebug LVLevel = 0
	LVInfo  LVLevel = 1
	LVWarn  LVLevel = 2
	LVError LVLevel = 3
)

var logLevelLabels = [...]string{"DBG", "INF", "WRN", "ERR"}
var logLevelIcons = [...]rune{'·', 'ℹ', '⚠', '✗'}

// LogViewerStyle holds styling.
type LogViewerStyle struct {
	Debug buffer.Style
	Info  buffer.Style
	Warn  buffer.Style
	Error buffer.Style
	Text  buffer.Style
	Time  buffer.Style
}

// DefaultLogViewerStyle returns defaults.
func DefaultLogViewerStyle() LogViewerStyle {
	return LogViewerStyle{
		Debug: buffer.Style{Fg: buffer.RGB(100, 116, 139)},
		Info:  buffer.Style{Fg: buffer.RGB(96, 165, 250)},
		Warn:  buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Error: buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Text:  buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		Time:  buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

const logViewerMaxEntries = 50

// logEntry holds a single log entry.
type logEntry struct {
	level   LVLevel
	message string
	seq     int
}

// LogViewer renders a scrollable log viewer.
type LogViewer struct {
	BaseComponent
	mu sync.Mutex

	entries [logViewerMaxEntries]logEntry
	count   int
	head    int
	seq     int
	width   int
	maxRows int
	style   LogViewerStyle
}

// NewLogViewer creates a LogViewer.
func NewLogViewer() *LogViewer {
	lv := &LogViewer{width: 50, maxRows: 10, style: DefaultLogViewerStyle()}
	lv.SetID(GenerateID("logview"))
	return lv
}

// AddEntry adds a log entry with level and message.
func (lv *LogViewer) AddEntry(level LVLevel, message string) *LogViewer {
	lv.mu.Lock()
	lv.entries[lv.head] = logEntry{level: level, message: message, seq: lv.seq}
	lv.head = (lv.head + 1) % logViewerMaxEntries
	if lv.count < logViewerMaxEntries {
		lv.count++
	}
	lv.seq++
	lv.mu.Unlock()
	return lv
}

// Clear removes all entries.
func (lv *LogViewer) Clear() *LogViewer {
	lv.mu.Lock()
	lv.count = 0
	lv.head = 0
	lv.mu.Unlock()
	return lv
}

// Count returns the number of entries.
func (lv *LogViewer) Count() int {
	lv.mu.Lock()
	defer lv.mu.Unlock()
	return lv.count
}

// SetMaxRows sets the maximum visible rows.
func (lv *LogViewer) SetMaxRows(n int) *LogViewer {
	lv.mu.Lock()
	if n < 1 {
		n = 1
	}
	lv.maxRows = n
	lv.mu.Unlock()
	return lv
}

// SetWidth sets the display width.
func (lv *LogViewer) SetWidth(w int) *LogViewer {
	lv.mu.Lock()
	if w < 20 {
		w = 20
	}
	lv.width = w
	lv.mu.Unlock()
	return lv
}

// SetStyle sets custom style.
func (lv *LogViewer) SetStyle(s LogViewerStyle) *LogViewer {
	lv.mu.Lock()
	lv.style = s
	lv.mu.Unlock()
	return lv
}

// Measure returns preferred size.
func (lv *LogViewer) Measure(cs Constraints) Size {
	lv.mu.Lock()
	h := lv.maxRows
	if h > lv.count {
		h = lv.count
	}
	lv.mu.Unlock()
	if h < 1 {
		h = 1
	}
	w := lv.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the log viewer.
func (lv *LogViewer) Paint(buf *buffer.Buffer) {
	lv.mu.Lock()
	defer lv.mu.Unlock()

	b := lv.Bounds()
	x, y := b.X, b.Y

	if lv.count == 0 {
		return
	}

	// Show the most recent entries (up to maxRows)
	nShow := lv.maxRows
	if nShow > lv.count {
		nShow = lv.count
	}

	for i := 0; i < nShow; i++ {
		idx := (lv.head - nShow + i + logViewerMaxEntries) % logViewerMaxEntries
		entry := lv.entries[idx]
		yy := y + i
		if yy >= buf.Height {
			break
		}

		var levelStyle buffer.Style
		switch entry.level {
		case LVDebug:
			levelStyle = lv.style.Debug
		case LVInfo:
			levelStyle = lv.style.Info
		case LVWarn:
			levelStyle = lv.style.Warn
		case LVError:
			levelStyle = lv.style.Error
		}
		textStyle := lv.style.Text

		col := x

		// Level icon
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: logLevelIcons[entry.level], Fg: levelStyle.Fg, Bg: levelStyle.Bg, Flags: levelStyle.Flags, Width: 1})
			col++
		}
		// Level label
		for _, r := range logLevelLabels[entry.level] {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: levelStyle.Fg, Bg: levelStyle.Bg, Flags: levelStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}

		// Message text
		for _, r := range entry.message {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (lv *LogViewer) Children() []Component { return nil }
