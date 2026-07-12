package block

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// TypeSubAgent is the BlockType for sub-agent messages.
var TypeSubAgent BlockType = 255 // Use high value to avoid conflicts

func init() {
	// Register type name
	subAgentTypes[TypeSubAgent] = "SubAgent"
}

var subAgentTypes = make(map[BlockType]string)

// SubAgentMessage represents a single message in a sub-agent conversation.
type SubAgentMessage struct {
	Role    string    `json:"role"`    // user/assistant
	Content string    `json:"content"` // message text
	Time    time.Time `json:"time"`    // when received
}

// SubAgentBlock displays sub-agent/swarm messages from spawned agents.
// It shows the agent name, task, status, and message history.
type SubAgentBlock struct {
	BaseBlock
	agentID   string
	agentName string
	task      string
	status    string // running/completed/failed
	messages  []SubAgentMessage
}

// NewSubAgentBlock creates a new sub-agent block.
func NewSubAgentBlock(id, agentID, agentName, task string) *SubAgentBlock {
	return &SubAgentBlock{
		BaseBlock: NewBaseBlock(id, TypeSubAgent),
		agentID:   agentID,
		agentName: agentName,
		task:      task,
		status:    "running",
	}
}

// AgentID returns the agent identifier.
func (b *SubAgentBlock) AgentID() string { return b.agentID }

// AgentName returns the agent display name.
func (b *SubAgentBlock) AgentName() string { return b.agentName }

// Task returns the task description.
func (b *SubAgentBlock) Task() string { return b.task }

// Status returns the current status.
func (b *SubAgentBlock) Status() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.status
}

// SetStatus updates the agent status.
func (b *SubAgentBlock) SetStatus(status string) {
	b.mu.Lock()
	b.status = status
	b.dirty = true
	b.mu.Unlock()
}

// Messages returns a copy of the message history.
func (b *SubAgentBlock) Messages() []SubAgentMessage {
	b.mu.RLock()
	defer b.mu.RUnlock()
	result := make([]SubAgentMessage, len(b.messages))
	copy(result, b.messages)
	return result
}

// AddMessage appends a message to the history.
func (b *SubAgentBlock) AddMessage(msg SubAgentMessage) {
	b.mu.Lock()
	b.messages = append(b.messages, msg)
	b.dirty = true
	b.mu.Unlock()
}

// MessageCount returns the number of messages.
func (b *SubAgentBlock) MessageCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.messages)
}

// Measure computes the desired height.
func (b *SubAgentBlock) Measure(cs component.Constraints) component.Size {
	b.mu.RLock()
	defer b.mu.RUnlock()

	w := cs.MaxWidth
	if w <= 0 {
		w = 80
	}

	// Header: 1 line + spacing
	h := 2
	// Messages: 1 line each (truncated)
	h += len(b.messages)
	// Bottom spacing
	h++

	return component.Size{W: w, H: h}
}

// SetBounds is called by the container to set the block's position.
func (b *SubAgentBlock) SetBounds(r component.Rect) {
	b.BaseBlock.SetBounds(r)
}

// Paint renders the sub-agent block.
func (b *SubAgentBlock) Paint(buf *buffer.Buffer) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	bounds := b.BaseBlock.Bounds()
	x, y := bounds.X, bounds.Y
	w := bounds.W
	if w <= 0 {
		return
	}

	// Colors
	var statusIcon buffer.Color
	switch b.status {
	case "running":
		statusIcon = buffer.RGB(241, 250, 140) // dracula yellow
	case "completed":
		statusIcon = buffer.RGB(80, 250, 123) // dracula green
	case "failed":
		statusIcon = buffer.RGB(255, 85, 85) // dracula red
	default:
		statusIcon = buffer.RGB(139, 143, 159) // dracula comment
	}

	nameColor := buffer.RGB(189, 147, 249) // dracula purple
	mutedColor := buffer.RGB(139, 143, 159)

	// Header line: icon status agent_name: task
	icon := "\u25CF" // ●
	header := fmt.Sprintf("%s %s: %s [%s]", icon, b.agentName, b.task, b.status)
	drawX := x
	for _, r := range header {
		if drawX >= x+w || drawX >= buf.Width {
			break
		}
		color := nameColor
		// Color the icon based on status
		if drawX == x {
			color = statusIcon
		}
		buf.SetCell(drawX, y, buffer.Cell{Rune: r, Width: 1, Fg: color})
		drawX++
	}

	// Messages
	msgY := y + 2
	for i := b.firstVisibleMessage(w); i < len(b.messages) && msgY < bounds.Y+bounds.H; i++ {
		msg := b.messages[i]
		rolePrefix := "  >"
		roleColor := mutedColor
		if msg.Role == "assistant" {
			rolePrefix = "  <"
			roleColor = buffer.RGB(139, 233, 253) // dracula cyan
		}

		drawX = x + 1
		for _, r := range rolePrefix {
			if drawX >= x+w || drawX >= buf.Width {
				break
			}
			buf.SetCell(drawX, msgY, buffer.Cell{Rune: r, Width: 1, Fg: roleColor})
			drawX++
		}

		// Content (truncated)
		maxContent := w - 6
		if maxContent < 1 {
			maxContent = 1
		}
		content := msg.Content
		if len([]rune(content)) > maxContent {
			content = string([]rune(content)[:maxContent-1]) + "\u2026"
		}
		for _, r := range content {
			if drawX >= x+w || drawX >= buf.Width {
				break
			}
			buf.SetCell(drawX, msgY, buffer.Cell{Rune: r, Width: 1, Fg: buffer.RGB(248, 248, 242)})
			drawX++
		}
		msgY++
	}
}

// firstVisibleMessage returns 0 (show all from start, could be enhanced with scroll).
func (b *SubAgentBlock) firstVisibleMessage(w int) int {
	return 0
}

// SerializeState serializes the block state to JSON.
func (b *SubAgentBlock) SerializeState() (json.RawMessage, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	data := struct {
		AgentID   string            `json:"agentID"`
		AgentName string            `json:"agentName"`
		Task      string            `json:"task"`
		Status    string            `json:"status"`
		Messages  []SubAgentMessage `json:"messages"`
	}{
		AgentID:   b.agentID,
		AgentName: b.agentName,
		Task:      b.task,
		Status:    b.status,
		Messages:  b.messages,
	}
	return json.Marshal(data)
}

// DeserializeState restores block state from JSON.
func (b *SubAgentBlock) DeserializeState(data json.RawMessage) error {
	var state struct {
		AgentID   string            `json:"agentID"`
		AgentName string            `json:"agentName"`
		Task      string            `json:"task"`
		Status    string            `json:"status"`
		Messages  []SubAgentMessage `json:"messages"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}
	b.mu.Lock()
	b.agentID = state.AgentID
	b.agentName = state.AgentName
	b.task = state.Task
	b.status = state.Status
	b.messages = state.Messages
	b.dirty = true
	b.mu.Unlock()
	return nil
}
