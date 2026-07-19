package component

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// MessageRole identifies who authored a message.
type MessageRole uint8

const (
	RoleUser MessageRole = iota
	RoleAssistant
	RoleSystem
	RoleTool
)

// String returns a human-readable role name.
func (r MessageRole) String() string {
	switch r {
	case RoleUser:
		return "You"
	case RoleAssistant:
		return "Assistant"
	case RoleSystem:
		return "System"
	case RoleTool:
		return "Tool"
	}
	return "Unknown"
}

// defaultAvatar returns a visual icon for each role.
func defaultAvatar(r MessageRole) string {
	switch r {
	case RoleUser:
		return "👤"
	case RoleAssistant:
		return "🤖"
	case RoleSystem:
		return "⚙"
	case RoleTool:
		return "🔧"
	}
	return "?"
}

// MessageBubble renders a chat message with role-based styling.
// User messages are right-aligned with accent background; assistant messages
// are left-aligned with muted background; system messages are centered with border.
//
// This is the unifying wrapper that connects specialized blocks (ToolCallView,
// CitationsBlock, etc.) into a coherent conversation flow.
//
// Thread-safe.
type MessageBubble struct {
	BaseComponent
	mu sync.Mutex

	role      MessageRole
	content   string
	timestamp time.Time
	modelName string
	streaming bool
	hasError  bool
	avatar    string
	cursorOn  bool // blink state for streaming cursor
}

// NewMessageBubble creates a message bubble with the given role and content.
func NewMessageBubble(role MessageRole, content string) *MessageBubble {
	return &MessageBubble{
		BaseComponent: BaseComponent{id: GenerateID("msg")},
		role:          role,
		content:       content,
		timestamp:     time.Now(),
		avatar:        defaultAvatar(role),
	}
}

// Role returns the message role.
func (m *MessageBubble) Role() MessageRole {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.role
}

// Content returns the message text.
func (m *MessageBubble) Content() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.content
}

// SetContent replaces the message text.
func (m *MessageBubble) SetContent(s string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.content = s
}

// SetStreaming toggles streaming mode (shows blinking cursor).
func (m *MessageBubble) SetStreaming(v bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streaming = v
}

// Streaming returns whether streaming mode is active.
func (m *MessageBubble) Streaming() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.streaming
}

// SetError marks the message as an error.
func (m *MessageBubble) SetError(v bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hasError = v
}

// HasError returns whether the message is an error.
func (m *MessageBubble) HasError() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.hasError
}

// SetModel sets the model name shown in the header (assistant messages).
func (m *MessageBubble) SetModel(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modelName = name
}

// Model returns the model name.
func (m *MessageBubble) Model() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.modelName
}

// SetAvatar sets a custom avatar icon.
func (m *MessageBubble) SetAvatar(s string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.avatar = s
}

// Avatar returns the current avatar.
func (m *MessageBubble) Avatar() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.avatar
}

// Timestamp returns the message creation time.
func (m *MessageBubble) Timestamp() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.timestamp
}

// SetTimestamp sets the message timestamp.
func (m *MessageBubble) SetTimestamp(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timestamp = t
}

// ToggleCursor flips the streaming cursor blink state.
func (m *MessageBubble) ToggleCursor() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cursorOn = !m.cursorOn
}

// Measure returns the desired size based on content and width.
func (m *MessageBubble) Measure(cs Constraints) Size {
	m.mu.Lock()
	defer m.mu.Unlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}

	// Header (1 line) + content lines
	h := 1
	if m.content != "" {
		contentW := m.contentWidthLocked(maxW)
		lines := wrapLines(m.content, contentW)
		h += len(lines)
	}
	if m.streaming {
		// cursor takes part of the last content line, no extra line
	}
	if m.hasError {
		h++ // error footer
	}
	if h < 1 {
		h = 1
	}
	return Size{W: maxW, H: h}
}

// contentWidthLocked returns the usable width for content text.
func (m *MessageBubble) contentWidthLocked(maxW int) int {
	// Reserve padding: 2 chars on each side for non-system, 1 for system
	if m.role == RoleSystem {
		if maxW < 2 {
			return 1
		}
		return maxW - 2
	}
	if maxW < 4 {
		return maxW
	}
	return maxW - 4
}

// Paint renders the message bubble.
func (m *MessageBubble) Paint(buf *buffer.Buffer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	bounds := m.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	switch m.role {
	case RoleSystem:
		m.paintSystem(buf, bounds)
	default:
		m.paintStandard(buf, bounds)
	}
}

// paintStandard renders user/assistant/tool messages with header + content.
func (m *MessageBubble) paintStandard(buf *buffer.Buffer, bounds Rect) {
	th := theme.Get()
	w := bounds.W
	y := bounds.Y

	// --- Header line ---
	headerStyle := buffer.Style{Fg: th.Muted}
	if m.hasError {
		headerStyle = buffer.Style{Fg: th.Error}
	}

	header := m.avatar + " " + m.role.String()
	if m.modelName != "" {
		header += " · " + m.modelName
	}
	header += "  " + m.timestamp.Format("15:04")
	if m.streaming {
		header += " …"
	}

	// Right-align for user
	if m.role == RoleUser {
		hr := []rune(header)
		if len(hr) < w {
			spaces := w - len(hr)
			buf.DrawText(bounds.X+spaces, y, header, headerStyle)
		} else {
			r := hr
			if len(r) > w {
				r = r[:w]
			}
			buf.DrawText(bounds.X, y, string(r), headerStyle)
		}
	} else {
		r := []rune(header)
		if len(r) > w {
			r = r[:w]
		}
		buf.DrawText(bounds.X, y, string(r), headerStyle)
	}
	y++

	// --- Content ---
	contentStyle := buffer.Style{Fg: th.Fg}
	if m.hasError {
		contentStyle = buffer.Style{Fg: th.Error}
	}

	if m.content == "" && m.streaming {
		// Show just the cursor
		cursor := " ▊"
		if !m.cursorOn {
			cursor = "  "
		}
		buf.DrawText(bounds.X+1, y, cursor, contentStyle)
		return
	}

	contentW := m.contentWidthLocked(w)
	padX := bounds.X + 2 // left padding for non-system

	// Draw content using DrawTextClamped to avoid []rune allocations
	for _, para := range strings.Split(m.content, "\n") {
		if y >= bounds.Y+bounds.H {
			return
		}
		if para == "" {
			y++
			continue
		}
		// Word-wrap within the paragraph
		words := strings.Fields(para)
		if len(words) == 0 {
			y++
			continue
		}
		current := words[0]
		curLen := utf8.RuneCountInString(current)
		for _, wd := range words[1:] {
			wdLen := utf8.RuneCountInString(wd)
			if curLen+1+wdLen <= contentW {
				current += " " + wd
				curLen += 1 + wdLen
			} else {
				m.drawLineLocked(buf, padX, y, contentW, current, contentStyle)
				y++
				if y >= bounds.Y+bounds.H {
					return
				}
				current = wd
			}
		}
		// Last line of this paragraph (append cursor if streaming)
		if m.streaming {
			current += "▊"
		}
		m.drawLineLocked(buf, padX, y, contentW, current, contentStyle)
		y++
	}

	// --- Error footer ---
	if m.hasError && y < bounds.Y+bounds.H {
		buf.DrawText(padX, y, "⚠ Error", buffer.Style{Fg: th.Error})
	}
}

// paintSystem renders a centered system message with visual separation.
func (m *MessageBubble) paintSystem(buf *buffer.Buffer, bounds Rect) {
	th := theme.Get()
	w := bounds.W
	y := bounds.Y

	sysStyle := buffer.Style{Fg: th.Muted}

	contentW := w - 2
	if contentW < 1 {
		contentW = 1
	}

	lines := wrapLines(m.content, contentW)
	for _, line := range lines {
		if y >= bounds.Y+bounds.H {
			break
		}
		// Center-align
		lr := []rune(line)
		if len(lr) > contentW {
			lr = lr[:contentW]
		}
		spaces := (w - len(lr)) / 2
		buf.DrawText(bounds.X+spaces, y, string(lr), sysStyle)
		y++
	}
}

// drawLineLocked draws a single content line, right-aligned for user messages.
func (m *MessageBubble) drawLineLocked(buf *buffer.Buffer, x, y, maxW int, text string, style buffer.Style) {
	r := []rune(text)
	if len(r) > maxW {
		text = string(r[:maxW])
		r = []rune(text)
	}
	if m.role == RoleUser {
		if len(r) < maxW {
			x += maxW - len(r)
		}
	}
	buf.DrawText(x, y, text, style)
}

// wrapLines wraps text to fit within the given width, returning a slice of lines.
func wrapLines(text string, width int) []string {
	if width < 1 {
		width = 1
	}
	var result []string
	for _, para := range strings.Split(text, "\n") {
		if para == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(para)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		current := words[0]
		for _, w := range words[1:] {
			if len([]rune(current))+1+len([]rune(w)) <= width {
				current += " " + w
			} else {
				result = append(result, current)
				current = w
			}
		}
		result = append(result, current)
	}
	if len(result) == 0 {
		result = []string{""}
	}
	return result
}

// MessageRoleFromString parses a role name string into a MessageRole.
func MessageRoleFromString(s string) MessageRole {
	switch strings.ToLower(s) {
	case "user", "human":
		return RoleUser
	case "assistant", "ai", "bot":
		return RoleAssistant
	case "system":
		return RoleSystem
	case "tool":
		return RoleTool
	}
	return RoleAssistant
}

// String returns a short label for the bubble (for debugging).
func (m *MessageBubble) String() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprintf("MessageBubble{%s: %q}", m.role, truncateStr(m.content, 30))
}
