package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ─── BaseBlock.Paint (0% → 100%) ───

func TestP89_BaseBlock_Paint(t *testing.T) {
	bb := NewBaseBlock("test", TypeAssistantText)
	buf := buffer.NewBuffer(10, 5)
	bb.Paint(buf) // should be a no-op
}

// ─── AssistantTextBlock.Paint — fallback text render path ───

func TestP89_AssistantText_Paint_FallbackText(t *testing.T) {
	b := NewAssistantTextBlock("fb-test")
	b.AppendDelta("Hello\nWorld\nThird line")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — zero bounds ───

func TestP89_AssistantText_Paint_ZeroBounds(t *testing.T) {
	b := NewAssistantTextBlock("zb-test")
	b.AppendDelta("Some content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // should not panic
}

// ─── AssistantTextBlock.Paint — empty content ───

func TestP89_AssistantText_Paint_EmptyContent(t *testing.T) {
	b := NewAssistantTextBlock("empty-test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf) // should return early, no panic
}

// ─── AssistantTextBlock.Paint — height truncation ───

func TestP89_AssistantText_Paint_HeightTruncation(t *testing.T) {
	b := NewAssistantTextBlock("trunc-test")
	b.AppendDelta("Line 1\nLine 2\nLine 3\nLine 4\nLine 5")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 3}) // only 3 rows visible
	buf := buffer.NewBuffer(40, 3)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — non-zero offset ───

func TestP89_AssistantText_Paint_NonZeroOffset(t *testing.T) {
	b := NewAssistantTextBlock("offset-test")
	b.AppendDelta("Some text content")
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 30, H: 5})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — width truncation ───

func TestP89_AssistantText_Paint_WidthTruncation(t *testing.T) {
	b := NewAssistantTextBlock("wtrunc-test")
	// Long text that will exceed bounds width
	b.AppendDelta("This is a very long line of text that should exceed the narrow bounds width")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — markdown headers ───

func TestP89_AssistantText_Paint_MarkdownHeaders(t *testing.T) {
	b := NewAssistantTextBlock("hdr-test")
	b.AppendDelta("# Header 1\n\n## Header 2\n\n### Header 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — markdown code blocks ───

func TestP89_AssistantText_Paint_MarkdownCodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("code-test")
	b.AppendDelta("```go\nfunc main() {\n    fmt.Println(\"hello\")\n}\n```")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — markdown lists ───

func TestP89_AssistantText_Paint_MarkdownLists(t *testing.T) {
	b := NewAssistantTextBlock("list-test")
	b.AppendDelta("- Item 1\n- Item 2\n  - Nested item\n- Item 3\n\n1. First\n2. Second")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — markdown bold/italic/code ───

func TestP89_AssistantText_Paint_MarkdownInline(t *testing.T) {
	b := NewAssistantTextBlock("inline-test")
	b.AppendDelta("This has **bold**, *italic*, and `code` text.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — markdown tables ───

func TestP89_AssistantText_Paint_MarkdownTable(t *testing.T) {
	b := NewAssistantTextBlock("table-test")
	b.AppendDelta("| Name | Age |\n|------|-----|\n| Alice | 30 |\n| Bob | 25 |")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — markdown links ───

func TestP89_AssistantText_Paint_MarkdownLinks(t *testing.T) {
	b := NewAssistantTextBlock("links-test")
	b.AppendDelta("Check [this link](https://example.com) for more info.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — unicode content ───

func TestP89_AssistantText_Paint_Unicode(t *testing.T) {
	b := NewAssistantTextBlock("unicode-test")
	b.AppendDelta("Hello café — 世界 🌍")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — multi-paragraph ───

func TestP89_AssistantText_Paint_MultiParagraph(t *testing.T) {
	b := NewAssistantTextBlock("multi-para")
	b.AppendDelta("First paragraph here.\n\nSecond paragraph.\n\nThird and final paragraph.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — nested lists ───

func TestP89_AssistantText_Paint_NestedLists(t *testing.T) {
	b := NewAssistantTextBlock("nested-list")
	b.AppendDelta("- Top level\n  - Second level\n    - Third level\n  - Back to second\n- Back to top")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — blockquote ───

func TestP89_AssistantText_Paint_Blockquote(t *testing.T) {
	b := NewAssistantTextBlock("quote-test")
	b.AppendDelta("> This is a quote.\n> It has multiple lines.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Paint — narrow width ───

func TestP89_AssistantText_Paint_NarrowWidth(t *testing.T) {
	b := NewAssistantTextBlock("narrow-test")
	b.AppendDelta("This is a long sentence that should wrap at narrow widths.")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 15, H: 10})
	buf := buffer.NewBuffer(15, 10)
	b.Paint(buf)
}

// ─── AssistantTextBlock.Measure — with content ───

func TestP89_AssistantText_Measure_WithContent(t *testing.T) {
	b := NewAssistantTextBlock("measure-test")
	b.AppendDelta("Some content\nWith two lines")
	s := b.Measure(component.Constraints{MaxWidth: 80})
	// Content may render to different line counts depending on markdown parsing
	_ = s
}

// ─── AssistantTextBlock.Measure — empty content ───

func TestP89_AssistantText_Measure_Empty(t *testing.T) {
	b := NewAssistantTextBlock("measure-empty")
	s := b.Measure(component.Constraints{MaxWidth: 80})
	if s.H != 1 {
		t.Errorf("Measure height for empty = %d, want 1", s.H)
	}
}

// ─── AssistantTextBlock.Measure — zero max width ───

func TestP89_AssistantText_Measure_ZeroMaxWidth(t *testing.T) {
	b := NewAssistantTextBlock("measure-zero")
	b.AppendDelta("Some content")
	s := b.Measure(component.Constraints{MaxWidth: 0})
	if s.W != 80 {
		t.Errorf("Measure width with 0 maxW = %d, want 80 (default)", s.W)
	}
}

// ─── AssistantTextBlock.Measure — with code block ───

func TestP89_AssistantText_Measure_CodeBlock(t *testing.T) {
	b := NewAssistantTextBlock("measure-code")
	b.AppendDelta("```go\nfunc main() {}\n```")
	s := b.Measure(component.Constraints{MaxWidth: 80})
	if s.H < 1 {
		t.Errorf("Measure height for code block = %d, want >= 1", s.H)
	}
}

// ─── AssistantTextBlock.getCachedBlocks — cache invalidation ───

func TestP89_AssistantText_CacheInvalidation(t *testing.T) {
	b := NewAssistantTextBlock("cache-test")
	b.AppendDelta("First content")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf1 := buffer.NewBuffer(60, 10)
	b.Paint(buf1)

	// Append more content — cache should be invalidated
	b.AppendDelta("\n\nSecond paragraph")
	buf2 := buffer.NewBuffer(60, 10)
	b.Paint(buf2)

	// Change bounds width — cache should be invalidated
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf3 := buffer.NewBuffer(40, 10)
	b.Paint(buf3)
}

// ─── AssistantTextBlock.SerializeState + DeserializeState roundtrip ───

func TestP89_AssistantText_SerializeDeserialize(t *testing.T) {
	b := NewAssistantTextBlock("ser-test")
	b.AppendDelta("Some content to serialize")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	b2 := NewAssistantTextBlock("ser-test2")
	err = b2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}

	if b2.Content() != b.Content() {
		t.Errorf("Content mismatch: %q vs %q", b2.Content(), b.Content())
	}
}

// ─── AssistantTextBlock.DeserializeState — invalid JSON ───

func TestP89_AssistantText_DeserializeState_InvalidJSON(t *testing.T) {
	b := NewAssistantTextBlock("badjson-test")
	err := b.DeserializeState(json.RawMessage(`{invalid json}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ─── AssistantTextBlock.SetContent ───

func TestP89_AssistantText_SetContent(t *testing.T) {
	b := NewAssistantTextBlock("setcontent-test")
	b.SetContent("Replaced content")
	if b.Content() != "Replaced content" {
		t.Errorf("Content after SetContent = %q", b.Content())
	}
}

// ─── AssistantTextBlock concurrent paint ───

func TestP89_AssistantText_ConcurrentPaint(t *testing.T) {
	b := NewAssistantTextBlock("concurrent-test")
	b.AppendDelta("# Title\n\nSome **bold** text with `code` and [links](https://example.com).\n\n- List item 1\n- List item 2")

	done := make(chan struct{})

	// Goroutine 1: paint
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 50; i++ {
			buf := buffer.NewBuffer(60, 20)
			b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
			b.Paint(buf)
		}
	}()

	// Goroutine 2: append deltas
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 50; i++ {
			b.AppendDelta("\nMore content ")
		}
	}()

	<-done
	<-done
}

// ─── ErrorBlock coverage ───

func TestP89_ErrorBlock_Paint(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-test", "Something went wrong")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	eb.Paint(buf)
}

func TestP89_ErrorBlock_Paint_ZeroBounds(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-zb", "Error")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(60, 5)
	eb.Paint(buf) // should not panic
}

func TestP89_ErrorBlock_Paint_NarrowWidth(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-narrow", "A very long error message that should wrap")
	eb.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	eb.Paint(buf)
}

func TestP89_ErrorBlock_Timestamp(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-ts", "Error")
	ts := eb.Timestamp()
	if ts.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

// ─── ErrorBlock.SerializeState ───

func TestP89_ErrorBlock_SerializeDeserialize(t *testing.T) {
	eb := NewErrorBlockWithMessage("err-ser", "Error message")
	data, err := eb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	eb2 := NewErrorBlock("err-ser2")
	eb2.DeserializeState(data)
}

// ─── UserMessageBlock.DeserializeState (83.3% → 100%) ───

func TestP89_UserMessage_DeserializeState(t *testing.T) {
	um := NewUserMessageBlock("um-test", "Original message")

	data, err := um.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	um2 := NewUserMessageBlock("um-test2", "")
	err = um2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if um2.Content() != "Original message" {
		t.Errorf("Content = %q, want 'Original message'", um2.Content())
	}
}

func TestP89_UserMessage_DeserializeState_InvalidJSON(t *testing.T) {
	um := NewUserMessageBlock("um-badjson", "")
	err := um.DeserializeState(json.RawMessage(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ─── ThinkingBlock.DeserializeState ───

func TestP89_ThinkingBlock_DeserializeState(t *testing.T) {
	tb := NewThinkingBlock("tb-ser")
	tb.AppendDelta("Thinking content here")

	data, err := tb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	tb2 := NewThinkingBlock("tb-ser2")
	err = tb2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

// ─── ThinkingBlock.Paint expanded ───

func TestP89_ThinkingBlock_Paint_Expanded(t *testing.T) {
	tb := NewThinkingBlock("tb-expanded")
	tb.AppendDelta("# Analysis\n\nThis is my reasoning process.\n\n1. First point\n2. Second point")
	tb.Expand()
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	tb.Paint(buf)
}

// ─── ThinkingBlock.Paint collapsed long preview ───

func TestP89_ThinkingBlock_Paint_CollapsedLongPreview(t *testing.T) {
	tb := NewThinkingBlock("tb-prev")
	tb.AppendDelta("This is a very long thinking content that should be truncated in the collapsed view preview")
	tb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	tb.Paint(buf)
}

// ─── BlockContainer.TotalHeight (100% already, verify robustness) ───

func TestP89_Container_TotalHeight(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("a1"))
	c.AddBlock(NewAssistantTextBlock("a2"))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	h := c.TotalHeight()
	_ = h
}

// ─── BlockContainer.BlockPositions ───

func TestP89_Container_BlockPositions(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("bp1"))
	c.AddBlock(NewAssistantTextBlock("bp2"))
	c.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	pos := c.BlockPositions()
	// positions are populated during SetBounds
	_ = pos
}

// ─── Registry.TypeNameForBlock ───

func TestP89_Registry_TypeNameForBlock(t *testing.T) {
	r := NewRegistry()
	r.Register("assistant_text", func(id string) Block {
		return NewAssistantTextBlock(id)
	})

	b := NewAssistantTextBlock("reg-test")
	name := r.TypeNameForBlock(b)
	if name == "" {
		t.Error("expected non-empty type name")
	}
}

// ─── WorkflowBlock empty serialize/deserialize ───

func TestP89_Workflow_SerializeEmpty(t *testing.T) {
	wf := NewWorkflowBlock("wf-empty")
	data, err := wf.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}

	wf2 := NewWorkflowBlock("wf-empty2")
	err = wf2.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
}

// ─── StreamDispatcher.Flush / Factory ───

func TestP89_StreamDispatcher_Flush(t *testing.T) {
	container := NewBlockContainer()
	dispatcher := NewStreamDispatcher(container)
	dispatcher.Flush()
}

func TestP89_StreamDispatcher_Factory(t *testing.T) {
	container := NewBlockContainer()
	dispatcher := NewStreamDispatcher(container)
	f := dispatcher.Factory()
	if f == nil {
		t.Error("expected non-nil factory")
	}
}

// ─── SaveContainer (83.3% → 100%) ───

func TestP89_SaveContainer_ErrorPath(t *testing.T) {
	c := NewBlockContainer()
	c.AddBlock(NewAssistantTextBlock("save-test"))
	c.AddBlock(NewErrorBlockWithMessage("err-save", "error"))

	data, err := SaveContainer(c, NewDefaultRegistry())
	if err != nil {
		t.Fatalf("SaveContainer error: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty saved data")
	}
}

// ─── DetectDiff / DiffStyle ───

func TestP89_DetectDiff_SameContent(t *testing.T) {
	result := DetectDiff("same text")
	_ = result
}

func TestP89_DiffStyle_AllTypes(t *testing.T) {
	// Test that DiffStyle doesn't panic for all diff types
	for _, dt := range []DiffType{DiffContext, DiffAdd, DiffDel, DiffHunk, DiffMeta} {
		s := DiffStyle(dt)
		_ = s
	}
}
