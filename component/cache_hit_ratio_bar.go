package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CacheHitRatioBar: Cache Hit/Miss Ratio Display ───
//
// CacheHitRatioBar renders a horizontal stacked bar showing cache hit vs miss
// ratio with percentage labels and a sliding window indicator.
//
// Usage:
//
//	ch := NewCacheHitRatioBar()
//	ch.SetHits(850)
//	ch.SetMisses(150)
//	ch.Paint(buf)

// CacheHitRatioStyle holds styling.
type CacheHitRatioStyle struct {
	Hit     buffer.Style
	Miss    buffer.Style
	Label   buffer.Style
	Percent buffer.Style
	Border  buffer.Style
}

// DefaultCacheHitRatioStyle returns defaults.
func DefaultCacheHitRatioStyle() CacheHitRatioStyle {
	hit := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	miss := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	label := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	pct := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return CacheHitRatioStyle{Hit: hit, Miss: miss, Label: label, Percent: pct, Border: border}
}

// CacheHitRatioBar renders a cache hit/miss stacked bar.
type CacheHitRatioBar struct {
	BaseComponent
	mu sync.Mutex

	hits  int
	miss  int
	width int
	style CacheHitRatioStyle
	// cached
	hitPctStr string
	missPctStr string
}

// NewCacheHitRatioBar creates a CacheHitRatioBar.
func NewCacheHitRatioBar() *CacheHitRatioBar {
	ch := &CacheHitRatioBar{width: 30, style: DefaultCacheHitRatioStyle()}
	ch.SetID(GenerateID("cachehit"))
	ch.hitPctStr = "0%"
	ch.missPctStr = "0%"
	return ch
}

// SetHits sets cache hit count (caches display strings).
func (ch *CacheHitRatioBar) SetHits(n int) *CacheHitRatioBar {
	ch.mu.Lock()
	ch.hits = n
	ch.recomputeLocked()
	ch.mu.Unlock()
	return ch
}

// SetMisses sets cache miss count.
func (ch *CacheHitRatioBar) SetMisses(n int) *CacheHitRatioBar {
	ch.mu.Lock()
	ch.miss = n
	ch.recomputeLocked()
	ch.mu.Unlock()
	return ch
}

func (ch *CacheHitRatioBar) recomputeLocked() {
	total := ch.hits + ch.miss
	if total == 0 {
		ch.hitPctStr = "0%"
		ch.missPctStr = "0%"
		return
	}
	hitPct := ch.hits * 100 / total
	ch.hitPctStr = itoa(hitPct) + "%"
	ch.missPctStr = itoa(100-hitPct) + "%"
}

// SetWidth sets bar width.
func (ch *CacheHitRatioBar) SetWidth(w int) *CacheHitRatioBar {
	ch.mu.Lock()
	if w < 10 { w = 10 }
	ch.width = w
	ch.mu.Unlock()
	return ch
}

// HitPercent returns the hit percentage.
func (ch *CacheHitRatioBar) HitPercent() int {
	ch.mu.Lock()
	defer ch.mu.Unlock()
	total := ch.hits + ch.miss
	if total == 0 { return 0 }
	return ch.hits * 100 / total
}

// SetStyle sets custom style.
func (ch *CacheHitRatioBar) SetStyle(s CacheHitRatioStyle) *CacheHitRatioBar {
	ch.mu.Lock()
	ch.style = s
	ch.mu.Unlock()
	return ch
}

// Measure returns preferred size.
func (ch *CacheHitRatioBar) Measure(cs Constraints) Size {
	w := ch.width + 20
	h := 3
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h}
}

// Paint renders the cache hit ratio bar into the buffer.
func (ch *CacheHitRatioBar) Paint(buf *buffer.Buffer) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	b := ch.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 20 { w = 50 }

	total := ch.hits + ch.miss
	if total == 0 { return }

	barW := w - 2
	hitW := ch.hits * barW / total
	missW := barW - hitW

	hitStyle := ch.style.Hit
	missStyle := ch.style.Miss
	labelStyle := ch.style.Label

	col := x + 1
	// Hit portion
	for i := 0; i < hitW; i++ {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: hitStyle.Fg, Bg: hitStyle.Bg, Flags: hitStyle.Flags, Width: 1})
		col++
	}
	// Miss portion
	for i := 0; i < missW; i++ {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: '▓', Fg: missStyle.Fg, Bg: missStyle.Bg, Flags: missStyle.Flags, Width: 1})
		col++
	}

	// Labels on row 1
	col = x + 1
	hitLabel := "Hit: " + ch.hitPctStr
	for _, r := range hitLabel {
		if col >= buf.Width { break }
		buf.SetCell(col, y+1, buffer.Cell{Rune: r, Fg: hitStyle.Fg, Bg: hitStyle.Bg, Flags: hitStyle.Flags, Width: 1})
		col++
	}
	// Miss label right-aligned
	missLabel := "Miss: " + ch.missPctStr
	missStart := x + w - 1 - len(missLabel)
	if missStart < col { missStart = col }
	for c := col; c < missStart && c < buf.Width; c++ {
		buf.SetCell(c, y+1, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
	}
	for i, r := range missLabel {
		cx := missStart + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y+1, buffer.Cell{Rune: r, Fg: missStyle.Fg, Bg: missStyle.Bg, Flags: missStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (ch *CacheHitRatioBar) Children() []Component { return nil }
