package block

import (
	"testing"
)

func TestRegistryEmpty(t *testing.T) {
	r := NewRegistry()

	if len(r.Types()) != 0 {
		t.Fatalf("expected 0 types, got %d", len(r.Types()))
	}
	if r.Has("error") {
		t.Fatal("expected Has(\"error\") to be false")
	}
}

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	r.Register("error", func(id string) Block {
		return NewErrorBlock(id)
	})

	if !r.Has("error") {
		t.Fatal("expected Has(\"error\") to be true")
	}
	if len(r.Types()) != 1 {
		t.Fatalf("expected 1 type, got %d", len(r.Types()))
	}
}

func TestRegistryCreate(t *testing.T) {
	r := NewRegistry()
	r.Register("error", func(id string) Block {
		return NewErrorBlock(id)
	})

	blk, err := r.Create("error", "test-err")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	eb, ok := blk.(*ErrorBlock)
	if !ok {
		t.Fatal("expected *ErrorBlock")
	}
	if eb.ID() != "test-err" {
		t.Fatalf("expected ID 'test-err', got %q", eb.ID())
	}
	if eb.Type() != TypeError {
		t.Fatalf("expected type %s, got %s", TypeError, eb.Type())
	}
}

func TestRegistryUnknownType(t *testing.T) {
	r := NewRegistry()

	_, err := r.Create("error", "test-err")
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestRegistryReplace(t *testing.T) {
	r := NewRegistry()

	callCount := 0
	r.Register("error", func(id string) Block {
		callCount++
		return NewErrorBlock(id)
	})

	// Replace factory
	r.Register("error", func(id string) Block {
		callCount += 100
		return NewErrorBlock(id)
	})

	r.Create("error", "id")
	if callCount != 100 {
		t.Fatalf("expected replaced factory to be called (callCount=100), got %d", callCount)
	}
}

func TestRegistryMultipleTypes(t *testing.T) {
	r := NewRegistry()
	r.Register("thinking", func(id string) Block { return NewThinkingBlock(id) })
	r.Register("tool_result", func(id string) Block { return NewToolResultBlock(id) })
	r.Register("error", func(id string) Block { return NewErrorBlock(id) })

	if len(r.Types()) != 3 {
		t.Fatalf("expected 3 types, got %d", len(r.Types()))
	}

	// Create each type
	blk1, err := r.Create("thinking", "t1")
	if err != nil {
		t.Fatalf("thinking: unexpected error: %v", err)
	}
	if _, ok := blk1.(*ThinkingBlock); !ok {
		t.Fatalf("thinking: unexpected type %T", blk1)
	}

	blk2, err := r.Create("tool_result", "t2")
	if err != nil {
		t.Fatalf("tool_result: unexpected error: %v", err)
	}
	if _, ok := blk2.(*ToolResultBlock); !ok {
		t.Fatalf("tool_result: unexpected type %T", blk2)
	}

	blk3, err := r.Create("error", "t3")
	if err != nil {
		t.Fatalf("error: unexpected error: %v", err)
	}
	if _, ok := blk3.(*ErrorBlock); !ok {
		t.Fatalf("error: unexpected type %T", blk3)
	}
}

func TestNewDefaultRegistry(t *testing.T) {
	r := NewDefaultRegistry()

	// Should have all 6 built-in types registered.
	expectedTypes := []string{
		"thinking",
		"assistant_text",
		"tool_call",
		"tool_result",
		"error",
		"user_message",
	}

	for _, name := range expectedTypes {
		if !r.Has(name) {
			t.Fatalf("expected %q to be registered", name)
		}
	}

	// Create one of each type
	for _, name := range expectedTypes {
		blk, err := r.Create(name, "test-"+name)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", name, err)
		}
		if blk.ID() != "test-"+name {
			t.Fatalf("%s: expected ID %q, got %q", name, "test-"+name, blk.ID())
		}
	}
}

func TestRegistryTypes(t *testing.T) {
	r := NewRegistry()
	r.Register("thinking", func(id string) Block { return NewThinkingBlock(id) })
	r.Register("error", func(id string) Block { return NewErrorBlock(id) })

	types := r.Types()
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d", len(types))
	}
}
