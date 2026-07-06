package block

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AssistantTextBlock.Paint edge cases ───

func TestP78_AssistantText_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("p78-h1")
	b.AppendDelta("# Big Header\n\nText after header.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	if buf.GetCell(0, 0).Rune != 'B' { // "Big Header" (header is bold)
		t.Errorf("header render: cell(0,0) = %q, want 'B'", buf.GetCell(0, 0).Rune)
	}
}

func TestP78_AssistantText_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("p78-code")
	b.AppendDelta("```go\nfmt.Println(\"hello\")\n```\n")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // should render code block without panic
}

func TestP78_AssistantText_Paint_UnorderedList(t *testing.T) {
	b := NewAssistantTextBlock("p78-list")
	b.AppendDelta("- Item one\n- Item two\n- Item three")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("nothing painted for list item")
	}
}

func TestP78_AssistantText_Paint_OrderedList(t *testing.T) {
	b := NewAssistantTextBlock("p78-olist")
	b.AppendDelta("1. First\n2. Second\n3. Third")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP78_AssistantText_Paint_BoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("p78-bold")
	b.AppendDelta("**bold** and *italic* and `code`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

func TestP78_AssistantText_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("p78-table")
	b.AppendDelta("| A | B |\n|---|---|\n| 1 | 2 |\n")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP78_AssistantText_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("p78-offset")
	b.AppendDelta("Hello World")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 5})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
	cell := buf.GetCell(5, 3)
	if cell.Rune != 'H' {
		t.Errorf("offset render: cell(5,3) = %q, want 'H'", cell.Rune)
	}
}

func TestP78_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("p78-narrow")
	b.AppendDelta("This is a very long line that needs wrapping")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 20})
	buf := buffer.NewBuffer(5, 20)
	b.Paint(buf) // should wrap without panic
}

func TestP78_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("p78-trunc")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf) // should stop at height 3 without panic
}

func TestP78_AssistantText_Paint_CellWidthZero(t *testing.T) {
	b := NewAssistantTextBlock("p78-zero")
	b.AppendDelta("café") // é is 2 bytes but 1 width
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf) // should handle unicode without panic
}

// ─── WorkflowBlock.SpinnerFrame ───

func TestP78_Workflow_SpinnerFrame_Advances(t *testing.T) {
	w := NewWorkflowBlock("p78-wf1")
	w.SetTitle("Test Workflow")
	w.AddStep("Step 1", "")
	w.SetStepStatus(0, StepRunning)
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})

	r1 := w.SpinnerFrame()
	time.Sleep(150 * time.Millisecond)
	r2 := w.SpinnerFrame()
	if r1 == r2 {
		t.Error("spinner should advance after 100ms")
	}
}

func TestP78_Workflow_SpinnerFrame_NoAdvanceWithinInterval(t *testing.T) {
	w := NewWorkflowBlock("p78-wf2")
	w.AddStep("Step 1", "")

	r1 := w.SpinnerFrame()
	r2 := w.SpinnerFrame() // immediately, no sleep
	if r1 != r2 {
		t.Error("spinner should not advance within 100ms interval")
	}
}

func TestP78_Workflow_SpinnerFrame_OutOfBounds(t *testing.T) {
	w := NewWorkflowBlock("p78-wf3")
	for i := 0; i < 20; i++ {
		w.AdvanceSpinner()
	}
	_ = w.SpinnerFrame() // should not panic
}

// ─── WorkflowBlock.Paint edge cases ───

func TestP78_Workflow_Paint_WithDescription(t *testing.T) {
	w := NewWorkflowBlock("p78-wf4")
	w.SetTitle("Build Pipeline")
	w.AddStep("compile", "Go build main package")
	w.SetStepStatus(0, StepDone)
	w.AddStep("test", "Running unit tests")
	w.SetStepStatus(1, StepRunning)
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	w.Paint(buf)
}

func TestP78_Workflow_Paint_NarrowWidth(t *testing.T) {
	w := NewWorkflowBlock("p78-wf6")
	w.AddStep("Step 1", "")
	w.SetStepStatus(0, StepDone)
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 5})
	buf := buffer.NewBuffer(3, 5)
	w.Paint(buf) // should not panic with narrow bounds (no progress bar)
}

func TestP78_Workflow_Paint_NoSteps(t *testing.T) {
	w := NewWorkflowBlock("p78-wf7")
	w.SetTitle("Empty")
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	w.Paint(buf) // should not panic
}

func TestP78_Workflow_Paint_AllStatuses(t *testing.T) {
	w := NewWorkflowBlock("p78-wf8")
	w.AddStep("pending", "")
	w.AddStep("running", "")
	w.SetStepStatus(1, StepRunning)
	w.AddStep("done", "")
	w.SetStepStatus(2, StepDone)
	w.AddStep("failed", "")
	w.SetStepStatus(3, StepFailed)
	w.AddStep("skipped", "")
	w.SetStepStatus(4, StepSkipped)
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	w.Paint(buf)
}

func TestP78_Workflow_Serialize_Deserialize(t *testing.T) {
	w := NewWorkflowBlock("p78-wf9")
	w.SetTitle("Pipeline")
	w.AddStep("build", "Build the project")
	w.SetStepStatus(0, StepDone)
	w.AddStep("test", "Run tests")
	w.SetStepStatus(1, StepRunning)

	data, err := w.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	w2 := NewWorkflowBlock("p78-wf9-2")
	err = w2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

func TestP78_Workflow_Deserialize_InvalidJSON(t *testing.T) {
	w := NewWorkflowBlock("p78-wf10")
	err := w.DeserializeState([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP78_Workflow_HasRunning_HasFailed(t *testing.T) {
	w := NewWorkflowBlock("p78-wf11")
	w.AddStep("s1", "")
	w.SetStepStatus(0, StepDone)
	if w.HasRunning() {
		t.Error("HasRunning should be false")
	}
	w.SetStepStatus(0, StepRunning)
	if !w.HasRunning() {
		t.Error("HasRunning should be true")
	}
	w.SetStepStatus(0, StepFailed)
	if !w.HasFailed() {
		t.Error("HasFailed should be true")
	}
}

func TestP78_Workflow_ProgressText_MixedStatuses(t *testing.T) {
	w := NewWorkflowBlock("p78-wf12")
	w.AddStep("a", "")
	w.SetStepStatus(0, StepDone)
	w.AddStep("b", "")
	w.SetStepStatus(1, StepRunning)
	w.AddStep("c", "")
	w.SetStepStatus(2, StepFailed)
	w.AddStep("d", "")

	pt := w.ProgressText()
	if pt == "" {
		t.Error("ProgressText should not be empty")
	}
}
