package component

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// P295: ToolCallView component tests

func TestP295_NewToolCallView(t *testing.T) {
	tc := NewToolCallView("read_file", `{"path":"/tmp/test.go"}`)
	if tc.ToolName() != "read_file" {
		t.Errorf("ToolName = %q, want read_file", tc.ToolName())
	}
	if tc.Status() != ToolCallRunning {
		t.Errorf("Status = %v, want Running", tc.Status())
	}
	if tc.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestP295_PrettyPrintArgs(t *testing.T) {
	// Valid JSON gets pretty-printed
	tc := NewToolCallView("exec", `{"cmd":"ls","args":["-la"]}`)
	tc.mu.Lock()
	if tc.prettyArg == tc.args {
		t.Error("expected pretty-printed args to differ from raw")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(tc.prettyArg), &parsed); err != nil {
		t.Errorf("prettyArg is not valid JSON: %v", err)
	}
	tc.mu.Unlock()
}

func TestP295_PrettyPrintArgs_NonJSON(t *testing.T) {
	tc := NewToolCallView("exec", "not json at all")
	if tc.Args() != "not json at all" {
		t.Errorf("Args = %q", tc.Args())
	}
}

func TestP295_PrettyPrintArgs_Empty(t *testing.T) {
	tc := NewToolCallView("exec", "")
	if tc.Args() != "" {
		t.Errorf("Args = %q, want empty", tc.Args())
	}
}

func TestP295_SetResult(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.SetResult("hello world")
	if tc.Result() != "hello world" {
		t.Errorf("Result = %q", tc.Result())
	}
}

func TestP295_AppendResult(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.AppendResult("hello ")
	tc.AppendResult("world")
	if tc.Result() != "hello world" {
		t.Errorf("Result = %q, want 'hello world'", tc.Result())
	}
}

func TestP295_Complete(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	time.Sleep(2 * time.Millisecond)
	tc.Complete()
	if tc.Status() != ToolCallCompleted {
		t.Errorf("Status = %v, want Completed", tc.Status())
	}
	if tc.Duration() < time.Millisecond {
		t.Errorf("Duration = %v, expected > 1ms", tc.Duration())
	}
}

func TestP295_Error(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.Error()
	if tc.Status() != ToolCallErrored {
		t.Errorf("Status = %v, want Errored", tc.Status())
	}
}

func TestP295_Toggle(t *testing.T) {
	tc := NewToolCallView("exec", `{"a":1}`)
	if tc.Expanded() {
		t.Error("should start collapsed")
	}
	tc.Toggle()
	if !tc.Expanded() {
		t.Error("should be expanded after toggle")
	}
	tc.Toggle()
	if tc.Expanded() {
		t.Error("should be collapsed after second toggle")
	}
}

func TestP295_SetExpanded(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.SetExpanded(true)
	if !tc.Expanded() {
		t.Error("should be expanded")
	}
	tc.SetExpanded(false)
	if tc.Expanded() {
		t.Error("should be collapsed")
	}
}

func TestP295_SetShowFull(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	if tc.ShowFull() {
		t.Error("should start not showing full")
	}
	tc.SetShowFull(true)
	if !tc.ShowFull() {
		t.Error("should show full")
	}
}

func TestP295_SetMaxResultPreview(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.SetMaxResultPreview(10)
	tc.mu.Lock()
	if tc.maxResultPreview != 10 {
		t.Errorf("maxResultPreview = %d, want 10", tc.maxResultPreview)
	}
	tc.mu.Unlock()
	// clamp to 1
	tc.SetMaxResultPreview(0)
	tc.mu.Lock()
	if tc.maxResultPreview != 1 {
		t.Errorf("maxResultPreview = %d, want 1 (clamped)", tc.maxResultPreview)
	}
	tc.mu.Unlock()
}

func TestP295_AdvanceSpinner(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.mu.Lock()
	f0 := tc.spinnerF
	tc.mu.Unlock()
	tc.AdvanceSpinner()
	tc.mu.Lock()
	if tc.spinnerF != f0+1 {
		t.Errorf("spinnerF = %d, want %d", tc.spinnerF, f0+1)
	}
	tc.mu.Unlock()
}

func TestP295_Measure_Collapsed(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls"}`)
	s := tc.Measure(Constraints{MaxWidth: 60, MaxHeight: 20})
	if s.H != 1 {
		t.Errorf("collapsed height = %d, want 1", s.H)
	}
	if s.W != 60 {
		t.Errorf("width = %d, want 60", s.W)
	}
}

func TestP295_Measure_Expanded(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls","flags":["-la","-h"]}`)
	tc.SetExpanded(true)
	s := tc.Measure(Constraints{MaxWidth: 60, MaxHeight: 20})
	// header(1) + border(1) + arg_lines + border(1)
	if s.H < 4 {
		t.Errorf("expanded height = %d, expected >= 4", s.H)
	}
}

func TestP295_Measure_WithResult(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls"}`)
	tc.SetResult("line1\nline2\nline3\nline4\nline5")
	s := tc.Measure(Constraints{MaxWidth: 60, MaxHeight: 30})
	// header(1) + result border(1) + preview(3) + border(1) + "show more"(1) = 7
	if s.H < 6 {
		t.Errorf("height with result = %d, expected >= 6", s.H)
	}
}

func TestP295_Measure_FullResult(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls"}`)
	tc.SetResult("a\nb\nc\nd\ne")
	tc.SetShowFull(true)
	s := tc.Measure(Constraints{MaxWidth: 60, MaxHeight: 30})
	// header(1) + result border(1) + 5 lines + border(1) = 8
	if s.H < 7 {
		t.Errorf("height full result = %d, expected >= 7", s.H)
	}
}

func TestP295_Measure_DefaultWidth(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	s := tc.Measure(Constraints{})
	if s.W != 80 {
		t.Errorf("default width = %d, want 80", s.W)
	}
}

func TestP295_Paint_RunningCollapsed(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"echo hello"}`)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	tc.Paint(buf)
}

func TestP295_Paint_Completed(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"echo hi"}`)
	tc.Complete()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	tc.Paint(buf)
}

func TestP295_Paint_Errored(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"fail"}`)
	tc.Error()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	tc.Paint(buf)
}

func TestP295_Paint_Expanded(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls","args":["-la","-h"]}`)
	tc.SetExpanded(true)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	tc.Paint(buf)
}

func TestP295_Paint_WithResult(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls"}`)
	tc.SetResult("file1.go\nfile2.go\nfile3.go\nfile4.go\nfile5.go\nfile6.go")
	tc.Complete()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	tc.Paint(buf)
}

func TestP295_Paint_WithResult_Full(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls"}`)
	tc.SetResult("a\nb\nc\nd\ne\nf")
	tc.SetShowFull(true)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	tc.Paint(buf)
}

func TestP295_Paint_ExpandedWithResult(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls","flags":["-la"]}`)
	tc.SetExpanded(true)
	tc.SetResult("file1.go\nfile2.go")
	tc.Complete()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	tc.Paint(buf)
}

func TestP295_Paint_ZeroBounds(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tc.Paint(buf) // should not panic
}

func TestP295_Paint_NonZeroOffset(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"ls"}`)
	tc.SetExpanded(true)
	tc.SetBounds(Rect{X: 5, Y: 3, W: 50, H: 15})
	buf := buffer.NewBuffer(60, 20)
	tc.Paint(buf)
}

func TestP295_Paint_NarrowWidth(t *testing.T) {
	tc := NewToolCallView("very_long_tool_name_here", `{"key":"value"}`)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	tc.Paint(buf)
}

func TestP295_SetArgs(t *testing.T) {
	tc := NewToolCallView("exec", "{}")
	tc.SetArgs(`{"new":"value"}`)
	if tc.Args() != `{"new":"value"}` {
		t.Errorf("Args = %q", tc.Args())
	}
}

func TestP295_FormatToolDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{time.Microsecond, "0ms"},
		{time.Millisecond, "1ms"},
		{500 * time.Millisecond, "500ms"},
		{time.Second, "1.0s"},
		{2500 * time.Millisecond, "2.5s"},
	}
	for _, tt := range tests {
		got := formatToolDuration(tt.d)
		if got != tt.want {
			t.Errorf("formatToolDuration(%v) = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestP295_Concurrent(t *testing.T) {
	tc := NewToolCallView("exec", `{"cmd":"test"}`)
	done := make(chan struct{})
	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			tc.AppendResult("x")
		}
		close(done)
	}()
	// Reader goroutine
	for i := 0; i < 50; i++ {
		_ = tc.Result()
		_ = tc.Status()
	}
	<-done
}

// Verify ToolCallView satisfies Component interface
func TestP295_SatisfiesComponent(t *testing.T) {
	var _ Component = (*ToolCallView)(nil)
}
