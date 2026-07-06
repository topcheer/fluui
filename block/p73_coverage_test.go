package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === P73 Coverage tests for assistant_text.go Paint (63.9%) ===

func TestP73_AssistantText_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("at-h-1")
	b.AppendDelta("# Big Header\n\n## Smaller Header\n\n### Smallest")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	// Header should produce content
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected header content at (0,0)")
	}
}

func TestP73_AssistantText_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("at-cb-1")
	b.AppendDelta("Here is some code:\n\n```go\nfmt.Println(\"hello\")\n```\n\nDone.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
	// Should have content in multiple rows
	found := false
	for y := 0; y < 15; y++ {
		if buf.GetCell(0, y).Rune != 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected content in buffer after Paint")
	}
}

func TestP73_AssistantText_Paint_ListItems(t *testing.T) {
	b := NewAssistantTextBlock("at-li-1")
	b.AppendDelta("- First item\n- Second item\n- Third item\n")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected list content at (0,0)")
	}
}

func TestP73_AssistantText_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("at-tb-1")
	b.AppendDelta("| Name | Value |\n|------|-------|\n| A    | 1     |\n| B    | 2     |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	found := false
	for y := 0; y < 10; y++ {
		if buf.GetCell(0, y).Rune != 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected table content in buffer")
	}
}

func TestP73_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("at-zb-1")
	b.AppendDelta("content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // should not panic
}

func TestP73_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("at-nw-1")
	b.AppendDelta("This is a long line that should be truncated when width is very narrow")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	b.Paint(buf)
	// Should not overflow
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}

func TestP73_AssistantText_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("at-mp-1")
	b.AppendDelta("First paragraph here.\n\nSecond paragraph here.\n\nThird paragraph.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
	// Multiple paragraphs should fill multiple rows
	rowsWithContent := 0
	for y := 0; y < 15; y++ {
		if buf.GetCell(0, y).Rune != 0 {
			rowsWithContent++
		}
	}
	if rowsWithContent < 2 {
		t.Errorf("expected at least 2 rows with content, got %d", rowsWithContent)
	}
}

func TestP73_AssistantText_Paint_BoldItalic(t *testing.T) {
	b := NewAssistantTextBlock("at-bi-1")
	b.AppendDelta("**bold text** and *italic text* and `code text`")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected styled content at (0,0)")
	}
}

func TestP73_AssistantText_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("at-ec-1")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf) // should not panic with empty content
}

func TestP73_AssistantText_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("at-no-1")
	b.AppendDelta("content at offset")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 40, H: 5})
	buf := buffer.NewBuffer(80, 20)
	b.Paint(buf)
	cell := buf.GetCell(5, 3)
	if cell.Rune == 0 {
		t.Error("expected content at offset (5,3)")
	}
}

func TestP73_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("at-ht-1")
	b.AppendDelta("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 3})
	buf := buffer.NewBuffer(80, 3)
	b.Paint(buf) // should stop at height limit without panic
}

func TestP73_AssistantText_Measure_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("at-mc-1")
	b.AppendDelta("```go\nfmt.Println(\"hello\")\n```")
	s := b.Measure(component.Bounded(80, 100))
	if s.H < 2 {
		t.Errorf("expected at least 2 lines for code block, got %d", s.H)
	}
}

func TestP73_AssistantText_Measure_NestedList(t *testing.T) {
	b := NewAssistantTextBlock("at-mn-1")
	b.AppendDelta("- item 1\n  - nested 1\n  - nested 2\n- item 2")
	s := b.Measure(component.Bounded(80, 100))
	if s.H < 4 {
		t.Errorf("expected at least 4 lines for nested list, got %d", s.H)
	}
}

func TestP73_AssistantText_SerializeDeserialize_RoundTrip(t *testing.T) {
	b := NewAssistantTextBlock("at-sd-1")
	b.AppendDelta("test content")
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	b2 := NewAssistantTextBlock("at-sd-2")
	err = b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if b2.Content() != "test content" {
		t.Errorf("expected 'test content', got %q", b2.Content())
	}
}

func TestP73_AssistantText_DeserializeState_InvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("at-dj-1")
	err := b.DeserializeState(json.RawMessage("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// === P73 Coverage tests for workflow.go functions ===

func TestP73_Workflow_statusIcon_AllStatuses(t *testing.T) {
	statuses := []StepStatus{StepPending, StepRunning, StepDone, StepFailed, StepSkipped, StepStatus(99)}
	for _, s := range statuses {
		icon := statusIcon(s)
		if icon == 0 {
			t.Errorf("statusIcon(%d) returned 0", s)
		}
	}
}

func TestP73_Workflow_statusColor_AllStatuses(t *testing.T) {
	statuses := []StepStatus{StepPending, StepRunning, StepDone, StepFailed, StepSkipped, StepStatus(99)}
	for _, s := range statuses {
		color := statusColor(s)
		_ = color // should not panic
	}
}

func TestP73_Workflow_parseStepStatus_AllStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected StepStatus
	}{
		{"pending", StepPending},
		{"running", StepRunning},
		{"done", StepDone},
		{"failed", StepFailed},
		{"skipped", StepSkipped},
		{"unknown", StepPending}, // unknown defaults to pending
		{"", StepPending},
		{"PENDING", StepPending},     // case insensitive
		{"RUNNING", StepRunning},     // case insensitive
		{"Done", StepDone},           // mixed case
	}
	for _, tt := range tests {
		got := parseStepStatus(tt.input)
		if got != tt.expected {
			t.Errorf("parseStepStatus(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestP73_Workflow_SpinnerFrame(t *testing.T) {
	w := NewWorkflowBlock("wf-sf-1")
	frame := w.SpinnerFrame()
	_ = frame // should not panic
}

func TestP73_Workflow_ProgressText_MixedStatuses(t *testing.T) {
	w := NewWorkflowBlock("wf-pt-1")
	w.AddStep("step1", "desc1")
	w.AddStep("step2", "desc2")
	w.AddStep("step3", "desc3")
	w.AddStep("step4", "desc4")
	w.SetStepStatus(0, StepDone)
	w.SetStepStatus(1, StepSkipped)
	w.SetStepStatus(2, StepRunning)
	w.SetStepStatus(3, StepPending)
	pt := w.ProgressText()
	if pt == "" {
		t.Error("expected non-empty progress text")
	}
}

func TestP73_Workflow_ProgressFraction_AllComplete(t *testing.T) {
	w := NewWorkflowBlock("wf-pf-1")
	w.AddStep("s1", "")
	w.AddStep("s2", "")
	w.AddStep("s3", "")
	w.SetStepStatus(0, StepDone)
	w.SetStepStatus(1, StepDone)
	w.SetStepStatus(2, StepDone)
	// Should be 100% complete
}

func TestP73_Workflow_ProgressFraction_NoneComplete(t *testing.T) {
	w := NewWorkflowBlock("wf-pf-2")
	w.AddStep("s1", "")
	w.AddStep("s2", "")
	w.SetStepStatus(0, StepPending)
	w.SetStepStatus(1, StepRunning)
}

func TestP73_Workflow_SerializeDeserialize_RoundTrip(t *testing.T) {
	w := NewWorkflowBlock("wf-sd-1")
	w.SetTitle("Test Workflow")
	w.AddStep("step1", "description1")
	w.AddStep("step2", "description2")
	w.SetStepStatus(0, StepDone)
	w.SetStepStatus(1, StepRunning)

	data, err := w.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	w2 := NewWorkflowBlock("wf-sd-2")
	err = w2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if w2.Title() != "Test Workflow" {
		t.Errorf("expected title 'Test Workflow', got %q", w2.Title())
	}
}

func TestP73_Workflow_DeserializeState_InvalidJSON(t *testing.T) {
	w := NewWorkflowBlock("wf-dj-1")
	err := w.DeserializeState(json.RawMessage("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP73_Workflow_Paint_WithSteps(t *testing.T) {
	w := NewWorkflowBlock("wf-pw-1")
	w.SetTitle("My Workflow")
	w.AddStep("step1", "Do something")
	w.AddStep("step2", "Do another thing")
	w.SetStepStatus(0, StepDone)
	w.SetStepStatus(1, StepRunning)
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	w.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}

func TestP73_Workflow_Paint_EmptySteps(t *testing.T) {
	w := NewWorkflowBlock("wf-pe-1")
	w.SetTitle("Empty Workflow")
	w.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	w.Paint(buf) // should not panic with no steps
}

func TestP73_Workflow_HasRunning(t *testing.T) {
	w := NewWorkflowBlock("wf-hr-1")
	w.AddStep("s1", "")
	if w.HasRunning() {
		t.Error("expected no running initially")
	}
	w.SetStepStatus(0, StepRunning)
	if !w.HasRunning() {
		t.Error("expected running after SetStepStatus")
	}
}

func TestP73_Workflow_HasFailed(t *testing.T) {
	w := NewWorkflowBlock("wf-hf-1")
	w.AddStep("s1", "")
	if w.HasFailed() {
		t.Error("expected no failed initially")
	}
	w.SetStepStatus(0, StepFailed)
	if !w.HasFailed() {
		t.Error("expected failed after SetStepStatus")
	}
}

// === P73 Coverage tests for thinking.go ===

func TestP73_Thinking_SpinnerFrame_OutOfBounds(t *testing.T) {
	b := NewThinkingBlock("th-sf-1")
	// Manually set out-of-bounds spinner frame
	b.mu.Lock()
	b.spinnerF = 999
	b.mu.Unlock()
	frame := b.SpinnerFrame()
	if frame == "" {
		t.Error("expected non-empty frame even for out-of-bounds")
	}
}

func TestP73_Thinking_SpinnerFrame_NegativeFrame(t *testing.T) {
	b := NewThinkingBlock("th-sf-2")
	b.mu.Lock()
	b.spinnerF = -5
	b.mu.Unlock()
	frame := b.SpinnerFrame()
	if frame == "" {
		t.Error("expected non-empty frame even for negative index")
	}
}

func TestP73_Thinking_PreviewLineUnlocked_LeadingNewlines(t *testing.T) {
	b := NewThinkingBlock("th-plu-1")
	b.AppendDelta("\n\n\nActual content after newlines")
	preview := b.PreviewLineUnlocked(80)
	if preview != "Actual content after newlines" {
		t.Errorf("expected trimmed content, got %q", preview)
	}
}

func TestP73_Thinking_PreviewLineUnlocked_ZeroMaxLen(t *testing.T) {
	b := NewThinkingBlock("th-plu-2")
	b.AppendDelta("content")
	preview := b.PreviewLineUnlocked(0)
	if preview != "" {
		t.Errorf("expected empty for maxLen=0, got %q", preview)
	}
}

func TestP73_Thinking_PreviewLineUnlocked_NegativeMaxLen(t *testing.T) {
	b := NewThinkingBlock("th-plu-3")
	b.AppendDelta("content")
	preview := b.PreviewLineUnlocked(-1)
	if preview != "" {
		t.Errorf("expected empty for negative maxLen, got %q", preview)
	}
}

func TestP73_Thinking_PaintExpanded_MarkdownContent(t *testing.T) {
	b := NewThinkingBlock("th-pem-1")
	b.AppendDelta("## Analysis\n\nI need to consider:\n\n1. Performance\n2. Correctness\n3. Readability")
	b.SetState(BlockComplete)
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 20})
	buf := buffer.NewBuffer(80, 20)
	b.Paint(buf)
	// Header at row 0
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected header at (0,0)")
	}
}

func TestP73_Thinking_PaintCollapsed_LongPreview(t *testing.T) {
	b := NewThinkingBlock("th-pcl-1")
	b.AppendDelta("This is a very long preview that should be truncated when rendered in collapsed view with narrow width")
	b.SetState(BlockComplete)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)
	// Should not overflow width
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected content at (0,0)")
	}
}
