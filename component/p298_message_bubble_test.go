package component

import (
	"strings"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// === MessageBubble tests ===

func TestP298_NewMessageBubble(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "Hello")
	if mb.Role() != RoleUser {
		t.Errorf("Role = %v", mb.Role())
	}
	if mb.Content() != "Hello" {
		t.Errorf("Content = %q", mb.Content())
	}
	if mb.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestP298_MessageRole_String(t *testing.T) {
	tests := []struct {
		role MessageRole
		want string
	}{
		{RoleUser, "You"},
		{RoleAssistant, "Assistant"},
		{RoleSystem, "System"},
		{RoleTool, "Tool"},
	}
	for _, tt := range tests {
		if got := tt.role.String(); got != tt.want {
			t.Errorf("Role(%d).String() = %q, want %q", tt.role, got, tt.want)
		}
	}
}

func TestP298_MessageRoleFromString(t *testing.T) {
	tests := []struct {
		input string
		want  MessageRole
	}{
		{"user", RoleUser},
		{"Human", RoleUser},
		{"assistant", RoleAssistant},
		{"AI", RoleAssistant},
		{"system", RoleSystem},
		{"tool", RoleTool},
		{"unknown", RoleAssistant}, // default
	}
	for _, tt := range tests {
		if got := MessageRoleFromString(tt.input); got != tt.want {
			t.Errorf("MessageRoleFromString(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestP298_SetContent(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "initial")
	mb.SetContent("updated")
	if mb.Content() != "updated" {
		t.Errorf("Content = %q", mb.Content())
	}
}

func TestP298_SetStreaming(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "hi")
	mb.SetStreaming(true)
	if !mb.Streaming() {
		t.Error("should be streaming")
	}
	mb.SetStreaming(false)
	if mb.Streaming() {
		t.Error("should not be streaming")
	}
}

func TestP298_SetError(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "oops")
	mb.SetError(true)
	if !mb.HasError() {
		t.Error("should have error")
	}
	mb.SetError(false)
	if mb.HasError() {
		t.Error("should not have error")
	}
}

func TestP298_SetModel(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "response")
	mb.SetModel("gpt-4")
	if mb.Model() != "gpt-4" {
		t.Errorf("Model = %q", mb.Model())
	}
}

func TestP298_SetAvatar(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "hi")
	mb.SetAvatar("🧑")
	if mb.Avatar() != "🧑" {
		t.Errorf("Avatar = %q", mb.Avatar())
	}
}

func TestP298_DefaultAvatar(t *testing.T) {
	if defaultAvatar(RoleUser) == "" {
		t.Error("user avatar should not be empty")
	}
	if defaultAvatar(RoleAssistant) == "" {
		t.Error("assistant avatar should not be empty")
	}
}

func TestP298_Timestamp(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "test")
	ts := time.Now().Add(-1 * time.Hour)
	mb.SetTimestamp(ts)
	if !mb.Timestamp().Equal(ts) {
		t.Errorf("Timestamp = %v, want %v", mb.Timestamp(), ts)
	}
}

func TestP298_ToggleCursor(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "streaming")
	mb.ToggleCursor()
	mb.mu.Lock()
	on1 := mb.cursorOn
	mb.mu.Unlock()
	mb.ToggleCursor()
	mb.mu.Lock()
	on2 := mb.cursorOn
	mb.mu.Unlock()
	if on1 == on2 {
		t.Error("cursor should toggle")
	}
}

func TestP298_Measure_Simple(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "Hello world")
	s := mb.Measure(Constraints{MaxWidth: 80})
	// header(1) + content(1) = 2
	if s.H < 2 {
		t.Errorf("H = %d, expected >= 2", s.H)
	}
}

func TestP298_Measure_LongContent(t *testing.T) {
	long := strings.Repeat("word ", 50)
	mb := NewMessageBubble(RoleAssistant, long)
	s := mb.Measure(Constraints{MaxWidth: 40})
	if s.H <= 2 {
		t.Errorf("H = %d, expected wrapping to produce many lines", s.H)
	}
}

func TestP298_Measure_Empty(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "")
	s := mb.Measure(Constraints{MaxWidth: 80})
	if s.H < 1 {
		t.Errorf("H = %d, expected >= 1", s.H)
	}
}

func TestP298_Measure_Error(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "fail")
	mb.SetError(true)
	s := mb.Measure(Constraints{MaxWidth: 80})
	// header(1) + content(1) + error(1) = 3
	if s.H < 3 {
		t.Errorf("H = %d, expected >= 3 with error", s.H)
	}
}

func TestP298_Measure_DefaultWidth(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "test")
	s := mb.Measure(Constraints{})
	if s.W != 80 {
		t.Errorf("default W = %d, want 80", s.W)
	}
}

func TestP298_Paint_User(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "Hello there!")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	mb.Paint(buf)
}

func TestP298_Paint_Assistant(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "Hi! How can I help?")
	mb.SetModel("claude-3")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	mb.Paint(buf)
}

func TestP298_Paint_AssistantStreaming(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "Let me think about")
	mb.SetStreaming(true)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	mb.Paint(buf)
	mb.ToggleCursor()
	mb.Paint(buf)
}

func TestP298_Paint_StreamingEmpty(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "")
	mb.SetStreaming(true)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	mb.Paint(buf) // should show cursor only, not panic
}

func TestP298_Paint_System(t *testing.T) {
	mb := NewMessageBubble(RoleSystem, "System message about the conversation")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	mb.Paint(buf)
}

func TestP298_Paint_Tool(t *testing.T) {
	mb := NewMessageBubble(RoleTool, "Tool output result here")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 3})
	buf := buffer.NewBuffer(60, 3)
	mb.Paint(buf)
}

func TestP298_Paint_Error(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "Something went wrong")
	mb.SetError(true)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	mb.Paint(buf)
}

func TestP298_Paint_LongContent(t *testing.T) {
	long := strings.Repeat("The quick brown fox ", 10)
	mb := NewMessageBubble(RoleAssistant, long)
	mb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	mb.Paint(buf)
}

func TestP298_Paint_ZeroBounds(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "test")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	mb.Paint(buf) // should not panic
}

func TestP298_Paint_NonZeroOffset(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "test")
	mb.SetBounds(Rect{X: 5, Y: 3, W: 50, H: 5})
	buf := buffer.NewBuffer(60, 10)
	mb.Paint(buf)
}

func TestP298_Paint_NarrowWidth(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "very long text that overflows")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	mb.Paint(buf)
}

func TestP298_String(t *testing.T) {
	mb := NewMessageBubble(RoleUser, "Hello")
	s := mb.String()
	if !strings.Contains(s, "You") {
		t.Errorf("String() = %q, should contain role", s)
	}
}

func TestP298_WrapLines(t *testing.T) {
	count := countWrapLines("hello world foo bar", 10)
	if count < 2 {
		t.Errorf("expected wrapping, got %d lines", count)
	}
}

func TestP298_WrapLines_Newlines(t *testing.T) {
	count := countWrapLines("line1\nline2\nline3", 80)
	if count != 3 {
		t.Errorf("expected 3 lines, got %d", count)
	}
}

func TestP298_WrapLines_Empty(t *testing.T) {
	count := countWrapLines("", 80)
	if count != 1 {
		t.Errorf("expected 1 line for empty, got %d", count)
	}
}

func TestP298_Concurrent(t *testing.T) {
	mb := NewMessageBubble(RoleAssistant, "init")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			mb.SetContent("update")
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		_ = mb.Content()
		_ = mb.Streaming()
	}
	<-done
}

func TestP298_SatisfiesComponent(t *testing.T) {
	var _ Component = (*MessageBubble)(nil)
}

// drawStyledText: truncation path
func TestP298_DrawStyledText_Truncate(t *testing.T) {
	tc := NewToolCallView("test", "{}")
	buf := buffer.NewBuffer(10, 1)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	// Trigger drawStyledText via paintCollapsed with long text
	tc.mu.Lock()
	tc.drawStyledText(buf, 0, 0, 3, "hello world", buffer.DefaultStyle)
	tc.mu.Unlock()
}

// contentWidthLocked: system narrow, non-system narrow
func TestP298_ContentWidth_Edges(t *testing.T) {
	// System with maxW < 2
	mb := NewMessageBubble(RoleSystem, "x")
	mb.mu.Lock()
	if w := mb.contentWidthLocked(1); w != 1 {
		t.Errorf("system maxW=1: got %d, want 1", w)
	}
	if w := mb.contentWidthLocked(10); w != 8 {
		t.Errorf("system maxW=10: got %d, want 8", w)
	}
	mb.mu.Unlock()

	// User with maxW < 4
	mb2 := NewMessageBubble(RoleUser, "x")
	mb2.mu.Lock()
	if w := mb2.contentWidthLocked(2); w != 2 {
		t.Errorf("user maxW=2: got %d, want 2", w)
	}
	if w := mb2.contentWidthLocked(10); w != 6 {
		t.Errorf("user maxW=10: got %d, want 6", w)
	}
	mb2.mu.Unlock()
}
