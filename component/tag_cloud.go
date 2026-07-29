package component

import (
	"sort"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── TagCloud: Weighted Tag Cloud Display ───
//
// TagCloud displays tags with sizes proportional to their weight/frequency.
// Tags are sorted alphabetically and rendered with size-based coloring.
//
// Usage:
//
//	tc := NewTagCloud()
//	tc.AddTag("go", 50)
//	tc.AddTag("tui", 30)
//	tc.AddTag("ai", 80)
//	tc.Paint(buf)

// TagItem represents a single tag with name and weight.
type TagItem struct {
	Name   string
	Weight int
}

// TagCloudStyle holds styling for TagCloud.
type TagCloudStyle struct {
	Small  buffer.Style  // low weight
	Medium buffer.Style  // medium weight
	Large  buffer.Style  // high weight
	Border buffer.Style
}

// DefaultTagCloudStyle returns sensible defaults.
func DefaultTagCloudStyle() TagCloudStyle {
	small := buffer.Style{Fg: buffer.RGB(100, 116, 139)}   // slate-500
	medium := buffer.Style{Fg: buffer.RGB(96, 165, 250)}   // blue-400
	large := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold} // violet-400 bold
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}    // slate-600
	return TagCloudStyle{Small: small, Medium: medium, Large: large, Border: border}
}

// TagCloud displays weighted tags with size-based styling.
type TagCloud struct {
	BaseComponent
	mu sync.Mutex

	tags  []TagItem
	style TagCloudStyle

	// cached sorted tags and computed sizes
	cachedSorted []TagItem
	cachedDirty  bool
}

// NewTagCloud creates a TagCloud with defaults.
func NewTagCloud() *TagCloud {
	tc := &TagCloud{
		style:       DefaultTagCloudStyle(),
		cachedDirty: true,
	}
	tc.SetID(GenerateID("tagcloud"))
	return tc
}

// AddTag adds a tag with weight.
func (tc *TagCloud) AddTag(name string, weight int) *TagCloud {
	tc.mu.Lock()
	tc.tags = append(tc.tags, TagItem{Name: name, Weight: weight})
	tc.cachedDirty = true
	tc.mu.Unlock()
	return tc
}

// SetTags replaces all tags.
func (tc *TagCloud) SetTags(tags []TagItem) *TagCloud {
	tc.mu.Lock()
	tc.tags = tags
	tc.cachedDirty = true
	tc.mu.Unlock()
	return tc
}

// TagCount returns the number of tags.
func (tc *TagCloud) TagCount() int {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return len(tc.tags)
}

// SetStyle sets the custom style.
func (tc *TagCloud) SetStyle(s TagCloudStyle) *TagCloud {
	tc.mu.Lock()
	tc.style = s
	tc.mu.Unlock()
	return tc
}

// computeSortedLocked sorts tags alphabetically. Caller must hold lock.
func (tc *TagCloud) computeSortedLocked() {
	if !tc.cachedDirty {
		return
	}
	tc.cachedDirty = false
	tc.cachedSorted = tc.cachedSorted[:0]
	tc.cachedSorted = append(tc.cachedSorted, tc.tags...)
	sort.Slice(tc.cachedSorted, func(i, j int) bool {
		return tc.cachedSorted[i].Name < tc.cachedSorted[j].Name
	})
}

// maxWeightLocked returns the maximum weight.
func (tc *TagCloud) maxWeightLocked() int {
	max := 1
	for _, t := range tc.cachedSorted {
		if t.Weight > max {
			max = t.Weight
		}
	}
	return max
}

// tagStyleForWeight returns the style based on weight ratio.
func (tc *TagCloud) tagStyleForWeight(weight, max int) buffer.Style {
	if max <= 0 { max = 1 }
	ratio := float64(weight) / float64(max)
	if ratio >= 0.66 {
		return tc.style.Large
	}
	if ratio >= 0.33 {
		return tc.style.Medium
	}
	return tc.style.Small
}

// Measure returns the preferred size.
func (tc *TagCloud) Measure(cs Constraints) Size {
	w := 50
	h := 5
	tc.mu.Lock()
	if len(tc.tags) > 10 { h = len(tc.tags)/4 + 3 }
	tc.mu.Unlock()
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	if cs.MaxHeight > 0 && h > cs.MaxHeight { h = cs.MaxHeight }
	return Size{W: w, H: h}
}

// Paint renders the tag cloud into the buffer.
func (tc *TagCloud) Paint(buf *buffer.Buffer) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	b := tc.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 { w = 50 }
	if h < 3 { h = 5 }

	tc.computeSortedLocked()
	if len(tc.cachedSorted) == 0 { return }

	maxW := tc.maxWeightLocked()

	// Draw border
	bs := tc.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 { ch = '┌' } else if row == 0 && col == w-1 { ch = '┐' } else if row == h-1 && col == 0 { ch = '└' } else if row == h-1 && col == w-1 { ch = '┘' } else if row == 0 || row == h-1 { ch = '─' } else if col == 0 || col == w-1 { ch = '│' }
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	// Draw tags in a flowing layout
	col := x + 1
	rowY := y + 1
	spacing := 2

	for _, tag := range tc.cachedSorted {
		if rowY >= y+h-1 || rowY >= buf.Height { break }

		tagLen := len(tag.Name) + spacing
		if col+tagLen > x+w-1 {
			rowY++
			col = x + 1
			if rowY >= y+h-1 || rowY >= buf.Height { break }
		}

		style := tc.tagStyleForWeight(tag.Weight, maxW)

		for _, r := range tag.Name {
			if col >= x+w-1 || col >= buf.Width { break }
			buf.SetCell(col, rowY, buffer.Cell{Rune: r, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			col++
		}
		// Spacing
		for i := 0; i < spacing; i++ {
			if col < x+w-1 && col < buf.Width {
				buf.SetCell(col, rowY, buffer.Cell{Rune: ' ', Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
			}
			col++
		}
	}
}

// Children returns nil.
func (tc *TagCloud) Children() []Component { return nil }
