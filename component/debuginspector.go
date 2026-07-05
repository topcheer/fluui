package component

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// InspectorMode controls what the DebugInspector panel shows.
type InspectorMode int

const (
	// InspectTree shows the component tree hierarchy.
	InspectTree InspectorMode = iota
	// InspectEvents shows a scrolling log of recent events.
	InspectEvents
	// InspectStats shows render and performance metrics.
	InspectStats
)

// EventType identifies the kind of recorded event.
type EventType int

const (
	EventKey EventType = iota
	EventMouse
	EventResize
	EventCustom
)

// EventEntry is a single recorded event for the inspector log.
type EventEntry struct {
	Type      EventType
	Timestamp time.Time
	Summary   string
}

// RenderStats tracks rendering performance metrics.
type RenderStats struct {
	FrameCount    int
	TotalCells    int
	TotalRenderNs int64
	LastRenderNs  int64
	DirtyCount    int
}

// DebugInspector is a runtime debugging overlay that shows the component tree,
// event log, and render statistics. It is toggled on/off at runtime and renders
// as a panel overlay on top of the main content.
//
// Usage:
//
//	inspector := NewDebugInspector()
//	app.OnKey(func(k *term.KeyEvent) {
//	    if k.Key == term.KeyF12 {
//	        inspector.Toggle()
//	    }
//	    inspector.RecordKey(k)
//	})
//	app.OnPaint(func(buf *buffer.Buffer) {
//	    root.Paint(buf)
//	    inspector.Paint(buf) // overlays on top
//	})
type DebugInspector struct {
	BaseComponent
	mu sync.RWMutex

	visible bool
	mode    InspectorMode

	// Event log (ring buffer)
	events    []EventEntry
	maxEvents int
	eventHead int // next write position

	// Render stats
	stats RenderStats

	// Optional root component for tree inspection
	root Component

	// Panel dimensions
	bounds    Rect
	panelW    int
	panelH    int
	panelX    int
	panelY    int

	// Scroll offset for event log
	scrollOffset int

	// Title
	title string
}

// NewDebugInspector creates a new DebugInspector with default settings.
func NewDebugInspector() *DebugInspector {
	di := &DebugInspector{
		visible:   false,
		mode:      InspectTree,
		maxEvents: 200,
		events:    make([]EventEntry, 0, 200),
		panelW:    50,
		panelH:    20,
		title:     "Debug Inspector",
	}
	di.SetID(GenerateID("debug-inspector"))
	return di
}

// Visible returns whether the inspector overlay is shown.
func (di *DebugInspector) Visible() bool {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.visible
}

// Show makes the inspector overlay visible.
func (di *DebugInspector) Show() {
	di.mu.Lock()
	di.visible = true
	di.mu.Unlock()
}

// Hide hides the inspector overlay.
func (di *DebugInspector) Hide() {
	di.mu.Lock()
	di.visible = false
	di.mu.Unlock()
}

// Toggle flips the visibility and returns the new state.
func (di *DebugInspector) Toggle() bool {
	di.mu.Lock()
	di.visible = !di.visible
	newState := di.visible
	di.mu.Unlock()
	return newState
}

// SetVisible directly sets visibility.
func (di *DebugInspector) SetVisible(v bool) {
	di.mu.Lock()
	di.visible = v
	di.mu.Unlock()
}

// Mode returns the current inspector mode.
func (di *DebugInspector) Mode() InspectorMode {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.mode
}

// SetMode sets the inspector display mode.
func (di *DebugInspector) SetMode(m InspectorMode) {
	di.mu.Lock()
	di.mode = m
	di.scrollOffset = 0
	di.mu.Unlock()
}

// NextMode cycles to the next inspector mode.
func (di *DebugInspector) NextMode() {
	di.mu.Lock()
	di.mode = (di.mode + 1) % 3
	di.scrollOffset = 0
	di.mu.Unlock()
}

// SetRoot sets the root component for tree inspection.
func (di *DebugInspector) SetRoot(c Component) {
	di.mu.Lock()
	di.root = c
	di.mu.Unlock()
}

// SetPanelSize sets the overlay panel dimensions.
func (di *DebugInspector) SetPanelSize(w, h int) {
	di.mu.Lock()
	di.panelW = w
	di.panelH = h
	di.mu.Unlock()
}

// SetTitle sets the panel title.
func (di *DebugInspector) SetTitle(title string) {
	di.mu.Lock()
	di.title = title
	di.mu.Unlock()
}

// Title returns the panel title.
func (di *DebugInspector) Title() string {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.title
}

// ScrollUp scrolls the event log up by n lines.
func (di *DebugInspector) ScrollUp(n int) {
	di.mu.Lock()
	di.scrollOffset -= n
	if di.scrollOffset < 0 {
		di.scrollOffset = 0
	}
	di.mu.Unlock()
}

// ScrollDown scrolls the event log down by n lines.
func (di *DebugInspector) ScrollDown(n int) {
	di.mu.Lock()
	di.scrollOffset += n
	di.mu.Unlock()
}

// Events returns a copy of the recorded event entries.
func (di *DebugInspector) Events() []EventEntry {
	di.mu.RLock()
	defer di.mu.RUnlock()
	result := make([]EventEntry, len(di.events))
	copy(result, di.events)
	return result
}

// Stats returns a copy of the current render statistics.
func (di *DebugInspector) Stats() RenderStats {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.stats
}

// RecordKey logs a key event.
func (di *DebugInspector) RecordKey(k *term.KeyEvent) {
	if k == nil {
		return
	}
	summary := keyEventSummary(k)
	di.recordEvent(EventEntry{
		Type:      EventKey,
		Timestamp: time.Now(),
		Summary:   summary,
	})
}

// RecordMouse logs a mouse event.
func (di *DebugInspector) RecordMouse(m *term.MouseEvent) {
	if m == nil {
		return
	}
	summary := fmt.Sprintf("Mouse (%d,%d) btn=%d act=%d", m.X, m.Y, m.Button, m.Action)
	di.recordEvent(EventEntry{
		Type:      EventMouse,
		Timestamp: time.Now(),
		Summary:   summary,
	})
}

// RecordResize logs a resize event.
func (di *DebugInspector) RecordResize(w, h int) {
	di.recordEvent(EventEntry{
		Type:      EventResize,
		Timestamp: time.Now(),
		Summary:   fmt.Sprintf("Resize %dx%d", w, h),
	})
}

// RecordCustom logs a custom event with an arbitrary summary string.
func (di *DebugInspector) RecordCustom(summary string) {
	di.recordEvent(EventEntry{
		Type:      EventCustom,
		Timestamp: time.Now(),
		Summary:   summary,
	})
}

// RecordRender updates render statistics. Call this after each frame.
func (di *DebugInspector) RecordRender(duration time.Duration, cellCount int, dirty bool) {
	di.mu.Lock()
	di.stats.FrameCount++
	di.stats.TotalCells += cellCount
	ns := duration.Nanoseconds()
	di.stats.TotalRenderNs += ns
	di.stats.LastRenderNs = ns
	if dirty {
		di.stats.DirtyCount++
	}
	di.mu.Unlock()
}

// ResetStats clears all render statistics.
func (di *DebugInspector) ResetStats() {
	di.mu.Lock()
	di.stats = RenderStats{}
	di.mu.Unlock()
}

// ClearEvents clears the event log.
func (di *DebugInspector) ClearEvents() {
	di.mu.Lock()
	di.events = di.events[:0]
	di.eventHead = 0
	di.scrollOffset = 0
	di.mu.Unlock()
}

// HandleKey processes keyboard input for the inspector.
// Tab cycles modes, Up/Down scroll, Esc hides.
// Returns true if the key was consumed.
func (di *DebugInspector) HandleKey(k *term.KeyEvent) bool {
	if k == nil {
		return false
	}

	di.mu.RLock()
	visible := di.visible
	di.mu.RUnlock()

	if !visible {
		return false
	}

	switch {
	case k.Key == term.KeyEscape:
		di.Hide()
		return true
	case k.Key == term.KeyTab:
		di.NextMode()
		return true
	case k.Key == term.KeyUp:
		di.ScrollUp(1)
		return true
	case k.Key == term.KeyDown:
		di.ScrollDown(1)
		return true
	case k.Key == term.KeyPageUp:
		di.ScrollUp(10)
		return true
	case k.Key == term.KeyPageDown:
		di.ScrollDown(10)
		return true
	}

	return false
}

// recordEvent adds an event to the ring buffer.
func (di *DebugInspector) recordEvent(e EventEntry) {
	di.mu.Lock()
	defer di.mu.Unlock()

	if len(di.events) < di.maxEvents {
		di.events = append(di.events, e)
	} else {
		di.events[di.eventHead] = e
		di.eventHead = (di.eventHead + 1) % di.maxEvents
	}
}

// Measure returns the desired panel size.
func (di *DebugInspector) Measure(cs Constraints) Size {
	di.mu.RLock()
	w := di.panelW
	h := di.panelH
	di.mu.RUnlock()

	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// SetBounds sets the overlay position within the terminal.
func (di *DebugInspector) SetBounds(r Rect) {
	di.mu.Lock()
	di.bounds = r
	// Position panel at top-right corner
	pw := di.panelW
	ph := di.panelH
	if pw > r.W {
		pw = r.W
	}
	if ph > r.H {
		ph = r.H
	}
	di.panelX = r.X + r.W - pw
	di.panelY = r.Y
	di.mu.Unlock()
}

// Bounds returns the current bounds.
func (di *DebugInspector) Bounds() Rect {
	di.mu.RLock()
	defer di.mu.RUnlock()
	return di.bounds
}

// Paint renders the inspector overlay panel.
func (di *DebugInspector) Paint(buf *buffer.Buffer) {
	di.mu.RLock()
	defer di.mu.RUnlock()

	if !di.visible {
		return
	}

	// Clamp panel to buffer
	pw := di.panelW
	ph := di.panelH
	if di.panelX+pw > di.bounds.X+di.bounds.W {
		pw = di.bounds.X + di.bounds.W - di.panelX
	}
	if di.panelY+ph > di.bounds.Y+di.bounds.H {
		ph = di.bounds.Y + di.bounds.H - di.panelY
	}
	if pw < 5 || ph < 3 {
		return
	}

	// Draw panel background and border
	borderStyle := buffer.Style{
		Fg: buffer.NamedColor(buffer.NamedCyan),
	}
	bgStyle := buffer.Style{
		Bg: buffer.RGB(0x1e, 0x1e, 0x2e),
		Fg: buffer.NamedColor(buffer.NamedWhite),
	}

	// Fill background
	for y := di.panelY; y < di.panelY+ph && y < buf.Height; y++ {
		for x := di.panelX; x < di.panelX+pw && x < buf.Width; x++ {
			buf.SetCell(x, y, buffer.NewCell(' ', bgStyle))
		}
	}

	// Draw border
	for x := di.panelX; x < di.panelX+pw && x < buf.Width; x++ {
		buf.SetCell(x, di.panelY, buffer.NewCell('─', borderStyle))
		buf.SetCell(x, di.panelY+ph-1, buffer.NewCell('─', borderStyle))
	}
	for y := di.panelY; y < di.panelY+ph && y < buf.Height; y++ {
		buf.SetCell(di.panelX, y, buffer.NewCell('│', borderStyle))
		buf.SetCell(di.panelX+pw-1, y, buffer.NewCell('│', borderStyle))
	}

	// Corners
	if di.panelX < buf.Width && di.panelY < buf.Height {
		buf.SetCell(di.panelX, di.panelY, buffer.NewCell('┌', borderStyle))
		buf.SetCell(di.panelX+pw-1, di.panelY, buffer.NewCell('┐', borderStyle))
		buf.SetCell(di.panelX, di.panelY+ph-1, buffer.NewCell('└', borderStyle))
		buf.SetCell(di.panelX+pw-1, di.panelY+ph-1, buffer.NewCell('┘', borderStyle))
	}

	// Draw title
	titleText := " " + di.title + " "
	tx := di.panelX + 2
	for _, r := range titleText {
		if tx >= di.panelX+pw-1 {
			break
		}
		buf.SetCell(tx, di.panelY, buffer.NewCell(r, buffer.Style{
			Fg:    buffer.NamedColor(buffer.NamedCyan),
			Flags: buffer.Bold,
		}))
		tx++
	}

	// Draw mode tabs
	modeText := "[1]Tree [2]Events [3]Stats"
	mx := di.panelX + 1
	my := di.panelY + 1
	for _, r := range modeText {
		if mx >= di.panelX+pw-1 {
			break
		}
		style := bgStyle
		isActive := false
		switch r {
		case 'T':
			isActive = di.mode == InspectTree
		case 'E':
			isActive = di.mode == InspectEvents
		case 'S':
			isActive = di.mode == InspectStats
		}
		if isActive {
			style = buffer.Style{
				Fg:    buffer.NamedColor(buffer.NamedYellow),
				Bg:    buffer.RGB(0x1e, 0x1e, 0x2e),
				Flags: buffer.Bold,
			}
		}
		buf.SetCell(mx, my, buffer.NewCell(r, style))
		mx++
	}

	// Draw content area
	contentY := di.panelY + 2
	contentH := ph - 3 // minus top border, tabs, bottom border
	contentX := di.panelX + 1
	contentW := pw - 2

	switch di.mode {
	case InspectTree:
		di.paintTreeLocked(buf, contentX, contentY, contentW, contentH)
	case InspectEvents:
		di.paintEventsLocked(buf, contentX, contentY, contentW, contentH)
	case InspectStats:
		di.paintStatsLocked(buf, contentX, contentY, contentW, contentH)
	}
}

// paintTreeLocked draws the component tree.
func (di *DebugInspector) paintTreeLocked(buf *buffer.Buffer, x, y, w, h int) {
	if di.root == nil {
		di.drawText(buf, x, y, w, "(no root component set)", buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedBrightBlack),
		})
		return
	}

	row := 0
	var walk func(c Component, depth int)
	walk = func(c Component, depth int) {
		if row >= h || y+row >= buf.Height {
			return
		}

		indent := strings.Repeat("  ", depth)
		typeName := componentTypeName(c)
		bounds := c.Bounds()
		info := fmt.Sprintf("%s%s [%dx%d @(%d,%d)]", indent, typeName, bounds.W, bounds.H, bounds.X, bounds.Y)

		style := buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedWhite),
			Bg: buffer.RGB(0x1e, 0x1e, 0x2e),
		}
		if depth == 0 {
			style.Flags = buffer.Bold
		}

		di.drawText(buf, x, y+row, w, info, style)
		row++

		for _, child := range c.Children() {
			walk(child, depth+1)
		}
	}
	walk(di.root, 0)
}

// paintEventsLocked draws the event log.
func (di *DebugInspector) paintEventsLocked(buf *buffer.Buffer, x, y, w, h int) {
	if len(di.events) == 0 {
		di.drawText(buf, x, y, w, "(no events recorded)", buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedBrightBlack),
		})
		return
	}

	// Determine visible range with scroll
	total := len(di.events)
	start := 0
	if total > h {
		start = total - h - di.scrollOffset
		if start < 0 {
			start = 0
		}
	}
	end := start + h
	if end > total {
		end = total
	}

	row := 0
	for i := start; i < end && row < h; i++ {
		e := di.events[i]
		ts := e.Timestamp.Format("15:04:05.000")

		var typeIcon string
		var typeColor buffer.Color
		switch e.Type {
		case EventKey:
			typeIcon = "K"
			typeColor = buffer.NamedColor(buffer.NamedGreen)
		case EventMouse:
			typeIcon = "M"
			typeColor = buffer.NamedColor(buffer.NamedBlue)
		case EventResize:
			typeIcon = "R"
			typeColor = buffer.NamedColor(buffer.NamedYellow)
		case EventCustom:
			typeIcon = "C"
			typeColor = buffer.NamedColor(buffer.NamedMagenta)
		}

		line := fmt.Sprintf("%s %s %s", ts, typeIcon, truncate(e.Summary, w-14))
		di.drawText(buf, x, y+row, w, line, buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedWhite),
			Bg: buffer.RGB(0x1e, 0x1e, 0x2e),
		})

		// Color the type icon
		buf.SetCell(x+12, y+row, buffer.NewCell([]rune(typeIcon)[0], buffer.Style{
			Fg:    typeColor,
			Bg:    buffer.RGB(0x1e, 0x1e, 0x2e),
			Flags: buffer.Bold,
		}))

		row++
	}
}

// paintStatsLocked draws the render statistics.
func (di *DebugInspector) paintStatsLocked(buf *buffer.Buffer, x, y, w, h int) {
	s := di.stats

	avgNs := int64(0)
	if s.FrameCount > 0 {
		avgNs = s.TotalRenderNs / int64(s.FrameCount)
	}
	avgCells := 0
	if s.FrameCount > 0 {
		avgCells = s.TotalCells / s.FrameCount
	}
	dirtyPct := 0.0
	if s.FrameCount > 0 {
		dirtyPct = float64(s.DirtyCount) * 100.0 / float64(s.FrameCount)
	}

	lines := []struct {
		label string
		value string
		color buffer.Color
	}{
		{"Frames", fmt.Sprintf("%d", s.FrameCount), buffer.NamedColor(buffer.NamedWhite)},
		{"Last Render", fmt.Sprintf("%.2f ms", float64(s.LastRenderNs)/1e6), buffer.NamedColor(buffer.NamedGreen)},
		{"Avg Render", fmt.Sprintf("%.2f ms", float64(avgNs)/1e6), buffer.NamedColor(buffer.NamedCyan)},
		{"Avg Cells", fmt.Sprintf("%d", avgCells), buffer.NamedColor(buffer.NamedWhite)},
		{"Dirty Frames", fmt.Sprintf("%d (%.1f%%)", s.DirtyCount, dirtyPct), buffer.NamedColor(buffer.NamedYellow)},
		{"Total Cells", fmt.Sprintf("%d", s.TotalCells), buffer.NamedColor(buffer.NamedWhite)},
	}

	for i, ln := range lines {
		if i >= h {
			break
		}
		labelText := fmt.Sprintf("  %-14s", ln.label)
		di.drawText(buf, x, y+i, w, labelText, buffer.Style{
			Fg: buffer.NamedColor(buffer.NamedBrightBlack),
			Bg: buffer.RGB(0x1e, 0x1e, 0x2e),
		})
		valueText := ln.value
		for j, r := range valueText {
			vx := x + 16 + j
			if vx >= x+w {
				break
			}
			buf.SetCell(vx, y+i, buffer.NewCell(r, buffer.Style{
				Fg: ln.color,
				Bg: buffer.RGB(0x1e, 0x1e, 0x2e),
				Flags: buffer.Bold,
			}))
		}
	}
}

// drawText writes a string at the given position with a style.
func (di *DebugInspector) drawText(buf *buffer.Buffer, x, y, maxW int, text string, style buffer.Style) {
	col := 0
	for _, r := range text {
		if col >= maxW || x+col >= buf.Width {
			break
		}
		buf.SetCell(x+col, y, buffer.NewCell(r, style))
		col++
	}
}

// Children returns nil (inspector has no children in the component tree).
func (di *DebugInspector) Children() []Component {
	return nil
}

// --- Helpers ---

// keyEventSummary creates a readable summary of a key event.
func keyEventSummary(k *term.KeyEvent) string {
	var parts []string

	if k.Modifiers != 0 {
		if k.Modifiers&term.ModCtrl != 0 {
			parts = append(parts, "Ctrl")
		}
		if k.Modifiers&term.ModAlt != 0 {
			parts = append(parts, "Alt")
		}
		if k.Modifiers&term.ModShift != 0 {
			parts = append(parts, "Shift")
		}
	}

	if k.Rune != 0 {
		parts = append(parts, fmt.Sprintf("%q", string(k.Rune)))
	} else {
		parts = append(parts, keyName(k.Key))
	}

	return strings.Join(parts, "+")
}

// keyName returns a human-readable name for a KeyCode.
func keyName(k term.KeyCode) string {
	names := map[term.KeyCode]string{
		term.KeyUp:        "Up",
		term.KeyDown:      "Down",
		term.KeyLeft:      "Left",
		term.KeyRight:     "Right",
		term.KeyEnter:     "Enter",
		term.KeyTab:       "Tab",
		term.KeyBackspace: "Backspace",
		term.KeyDelete:    "Delete",
		term.KeyInsert:    "Insert",
		term.KeyHome:      "Home",
		term.KeyEnd:       "End",
		term.KeyPageUp:    "PageUp",
		term.KeyPageDown:  "PageDown",
		term.KeyEscape:    "Esc",
		term.KeySpace:     "Space",
		term.KeyF1:        "F1",
		term.KeyF2:        "F2",
		term.KeyF3:        "F3",
		term.KeyF4:        "F4",
		term.KeyF5:        "F5",
		term.KeyF6:        "F6",
		term.KeyF7:        "F7",
		term.KeyF8:        "F8",
		term.KeyF9:        "F9",
		term.KeyF10:       "F10",
		term.KeyF11:       "F11",
		term.KeyF12:       "F12",
	}
	if name, ok := names[k]; ok {
		return name
	}
	return fmt.Sprintf("Key(%d)", k)
}

// componentTypeName extracts a short type name from a component.
func componentTypeName(c Component) string {
	if c == nil {
		return "nil"
	}
	id := c.ID()
	// Try to extract component type from ID prefix
	if idx := indexOf(id, "-"); idx > 0 {
		return id[:idx]
	}
	return id
}

// indexOf returns the index of the first occurrence of sep in s, or -1.
func indexOf(s, sep string) int {
	return strings.Index(s, sep)
}

// truncate shortens a string to maxLen, adding "…" if truncated.
func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(runes[:maxLen-1]) + "…"
}
