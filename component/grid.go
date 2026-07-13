package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Grid: 2D Layout with Row/Column Spans ───
//
// Grid is a 2D layout container that arranges children in a grid of rows
// and columns. Each child can span multiple rows and/or columns.
// Inspired by tview's Grid and CSS Grid Layout.
//
// Usage:
//	grid := NewGrid()
//	grid.SetRows(3, 0, 3)     // 3 rows: fixed top, flexible middle, fixed bottom
//	grid.SetColumns(20, 0, 20) // 3 cols: fixed left, flexible center, fixed right
//	grid.AddItem(sidebar, 0, 0, 3, 1)  // spans all rows, col 0
//	grid.AddItem(content, 0, 1, 2, 1)  // rows 0-1, col 1
//	grid.AddItem(statusbar, 2, 1, 1, 2) // row 2, cols 1-2

// GridItem holds a child component and its grid placement.
type GridItem struct {
	Component Component
	Row       int
	Col       int
	RowSpan   int
	ColSpan   int
}

// Grid is a 2D grid layout container with flexible rows/columns.
type Grid struct {
	mu       sync.RWMutex
	BaseComponent
	items     []GridItem
	rowSizes  []int // 0 = proportional (flex), >0 = fixed chars
	colSizes  []int // 0 = proportional (flex), >0 = fixed chars
	rowGap    int
	colGap    int
}

// NewGrid creates an empty grid.
func NewGrid() *Grid {
	return &Grid{}
}

// SetRows defines row sizes. 0 means proportional (takes remaining space),
// >0 means fixed character height.
func (g *Grid) SetRows(sizes ...int) {
	g.mu.Lock()
	g.rowSizes = sizes
	g.mu.Unlock()
}

// SetColumns defines column sizes. 0 means proportional, >0 means fixed width.
func (g *Grid) SetColumns(sizes ...int) {
	g.mu.Lock()
	g.colSizes = sizes
	g.mu.Unlock()
}

// SetRowGap sets vertical gap between rows (in characters).
func (g *Grid) SetRowGap(n int) {
	g.mu.Lock()
	g.rowGap = n
	g.mu.Unlock()
}

// SetColGap sets horizontal gap between columns (in characters).
func (g *Grid) SetColGap(n int) {
	g.mu.Lock()
	g.colGap = n
	g.mu.Unlock()
}

// AddItem places a component at (row, col) with optional span.
// RowSpan/ColSpan of 0 means span 1.
func (g *Grid) AddItem(child Component, row, col, rowSpan, colSpan int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if rowSpan <= 0 {
		rowSpan = 1
	}
	if colSpan <= 0 {
		colSpan = 1
	}
	g.items = append(g.items, GridItem{
		Component: child,
		Row:       row,
		Col:       col,
		RowSpan:   rowSpan,
		ColSpan:   colSpan,
	})
}

// RemoveItem removes a component by reference.
func (g *Grid) RemoveItem(child Component) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, item := range g.items {
		if item.Component == child {
			g.items = append(g.items[:i], g.items[i+1:]...)
			return
		}
	}
}

// Clear removes all items.
func (g *Grid) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.items = nil
}

// Items returns a copy of all grid items.
func (g *Grid) Items() []GridItem {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]GridItem, len(g.items))
	copy(result, g.items)
	return result
}

// ItemCount returns the number of items.
func (g *Grid) ItemCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.items)
}

func (g *Grid) Measure(constraints Constraints) Size {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return Size{W: constraints.MaxWidth, H: constraints.MaxHeight}
}

func (g *Grid) computeLayoutLocked(w, h int) (rowHeights, colWidths []int) {
	numRows := len(g.rowSizes)
	numCols := len(g.colSizes)

	// If no rows/cols defined, auto-detect from items
	if numRows == 0 {
		maxRow := 0
		for _, item := range g.items {
			end := item.Row + item.RowSpan
			if end > maxRow {
				maxRow = end
			}
		}
		numRows = maxRow
		g.rowSizes = make([]int, numRows) // all proportional
	}
	if numCols == 0 {
		maxCol := 0
		for _, item := range g.items {
			end := item.Col + item.ColSpan
			if end > maxCol {
				maxCol = end
			}
		}
		numCols = maxCol
		g.colSizes = make([]int, numCols) // all proportional
	}

	// Compute column widths
	colWidths = make([]int, numCols)
	fixedColW := 0
	flexCols := 0
	for i, s := range g.colSizes {
		if i >= numCols {
			break
		}
		if s > 0 {
			colWidths[i] = s
			fixedColW += s
		} else {
			flexCols++
		}
	}
	// Account for gaps
	totalGapW := g.colGap * (numCols - 1)
	if totalGapW < 0 {
		totalGapW = 0
	}
	availFlexW := w - fixedColW - totalGapW
	if availFlexW < 0 {
		availFlexW = 0
	}
	flexW := 0
	if flexCols > 0 {
		flexW = availFlexW / flexCols
	}
	for i, s := range g.colSizes {
		if i >= numCols {
			break
		}
		if s == 0 {
			colWidths[i] = flexW
		}
	}

	// Compute row heights
	rowHeights = make([]int, numRows)
	fixedRowH := 0
	flexRows := 0
	for i, s := range g.rowSizes {
		if i >= numRows {
			break
		}
		if s > 0 {
			rowHeights[i] = s
			fixedRowH += s
		} else {
			flexRows++
		}
	}
	totalGapH := g.rowGap * (numRows - 1)
	if totalGapH < 0 {
		totalGapH = 0
	}
	availFlexH := h - fixedRowH - totalGapH
	if availFlexH < 0 {
		availFlexH = 0
	}
	flexH := 0
	if flexRows > 0 {
		flexH = availFlexH / flexRows
	}
	for i, s := range g.rowSizes {
		if i >= numRows {
			break
		}
		if s == 0 {
			rowHeights[i] = flexH
		}
	}

	return rowHeights, colWidths
}

func (g *Grid) Paint(buf *buffer.Buffer) {
	g.mu.Lock()
	defer g.mu.Unlock()

	bounds := g.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	rowHeights, colWidths := g.computeLayoutLocked(bounds.W, bounds.H)

	// Compute cumulative positions
	colX := make([]int, len(colWidths)+1)
	x := 0
	for i, w := range colWidths {
		colX[i] = x
		x += w + g.colGap
	}
	colX[len(colWidths)] = x

	rowY := make([]int, len(rowHeights)+1)
	y := 0
	for i, h := range rowHeights {
		rowY[i] = y
		y += h + g.rowGap
	}
	rowY[len(rowHeights)] = y

	// Paint each item
	for _, item := range g.items {
		if item.Component == nil {
			continue
		}

		startRow := item.Row
		if startRow >= len(rowHeights) {
			continue
		}
		startCol := item.Col
		if startCol >= len(colWidths) {
			continue
		}

		// Compute cell bounds
		cellX := bounds.X + colX[startCol]
		cellY := bounds.Y + rowY[startRow]

		cellW := 0
		endCol := startCol + item.ColSpan
		if endCol > len(colWidths) {
			endCol = len(colWidths)
		}
		for c := startCol; c < endCol; c++ {
			cellW += colWidths[c]
			if c < endCol-1 {
				cellW += g.colGap
			}
		}

		cellH := 0
		endRow := startRow + item.RowSpan
		if endRow > len(rowHeights) {
			endRow = len(rowHeights)
		}
		for r := startRow; r < endRow; r++ {
			cellH += rowHeights[r]
			if r < endRow-1 {
				cellH += g.rowGap
			}
		}

		if cellW <= 0 || cellH <= 0 {
			continue
		}

		item.Component.SetBounds(Rect{
			X: cellX,
			Y: cellY,
			W: cellW,
			H: cellH,
		})
		item.Component.Paint(buf)
	}
}

func (g *Grid) Children() []Component {
	g.mu.RLock()
	defer g.mu.RUnlock()
	result := make([]Component, 0, len(g.items))
	for _, item := range g.items {
		if item.Component != nil {
			result = append(result, item.Component)
		}
	}
	return result
}