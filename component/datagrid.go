package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// DataGridCell holds a single cell's value and optional style override.
type DataGridCell struct {
	Value string
}

// DataGridColumn defines a column in the grid.
type DataGridColumn struct {
	Title string
	Width int // 0 = auto
}

// DataGrid renders tabular data with headers, zebra striping, and
// horizontal scrolling. Unlike Table (which is form-oriented), DataGrid
// is optimized for large read-only datasets like logs, metrics, and
// AI conversation history.
//
// Thread-safe.
type DataGrid struct {
	BaseComponent
	mu      sync.Mutex
	columns []DataGridColumn
	rows    [][]DataGridCell
	scrollX int
	scrollY int
	cursorRow int
}

// NewDataGrid creates a grid with the given columns.
func NewDataGrid(columns []DataGridColumn) *DataGrid {
	return &DataGrid{
		BaseComponent: BaseComponent{id: GenerateID("datagrid")},
		columns:       columns,
	}
}

// SetColumns replaces columns.
func (g *DataGrid) SetColumns(cols []DataGridColumn) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.columns = cols
}

// SetRows replaces all data rows.
func (g *DataGrid) SetRows(rows [][]DataGridCell) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rows = rows
}

// AddRow appends a single row.
func (g *DataGrid) AddRow(row []DataGridCell) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rows = append(g.rows, row)
}

// RowCount returns the number of data rows.
func (g *DataGrid) RowCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return len(g.rows)
}

// ScrollLeft shifts the view left.
func (g *DataGrid) ScrollLeft(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.scrollX -= n
	if g.scrollX < 0 {
		g.scrollX = 0
	}
}

// ScrollRight shifts the view right.
func (g *DataGrid) ScrollRight(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.scrollX += n
}

// ScrollUp shifts the view up.
func (g *DataGrid) ScrollUp(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.scrollY -= n
	if g.scrollY < 0 {
		g.scrollY = 0
	}
}

// ScrollDown shifts the view down.
func (g *DataGrid) ScrollDown(n int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.scrollY += n
	max := g.maxScrollY()
	if g.scrollY > max {
		g.scrollY = max
	}
}

// CursorRow returns the highlighted row index (-1 = none).
func (g *DataGrid) CursorRow() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.cursorRow
}

// SetCursorRow sets the highlighted row.
func (g *DataGrid) SetCursorRow(idx int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cursorRow = idx
}

// Clear removes all rows.
func (g *DataGrid) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rows = nil
	g.scrollX = 0
	g.scrollY = 0
	g.cursorRow = -1
}

func (g *DataGrid) maxScrollY() int {
	if len(g.rows) == 0 {
		return 0
	}
	return len(g.rows) - 1
}

// colWidth returns the width for a column (default 10 if 0).
func (g *DataGrid) colWidth(idx int) int {
	if idx >= len(g.columns) {
		return 10
	}
	w := g.columns[idx].Width
	if w <= 0 {
		return 10
	}
	return w
}

// Measure returns the desired size.
func (g *DataGrid) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 20
	}
	return Size{W: maxW, H: maxH}
}

// Paint renders the data grid.
func (g *DataGrid) Paint(buf *buffer.Buffer) {
	g.mu.Lock()
	columns := g.columns
	rows := g.rows
	scrollX := g.scrollX
	scrollY := g.scrollY
	cursor := g.cursorRow
	g.mu.Unlock()

	b := g.Bounds()
	if b.W <= 0 || b.H <= 0 || len(columns) == 0 {
		return
	}

	th := theme.Get()
	headerStyle := buffer.Style{Fg: th.Accent, Bg: th.CodeBg}
	normalStyle := buffer.Style{Fg: th.Fg}
	zebraStyle := buffer.Style{Fg: th.Fg, Bg: th.CodeBg}
	cursorStyle := buffer.Style{Fg: th.Bg, Bg: th.Accent}
	mutedStyle := buffer.Style{Fg: th.Muted}

	// Header row
	y := b.Y
	x := b.X
	startCol := scrollX
	if startCol >= len(columns) {
		startCol = 0
	}

	for ci := startCol; ci < len(columns); ci++ {
		cw := g.colWidth(ci)
		if x+cw > b.X+b.W {
			cw = b.X + b.W - x
			if cw <= 0 {
				break
			}
		}
		title := columns[ci].Title
		if utf8.RuneCountInString(title) > cw {
			title = truncateRunes(title, cw-1) + "\u2026"
		}
		// Fill header background
		for i := 0; i < cw; i++ {
			buf.SetCell(x+i, y, buffer.Cell{Rune: ' ', Width: 1, Bg: th.CodeBg})
		}
		buf.DrawText(x, y, title, headerStyle)
		x += cw + 1 // +1 gap
		if x >= b.X+b.W {
			break
		}
	}

	// Data rows
	dataTop := b.Y + 1
	dataH := b.H - 1
	if dataH < 0 {
		dataH = 0
	}

	for vi := 0; vi < dataH; vi++ {
		rowIdx := scrollY + vi
		if rowIdx >= len(rows) {
			break
		}
		y := dataTop + vi
		row := rows[rowIdx]

		style := normalStyle
		if vi%2 == 1 {
			style = zebraStyle
		}
		if rowIdx == cursor {
			style = cursorStyle
		}

		x = b.X
		for ci := startCol; ci < len(columns) && ci < len(row); ci++ {
			cw := g.colWidth(ci)
			if x+cw > b.X+b.W {
				cw = b.X + b.W - x
				if cw <= 0 {
					break
				}
			}

			val := row[ci].Value
			if utf8.RuneCountInString(val) > cw {
				val = truncateRunes(val, cw-1) + "\u2026"
			}
			buf.DrawText(x, y, val, style)

			// Pad remaining width
			drawn := utf8.RuneCountInString(val)
			for i := drawn; i < cw; i++ {
				buf.SetCell(x+i, y, buffer.Cell{Rune: ' ', Width: 1, Fg: style.Fg, Bg: style.Bg})
			}

			x += cw + 1
			if x >= b.X+b.W {
				break
			}
		}
	}

	// Scroll indicators
	if scrollX > 0 && b.W > 2 {
		buf.DrawText(b.X, b.Y, "\u25c0", mutedStyle) // ◀
	}
	if scrollY > 0 && b.H > 2 {
		buf.DrawText(b.X+b.W-1, b.Y, "\u25b2", mutedStyle) // ▲
	}
}
