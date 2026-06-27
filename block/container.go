package block

import (
	"sync"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// BlockContainer manages a list of blocks and handles layout.
// It is itself a Component so it can be placed inside a ScrollView.
type BlockContainer struct {
	component.BaseComponent

	mu          sync.RWMutex
	blocks      []Block
	spacing     int // pixels between blocks
	dirty       bool
	positions   []BlockPosition // cached during SetBounds
	totalHeight int             // cached during Measure
}

// NewBlockContainer creates an empty container with default spacing of 1.
func NewBlockContainer() *BlockContainer {
	return &BlockContainer{
		spacing: 1,
		dirty:   true,
	}
}

// AddBlock appends a block to the container.
func (c *BlockContainer) AddBlock(b Block) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blocks = append(c.blocks, b)
	c.dirty = true
}

// RemoveBlock removes the block with the given ID. Returns true if found.
func (c *BlockContainer) RemoveBlock(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, b := range c.blocks {
		if b.ID() == id {
			c.blocks = append(c.blocks[:i], c.blocks[i+1:]...)
			c.dirty = true
			return true
		}
	}
	return false
}

// LastBlock returns the most recently added block, or nil if empty.
func (c *BlockContainer) LastBlock() Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.blocks) == 0 {
		return nil
	}
	return c.blocks[len(c.blocks)-1]
}

// Blocks returns a copy of all blocks in order.
func (c *BlockContainer) Blocks() []Block {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]Block, len(c.blocks))
	copy(result, c.blocks)
	return result
}

// Len returns the number of blocks.
func (c *BlockContainer) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.blocks)
}

// SetSpacing sets the vertical spacing between blocks.
func (c *BlockContainer) SetSpacing(s int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.spacing = s
	c.dirty = true
}

// Measure calculates the total size of all blocks plus spacing.
func (c *BlockContainer) Measure(cs component.Constraints) component.Size {
	c.mu.RLock()
	defer c.mu.RUnlock()

	maxBlockW := 0
	totalH := 0

	for i, b := range c.blocks {
		sz := b.Measure(component.Unbounded())
		if sz.W > maxBlockW {
			maxBlockW = sz.W
		}
		totalH += sz.H
		if i < len(c.blocks)-1 {
			totalH += c.spacing
		}
	}

	c.totalHeight = totalH // cache for virtual scrolling

	w := maxBlockW
	h := totalH

	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	return component.Size{W: w, H: h}
}

// SetBounds lays out all blocks vertically within the given bounds.
func (c *BlockContainer) SetBounds(r component.Rect) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.BaseComponent.SetBounds(r)

	// Reallocate positions cache
	c.positions = make([]BlockPosition, len(c.blocks))

	// Give each block the full width and stack them vertically.
	y := r.Y
	localY := 0 // track Y in container-local coordinates
	for i, b := range c.blocks {
		sz := b.Measure(component.Constraints{MaxWidth: r.W})
		blockH := sz.H
		b.SetBounds(component.Rect{
			X: r.X,
			Y: y,
			W: r.W,
			H: blockH,
		})
		c.positions[i] = BlockPosition{Y: localY, H: blockH}
		y += blockH
		localY += blockH
		if i < len(c.blocks)-1 {
			y += c.spacing
			localY += c.spacing
		}
	}
}

// Paint renders all blocks into the buffer at their computed positions.
func (c *BlockContainer) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, b := range c.blocks {
		b.Paint(buf)
	}
}

// Children returns all blocks as children (for component tree traversal).
func (c *BlockContainer) Children() []component.Component {
	c.mu.RLock()
	defer c.mu.RUnlock()
	children := make([]component.Component, len(c.blocks))
	for i, b := range c.blocks {
		children[i] = b
	}
	return children
}

// IsDirty returns true if the container or any child block is dirty.
func (c *BlockContainer) IsDirty() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.dirty {
		return true
	}
	for _, b := range c.blocks {
		if b.IsDirty() {
			return true
		}
	}
	return false
}

// ClearDirty clears dirty flags on the container and all children.
func (c *BlockContainer) ClearDirty() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dirty = false
	for _, b := range c.blocks {
		b.ClearDirty()
	}
}
