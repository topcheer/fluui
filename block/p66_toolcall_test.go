package block

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestP66_NewToolCall_BasicFields(t *testing.T) {
	b := NewToolCallBlock("t1", "read_file", `{"path": "/tmp/test.go"}`)
	if b.ToolName() != "read_file" {
		t.Errorf("expected read_file, got %s", b.ToolName())
	}
	if b.RawArgs() != `{"path": "/tmp/test.go"}` {
		t.Error("unexpected raw args")
	}
	if b.Expanded() {
		t.Error("expected collapsed by default")
	}
}

func TestP66_NewToolCall_PrettyPrintsJSON(t *testing.T) {
	b := NewToolCallBlock("t1", "search", `{"query":"test","limit":10}`)
	if !b.hasPretty {
		t.Error("expected args to be pretty-printable JSON")
	}
	if b.prettyArg == "" {
		t.Error("expected non-empty pretty args")
	}
	// Pretty-printed should have newlines (indentation)
	if !contains(b.prettyArg, "\n") {
		t.Error("expected pretty args to contain newlines")
	}
}

func TestP66_NewToolCall_InvalidJSON(t *testing.T) {
	b := NewToolCallBlock("t1", "echo", `not json`)
	if b.hasPretty {
		t.Error("expected hasPretty=false for invalid JSON")
	}
	if b.prettyArg != "" {
		t.Error("expected empty prettyArg for invalid JSON")
	}
}

func TestP66_NewToolCall_EmptyArgs(t *testing.T) {
	b := NewToolCallBlock("t1", "list", "")
	if b.hasPretty {
		t.Error("expected hasPretty=false for empty args")
	}
}

func TestP66_Toggle(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	if b.Expanded() {
		t.Error("expected collapsed initially")
	}
	b.Toggle()
	if !b.Expanded() {
		t.Error("expected expanded after toggle")
	}
	b.Toggle()
	if b.Expanded() {
		t.Error("expected collapsed after second toggle")
	}
}

func TestP66_SetExpanded(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	b.SetExpanded(true)
	if !b.Expanded() {
		t.Error("expected expanded")
	}
	b.SetExpanded(false)
	if b.Expanded() {
		t.Error("expected collapsed")
	}
}

func TestP66_AdvanceSpinner(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	original := b.spinnerF
	b.AdvanceSpinner()
	if b.spinnerF != original+1 {
		t.Error("expected spinner to advance")
	}
	// Verify it cycles
	for i := 0; i < len(toolCallSpinnerFrames)+5; i++ {
		b.AdvanceSpinner()
	}
	// Should not panic and should be within range via modulo
	if b.spinnerF < 0 {
		t.Error("spinner should not be negative")
	}
}

func TestP66_Duration(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	d := b.Duration()
	if d < 0 {
		t.Error("expected non-negative duration")
	}
	// Should be close to 0 for a freshly created block
	if d > 100*time.Millisecond {
		t.Error("expected duration < 100ms for new block")
	}
}

func TestP66_Measure_Collapsed(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	s := b.Measure(component.Constraints{MaxWidth: 80})
	if s.H != 1 {
		t.Errorf("expected H=1 when collapsed, got %d", s.H)
	}
	if s.W != 80 {
		t.Errorf("expected W=80, got %d", s.W)
	}
}

func TestP66_Measure_Expanded(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"key": "value", "num": 42}`)
	b.SetExpanded(true)
	s := b.Measure(component.Constraints{MaxWidth: 80})
	if s.H <= 1 {
		t.Error("expected H > 1 when expanded")
	}
}

func TestP66_Measure_ExpandedEmptyArgs(t *testing.T) {
	b := NewToolCallBlock("t1", "test", "")
	b.SetExpanded(true)
	s := b.Measure(component.Constraints{MaxWidth: 80})
	if s.H != 1 {
		t.Errorf("expected H=1 when expanded but no args, got %d", s.H)
	}
}

func TestP66_Measure_DefaultMaxWidth(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	s := b.Measure(component.Constraints{})
	if s.W != 80 {
		t.Errorf("expected default W=80, got %d", s.W)
	}
}

func TestP66_Paint_Collapsed(t *testing.T) {
	b := NewToolCallBlock("t1", "read_file", `{"path":"/tmp"}`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
	// Should have non-empty content at row 0
	cell := buf.GetCell(0, 0)
	if cell.Rune == ' ' || cell.Rune == 0 {
		t.Error("expected non-empty content at (0,0)")
	}
}

func TestP66_Paint_CollapsedComplete(t *testing.T) {
	b := NewToolCallBlock("t1", "read_file", `{"path":"/tmp"}`)
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
	// Should show ✓ icon
	found := false
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '✓' {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ✓ icon in completed tool call")
	}
}

func TestP66_Paint_CollapsedError(t *testing.T) {
	b := NewToolCallBlock("t1", "read_file", `{"path":"/tmp"}`)
	b.Fail(errTool("file not found"))
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
	// Should show ✗ icon
	found := false
	for x := 0; x < 10; x++ {
		c := buf.GetCell(x, 0)
		if c.Rune == '✗' {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ✗ icon in failed tool call")
	}
}

func TestP66_Paint_Expanded(t *testing.T) {
	b := NewToolCallBlock("t1", "search", `{"query":"test","limit":10}`)
	b.SetExpanded(true)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.Paint(buf)
	// Should have border characters
	foundBorder := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 60; x++ {
			c := buf.GetCell(x, y)
			if c.Rune == '╭' || c.Rune == '╮' || c.Rune == '╰' || c.Rune == '╯' {
				foundBorder = true
			}
		}
	}
	if !foundBorder {
		t.Error("expected border characters in expanded view")
	}
}

func TestP66_Paint_ExpandedInvalidJSON(t *testing.T) {
	b := NewToolCallBlock("t1", "echo", `plain text args`)
	b.SetExpanded(true)
	// SetBounds with enough height
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
	// Should show raw text in bordered area
}

func TestP66_Paint_ZeroBounds(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(10, 5)
	b.Paint(buf) // should not panic
}

func TestP66_SerializeState(t *testing.T) {
	b := NewToolCallBlock("t1", "read_file", `{"path":"/tmp/test"}`)
	b.SetExpanded(true)
	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState failed: %v", err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	if m["tool_name"] != "read_file" {
		t.Error("expected tool_name in serialized state")
	}
	if m["expanded"] != true {
		t.Error("expected expanded=true in serialized state")
	}
}

func TestP66_DeserializeState(t *testing.T) {
	jsonData := `{"tool_name":"write_file","args":"{\"path\":\"/out\"}","expanded":true}`
	b := NewToolCallBlock("t0", "", "")
	err := b.DeserializeState(json.RawMessage(jsonData))
	if err != nil {
		t.Fatalf("DeserializeState failed: %v", err)
	}
	if b.ToolName() != "write_file" {
		t.Errorf("expected write_file, got %s", b.ToolName())
	}
	if !b.Expanded() {
		t.Error("expected expanded=true")
	}
}

func TestP66_DeserializeState_RePrettyPrints(t *testing.T) {
	jsonData := `{"tool_name":"test","args":"{\"a\":1}","expanded":false}`
	b := NewToolCallBlock("t0", "", "")
	err := b.DeserializeState(json.RawMessage(jsonData))
	if err != nil {
		t.Fatalf("DeserializeState failed: %v", err)
	}
	if !b.hasPretty {
		t.Error("expected re-pretty-printed args after deserialize")
	}
}

func TestP66_PaintWithDuration(t *testing.T) {
	b := NewToolCallBlock("t1", "long_task", `{"id":123}`)
	b.Complete()
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 5)
	b.Paint(buf)
	// Should render without panic and contain non-empty output
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 || cell.Rune == ' ' {
		t.Error("expected non-empty output")
	}
}

func TestP66_TruncateRunes(t *testing.T) {
	if truncateRunes("hello", 10) != "hello" {
		t.Error("expected no truncation for short string")
	}
	if truncateRunes("hello", 3) != "he…" {
		t.Error("expected he… for truncation to 3")
	}
	if truncateRunes("hello", 1) != "…" {
		t.Error("expected … for truncation to 1")
	}
}

func TestP66_RuneCount(t *testing.T) {
	if runeCount("hello") != 5 {
		t.Error("expected 5 for hello")
	}
	if runeCount("héllo") != 5 {
		t.Error("expected 5 for héllo")
	}
	if runeCount("") != 0 {
		t.Error("expected 0 for empty")
	}
}

func TestP66_ConcurrentPaintAndToggle(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 10})

	done := make(chan struct{})
	// Concurrent toggler
	go func() {
		for i := 0; i < 100; i++ {
			b.Toggle()
		}
		close(done)
	}()

	// Concurrent painter
	for i := 0; i < 100; i++ {
		buf := buffer.NewBuffer(40, 10)
		b.Paint(buf)
	}
	<-done
}

func TestP66_ConcurrentAdvanceSpinner(t *testing.T) {
	b := NewToolCallBlock("t1", "test", `{"a":1}`)
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 1})

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			b.AdvanceSpinner()
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		buf := buffer.NewBuffer(40, 1)
		b.Paint(buf)
	}
	<-done
}

// helper
type errTool string

func (e errTool) Error() string { return string(e) }
func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
