package block

import (
	"errors"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestToolCallCreation(t *testing.T) {
	b := NewToolCallBlock("tc-1", "ReadFile", `{"path":"/tmp/a.go"}`)

	if b.ID() != "tc-1" {
		t.Errorf("ID() = %q, want 'tc-1'", b.ID())
	}
	if b.Type() != TypeToolCall {
		t.Errorf("Type() = %v, want TypeToolCall", b.Type())
	}
	if b.State() != BlockStreaming {
		t.Errorf("State() = %v, want BlockStreaming (NewToolCallBlock starts streaming)", b.State())
	}
	if b.ToolName() != "ReadFile" {
		t.Errorf("ToolName() = %q, want 'ReadFile'", b.ToolName())
	}
	if b.RawArgs() != `{"path":"/tmp/a.go"}` {
		t.Errorf("RawArgs() = %q, want raw args JSON", b.RawArgs())
	}
	if !b.IsDirty() {
		t.Error("new block should be dirty")
	}
}

func TestToolCallMeasure(t *testing.T) {
	b := NewToolCallBlock("tc-2", "WriteFile", `{}`)

	// With explicit max width
	size := b.Measure(component.Constraints{MaxWidth: 60})
	if size.H != 1 {
		t.Errorf("Measure H = %d, want 1", size.H)
	}
	if size.W != 60 {
		t.Errorf("Measure W = %d, want 60", size.W)
	}

	// With zero width — falls back to 80
	size = b.Measure(component.Unbounded())
	if size.W != 80 {
		t.Errorf("Measure W (unbounded) = %d, want 80", size.W)
	}
}

func TestToolCallPaintStreaming(t *testing.T) {
	b := NewToolCallBlock("tc-3", "Grep", `"pattern"`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == ' ' || cell.Rune == 0 {
		t.Errorf("Paint (streaming): cell (0,0) should have content, got rune %q", cell.Rune)
	}
}

func TestToolCallPaintComplete(t *testing.T) {
	b := NewToolCallBlock("tc-4", "Edit", `"file.go"`)
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	// Find the checkmark ✓ in the painted output
	found := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '✓' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Paint (complete): expected ✓ in buffer")
	}
}

func TestToolCallPaintError(t *testing.T) {
	b := NewToolCallBlock("tc-5", "Bash", `"ls"`)
	b.Fail(errors.New("command not found"))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	buf := buffer.NewBuffer(40, 1)
	b.Paint(buf)

	// Find the cross ✗ in the painted output
	found := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '✗' {
			found = true
			break
		}
	}
	if !found {
		t.Error("Paint (error): expected ✗ in buffer")
	}
}
