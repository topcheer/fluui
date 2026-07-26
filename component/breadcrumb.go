package component

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// Breadcrumb displays a navigation path as clickable segments separated
// by a delimiter (e.g., "Home > Settings > AI > Models"). Useful in
// file explorers, settings panels, and multi-level navigation.
//
// Thread-safe. Zero-alloc Paint.
type Breadcrumb struct {
	BaseComponent
	mu sync.Mutex

	items     []string
	delimiter string
	activeIdx int
}

// NewBreadcrumb creates a breadcrumb with the given path items.
func NewBreadcrumb(items []string) *Breadcrumb {
	return &Breadcrumb{
		BaseComponent: BaseComponent{id: GenerateID("breadcrumb")},
		items:         items,
		delimiter:    " \u203a ", // ›
		activeIdx:    -1,
	}
}

// SetItems replaces the path items.
func (b *Breadcrumb) SetItems(items []string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = items
}

// Items returns a copy of the current path items.
func (b *Breadcrumb) Items() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, len(b.items))
	copy(out, b.items)
	return out
}

// SetDelimiter changes the separator between items.
func (b *Breadcrumb) SetDelimiter(d string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.delimiter = d
}

// SetActive highlights the item at the given index (-1 = none).
func (b *Breadcrumb) SetActive(idx int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.activeIdx = idx
}

// ActiveIndex returns the highlighted item index.
func (b *Breadcrumb) ActiveIndex() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.activeIdx
}

// ItemCount returns the number of path items.
func (b *Breadcrumb) ItemCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.items)
}

// Push appends a path item to the end.
func (b *Breadcrumb) Push(item string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.items = append(b.items, item)
}

// Pop removes and returns the last path item.
func (b *Breadcrumb) Pop() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.items) == 0 {
		return ""
	}
	last := b.items[len(b.items)-1]
	b.items = b.items[:len(b.items)-1]
	return last
}

// Measure returns the desired size.
func (b *Breadcrumb) Measure(cs Constraints) Size {
	b.mu.Lock()
	items := b.items
	delim := b.delimiter
	b.mu.Unlock()

	w := 0
	delimW := utf8.RuneCountInString(delim)
	for i, item := range items {
		if i > 0 {
			w += delimW
		}
		w += utf8.RuneCountInString(item)
	}
	if w < 1 {
		w = 1
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 1
	}
	return Size{W: w, H: 1}
}

// Paint renders the breadcrumb path.
func (b *Breadcrumb) Paint(buf *buffer.Buffer) {
	b.mu.Lock()
	items := b.items
	delim := b.delimiter
	active := b.activeIdx
	b.mu.Unlock()

	if len(items) == 0 {
		return
	}

	bounds := b.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	th := theme.Get()
	normalStyle := buffer.Style{Fg: th.Muted}
	activeStyle := buffer.Style{Fg: th.Accent}
	delimStyle := buffer.Style{Fg: th.Border}

	x := bounds.X
	maxX := bounds.X + bounds.W

	for i, item := range items {
		style := normalStyle
		if i == active {
			style = activeStyle
		}

		// Draw delimiter (except first item)
		if i > 0 {
			if x+utf8.RuneCountInString(delim) > maxX {
				// Truncate with ellipsis
				if x < maxX {
					buf.DrawText(x, bounds.Y, "\u2026", normalStyle)
				}
				break
			}
			x = buf.DrawText(x, bounds.Y, delim, delimStyle)
		}

		// Draw item text
		itemW := utf8.RuneCountInString(item)
		if x+itemW > maxX {
			// Truncate
			remaining := maxX - x
			if remaining > 1 {
				trunc := truncateRunes(item, remaining-1)
				buf.DrawText(x, bounds.Y, trunc+"\u2026", style)
			} else if remaining == 1 {
				buf.DrawText(x, bounds.Y, "\u2026", style)
			}
			break
		}
		x = buf.DrawText(x, bounds.Y, item, style)
	}
}

// String returns the breadcrumb as a single string with delimiters.
func (b *Breadcrumb) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.Join(b.items, b.delimiter)
}
