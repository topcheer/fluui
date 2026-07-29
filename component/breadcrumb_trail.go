package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── BreadcrumbTrail: Path Navigation with Auto-Truncation ───
//
// BreadcrumbTrail renders a path as breadcrumbs with separators, automatically
// truncating early segments with an ellipsis when the total width exceeds the
// available space.
//
// Usage:
//
//	bt := NewBreadcrumbTrail()
//	bt.AddCrumb("Home")
//	bt.AddCrumb("Projects")
//	bt.AddCrumb("Fluui")
//	bt.SetSeparator('›')
//	bt.Paint(buf)

// BreadcrumbStyle holds styling for BreadcrumbTrail.
type BreadcrumbStyle struct {
	Crumb     buffer.Style
	Active    buffer.Style // last crumb (current location)
	Separator buffer.Style
	Ellipsis  buffer.Style
	Border    buffer.Style
}

// DefaultBreadcrumbStyle returns sensible defaults.
func DefaultBreadcrumbStyle() BreadcrumbStyle {
	crumb := buffer.Style{Fg: buffer.RGB(148, 163, 184)}    // slate-400
	active := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold} // slate-200 bold
	sep := buffer.Style{Fg: buffer.RGB(71, 85, 105)}        // slate-600
	ellipsis := buffer.Style{Fg: buffer.RGB(100, 116, 139)} // slate-500
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}     // slate-600
	return BreadcrumbStyle{Crumb: crumb, Active: active, Separator: sep, Ellipsis: ellipsis, Border: border}
}

// BreadcrumbTrail renders a path with separators and auto-truncation.
type BreadcrumbTrail struct {
	BaseComponent
	mu sync.Mutex

	crumbs    []string
	separator rune
	style     BreadcrumbStyle
}

// NewBreadcrumbTrail creates a BreadcrumbTrail with defaults.
func NewBreadcrumbTrail() *BreadcrumbTrail {
	bt := &BreadcrumbTrail{
		separator: '›',
		style:     DefaultBreadcrumbStyle(),
	}
	bt.SetID(GenerateID("breadcrumb"))
	return bt
}

// AddCrumb adds a path segment.
func (bt *BreadcrumbTrail) AddCrumb(label string) *BreadcrumbTrail {
	bt.mu.Lock()
	bt.crumbs = append(bt.crumbs, label)
	bt.mu.Unlock()
	return bt
}

// SetCrumbs replaces all crumbs.
func (bt *BreadcrumbTrail) SetCrumbs(labels []string) *BreadcrumbTrail {
	bt.mu.Lock()
	bt.crumbs = labels
	bt.mu.Unlock()
	return bt
}

// SetSeparator sets the separator character.
func (bt *BreadcrumbTrail) SetSeparator(r rune) *BreadcrumbTrail {
	bt.mu.Lock()
	bt.separator = r
	bt.mu.Unlock()
	return bt
}

// CrumbCount returns the number of crumbs.
func (bt *BreadcrumbTrail) CrumbCount() int {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	return len(bt.crumbs)
}

// SetStyle sets the custom style.
func (bt *BreadcrumbTrail) SetStyle(s BreadcrumbStyle) *BreadcrumbTrail {
	bt.mu.Lock()
	bt.style = s
	bt.mu.Unlock()
	return bt
}

// Measure returns the preferred size.
func (bt *BreadcrumbTrail) Measure(cs Constraints) Size {
	w := 60
	h := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the breadcrumb trail into the buffer.
func (bt *BreadcrumbTrail) Paint(buf *buffer.Buffer) {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	b := bt.Bounds()
	x, y := b.X, b.Y
	w := b.W
	if w < 10 {
		w = 60
	}

	if len(bt.crumbs) == 0 {
		return
	}

	// Calculate total width needed
	sepStr := [3]rune{' ', bt.separator, ' '}
	sepW := len(sepStr)
	totalW := 0
	crumbsW := [32]int{} // stack-allocated, max 32 crumbs
	crumbCount := len(bt.crumbs)
	if crumbCount > 32 {
		crumbCount = 32
	}
	for i := 0; i < crumbCount; i++ {
		w := 0
		for range bt.crumbs[i] {
			w++
		}
		crumbsW[i] = w
		totalW += w
		if i < crumbCount-1 {
			totalW += sepW
		}
	}

	availW := w
	if bt.style.Border.Fg != bt.style.Crumb.Fg {
		availW = w // no border in breadcrumb
	}

	// Determine which crumbs to show vs truncate
	col := x
	showAll := totalW <= availW

	if showAll {
		// Draw all crumbs normally
		for i, c := range bt.crumbs {
			var style buffer.Style
			if i == len(bt.crumbs)-1 {
				style = bt.style.Active
			} else {
				style = bt.style.Crumb
			}
			for _, r := range c {
				if col >= x+w || col >= buf.Width {
					break
				}
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
				col++
			}
			// Separator
			if i < len(bt.crumbs)-1 {
				sepStyle := bt.style.Separator
				for _, r := range sepStr {
					if col >= x+w || col >= buf.Width {
						break
					}
					buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
					col++
				}
			}
		}
	} else {
		// Truncate: show last crumb fully, add ellipsis for hidden ones
		ellipsisStyle := bt.style.Ellipsis
		activeStyle := bt.style.Active
		crumbStyle := bt.style.Crumb
		sepStyle := bt.style.Separator

		// Always show last crumb
		lastW := crumbsW[len(bt.crumbs)-1]

		// Work backwards from second-to-last to fit as many as possible
		remainingSpace := availW - lastW
		showFromIdx := len(bt.crumbs) - 1 // default: only last
		hasEllipsis := false

		for i := len(bt.crumbs) - 2; i >= 0; i-- {
			needed := crumbsW[i] + sepW
			if remainingSpace >= needed {
				remainingSpace -= needed
				showFromIdx = i
			} else {
				// Can't fit this one — add ellipsis if space allows
				if remainingSpace >= 4 { // "… › " = ~4 chars
					hasEllipsis = true
					remainingSpace -= 4
				}
				break
			}
		}

		// Draw ellipsis if needed
		if hasEllipsis {
			ellipsisRunes := []rune{'…'}
			for _, r := range ellipsisRunes {
				if col >= x+w || col >= buf.Width {
					break
				}
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: ellipsisStyle.Fg, Bg: ellipsisStyle.Bg, Flags: ellipsisStyle.Flags, Width: 1})
				col++
			}
			// Separator after ellipsis
			for _, r := range sepStr {
				if col >= x+w || col >= buf.Width {
					break
				}
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
				col++
			}
		}

		// Draw visible crumbs from showFromIdx
		for i := showFromIdx; i < len(bt.crumbs); i++ {
			var style buffer.Style
			if i == len(bt.crumbs)-1 {
				style = activeStyle
			} else {
				style = crumbStyle
			}
			for _, r := range bt.crumbs[i] {
				if col >= x+w || col >= buf.Width {
					break
				}
				buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
				col++
			}
			if i < len(bt.crumbs)-1 {
				for _, r := range sepStr {
					if col >= x+w || col >= buf.Width {
						break
					}
					buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: sepStyle.Fg, Bg: sepStyle.Bg, Flags: sepStyle.Flags, Width: 1})
					col++
				}
			}
		}
	}
}

// Children returns nil.
func (bt *BreadcrumbTrail) Children() []Component { return nil }
