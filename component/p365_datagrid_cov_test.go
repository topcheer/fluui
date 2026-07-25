package component

import (
	"testing"
)

// TestP365_DataGrid_MaxScrollY_Empty covers the len==0 branch.
func TestP365_DataGrid_MaxScrollY_Empty(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	// maxScrollY with 0 rows → should return 0 (covered via ScrollDown)
	g.ScrollDown(5) // should clamp to 0
}

// TestP365_DataGrid_ColWidth_OutOfRange covers idx >= len(columns).
func TestP365_DataGrid_ColWidth_OutOfRange(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	// colWidth with idx >= len → should return 10 (default)
	// Covered indirectly via Paint with more row cells than columns
	g.SetRows([][]DataGridCell{{{Value: "a"}, {Value: "b"}, {Value: "c"}, {Value: "d"}}})
	g.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := newBenchBuffer(50, 5)
	g.Paint(buf)
}

// TestP365_DataGrid_ColWidth_ZeroWidth covers Width=0 auto-default.
func TestP365_DataGrid_ColWidth_ZeroWidth(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 0}})
	g.SetRows([][]DataGridCell{{{Value: "test"}}})
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := newBenchBuffer(20, 5)
	g.Paint(buf)
}
