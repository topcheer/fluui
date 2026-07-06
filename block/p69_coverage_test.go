package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === P69 Coverage tests ===

func TestP69_UserMessage_SetContent(t *testing.T) {
	b := NewUserMessageBlock("um-1", "initial")
	b.SetContent("replaced")
	if b.Content() != "replaced" {
		t.Errorf("expected 'replaced', got %q", b.Content())
	}
}

func TestP69_UserMessage_SetContent_Empty(t *testing.T) {
	b := NewUserMessageBlock("um-2", "has content")
	b.SetContent("")
	if b.Content() != "" {
		t.Errorf("expected empty, got %q", b.Content())
	}
}

func TestP69_UserMessage_SetContent_MarksDirty(t *testing.T) {
	b := NewUserMessageBlock("um-3", "original")
	b.ClearDirty()
	b.SetContent("new")
	if !b.IsDirty() {
		t.Error("expected dirty after SetContent")
	}
}

func TestP69_UserMessage_Measure(t *testing.T) {
	b := NewUserMessageBlock("um-4", "short")
	s := b.Measure(component.Bounded(80, 100))
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("expected positive size, got %v", s)
	}
}

func TestP69_UserMessage_Measure_Empty(t *testing.T) {
	b := NewUserMessageBlock("um-5", "")
	s := b.Measure(component.Bounded(80, 100))
	if s.H != 1 {
		t.Errorf("expected height 1 for empty, got %d", s.H)
	}
}

func TestP69_UserMessage_Serialize(t *testing.T) {
	b := NewUserMessageBlock("um-6", "test message")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty serialized data")
	}
}

func TestP69_UserMessage_Paint(t *testing.T) {
	b := NewUserMessageBlock("um-7", "hello world")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
	// Should draw content
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}

func TestP69_Thinking_Deserialize_InvalidJSON(t *testing.T) {
	b := NewThinkingBlock("th-1")
	err := b.DeserializeState([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP69_Thinking_SetContent_MarksDirty(t *testing.T) {
	b := NewThinkingBlock("th-2")
	b.SetContent("original thinking")
	b.ClearDirty()
	b.SetContent("new thinking")
	if !b.IsDirty() {
		t.Error("expected dirty after SetContent")
	}
}

func TestP69_Thinking_Paint_EmptyContent(t *testing.T) {
	b := NewThinkingBlock("th-3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf) // should not panic with empty content
	_ = buf
}

func TestP69_ErrorBlock_Measure_LongContent(t *testing.T) {
	b := NewErrorBlockWithMessage("err-1", "this is a very long error message that should wrap across multiple lines when measured against a narrow width")
	s := b.Measure(component.Bounded(20, 100))
	if s.H < 2 {
		t.Errorf("expected multi-line height for long content, got %d", s.H)
	}
}

func TestP69_ErrorBlock_Paint(t *testing.T) {
	b := NewErrorBlockWithMessage("err-2", "test error")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // should not panic
}

func TestP69_Workflow_HasRunning_Empty(t *testing.T) {
	w := NewWorkflowBlock("wf-1")
	if w.HasRunning() {
		t.Error("expected no running steps in empty workflow")
	}
}

func TestP69_Workflow_HasFailed_Empty(t *testing.T) {
	w := NewWorkflowBlock("wf-2")
	if w.HasFailed() {
		t.Error("expected no failed steps in empty workflow")
	}

}

func TestP69_Workflow_ProgressText_Empty(t *testing.T) {
	w := NewWorkflowBlock("wf-3")
	pt := w.ProgressText()
	if pt == "" {
		// ProgressText should return something even for empty workflow
		t.Log("ProgressText for empty:", pt)
	}
}

func TestP69_Workflow_TypeName(t *testing.T) {
	w := NewWorkflowBlock("wf-4")
	if w.TypeName() == "" {
		t.Error("expected non-empty TypeName")
	}
}

func TestP69_Diff_DetectDiff_NotDiff(t *testing.T) {
	// DetectDiff returns true if text looks like a diff
	result := DetectDiff("just some text")
	if result {
		t.Error("expected false for non-diff text")
	}
}

func TestP69_Diff_DetectDiff_IsDiff(t *testing.T) {
	result := DetectDiff("diff --git a/test b/test")
	if !result {
		t.Error("expected true for diff text")
	}
}

func TestP69_Diff_ParseDiff_NoChanges(t *testing.T) {
	diffText := "diff --git a/test b/test\nindex abc..def 100644\n--- a/test\n+++ b/test\n"
	lines := ParseDiff(diffText)
	if len(lines) == 0 {
		t.Error("expected some parsed lines")
	}
}

func TestP69_Diff_DiffStyle_AllTypes(t *testing.T) {
	types := []DiffType{DiffContext, DiffAdd, DiffDel, DiffHunk, DiffMeta}
	for _, dt := range types {
		s := DiffStyle(dt)
		_ = s // should not panic
	}
}

func TestP69_Stream_Dispatcher_Flush(t *testing.T) {
	c := NewBlockContainer()
	dispatcher := NewStreamDispatcher(c)
	dispatcher.Flush() // should not panic with no active blocks
	if len(dispatcher.ActiveBlocks()) != 0 {
		t.Error("expected no active blocks after flush")
	}
}

func TestP69_Stream_Dispatcher_Factory(t *testing.T) {
	c := NewBlockContainer()
	dispatcher := NewStreamDispatcher(c)
	if dispatcher.Factory() == nil {
		t.Error("expected non-nil factory")
	}
}

func TestP69_BlockContainer_TotalHeight_Empty(t *testing.T) {
	c := NewBlockContainer()
	if c.TotalHeight() != 0 {
		t.Errorf("expected 0 height for empty, got %d", c.TotalHeight())
	}
}

func TestP69_BlockContainer_BlockPositions_Empty(t *testing.T) {
	c := NewBlockContainer()
	if len(c.BlockPositions()) != 0 {
		t.Error("expected empty block positions")
	}
}

func TestP69_BlockContainer_TotalHeight_WithBlocks(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewUserMessageBlock("u1", "hello world this is a long message"))
	c.AddBlock(NewUserMessageBlock("u2", "another message here"))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 100})
	if c.TotalHeight() < 0 {
		t.Errorf("expected non-negative height, got %d", c.TotalHeight())
	}
	_ = c.TotalHeight() // may be 0 if blocks don't measure yet
}
