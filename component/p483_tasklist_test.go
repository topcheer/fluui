package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownTaskListBasic(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] Pending\n- [x] Done\n- [ ] Another")

	if tl.TotalCount() != 3 {
		t.Errorf("TotalCount = %d, want 3", tl.TotalCount())
	}
	if tl.TaskCount() != 3 {
		t.Errorf("TaskCount = %d, want 3", tl.TaskCount())
	}
	if tl.CompletedCount() != 1 {
		t.Errorf("CompletedCount = %d, want 1", tl.CompletedCount())
	}
}

func TestMarkdownTaskListToggle(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] Task 1\n- [ ] Task 2")

	tl.ToggleItem(0) // toggle first to checked
	if tl.CompletedCount() != 1 {
		t.Errorf("After toggle: CompletedCount = %d, want 1", tl.CompletedCount())
	}
	tl.ToggleItem(0) // toggle back
	if tl.CompletedCount() != 0 {
		t.Errorf("After toggle back: CompletedCount = %d, want 0", tl.CompletedCount())
	}
}

func TestMarkdownTaskListToggleOutOfRange(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] Task 1")
	tl.ToggleItem(5)  // out of range — should not panic
	tl.ToggleItem(-1) // negative — should not panic
	if tl.CompletedCount() != 0 {
		t.Errorf("CompletedCount after invalid toggle = %d, want 0", tl.CompletedCount())
	}
}

func TestMarkdownTaskListCounts(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [x] Done\n- [ ] Pending\n- Plain item\n- [x] Also done")

	if tl.TotalCount() != 4 {
		t.Errorf("TotalCount = %d, want 4", tl.TotalCount())
	}
	if tl.TaskCount() != 3 {
		t.Errorf("TaskCount = %d, want 3 (excludes plain item)", tl.TaskCount())
	}
	if tl.CompletedCount() != 2 {
		t.Errorf("CompletedCount = %d, want 2", tl.CompletedCount())
	}
}

func TestMarkdownTaskListEmpty(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("")
	if tl.TotalCount() != 0 {
		t.Errorf("TotalCount = %d, want 0", tl.TotalCount())
	}
}

func TestMarkdownTaskListPlainItems(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- Plain item\n- Another item")

	if tl.TotalCount() != 2 {
		t.Errorf("TotalCount = %d, want 2", tl.TotalCount())
	}
	if tl.TaskCount() != 0 {
		t.Errorf("TaskCount = %d, want 0 (no checkboxes)", tl.TaskCount())
	}
}

func TestMarkdownTaskListMeasure(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] A\n- [ ] B\n- [ ] C")
	s := tl.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestMarkdownTaskListPaint(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- [ ] Pending\n- [x] Done\n- [ ] Another")
	tl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})

	buf := buffer.NewBuffer(40, 6)
	tl.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}

	// Check for checkbox characters
	foundUnchecked := false
	foundChecked := false
	for y := 0; y < 6; y++ {
		for x := 0; x < 40; x++ {
			r := buf.GetCell(x, y).Rune
			if r == '☐' {
				foundUnchecked = true
			}
			if r == '☑' {
				foundChecked = true
			}
		}
	}
	if !foundUnchecked {
		t.Error("unchecked checkbox ☐ not found")
	}
	if !foundChecked {
		t.Error("checked checkbox ☑ not found")
	}
}

func TestMarkdownTaskListPaintPlainItems(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("- Plain item\n- [ ] Task")
	tl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})

	buf := buffer.NewBuffer(40, 5)
	tl.Paint(buf)

	// Find bullet char for plain item
	foundBullet := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 1).Rune == '•' {
			foundBullet = true
			break
		}
	}
	if !foundBullet {
		t.Error("bullet char not found for plain item")
	}
}

func TestMarkdownTaskListChildren(t *testing.T) {
	tl := NewMarkdownTaskList()
	if tl.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownTaskListStyle(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetStyle(TaskListStyle{
		Checked:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Unchecked:  buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		NormalItem: buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Checkbox:   buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Border:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	tl.SetMarkdown("- [x] Done\n- [ ] Pending\n- Plain")
	tl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	tl.Paint(buf)
}

func TestMarkdownTaskListStarAsterisk(t *testing.T) {
	tl := NewMarkdownTaskList()
	tl.SetMarkdown("* [ ] Star task\n+ [x] Plus task")
	if tl.TaskCount() != 2 {
		t.Errorf("TaskCount = %d, want 2 (should parse * and + prefixes)", tl.TaskCount())
	}
}
