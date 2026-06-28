package app

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// SelectionPoint represents a screen coordinate within the buffer.
type SelectionPoint struct {
	X, Y int
}

// Selection tracks a text selection region on screen.
// The selection always has Start <= End after normalization.
type Selection struct {
	Start SelectionPoint
	End   SelectionPoint
}

// SelectionManager manages text selection state for mouse drag
// and keyboard selection modes.
//
// Mouse drag: left-click-drag selects a rectangular region.
// Keyboard: Shift+Arrow extends the selection from the anchor point.
// Copy: Ctrl+Shift+C copies the selected text via OSC52 to the
// system clipboard.
type SelectionManager struct {
	mu sync.RWMutex

	// selection is the current normalized selection (Start <= End).
	selection Selection

	// active indicates whether there is an active selection.
	active bool

	// selecting indicates the mouse button is held down (drag in progress).
	selecting bool

	// anchor is the fixed starting point for keyboard selection.
	// When the user presses Shift+Arrow, the anchor is where the
	// selection began and End moves with the arrow keys.
	anchor SelectionPoint

	// cursor is the current keyboard cursor position (for keyboard selection).
	cursor SelectionPoint
}

// NewSelectionManager creates a new SelectionManager with no active selection.
func NewSelectionManager() *SelectionManager {
	return &SelectionManager{}
}

// --- Mouse interaction ---

// StartSelection begins a mouse-drag selection at the given coordinates.
func (sm *SelectionManager) StartSelection(x, y int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.selecting = true
	sm.active = true
	sm.selection.Start = SelectionPoint{X: x, Y: y}
	sm.selection.End = SelectionPoint{X: x, Y: y}
	sm.anchor = SelectionPoint{X: x, Y: y}
	sm.cursor = SelectionPoint{X: x, Y: y}
}

// ExtendSelection extends the current selection to the given coordinates.
// Called during mouse drag.
func (sm *SelectionManager) ExtendSelection(x, y int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.selecting {
		return
	}

	sm.selection.End = SelectionPoint{X: x, Y: y}
	sm.cursor = SelectionPoint{X: x, Y: y}
}

// EndSelection finalizes the mouse-drag selection.
// If Start == End, the selection is cleared (it was just a click).
func (sm *SelectionManager) EndSelection() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.selecting = false

	// If start == end, it was a click, not a drag — clear selection.
	sel := sm.normalizedLocked()
	if sel.Start.X == sel.End.X && sel.Start.Y == sel.End.Y {
		sm.active = false
	}
}

// --- Keyboard selection ---

// StartKeyboardSelection begins keyboard-based selection from the given
// cursor position. Subsequent ExtendKeyboardSelection calls move the
// selection endpoint.
func (sm *SelectionManager) StartKeyboardSelection(x, y int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.active = true
	sm.selecting = false
	sm.anchor = SelectionPoint{X: x, Y: y}
	sm.cursor = SelectionPoint{X: x, Y: y}
	sm.selection.Start = SelectionPoint{X: x, Y: y}
	sm.selection.End = SelectionPoint{X: x, Y: y}
}

// ExtendKeyboardSelection moves the selection endpoint by (dx, dy).
// The anchor remains fixed; only the cursor (End) moves.
// bufWidth and bufHeight are used to clamp the cursor position.
func (sm *SelectionManager) ExtendKeyboardSelection(dx, dy int, bufWidth, bufHeight int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.active {
		return
	}

	newX := sm.cursor.X + dx
	newY := sm.cursor.Y + dy

	// Clamp to buffer bounds.
	if newX < 0 {
		newX = 0
	}
	if newX >= bufWidth {
		newX = bufWidth - 1
	}
	if newY < 0 {
		newY = 0
	}
	if newY >= bufHeight {
		newY = bufHeight - 1
	}

	sm.cursor = SelectionPoint{X: newX, Y: newY}
	sm.selection.End = SelectionPoint{X: newX, Y: newY}
}

// --- State queries ---

// HasSelection returns true if there is an active non-empty selection.
func (sm *SelectionManager) HasSelection() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.active
}

// IsSelecting returns true if a mouse drag is in progress.
func (sm *SelectionManager) IsSelecting() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.selecting
}

// SelectionRange returns the normalized selection range (Start <= End).
// Returns the range and true if there is an active selection.
func (sm *SelectionManager) SelectionRange() (start, end SelectionPoint, ok bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.active {
		return SelectionPoint{}, SelectionPoint{}, false
	}

	normalized := sm.normalizedLocked()
	return normalized.Start, normalized.End, true
}

// Clear clears the current selection.
func (sm *SelectionManager) Clear() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.active = false
	sm.selecting = false
}

// --- Text extraction ---

// GetSelectedText extracts the text content from the buffer within the
// current selection region. Returns "" if no selection is active.
//
// For selections spanning multiple rows, rows are joined with newlines.
// Within a single row, trailing spaces are trimmed.
func (sm *SelectionManager) GetSelectedText(buf *buffer.Buffer) string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if !sm.active || buf == nil {
		return ""
	}

	sel := sm.normalizedLocked()

	var sb strings.Builder
	for y := sel.Start.Y; y <= sel.End.Y; y++ {
		if y < 0 || y >= buf.Height {
			continue
		}

		startX := 0
		endX := buf.Width - 1

		if y == sel.Start.Y {
			startX = sel.Start.X
		}
		if y == sel.End.Y {
			endX = sel.End.X
		}

		// Clamp to buffer bounds.
		if startX < 0 {
			startX = 0
		}
		if endX >= buf.Width {
			endX = buf.Width - 1
		}
		if startX > endX {
			continue
		}

		// Extract text from this row.
		var rowSB strings.Builder
		for x := startX; x <= endX; x++ {
			cell := buf.GetCell(x, y)
			if cell.Width == 0 {
				continue // skip continuation cells of wide characters
			}
			rowSB.WriteRune(cell.Rune)
		}

		// Trim trailing spaces on each line.
		rowText := strings.TrimRight(rowSB.String(), " ")
		sb.WriteString(rowText)

		// Add newline between rows (but not after the last row).
		if y < sel.End.Y {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// --- OSC52 copy ---

// CopySelection returns an OSC52 escape sequence containing the selected
// text, ready to be written to the terminal. Returns "" if no selection
// is active or the selection has no text.
func (sm *SelectionManager) CopySelection(buf *buffer.Buffer) string {
	text := sm.GetSelectedText(buf)
	if text == "" {
		return ""
	}
	return term.CopyOSC52(text)
}

// CopySelectionSource returns an OSC52 escape sequence with a specific
// clipboard source (system, primary, etc.).
func (sm *SelectionManager) CopySelectionSource(buf *buffer.Buffer, source term.ClipboardSource) string {
	text := sm.GetSelectedText(buf)
	if text == "" {
		return ""
	}
	return term.CopyOSC52Source(text, source)
}

// --- Key handling ---

// HandleKey processes keyboard events for selection mode.
// It handles:
//   - Shift+Arrow keys: extend/reduce keyboard selection
//   - Escape: clear selection
//
// cursorX and cursorY are the current cursor position (for starting
// a new keyboard selection). bufWidth and bufHeight are the buffer
// dimensions for clamping.
//
// Returns true if the key was consumed.
func (sm *SelectionManager) HandleKey(key *term.KeyEvent, cursorX, cursorY, bufWidth, bufHeight int) bool {
	if key == nil {
		return false
	}

	// Escape clears selection.
	if key.Key == term.KeyEscape {
		if sm.HasSelection() {
			sm.Clear()
			return true
		}
		return false
	}

	// Shift+Arrow: keyboard selection.
	if key.Modifiers&term.ModShift != 0 {
		if !sm.HasSelection() {
			sm.StartKeyboardSelection(cursorX, cursorY)
		}

		consumed := true
		switch key.Key {
		case term.KeyUp:
			sm.ExtendKeyboardSelection(0, -1, bufWidth, bufHeight)
		case term.KeyDown:
			sm.ExtendKeyboardSelection(0, 1, bufWidth, bufHeight)
		case term.KeyLeft:
			sm.ExtendKeyboardSelection(-1, 0, bufWidth, bufHeight)
		case term.KeyRight:
			sm.ExtendKeyboardSelection(1, 0, bufWidth, bufHeight)
		default:
			consumed = false
		}
		return consumed
	}

	return false
}

// --- Rendering ---

// ApplyHighlight applies reverse-video highlighting to the selected
// region in the buffer. This should be called after the normal render
// pass to overlay the selection visual.
//
// For each cell in the selection, the foreground and background colors
// are swapped to create the highlight effect.
func (sm *SelectionManager) ApplyHighlight(buf *buffer.Buffer) {
	if buf == nil {
		return
	}

	sm.mu.RLock()
	if !sm.active {
		sm.mu.RUnlock()
		return
	}

	sel := sm.normalizedLocked()
	sm.mu.RUnlock()

	// Iterate over the selection region and swap Fg/Bg for each cell.
	for y := sel.Start.Y; y <= sel.End.Y; y++ {
		if y < 0 || y >= buf.Height {
			continue
		}

		startX := 0
		endX := buf.Width - 1

		if y == sel.Start.Y {
			startX = sel.Start.X
		}
		if y == sel.End.Y {
			endX = sel.End.X
		}

		// Clamp.
		if startX < 0 {
			startX = 0
		}
		if endX >= buf.Width {
			endX = buf.Width - 1
		}

		for x := startX; x <= endX; x++ {
			cell := buf.GetCell(x, y)
			// Swap Fg and Bg for reverse-video effect.
			cell.Fg, cell.Bg = cell.Bg, cell.Fg
			buf.SetCell(x, y, cell)
		}
	}
}

// --- Mouse event integration ---

// HandleMouse processes a mouse event for selection tracking.
// It should be called for every mouse event before other handlers.
//
// - MouseDown (left button): starts selection
// - MouseDrag: extends selection
// - MouseUp: finalizes selection
//
// Returns true if the event was consumed by selection tracking.
func (sm *SelectionManager) HandleMouse(mouse *term.MouseEvent) bool {
	if mouse == nil {
		return false
	}

	// Only left button for selection.
	if mouse.Button != term.MouseLeft && mouse.Action != term.MouseUp {
		return false
	}

	switch mouse.Action {
	case term.MouseDown:
		if mouse.Button == term.MouseLeft {
			sm.StartSelection(mouse.X, mouse.Y)
			return true
		}
	case term.MouseDrag:
		if mouse.Button == term.MouseLeft {
			sm.ExtendSelection(mouse.X, mouse.Y)
			return true
		}
	case term.MouseUp:
		if sm.IsSelecting() {
			sm.EndSelection()
			return true
		}
	}

	return false
}

// --- Helpers ---

// normalizedLocked returns the selection with Start <= End.
// Caller must hold at least RLock.
func (sm *SelectionManager) normalizedLocked() Selection {
	s := sm.selection

	// Normalize so Start is top-left and End is bottom-right.
	if s.Start.Y > s.End.Y || (s.Start.Y == s.End.Y && s.Start.X > s.End.X) {
		s.Start, s.End = s.End, s.Start
	}

	return s
}

// Cursor returns the current cursor position (end of selection).
func (sm *SelectionManager) Cursor() SelectionPoint {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.cursor
}

// Anchor returns the selection anchor point (start of selection).
func (sm *SelectionManager) Anchor() SelectionPoint {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.anchor
}

// Selection returns the raw (unnormalized) selection.
func (sm *SelectionManager) Selection() Selection {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.selection
}
