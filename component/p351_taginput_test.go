package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP351_TagInput_Create(t *testing.T) {
	ti := NewTagInput("Add tags...")
	if ti.TagCount() != 0 {
		t.Errorf("count = %d, want 0", ti.TagCount())
	}
	if ti.Input() != "" {
		t.Error("input should be empty")
	}
}

func TestP351_TagInput_AddTag(t *testing.T) {
	ti := NewTagInput("")
	if !ti.AddTag("golang") {
		t.Error("should add tag")
	}
	if !ti.AddTag("tui") {
		t.Error("should add tag")
	}
	if ti.TagCount() != 2 {
		t.Errorf("count = %d, want 2", ti.TagCount())
	}
}

func TestP351_TagInput_AddDuplicate(t *testing.T) {
	ti := NewTagInput("")
	ti.AddTag("go")
	if ti.AddTag("go") {
		t.Error("should not add duplicate")
	}
	if ti.TagCount() != 1 {
		t.Errorf("count = %d after duplicate", ti.TagCount())
	}
}

func TestP351_TagInput_AddEmpty(t *testing.T) {
	ti := NewTagInput("")
	if ti.AddTag("") {
		t.Error("should not add empty tag")
	}
}

func TestP351_TagInput_MaxTags(t *testing.T) {
	ti := NewTagInput("")
	ti.SetMaxTags(2)
	ti.AddTag("a")
	ti.AddTag("b")
	if ti.AddTag("c") {
		t.Error("should not exceed maxTags")
	}
	if ti.TagCount() != 2 {
		t.Errorf("count = %d, want 2", ti.TagCount())
	}
}

func TestP351_TagInput_RemoveTag(t *testing.T) {
	ti := NewTagInput("")
	ti.AddTag("a")
	ti.AddTag("b")
	ti.AddTag("c")
	ti.RemoveTag(1) // remove "b"
	tags := ti.Tags()
	if len(tags) != 2 {
		t.Fatalf("count = %d, want 2", len(tags))
	}
	if tags[0] != "a" || tags[1] != "c" {
		t.Errorf("tags = %v, want [a c]", tags)
	}
}

func TestP351_TagInput_RemoveTag_Invalid(t *testing.T) {
	ti := NewTagInput("")
	ti.AddTag("a")
	ti.RemoveTag(-1)
	ti.RemoveTag(99)
	if ti.TagCount() != 1 {
		t.Errorf("count = %d after invalid remove", ti.TagCount())
	}
}

func TestP351_TagInput_RemoveLast(t *testing.T) {
	ti := NewTagInput("")
	ti.AddTag("a")
	ti.AddTag("b")
	last := ti.RemoveLast()
	if last != "b" {
		t.Errorf("last = %q, want b", last)
	}
	if ti.TagCount() != 1 {
		t.Errorf("count = %d", ti.TagCount())
	}
}

func TestP351_TagInput_RemoveLast_Empty(t *testing.T) {
	ti := NewTagInput("")
	if ti.RemoveLast() != "" {
		t.Error("should return empty for empty input")
	}
}

func TestP351_TagInput_SetInput(t *testing.T) {
	ti := NewTagInput("")
	ti.SetInput("typing")
	if ti.Input() != "typing" {
		t.Errorf("input = %q", ti.Input())
	}
}

func TestP351_TagInput_CommitInput(t *testing.T) {
	ti := NewTagInput("")
	ti.SetInput("newtag")
	if !ti.CommitInput() {
		t.Error("should commit tag")
	}
	if ti.Input() != "" {
		t.Error("input should be cleared after commit")
	}
	if ti.TagCount() != 1 {
		t.Errorf("count = %d", ti.TagCount())
	}
}

func TestP351_TagInput_CommitInput_Empty(t *testing.T) {
	ti := NewTagInput("")
	ti.SetInput("")
	if ti.CommitInput() {
		t.Error("should not commit empty input")
	}
}

func TestP351_TagInput_SetTags(t *testing.T) {
	ti := NewTagInput("")
	ti.SetTags([]string{"x", "y", "z"})
	if ti.TagCount() != 3 {
		t.Errorf("count = %d", ti.TagCount())
	}
}

func TestP351_TagInput_Clear(t *testing.T) {
	ti := NewTagInput("")
	ti.AddTag("a")
	ti.SetInput("text")
	ti.Clear()
	if ti.TagCount() != 0 {
		t.Error("tags should be cleared")
	}
	if ti.Input() != "" {
		t.Error("input should be cleared")
	}
}

func TestP351_TagInput_Measure(t *testing.T) {
	ti := NewTagInput("")
	s := ti.Measure(Constraints{MaxWidth: 40, MaxHeight: 1})
	if s.H != 1 {
		t.Errorf("height = %d, want 1", s.H)
	}
}

func TestP351_TagInput_Paint(t *testing.T) {
	ti := NewTagInput("Add tags...")
	ti.AddTag("go")
	ti.AddTag("tui")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ti.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP351_TagInput_Paint_WithInput(t *testing.T) {
	ti := NewTagInput("")
	ti.AddTag("tag1")
	ti.SetInput("typing")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ti.Paint(buf)
}

func TestP351_TagInput_Paint_Placeholder(t *testing.T) {
	ti := NewTagInput("Enter tags...")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ti.Paint(buf)
}

func TestP351_TagInput_Paint_Empty(t *testing.T) {
	ti := NewTagInput("")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	ti.Paint(buf)
}

func TestP351_TagInput_Paint_TooManyTags(t *testing.T) {
	ti := NewTagInput("")
	for i := 0; i < 20; i++ {
		ti.AddTag("verylongtag" + string(rune('a'+i)))
	}
	ti.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	ti.Paint(buf) // should clip without panic
}

func BenchmarkTagInput_Paint(b *testing.B) {
	ti := NewTagInput("")
	ti.AddTag("golang")
	ti.AddTag("tui")
	ti.AddTag("ai")
	ti.SetInput("typ")
	ti.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ti.Paint(buf)
	}
}
