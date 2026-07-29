package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── NotificationStack: Stacked Toast/Notification Cards ───
//
// NotificationStack displays a vertical stack of notification cards, each with
// title, message, severity (info/warn/error/success), and distinct coloring.
// Common in chat apps, build tools, and system dashboards.
//
// Usage:
//
//	ns := NewNotificationStack()
//	ns.AddNotification("Build", "Compiled successfully", NotifSuccess)
//	ns.AddNotification("Warning", "Deprecation in main.go", NotifWarning)
//	ns.Paint(buf)

// NotifSeverity describes notification severity level.
type NotifSeverity int

const (
	NotifInfo NotifSeverity = iota
	NotifWarning
	NotifError
	NotifSuccess
)

// NotificationEntry represents a single notification in the stack.
type NotificationEntry struct {
	Title    string
	Message  string
	Severity NotifSeverity
}

// NotificationStackStyle holds styling for NotificationStack.
type NotificationStackStyle struct {
	Title    [4]buffer.Style // [info, warning, error, success]
	Message  [4]buffer.Style
	Icon     [4]buffer.Style
	Border   [4]buffer.Style
}

// DefaultNotificationStackStyle returns sensible defaults.
func DefaultNotificationStackStyle() NotificationStackStyle {
	info := buffer.Style{Fg: buffer.RGB(96, 165, 250)}     // blue-400
	warn := buffer.Style{Fg: buffer.RGB(234, 179, 8)}      // yellow-500
	errS := buffer.Style{Fg: buffer.RGB(239, 68, 68)}      // red-500
	succ := buffer.Style{Fg: buffer.RGB(34, 197, 94)}      // green-500
	msg := buffer.Style{Fg: buffer.RGB(203, 213, 225)}     // slate-300
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}    // slate-600
	return NotificationStackStyle{
		Title:   [4]buffer.Style{{Fg: info.Fg, Flags: buffer.Bold}, {Fg: warn.Fg, Flags: buffer.Bold}, {Fg: errS.Fg, Flags: buffer.Bold}, {Fg: succ.Fg, Flags: buffer.Bold}},
		Message: [4]buffer.Style{msg, msg, msg, msg},
		Icon:    [4]buffer.Style{info, warn, errS, succ},
		Border:  [4]buffer.Style{border, border, border, border},
	}
}

// notifIcon returns the icon rune for a severity.
func notifIcon(s NotifSeverity) rune {
	switch s {
	case NotifInfo:
		return 'ℹ'
	case NotifWarning:
		return '⚠'
	case NotifError:
		return '✗'
	case NotifSuccess:
		return '✓'
	default:
		return '·'
	}
}

// NotificationStack displays stacked notification cards.
type NotificationStack struct {
	BaseComponent
	mu sync.Mutex

	entries []NotificationEntry
	style   NotificationStackStyle
}

// NewNotificationStack creates a NotificationStack with defaults.
func NewNotificationStack() *NotificationStack {
	ns := &NotificationStack{
		style: DefaultNotificationStackStyle(),
	}
	ns.SetID(GenerateID("notifstack"))
	return ns
}

// AddNotification adds a notification to the stack.
func (ns *NotificationStack) AddNotification(title, message string, severity NotifSeverity) *NotificationStack {
	ns.mu.Lock()
	ns.entries = append(ns.entries, NotificationEntry{Title: title, Message: message, Severity: severity})
	ns.mu.Unlock()
	return ns
}

// Dismiss removes a notification by index.
func (ns *NotificationStack) Dismiss(index int) *NotificationStack {
	ns.mu.Lock()
	if index >= 0 && index < len(ns.entries) {
		ns.entries = append(ns.entries[:index], ns.entries[index+1:]...)
	}
	ns.mu.Unlock()
	return ns
}

// Count returns the number of notifications.
func (ns *NotificationStack) Count() int {
	ns.mu.Lock()
	defer ns.mu.Unlock()
	return len(ns.entries)
}

// Clear removes all notifications.
func (ns *NotificationStack) Clear() *NotificationStack {
	ns.mu.Lock()
	ns.entries = ns.entries[:0]
	ns.mu.Unlock()
	return ns
}

// SetStyle sets the custom style.
func (ns *NotificationStack) SetStyle(s NotificationStackStyle) *NotificationStack {
	ns.mu.Lock()
	ns.style = s
	ns.mu.Unlock()
	return ns
}

// Measure returns the preferred size.
func (ns *NotificationStack) Measure(cs Constraints) Size {
	ns.mu.Lock()
	count := len(ns.entries)
	ns.mu.Unlock()

	w := 50
	h := count*3 + 1 // each card: title line + message line + gap
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

// Paint renders the notification stack into the buffer.
func (ns *NotificationStack) Paint(buf *buffer.Buffer) {
	ns.mu.Lock()
	defer ns.mu.Unlock()

	b := ns.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 20 {
		w = 50
	}

	for idx, entry := range ns.entries {
		cardY := y + idx*3
		if cardY+2 >= buf.Height {
			break
		}

		sev := int(entry.Severity)
		if sev < 0 || sev > 3 {
			sev = 0
		}
		borderStyle := ns.style.Border[sev]
		titleStyle := ns.style.Title[sev]
		iconStyle := ns.style.Icon[sev]
		msgStyle := ns.style.Message[sev]

		// Draw top border
		for col := 0; col < w && x+col < buf.Width; col++ {
			buf.SetCell(x+col, cardY, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
		}

		// Title row with icon
		col := x
		icon := notifIcon(entry.Severity)
		if col < buf.Width {
			buf.SetCell(col, cardY+1, buffer.Cell{Rune: icon, Fg: iconStyle.Fg, Bg: iconStyle.Bg, Flags: iconStyle.Flags, Width: 1})
		}
		col++
		if col < buf.Width {
			buf.SetCell(col, cardY+1, buffer.Cell{Rune: ' ', Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
		}
		col++
		for _, r := range entry.Title {
			if col >= x+w || col >= buf.Width {
				break
			}
			buf.SetCell(col, cardY+1, buffer.Cell{Rune: r, Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
			col++
		}

		// Message row
		col = x + 3
		for _, r := range entry.Message {
			if col >= x+w || col >= buf.Width {
				break
			}
			buf.SetCell(col, cardY+2, buffer.Cell{Rune: r, Fg: msgStyle.Fg, Bg: msgStyle.Bg, Flags: msgStyle.Flags, Width: 1})
			col++
		}
	}
}

// Children returns nil.
func (ns *NotificationStack) Children() []Component { return nil }
