package block

import (
	"encoding/json"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestP153_SubAgent_Measure(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "Starting build..."})
	b.AddMessage(SubAgentMessage{Role: "user", Content: "Go ahead"})
	size := b.Measure(component.Constraints{MaxWidth: 80, MaxHeight: 20})
	if size.W <= 0 || size.H <= 0 {
		t.Errorf("expected positive size, got %+v", size)
	}
}

func TestP153_SubAgent_Measure_NoMessages(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	size := b.Measure(component.Constraints{MaxWidth: 80, MaxHeight: 20})
	if size.H < 2 {
		t.Errorf("expected at least 2 lines (header), got %d", size.H)
	}
}

func TestP153_SubAgent_Measure_ZeroWidth(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	size := b.Measure(component.Constraints{})
	if size.W != 80 {
		t.Errorf("expected default W=80, got %d", size.W)
	}
}

func TestP153_SubAgent_SetBounds(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	// Should not panic, should set bounds
}

func TestP153_SubAgent_Paint(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "Building..."})
	b.AddMessage(SubAgentMessage{Role: "user", Content: "OK go ahead"})
	b.SetStatus("running")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.Paint(buf)
}

func TestP153_SubAgent_Paint_AllStatuses(t *testing.T) {
	statuses := []string{"pending", "running", "completed", "failed", "unknown"}
	for _, status := range statuses {
		b := NewSubAgentBlock("block1", "agent1", "Worker", "Task: "+status)
		b.SetStatus(status)
		b.AddMessage(SubAgentMessage{Role: "assistant", Content: "Message"})
		b.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 15})
		buf := buffer.NewBuffer(60, 15)
		b.Paint(buf)
	}
}

func TestP153_SubAgent_Paint_NarrowWidth(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "A very long message that needs truncation"})
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 10, H: 10})
	buf := buffer.NewBuffer(10, 10)
	b.Paint(buf)
}

func TestP153_SubAgent_Paint_ZeroBounds(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(60, 15)
	b.Paint(buf) // should not panic
}

func TestP153_SubAgent_Paint_LongContent(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project with a very long task name")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "A very very very long message that should be truncated when rendered"})
	b.AddMessage(SubAgentMessage{Role: "user", Content: "Short"})
	b.SetBounds(component.Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	b.Paint(buf)
}

func TestP153_SubAgent_Paint_NonZeroOffset(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "Hello"})
	b.SetBounds(component.Rect{X: 5, Y: 3, W: 50, H: 10})
	buf := buffer.NewBuffer(60, 15)
	b.Paint(buf)
}

func TestP153_SubAgent_FirstVisibleMessage(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "msg1"})
	b.AddMessage(SubAgentMessage{Role: "user", Content: "msg2"})
	if b.firstVisibleMessage(80) != 0 {
		t.Error("expected 0")
	}
}

func TestP153_SubAgent_SerializeDeserialize(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.SetStatus("running")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "Building..."})
	b.AddMessage(SubAgentMessage{Role: "user", Content: "OK"})

	state, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	b2 := NewSubAgentBlock("b1","","","")
	if err := b2.DeserializeState(state); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}
	if b2.AgentID() != "agent1" {
		t.Errorf("expected agentID 'agent1', got %q", b2.AgentID())
	}
	if b2.AgentName() != "Worker" {
		t.Errorf("expected agentName 'Worker', got %q", b2.AgentName())
	}
	if b2.Status() != "running" {
		t.Errorf("expected status 'running', got %q", b2.Status())
	}
	if b2.MessageCount() != 2 {
		t.Errorf("expected 2 messages, got %d", b2.MessageCount())
	}
}

func TestP153_SubAgent_DeserializeInvalidJSON(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	err := b.DeserializeState(json.RawMessage("invalid"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestP153_SubAgent_Getters(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	if b.AgentID() != "agent1" {
		t.Errorf("expected 'agent1', got %q", b.AgentID())
	}
	if b.AgentName() != "Worker" {
		t.Errorf("expected 'Worker', got %q", b.AgentName())
	}
	if b.Task() != "Build project" {
		t.Errorf("expected 'Build project', got %q", b.Task())
	}
	if b.Status() != "running" {
		t.Errorf("expected 'running', got %q", b.Status())
	}
}

func TestP153_SubAgent_SetStatus(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.SetStatus("completed")
	if b.Status() != "completed" {
		t.Errorf("expected 'completed', got %q", b.Status())
	}
}

func TestP153_SubAgent_Messages(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	b.AddMessage(SubAgentMessage{Role: "assistant", Content: "Hello"})
	b.AddMessage(SubAgentMessage{Role: "user", Content: "Hi"})
	msgs := b.Messages()
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
	if msgs[0].Role != "assistant" || msgs[0].Content != "Hello" {
		t.Errorf("unexpected msg[0]: %+v", msgs[0])
	}
}

func TestP153_SubAgent_MessageCount(t *testing.T) {
	b := NewSubAgentBlock("block1", "agent1", "Worker", "Build project")
	if b.MessageCount() != 0 {
		t.Errorf("expected 0, got %d", b.MessageCount())
	}
	b.AddMessage(SubAgentMessage{Role: "user", Content: "test"})
	if b.MessageCount() != 1 {
		t.Errorf("expected 1, got %d", b.MessageCount())
	}
}