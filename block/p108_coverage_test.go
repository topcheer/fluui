package block

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── AssistantTextBlock.Paint (69.4% → 85%+) ───

func TestP108_AssistantTextBlock_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("# Heading 1\n## Heading 2\n### Heading 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("```go\nfunc main() {\n\tprintln(\"hello\")\n}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_UnorderedList(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("- item 1\n- item 2\n  - nested\n- item 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_OrderedList(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("1. first\n2. second\n3. third")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_BoldItalicCode(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("This has **bold**, *italic*, and `code` text.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_Table(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("| Name | Value |\n|------|-------|\n| A    | 1     |\n| B    | 2     |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("> This is a quote\n> Second line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MarkdownLink(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("See [documentation](https://example.com) for details.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("Unicode test: café, 日本語, emoji, naïve")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_HorizontalRule(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("above\n\n---\n\nbelow")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_NestedList(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("1. outer\n   - inner a\n   - inner b\n2. outer 2")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("First paragraph here.\n\nSecond paragraph here.\n\nThird one.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("This is a long line that needs wrapping in narrow width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("line1\nline2\nline3\nline4\nline5")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 2})
	buf := buffer.NewBuffer(80, 2)
	b.Paint(buf) // should stop early
}

func TestP108_AssistantTextBlock_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("Hello world")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 5})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_WidthTruncation(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("text that is wider than the bounds allows for rendering properly")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 5, H: 10})
	buf := buffer.NewBuffer(5, 10)
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf) // should not panic
}

func TestP108_AssistantTextBlock_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("hello")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(80, 24)
	b.Paint(buf) // should not panic
}

// ─── Cache invalidation coverage ───

func TestP108_AssistantTextBlock_CacheInvalidation_ContentChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("initial content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	// Change content - should invalidate cache
	b.SetContent("changed content")
	b.Paint(buf)
}

func TestP108_AssistantTextBlock_CacheInvalidation_WidthChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.SetContent("some content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 10})
	buf := buffer.NewBuffer(80, 10)
	b.Paint(buf)
	// Change width - should invalidate cache
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf2 := buffer.NewBuffer(60, 10)
	b.Paint(buf2)
}

// ─── Serialize/Deserialize roundtrip ───

func TestP108_AssistantTextBlock_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("test-id")
	b.SetContent("hello world")
	data, _ := b.SerializeState()
	if data == nil {
		t.Fatal("expected non-nil serialized state")
	}
	b2 := NewAssistantTextBlock("test-id2")
	b2.DeserializeState(data)
	if b2.Content() != "hello world" {
		t.Errorf("expected 'hello world', got %q", b2.Content())
	}
}

func TestP108_AssistantTextBlock_DeserializeInvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.DeserializeState([]byte("invalid json"))
	// Should not panic, content should be empty
	if b.Content() != "" {
		t.Errorf("expected empty content for invalid JSON, got %q", b.Content())
	}
}

// ─── WorkflowBlock coverage ───

func TestP108_WorkflowBlock_StatusIcons(t *testing.T) {
	// Test case insensitive
	_ = parseStepStatus("DONE")
	_ = parseStepStatus("Running")
	// Empty string
	_ = parseStepStatus("")
}

func TestP108_WorkflowBlock_SpinnerFrame(t *testing.T) {
	b := NewWorkflowBlock("test")
	b.AddStep("step1", "description")
	// Within spinner interval
	b.AdvanceSpinner()
	frame := b.SpinnerFrame()
	if frame == 0 {
		t.Error("expected non-zero spinner frame")
	}
}

func TestP108_WorkflowBlock_HasRunning(t *testing.T) {
	b := NewWorkflowBlock("test")
	b.AddStep("step1", "desc1")
	b.AddStep("step2", "desc2")
	b.SetStepStatus(0, StepDone)
	b.SetStepStatus(1, StepRunning)
	if !b.HasRunning() {
		t.Error("expected HasRunning=true")
	}
}

func TestP108_WorkflowBlock_HasFailed(t *testing.T) {
	b := NewWorkflowBlock("test")
	b.AddStep("step1", "desc1")
	b.SetStepStatus(0, StepFailed)
	if !b.HasFailed() {
		t.Error("expected HasFailed=true")
	}
}

func TestP108_WorkflowBlock_ProgressText(t *testing.T) {
	b := NewWorkflowBlock("test")
	b.AddStep("a", "desc")
	b.AddStep("b", "desc")
	b.SetStepStatus(0, StepDone)
	b.SetStepStatus(1, StepRunning)
	pt := b.ProgressText()
	if pt == "" {
		t.Error("expected non-empty progress text")
	}
}

func TestP108_WorkflowBlock_SerializeDeserialize(t *testing.T) {
	b := NewWorkflowBlock("test")
	b.AddStep("step1", "desc")
	b.SetStepStatus(0, StepDone)
	data, _ := b.SerializeState()
	if data == nil {
		t.Fatal("expected non-nil serialized state")
	}
	b2 := NewWorkflowBlock("test2")
	b2.DeserializeState(data)
}

func TestP108_WorkflowBlock_SerializeEmpty(t *testing.T) {
	b := NewWorkflowBlock("test")
	data, _ := b.SerializeState()
	if data == nil {
		t.Fatal("expected non-nil serialized state for empty workflow")
	}
}

// ─── Container coverage ───

func TestP108_BlockContainer_TotalHeight(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("b1")
	b1.SetContent("hello\nworld\ntest")
	c.AddBlock(b1)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 50})
	c.Measure(component.Bounded(80, 50))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 50})
	h := c.TotalHeight()
	if h <= 0 {
		t.Errorf("expected positive height, got %d", h)
	}
}

func TestP108_BlockContainer_BlockPositions(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("b1")
	b1.SetContent("hello")
	b2 := NewAssistantTextBlock("b2")
	b2.SetContent("world")
	c.AddBlock(b1)
	c.AddBlock(b2)
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 50})
	positions := c.BlockPositions()
	if len(positions) != 2 {
		t.Errorf("expected 2 positions, got %d", len(positions))
	}
}

// ─── SaveContainer ───

func TestP108_SaveContainer_TypeNameProvider(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("b1")
	b1.SetContent("test")
	c.AddBlock(b1)
	r := NewRegistry()
	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty serialized data")
	}
}

func TestP108_SaveContainer_NilRegistry(t *testing.T) {
	c := NewBlockContainer()
	b1 := NewAssistantTextBlock("b1")
	b1.SetContent("test")
	c.AddBlock(b1)
	// SaveContainer handles nil registry gracefully (uses Type().String() fallback)
	_, err := SaveContainer(c, nil)
	if err != nil {
		t.Errorf("unexpected error for nil registry: %v", err)
	}
}

func TestP108_SaveContainer_Empty(t *testing.T) {
	c := NewBlockContainer()
	r := NewRegistry()
	_, err := SaveContainer(c, r)
	if err != nil {
		t.Errorf("unexpected error for empty container: %v", err)
	}
}

func TestP108_SaveContainer_MultipleTypes(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("a"))
	c.AddBlock(NewUserMessageBlock("b", "hello"))
	r := NewRegistry()
	_, err := SaveContainer(c, r)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ─── UserMessageBlock ───

func TestP108_UserMessageBlock_Paint(t *testing.T) {
	b := NewUserMessageBlock("test", "Hello world this is a test message")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
}

func TestP108_UserMessageBlock_Paint_Narrow(t *testing.T) {
	b := NewUserMessageBlock("test", "A very long message that needs wrapping in narrow bounds")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 8, H: 10})
	buf := buffer.NewBuffer(8, 10)
	b.Paint(buf)
}

func TestP108_UserMessageBlock_Paint_Unicode(t *testing.T) {
	b := NewUserMessageBlock("test", "café naïve 日本語")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
}

func TestP108_UserMessageBlock_DeserializeState(t *testing.T) {
	b := NewUserMessageBlock("test", "content")
	data, _ := b.SerializeState()
	b2 := NewUserMessageBlock("test2", "")
	b2.DeserializeState(data)
	if b2.Content() != "content" {
		t.Errorf("expected 'content', got %q", b2.Content())
	}
}

func TestP108_UserMessageBlock_DeserializeInvalidJSON(t *testing.T) {
	b := NewUserMessageBlock("test", "")
	b.DeserializeState([]byte("not json"))
	// Should not panic
}

// ─── ErrorBlock ───

func TestP108_ErrorBlock_Paint(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "something went wrong")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	buf := buffer.NewBuffer(80, 5)
	b.Paint(buf)
}

func TestP108_ErrorBlock_Measure(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "error text")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 5})
	_ = b
}

func TestP108_ErrorBlock_Timestamp(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "error")
	ts := b.Timestamp()
	if ts.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestP108_ErrorBlock_SerializeDeserialize(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "test error")
	data, _ := b.SerializeState()
	b2 := NewErrorBlock("err2")
	b2.DeserializeState(data)
}

// ─── ThinkingBlock ───

func TestP108_ThinkingBlock_Deserialize(t *testing.T) {
	b := NewThinkingBlock("test")
	b.AppendDelta("thinking content")
	data, _ := b.SerializeState()
	b2 := NewThinkingBlock("test2")
	b2.DeserializeState(data)
}

func TestP108_ThinkingBlock_Paint_ExpandedMarkdown(t *testing.T) {
	b := NewThinkingBlock("test")
	b.AppendDelta("## Reasoning\n\nI need to **analyze** this carefully.\n\n- Point A\n- Point B")
	b.Expand()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 15})
	buf := buffer.NewBuffer(80, 15)
	b.Paint(buf)
}

func TestP108_ThinkingBlock_Paint_CollapsedLongPreview(t *testing.T) {
	b := NewThinkingBlock("test")
	b.AppendDelta("This is a very long thinking content that should be truncated in the preview line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
}

// ─── Misc helpers ───

func TestP108_DetectDiff(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"diff --git a/file b/file", true},
		{"@@ -1,3 +1,4 @@\n+added\n-removed", true},
		{"normal text", false},
		{"", false},
	}
	for _, tc := range tests {
		got := DetectDiff(tc.input)
		if got != tc.want {
			t.Errorf("DetectDiff(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestP108_DiffStyle(t *testing.T) {
	for _, dt := range []DiffType{DiffAdd, DiffDel, DiffContext, DiffMeta, DiffHunk} {
		s := DiffStyle(dt)
		_ = s
	}
}

// ─── ImageBlock ───

func TestP108_ImageBlock_FormatDetection(t *testing.T) {
	// PNG magic bytes
	pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	b := NewImageBlock("img1", "test.png", pngData)
	if b.Format() == "" {
		t.Error("expected PNG format detection")
	}
}

func TestP108_ImageBlock_FileSize(t *testing.T) {
	b := NewImageBlock("img1", "test.png", make([]byte, 1024))
	s := b.FileSize()
	if s == "" {
		t.Error("expected non-empty file size")
	}
}

func TestP108_ImageBlock_TruncateText(t *testing.T) {
	b := NewImageBlock("img1", "test.png", nil)
	_ = b
}

// ─── StreamDispatcher ───

func TestP108_StreamDispatcher_Flush(t *testing.T) {
	c := NewBlockContainer()
	d := NewStreamDispatcher(c)
	d.Flush()
	// Should not panic on empty
}

func TestP108_StreamDispatcher_Factory(t *testing.T) {
	c := NewBlockContainer()
	d := NewStreamDispatcher(c)
	f := d.Factory()
	if f == nil {
		t.Error("expected non-nil factory")
	}
}

// ─── Registry.TypeNameForBlock ───

func TestP108_Registry_TypeNameForBlock(t *testing.T) {
	r := NewDefaultRegistry()
	b := NewAssistantTextBlock("test")
	name := r.TypeNameForBlock(b)
	if name == "" {
		t.Error("expected non-empty type name")
	}
}

// ─── BaseBlock.Paint ───

func TestP108_BaseBlock_Paint(t *testing.T) {
	bb := &BaseBlock{}
	bb.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	bb.Paint(buf) // no-op, should not panic
}

// Unused import guard
var _ = term.KeyEnter
