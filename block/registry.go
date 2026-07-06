package block

import (
	"fmt"
)

// BlockFactoryFn is a factory function that creates a new Block with the given ID.
type BlockFactoryFn func(id string) Block

// Registry maps block type names to factory functions.
// It enables dynamic block creation by type name, useful for
// deserialization, plugins, or generic block creation utilities.
type Registry struct {
	factories map[string]BlockFactoryFn
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]BlockFactoryFn),
	}
}

// Register associates a type name with a factory function.
// If the type is already registered, the new factory replaces the old one.
func (r *Registry) Register(typeName string, fn BlockFactoryFn) {
	r.factories[typeName] = fn
}

// Create produces a new Block of the given type with the given ID.
// Returns an error if the type is not registered.
func (r *Registry) Create(typeName string, id string) (Block, error) {
	fn, ok := r.factories[typeName]
	if !ok {
		return nil, fmt.Errorf("block type %q is not registered", typeName)
	}
	return fn(id), nil
}

// Has reports whether the given type name is registered.
func (r *Registry) Has(typeName string) bool {
	_, ok := r.factories[typeName]
	return ok
}

// Types returns all registered type names.
func (r *Registry) Types() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// TypeNameForBlock finds the registered type name for a given block instance
// by creating a probe from each factory and comparing BlockType values.
// Returns "" if no match is found. This enables serialization of custom blocks
// whose BlockType falls outside the built-in constants.
func (r *Registry) TypeNameForBlock(b Block) string {
	bt := b.Type()
	for name, fn := range r.factories {
		probe := fn("__probe__")
		if probe.Type() == bt {
			return name
		}
	}
	return ""
}

// NewDefaultRegistry creates a Registry pre-populated with all built-in
// block types.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register("thinking", func(id string) Block { return NewThinkingBlock(id) })
	r.Register("assistant_text", func(id string) Block { return NewAssistantTextBlock(id) })
	r.Register("tool_call", func(id string) Block { return NewToolCallBlock(id, "", "") })
	r.Register("tool_result", func(id string) Block { return NewToolResultBlock(id) })
	r.Register("error", func(id string) Block { return NewErrorBlock(id) })
	r.Register("user_message", func(id string) Block { return NewUserMessageBlock(id, "") })
	r.Register("image", func(id string) Block { return NewImageBlock(id, "", nil) })
	return r
}
