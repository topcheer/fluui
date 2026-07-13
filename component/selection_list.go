package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── SelectionList: Multi-Select List with Toggle ───
//
// SelectionList is a multi-selection list where users toggle items on/off.
// Inspired by Textual's SelectionList.
//
// Usage:
//	sl := NewSelectionList([]string{"Apple", "Banana", "Cherry"})
//	sl.Toggle(1)               // toggle Banana
//	sl.IsSelected(0)           // check Apple
//	sl.SelectedItems()         // [0, 2] (indices of selected)

// SelectionList is a multi-select list widget.
type SelectionList struct {
	mu        sync.RWMutex
	BaseComponent
	items     []SelectionItem
	cursor    int
	scrollY   int
	onChange  func()
	style     SelectionListStyle
}

// SelectionItem holds a single selectable item.
type SelectionItem struct {
	Label    string
	Selected bool
	Disabled bool
}

// SelectionListStyle holds colors.
type SelectionListStyle struct {
	Fg        buffer.Color
	Bg        buffer.Color
	SelectedFg buffer.Color
	CursorBg  buffer.Color
	DisabledFg buffer.Color
}

func defaultSelectionListStyle() SelectionListStyle {
	return SelectionListStyle{
		Fg:         buffer.NamedColor(buffer.NamedWhite),
		Bg:         buffer.Color{Type: buffer.ColorNone},
		SelectedFg: buffer.NamedColor(buffer.NamedGreen),
		CursorBg:   buffer.RGB(68, 71, 90),
		DisabledFg: buffer.RGB(98, 114, 164),
	}
}

// NewSelectionList creates a selection list from string labels.
func NewSelectionList(labels []string) *SelectionList {
	items := make([]SelectionItem, len(labels))
	for i, l := range labels {
		items[i] = SelectionItem{Label: l}
	}
	return &SelectionList{
		items: items,
		style: defaultSelectionListStyle(),
	}
}

// Toggle flips the selection state of item at index.
func (s *SelectionList) Toggle(idx int) {
	s.mu.Lock()
	if idx >= 0 && idx < len(s.items) && !s.items[idx].Disabled {
		s.items[idx].Selected = !s.items[idx].Selected
	}
	s.mu.Unlock()
	s.notifyChange()
}

// IsSelected returns whether item at index is selected.
func (s *SelectionList) IsSelected(idx int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if idx < 0 || idx >= len(s.items) {
		return false
	}
	return s.items[idx].Selected
}

// SelectedItems returns indices of all selected items.
func (s *SelectionList) SelectedItems() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []int
	for i, item := range s.items {
		if item.Selected {
			result = append(result, i)
		}
	}
	return result
}

// SelectedLabels returns labels of all selected items.
func (s *SelectionList) SelectedLabels() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []string
	for _, item := range s.items {
		if item.Selected {
			result = append(result, item.Label)
		}
	}
	return result
}

// SelectAll selects all non-disabled items.
func (s *SelectionList) SelectAll() {
	s.mu.Lock()
	for i := range s.items {
		if !s.items[i].Disabled {
			s.items[i].Selected = true
		}
	}
	s.mu.Unlock()
	s.notifyChange()
}

// DeselectAll deselects all items.
func (s *SelectionList) DeselectAll() {
	s.mu.Lock()
	for i := range s.items {
		s.items[i].Selected = false
	}
	s.mu.Unlock()
	s.notifyChange()
}

// SetItems replaces all items.
func (s *SelectionList) SetItems(items []SelectionItem) {
	s.mu.Lock()
	s.items = items
	s.cursor = 0
	s.scrollY = 0
	s.mu.Unlock()
	s.notifyChange()
}

// Items returns a copy of all items.
func (s *SelectionList) Items() []SelectionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]SelectionItem, len(s.items))
	copy(result, s.items)
	return result
}

// Cursor returns the cursor index.
func (s *SelectionList) Cursor() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cursor
}

// SetCursor sets the cursor index (clamped).
func (s *SelectionList) SetCursor(idx int) {
	s.mu.Lock()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(s.items) {
		idx = len(s.items) - 1
	}
	s.cursor = idx
	s.mu.Unlock()
}

// ItemCount returns the number of items.
func (s *SelectionList) ItemCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// SetDisabled sets the disabled state of an item.
func (s *SelectionList) SetDisabled(idx int, disabled bool) {
	s.mu.Lock()
	if idx >= 0 && idx < len(s.items) {
		s.items[idx].Disabled = disabled
	}
	s.mu.Unlock()
}

// SetOnChange sets callback.
func (s *SelectionList) SetOnChange(fn func()) {
	s.mu.Lock()
	s.onChange = fn
	s.mu.Unlock()
}

func (s *SelectionList) notifyChange() {
	s.mu.RLock()
	cb := s.onChange
	s.mu.RUnlock()
	if cb != nil {
		cb()
	}
}

// HandleKey processes keyboard input.
func (s *SelectionList) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}

	switch ev.Key {
	case term.KeyUp:
		s.moveCursor(-1)
		return true
	case term.KeyDown:
		s.moveCursor(1)
		return true
	case term.KeyEnter, term.KeySpace:
		s.Toggle(s.Cursor())
		return true
	}

	if ev.Rune == 'j' {
		s.moveCursor(1)
		return true
	}
	if ev.Rune == 'k' {
		s.moveCursor(-1)
		return true
	}
	if ev.Rune == ' ' {
		s.Toggle(s.Cursor())
		return true
	}
	if ev.Rune == 'a' {
		s.SelectAll()
		return true
	}
	if ev.Rune == 'd' {
		s.DeselectAll()
		return true
	}

	return false
}

func (s *SelectionList) moveCursor(delta int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cursor += delta
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(s.items) {
		s.cursor = len(s.items) - 1
	}
}

func (s *SelectionList) Measure(constraints Constraints) Size {
	s.mu.RLock()
	defer s.mu.RUnlock()
	maxW := 0
	for _, item := range s.items {
		w := len([]rune(item.Label)) + 4 // checkbox + space + label
		if w > maxW {
			maxW = w
		}
	}
	if maxW > constraints.MaxWidth && constraints.MaxWidth > 0 {
		maxW = constraints.MaxWidth
	}
	return Size{W: maxW, H: len(s.items)}
}

func (s *SelectionList) Paint(buf *buffer.Buffer) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bounds := s.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	for i, item := range s.items {
		y := bounds.Y + i
		if y >= bounds.Y+bounds.H {
			break
		}

		isCursor := i == s.cursor
		checkBox := "[ ] "
		fg := s.style.Fg
		if item.Selected {
			checkBox = "[x] "
			fg = s.style.SelectedFg
		}
		if item.Disabled {
			fg = s.style.DisabledFg
		}

		bg := s.style.Bg
		if isCursor {
			bg = s.style.CursorBg
		}

		// Draw checkbox
		for j, r := range checkBox {
			if j >= bounds.W {
				break
			}
			buf.SetCell(bounds.X+j, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    fg,
				Bg:    bg,
			})
		}

		// Draw label
		label := item.Label
		if len([]rune(label))+4 > bounds.W {
			label = truncateLabel(label, bounds.W-4)
		}
		for j, r := range label {
			if 4+j >= bounds.W {
				break
			}
			buf.SetCell(bounds.X+4+j, y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    fg,
				Bg:    bg,
			})
		}
	}
}

func truncateLabel(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}
	return string(runes[:maxLen-3]) + "..."
}

// ─── LineGauge: Compact Single-Line Progress Bar ───
//
// LineGauge is a thin, compact progress indicator.
// Inspired by Ratatui's LineGauge.
//
// Usage:
//	lg := NewLineGauge()
//	lg.SetPercent(65)
//	lg.SetLabel("Uploading...")

// LineGauge is a compact single-line progress gauge.
type LineGauge struct {
	mu      sync.RWMutex
	BaseComponent
	percent  float64
	label    string
	fg       buffer.Color
	bg       buffer.Color
	fillChar rune
	emptyChar rune
}

// NewLineGauge creates a compact line gauge.
func NewLineGauge() *LineGauge {
	return &LineGauge{
		fg:        buffer.NamedColor(buffer.NamedGreen),
		bg:        buffer.RGB(68, 71, 90),
		fillChar:  '█',
		emptyChar: '░',
	}
}

// SetPercent sets the progress (0-100).
func (g *LineGauge) SetPercent(p float64) {
	g.mu.Lock()
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	g.percent = p
	g.mu.Unlock()
}

// Percent returns the current progress.
func (g *LineGauge) Percent() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.percent
}

// SetLabel sets the text label shown on the gauge.
func (g *LineGauge) SetLabel(label string) {
	g.mu.Lock()
	g.label = label
	g.mu.Unlock()
}

// SetFg sets the filled portion color.
func (g *LineGauge) SetFg(c buffer.Color) {
	g.mu.Lock()
	g.fg = c
	g.mu.Unlock()
}

// SetBg sets the empty portion color.
func (g *LineGauge) SetBg(c buffer.Color) {
	g.mu.Lock()
	g.bg = c
	g.mu.Unlock()
}

func (g *LineGauge) Measure(constraints Constraints) Size {
	return Size{W: constraints.MaxWidth, H: 1}
}

func (g *LineGauge) Paint(buf *buffer.Buffer) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	bounds := g.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	w := bounds.W
	fillW := int(float64(w) * g.percent / 100.0)
	if fillW > w {
		fillW = w
	}

	for x := 0; x < w; x++ {
		var r rune
		var fg, bg buffer.Color
		if x < fillW {
			r = g.fillChar
			fg = g.fg
			bg = g.bg
		} else {
			r = g.emptyChar
			fg = g.bg
			bg = g.bg
		}
		buf.SetCell(bounds.X+x, bounds.Y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    fg,
			Bg:    bg,
		})
	}

	// Draw label centered
	if g.label != "" {
		labelX := bounds.X + w/2 - len([]rune(g.label))/2
		if labelX < bounds.X {
			labelX = bounds.X
		}
		for i, r := range g.label {
			if labelX+i >= bounds.X+w {
				break
			}
			// Determine which side this char is on
			fg := g.bg // use bg color for text contrast
			if labelX+i-bounds.X < fillW {
				fg = g.bg
			}
			buf.SetCell(labelX+i, bounds.Y, buffer.Cell{
				Rune:  r,
				Width: 1,
				Fg:    fg,
				Bg:    g.bg,
				Flags:  buffer.Reverse,
			})
		}
	}
}