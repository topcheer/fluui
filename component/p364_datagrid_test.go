package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP364_DataGrid_Create(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{
		{Title: "Name", Width: 15},
		{Title: "Model", Width: 12},
		{Title: "Tokens", Width: 10},
	})
	if g.RowCount() != 0 {
		t.Errorf("rows = %d, want 0", g.RowCount())
	}
}

func TestP364_DataGrid_AddRow(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.AddRow([]DataGridCell{{Value: "test1"}})
	g.AddRow([]DataGridCell{{Value: "test2"}})
	if g.RowCount() != 2 {
		t.Errorf("rows = %d, want 2", g.RowCount())
	}
}

func TestP364_DataGrid_SetRows(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.SetRows([][]DataGridCell{
		{{Value: "r1"}},
		{{Value: "r2"}},
		{{Value: "r3"}},
	})
	if g.RowCount() != 3 {
		t.Errorf("rows = %d", g.RowCount())
	}
}

func TestP364_DataGrid_ScrollDown(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	for i := 0; i < 10; i++ {
		g.AddRow([]DataGridCell{{Value: "row"}})
	}
	g.ScrollDown(3)
	g.ScrollDown(100) // should clamp
}

func TestP364_DataGrid_ScrollUp(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.ScrollUp(5) // should clamp to 0
}

func TestP364_DataGrid_ScrollLeftRight(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{
		{Title: "A", Width: 10},
		{Title: "B", Width: 10},
		{Title: "C", Width: 10},
	})
	g.ScrollRight(1)
	g.ScrollLeft(1)
	g.ScrollLeft(100) // clamp to 0
}

func TestP364_DataGrid_CursorRow(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.SetCursorRow(5)
	if g.CursorRow() != 5 {
		t.Errorf("cursor = %d", g.CursorRow())
	}
}

func TestP364_DataGrid_Clear(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.AddRow([]DataGridCell{{Value: "x"}})
	g.SetCursorRow(0)
	g.Clear()
	if g.RowCount() != 0 {
		t.Error("should be empty")
	}
	if g.CursorRow() != -1 {
		t.Error("cursor should be -1 after clear")
	}
}

func TestP364_DataGrid_SetColumns(t *testing.T) {
	g := NewDataGrid(nil)
	g.SetColumns([]DataGridColumn{{Title: "New", Width: 20}})
}

func TestP364_DataGrid_Measure(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	s := g.Measure(Constraints{MaxWidth: 0, MaxHeight: 0})
	if s.W != 80 || s.H != 20 {
		t.Errorf("defaults = %dx%d, want 80x20", s.W, s.H)
	}
}

func TestP364_DataGrid_Paint(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{
		{Title: "Name", Width: 15},
		{Title: "Model", Width: 12},
	})
	g.AddRow([]DataGridCell{{Value: "GPT-4"}, {Value: "OpenAI"}})
	g.AddRow([]DataGridCell{{Value: "Claude"}, {Value: "Anthropic"}})
	g.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	g.Paint(buf)

	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected header content")
	}
}

func TestP364_DataGrid_Paint_Cursor(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.AddRow([]DataGridCell{{Value: "r1"}})
	g.AddRow([]DataGridCell{{Value: "r2"}})
	g.SetCursorRow(1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	g.Paint(buf)
}

func TestP364_DataGrid_Paint_Scrolled(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{
		{Title: "A", Width: 10},
		{Title: "B", Width: 10},
		{Title: "C", Width: 10},
	})
	for i := 0; i < 20; i++ {
		g.AddRow([]DataGridCell{{Value: "a"}, {Value: "b"}, {Value: "c"}})
	}
	g.ScrollDown(5)
	g.ScrollRight(1)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	g.Paint(buf)
}

func TestP364_DataGrid_Paint_LongValues(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "Data", Width: 5}})
	g.AddRow([]DataGridCell{{Value: "This is a very long value"}})
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	g.Paint(buf) // should truncate
}

func TestP364_DataGrid_Paint_Empty(t *testing.T) {
	g := NewDataGrid(nil)
	g.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	g.Paint(buf)
}

func TestP364_DataGrid_Paint_ZeroBounds(t *testing.T) {
	g := NewDataGrid([]DataGridColumn{{Title: "A", Width: 10}})
	g.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(20, 5)
	g.Paint(buf)
}

func BenchmarkDataGrid_Paint(b *testing.B) {
	g := NewDataGrid([]DataGridColumn{
		{Title: "ID", Width: 6},
		{Title: "Name", Width: 15},
		{Title: "Model", Width: 12},
		{Title: "Tokens", Width: 8},
		{Title: "Status", Width: 8},
	})
	for i := 0; i < 50; i++ {
		g.AddRow([]DataGridCell{
			{Value: "001"},
			{Value: "TestItem"},
			{Value: "gpt-4"},
			{Value: "1500"},
			{Value: "ok"},
		})
	}
	g.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g.Paint(buf)
	}
}
