package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestTaskList_Unchecked(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("- [ ] Buy milk\n- [ ] Walk dog")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	text := blocksToText(blocks[0])
	if !containsStr(text, "\u2610") { // ☐
		t.Errorf("expected ballot box (unchecked), got: %q", text)
	}
}

func TestTaskList_Checked(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("- [x] Done task")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	text := blocksToText(blocks[0])
	if !containsStr(text, "\u2611") { // ☑
		t.Errorf("expected ballot box with check, got: %q", text)
	}
}

func TestTaskList_Mixed(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("- [ ] Pending\n- [x] Completed\n- [ ] Also pending")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	text := blocksToText(blocks[0])
	if !containsStr(text, "\u2610") { // ☐ for unchecked
		t.Errorf("expected unchecked boxes, got: %q", text)
	}
	if !containsStr(text, "\u2611") { // ☑ for checked
		t.Errorf("expected checked boxes, got: %q", text)
	}
}

func TestTaskList_NotATask(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("- Regular list item")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	text := blocksToText(blocks[0])
	if containsStr(text, "\u2610") || containsStr(text, "\u2611") {
		t.Errorf("regular list should not have checkboxes, got: %q", text)
	}
	if !containsStr(text, "\u2022") { // • bullet
		t.Errorf("expected bullet, got: %q", text)
	}
}

func TestTaskList_CheckboxTextStripped(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("- [x] Implementation done")
	if err != nil {
		t.Fatal(err)
	}
	text := blocksToText(blocks[0])
	// Should NOT contain literal "[x]" — it should be replaced with ☑
	if containsStr(text, "[x]") {
		t.Errorf("checkbox prefix should be stripped, got: %q", text)
	}
}

func TestTaskList_ContentPreserved(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("- [ ] Important task to do")
	if err != nil {
		t.Fatal(err)
	}
	text := blocksToText(blocks[0])
	if !containsStr(text, "Important task to do") {
		t.Errorf("content should be preserved, got: %q", text)
	}
}

func TestDetectTaskListItem(t *testing.T) {
	// Test the helper directly with source bytes
	source := []byte("- [ ] task")
	_ = source
	// The list item text should start with "[ ]"
	// detectTaskListItem checks if text starts with [ ]/[x]/[X]
	// For unit-level test, we verify the stripTaskPrefix function

	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: ' ', Width: 1},
		{Rune: ']', Width: 1},
		{Rune: ' ', Width: 1},
		{Rune: 'h', Width: 1},
		{Rune: 'i', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 2 {
		t.Errorf("expected 2 cells after strip, got %d", len(result))
	}
	if result[0].Rune != 'h' || result[1].Rune != 'i' {
		t.Errorf("expected 'hi' after strip, got %c%c", result[0].Rune, result[1].Rune)
	}
}

func TestStripTaskPrefix_NoSpaceAfter(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: '[', Width: 1},
		{Rune: 'x', Width: 1},
		{Rune: ']', Width: 1},
	}
	result := stripTaskPrefix(cells)
	// 3 cells is below the minimum (4), so function returns unchanged
	if len(result) != 3 {
		t.Errorf("expected 3 cells (too short to strip), got %d", len(result))
	}
}

func TestStripTaskPrefix_NotCheckbox(t *testing.T) {
	cells := []buffer.Cell{
		{Rune: 'a', Width: 1},
		{Rune: 'b', Width: 1},
	}
	result := stripTaskPrefix(cells)
	if len(result) != 2 {
		t.Errorf("expected unchanged cells, got %d", len(result))
	}
}

func TestStripTaskPrefix_TooShort(t *testing.T) {
	cells := []buffer.Cell{{Rune: 'a', Width: 1}}
	result := stripTaskPrefix(cells)
	if len(result) != 1 {
		t.Errorf("expected unchanged for short input, got %d", len(result))
	}
}

func TestTaskList_OrderedListUnchanged(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 60)
	blocks, err := r.Render("1. [ ] Should be regular numbered")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	text := blocksToText(blocks[0])
	// Ordered list should keep "[ ]" as text, not convert to checkbox
	if containsStr(text, "\u2610") {
		t.Errorf("ordered list should not get checkbox, got: %q", text)
	}
}

// helpers

func blocksToText(b *Block) string {
	var runes []rune
	for _, line := range b.Cells {
		for _, cell := range line {
			runes = append(runes, cell.Rune)
		}
		runes = append(runes, '\n')
	}
	return string(runes)
}

// Ensure buffer import is used
var _ buffer.Cell
