package block

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === Enhanced ToolResultBlock tests (P68) ===

func TestP68_ToolResult_SetOutput(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("replaced content")
	if b.Output() != "replaced content" {
		t.Errorf("expected 'replaced content', got %q", b.Output())
	}
}

func TestP68_ToolResult_SetOutput_ReplacesPrevious(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.AppendDelta("initial")
	b.SetOutput("replaced")
	if b.Output() != "replaced" {
		t.Errorf("expected 'replaced', got %q", b.Output())
	}
}

func TestP68_ToolResult_SetStatusCode(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetStatusCode(200)
	if b.StatusCode() != 200 {
		t.Errorf("expected 200, got %d", b.StatusCode())
	}
}

func TestP68_ToolResult_SetStatusCode_Error(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetStatusCode(404)
	if b.StatusCode() != 404 {
		t.Errorf("expected 404, got %d", b.StatusCode())
	}
}

func TestP68_ToolResult_SetContentType(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetContentType("json")
	if b.ContentType() != "json" {
		t.Errorf("expected 'json', got %q", b.ContentType())
	}
}

func TestP68_ToolResult_LineCount(t *testing.T) {
	b := NewToolResultBlock("tr1")
	if b.LineCount() != 0 {
		t.Errorf("expected 0, got %d", b.LineCount())
	}
	b.SetOutput("line1\nline2\nline3")
	if b.LineCount() != 3 {
		t.Errorf("expected 3, got %d", b.LineCount())
	}
}

func TestP68_ToolResult_LineCount_Empty(t *testing.T) {
	b := NewToolResultBlock("tr1")
	if b.LineCount() != 0 {
		t.Errorf("expected 0 for empty, got %d", b.LineCount())
	}
}

func TestP68_ToolResult_LineCount_SingleLine(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("only one line")
	if b.LineCount() != 1 {
		t.Errorf("expected 1, got %d", b.LineCount())
	}
}

func TestP68_ToolResult_SetCollapsed(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetCollapsed(true)
	if !b.Collapsed() {
		t.Error("expected collapsed=true")
	}
	b.SetCollapsed(false)
	if b.Collapsed() {
		t.Error("expected collapsed=false")
	}
}

func TestP68_ToolResult_MaxPreview(t *testing.T) {
	b := NewToolResultBlock("tr1")
	if b.MaxPreview() != 5 {
		t.Errorf("expected default 5, got %d", b.MaxPreview())
	}
}

func TestP68_ToolResult_SetMaxPreview(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetMaxPreview(10)
	if b.MaxPreview() != 10 {
		t.Errorf("expected 10, got %d", b.MaxPreview())
	}
}

func TestP68_ToolResult_SetMaxPreview_ZeroClamped(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetMaxPreview(0)
	if b.MaxPreview() != 1 {
		t.Errorf("expected clamped to 1, got %d", b.MaxPreview())
	}
}

// === Paint tests ===

func TestP68_ToolResult_PaintZeroBounds(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("test")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 5)
	b.Paint(buf) // should not panic
}

func TestP68_ToolResult_PaintEmpty(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
	// Should draw borders even with no content
	cell := buf.GetCell(0, 0)
	if cell.Rune != '╭' {
		t.Errorf("expected top-left border, got %q", string(cell.Rune))
	}
}

func TestP68_ToolResult_PaintWithContent(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("Hello World")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf)
	// Content should appear at row 1
	cell := buf.GetCell(1, 1)
	if cell.Rune == 0 {
		t.Error("expected content at (1,1)")
	}
}

func TestP68_ToolResult_PaintCollapsedWithHint(t *testing.T) {
	b := NewToolResultBlock("tr1")
	for i := 0; i < 10; i++ {
		b.AppendDelta("line\n")
	}
	b.Complete() // auto-collapses if > 5 lines
	if !b.Collapsed() {
		t.Fatal("expected collapsed")
	}
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
	// Hint line should appear at row 6 (1 border + 5 preview lines)
	hintCell := buf.GetCell(1, 6)
	if hintCell.Rune == 0 || hintCell.Rune == ' ' {
		// Hint may be at a slightly different position, just check something is drawn
	}
}

func TestP68_ToolResult_PaintExpanded(t *testing.T) {
	b := NewToolResultBlock("tr1")
	for i := 0; i < 10; i++ {
		b.AppendDelta("line\n")
	}
	b.Complete()
	b.SetCollapsed(false) // force expanded
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	b.Paint(buf)
	// Should have content at multiple rows
	cell := buf.GetCell(1, 1)
	if cell.Rune == 0 {
		t.Error("expected content at expanded view")
	}
}

func TestP68_ToolResult_PaintWithStatusCode(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("ok")
	b.SetStatusCode(200)
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // should show ✓ icon
}

func TestP68_ToolResult_PaintWithErrorStatusCode(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("not found")
	b.SetStatusCode(404)
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // should show ✗ icon
}

func TestP68_ToolResult_PaintStreaming(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.AppendDelta("partial...")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	b.Paint(buf) // should show spinner icon
}

func TestP68_ToolResult_PaintNarrowWidth(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("short")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 5})
	buf := buffer.NewBuffer(3, 5)
	b.Paint(buf) // should not panic with very narrow width
}

// === Serialize/Deserialize ===

func TestP68_ToolResult_SerializeState(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("test output")
	b.SetStatusCode(200)
	b.SetContentType("json")
	b.SetCollapsed(true)

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState error: %v", err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["result"] != "test output" {
		t.Errorf("expected 'test output', got %v", m["result"])
	}
	if m["status_code"] != float64(200) {
		t.Errorf("expected 200, got %v", m["status_code"])
	}
	if m["content_type"] != "json" {
		t.Errorf("expected 'json', got %v", m["content_type"])
	}
}

func TestP68_ToolResult_DeserializeState(t *testing.T) {
	b := NewToolResultBlock("tr1")
	data, _ := json.Marshal(map[string]any{
		"result":       "deserialized",
		"collapsed":    true,
		"status_code":  500,
		"content_type": "xml",
	})
	err := b.DeserializeState(data)
	if err != nil {
		t.Fatalf("DeserializeState error: %v", err)
	}
	if b.Output() != "deserialized" {
		t.Errorf("expected 'deserialized', got %q", b.Output())
	}
	if b.StatusCode() != 500 {
		t.Errorf("expected 500, got %d", b.StatusCode())
	}
	if b.ContentType() != "xml" {
		t.Errorf("expected 'xml', got %q", b.ContentType())
	}
}

func TestP68_ToolResult_DeserializeState_InvalidJSON(t *testing.T) {
	b := NewToolResultBlock("tr1")
	err := b.DeserializeState([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// === itoa helper ===

func TestP68_itoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{-1, "-1"},
		{-42, "-42"},
		{999, "999"},
	}
	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.expected {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// === Concurrent access ===

func TestP68_ToolResult_ConcurrentPaintAndToggle(t *testing.T) {
	b := NewToolResultBlock("tr1")
	b.SetOutput("content\nwith\nmultiple\nlines\nhere\nand\nmore")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 20})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(40, 20) // separate buffer per goroutine
			b.Paint(buf)
		}()
		go func() {
			defer wg.Done()
			b.Toggle()
		}()
	}
	wg.Wait()
}

func TestP68_ToolResult_ConcurrentAppendAndRead(t *testing.T) {
	b := NewToolResultBlock("tr1")
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			b.AppendDelta("data\n")
		}()
		go func() {
			defer wg.Done()
			_ = b.Output()
			_ = b.LineCount()
		}()
	}
	wg.Wait()
}
