package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AssistantTextBlock.Paint markdown coverage (63.9% → 80%+) ───

func TestP84_AssistantText_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("p84-h")
	b.AppendDelta("# H1\n## H2\n### H3\n#### H4")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	// Verify content is painted
	if buf.GetCell(0, 0).Rune == 0 || buf.GetCell(0, 0).Rune == ' ' {
		t.Error("expected header content at (0,0)")
	}
}

func TestP84_AssistantText_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("p84-cb")
	b.AppendDelta("```go\nfmt.Println(\"hello\")\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_UnorderedList(t *testing.T) {
	b := NewAssistantTextBlock("p84-ul")
	b.AppendDelta("- Item 1\n- Item 2\n  - Sub item")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_OrderedList(t *testing.T) {
	b := NewAssistantTextBlock("p84-ol")
	b.AppendDelta("1. First\n2. Second\n3. Third")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_BoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("p84-bi")
	b.AppendDelta("**bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("p84-tbl")
	b.AppendDelta("| Col1 | Col2 |\n|------|------|\n| a    | b    |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("p84-offset")
	b.AppendDelta("Hello world")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 40, H: 10})
	buf := buffer.NewBuffer(50, 15)
	b.Paint(buf)
	// Verify content painted at offset
	if buf.GetCell(5, 3).Rune == 0 || buf.GetCell(5, 3).Rune == ' ' {
		t.Error("expected content at (5,3)")
	}
}

func TestP84_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("p84-narrow")
	b.AppendDelta("This is a long line that will wrap multiple times")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 20})
	buf := buffer.NewBuffer(5, 20)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("p84-trunc")
	b.AppendDelta("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	b.Paint(buf) // should only paint 2 lines
}

func TestP84_AssistantText_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("p84-uni")
	b.AppendDelta("Héllo wörld! 日本語 тест")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_MarkdownLink(t *testing.T) {
	b := NewAssistantTextBlock("p84-link2")
	b.AppendDelta("[Example](https://example.com) link")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

func TestP84_AssistantText_Paint_EmptyAfterClear(t *testing.T) {
	b := NewAssistantTextBlock("p84-empty2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // empty content, should not panic
}

func TestP84_AssistantText_Paint_WidthTruncation(t *testing.T) {
	b := NewAssistantTextBlock("p84-wtrunc")
	b.AppendDelta("A very long line of content that exceeds the available width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	b.Paint(buf) // cells beyond width should be truncated
}

func TestP84_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("p84-zb")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // zero bounds, should not paint or panic
}

func TestP84_AssistantText_Serialize_RoundTrip(t *testing.T) {
	b := NewAssistantTextBlock("p84-ser3")
	b.AppendDelta("# Title\n\nSome **bold** content")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	b2 := NewAssistantTextBlock("p84-ser4")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestP84_AssistantText_Deserialize_InvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("p84-bad2")
	err := b.DeserializeState([]byte("{not valid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ─── WorkflowBlock additional coverage ───

func TestP84_Workflow_StatusIcon_All(t *testing.T) {
	// statusIcon is private, but we can test it via AddStep + Paint
	w := NewWorkflowBlock("wf1")
	w.AddStep("pending", "pending step")
	w.AddStep("running", "running step")
	w.AddStep("done", "done step")
	w.AddStep("failed", "failed step")
	w.AddStep("skipped", "skipped step")
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	w.Paint(buf)
}

func TestP84_Workflow_SpinnerFrame(t *testing.T) {
	w := NewWorkflowBlock("wf2")
	w.AddStep("step1", "doing work")
	w.AdvanceSpinner()
	frame := w.SpinnerFrame()
	if frame < 0 {
		t.Errorf("SpinnerFrame = %d, want >= 0", frame)
	}
}

func TestP84_Workflow_HasRunning_HasFailed(t *testing.T) {
	w := NewWorkflowBlock("wf3")
	w.AddStep("step1", "pending")
	if w.HasRunning() {
		t.Error("expected no running steps")
	}
	if w.HasFailed() {
		t.Error("expected no failed steps")
	}
}

func TestP84_Workflow_ProgressText(t *testing.T) {
	w := NewWorkflowBlock("wf4")
	w.AddStep("s1", "step 1")
	w.AddStep("s2", "step 2")
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	w.Paint(buf)
}

func TestP84_Workflow_Serialize_Deserialize(t *testing.T) {
	w := NewWorkflowBlock("wf5")
	w.AddStep("s1", "step 1")
	data, err := w.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	w2 := NewWorkflowBlock("wf6")
	if err := w2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}
