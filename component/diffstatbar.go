package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// DiffStatBarStyle controls the visual style of a DiffStatBar.
type DiffStatBarStyle int

const (
	// DiffStatStyleBar renders a proportional bar: ████████░░ +12 -5
	DiffStatStyleBar DiffStatBarStyle = iota
	// DiffStatStyleText renders compact text: +12 -5 (5 files)
	DiffStatStyleText
	// DiffStatStyleFull renders bar + text: ████████░░ +12 -5 (3 files)
	DiffStatStyleFull
)

// DiffStatBar renders a compact diff statistics bar (like GitHub's
// "+12 -5" green/red indicator). Essential for AI code review UIs
// where the assistant proposes changes and the user needs a quick
// visual summary of additions vs deletions.
//
// Thread-safe.
type DiffStatBar struct {
	BaseComponent
	mu sync.RWMutex

	additions  int
	deletions  int
	files      int
	barWidth   int
	style      DiffStatBarStyle
	customAdd  buffer.Color // zero = auto (theme.Success)
	customDel  buffer.Color // zero = auto (theme.Error)
}

// NewDiffStatBar creates a diff stat bar with default settings.
func NewDiffStatBar(additions, deletions int) *DiffStatBar {
	return &DiffStatBar{
		BaseComponent: BaseComponent{id: GenerateID("diffstat")},
		additions:     additions,
		deletions:     deletions,
		barWidth:      10,
		style:         DiffStatStyleFull,
	}
}

// SetStats updates the diff statistics.
func (d *DiffStatBar) SetStats(additions, deletions, files int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.additions = additions
	d.deletions = deletions
	d.files = files
}

// Additions returns the additions count.
func (d *DiffStatBar) Additions() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.additions
}

// Deletions returns the deletions count.
func (d *DiffStatBar) Deletions() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.deletions
}

// Files returns the changed files count.
func (d *DiffStatBar) Files() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.files
}

// SetBarWidth sets the proportional bar width (default 10).
func (d *DiffStatBar) SetBarWidth(w int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if w < 1 {
		w = 1
	}
	d.barWidth = w
}

// BarWidth returns the current bar width.
func (d *DiffStatBar) BarWidth() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.barWidth
}

// SetStyle sets the display style (bar, text, or full).
func (d *DiffStatBar) SetStyle(s DiffStatBarStyle) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.style = s
}

// Style returns the current display style.
func (d *DiffStatBar) Style() DiffStatBarStyle {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.style
}

// SetColors overrides the addition/deletion colors.
// Pass buffer.Color{} (zero) for either to revert to theme defaults.
func (d *DiffStatBar) SetColors(add, del buffer.Color) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.customAdd = add
	d.customDel = del
}

// resolveAddColor returns the additions color.
func (d *DiffStatBar) resolveAddColorLocked() buffer.Color {
	if d.customAdd.Type != buffer.ColorNone {
		return d.customAdd
	}
	return theme.Get().Success
}

// resolveDelColor returns the deletions color.
func (d *DiffStatBar) resolveDelColorLocked() buffer.Color {
	if d.customDel.Type != buffer.ColorNone {
		return d.customDel
	}
	return theme.Get().Error
}

// Measure returns the preferred size (always 1 row).
func (d *DiffStatBar) Measure(cs Constraints) Size {
	d.mu.RLock()
	defer d.mu.RUnlock()

	w := d.measureWidthLocked()
	h := 1

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

func (d *DiffStatBar) measureWidthLocked() int {
	w := 0
	switch d.style {
	case DiffStatStyleBar:
		w = d.barWidth + 1 // bar + space
		w += d.textWidthLocked()
	case DiffStatStyleText:
		w = d.textWidthLocked()
	case DiffStatStyleFull:
		w = d.barWidth + 1 // bar + space
		w += d.textWidthLocked()
		if d.files > 0 {
			// " (N files)" or " (N file)"
			w += 3 + numDigits(d.files) + 6
			if d.files == 1 {
				w-- // "file" not "files"
			}
		}
	}
	return w
}

// textWidthLocked returns the width of "+N -M" text.
func (d *DiffStatBar) textWidthLocked() int {
	w := 1 + numDigits(d.additions) // "+N"
	w += 1 + 1 + numDigits(d.deletions) // " -M"
	return w
}

// Paint draws the diff stat bar. Zero allocations.
func (d *DiffStatBar) Paint(buf *buffer.Buffer) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	bounds := d.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	addColor := d.resolveAddColorLocked()
	delColor := d.resolveDelColorLocked()

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	switch d.style {
	case DiffStatStyleText:
		x = d.drawTextLocked(buf, x, y, maxX, addColor, delColor)

	case DiffStatStyleBar:
		x = d.drawBarLocked(buf, x, y, maxX, addColor, delColor)
		if x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
			x++
		}
		x = d.drawTextLocked(buf, x, y, maxX, addColor, delColor)

	case DiffStatStyleFull:
		x = d.drawBarLocked(buf, x, y, maxX, addColor, delColor)
		if x < maxX {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
			x++
		}
		x = d.drawTextLocked(buf, x, y, maxX, addColor, delColor)
		// Append file count if space allows
		if d.files > 0 && x < maxX {
			var fb [32]byte
			fbs := fb[:0]
			fbs = append(fbs, " ("...)
			fbs = strconv.AppendInt(fbs, int64(d.files), 10)
			if d.files == 1 {
				fbs = append(fbs, " file)"...)
			} else {
				fbs = append(fbs, " files)"...)
			}
			muted := buffer.Style{Fg: theme.Get().Muted}
			x = buf.DrawBytes(x, y, fbs, muted)
		}
	}
}

// drawBarLocked draws the proportional bar (green additions / red deletions).
func (d *DiffStatBar) drawBarLocked(buf *buffer.Buffer, x, y, maxX int, addColor, delColor buffer.Color) int {
	total := d.additions + d.deletions
	if total == 0 {
		// All neutral (empty diff)
		for i := 0; i < d.barWidth && x < maxX; i++ {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2591', Width: 1, Fg: theme.Get().Muted})
			x++
		}
		return x
	}

	addCells := d.barWidth * d.additions / total
	if d.additions > 0 && addCells == 0 {
		addCells = 1 // ensure at least 1 green cell if there are additions
	}

	for i := 0; i < d.barWidth && x < maxX; i++ {
		if i < addCells {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2593', Width: 1, Fg: addColor})
		} else {
			buf.SetCell(x, y, buffer.Cell{Rune: '\u2593', Width: 1, Fg: delColor})
		}
		x++
	}
	return x
}

// drawTextLocked draws "+N -M" using stack buffers.
func (d *DiffStatBar) drawTextLocked(buf *buffer.Buffer, x, y, maxX int, addColor, delColor buffer.Color) int {
	addStyle := buffer.Style{Fg: addColor, Flags: buffer.Bold}
	delStyle := buffer.Style{Fg: delColor, Flags: buffer.Bold}

	// "+N"
	var ab [20]byte
	abs := ab[:0]
	abs = append(abs, '+')
	abs = strconv.AppendInt(abs, int64(d.additions), 10)
	x = buf.DrawBytes(x, y, abs, addStyle)

	// " -M"
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1})
		x++
	}
	if x < maxX {
		var db [20]byte
		dbs := db[:0]
		dbs = append(dbs, '-')
		dbs = strconv.AppendInt(dbs, int64(d.deletions), 10)
		x = buf.DrawBytes(x, y, dbs, delStyle)
	}

	return x
}

// numDigits returns the number of decimal digits in n (minimum 1).
func numDigits(n int) int {
	if n < 0 {
		n = -n
	}
	if n < 10 {
		return 1
	}
	digits := 0
	for n > 0 {
		digits++
		n /= 10
	}
	return digits
}
