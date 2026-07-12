package block

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSubAgentBlock_Basic(t *testing.T) {
	b := NewSubAgentBlock("sa1", "agent-001", "Researcher", "Analyze codebase")
	if b.AgentID() != "agent-001" {
		t.Fatalf("expected agent-001, got %s", b.AgentID())
	}
	if b.AgentName() != "Researcher" {
		t.Fatalf("expected Researcher, got %s", b.AgentName())
	}
	if b.Task() != "Analyze codebase" {
		t.Fatalf("expected 'Analyze codebase', got %s", b.Task())
	}
	if b.Status() != "running" {
		t.Fatalf("expected running, got %s", b.Status())
	}
}

func TestSubAgentBlock_SetStatus(t *testing.T) {
	b := NewSubAgentBlock("sa1", "a1", "Bot", "Task")
	b.SetStatus("completed")
	if b.Status() != "completed" {
		t.Fatalf("expected completed, got %s", b.Status())
	}
}

func TestSubAgentBlock_AddMessage(t *testing.T) {
	b := NewSubAgentBlock("sa1", "a1", "Bot", "Task")
	if b.MessageCount() != 0 {
		t.Fatal("should start with 0 messages")
	}

	msg := SubAgentMessage{
		Role:    "user",
		Content: "Do the thing",
		Time:    time.Now(),
	}
	b.AddMessage(msg)

	if b.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", b.MessageCount())
	}

	msgs := b.Messages()
	if len(msgs) != 1 || msgs[0].Content != "Do the thing" {
		t.Fatal("message content mismatch")
	}
}

func TestSubAgentBlock_SerializeDeserialize(t *testing.T) {
	b := NewSubAgentBlock("sa1", "a1", "Agent", "Do work")
	b.AddMessage(SubAgentMessage{
		Role:    "assistant",
		Content: "Working on it",
		Time:    time.Now(),
	})
	b.SetStatus("completed")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatal(err)
	}

	b2 := NewSubAgentBlock("sa1", "", "", "")
	if err := b2.DeserializeState(data); err != nil {
		t.Fatal(err)
	}

	if b2.AgentName() != "Agent" || b2.Task() != "Do work" || b2.Status() != "completed" {
		t.Fatal("deserialized state mismatch")
	}
	if b2.MessageCount() != 1 {
		t.Fatal("expected 1 message after deserialize")
	}
}

func TestSubAgentBlock_DeserializeInvalid(t *testing.T) {
	b := NewSubAgentBlock("sa1", "a1", "A", "T")
	err := b.DeserializeState(json.RawMessage(`{invalid json`))
	if err == nil {
		t.Fatal("expected error on invalid JSON")
	}
}
