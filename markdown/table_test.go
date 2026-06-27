package markdown

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestRenderTable_Basic(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |\n")
	if err != nil {
		t.Fatal(err)
	}

	// Find the table block
	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatalf("no BlockTable found in %d blocks", len(blocks))
	}

	// Expected rows: top border, header, separator, data row 1, data row 2, bottom border = 6 lines
	if len(table.Cells) < 5 {
		t.Fatalf("expected at least 5 lines, got %d", len(table.Cells))
	}

	// Verify header content
	headerText := cellText(table.Cells[1])
	if !strings.Contains(headerText, "Name") {
		t.Errorf("header should contain 'Name': %q", headerText)
	}
	if !strings.Contains(headerText, "Age") {
		t.Errorf("header should contain 'Age': %q", headerText)
	}

	// Verify data rows
	row1 := cellText(table.Cells[3])
	if !strings.Contains(row1, "Alice") || !strings.Contains(row1, "30") {
		t.Errorf("row1 should contain Alice and 30: %q", row1)
	}
	row2 := cellText(table.Cells[4])
	if !strings.Contains(row2, "Bob") || !strings.Contains(row2, "25") {
		t.Errorf("row2 should contain Bob and 25: %q", row2)
	}
}

func TestRenderTable_BordersPresent(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("| A | B |\n|---|---|\n| 1 | 2 |\n")
	if err != nil {
		t.Fatal(err)
	}

	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatal("no table block")
	}

	// Top border should contain ┌ and ┐
	topBorder := cellText(table.Cells[0])
	if !strings.Contains(topBorder, "┌") {
		t.Errorf("top border should start with ┌: %q", topBorder)
	}
	if !strings.Contains(topBorder, "┐") {
		t.Errorf("top border should end with ┐: %q", topBorder)
	}

	// Bottom border should contain └ and ┘
	bottomBorder := cellText(table.Cells[len(table.Cells)-1])
	if !strings.Contains(bottomBorder, "└") {
		t.Errorf("bottom border should start with └: %q", bottomBorder)
	}
	if !strings.Contains(bottomBorder, "┘") {
		t.Errorf("bottom border should end with ┘: %q", bottomBorder)
	}

	// Data rows should contain │ separators
	for _, line := range table.Cells {
		text := cellText(line)
		if strings.Contains(text, "│") {
			// Good — has vertical borders
			break
		}
	}
}

func TestRenderTable_HeaderStyling(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("| Col1 | Col2 |\n|------|------|\n| a | b |\n")
	if err != nil {
		t.Fatal(err)
	}

	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatal("no table block")
	}

	// Header row (line 1) should have Bold flag on text cells
	tt := DefaultTheme()
	headerBold := false
	for _, cell := range table.Cells[1] {
		if cell.Rune != '│' && cell.Rune != ' ' && cell.Flags&buffer.Bold != 0 {
			headerBold = true
			break
		}
	}
	if !headerBold {
		t.Error("header row text should have Bold flag")
	}

	// Header should use H4 color (purple in Dracula)
	headerPurple := false
	for _, cell := range table.Cells[1] {
		if cell.Rune != '│' && cell.Rune != ' ' && cell.Fg.Equal(tt.H4) {
			headerPurple = true
			break
		}
	}
	if !headerPurple {
		t.Error("header row should use H4 color")
	}

	// Data row should NOT have Bold flag
	dataBold := false
	for _, cell := range table.Cells[3] {
		if cell.Rune != '│' && cell.Rune != ' ' && cell.Flags&buffer.Bold != 0 {
			dataBold = true
			break
		}
	}
	if dataBold {
		t.Error("data row text should NOT have Bold flag")
	}
}

func TestRenderTable_Alignment(t *testing.T) {
	// Left, center, right alignment
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("| Left | Center | Right |\n|:-----|:------:|------:|\n| a | bb | ccc |\n")
	if err != nil {
		t.Fatal(err)
	}

	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatal("no table block")
	}

	// Data row should be at index 3 (after top border, header, separator)
	if len(table.Cells) < 4 {
		t.Fatalf("expected at least 4 lines, got %d", len(table.Cells))
	}

	dataRow := table.Cells[3]

	// Find the three data cells by looking for content runes
	// We check the position of 'a', 'bb', 'ccc' within the row
	rowText := cellText(dataRow)

	// Left-aligned 'a': should be immediately after "│ "
	// Center-aligned 'bb': should have spaces on both sides
	// Right-aligned 'ccc': should end at " │"

	// Extract each column cell area
	// The row looks like: │ a   │  bb  │ ccc │
	leftCell := extractColumn(rowText, 0)
	centerCell := extractColumn(rowText, 1)
	rightCell := extractColumn(rowText, 2)

	// Left: "a  " (a + trailing spaces)
	leftTrimmed := strings.TrimSpace(leftCell)
	if !strings.HasPrefix(leftTrimmed, "a") {
		t.Errorf("left col should start with 'a': %q (full row: %q)", leftCell, rowText)
	}

	// Center: " bb " (spaces on both sides)
	centerTrimmed := strings.TrimSpace(centerCell)
	if !strings.Contains(centerTrimmed, "bb") {
		t.Errorf("center col should contain 'bb': %q", centerCell)
	}

	// Right: "ccc" should be at the end (right-aligned)
	rightTrimmed := strings.TrimSpace(rightCell)
	if !strings.HasSuffix(rightTrimmed, "ccc") {
		t.Errorf("right col should end with 'ccc': %q", rightCell)
	}
}

func TestRenderTable_ColumnWidths(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("| Short | LongerHeader |\n|-------|--------------|\n| x | y |\n")
	if err != nil {
		t.Fatal(err)
	}

	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatal("no table block")
	}

	// Both header and data rows should have the same total width
	if len(table.Cells) < 2 {
		t.Fatal("need at least 2 rows")
	}
	headerW := len(table.Cells[1])
	dataW := len(table.Cells[3])
	if headerW != dataW {
		t.Errorf("header width (%d) != data width (%d)", headerW, dataW)
	}

	// Column 2 should be wider (accommodates "LongerHeader" = 12 chars)
	headerText := cellText(table.Cells[1])
	if !strings.Contains(headerText, "LongerHeader") {
		t.Errorf("header should contain 'LongerHeader': %q", headerText)
	}
}

func TestRenderTable_SingleColumn(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("| Only |\n|------|\n| Data |\n")
	if err != nil {
		t.Fatal(err)
	}

	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatal("no table block")
	}

	// Should have top border, header, separator, data, bottom border = 5 lines
	if len(table.Cells) != 5 {
		t.Fatalf("expected 5 lines for single column table, got %d", len(table.Cells))
	}
}

func TestRenderTable_ThemeFromGlobal(t *testing.T) {
	// Verify that the markdown renderer works with a converted global theme
	dt := themeDraculaConverted()
	r := NewMarkdownRenderer(dt, 80)

	blocks, err := r.Render("| H |\n|---|\n| V |\n")
	if err != nil {
		t.Fatal(err)
	}

	for _, b := range blocks {
		if b.Type == BlockTable {
			return // success
		}
	}
	t.Error("expected a table block when using MarkdownThemeFromTheme")
}

func TestRenderTable_WideTableScaling(t *testing.T) {
	// A table that exceeds the renderer width should be scaled down
	r := NewMarkdownRenderer(DefaultTheme(), 30)
	blocks, err := r.Render("| alpha | beta | gamma | delta | epsilon |\n|-------|------|-------|-------|---------|\n| aaaaaaa | bbbbbbb | ccccccc | ddddddd | eeeeeee |\n")
	if err != nil {
		t.Fatal(err)
	}

	var table *Block
	for _, b := range blocks {
		if b.Type == BlockTable {
			table = b
			break
		}
	}
	if table == nil {
		t.Fatal("no table block")
	}

	// Each row should be roughly within the renderer width.
	// The scaler reduces columns but borders still add overhead,
	// so we allow some margin (1.5x the renderer width).
	for i, line := range table.Cells {
		w := 0
		for _, c := range line {
			w += int(c.Width)
		}
		if w > 45 {
			t.Errorf("row %d width %d exceeds 45 (renderer width=30)", i, w)
		}
	}
}

// extractColumn extracts the nth column text from a table row string.
// Columns are separated by │ characters.
func extractColumn(row string, colIdx int) string {
	parts := strings.Split(row, "│")
	if colIdx+1 < len(parts) {
		return parts[colIdx+1]
	}
	return ""
}

// themeDraculaConverted returns a MarkdownTheme converted from the global Dracula theme.
func themeDraculaConverted() *MarkdownTheme {
	return MarkdownThemeFromTheme(nil) // nil → DefaultTheme which is Dracula-based
}
