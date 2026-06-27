package block

import (
	"errors"
	"fmt"
	"sync"
)

// StreamDelta represents one increment of AI streaming output.
type StreamDelta struct {
	Type     string // "thinking", "text", "tool_call", "tool_result", "error"
	Content  string // delta text
	ToolName string // only for tool_call
	ToolArgs string // only for tool_call
}

// BlockFactory creates blocks with auto-incrementing IDs.
type BlockFactory struct {
	mu        sync.Mutex
	idCounter int
}

// NewBlockFactory creates a BlockFactory.
func NewBlockFactory() *BlockFactory {
	return &BlockFactory{}
}

// NextID generates a unique ID with the given prefix.
func (f *BlockFactory) NextID(prefix string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.idCounter++
	return fmt.Sprintf("%s-%d", prefix, f.idCounter)
}

// CreateThinkingBlock creates a new ThinkingBlock with a unique ID.
func (f *BlockFactory) CreateThinkingBlock() *ThinkingBlock {
	return NewThinkingBlock(f.NextID("thinking"))
}

// CreateToolCallBlock creates a new ToolCallBlock with a unique ID.
func (f *BlockFactory) CreateToolCallBlock(toolName, rawArgs string) *ToolCallBlock {
	return NewToolCallBlock(f.NextID("tool-call"), toolName, rawArgs)
}

// CreateToolResultBlock creates a new ToolResultBlock with a unique ID.
func (f *BlockFactory) CreateToolResultBlock() *ToolResultBlock {
	return NewToolResultBlock(f.NextID("tool-result"))
}

// CreateAssistantTextBlock creates a new AssistantTextBlock with a unique ID.
func (f *BlockFactory) CreateAssistantTextBlock() *AssistantTextBlock {
	return NewAssistantTextBlock(f.NextID("text"))
}

// CreateUserMessageBlock creates a new UserMessageBlock with a unique ID.
func (f *BlockFactory) CreateUserMessageBlock(content string) *UserMessageBlock {
	return NewUserMessageBlock(f.NextID("user"), content)
}

// StreamDispatcher routes streaming deltas to the correct block type.
// It maintains at most one active block per streamable type (thinking, text,
// tool_result). Tool calls always create a new block because each call is
// distinct.
type StreamDispatcher struct {
	mu         sync.Mutex
	container  *BlockContainer
	factory    *BlockFactory
	current    map[BlockType]Block // active streaming blocks by type
	currentErr Block               // active error block (if any)
}

// NewStreamDispatcher creates a dispatcher bound to the given container.
func NewStreamDispatcher(container *BlockContainer) *StreamDispatcher {
	return &StreamDispatcher{
		container: container,
		factory:   NewBlockFactory(),
		current:   make(map[BlockType]Block),
	}
}

// Dispatch processes a single streaming delta, creating or updating blocks
// as needed.
func (d *StreamDispatcher) Dispatch(delta StreamDelta) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch delta.Type {
	case "thinking":
		blk := d.current[TypeThinking]
		if blk == nil {
			blk = d.factory.CreateThinkingBlock()
			d.container.AddBlock(blk)
			d.current[TypeThinking] = blk
		}
		blk.AppendDelta(delta.Content)

	case "text":
		blk := d.current[TypeAssistantText]
		if blk == nil {
			blk = d.factory.CreateAssistantTextBlock()
			d.container.AddBlock(blk)
			d.current[TypeAssistantText] = blk
		}
		blk.AppendDelta(delta.Content)

	case "tool_call":
		// Each tool call is a distinct block.
		blk := d.factory.CreateToolCallBlock(delta.ToolName, delta.ToolArgs)
		d.container.AddBlock(blk)
		// Track as current tool_call so we can complete it on flush.
		d.current[TypeToolCall] = blk

	case "tool_result":
		blk := d.current[TypeToolResult]
		if blk == nil {
			blk = d.factory.CreateToolResultBlock()
			d.container.AddBlock(blk)
			d.current[TypeToolResult] = blk
		}
		blk.AppendDelta(delta.Content)

	case "error":
		if d.currentErr != nil {
			d.currentErr.AppendDelta(delta.Content)
		} else {
			// Fail the current active block if one exists.
			failed := false
			for _, b := range d.current {
				b.Fail(errors.New(delta.Content))
				failed = true
			}
			_ = failed
		}

	default:
		return fmt.Errorf("unknown delta type: %q", delta.Type)
	}

	return nil
}

// Flush completes all currently active blocks and clears the active map.
func (d *StreamDispatcher) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, blk := range d.current {
		if blk.State() == BlockStreaming || blk.State() == BlockPending {
			blk.Complete()
		}
	}
	d.current = make(map[BlockType]Block)
	d.currentErr = nil
}

// ActiveBlocks returns the currently streaming blocks.
func (d *StreamDispatcher) ActiveBlocks() []Block {
	d.mu.Lock()
	defer d.mu.Unlock()

	blocks := make([]Block, 0, len(d.current))
	for _, b := range d.current {
		blocks = append(blocks, b)
	}
	return blocks
}

// Factory returns the underlying BlockFactory.
func (d *StreamDispatcher) Factory() *BlockFactory { return d.factory }
