package component

import (
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// Alignment specifies horizontal text alignment within a column.
type Alignment int

const (
	AlignLeft   Alignment = 0
	AlignCenter Alignment = 1
	AlignRight  Alignment = 2
)

// TableColumn holds metadata for a single column.
type TableColumn struct {
	Title string
	Align Alignment
	Width int // computed column width (max of header and all cell values)
}

// Table is a scrollable, sortable table widget with headers, zebra striping,
// and keyboard navigation.
//
// It implements the Component interface (ID, Measure, SetBounds, Bounds, Paint, Children)
// and provides HandleKey for keyboard interaction.
type Table struct {
	BaseComponent

	mu sync.RWMutex

	headers []string
	rows    [][]string
	columns []TableColumn

	// selection
	selectedRow int

	// scrolling
	scrollY int // vertical scroll offset (row index of first visible data row)
	scrollX int // horizontal scroll offset (column index of first visible column)

	// sorting
	sortCol       int  // column index currently sorted, -1 = unsorted
	sortAscending bool // true = ascending

	// column alignment overrides (key = column index)
	colAligns map[int]Alignment

	// display options
	zebra       bool
	showRowNum  bool

	// filtering
	filter       string // case-insensitive substring filter
	filteredRows [][]string // computed subset when filter is active

	// callback
	onSelect func(row int, rowData []string)

	// cached bounds
	boundsW int
	boundsH int
}

// NewTable creates a Table with the given headers and optional initial rows.
func NewTable(headers []string, rows ...[]string) *Table {
	t := &Table{
		headers:      make([]string, len(headers)),
		sortCol:      -1,
		colAligns:    make(map[int]Alignment),
		zebra:        true,
		selectedRow:  0,
		scrollY:      0,
		scrollX:      0,
	}
	copy(t.headers, headers)

	// Deep-copy rows.
	for _, r := range rows {
		row := make([]string, len(r))
		copy(row, r)
		t.rows = append(t.rows, row)
	}

	t.recomputeColumns()
	return t
}

// --- Data accessors ---

// Headers returns a copy of the column headers.
func (t *Table) Headers() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]string, len(t.headers))
	copy(result, t.headers)
	return result
}

// Rows returns a deep copy of all data rows.
func (t *Table) Rows() [][]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([][]string, len(t.rows))
	for i, r := range t.rows {
		result[i] = make([]string, len(r))
		copy(result[i], r)
	}
	return result
}

// SetHeaders replaces all headers.
func (t *Table) SetHeaders(headers []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.headers = make([]string, len(headers))
	copy(t.headers, headers)
	t.recomputeColumnsLocked()
}

// SetRows replaces all data rows.
func (t *Table) SetRows(rows [][]string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rows = t.rows[:0]
	for _, r := range rows {
		row := make([]string, len(r))
		copy(row, r)
		t.rows = append(t.rows, row)
	}
	t.recomputeColumnsLocked()
	t.applyFilterLocked()
	t.clampSelectionLocked()
}

// AddRow appends a single data row.
func (t *Table) AddRow(row []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	r := make([]string, len(row))
	copy(r, row)
	t.rows = append(t.rows, r)
	t.recomputeColumnsLocked()
	t.applyFilterLocked()
}

// RowCount returns the number of visible rows (filtered if filter active).
func (t *Table) RowCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.displayRowsLocked())
}

// SelectedRow returns the index of the currently selected row.
func (t *Table) SelectedRow() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.selectedRow
}

// SelectedRowData returns a copy of the currently selected row's data.
func (t *Table) SelectedRowData() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	rows := t.displayRowsLocked()
	if t.selectedRow < 0 || t.selectedRow >= len(rows) {
		return nil
	}
	row := make([]string, len(rows[t.selectedRow]))
	copy(row, rows[t.selectedRow])
	return row
}

// SetSelectedRow sets the selection, clamping to valid range, and adjusts scroll.
func (t *Table) SetSelectedRow(idx int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.selectedRow = idx
	t.clampSelectionLocked()
}

// --- Alignment ---

// SetColumnAlign sets the alignment for a specific column.
func (t *Table) SetColumnAlign(col int, align Alignment) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.colAligns[col] = align
}

// ColumnAlign returns the alignment for a column (default AlignLeft).
func (t *Table) ColumnAlign(col int) Alignment {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if a, ok := t.colAligns[col]; ok {
		return a
	}
	return AlignLeft
}

// Columns returns a copy of the computed column metadata.
func (t *Table) Columns() []TableColumn {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]TableColumn, len(t.columns))
	copy(result, t.columns)
	return result
}

// --- Display options ---

// SetZebra enables or disables zebra striping.
func (t *Table) SetZebra(enable bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.zebra = enable
}

// Zebra returns whether zebra striping is enabled.
func (t *Table) Zebra() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.zebra
}

// SetShowRowNumbers toggles row number display.
func (t *Table) SetShowRowNumbers(enable bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.showRowNum = enable
}

// --- Sorting ---

// SortBy sorts the table by the given column index.
func (t *Table) SortBy(col int, ascending bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if col < 0 || col >= len(t.headers) {
		return
	}
	t.sortCol = col
	t.sortAscending = ascending
	sort.SliceStable(t.rows, func(i, j int) bool {
		a, b := "", ""
		if col < len(t.rows[i]) {
			a = t.rows[i][col]
		}
		if col < len(t.rows[j]) {
			b = t.rows[j][col]
		}
		if ascending {
			return a < b
		}
		return a > b
	})
}

// SortCol returns the currently sorted column (-1 if unsorted).
func (t *Table) SortCol() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.sortCol
}

// SortAscending returns true if the current sort is ascending.
func (t *Table) SortAscending() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.sortAscending
}

// ToggleSort sorts by the given column. If already sorted on that column,
// toggles the sort direction. Otherwise sorts ascending.
func (t *Table) ToggleSort(col int) {
	t.mu.RLock()
	if t.sortCol == col {
		asc := !t.sortAscending
		t.mu.RUnlock()
		t.SortBy(col, asc)
		return
	}
	t.mu.RUnlock()
	t.SortBy(col, true)
}

// --- Callback ---

// OnSelect sets the callback invoked when the user presses Enter on a row.
func (t *Table) OnSelect(fn func(row int, rowData []string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onSelect = fn
}

// --- Scrolling ---

// ScrollY returns the current vertical scroll offset.
func (t *Table) ScrollY() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.scrollY
}

// ScrollX returns the current horizontal scroll offset (column index).
func (t *Table) ScrollX() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.scrollX
}

// ScrollUp moves the viewport up by n rows.
func (t *Table) ScrollUp(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.scrollY -= n
	if t.scrollY < 0 {
		t.scrollY = 0
	}
}

// ScrollDown moves the viewport down by n rows.
func (t *Table) ScrollDown(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.scrollY += n
	t.clampScrollYLocked()
}

// ensureVisibleLocked adjusts scrollY so the selected row is visible.
func (t *Table) ensureVisibleLocked(viewportH int) {
	if viewportH <= 0 {
		return
	}
	// Header takes 1 row, so data viewport = viewportH - 1.
	dataH := viewportH - 1
	if dataH <= 0 {
		return
	}

	if t.selectedRow < t.scrollY {
		t.scrollY = t.selectedRow
	}
	if t.selectedRow >= t.scrollY+dataH {
		t.scrollY = t.selectedRow - dataH + 1
	}
	if t.scrollY < 0 {
		t.scrollY = 0
	}
}

// --- Keyboard handling ---

// HandleKey processes keyboard navigation. Returns true if the key was consumed.
func (t *Table) HandleKey(key *term.KeyEvent) bool {
	if key == nil {
		return false
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	switch key.Key {
	case term.KeyUp:
		if t.selectedRow > 0 {
			t.selectedRow--
		}
		t.ensureVisibleLocked(t.boundsH)
		return true

	case term.KeyDown:
		rows := t.displayRowsLocked()
		if t.selectedRow < len(rows)-1 {
			t.selectedRow++
		}
		t.ensureVisibleLocked(t.boundsH)
		return true

	case term.KeyLeft:
		if t.scrollX > 0 {
			t.scrollX--
		}
		return true

	case term.KeyRight:
		maxScroll := len(t.columns) - 1
		if t.scrollX < maxScroll {
			t.scrollX++
		}
		return true

	case term.KeyHome:
		t.selectedRow = 0
		t.scrollY = 0
		return true

	case term.KeyEnd:
		rows := t.displayRowsLocked()
		if len(rows) > 0 {
			t.selectedRow = len(rows) - 1
		}
		t.ensureVisibleLocked(t.boundsH)
		return true

	case term.KeyPageUp:
		dataH := t.boundsH - 1
		if dataH <= 0 {
			dataH = 1
		}
		t.selectedRow -= dataH
		if t.selectedRow < 0 {
			t.selectedRow = 0
		}
		t.scrollY -= dataH
		if t.scrollY < 0 {
			t.scrollY = 0
		}
		return true

	case term.KeyPageDown:
		dataH := t.boundsH - 1
		if dataH <= 0 {
			dataH = 1
		}
		t.selectedRow += dataH
		rows := t.displayRowsLocked()
		if t.selectedRow >= len(rows) {
			t.selectedRow = len(rows) - 1
		}
		if t.selectedRow < 0 {
			t.selectedRow = 0
		}
		t.ensureVisibleLocked(t.boundsH)
		return true

	case term.KeyEnter:
		rows := t.displayRowsLocked()
		if t.onSelect != nil && t.selectedRow >= 0 && t.selectedRow < len(rows) {
			row := make([]string, len(rows[t.selectedRow]))
			copy(row, rows[t.selectedRow])
			cb := t.onSelect
			t.mu.Unlock()
			cb(t.selectedRow, row)
			t.mu.Lock()
		}
		return true

	default:
		// Check for sort shortcuts: Ctrl+1 through Ctrl+9
		if key.Modifiers&term.ModCtrl != 0 && key.Key == term.KeyUnknown {
			if key.Rune >= '1' && key.Rune <= '9' {
				col := int(key.Rune - '1')
				if col >= 0 && col < len(t.headers) {
					if t.sortCol == col {
						t.sortAscending = !t.sortAscending
					} else {
						t.sortCol = col
						t.sortAscending = true
					}
					t.sortRowsLocked()
				}
				return true
			}
		}
		return false
	}
}

// --- Component interface ---

// Measure returns the desired size of the table.
// Width is the sum of all column widths plus separators.
// Height is header (1) + row count.
func (t *Table) Measure(cs Constraints) Size {
	t.mu.RLock()
	defer t.mu.RUnlock()

	w := t.totalWidthLocked()
	h := 1 + len(t.displayRowsLocked())

	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// SetBounds sets the component bounds and adjusts scroll to keep selection visible.
func (t *Table) SetBounds(r Rect) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bounds = r
	t.boundsW = r.W
	t.boundsH = r.H
	t.clampSelectionLocked()
	t.ensureVisibleLocked(r.H)
}

// Paint renders the table into the given buffer.
func (t *Table) Paint(buf *buffer.Buffer) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	x, y := t.bounds.X, t.bounds.Y
	w, h := t.bounds.W, t.bounds.H
	if w < 1 || h < 1 {
		return
	}

	th := theme.Get()

	// --- Header row ---
	headerStyle := buffer.Style{
		Fg:    th.Accent,
		Bg:    th.Bg,
		Flags: buffer.Bold,
	}

	colX := x
	visibleCols := t.visibleColumnsLocked(w)

	for _, ci := range visibleCols {
		col := t.columns[ci]
		title := t.truncateToWidth(col.Title, col.Width)
		t.drawCellLocked(buf, colX, y, title, col.Width, ci, headerStyle)
		colX += int(col.Width) + 1 // +1 for separator space
		if colX >= x+w {
			break
		}
	}

	// Sort indicator on the sorted column header.
	if t.sortCol >= 0 && t.sortCol < len(t.columns) {
		arrow := " ▲"
		if !t.sortAscending {
			arrow = " ▼"
		}
		indicatorX := x
		for i := 0; i < t.sortCol; i++ {
			indicatorX += int(t.columns[i].Width) + 1
		}
		indicatorX += int(t.columns[t.sortCol].Width) - 1
		if indicatorX < x+w-1 {
			for _, r := range arrow {
				buf.SetCell(indicatorX, y, buffer.Cell{
					Rune:  r,
					Width: 1,
					Fg:    th.Accent,
					Bg:    th.Bg,
					Flags: buffer.Bold,
				})
				indicatorX++
			}
		}
	}

	// --- Separator line below header ---
	if h >= 2 {
		sepY := y + 1
		sepStyle := buffer.Style{Fg: th.Border, Bg: th.Bg}
		for i := 0; i < w; i++ {
			buf.SetCell(x+i, sepY, buffer.Cell{
				Rune:  '─',
				Width: 1,
				Fg:    sepStyle.Fg,
				Bg:    sepStyle.Bg,
			})
		}
	}

	// --- Data rows ---
	rows := t.displayRowsLocked()
	dataStartY := y + 2 // header + separator
	dataH := h - 2
	if dataH < 0 {
		dataH = 0
	}

	rowStyle := buffer.Style{Fg: th.Fg, Bg: th.Bg}
	selectedStyle := buffer.Style{Fg: th.Bg, Bg: th.Accent, Flags: buffer.Bold}
	zebraStyle := buffer.Style{Fg: th.Fg, Bg: th.CodeBg}

	for vi := 0; vi < dataH; vi++ {
		rowIdx := t.scrollY + vi
		if rowIdx >= len(rows) {
			break
		}

		rowY := dataStartY + vi
		isSelected := rowIdx == t.selectedRow
		isZebra := t.zebra && rowIdx%2 == 1

		colX = x
		for _, ci := range visibleCols {
			col := t.columns[ci]
				cellText := ""
			if ci < len(rows[rowIdx]) {
				cellText = rows[rowIdx][ci]
			}
			cellText = t.truncateToWidth(cellText, col.Width)

			var style buffer.Style
			switch {
			case isSelected:
				style = selectedStyle
			case isZebra:
				style = zebraStyle
			default:
				style = rowStyle
			}

			t.drawCellLocked(buf, colX, rowY, cellText, col.Width, ci, style)
			colX += int(col.Width) + 1
			if colX >= x+w {
				break
			}
		}

		// Fill rest of the row with background.
		for i := colX; i < x+w; i++ {
			if i < x+w {
				var bg buffer.Color
				switch {
				case isSelected:
					bg = selectedStyle.Bg
				case isZebra:
					bg = zebraStyle.Bg
				default:
					bg = rowStyle.Bg
				}
				buf.SetCell(i, rowY, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
			}
		}
	}

	// --- Scrollbar indicator (if more rows than viewport) ---
	if len(t.rows) > dataH && dataH > 0 && w >= 3 {
		sbX := x + w - 1
		sbH := dataH
		for i := 0; i < sbH; i++ {
			ratio := float64(i) / float64(sbH)
			rowAtScroll := t.scrollY + i
			trackRune := '░'
			_ = ratio

			if rowAtScroll >= t.scrollY && rowAtScroll < t.scrollY+sbH {
				if rowAtScroll == t.selectedRow-t.scrollY+t.scrollY {
					// This is just the track; thumb is computed below
				}
			}

			buf.SetCell(sbX, dataStartY+i, buffer.Cell{
				Rune:  trackRune,
				Width: 1,
				Fg:    th.BorderMuted,
				Bg:    th.Bg,
			})
		}

		// Thumb position.
		if len(t.rows) > 0 {
			thumbRatio := float64(t.scrollY) / float64(len(t.rows))
			thumbPos := int(thumbRatio * float64(sbH))
			if thumbPos >= sbH {
				thumbPos = sbH - 1
			}
			buf.SetCell(sbX, dataStartY+thumbPos, buffer.Cell{
				Rune:  '█',
				Width: 1,
				Fg:    th.Accent,
				Bg:    th.Bg,
			})
		}
	}
}

// Children returns nil (Table is a leaf component).
func (t *Table) Children() []Component { return nil }

// --- Filtering ---

// SetFilter applies a case-insensitive substring filter to the table.
// Only rows where any column contains the query are shown.
// An empty string clears the filter.
func (t *Table) SetFilter(query string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.filter = query
	t.applyFilterLocked()
	t.selectedRow = 0
	t.scrollY = 0
}

// Filter returns the current filter string (empty = no filter).
func (t *Table) Filter() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.filter
}

// ClearFilter removes any active filter.
func (t *Table) ClearFilter() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.filter = ""
	t.filteredRows = nil
	t.selectedRow = 0
	t.scrollY = 0
}

// IsFiltered returns true if a filter is currently active.
func (t *Table) IsFiltered() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.filter != ""
}

// displayRowsLocked returns the rows that should be displayed (filtered or all).
func (t *Table) displayRowsLocked() [][]string {
	if t.filter != "" {
		return t.filteredRows
	}
	return t.rows
}

// applyFilterLocked recomputes filteredRows from the current filter.
func (t *Table) applyFilterLocked() {
	if t.filter == "" {
		t.filteredRows = nil
		return
	}
	t.filteredRows = t.filteredRows[:0]
	for _, row := range t.rows {
		if rowMatchesFilter(row, t.filter) {
			t.filteredRows = append(t.filteredRows, row)
		}
	}
}

// rowMatchesFilter checks if any cell in the row contains the query.
func rowMatchesFilter(row []string, query string) bool {
	q := strings.ToLower(query)
	for _, cell := range row {
		if strings.Contains(strings.ToLower(cell), q) {
			return true
		}
	}
	return false
}

// --- Internal helpers ---

// recomputeColumns recomputes column widths from headers and rows.
// Must NOT be called under lock — this version acquires the lock.
func (t *Table) recomputeColumns() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.recomputeColumnsLocked()
}

// recomputeColumnsLocked recomputes column widths. Caller must hold the lock.
func (t *Table) recomputeColumnsLocked() {
	numCols := len(t.headers)
	// Ensure at least enough columns for all rows.
	for _, r := range t.rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}

	t.columns = make([]TableColumn, numCols)
	for i := 0; i < numCols; i++ {
		w := 0
		if i < len(t.headers) {
			w = buffer.StringWidth(t.headers[i])
		}
		for _, r := range t.rows {
			if i < len(r) {
				rw := buffer.StringWidth(r[i])
				if rw > w {
					w = rw
				}
			}
		}
		title := ""
		if i < len(t.headers) {
			title = t.headers[i]
		}
		align := AlignLeft
		if a, ok := t.colAligns[i]; ok {
			align = a
		}
		// Minimum column width of 3.
		if w < 3 {
			w = 3
		}
		t.columns[i] = TableColumn{
			Title: title,
			Align: align,
			Width: w,
		}
	}
}

// totalWidthLocked returns the total width including separators.
func (t *Table) totalWidthLocked() int {
	w := 0
	for i, c := range t.columns {
		w += int(c.Width)
		if i < len(t.columns)-1 {
			w++ // separator
		}
	}
	return w
}

// visibleColumnsLocked returns the indices of columns visible given the current scrollX and width.
func (t *Table) visibleColumnsLocked(maxW int) []int {
	if len(t.columns) == 0 {
		return nil
	}
	start := t.scrollX
	if start < 0 {
		start = 0
	}
	if start >= len(t.columns) {
		return nil
	}

	var result []int
	usedW := 0
	for i := start; i < len(t.columns); i++ {
		colW := int(t.columns[i].Width) + 1 // +1 for separator
		if usedW+colW > maxW && len(result) > 0 {
			break
		}
		result = append(result, i)
		usedW += colW
		if usedW >= maxW {
			break
		}
	}
	return result
}

// drawCellLocked draws a single cell with the column's alignment.
func (t *Table) drawCellLocked(buf *buffer.Buffer, x, y int, text string, colWidth int, colIdx int, style buffer.Style) {
	align := AlignLeft
	if colIdx < len(t.columns) {
		align = t.columns[colIdx].Align
	}

	textW := buffer.StringWidth(text)
	if textW > colWidth {
		textW = colWidth
	}

	var startX int
	switch align {
	case AlignCenter:
		startX = x + (colWidth-textW)/2
	case AlignRight:
		startX = x + colWidth - textW
	default:
		startX = x
	}

	if startX < x {
		startX = x
	}

	// Fill background for the full column width.
	for i := x; i < x+colWidth && i < buf.Width; i++ {
		buf.SetCell(i, y, buffer.Cell{Rune: ' ', Width: 1, Bg: style.Bg, Fg: style.Fg, Flags: style.Flags})
	}

	// Draw text.
	buf.DrawTextClamped(startX, y, text, style)
}

// truncateToWidth truncates text to fit within width display columns.
func (t *Table) truncateToWidth(text string, width int) string {
	if width <= 0 {
		return ""
	}
	w := buffer.StringWidth(text)
	if w <= width {
		return text
	}
	// Truncate character by character.
	var result []rune
	curW := 0
	for _, r := range text {
		rw := buffer.RuneWidth(r)
		if curW+rw > width {
			break
		}
		result = append(result, r)
		curW += rw
	}
	// Add ellipsis if there's room.
	if width >= 3 && curW+1 <= width {
		// Replace last char(s) with ellipsis.
		if len(result) > 0 {
			lastW := buffer.RuneWidth(result[len(result)-1])
			if curW-lastW+3 <= width {
				result = result[:len(result)-1]
				result = append(result, '…')
			}
		}
	}
	return string(result)
}

// sortRowsLocked sorts rows by sortCol/sortAscending. Caller must hold lock.
func (t *Table) sortRowsLocked() {
	if t.sortCol < 0 || t.sortCol >= len(t.headers) {
		return
	}
	col := t.sortCol
	asc := t.sortAscending
	sort.SliceStable(t.rows, func(i, j int) bool {
		a, b := "", ""
		if col < len(t.rows[i]) {
			a = t.rows[i][col]
		}
		if col < len(t.rows[j]) {
			b = t.rows[j][col]
		}
		if asc {
			return a < b
		}
		return a > b
	})
}

// clampSelectionLocked ensures selectedRow is within valid range.
func (t *Table) clampSelectionLocked() {
	rows := t.displayRowsLocked()
	if t.selectedRow < 0 {
		t.selectedRow = 0
	}
	if len(rows) > 0 && t.selectedRow >= len(rows) {
		t.selectedRow = len(rows) - 1
	}
	if len(rows) == 0 {
		t.selectedRow = 0
	}
}

// clampScrollYLocked ensures scrollY is within valid range.
func (t *Table) clampScrollYLocked() {
	rows := t.displayRowsLocked()
	maxScroll := len(rows)
	if t.scrollY > maxScroll {
		t.scrollY = maxScroll
	}
	if t.scrollY < 0 {
		t.scrollY = 0
	}
}

// utf8RuneCount is a safe rune counter.
func utf8RuneCount(s string) int {
	return utf8.RuneCountInString(s)
}
