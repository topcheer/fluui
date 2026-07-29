package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ContextTrimmer: Context Window Trimming Visualizer ───
//
// ContextTrimmer renders a visual showing how much of the conversation
// context will be trimmed when the context window overflows. Shows
// keep/retrim/discard zones.
//
// Usage:
//
//	ct := NewContextTrimmer()
//	ct.SetSegments(2000, 6000, 2000) // keep=2000, retrim=6000, discard=2000
//	ct.Paint(buf)

// ContextTrimmerStyle holds styling.
type ContextTrimmerStyle struct {
	Keep    buffer.Style
	Retrim  buffer.Style
	Discard buffer.Style
	Label   buffer.Style
}

// DefaultContextTrimmerStyle returns defaults.
func DefaultContextTrimmerStyle() ContextTrimmerStyle {
	return ContextTrimmerStyle{
		Keep:    buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Retrim:  buffer.Style{Fg: buffer.RGB(245, 158, 11)},
		Discard: buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
	}
}

// ContextTrimmer renders a context trimming visualization.
type ContextTrimmer struct {
	BaseComponent
	mu sync.Mutex

	keep    int
	retrim  int
	discard int
	width   int
	style   ContextTrimmerStyle
	// cached
	keepStr    string
	retrimStr  string
	discardStr string
	barKeep    int
	barRetrim  int
	barDiscard int
}

// NewContextTrimmer creates a ContextTrimmer.
func NewContextTrimmer() *ContextTrimmer {
	ct := &ContextTrimmer{width: 36, style: DefaultContextTrimmerStyle()}
	ct.SetID(GenerateID("ctxtrim"))
	ct.recomputeLocked()
	return ct
}

// SetSegments sets keep, retrim, and discard token counts.
func (ct *ContextTrimmer) SetSegments(keep, retrim, discard int) *ContextTrimmer {
	ct.mu.Lock()
	if keep < 0 { keep = 0 }
	if retrim < 0 { retrim = 0 }
	if discard < 0 { discard = 0 }
	ct.keep = keep
	ct.retrim = retrim
	ct.discard = discard
	ct.recomputeLocked()
	ct.mu.Unlock()
	return ct
}

func (ct *ContextTrimmer) recomputeLocked() {
	total := ct.keep + ct.retrim + ct.discard
	ct.keepStr = itoa(ct.keep)
	ct.retrimStr = itoa(ct.retrim)
	ct.discardStr = itoa(ct.discard)

	if total == 0 {
		ct.barKeep = 0
		ct.barRetrim = 0
		ct.barDiscard = 0
		return
	}
	const barW = 30
	ct.barKeep = ct.keep * barW / total
	ct.barRetrim = ct.retrim * barW / total
	ct.barDiscard = barW - ct.barKeep - ct.barRetrim
	if ct.barDiscard < 0 { ct.barDiscard = 0 }
}

// TotalTokens returns total tokens across all segments.
func (ct *ContextTrimmer) TotalTokens() int {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	return ct.keep + ct.retrim + ct.discard
}

// SetWidth sets the display width.
func (ct *ContextTrimmer) SetWidth(w int) *ContextTrimmer {
	ct.mu.Lock()
	if w < 20 { w = 20 }
	ct.width = w
	ct.mu.Unlock()
	return ct
}

// SetStyle sets custom style.
func (ct *ContextTrimmer) SetStyle(s ContextTrimmerStyle) *ContextTrimmer {
	ct.mu.Lock()
	ct.style = s
	ct.mu.Unlock()
	return ct
}

// Measure returns preferred size.
func (ct *ContextTrimmer) Measure(cs Constraints) Size {
	w := ct.width + 10
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 3}
}

// Paint renders the context trimmer.
func (ct *ContextTrimmer) Paint(buf *buffer.Buffer) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	b := ct.Bounds()
	x, y := b.X, b.Y

	keepStyle := ct.style.Keep
	retrimStyle := ct.style.Retrim
	discardStyle := ct.style.Discard
	labelStyle := ct.style.Label

	// Row 0: segmented bar
	col := x
	for i := 0; i < ct.barKeep; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: keepStyle.Fg, Bg: keepStyle.Bg, Flags: keepStyle.Flags, Width: 1})
		col++
	}
	for i := 0; i < ct.barRetrim; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '▓', Fg: retrimStyle.Fg, Bg: retrimStyle.Bg, Flags: retrimStyle.Flags, Width: 1})
		col++
	}
	for i := 0; i < ct.barDiscard; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: discardStyle.Fg, Bg: discardStyle.Bg, Flags: discardStyle.Flags, Width: 1})
		col++
	}

	// Row 1: labels
	col = x
	label1 := "Keep:" + ct.keepStr + " Trim:" + ct.retrimStr + " Drop:" + ct.discardStr
	for _, r := range label1 {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Row 2: legend
	col = x
	legend := "█keep ▓trim ░drop"
	for _, r := range legend {
		if col >= buf.Width { break }
		var st buffer.Style
		switch {
		case r == '█':
			st = keepStyle
		case r == '▓':
			st = retrimStyle
		case r == '░':
			st = discardStyle
		default:
			st = labelStyle
		}
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: st.Fg, Bg: st.Bg, Flags: st.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (ct *ContextTrimmer) Children() []Component { return nil }
