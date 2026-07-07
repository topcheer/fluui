package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestTable_SetFilter(t *testing.T) {
	tbl := NewTable([]string{"Name", "City", "Age"})
	tbl.SetRows([][]string{
		{"Alice", "NYC", "30"},
		{"Bob", "LA", "25"},
		{"Charlie", "NYC", "35"},
		{"Diana", "SF", "28"},
	})
	tbl.SetFilter("NYC")

	if !tbl.IsFiltered() {
		t.Error("expected filtered")
	}
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 filtered rows, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_CaseInsensitive(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Alice"}, {"bob"}, {"Charlie"}})
	tbl.SetFilter("ALICE")
	if tbl.RowCount() != 1 {
		t.Errorf("expected 1 match, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_EmptyClears(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Alice"}, {"Bob"}})
	tbl.SetFilter("ali")
	tbl.SetFilter("")
	if tbl.IsFiltered() {
		t.Error("expected no filter after empty string")
	}
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 rows after clearing filter, got %d", tbl.RowCount())
	}
}

func TestTable_ClearFilter(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Alice"}, {"Bob"}})
	tbl.SetFilter("ali")
	tbl.ClearFilter()
	if tbl.IsFiltered() {
		t.Error("expected no filter after ClearFilter")
	}
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 rows, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_MatchesAnyColumn(t *testing.T) {
	tbl := NewTable([]string{"Name", "City"})
	tbl.SetRows([][]string{
		{"Alice", "NYC"},
		{"Bob", "LA"},
		{"Charlie", "NYC"},
	})
	// "NYC" matches column 2
	tbl.SetFilter("NYC")
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_NoMatches(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Alice"}, {"Bob"}})
	tbl.SetFilter("xyz")
	if tbl.RowCount() != 0 {
		t.Errorf("expected 0 rows, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_Paint(t *testing.T) {
	tbl := NewTable([]string{"Name", "City"})
	tbl.SetRows([][]string{
		{"Alice", "NYC"},
		{"Bob", "LA"},
		{"Charlie", "SF"},
	})
	tbl.SetFilter("a") // matches Alice, LA, Charlie, SF
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	tbl.Paint(buf)
}

func TestTable_Filter_SelectedRowData(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{
		{"Alice"},
		{"Bob"},
		{"Charlie"},
	})
	tbl.SetFilter("b") // matches Bob
	data := tbl.SelectedRowData()
	if len(data) == 0 || data[0] != "Bob" {
		t.Errorf("expected Bob, got %v", data)
	}
}

func TestTable_Filter_Navigation(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{
		{"Alice"}, {"Bob"}, {"Charlie"}, {"Diana"}, {"Eve"},
	})
	tbl.SetFilter("a") // matches Alice, Charlie, Diana
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	// Should be 3 rows
	if tbl.RowCount() != 3 {
		t.Errorf("expected 3, got %d", tbl.RowCount())
	}

	// Navigate down
	tbl.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if tbl.SelectedRow() != 1 {
		t.Errorf("expected row 1, got %d", tbl.SelectedRow())
	}

	// Navigate to end
	tbl.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if tbl.SelectedRow() != 2 {
		t.Errorf("expected row 2, got %d", tbl.SelectedRow())
	}

	// Navigate up
	tbl.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if tbl.SelectedRow() != 1 {
		t.Errorf("expected row 1, got %d", tbl.SelectedRow())
	}
}

func TestTable_Filter_AddRowReapplies(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Apple"}, {"Banana"}})
	tbl.SetFilter("an") // matches Banana
	if tbl.RowCount() != 1 {
		t.Errorf("expected 1, got %d", tbl.RowCount())
	}
	// Add a new matching row
	tbl.AddRow([]string{"Mango"})
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 after adding Mango, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_SetRowsReapplies(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Apple"}})
	tbl.SetFilter("ap")
	if tbl.RowCount() != 1 {
		t.Errorf("expected 1, got %d", tbl.RowCount())
	}
	tbl.SetRows([][]string{{"Apricot"}, {"Grape"}, {"Cherry"}})
	if tbl.RowCount() != 2 {
		t.Errorf("expected 2 (Apricot, Grape), got %d", tbl.RowCount())
	}
}

func TestTable_Filter_Measure(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Alice"}, {"Bob"}, {"Charlie"}})
	tbl.SetFilter("b")
	s := tbl.Measure(Bounded(40, 100))
	// Measure should reflect filtered count: 1 header + 1 data row = 2
	if s.H != 2 {
		t.Errorf("expected H=2, got H=%d", s.H)
	}
}

func TestTable_Filter_SortInteraction(t *testing.T) {
	tbl := NewTable([]string{"Name", "Score"})
	tbl.SetRows([][]string{
		{"Alice", "90"},
		{"Bob", "85"},
		{"Charlie", "95"},
		{"Diana", "80"},
	})
	tbl.SetFilter("a") // matches Alice, Charlie, Diana
	tbl.SortBy(1, false) // sort by Score descending
	// After sort + filter, should show Charlie(95), Alice(90), Diana(80)
	if tbl.RowCount() != 3 {
		t.Errorf("expected 3, got %d", tbl.RowCount())
	}
}

func TestTable_Filter_HomeEnd(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	tbl.SetRows([][]string{{"Alice"}, {"Bob"}, {"Charlie"}, {"Diana"}})
	tbl.SetFilter("a")
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	tbl.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if tbl.SelectedRow() != tbl.RowCount()-1 {
		t.Errorf("expected last row")
	}

	tbl.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if tbl.SelectedRow() != 0 {
		t.Errorf("expected first row")
	}
}

func TestTable_Filter_ReturnsEmptyStringWhenNone(t *testing.T) {
	tbl := NewTable([]string{"Name"})
	if tbl.Filter() != "" {
		t.Error("expected empty filter string")
	}
}

func TestRowMatchesFilter(t *testing.T) {
	if !rowMatchesFilter([]string{"hello", "world"}, "ell") {
		t.Error("expected match")
	}
	if rowMatchesFilter([]string{"hello", "world"}, "xyz") {
		t.Error("expected no match")
	}
	if !rowMatchesFilter([]string{"HELLO"}, "hello") {
		t.Error("expected case-insensitive match")
	}
	if rowMatchesFilter([]string{}, "anything") {
		t.Error("expected no match for empty row")
	}
}
