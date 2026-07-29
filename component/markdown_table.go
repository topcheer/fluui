package component

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── MarkdownTable: Render Markdown Pipe Tables ───
//
// MarkdownTable parses and renders markdown pipe-delimited tables (GFM syntax)
// into a styled TUI table with column alignment. Supports left/center/right
// alignment via the separator row (|:---|:---:|---:|).
//
// Usage:
//
//	mt := NewMarkdownTable()
//	mt.SetMarkdown(`| Name | Age | City |
//	|:-----|----:|:----:|
//	| Alice | 30 | NYC |
//	| Bob | 25 | LA |`)
//	mt.Paint(buf)

// MarkdownTableStyle holds styling for MarkdownTable.
type MarkdownTableStyle struct {
	Header     buffer.Style
	Cell       buffer.Style
	Separator  buffer.Style
	Border     buffer.Style
	AltRow     buffer.Style // alternating row color
}

// DefaultMarkdownTableStyle returns sensible defaults.
func DefaultMarkdownTableStyle() MarkdownTableStyle {
	header := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold} // violet-400 bold
	cell := buffer.Style{Fg: buffer.RGB(226, 232, 240)}                        // slate-200
	sep := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                           // slate-600
	border := buffer.Style{Fg: buffer.RGB(51, 65, 85)}                         // slate-700
	alt := buffer.Style{Fg: buffer.RGB(203, 213, 225)}                         // slate-300
	return MarkdownTableStyle{Header: header, Cell: cell, Separator: sep, Border: border, AltRow: alt}
}

// columnAlign represents text alignment for a column.
type columnAlign int

const (
	alignLeft   columnAlign = 0
	alignCenter columnAlign = 1
	alignRight  columnAlign = 2
)

// MarkdownTable renders parsed markdown pipe tables.
type MarkdownTable struct {
	BaseComponent
	mu sync.Mutex

	rawSource string
	style     MarkdownTableStyle

	// cached parsed data
	cachedHeaders  []string
	cachedRows     [][]string
	cachedAligns   []columnAlign
	cachedColWidth []int
}

// NewMarkdownTable creates a MarkdownTable with defaults.
func NewMarkdownTable() *MarkdownTable {
	mt := &MarkdownTable{
		style: DefaultMarkdownTableStyle(),
	}
	mt.SetID(GenerateID("mdtable"))
	return mt
}

// SetMarkdown sets the raw markdown source and triggers parsing.
func (mt *MarkdownTable) SetMarkdown(source string) *MarkdownTable {
	mt.mu.Lock()
	mt.rawSource = source
	mt.parseLocked()
	mt.mu.Unlock()
	return mt
}

// Markdown returns the raw markdown source.
func (mt *MarkdownTable) Markdown() string {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	return mt.rawSource
}

// SetStyle sets the custom style.
func (mt *MarkdownTable) SetStyle(s MarkdownTableStyle) *MarkdownTable {
	mt.mu.Lock()
	mt.style = s
	mt.mu.Unlock()
	return mt
}

// RowCount returns the number of data rows (excluding header).
func (mt *MarkdownTable) RowCount() int {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	return len(mt.cachedRows)
}

// ColumnCount returns the number of columns.
func (mt *MarkdownTable) ColumnCount() int {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	return len(mt.cachedHeaders)
}

// parseLocked parses the markdown source. Caller must hold lock.
func (mt *MarkdownTable) parseLocked() {
	mt.cachedHeaders = mt.cachedHeaders[:0]
	mt.cachedRows = mt.cachedRows[:0]
	mt.cachedAligns = mt.cachedAligns[:0]
	mt.cachedColWidth = mt.cachedColWidth[:0]

	lines := strings.Split(mt.rawSource, "\n")
	rowIdx := 0
	separatorSeen := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !strings.Contains(line, "|") {
			continue
		}

		cells := parsePipeRow(line)
		if len(cells) == 0 {
			continue
		}

		if rowIdx == 0 {
			// Header row
			mt.cachedHeaders = cells
		} else if rowIdx == 1 && isSeparatorRow(line) {
			// Separator row — parse alignment
			mt.cachedAligns = make([]columnAlign, len(cells))
			for i, c := range cells {
				mt.cachedAligns[i] = parseAlign(c)
			}
			separatorSeen = true
		} else {
			mt.cachedRows = append(mt.cachedRows, cells)
		}
		rowIdx++
	}

	// Default alignment if no separator
	if !separatorSeen && len(mt.cachedAligns) == 0 {
		mt.cachedAligns = make([]columnAlign, len(mt.cachedHeaders))
		for i := range mt.cachedAligns {
			mt.cachedAligns[i] = alignLeft
		}
	}

	// Calculate column widths
	colCount := len(mt.cachedHeaders)
	if colCount == 0 {
		return
	}
	mt.cachedColWidth = make([]int, colCount)
	for i := 0; i < colCount; i++ {
		if i < len(mt.cachedHeaders) {
			mt.cachedColWidth[i] = len(mt.cachedHeaders[i])
		}
	}
	for _, row := range mt.cachedRows {
		for i, cell := range row {
			if i < colCount && len(cell) > mt.cachedColWidth[i] {
				mt.cachedColWidth[i] = len(cell)
			}
		}
	}
}

// parsePipeRow splits a pipe-delimited row into cells.
func parsePipeRow(line string) []string {
	// Trim leading/trailing pipes
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	return strings.Split(line, "|")
}

// isSeparatorRow checks if a line is a markdown table separator.
func isSeparatorRow(line string) bool {
	cleaned := strings.ReplaceAll(line, "|", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ":", "")
	return cleaned == ""
}

// parseAlign extracts alignment from a separator cell.
func parseAlign(cell string) columnAlign {
	cell = strings.TrimSpace(cell)
	leftCol := strings.HasPrefix(cell, ":")
	rightCol := strings.HasSuffix(cell, ":")
	if leftCol && rightCol {
		return alignCenter
	}
	if rightCol {
		return alignRight
	}
	return alignLeft
}

// Measure returns the preferred size.
func (mt *MarkdownTable) Measure(cs Constraints) Size {
	mt.mu.Lock()
	colCount := len(mt.cachedHeaders)
	rowCount := len(mt.cachedRows)
	totalW := 0
	for _, w := range mt.cachedColWidth {
		totalW += w + 3 // content + " | "
	}
	totalW += 1 // leading border
	mt.mu.Unlock()

	w := totalW
	if w < 20 {
		w = 20
	}
	h := rowCount + 4 // border + header + separator + rows + border
	if h < 5 {
		h = 5
	}
	_ = colCount
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the markdown table into the buffer.
func (mt *MarkdownTable) Paint(buf *buffer.Buffer) {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	b := mt.Bounds()
	x, y := b.X, b.Y

	if len(mt.cachedHeaders) == 0 {
		return
	}

	borderStyle := mt.style.Border
	headerStyle := mt.style.Header
	cellStyle := mt.style.Cell
	sepStyle := mt.style.Separator
	altStyle := mt.style.AltRow

	rowH := y
	colCount := len(mt.cachedHeaders)

	// Build column start positions
	colStarts := make([]int, colCount+1)
	colStarts[0] = x + 1
	for i := 0; i < colCount; i++ {
		colStarts[i+1] = colStarts[i] + mt.cachedColWidth[i] + 3 // "content | "
	}

	// Top border
	totalW := colStarts[colCount] - x
	for i := 0; i < totalW+1 && x+i < buf.Width; i++ {
		buf.SetCell(x+i, rowH, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
	}

	// Header row
	rowH++
	for i := 0; i < colCount; i++ {
		text := ""
		if i < len(mt.cachedHeaders) {
			text = mt.cachedHeaders[i]
		}
		mt.drawCellLocked(buf, colStarts[i], rowH, text, mt.cachedColWidth[i], i, headerStyle)
	}

	// Separator row
	rowH++
	for i := 0; i < colCount; i++ {
		align := alignLeft
		if i < len(mt.cachedAligns) {
			align = mt.cachedAligns[i]
		}
		sepText := ""
		switch align {
		case alignLeft:
			sepText = ":" + strings.Repeat("-", mt.cachedColWidth[i]-1)
		case alignRight:
			sepText = strings.Repeat("-", mt.cachedColWidth[i]-1) + ":"
		case alignCenter:
			if mt.cachedColWidth[i] >= 2 {
				sepText = ":" + strings.Repeat("-", mt.cachedColWidth[i]-2) + ":"
			} else {
				sepText = "--"
			}
		}
		mt.drawCellLocked(buf, colStarts[i], rowH, sepText, mt.cachedColWidth[i], i, sepStyle)
	}

	// Data rows
	for rIdx, row := range mt.cachedRows {
		rowH++
		if rowH >= buf.Height {
			break
		}
		// Alternating row color
		var style buffer.Style
		if rIdx%2 == 1 {
			style = altStyle
		} else {
			style = cellStyle
		}
		for i := 0; i < colCount; i++ {
			text := ""
			if i < len(row) {
				text = row[i]
			}
			mt.drawCellLocked(buf, colStarts[i], rowH, text, mt.cachedColWidth[i], i, style)
		}
	}

	// Bottom border
	rowH++
	for i := 0; i < totalW+1 && x+i < buf.Width; i++ {
		if rowH < buf.Height {
			buf.SetCell(x+i, rowH, buffer.Cell{Rune: '─', Fg: borderStyle.Fg, Bg: borderStyle.Bg, Flags: borderStyle.Flags, Width: 1})
		}
	}
}

// drawCellLocked draws a cell with alignment. Caller must hold lock.
func (mt *MarkdownTable) drawCellLocked(buf *buffer.Buffer, x, y int, text string, width, colIdx int, style buffer.Style) {
	align := alignLeft
	if colIdx < len(mt.cachedAligns) {
		align = mt.cachedAligns[colIdx]
	}

	// Truncate text to width
	textRunes := []rune(text)
	if len(textRunes) > width {
		textRunes = textRunes[:width]
	}

	// Calculate position based on alignment
	startOff := 0
	switch align {
	case alignRight:
		startOff = width - len(textRunes)
	case alignCenter:
		startOff = (width - len(textRunes)) / 2
	}

	for i := 0; i < width; i++ {
		if x+i >= buf.Width {
			break
		}
		var ch rune = ' '
		if i >= startOff && i < startOff+len(textRunes) {
			ch = textRunes[i-startOff]
		}
		buf.SetCell(x+i, y, buffer.Cell{Rune: ch, Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
	}

	// Pipe separator
	sepX := x + width
	if sepX < buf.Width {
		buf.SetCell(sepX, y, buffer.Cell{Rune: '|', Fg: style.Fg, Bg: style.Bg, Flags: style.Flags, Width: 1})
	}
}

// Children returns nil.
func (mt *MarkdownTable) Children() []Component { return nil }
