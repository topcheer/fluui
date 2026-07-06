package component

import (
	"fmt"
	"sort"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ThemeStudio is an interactive theme editor component.
//
// It displays a list of all theme color slots (Background, Foreground, Accent,
// Border, Success, Error, etc.) with their current values. The user can:
//   - Navigate the list with arrow keys / j/k
//   - Press Enter to open the built-in ColorPicker for the selected slot
//   - Change colors and see the effect immediately via OnChange callback
//   - Press 'r' to reset to the original theme
//   - Press 's' to save (via OnSave callback)
//
// This is the only interactive theme editor in any Go TUI library.
// No equivalent exists in Bubble Tea, tview, termui, or Ratatui.
type ThemeStudio struct {
	BaseComponent
	mu sync.RWMutex

	slots     []themeColorSlot
	cursor    int
	pickerOpen bool
	picker    *ColorPicker

	originalTheme *theme.Theme // for reset
	changed       bool

	// Callbacks
	OnChange func() // fired when a color changes
	OnSave   func() // fired when 's' is pressed

	// Styles
	style ThemeStudioStyle
}

// themeColorSlot represents one editable color in the theme.
type themeColorSlot struct {
	Name     string
	Category string
	getter   func(*theme.Theme) buffer.Color
	setter   func(*theme.Theme, buffer.Color)
}

// ThemeStudioStyle holds visual styles for the ThemeStudio.
type ThemeStudioStyle struct {
	Title      buffer.Style
	Category   buffer.Style
	SlotName   buffer.Style
	ColorValue buffer.Style
	Swatch     buffer.Style
	Cursor     buffer.Style
	Help       buffer.Style
}

// DefaultThemeStudioStyle returns a sensible default style.
func DefaultThemeStudioStyle() ThemeStudioStyle {
	return ThemeStudioStyle{
		Title:      buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Category:   buffer.Style{Fg: buffer.Cyan, Flags: buffer.Bold | buffer.Underline},
		SlotName:   buffer.Style{Fg: buffer.White},
		ColorValue: buffer.Style{Fg: buffer.RGB(180, 180, 180)},
		Swatch:     buffer.Style{Fg: buffer.RGB(80, 80, 80)},
		Cursor:     buffer.Style{Fg: buffer.Yellow, Flags: buffer.Bold},
		Help:       buffer.Style{Fg: buffer.RGB(120, 120, 120), Flags: buffer.Dim},
	}
}

// NewThemeStudio creates a ThemeStudio component editing the given theme.
// Pass theme.Get() to edit the current active theme.
func NewThemeStudio(t *theme.Theme) *ThemeStudio {
	ts := &ThemeStudio{
		style:         DefaultThemeStudioStyle(),
		originalTheme: copyTheme(t),
	}
	ts.initSlots(t)
	ts.picker = NewColorPicker()
	ts.picker.OnChange = func(c buffer.Color) {
		ts.mu.Lock()
		defer ts.mu.Unlock()
		if ts.cursor < 0 || ts.cursor >= len(ts.slots) {
			return
		}
		ts.slots[ts.cursor].setter(theme.Get(), c)
		ts.changed = true
		ts.fireChangeLocked()
	}
	return ts
}

// ─── Public API ───

// Cursor returns the current cursor position.
func (ts *ThemeStudio) Cursor() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.cursor
}

// SetCursor sets the cursor position with clamping.
func (ts *ThemeStudio) SetCursor(idx int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.setCursorLocked(idx)
}

// SlotCount returns the number of color slots.
func (ts *ThemeStudio) SlotCount() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return len(ts.slots)
}

// Changed returns true if any color has been modified from the original.
func (ts *ThemeStudio) Changed() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.changed
}

// IsPickerOpen returns true if the ColorPicker overlay is open.
func (ts *ThemeStudio) IsPickerOpen() bool {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.pickerOpen
}

// ClosePicker closes the ColorPicker overlay if open.
func (ts *ThemeStudio) ClosePicker() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.pickerOpen = false
}

// Reset restores all colors to the original theme.
func (ts *ThemeStudio) Reset() {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	theme.SetActive(copyTheme(ts.originalTheme))
	ts.changed = false
	ts.fireChangeLocked()
}

// SetStyle sets the visual style.
func (ts *ThemeStudio) SetStyle(s ThemeStudioStyle) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.style = s
}

// HandleKey processes keyboard input.
func (ts *ThemeStudio) HandleKey(k *term.KeyEvent) bool {
	ts.mu.Lock()

	// If picker is open, route to picker first.
	// IMPORTANT: release mu before calling picker.HandleKey because the
	// picker's OnChange callback re-acquires mu. Holding mu would deadlock.
	if ts.pickerOpen {
		if k.Key == term.KeyEscape || k.Rune == 'q' {
			ts.pickerOpen = false
			ts.mu.Unlock()
			return true
		}
		ts.mu.Unlock()
		return ts.picker.HandleKey(k)
	}

	defer ts.mu.Unlock()

	switch {
	case k.Key == term.KeyUp || k.Rune == 'k':
		ts.setCursorLocked(ts.cursor - 1)
		return true
	case k.Key == term.KeyDown || k.Rune == 'j':
		ts.setCursorLocked(ts.cursor + 1)
		return true
	case k.Key == term.KeyHome || k.Rune == 'g':
		ts.setCursorLocked(0)
		return true
	case k.Key == term.KeyEnd || k.Rune == 'G':
		ts.setCursorLocked(len(ts.slots) - 1)
		return true
	case k.Key == term.KeyEnter:
		ts.openPickerLocked()
		return true
	case k.Rune == 'r':
		theme.SetActive(copyTheme(ts.originalTheme))
		ts.changed = false
		ts.fireChangeLocked()
		return true
	case k.Rune == 's':
		if cb := ts.OnSave; cb != nil {
			cb()
		}
		return true
	case k.Rune == 'q':
		ts.pickerOpen = false
		return true
	}
	return false
}

// Measure returns the desired size.
func (ts *ThemeStudio) Measure(cs Constraints) Size {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	// Width: category label + slot name + hex value + swatch
	// Typical: 20 + 20 + 10 + 4 = ~60 chars
	w := 60
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	h := len(ts.slots) + countCategories(ts.slots) + 2 // +2 for title + help
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the theme editor.
func (ts *ThemeStudio) Paint(buf *buffer.Buffer) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	bounds := ts.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	x, y := bounds.X, bounds.Y
	maxW := bounds.W

	// Title
	drawStringAt(buf, x, y, "Theme Studio", ts.style.Title)
	y++
	if y >= bounds.Y+bounds.H {
		return
	}

	// Slots grouped by category
	currentCategory := ""
	for i, slot := range ts.slots {
		if y >= bounds.Y+bounds.H-1 {
			break // leave room for help
		}
		// Category header
		if slot.Category != currentCategory {
			currentCategory = slot.Category
			if i > 0 {
				y++ // blank line between categories
				if y >= bounds.Y+bounds.H-1 {
					break
				}
			}
			drawStringAt(buf, x, y, slot.Category, ts.style.Category)
			y++
			if y >= bounds.Y+bounds.H-1 {
				break
			}
		}

		// Cursor indicator
		cursor := "  "
		if i == ts.cursor {
			cursor = "▶ "
		}

		// Slot name
		nameStyle := ts.style.SlotName
		if i == ts.cursor {
			nameStyle = ts.style.Cursor
		}
		drawStringAt(buf, x, y, cursor+slot.Name, nameStyle)

		// Color value (hex)
		col := slot.getter(theme.Get())
		hexStr := colorToHex(col)
		valX := x + 25
		if valX < x+maxW {
			drawStringAt(buf, valX, y, hexStr, ts.style.ColorValue)
		}

		// Color swatch (colored block)
		swatchX := x + 40
		if swatchX < x+maxW {
			paintSwatch(buf, swatchX, y, 3, col)
		}

		y++
	}

	// Help line
	if y < bounds.Y+bounds.H {
		help := "↑↓/jk nav · Enter edit · r reset · s save · q close"
		drawStringAt(buf, x, bounds.Y+bounds.H-1, help, ts.style.Help)
	}

	// Render picker overlay if open
	if ts.pickerOpen {
		ts.paintPickerOverlay(buf, bounds)
	}
}

// ─── Internal methods ───

func (ts *ThemeStudio) setCursorLocked(idx int) {
	if len(ts.slots) == 0 {
		ts.cursor = 0
		return
	}
	if idx < 0 {
		idx = len(ts.slots) - 1
	}
	if idx >= len(ts.slots) {
		idx = 0
	}
	ts.cursor = idx
}

func (ts *ThemeStudio) openPickerLocked() {
	if ts.cursor < 0 || ts.cursor >= len(ts.slots) {
		return
	}
	col := ts.slots[ts.cursor].getter(theme.Get())
	// Suppress OnChange during initial SetColor to avoid deadlock
	// (SetColor fires OnChange, but we don't want to apply during init).
	cb := ts.picker.OnChange
	ts.picker.OnChange = nil
	ts.picker.SetColor(col)
	ts.picker.OnChange = cb
	ts.pickerOpen = true
}

func (ts *ThemeStudio) applyCurrentColor(c buffer.Color) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if ts.cursor < 0 || ts.cursor >= len(ts.slots) {
		return
	}
	ts.slots[ts.cursor].setter(theme.Get(), c)
	ts.changed = true
	ts.fireChangeLocked()
}

func (ts *ThemeStudio) fireChangeLocked() {
	if cb := ts.OnChange; cb != nil {
		cb()
	}
}

func (ts *ThemeStudio) paintPickerOverlay(buf *buffer.Buffer, bounds Rect) {
	// Position picker in the right portion of the studio
	pickerBounds := Rect{
		X: bounds.X + bounds.W - 35,
		Y: bounds.Y + 3,
		W: 35,
		H: bounds.H - 5,
	}
	if pickerBounds.W < 30 {
		pickerBounds.W = 30
		pickerBounds.X = bounds.X + bounds.W - 30
	}
	if pickerBounds.H < 10 {
		pickerBounds.H = 10
	}
	ts.picker.SetBounds(pickerBounds)

	// Draw a semi-transparent background for the picker area
	for y := pickerBounds.Y; y < pickerBounds.Y+pickerBounds.H; y++ {
		for x := pickerBounds.X; x < pickerBounds.X+pickerBounds.W; x++ {
			c := buf.GetCell(x, y)
			c.Bg = buffer.RGB(20, 20, 30)
			buf.SetCell(x, y, c)
		}
	}

	ts.picker.Paint(buf)
}

// initSlots builds the list of editable color slots from a theme.
func (ts *ThemeStudio) initSlots(t *theme.Theme) {
	ts.slots = []themeColorSlot{
		// Base
		{"Background", "Base", func(t *theme.Theme) buffer.Color { return t.Bg }, func(t *theme.Theme, c buffer.Color) { t.Bg = c }},
		{"Foreground", "Base", func(t *theme.Theme) buffer.Color { return t.Fg }, func(t *theme.Theme, c buffer.Color) { t.Fg = c }},
		{"Accent", "Base", func(t *theme.Theme) buffer.Color { return t.Accent }, func(t *theme.Theme, c buffer.Color) { t.Accent = c }},

		// Borders
		{"Border", "Borders", func(t *theme.Theme) buffer.Color { return t.Border }, func(t *theme.Theme, c buffer.Color) { t.Border = c }},
		{"Border Active", "Borders", func(t *theme.Theme) buffer.Color { return t.BorderActive }, func(t *theme.Theme, c buffer.Color) { t.BorderActive = c }},
		{"Border Muted", "Borders", func(t *theme.Theme) buffer.Color { return t.BorderMuted }, func(t *theme.Theme, c buffer.Color) { t.BorderMuted = c }},

		// Status
		{"Success", "Status", func(t *theme.Theme) buffer.Color { return t.Success }, func(t *theme.Theme, c buffer.Color) { t.Success = c }},
		{"Error", "Status", func(t *theme.Theme) buffer.Color { return t.Error }, func(t *theme.Theme, c buffer.Color) { t.Error = c }},
		{"Warning", "Status", func(t *theme.Theme) buffer.Color { return t.Warning }, func(t *theme.Theme, c buffer.Color) { t.Warning = c }},
		{"Muted", "Status", func(t *theme.Theme) buffer.Color { return t.Muted }, func(t *theme.Theme, c buffer.Color) { t.Muted = c }},

		// Code
		{"Code Bg", "Code", func(t *theme.Theme) buffer.Color { return t.CodeBg }, func(t *theme.Theme, c buffer.Color) { t.CodeBg = c }},
		{"Code Fg", "Code", func(t *theme.Theme) buffer.Color { return t.CodeFg }, func(t *theme.Theme, c buffer.Color) { t.CodeFg = c }},

		// Diff
		{"Diff Add", "Diff", func(t *theme.Theme) buffer.Color { return t.DiffAdd }, func(t *theme.Theme, c buffer.Color) { t.DiffAdd = c }},
		{"Diff Del", "Diff", func(t *theme.Theme) buffer.Color { return t.DiffDel }, func(t *theme.Theme, c buffer.Color) { t.DiffDel = c }},
		{"Diff Meta", "Diff", func(t *theme.Theme) buffer.Color { return t.DiffMeta }, func(t *theme.Theme, c buffer.Color) { t.DiffMeta = c }},
		{"Diff Hunk", "Diff", func(t *theme.Theme) buffer.Color { return t.DiffHunk }, func(t *theme.Theme, c buffer.Color) { t.DiffHunk = c }},

		// Blocks
		{"User Msg Bg", "Blocks", func(t *theme.Theme) buffer.Color { return t.UserMsgBg }, func(t *theme.Theme, c buffer.Color) { t.UserMsgBg = c }},
		{"User Msg Fg", "Blocks", func(t *theme.Theme) buffer.Color { return t.UserMsgFg }, func(t *theme.Theme, c buffer.Color) { t.UserMsgFg = c }},
		{"Thinking Bg", "Blocks", func(t *theme.Theme) buffer.Color { return t.ThinkingBg }, func(t *theme.Theme, c buffer.Color) { t.ThinkingBg = c }},
		{"Thinking Fg", "Blocks", func(t *theme.Theme) buffer.Color { return t.ThinkingFg }, func(t *theme.Theme, c buffer.Color) { t.ThinkingFg = c }},
		{"Assistant Fg", "Blocks", func(t *theme.Theme) buffer.Color { return t.AssistantFg }, func(t *theme.Theme, c buffer.Color) { t.AssistantFg = c }},

		// Input
		{"Prompt Fg", "Input", func(t *theme.Theme) buffer.Color { return t.PromptFg }, func(t *theme.Theme, c buffer.Color) { t.PromptFg = c }},
		{"Separator", "Input", func(t *theme.Theme) buffer.Color { return t.Separator }, func(t *theme.Theme, c buffer.Color) { t.Separator = c }},

		// Search
		{"Search Bar Bg", "Search", func(t *theme.Theme) buffer.Color { return t.SearchBarBg }, func(t *theme.Theme, c buffer.Color) { t.SearchBarBg = c }},
		{"Search Bar Fg", "Search", func(t *theme.Theme) buffer.Color { return t.SearchBarFg }, func(t *theme.Theme, c buffer.Color) { t.SearchBarFg = c }},
		{"Search Match", "Search", func(t *theme.Theme) buffer.Color { return t.SearchMatch }, func(t *theme.Theme, c buffer.Color) { t.SearchMatch = c }},
	}

	// Sort by category, then name
	sort.SliceStable(ts.slots, func(i, j int) bool {
		if ts.slots[i].Category != ts.slots[j].Category {
			return ts.slots[i].Category < ts.slots[j].Category
		}
		return ts.slots[i].Name < ts.slots[j].Name
	})
}

// ─── Helpers ───

func copyTheme(t *theme.Theme) *theme.Theme {
	if t == nil {
		return nil
	}
	cp := *t
	return &cp
}

func countCategories(slots []themeColorSlot) int {
	cats := map[string]bool{}
	for _, s := range slots {
		cats[s.Category] = true
	}
	return len(cats)
}

func colorToHex(c buffer.Color) string {
	if c.Type == buffer.ColorTrue {
		return fmt.Sprintf("#%02X%02X%02X", c.R(), c.G(), c.B())
	}
	if c.Type == buffer.Color256 {
		return fmt.Sprintf("256:%d", c.Val)
	}
	return "default"
}

func paintSwatch(buf *buffer.Buffer, x, y, w int, c buffer.Color) {
	for i := 0; i < w; i++ {
		cell := buffer.Cell{Rune: ' ', Width: 1, Bg: c}
		buf.SetCell(x+i, y, cell)
	}
}

func drawStringAt(buf *buffer.Buffer, x, y int, s string, style buffer.Style) {
	for i, r := range s {
		buf.SetCell(x+i, y, buffer.Cell{Rune: r, Width: 1, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags})
	}
}
