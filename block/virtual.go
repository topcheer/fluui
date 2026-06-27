package block

import (
	"sort"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// BlockPosition records the Y offset and height of a block within the container.
// It is computed during SetBounds and used for virtual scrolling.
type BlockPosition struct {
	Y int // top Y in container-local coordinates
	H int // height in rows
}

// blockPositions returns the cached Y positions of all blocks.
// Call SetBounds first to populate the cache.
func (c *BlockContainer) BlockPositions() []BlockPosition {
	return c.positions
}

// TotalHeight returns the total content height of all blocks plus spacing.
// This is the same value computed during Measure.
func (c *BlockContainer) TotalHeight() int {
	return c.totalHeight
}

// PaintVisible implements component.VisiblePainter.
// It paints only blocks whose Y range intersects [visibleY0, visibleY1),
// skipping off-screen blocks for efficient rendering of large lists.
//
// Uses binary search (O(log n)) to find the first visible block, then
// linearly paints only the visible ones (O(k) where k ≈ 8 for a typical
// terminal). This is a significant improvement over O(n) linear scan when
// the container has thousands of blocks.
func (c *BlockContainer) PaintVisible(buf *buffer.Buffer, visibleY0, visibleY1 int) {
	if len(c.blocks) == 0 || len(c.positions) == 0 {
		return
	}

	// Binary search: find the first block whose bottom edge > visibleY0.
	// positions are sorted by Y (ascending) since SetBounds lays out
	// blocks top-to-bottom. We search for the first index where
	// positions[idx].Y + positions[idx].H > visibleY0.
	startIdx := sort.Search(len(c.positions), func(i int) bool {
		return c.positions[i].Y+c.positions[i].H > visibleY0
	})

	// Linearly paint from startIdx until we pass the visible range.
	for i := startIdx; i < len(c.blocks) && i < len(c.positions); i++ {
		pos := c.positions[i]
		if pos.Y >= visibleY1 {
			break // block is entirely below visible range
		}
		c.blocks[i].Paint(buf)
	}
}

// ensure BlockContainer satisfies component.VisiblePainter
var _ component.VisiblePainter = (*BlockContainer)(nil)
