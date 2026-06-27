// Package block implements the AI content block system.
//
// Blocks are the primary unit of content in a Fluui AI chat interface.
// Each block represents one semantic unit of AI output: a thinking process,
// a tool call, a tool result, or assistant text. Blocks are Components
// (they can Measure/SetBounds/Paint) with additional lifecycle methods
// for streaming updates.
package block

import (
	"fmt"
	"sync"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// BlockState represents the current lifecycle state of a block.
type BlockState uint8

const (
	BlockPending   BlockState = iota // Waiting for data (spinner)
	BlockStreaming                   // Currently receiving data
	BlockComplete                    // Data is finalized
	BlockError                       // An error occurred
)

// String returns a human-readable state name.
func (s BlockState) String() string {
	switch s {
	case BlockPending:
		return "pending"
	case BlockStreaming:
		return "streaming"
	case BlockComplete:
		return "complete"
	case BlockError:
		return "error"
	}
	return "unknown"
}

// BlockType identifies what kind of content a block contains.
// Note: We use distinct names from BlockState to avoid const conflicts.
type BlockType uint8

const (
	TypeThinking     BlockType = iota // AI thinking process
	TypeAssistantText                 // Streaming markdown text
	TypeToolCall                      // Tool invocation
	TypeToolResult                    // Tool output
	TypeCode                          // Standalone code block
	TypeError                         // Error message
	TypeUserMessage                   // User input echo
	TypeWorkflow                      // Agent workflow visualization
)

// String returns a human-readable type name.
func (t BlockType) String() string {
	switch t {
	case TypeThinking:
		return "thinking"
	case TypeAssistantText:
		return "assistant_text"
	case TypeToolCall:
		return "tool_call"
	case TypeToolResult:
		return "tool_result"
	case TypeCode:
		return "code"
	case TypeError:
		return "error"
	case TypeUserMessage:
		return "user_message"
	case TypeWorkflow:
		return "workflow"
	}
	return "unknown"
}

// Block is the interface for all AI content blocks. It extends
// component.Component with lifecycle and streaming methods.
type Block interface {
	component.Component

	// Identity
	ID() string
	Type() BlockType

	// Lifecycle
	State() BlockState
	SetState(BlockState)

	// Streaming (called from the event loop goroutine)
	AppendDelta(delta string)
	Complete()
	Fail(err error)

	// Metadata
	CreatedAt() time.Time
	Duration() time.Duration

	// Dirty flag for incremental rendering
	IsDirty() bool
	ClearDirty()
}

// BaseBlock provides common fields and methods for all Block implementations.
// Embed this in concrete block types to satisfy most of the Block interface.
type BaseBlock struct {
	component.BaseComponent

	id        string
	blockType BlockType
	state     BlockState
	startedAt time.Time
	endedAt   time.Time
	dirty     bool
	errMsg    string

	// mu protects content access in concrete blocks.
	// Pointer to avoid lock-copy when BaseBlock is embedded by value.
	mu *sync.RWMutex
}

// NewBaseBlock creates a BaseBlock with the given ID and type, in BlockStreaming state.
func NewBaseBlock(id string, bt BlockType) BaseBlock {
	return BaseBlock{
		id:        id,
		blockType: bt,
		state:     BlockStreaming,
		startedAt: time.Now(),
		dirty:     true,
		mu:        &sync.RWMutex{},
	}
}

func (b *BaseBlock) ID() string         { return b.id }
func (b *BaseBlock) Type() BlockType     { return b.blockType }
func (b *BaseBlock) State() BlockState   { return b.state }
// InitMu ensures the mutex is initialized.
// Call after manual construction (e.g. in tests).
func (b *BaseBlock) InitMu() {
	if b.mu == nil {
		b.mu = &sync.RWMutex{}
	}
}

func (b *BaseBlock) SetState(s BlockState) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = s
	b.dirty = true
}

func (b *BaseBlock) AppendDelta(_ string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dirty = true
}

func (b *BaseBlock) Complete() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = BlockComplete
	b.endedAt = time.Now()
	b.dirty = true
}

func (b *BaseBlock) Fail(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.state = BlockError
	b.endedAt = time.Now()
	b.errMsg = err.Error()
	b.dirty = true
}

func (b *BaseBlock) CreatedAt() time.Time { return b.startedAt }

func (b *BaseBlock) Duration() time.Duration {
	if b.endedAt.IsZero() {
		return time.Since(b.startedAt)
	}
	return b.endedAt.Sub(b.startedAt)
}

func (b *BaseBlock) IsDirty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.dirty
}
func (b *BaseBlock) ClearDirty() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dirty = false
}
func (b *BaseBlock) Error() string { return b.errMsg }
func (b *BaseBlock) MarkDirty() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dirty = true
}
// markDirtyLocked sets dirty=true without locking.
// Caller must already hold b.mu.
func (b *BaseBlock) markDirtyLocked() {
	b.dirty = true
}

// Paint is a no-op default; concrete blocks override this.
func (b *BaseBlock) Paint(_ *buffer.Buffer) {}

// String returns a debug representation.
func (b *BaseBlock) String() string {
	return fmt.Sprintf("Block{id=%s, type=%s, state=%s, dirty=%v}", b.id, b.blockType, b.state, b.dirty)
}
