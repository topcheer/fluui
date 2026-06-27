package block

import (
	"encoding/json"
	"fmt"
)

// SerializedBlock is the JSON representation of a block for persistence.
type SerializedBlock struct {
	Type      string            `json:"type"`            // registry type name (e.g. "thinking")
	ID        string            `json:"id"`              // unique block ID
	State     string            `json:"state"`           // lifecycle state ("streaming","complete","error")
	StateData json.RawMessage   `json:"state_data"`      // block-type-specific serialized data
	Children  []SerializedBlock `json:"children,omitempty"` // reserved for nested blocks
}

// Serializer is implemented by blocks that can serialize their state to JSON.
type Serializer interface {
	SerializeState() (json.RawMessage, error)
}

// TypeNameProvider is optionally implemented by blocks that have a custom
// registry type name. This is used by SaveContainer to serialize the correct
// type name for plugin-registered blocks whose BlockType value is not a
// built-in type. Built-in blocks don't need this — their Type().String()
// is used automatically.
type TypeNameProvider interface {
	TypeName() string
}

// Deserializer is implemented by blocks that can restore state from JSON.
type Deserializer interface {
	DeserializeState(data json.RawMessage) error
}

// SerializedContainer is the JSON representation of a BlockContainer.
type SerializedContainer struct {
	Version int              `json:"version"` // format version (currently 1)
	Blocks  []SerializedBlock `json:"blocks"`
}

// SaveContainer serializes all blocks in a container to JSON bytes.
// Blocks that implement Serializer will have their state captured.
// Blocks that don't implement Serializer are saved with null state_data.
func SaveContainer(c *BlockContainer, r *Registry) ([]byte, error) {
	blocks := c.Blocks()
	serialized := make([]SerializedBlock, 0, len(blocks))

	for _, b := range blocks {
		typeStr := b.Type().String()
		// For custom block types (e.g. from plugins), Type().String() returns "unknown".
		// First, check if the block explicitly declares its type name.
		if tnp, ok := b.(TypeNameProvider); ok {
			typeStr = tnp.TypeName()
		}
		// Fallback: try the registry reverse lookup via probe comparison.
		if r != nil && !r.Has(typeStr) {
			if name := r.TypeNameForBlock(b); name != "" {
				typeStr = name
			}
		}

		sb := SerializedBlock{
			Type:  typeStr,
			ID:    b.ID(),
			State: b.State().String(),
		}

		if s, ok := b.(Serializer); ok {
			data, err := s.SerializeState()
			if err != nil {
				return nil, fmt.Errorf("serialize block %q: %w", b.ID(), err)
			}
			sb.StateData = data
		}

		serialized = append(serialized, sb)
	}

	container := SerializedContainer{
		Version: 1,
		Blocks:  serialized,
	}

	return json.MarshalIndent(container, "", "  ")
}

// LoadContainer deserializes JSON bytes into a new BlockContainer.
// Each block is created via the Registry factory, then its state is
// restored via the Deserializer interface if the block implements it.
func LoadContainer(data []byte, r *Registry) (*BlockContainer, error) {
	var sc SerializedContainer
	if err := json.Unmarshal(data, &sc); err != nil {
		return nil, fmt.Errorf("unmarshal container: %w", err)
	}

	c := NewBlockContainer()

	for _, sb := range sc.Blocks {
		b, err := r.Create(sb.Type, sb.ID)
		if err != nil {
			return nil, fmt.Errorf("create block type %q id %q: %w", sb.Type, sb.ID, err)
		}

		// Restore block-specific state first (before lifecycle state),
		// because some blocks like ToolResultBlock.Complete() inspect content.
		if d, ok := b.(Deserializer); ok && len(sb.StateData) > 0 {
			if err := d.DeserializeState(sb.StateData); err != nil {
				return nil, fmt.Errorf("deserialize block %q: %w", sb.ID, err)
			}
		}

		// Restore lifecycle state. Use SetState instead of Fail/Complete
		// so we don't overwrite the deserialized content.
		switch sb.State {
		case "complete":
			b.Complete()
		case "error":
			b.SetState(BlockError)
		}

		c.AddBlock(b)
	}

	return c, nil
}
