package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// AccordionItem represents a single expandable section in an Accordion.
type AccordionItem struct {
	Title    string
	Content  string // raw text content
	Expanded bool
}

// Accordion is a vertical stack of expandable sections. Unlike a single
// Collapsible, the Accordion manages multiple sections and optionally
// enforces single-open mode (accordion behavior — opening one closes others).
//
// Common uses: settings panels, FAQ displays, configuration UIs, and
// AI tool call results with multiple sections.
//
// Thread-safe.
type Accordion struct {
	BaseComponent
	mu sync.Mutex

	items     []AccordionItem
	singleOpen bool // if true, only one item can be expanded at a time
}

// NewAccordion creates an accordion with the given items.
func NewAccordion(items []AccordionItem) *Accordion {
	return &Accordion{
		BaseComponent: BaseComponent{id: GenerateID("accordion")},
		items:         items,
	}
}

// SetSingleOpen controls whether only one item can be open at a time.
// When enabled, expanding an item automatically collapses all others.
func (a *Accordion) SetSingleOpen(v bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.singleOpen = v
	if v {
		// Collapse all but the first expanded item
		foundOpen := false
		for i := range a.items {
			if a.items[i].Expanded {
				if foundOpen {
					a.items[i].Expanded = false
				}
				foundOpen = true
			}
		}
	}
}

// Items returns a copy of all items.
func (a *Accordion) Items() []AccordionItem {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AccordionItem, len(a.items))
	copy(out, a.items)
	return out
}

// SetItems replaces all items.
func (a *Accordion) SetItems(items []AccordionItem) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.items = items
}

// ItemCount returns the number of items.
func (a *Accordion) ItemCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.items)
}

// Toggle expands or collapses the item at the given index.
func (a *Accordion) Toggle(idx int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if idx < 0 || idx >= len(a.items) {
		return
	}
	a.items[idx].Expanded = !a.items[idx].Expanded
	if a.singleOpen && a.items[idx].Expanded {
		for i := range a.items {
			if i != idx {
				a.items[i].Expanded = false
			}
		}
	}
}

// Expand opens the item at the given index.
func (a *Accordion) Expand(idx int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if idx < 0 || idx >= len(a.items) {
		return
	}
	a.items[idx].Expanded = true
	if a.singleOpen {
		for i := range a.items {
			if i != idx {
				a.items[i].Expanded = false
			}
		}
	}
}

// Collapse closes the item at the given index.
func (a *Accordion) Collapse(idx int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if idx >= 0 && idx < len(a.items) {
		a.items[idx].Expanded = false
	}
}

// ExpandAll opens all items (only works when singleOpen is false).
func (a *Accordion) ExpandAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.singleOpen {
		return
	}
	for i := range a.items {
		a.items[i].Expanded = true
	}
}

// CollapseAll closes all items.
func (a *Accordion) CollapseAll() {
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.items {
		a.items[i].Expanded = false
	}
}

// IsExpanded returns whether the item at idx is expanded.
func (a *Accordion) IsExpanded(idx int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if idx < 0 || idx >= len(a.items) {
		return false
	}
	return a.items[idx].Expanded
}

// Measure returns the desired size based on item states.
func (a *Accordion) Measure(cs Constraints) Size {
	a.mu.Lock()
	defer a.mu.Unlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 40
	}
	h := 0
	for _, item := range a.items {
		h++ // header line
		if item.Expanded && item.Content != "" {
			// Count content lines
			start := 0
			for i := 0; i <= len(item.Content); i++ {
				if i < len(item.Content) && item.Content[i] != '\n' {
					continue
				}
				h++
				start = i + 1
			}
			_ = start
		}
	}
	if h < 1 {
		h = 1
	}
	return Size{W: maxW, H: h}
}

// Paint renders the accordion.
func (a *Accordion) Paint(buf *buffer.Buffer) {
	a.mu.Lock()
	items := a.items
	a.mu.Unlock()

	if len(items) == 0 {
		return
	}

	b := a.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	headerStyle := buffer.Style{Fg: th.Accent}
	contentStyle := buffer.Style{Fg: th.Fg}
	mutedStyle := buffer.Style{Fg: th.Muted}

	y := b.Y
	for _, item := range items {
		if y >= b.Y+b.H {
			break
		}

		// Header: ▸ Title  (collapsed) or ▾ Title (expanded)
		marker := "\u25b8 " // ▸
		if item.Expanded {
			marker = "\u25be " // ▾
		}
		buf.DrawText(b.X, y, marker, headerStyle)
		headerText := item.Title
		if utf8.RuneCountInString(headerText) > b.W-2 {
			headerText = truncateRunes(headerText, b.W-3) + "\u2026"
		}
		buf.DrawText(b.X+2, y, headerText, headerStyle)
		y++

		// Content
		if item.Expanded && item.Content != "" {
			// Draw each line of content
			lineStart := 0
			for i := 0; i <= len(item.Content); i++ {
				if i < len(item.Content) && item.Content[i] != '\n' {
					continue
				}
				if y >= b.Y+b.H {
					break
				}
				line := item.Content[lineStart:i]
				lineStart = i + 1
				// Indent content by 2 spaces
				contentW := b.W - 2
				if utf8.RuneCountInString(line) > contentW {
					line = truncateRunes(line, contentW-1) + "\u2026"
				}
				buf.DrawText(b.X, y, "  ", mutedStyle)
				buf.DrawText(b.X+2, y, line, contentStyle)
				y++
			}
		}
	}
}
