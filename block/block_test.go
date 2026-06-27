package block

import (
	"errors"
	"testing"
	"time"
)

func TestBlockStateString(t *testing.T) {
	tests := []struct {
		state BlockState
		want  string
	}{
		{BlockPending, "pending"},
		{BlockStreaming, "streaming"},
		{BlockComplete, "complete"},
		{BlockError, "error"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("BlockState(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestBlockTypeString(t *testing.T) {
	tests := []struct {
		bt   BlockType
		want string
	}{
		{TypeThinking, "thinking"},
		{TypeAssistantText, "assistant_text"},
		{TypeToolCall, "tool_call"},
		{TypeToolResult, "tool_result"},
		{TypeCode, "code"},
		{TypeError, "error"},
		{TypeUserMessage, "user_message"},
	}
	for _, tt := range tests {
		if got := tt.bt.String(); got != tt.want {
			t.Errorf("BlockType(%d).String() = %q, want %q", tt.bt, got, tt.want)
		}
	}
}

func TestBaseBlockCreation(t *testing.T) {
	bb := NewBaseBlock("test-1", TypeThinking)
	if bb.ID() != "test-1" {
		t.Errorf("ID() = %q, want 'test-1'", bb.ID())
	}
	if bb.Type() != TypeThinking {
		t.Errorf("Type() = %v, want TypeThinking", bb.Type())
	}
	if bb.State() != BlockStreaming {
		t.Errorf("State() = %v, want BlockStreaming", bb.State())
	}
	if !bb.IsDirty() {
		t.Error("new block should be dirty")
	}
	if bb.CreatedAt().IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestBaseBlockLifecycle(t *testing.T) {
	bb := NewBaseBlock("test-2", TypeAssistantText)

	// AppendDelta marks dirty
	bb.ClearDirty()
	bb.AppendDelta("hello")
	if !bb.IsDirty() {
		t.Error("AppendDelta should mark dirty")
	}

	// Complete sets state and endedAt
	bb.ClearDirty()
	bb.Complete()
	if bb.State() != BlockComplete {
		t.Errorf("State() = %v, want BlockComplete", bb.State())
	}
	if !bb.IsDirty() {
		t.Error("Complete should mark dirty")
	}
	if bb.endedAt.IsZero() {
		t.Error("Complete should set endedAt")
	}

	// Duration should be positive
	if bb.Duration() < 0 {
		t.Error("Duration should be non-negative after Complete")
	}
}

func TestBaseBlockFail(t *testing.T) {
	bb := NewBaseBlock("test-3", TypeToolCall)
	bb.ClearDirty()

	err := errors.New("network timeout")
	bb.Fail(err)

	if bb.State() != BlockError {
		t.Errorf("State() = %v, want BlockError", bb.State())
	}
	if !bb.IsDirty() {
		t.Error("Fail should mark dirty")
	}
	if bb.Error() != "network timeout" {
		t.Errorf("Error() = %q, want 'network timeout'", bb.Error())
	}
}

func TestBaseBlockSetState(t *testing.T) {
	bb := NewBaseBlock("test-4", TypeCode)
	bb.ClearDirty()

	bb.SetState(BlockComplete)
	if bb.State() != BlockComplete {
		t.Errorf("State() = %v, want BlockComplete", bb.State())
	}
	if !bb.IsDirty() {
		t.Error("SetState should mark dirty")
	}
}

func TestBaseBlockClearDirty(t *testing.T) {
	bb := NewBaseBlock("test-5", TypeThinking)
	bb.ClearDirty()
	if bb.IsDirty() {
		t.Error("ClearDirty should clear dirty flag")
	}
}

func TestBaseBlockString(t *testing.T) {
	bb := NewBaseBlock("test-6", TypeToolCall)
	s := bb.String()
	if s == "" {
		t.Error("String() should not be empty")
	}
}

func TestBaseBlockDurationWhileStreaming(t *testing.T) {
	bb := NewBaseBlock("test-7", TypeAssistantText)
	time.Sleep(1 * time.Millisecond)
	// While streaming (endedAt is zero), Duration should be time since start
	d := bb.Duration()
	if d <= 0 {
		t.Errorf("Duration while streaming should be positive, got %v", d)
	}
}
