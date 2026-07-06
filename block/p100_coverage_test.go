package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// ═══════════════════════════════════════════════════════════════════════════
// P100 Coverage Tests — Block package low-coverage functions
// ═══════════════════════════════════════════════════════════════════════════

// ─── SaveContainer edge cases (83.3% → higher) ───

func TestP100_SaveContainer_TypeNameProvider(t *testing.T) {
	c := NewBlockContainer()
	r := NewRegistry()
	r.Register("text", func(id string) Block { return NewAssistantTextBlock(id) })

	b := NewAssistantTextBlock("test1")
	b.AppendDelta("Hello world")
	c.AddBlock(b)

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}

	// Verify it's valid JSON
	var sc SerializedContainer
	if err := json.Unmarshal(data, &sc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(sc.Blocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(sc.Blocks))
	}
}

func TestP100_SaveContainer_NilRegistry(t *testing.T) {
	c := NewBlockContainer()
	b := NewAssistantTextBlock("test1")
	b.AppendDelta("content")
	c.AddBlock(b)

	data, err := SaveContainer(c, nil)
	if err != nil {
		t.Fatalf("SaveContainer with nil registry: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

func TestP100_SaveContainer_EmptyContainer(t *testing.T) {
	c := NewBlockContainer()
	r := NewRegistry()
	r.Register("text", func(id string) Block { return NewAssistantTextBlock(id) })

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer empty: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data even for empty container")
	}
}

func TestP100_SaveContainer_MultipleBlockTypes(t *testing.T) {
	c := NewBlockContainer()
	r := NewRegistry()
	r.Register("text", func(id string) Block { return NewAssistantTextBlock(id) })
	r.Register("thinking", func(id string) Block { return NewThinkingBlock(id) })
	r.Register("user", func(id string) Block { return NewUserMessageBlock(id, "") })

	at := NewAssistantTextBlock("at1")
	at.AppendDelta("assistant text")
	c.AddBlock(at)

	th := NewThinkingBlock("th1")
	th.AppendDelta("thinking content")
	c.AddBlock(th)

	um := NewUserMessageBlock("um1", "user message")
	c.AddBlock(um)

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer multiple types: %v", err)
	}

	var sc SerializedContainer
	json.Unmarshal(data, &sc)
	if len(sc.Blocks) != 3 {
		t.Errorf("expected 3 blocks, got %d", len(sc.Blocks))
	}
}

// ─── getCachedBlocks edge cases (84.6% → higher) ───

func TestP100_GetCachedBlocks_WidthChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("# Heading\n\nSome text here")

	// First measure at width 80
	b.Measure(component.Bounded(80, 100))
	// Second measure at different width should invalidate cache
	b.Measure(component.Bounded(60, 100))
	// Third measure at same width should use cache
	b.Measure(component.Bounded(60, 100))
}

func TestP100_GetCachedBlocks_ContentChange(t *testing.T) {
	b := NewAssistantTextBlock("test")
	b.AppendDelta("initial content")
	b.Measure(component.Bounded(80, 100))
	// Content change should invalidate cache
	b.AppendDelta(" more content")
	b.Measure(component.Bounded(80, 100))
}

// ─── UserMessageBlock Paint (85.7% → higher) ───

func TestP100_UserMessage_Paint_MultiParagraph(t *testing.T) {
	b := NewUserMessageBlock("test", "Line 1\nLine 2\nLine 3")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP100_UserMessage_Paint_LongText(t *testing.T) {
	longText := "This is a very long line of text that will need to be wrapped across multiple lines to fit within the display bounds"
	b := NewUserMessageBlock("test", longText)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	b.Paint(buf)
}

func TestP100_UserMessage_Paint_Unicode(t *testing.T) {
	b := NewUserMessageBlock("test", "Hello 世界 café ☕ 日本語")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.Paint(buf)
}

func TestP100_UserMessage_Paint_NarrowWidth(t *testing.T) {
	b := NewUserMessageBlock("test", "some text that needs wrapping")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 8})
	buf := buffer.NewBuffer(10, 8)
	b.Paint(buf)
}

// ─── ErrorBlock Measure/Paint (90% → higher) ───

func TestP100_ErrorBlock_Measure_LongMessage(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "This is a very long error message that should still measure correctly even when truncated")
	size := b.Measure(component.Bounded(80, 100))
	if size.W <= 0 || size.H <= 0 {
		t.Errorf("Measure = (%d,%d), expected positive", size.W, size.H)
	}
}

func TestP100_ErrorBlock_Measure_NarrowWidth(t *testing.T) {
	b := NewErrorBlockWithMessage("err1", "error message")
	size := b.Measure(component.Bounded(10, 100))
	if size.W > 10 {
		t.Errorf("W = %d, should be <= 10", size.W)
	}
}

// ─── Container Measure edge cases (95.7% → higher) ───

func TestP100_Container_Measure_ZeroBlocks(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(1)
	size := c.Measure(component.Bounded(80, 100))
	_ = size
}

func TestP100_Container_Measure_Spacing(t *testing.T) {
	c := NewBlockContainer()
	c.SetSpacing(2)
	at := NewAssistantTextBlock("test")
	at.AppendDelta("content")
	c.AddBlock(at)
	c.AddBlock(NewUserMessageBlock("msg", "hello"))
	size := c.Measure(component.Bounded(80, 100))
	if size.H <= 0 {
		t.Errorf("H = %d, expected positive with spacing", size.H)
	}
}

// ─── ImageBlock Sequence (92.3% → higher) ───

func TestP100_ImageBlock_Sequence_Cached(t *testing.T) {
	b := NewImageBlock("img1", "test.png", []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A})
	seq1 := b.Sequence()
	// Second call should use cache
	seq2 := b.Sequence()
	if seq1 != seq2 {
		t.Error("cached sequence should match first call")
	}
}

func TestP100_ImageBlock_FormatDetection(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		data     []byte
		want     string
	}{
		{"png by ext", "photo.png", nil, "png"},
		{"png by magic", "unknown", []byte{0x89, 'P', 'N', 'G'}, "png"},
		{"jpeg by ext", "photo.jpg", nil, "jpeg"},
		{"jpeg by magic", "unknown", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "jpeg"},
		{"gif by ext", "anim.gif", nil, "gif"},
		{"gif by magic", "unknown", []byte{'G', 'I', 'F', '8'}, "gif"},
		{"bmp by ext", "old.bmp", nil, "bmp"},
		{"unknown", "noext", []byte{0x00, 0x01}, "unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := NewImageBlock("test", tc.filename, tc.data)
			if b.Format() != tc.want {
				t.Errorf("Format() = %q, want %q", b.Format(), tc.want)
			}
		})
	}
}

func TestP100_ImageBlock_FileSize(t *testing.T) {
	tests := []struct {
		bytes int
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
	}
	for _, tc := range tests {
		b := NewImageBlock("test", "x.png", make([]byte, tc.bytes))
		got := b.FileSize()
		if got != tc.want {
			t.Errorf("FileSize(%d) = %q, want %q", tc.bytes, got, tc.want)
		}
	}
}

func TestP100_ImageBlock_TruncateText(t *testing.T) {
	tests := []struct {
		text   string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hell…"},
		{"hi", 1, "…"},
		{"anything", 0, ""},
		{"short", -1, ""},
	}
	for _, tc := range tests {
		got := truncateText(tc.text, tc.maxLen)
		if got != tc.want {
			t.Errorf("truncateText(%q, %d) = %q, want %q", tc.text, tc.maxLen, got, tc.want)
		}
	}
}
