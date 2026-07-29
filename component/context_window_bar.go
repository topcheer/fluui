package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ContextWindowBar: AI Context Window Usage Visualizer ───
//
// ContextWindowBar renders a horizontal progress bar showing how much of
// an LLM's context window is consumed. Color transitions from green→yellow→red
// as usage approaches the limit. Common in AI chat interfaces and IDE plugins.
//
// Usage:
//
//	cwb := NewContextWindowBar()
//	cwb.SetContextLimit(128000) // 128K context window
//	cwb.SetUsed(95000)          // 95K tokens used
//	cwb.Paint(buf)

// ContextWindowBarStyle holds styling for ContextWindowBar.
type ContextWindowBarStyle struct {
	BarBg     buffer.Style
	Label     buffer.Style
	Threshold [3]buffer.Style // [normal, warning, critical]
}

// DefaultContextWindowBarStyle returns sensible defaults.
func DefaultContextWindowBarStyle() ContextWindowBarStyle {
	normal := buffer.Style{Fg: buffer.RGB(34, 197, 94)}    // green-500
	warning := buffer.Style{Fg: buffer.RGB(234, 179, 8)}   // yellow-500
	critical := buffer.Style{Fg: buffer.RGB(239, 68, 68)}  // red-500
	barBg := buffer.Style{Fg: buffer.RGB(51, 65, 85)}      // slate-700
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}   // slate-400
	return ContextWindowBarStyle{
		BarBg:     barBg,
		Label:     label,
		Threshold: [3]buffer.Style{normal, warning, critical},
	}
}

// ContextWindowBar displays context window token usage as a colored bar.
type ContextWindowBar struct {
	BaseComponent
	mu sync.Mutex

	limit     int
	used      int
	barWidth  int
	showLabel bool

	style ContextWindowBarStyle
}

// NewContextWindowBar creates a ContextWindowBar with defaults.
func NewContextWindowBar() *ContextWindowBar {
	cwb := &ContextWindowBar{
		limit:     8192,
		barWidth:  30,
		showLabel: true,
		style:     DefaultContextWindowBarStyle(),
	}
	cwb.SetID(GenerateID("ctxbar"))
	return cwb
}

// SetContextLimit sets the maximum context window size in tokens.
func (cwb *ContextWindowBar) SetContextLimit(n int) *ContextWindowBar {
	cwb.mu.Lock()
	cwb.limit = n
	cwb.mu.Unlock()
	return cwb
}

// ContextLimit returns the context window limit.
func (cwb *ContextWindowBar) ContextLimit() int {
	cwb.mu.Lock()
	defer cwb.mu.Unlock()
	return cwb.limit
}

// SetUsed sets the number of tokens currently used.
func (cwb *ContextWindowBar) SetUsed(n int) *ContextWindowBar {
	cwb.mu.Lock()
	cwb.used = n
	cwb.mu.Unlock()
	return cwb
}

// Used returns the current token usage.
func (cwb *ContextWindowBar) Used() int {
	cwb.mu.Lock()
	defer cwb.mu.Unlock()
	return cwb.used
}

// UsagePercent returns the percentage of context window used (0-100).
func (cwb *ContextWindowBar) UsagePercent() float64 {
	cwb.mu.Lock()
	defer cwb.mu.Unlock()
	if cwb.limit <= 0 {
		return 0
	}
	pct := float64(cwb.used) / float64(cwb.limit) * 100
	if pct > 100 {
		pct = 100
	}
	return pct
}

// SetBarWidth sets the width of the bar in characters.
func (cwb *ContextWindowBar) SetBarWidth(w int) *ContextWindowBar {
	cwb.mu.Lock()
	cwb.barWidth = w
	cwb.mu.Unlock()
	return cwb
}

// SetShowLabel toggles whether the label is shown.
func (cwb *ContextWindowBar) SetShowLabel(v bool) *ContextWindowBar {
	cwb.mu.Lock()
	cwb.showLabel = v
	cwb.mu.Unlock()
	return cwb
}

// SetStyle sets the custom style.
func (cwb *ContextWindowBar) SetStyle(s ContextWindowBarStyle) *ContextWindowBar {
	cwb.mu.Lock()
	cwb.style = s
	cwb.mu.Unlock()
	return cwb
}

// thresholdLevel returns 0=normal, 1=warning, 2=critical based on usage.
func (cwb *ContextWindowBar) thresholdLevel() int {
	pct := cwb.UsagePercent()
	switch {
	case pct >= 85:
		return 2
	case pct >= 60:
		return 1
	default:
		return 0
	}
}

// Measure returns the preferred size.
func (cwb *ContextWindowBar) Measure(cs Constraints) Size {
	w := cwb.barWidth + 20
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the context window bar into the buffer.
func (cwb *ContextWindowBar) Paint(buf *buffer.Buffer) {
	cwb.mu.Lock()
	defer cwb.mu.Unlock()

	b := cwb.Bounds()
	x, y := b.X, b.Y

	// Calculate filled portion
	pct := 0.0
	if cwb.limit > 0 {
		pct = float64(cwb.used) / float64(cwb.limit)
		if pct > 1.0 {
			pct = 1.0
		}
	}
	filled := int(pct * float64(cwb.barWidth))

	// Threshold color
	level := 0
	switch {
	case pct >= 0.85:
		level = 2
	case pct >= 0.60:
		level = 1
	}
	fillStyle := cwb.style.Threshold[level]

	// Draw bar background and fill
	col := x
	if cwb.showLabel {
		label := "ctx "
		for _, r := range label {
			if col < buf.Width {
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: cwb.style.Label.Fg, Bg: cwb.style.Label.Bg, Flags: cwb.style.Label.Flags, Width: 1})
			}
			col++
		}
	}

	barStart := col
	for i := 0; i < cwb.barWidth; i++ {
		if col >= buf.Width {
			break
		}
		if i < filled {
			buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
		} else {
			buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: cwb.style.BarBg.Fg, Bg: cwb.style.BarBg.Bg, Flags: cwb.style.BarBg.Flags, Width: 1})
		}
		col++
	}

	// Draw usage text after bar
	usageStr := " " + itoa(cwb.used) + "/" + formatTokenCount(cwb.limit)
	for _, r := range usageStr {
		if col < buf.Width {
			buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: cwb.style.Label.Fg, Bg: cwb.style.Label.Bg, Flags: cwb.style.Label.Flags, Width: 1})
		}
		col++
	}

	_ = barStart // barStart kept for potential future extensions
}

// formatTokenCount converts large token counts to compact notation (e.g. 128K).
func formatTokenCount(n int) string {
	if n >= 1000000 {
		return itoa(n/1000000) + "M"
	}
	if n >= 1000 {
		return itoa(n/1000) + "K"
	}
	return itoa(n)
}

// Children returns nil.
func (cwb *ContextWindowBar) Children() []Component { return nil }
