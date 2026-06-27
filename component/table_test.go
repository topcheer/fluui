package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- Test helpers ---

// keyChar creates a KeyEvent for a printable character.
func keyChar(r rune) *term.KeyEvent {
	return &term.KeyEvent{Key: term.KeyUnknown, Rune: r}
}

// keySpecial creates a KeyEvent for a non-printable key.
func keySpecial(k term.KeyCode) *term.KeyEvent {
	return &term.KeyEvent{Key: k}
}

// keyCtrl creates a Ctrl+key event.
func keyCtrl(r rune) *term.KeyEvent {
	return &term.KeyEvent{Key: term.KeyUnknown, Rune: r, Modifiers: term.ModCtrl}
}

func tableKeyEvent(key term.KeyCode) *term.KeyEvent {
	return &term.KeyEvent{Key: key}
}

func tableCtrlKey(r rune) *term.KeyEvent {
	return &term.KeyEvent{Key: term.KeyUnknown, Rune: r, Modifiers: term.ModCtrl}
}

// --- Construction tests ---

func TestTable_New(t *testing.T) {
	headers := []string{"Name", "Age", "City"}
	rows := [][]string{
		{"Alice", "30", "NYC"},
		{"Bob", "25", "LA"},
	}

	tbl := NewTable(headers, rows...)
	if tbl == nil {
		t.Fatal("NewTable returned nil")
	}

	h := tbl.Headers()
	if len(h) != 3 || h[0] != "Name" || h[1] != "Age" || h[2] != "City" {
		t.Fatalf("unexpected headers: %v", h)
	}

	if tbl.RowCount() != 2 {
		t.Fatalf("expected 2 rows, got %d", tbl.RowCount())
	}
}

func TestTable_NewEmpty(t *testing.T) {
	tbl := NewTable(nil)
	if tbl.RowCount() != 0 {
		t.Fatalf("expected 0 rows, got %d", tbl.RowCount())
	}
	if len(tbl.Headers()) != 0 {
		t.Fatal("expected empty headers")
	}
}

func TestTable_NewWithRows(t *testing.T) {
	tbl := NewTable(
		[]string{"A", "B"},
		[]string{"1", "2"},
		[]string{"3", "4"},
		[]string{"5", "6"},
	)
	if tbl.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", tbl.RowCount())
	}

	rows := tbl.Rows()
	if rows[0][0] != "1" || rows[2][1] != "6" {
		t.Fatalf("unexpected row data: %v", rows)
	}
}

// --- Data manipulation tests ---

func TestTable_SetHeaders(t *testing.T) {
	tbl := NewTable(nil)
	tbl.SetHeaders([]string{"X", "Y", "Z", "W"})

	h := tbl.Headers()
	if len(h) != 4 {
		t.Fatalf("expected 4 headers, got %d", len(h))
	}
}

func TestTable_SetRows(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})
	tbl.SetRows([][]string{
		{"x", "y"},
		{"p", "q"},
		{"m", "n"},
	})

	if tbl.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", tbl.RowCount())
	}

	// Verify data.
	rows := tbl.Rows()
	if rows[0][0] != "x" || rows[2][1] != "n" {
		t.Fatalf("unexpected data: %v", rows)
	}
}

func TestTable_AddRow(t *testing.T) {
	tbl := NewTable([]string{"Name"})

	tbl.AddRow([]string{"Alice"})
	if tbl.RowCount() != 1 {
		t.Fatalf("expected 1 row, got %d", tbl.RowCount())
	}

	tbl.AddRow([]string{"Bob"})
	tbl.AddRow([]string{"Charlie"})
	if tbl.RowCount() != 3 {
		t.Fatalf("expected 3 rows, got %d", tbl.RowCount())
	}

	rows := tbl.Rows()
	if rows[2][0] != "Charlie" {
		t.Fatalf("expected third row 'Charlie', got %q", rows[2][0])
	}
}

func TestTable_AddRow_ExternalMutation(t *testing.T) {
	tbl := NewTable([]string{"A"})
	data := []string{"hello"}
	tbl.AddRow(data)

	// Mutate original slice — table should be unaffected.
	data[0] = "world"

	rows := tbl.Rows()
	if rows[0][0] != "hello" {
		t.Fatal("AddRow should deep-copy data")
	}
}

func TestTable_Rows_DeepCopy(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"hello"})
	rows := tbl.Rows()
	rows[0][0] = "mutated"

	// Original should be unaffected.
	rows2 := tbl.Rows()
	if rows2[0][0] != "hello" {
		t.Fatal("Rows() should return a deep copy")
	}
}

// --- Selection tests ---

func TestTable_SelectedRow(t *testing.T) {
	tbl := NewTable([]string{"A"},
		[]string{"r1"},
		[]string{"r2"},
		[]string{"r3"},
	)

	if tbl.SelectedRow() != 0 {
		t.Fatalf("expected initial selection 0, got %d", tbl.SelectedRow())
	}
}

func TestTable_SetSelectedRow(t *testing.T) {
	tbl := NewTable([]string{"A"},
		[]string{"r1"},
		[]string{"r2"},
		[]string{"r3"},
	)

	tbl.SetSelectedRow(2)
	if tbl.SelectedRow() != 2 {
		t.Fatalf("expected selection 2, got %d", tbl.SelectedRow())
	}
}

func TestTable_SelectedRowData(t *testing.T) {
	tbl := NewTable([]string{"A", "B"},
		[]string{"a1", "b1"},
		[]string{"a2", "b2"},
	)

	tbl.SetSelectedRow(1)
	data := tbl.SelectedRowData()
	if len(data) != 2 || data[0] != "a2" || data[1] != "b2" {
		t.Fatalf("unexpected row data: %v", data)
	}
}

func TestTable_SelectedRowData_OutOfRange(t *testing.T) {
	tbl := NewTable([]string{"A"})
	data := tbl.SelectedRowData()
	if data != nil {
		t.Fatal("expected nil for empty table")
	}
}

// --- Keyboard navigation tests ---

func TestTable_HandleKey_Down(t *testing.T) {
	tbl := NewTable([]string{"A"},
		[]string{"r1"},
		[]string{"r2"},
		[]string{"r3"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	// Down → row 1.
	tbl.HandleKey(tableKeyEvent(term.KeyDown))
	if tbl.SelectedRow() != 1 {
		t.Fatalf("expected row 1, got %d", tbl.SelectedRow())
	}

	// Down → row 2.
	tbl.HandleKey(tableKeyEvent(term.KeyDown))
	if tbl.SelectedRow() != 2 {
		t.Fatalf("expected row 2, got %d", tbl.SelectedRow())
	}

	// Down at last row → stays at 2.
	tbl.HandleKey(tableKeyEvent(term.KeyDown))
	if tbl.SelectedRow() != 2 {
		t.Fatalf("expected row 2 (clamped), got %d", tbl.SelectedRow())
	}
}

func TestTable_HandleKey_Up(t *testing.T) {
	tbl := NewTable([]string{"A"},
		[]string{"r1"},
		[]string{"r2"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	tbl.SetSelectedRow(1)

	// Up → row 0.
	tbl.HandleKey(tableKeyEvent(term.KeyUp))
	if tbl.SelectedRow() != 0 {
		t.Fatalf("expected row 0, got %d", tbl.SelectedRow())
	}

	// Up at first row → stays at 0.
	tbl.HandleKey(tableKeyEvent(term.KeyUp))
	if tbl.SelectedRow() != 0 {
		t.Fatalf("expected row 0 (clamped), got %d", tbl.SelectedRow())
	}
}

func TestTable_HandleKey_HomeEnd(t *testing.T) {
	tbl := NewTable([]string{"A"},
		[]string{"r1"},
		[]string{"r2"},
		[]string{"r3"},
		[]string{"r4"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	tbl.SetSelectedRow(2)

	// Home → row 0.
	tbl.HandleKey(tableKeyEvent(term.KeyHome))
	if tbl.SelectedRow() != 0 {
		t.Fatalf("expected row 0 after Home, got %d", tbl.SelectedRow())
	}

	// End → last row.
	tbl.HandleKey(tableKeyEvent(term.KeyEnd))
	if tbl.SelectedRow() != 3 {
		t.Fatalf("expected row 3 after End, got %d", tbl.SelectedRow())
	}
}

func TestTable_HandleKey_LeftRight(t *testing.T) {
	tbl := NewTable([]string{"A", "B", "C", "D"},
		[]string{"a1", "b1", "c1", "d1"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	if tbl.ScrollX() != 0 {
		t.Fatalf("expected scrollX=0 initially")
	}

	// Right → scroll right.
	tbl.HandleKey(tableKeyEvent(term.KeyRight))
	if tbl.ScrollX() != 1 {
		t.Fatalf("expected scrollX=1, got %d", tbl.ScrollX())
	}

	// Right again.
	tbl.HandleKey(tableKeyEvent(term.KeyRight))
	if tbl.ScrollX() != 2 {
		t.Fatalf("expected scrollX=2, got %d", tbl.ScrollX())
	}

	// Left → scroll back.
	tbl.HandleKey(tableKeyEvent(term.KeyLeft))
	if tbl.ScrollX() != 1 {
		t.Fatalf("expected scrollX=1, got %d", tbl.ScrollX())
	}

	// Left at 0 → stays at 0.
	for i := 0; i < 10; i++ {
		tbl.HandleKey(tableKeyEvent(term.KeyLeft))
	}
	if tbl.ScrollX() != 0 {
		t.Fatalf("expected scrollX=0 (clamped), got %d", tbl.ScrollX())
	}
}

func TestTable_HandleKey_PageUpPageDown(t *testing.T) {
	tbl := NewTable([]string{"A"})
	for i := 0; i < 20; i++ {
		tbl.AddRow([]string{"row"})
	}
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10}) // dataH = 8

	// PageDown → skip ~9 rows (dataH = boundsH-1 = 9).
	tbl.HandleKey(tableKeyEvent(term.KeyPageDown))
	if tbl.SelectedRow() != 9 {
		t.Fatalf("expected row 9 after PageDown, got %d", tbl.SelectedRow())
	}

	// PageUp → go back 8.
	tbl.HandleKey(tableKeyEvent(term.KeyPageUp))
	if tbl.SelectedRow() != 0 {
		t.Fatalf("expected row 0 after PageUp, got %d", tbl.SelectedRow())
	}
}

func TestTable_HandleKey_Enter(t *testing.T) {
	tbl := NewTable([]string{"A"},
		[]string{"r1"},
		[]string{"r2"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	tbl.SetSelectedRow(1)

	called := false
	var selRow int
	var selData []string
	tbl.OnSelect(func(row int, data []string) {
		called = true
		selRow = row
		selData = data
	})

	tbl.HandleKey(tableKeyEvent(term.KeyEnter))

	if !called {
		t.Fatal("expected OnSelect callback to be called")
	}
	if selRow != 1 {
		t.Fatalf("expected row 1, got %d", selRow)
	}
	if selData[0] != "r2" {
		t.Fatalf("expected data 'r2', got %q", selData[0])
	}
}

func TestTable_HandleKey_Nil(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"r1"})
	if tbl.HandleKey(nil) {
		t.Fatal("HandleKey(nil) should return false")
	}
}

func TestTable_HandleKey_CtrlSort(t *testing.T) {
	tbl := NewTable([]string{"A", "B"},
		[]string{"banana", "2"},
		[]string{"apple", "1"},
		[]string{"cherry", "3"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	// Ctrl+1 → sort by column 0 ascending.
	consumed := tbl.HandleKey(tableCtrlKey('1'))
	if !consumed {
		t.Fatal("expected Ctrl+1 to be consumed")
	}
	rows := tbl.Rows()
	if rows[0][0] != "apple" {
		t.Fatalf("expected 'apple' first after sort, got %q", rows[0][0])
	}

	// Ctrl+1 again → toggle to descending.
	tbl.HandleKey(tableCtrlKey('1'))
	rows = tbl.Rows()
	if rows[0][0] != "cherry" {
		t.Fatalf("expected 'cherry' first after toggle sort, got %q", rows[0][0])
	}
}

// --- Sorting tests ---

func TestTable_SortBy_Ascending(t *testing.T) {
	tbl := NewTable([]string{"Name", "Age"},
		[]string{"Charlie", "30"},
		[]string{"Alice", "25"},
		[]string{"Bob", "35"},
	)

	tbl.SortBy(0, true) // Sort by Name ascending

	rows := tbl.Rows()
	if rows[0][0] != "Alice" || rows[1][0] != "Bob" || rows[2][0] != "Charlie" {
		t.Fatalf("sort ascending failed: %v", rows)
	}
}

func TestTable_SortBy_Descending(t *testing.T) {
	tbl := NewTable([]string{"Name"},
		[]string{"Alice"},
		[]string{"Charlie"},
		[]string{"Bob"},
	)

	tbl.SortBy(0, false)

	rows := tbl.Rows()
	if rows[0][0] != "Charlie" || rows[1][0] != "Bob" || rows[2][0] != "Alice" {
		t.Fatalf("sort descending failed: %v", rows)
	}
}

func TestTable_SortBy_InvalidColumn(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"r1"})
	// Sorting by invalid column should not panic.
	tbl.SortBy(99, true)
	tbl.SortBy(-1, false)
}

func TestTable_ToggleSort(t *testing.T) {
	tbl := NewTable([]string{"Name"},
		[]string{"B"},
		[]string{"A"},
		[]string{"C"},
	)

	tbl.ToggleSort(0)
	rows := tbl.Rows()
	if rows[0][0] != "A" {
		t.Fatalf("expected 'A' first, got %q", rows[0][0])
	}

	tbl.ToggleSort(0) // toggle
	rows = tbl.Rows()
	if rows[0][0] != "C" {
		t.Fatalf("expected 'C' first after toggle, got %q", rows[0][0])
	}
}

func TestTable_SortCol(t *testing.T) {
	tbl := NewTable([]string{"A", "B"}, []string{"1", "2"})
	if tbl.SortCol() != -1 {
		t.Fatalf("expected sortCol=-1 initially, got %d", tbl.SortCol())
	}

	tbl.SortBy(1, true)
	if tbl.SortCol() != 1 {
		t.Fatalf("expected sortCol=1, got %d", tbl.SortCol())
	}
	if !tbl.SortAscending() {
		t.Fatal("expected ascending=true")
	}
}

// --- Alignment tests ---

func TestTable_SetColumnAlign(t *testing.T) {
	tbl := NewTable([]string{"A", "B", "C"})

	tbl.SetColumnAlign(0, AlignLeft)
	tbl.SetColumnAlign(1, AlignCenter)
	tbl.SetColumnAlign(2, AlignRight)

	if tbl.ColumnAlign(0) != AlignLeft {
		t.Fatal("expected AlignLeft for col 0")
	}
	if tbl.ColumnAlign(1) != AlignCenter {
		t.Fatal("expected AlignCenter for col 1")
	}
	if tbl.ColumnAlign(2) != AlignRight {
		t.Fatal("expected AlignRight for col 2")
	}

	// Default alignment.
	if tbl.ColumnAlign(99) != AlignLeft {
		t.Fatal("expected default AlignLeft for unknown column")
	}
}

// --- Display options tests ---

func TestTable_Zebra(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"r1"})

	if !tbl.Zebra() {
		t.Fatal("expected zebra=true by default")
	}

	tbl.SetZebra(false)
	if tbl.Zebra() {
		t.Fatal("expected zebra=false after SetZebra(false)")
	}
}

func TestTable_ShowRowNumbers(t *testing.T) {
	tbl := NewTable([]string{"A"})
	tbl.SetShowRowNumbers(true)
	// Just verify no panic.
}

// --- Column computation tests ---

func TestTable_Columns(t *testing.T) {
	tbl := NewTable(
		[]string{"Name", "Age"},
		[]string{"Alice", "30"},
		[]string{"Bob", "999"},
	)

	cols := tbl.Columns()
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	// Width should be max of header and data.
	if cols[0].Width < 5 {
		t.Fatalf("expected col 0 width >=5 (Alice), got %d", cols[0].Width)
	}
	if cols[1].Width < 3 {
		t.Fatalf("expected col 1 width >=3 (999), got %d", cols[1].Width)
	}
}

func TestTable_Columns_RowsExceedHeaders(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"1", "2", "3"})
	tbl.AddRow([]string{"x", "y", "z", "w"})

	cols := tbl.Columns()
	if len(cols) < 4 {
		t.Fatalf("expected at least 4 columns (row has 4 fields), got %d", len(cols))
	}
}

// --- Measure tests ---

func TestTable_Measure(t *testing.T) {
	tbl := NewTable(
		[]string{"Name", "Age"},
		[]string{"Alice", "30"},
		[]string{"Bob", "25"},
	)

	size := tbl.Measure(Unbounded())
	if size.H != 3 { // 1 header + 2 rows
		t.Fatalf("expected height=3, got %d", size.H)
	}
	if size.W < 2 {
		t.Fatalf("expected width >=2, got %d", size.W)
	}
}

func TestTable_Measure_Empty(t *testing.T) {
	tbl := NewTable(nil)
	size := tbl.Measure(Unbounded())
	if size.H != 1 { // just header row
		t.Fatalf("expected height=1 (header only), got %d", size.H)
	}
}

func TestTable_Measure_Clamped(t *testing.T) {
	tbl := NewTable([]string{"A"})
	for i := 0; i < 100; i++ {
		tbl.AddRow([]string{"row"})
	}

	size := tbl.Measure(Bounded(50, 10))
	if size.W > 50 {
		t.Fatalf("expected width <=50, got %d", size.W)
	}
	if size.H > 10 {
		t.Fatalf("expected height <=10, got %d", size.H)
	}
}

// --- Paint tests ---

func TestTable_Paint(t *testing.T) {
	tbl := NewTable(
		[]string{"Name", "Age"},
		[]string{"Alice", "30"},
		[]string{"Bob", "25"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})

	buf := buffer.NewBuffer(30, 10)
	tbl.Paint(buf)

	// Header should be rendered — check first non-space in row 0.
	// Look for 'N' from "Name".
	found := false
	for x := 0; x < 30; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == 'N' {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find 'N' from 'Name' header in row 0")
	}
}

func TestTable_Paint_TooSmall(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"r1"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})

	buf := buffer.NewBuffer(1, 1)
	// Should not panic.
	tbl.Paint(buf)
}

func TestTable_Paint_Nil(t *testing.T) {
	tbl := NewTable(nil)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	buf := buffer.NewBuffer(10, 5)
	tbl.Paint(buf) // should not panic with empty table
}

// --- Component interface tests ---

func TestTable_ImplementsComponent(t *testing.T) {
	var _ Component = (*Table)(nil)
}

func TestTable_ID(t *testing.T) {
	tbl := NewTable([]string{"A"})
	tbl.SetID("my-table")
	if tbl.ID() != "my-table" {
		t.Fatalf("expected ID 'my-table', got %q", tbl.ID())
	}
}

func TestTable_Children(t *testing.T) {
	tbl := NewTable([]string{"A"})
	if tbl.Children() != nil {
		t.Fatal("Table should have no children")
	}
}

// --- Concurrency tests ---

func TestTable_Concurrent(t *testing.T) {
	tbl := NewTable([]string{"A", "B"})

	var wg sync.WaitGroup

	// Writer: AddRow
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			tbl.AddRow([]string{"a", "b"})
		}
	}()

	// Writer: SetRows
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			tbl.SetRows([][]string{{"x", "y"}})
		}
	}()

	// Reader: Rows
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = tbl.Rows()
		}
	}()

	// Reader: SelectedRow
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = tbl.SelectedRow()
		}
	}()

	// Sorter
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			tbl.SortBy(0, true)
		}
	}()

	// Keyboard handler
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			tbl.HandleKey(tableKeyEvent(term.KeyDown))
		}
	}()

	wg.Wait()
}

func TestTable_ConcurrentPaint(t *testing.T) {
	tbl := NewTable([]string{"A", "B"},
		[]string{"r1", "r2"},
		[]string{"r3", "r4"},
	)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	var wg sync.WaitGroup

	// Paint reader
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			buf := buffer.NewBuffer(20, 10)
			tbl.Paint(buf)
		}
	}()

	// Data writer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			tbl.AddRow([]string{"x", "y"})
		}
	}()

	wg.Wait()
}

// --- Scroll adjustment tests ---

func TestTable_ScrollFollowsSelection(t *testing.T) {
	tbl := NewTable([]string{"A"})
	for i := 0; i < 20; i++ {
		tbl.AddRow([]string{"row"})
	}
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5}) // dataH = 3

	// Go to last row via End.
	tbl.HandleKey(tableKeyEvent(term.KeyEnd))
	if tbl.SelectedRow() != 19 {
		t.Fatalf("expected row 19, got %d", tbl.SelectedRow())
	}

	// Scroll should have followed — scrollY should be > 0.
	if tbl.ScrollY() == 0 {
		t.Fatal("expected scrollY > 0 after navigating to last row")
	}
}

func TestTable_ScrollUpDown(t *testing.T) {
	tbl := NewTable([]string{"A"})
	for i := 0; i < 30; i++ {
		tbl.AddRow([]string{"row"})
	}

	tbl.ScrollDown(5)
	if tbl.ScrollY() != 5 {
		t.Fatalf("expected scrollY=5, got %d", tbl.ScrollY())
	}

	tbl.ScrollDown(100) // over-scroll
	// Should not go negative or exceed limits.
	if tbl.ScrollY() < 5 {
		t.Fatalf("expected scrollY >=5, got %d", tbl.ScrollY())
	}

	tbl.ScrollUp(3)
	if tbl.ScrollY() < 0 {
		t.Fatal("scrollY should never be negative")
	}
}
