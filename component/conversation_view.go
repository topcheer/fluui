package component

import (
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/theme"
)

// ConversationMessage represents a single entry in a conversation.
// It can be a text message (via Role/Content) or embed specialized
// components (ToolCallView, CitationsBlock).
type ConversationMessage struct {
	Role       MessageRole
	Content    string
	ModelName  string
	Timestamp  time.Time
	Streaming  bool
	ToolCall   *ToolCallView   // optional embedded tool call
	Citations  *CitationsBlock  // optional embedded citations
	Error      bool
}

// ConversationView is a scrollable chat history that orchestrates
// MessageBubbles, ToolCallViews, CitationsBlocks, and other AI-native
// components into a unified conversation flow.
//
// Features:
//   - Append messages with AddMessage / AddUserMessage / AddAssistantMessage
//   - Embedded tool calls and citations
//   - Auto-scroll to bottom when streaming
//   - Manual scroll up/down/to-bottom
//   - Thread-safe
type ConversationView struct {
	BaseComponent
	mu sync.Mutex

	messages   []ConversationMessage
	streaming  bool // global streaming flag (auto-scroll)
	scrollOff  int  // vertical scroll offset (0 = top)
	contentH   int  // total computed content height
	autoScroll bool // whether to stick to bottom
}

// NewConversationView creates an empty conversation view with auto-scroll enabled.
func NewConversationView() *ConversationView {
	return &ConversationView{
		BaseComponent: BaseComponent{id: GenerateID("conversation")},
		autoScroll:    true,
	}
}

// AddMessage appends a message to the conversation.
func (cv *ConversationView) AddMessage(msg ConversationMessage) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	cv.messages = append(cv.messages, msg)
}

// AddUserMessage is a convenience for appending a user text message.
func (cv *ConversationView) AddUserMessage(content string) {
	cv.AddMessage(ConversationMessage{
		Role:      RoleUser,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddAssistantMessage is a convenience for appending an assistant text message.
func (cv *ConversationView) AddAssistantMessage(content, model string) {
	cv.AddMessage(ConversationMessage{
		Role:      RoleAssistant,
		Content:   content,
		ModelName: model,
		Timestamp: time.Now(),
	})
}

// AddSystemMessage is a convenience for appending a system message.
func (cv *ConversationView) AddSystemMessage(content string) {
	cv.AddMessage(ConversationMessage{
		Role:      RoleSystem,
		Content:   content,
		Timestamp: time.Now(),
	})
}

// AddToolCall inserts a tool call into the conversation.
func (cv *ConversationView) AddToolCall(tc *ToolCallView) {
	cv.AddMessage(ConversationMessage{
		Role:     RoleTool,
		ToolCall: tc,
	})
}

// AddCitations appends a citations block as a conversation entry.
func (cv *ConversationView) AddCitations(cb *CitationsBlock) {
	// Citations are rendered as an assistant message with embedded block
	cv.AddMessage(ConversationMessage{
		Role:      RoleAssistant,
		Citations: cb,
	})
}

// Clear removes all messages.
func (cv *ConversationView) Clear() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.messages = nil
	cv.scrollOff = 0
	cv.contentH = 0
}

// MessageCount returns the number of messages.
func (cv *ConversationView) MessageCount() int {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	return len(cv.messages)
}

// Messages returns a copy of the messages slice.
func (cv *ConversationView) Messages() []ConversationMessage {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	out := make([]ConversationMessage, len(cv.messages))
	copy(out, cv.messages)
	return out
}

// SetStreaming enables/disables streaming mode (auto-scroll).
func (cv *ConversationView) SetStreaming(v bool) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.streaming = v
	if v {
		cv.autoScroll = true
	}
}

// IsStreaming returns whether streaming mode is active.
func (cv *ConversationView) IsStreaming() bool {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	return cv.streaming
}

// ScrollUp scrolls up by n lines.
func (cv *ConversationView) ScrollUp(n int) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.scrollOff -= n
	if cv.scrollOff < 0 {
		cv.scrollOff = 0
	}
	cv.autoScroll = false
}

// ScrollDown scrolls down by n lines.
func (cv *ConversationView) ScrollDown(n int) {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.scrollOff += n
	maxOff := cv.maxScrollLocked()
	if cv.scrollOff > maxOff {
		cv.scrollOff = maxOff
	}
	if cv.scrollOff >= maxOff {
		cv.autoScroll = true
	}
}

// ScrollToBottom scrolls to the most recent content.
func (cv *ConversationView) ScrollToBottom() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.scrollOff = cv.maxScrollLocked()
	cv.autoScroll = true
}

// ScrollToTop scrolls to the beginning.
func (cv *ConversationView) ScrollToTop() {
	cv.mu.Lock()
	defer cv.mu.Unlock()
	cv.scrollOff = 0
	cv.autoScroll = false
}

// HandleKey processes keyboard input for scrolling.
func (cv *ConversationView) HandleKey(ev *term.KeyEvent) bool {
	if ev == nil {
		return false
	}
	switch ev.Key {
	case term.KeyUp:
		cv.ScrollUp(1)
		return true
	case term.KeyDown:
		cv.ScrollDown(1)
		return true
	case term.KeyPageUp:
		cv.ScrollUp(10)
		return true
	case term.KeyPageDown:
		cv.ScrollDown(10)
		return true
	case term.KeyHome:
		cv.ScrollToTop()
		return true
	case term.KeyEnd:
		cv.ScrollToBottom()
		return true
	}
	return false
}

// maxScrollLocked returns the maximum scroll offset (caller must hold lock).
func (cv *ConversationView) maxScrollLocked() int {
	bounds := cv.bounds
	if bounds.H <= 0 {
		return 0
	}
	maxOff := cv.contentH - bounds.H
	if maxOff < 0 {
		maxOff = 0
	}
	return maxOff
}

// Measure computes the desired size. Always fills available space.
func (cv *ConversationView) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 24
	}
	return Size{W: maxW, H: maxH}
}

// Paint renders the conversation view.
func (cv *ConversationView) Paint(buf *buffer.Buffer) {
	cv.mu.Lock()
	defer cv.mu.Unlock()

	bounds := cv.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	if len(cv.messages) == 0 {
		cv.paintEmpty(buf, bounds)
		return
	}

	// Compute total content height
	cv.contentH = cv.computeContentHeightLocked(bounds.W)

	// Determine scroll offset
	if cv.autoScroll {
		cv.scrollOff = cv.maxScrollLocked()
	}
	// Clamp scroll
	maxOff := cv.maxScrollLocked()
	if cv.scrollOff > maxOff {
		cv.scrollOff = maxOff
	}

	// Render messages into a virtual buffer and blit the visible portion
	cv.renderMessages(buf, bounds)
}

// paintEmpty renders a placeholder when there are no messages.
func (cv *ConversationView) paintEmpty(buf *buffer.Buffer, bounds Rect) {
	muted := buffer.Style{Fg: theme.Get().Muted}
	hint := "No messages yet. Start a conversation."
	r := []rune(hint)
	if len(r) > bounds.W {
		r = r[:bounds.W]
	}
	spaces := (bounds.W - len(r)) / 2
	if spaces < 0 {
		spaces = 0
	}
	midY := bounds.Y + bounds.H/2
	buf.DrawText(bounds.X+spaces, midY, string(r), muted)
}

// computeContentHeightLocked calculates total height of all messages at given width.
func (cv *ConversationView) computeContentHeightLocked(width int) int {
	total := 0
	cs := Constraints{MaxWidth: width}
	for _, msg := range cv.messages {
		total += cv.measureMessageLocked(msg, cs)
		// 1 blank line between messages
		total++
	}
	if total > 0 {
		total-- // no trailing blank
	}
	return total
}

// measureMessageLocked returns the height needed for a single message.
func (cv *ConversationView) measureMessageLocked(msg ConversationMessage, cs Constraints) int {
	if msg.ToolCall != nil {
		return msg.ToolCall.Measure(cs).H
	}
	if msg.Citations != nil {
		return msg.Citations.Measure(cs).H
	}
	mb := NewMessageBubble(msg.Role, msg.Content)
	if msg.ModelName != "" {
		mb.SetModel(msg.ModelName)
	}
	mb.SetStreaming(msg.Streaming)
	mb.SetError(msg.Error)
	return mb.Measure(cs).H
}

// renderMessages paints messages at the correct scroll offset.
func (cv *ConversationView) renderMessages(buf *buffer.Buffer, bounds Rect) {
	width := bounds.W
	cs := Constraints{MaxWidth: width}
	y := 0 // virtual y in content space

	for _, msg := range cv.messages {
		h := cv.measureMessageLocked(msg, cs)

		// Determine if any part of this message is visible
		visTop := cv.scrollOff
		visBot := cv.scrollOff + bounds.H

		if y+h > visTop && y < visBot {
			// At least partially visible — render it
			// Calculate the screen Y for this message
			screenY := bounds.Y + (y - cv.scrollOff)
			if screenY < bounds.Y {
				screenY = bounds.Y
			}

			// Calculate how many lines are clipped from the top
			clipTop := 0
			if y < visTop {
				clipTop = visTop - y
			}

			// Available height for this message on screen
			availH := bounds.H - (screenY - bounds.Y)
			if availH > h-clipTop {
				availH = h - clipTop
			}
			if availH < 0 {
				availH = 0
			}

			if availH > 0 {
				msgBounds := Rect{X: bounds.X, Y: screenY, W: width, H: availH}
				cv.paintMessage(buf, msgBounds, msg, cs, clipTop)
			}
		}

		y += h + 1 // +1 for spacing between messages
	}
}

// paintMessage renders a single message into the buffer at the given bounds.
func (cv *ConversationView) paintMessage(buf *buffer.Buffer, bounds Rect, msg ConversationMessage, cs Constraints, clipTop int) {
	if msg.ToolCall != nil {
		// Render tool call directly
		toolBounds := Rect{X: bounds.X + 1, Y: bounds.Y, W: bounds.W - 2, H: bounds.H}
		msg.ToolCall.SetBounds(toolBounds)
		msg.ToolCall.Paint(buf)
		return
	}
	if msg.Citations != nil {
		citBounds := Rect{X: bounds.X + 1, Y: bounds.Y, W: bounds.W - 2, H: bounds.H}
		msg.Citations.SetBounds(citBounds)
		msg.Citations.Paint(buf)
		return
	}
	// Standard message bubble
	mb := NewMessageBubble(msg.Role, msg.Content)
	if msg.ModelName != "" {
		mb.SetModel(msg.ModelName)
	}
	mb.SetStreaming(msg.Streaming)
	mb.SetError(msg.Error)
	mb.SetTimestamp(msg.Timestamp)
	mb.SetBounds(bounds)
	mb.Paint(buf)
}
