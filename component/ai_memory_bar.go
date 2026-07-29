package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIMemoryBar: AI Conversation Memory Usage Bar ───
//
// AIMemoryBar renders a horizontal bar showing how much of the conversation
// memory window is currently used versus the model's total context limit.
// Displays segments for system prompt, conversation history, and available space.
//
// Usage:
//
//	m := NewAIMemoryBar()
//	m.SetUsage(2000, 8000, 4000) // system=2000, history=8000, limit=16000
//	m.Paint(buf)

// AIMemoryStyle holds styling.
type AIMemoryStyle struct {
	System  buffer.Style
	History buffer.Style
	Avail   buffer.Style
	Label   buffer.Style
	Warning buffer.Style
}

// DefaultAIMemoryStyle returns defaults.
func DefaultAIMemoryStyle() AIMemoryStyle {
	return AIMemoryStyle{
		System:  buffer.Style{Fg: buffer.RGB(168, 85, 247)},
		History: buffer.Style{Fg: buffer.RGB(59, 130, 246)},
		Avail:   buffer.Style{Fg: buffer.RGB(30, 41, 59)},
		Label:   buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Warning: buffer.Style{Fg: buffer.RGB(245, 158, 11), Flags: buffer.Bold},
	}
}

// AIMemoryBar renders a conversation memory usage bar.
type AIMemoryBar struct {
	BaseComponent
	mu sync.Mutex

	systemTokens  int
	historyTokens int
	contextLimit  int
	width         int
	style         AIMemoryStyle
	// cached
	usedStr    string
	limitStr   string
	sysStr     string
	histStr    string
	label1Str  string
	label2Str  string
	pctStr     string
	usedPct    int
	barSys     int
	barHist    int
	barAvail   int
	isWarning  bool
	warningStr string
}

// NewAIMemoryBar creates an AIMemoryBar.
func NewAIMemoryBar() *AIMemoryBar {
	m := &AIMemoryBar{width: 36, contextLimit: 16000, style: DefaultAIMemoryStyle()}
	m.SetID(GenerateID("aimem"))
	m.recomputeLocked()
	return m
}

// SetUsage sets system prompt tokens, conversation history tokens,
// and the model's total context window limit in tokens.
func (m *AIMemoryBar) SetUsage(system, history, limit int) *AIMemoryBar {
	m.mu.Lock()
	if system < 0 { system = 0 }
	if history < 0 { history = 0 }
	if limit < 1 { limit = 1 }
	m.systemTokens = system
	m.historyTokens = history
	m.contextLimit = limit
	m.recomputeLocked()
	m.mu.Unlock()
	return m
}

func (m *AIMemoryBar) recomputeLocked() {
	used := m.systemTokens + m.historyTokens
	if used > m.contextLimit { used = m.contextLimit }
	m.usedPct = used * 100 / m.contextLimit

	m.usedStr = itoa(used)
	m.limitStr = itoa(m.contextLimit)
	m.sysStr = itoa(m.systemTokens)
	m.histStr = itoa(m.historyTokens)
	m.label1Str = "Mem " + m.usedStr + "/" + m.limitStr
	m.label2Str = "sys:" + m.sysStr + " hist:" + m.histStr
	m.pctStr = itoa(m.usedPct) + "%"

	// Bar segments for 20-char bar
	const barW = 20
	m.barSys = m.systemTokens * barW / m.contextLimit
	m.barHist = m.historyTokens * barW / m.contextLimit
	m.barAvail = barW - m.barSys - m.barHist
	if m.barAvail < 0 { m.barAvail = 0 }

	m.isWarning = m.usedPct >= 80
	if m.isWarning {
		m.warningStr = "! " + itoa(m.usedPct) + "% used"
	} else {
		m.warningStr = ""
	}
}

// UsedPercent returns the percentage of context window used.
func (m *AIMemoryBar) UsedPercent() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usedPct
}

// SetWidth sets the bar width.
func (m *AIMemoryBar) SetWidth(w int) *AIMemoryBar {
	m.mu.Lock()
	if w < 20 { w = 20 }
	m.width = w
	m.mu.Unlock()
	return m
}

// SetStyle sets custom style.
func (m *AIMemoryBar) SetStyle(s AIMemoryStyle) *AIMemoryBar {
	m.mu.Lock()
	m.style = s
	m.mu.Unlock()
	return m
}

// Measure returns preferred size.
func (m *AIMemoryBar) Measure(cs Constraints) Size {
	w := m.width + 22
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: 3}
}

// Paint renders the AI memory bar.
func (m *AIMemoryBar) Paint(buf *buffer.Buffer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	b := m.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 20 { w = 50 }

	sysStyle := m.style.System
	histStyle := m.style.History
	availStyle := m.style.Avail
	labelStyle := m.style.Label
	warnStyle := m.style.Warning

	// Row 0: stacked bar
	col := x
	for i := 0; i < m.barSys; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: sysStyle.Fg, Bg: sysStyle.Bg, Flags: sysStyle.Flags, Width: 1})
		col++
	}
	for i := 0; i < m.barHist; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: histStyle.Fg, Bg: histStyle.Bg, Flags: histStyle.Flags, Width: 1})
		col++
	}
	for i := 0; i < m.barAvail; i++ {
		if col >= buf.Width { break }
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: availStyle.Fg, Bg: availStyle.Bg, Flags: availStyle.Flags, Width: 1})
		col++
	}

	// Row 1: labels
	label1 := m.label1Str
	col = x
	for _, r := range label1 {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Right-aligned warning or percent
	var rightLabel string
	var rightStyle buffer.Style
	if m.isWarning {
		rightLabel = m.warningStr
		rightStyle = warnStyle
	} else {
		rightLabel = m.pctStr
		rightStyle = labelStyle
	}
	rightStart := x + w - 1 - len(rightLabel)
	if rightStart < col { rightStart = col }
	for c := col; c < rightStart && c < buf.Width; c++ {
		buf.SetCell(c, y+1, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	}
	for i, r := range rightLabel {
		cx := rightStart + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: rightStyle.Fg, Bg: rightStyle.Bg, Flags: rightStyle.Flags, Width: 1})
	}

	// Row 2: detail
	label2 := m.label2Str
	col = x
	for _, r := range label2 {
		if col >= buf.Width { break }
		buf.SetCell(col, y+2, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (m *AIMemoryBar) Children() []Component { return nil }
