package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownAlert: GitHub-style Alert Callout ───
//
// MarkdownAlert renders a GitHub-style alert block (> [!NOTE], > [!WARNING], etc.)
// with an icon, colored left border, and message text.
//
// Usage:
//
//	a := NewMarkdownAlert()
//	a.SetLevel(AlertWarning)
//	a.SetText("This action cannot be undone.")
//	a.Paint(buf)

// AlertLevel represents the alert severity level.
type AlertLevel int

const (
	AlertNote      AlertLevel = 0
	AlertTip       AlertLevel = 1
	AlertImportant AlertLevel = 2
	AlertWarning   AlertLevel = 3
	AlertCaution   AlertLevel = 4
)

var alertLabels = [...]string{"Note", "Tip", "Important", "Warning", "Caution"}
var alertIcons = [...]rune{'ℹ', '💡', '❗', '⚠', '⛔'}

// MarkdownAlertStyle holds styling.
type MarkdownAlertStyle struct {
	Note      buffer.Style
	Tip       buffer.Style
	Important buffer.Style
	Warning   buffer.Style
	Caution   buffer.Style
	Text      buffer.Style
	Border    buffer.Style
}

// DefaultMarkdownAlertStyle returns defaults.
func DefaultMarkdownAlertStyle() MarkdownAlertStyle {
	return MarkdownAlertStyle{
		Note:      buffer.Style{Fg: buffer.RGB(59, 130, 246), Flags: buffer.Bold},
		Tip:       buffer.Style{Fg: buffer.RGB(34, 197, 94), Flags: buffer.Bold},
		Important: buffer.Style{Fg: buffer.RGB(168, 85, 247), Flags: buffer.Bold},
		Warning:   buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
		Caution:   buffer.Style{Fg: buffer.RGB(239, 68, 68), Flags: buffer.Bold},
		Text:      buffer.Style{Fg: buffer.RGB(203, 213, 225)},
		Border:    buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// MarkdownAlert renders a GitHub-style alert callout.
type MarkdownAlert struct {
	BaseComponent
	mu sync.Mutex

	level AlertLevel
	text  string
	width int
	style MarkdownAlertStyle
	// cached
	headerStr   string
	textRunes   []rune
	curStyle    buffer.Style
	borderStyle buffer.Style
}

// NewMarkdownAlert creates a MarkdownAlert.
func NewMarkdownAlert() *MarkdownAlert {
	a := &MarkdownAlert{level: AlertNote, width: 40, style: DefaultMarkdownAlertStyle()}
	a.SetID(GenerateID("mdalert"))
	a.recomputeLocked()
	return a
}

// SetLevel sets the alert severity level.
func (a *MarkdownAlert) SetLevel(l AlertLevel) *MarkdownAlert {
	a.mu.Lock()
	if int(l) < 0 || int(l) >= len(alertLabels) {
		l = AlertNote
	}
	a.level = l
	a.recomputeLocked()
	a.mu.Unlock()
	return a
}

// SetText sets the alert message text.
func (a *MarkdownAlert) SetText(s string) *MarkdownAlert {
	a.mu.Lock()
	a.text = s
	a.recomputeLocked()
	a.mu.Unlock()
	return a
}

func (a *MarkdownAlert) recomputeLocked() {
	a.headerStr = alertLabels[a.level]
	a.textRunes = []rune(a.text)

	switch a.level {
	case AlertNote:
		a.curStyle = a.style.Note
	case AlertTip:
		a.curStyle = a.style.Tip
	case AlertImportant:
		a.curStyle = a.style.Important
	case AlertWarning:
		a.curStyle = a.style.Warning
	case AlertCaution:
		a.curStyle = a.style.Caution
	}

	// Border color matches level
	a.borderStyle.Fg = a.curStyle.Fg
	a.borderStyle.Bg = a.curStyle.Bg
	a.borderStyle.Flags = 0
}

// Level returns the current alert level.
func (a *MarkdownAlert) Level() AlertLevel {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.level
}

// SetWidth sets the alert width.
func (a *MarkdownAlert) SetWidth(w int) *MarkdownAlert {
	a.mu.Lock()
	if w < 10 {
		w = 10
	}
	a.width = w
	a.mu.Unlock()
	return a
}

// SetStyle sets custom style.
func (a *MarkdownAlert) SetStyle(s MarkdownAlertStyle) *MarkdownAlert {
	a.mu.Lock()
	a.style = s
	a.recomputeLocked()
	a.mu.Unlock()
	return a
}

// Measure returns preferred size.
func (a *MarkdownAlert) Measure(cs Constraints) Size {
	w := a.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 2} // header + text
}

// Paint renders the alert callout.
func (a *MarkdownAlert) Paint(buf *buffer.Buffer) {
	a.mu.Lock()
	defer a.mu.Unlock()

	b := a.Bounds()
	x, y := b.X, b.Y
	w := a.width

	headerStyle := a.curStyle
	textStyle := a.style.Text
	borderStyle := a.borderStyle

	// Left border on both rows
	if x < buf.Width {
		buf.SetCell(x, y, buffer.Cell{Rune: '▌', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}
	if x < buf.Width {
		buf.SetCell(x, y+1, buffer.Cell{Rune: '▌', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}

	// Row 0: icon + header label
	col := x + 1
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: alertIcons[a.level], Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
		col++
	}
	for _, r := range a.headerStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: headerStyle.Fg, Bg: headerStyle.Bg, Flags: headerStyle.Flags, Width: 1})
		col++
	}

	// Row 1: message text (wrapped/truncated to width-2)
	col = x + 1
	maxW := w - 2
	for i, r := range a.textRunes {
		if i >= maxW {
			break
		}
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: textStyle.Fg, Bg: textStyle.Bg, Flags: textStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (a *MarkdownAlert) Children() []Component { return nil }
