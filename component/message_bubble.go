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
		if utf8.RuneCountInString(header) < w {
			spaces := w - utf8.RuneCountInString(header)
			buf.DrawText(bounds.X+spaces, y, header, headerStyle)
		} else {
			buf.DrawText(bounds.X, y, truncateStr(header, w), headerStyle)
		}
	} else {
		if utf8.RuneCountInString(header) > w {
			header = truncateStr(header, w)
		}
		buf.DrawText(bounds.X, y, header, headerStyle)
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

	// Draw content using inline split to avoid strings.Split allocation
	remaining := m.content
	for len(remaining) > 0 {
		if y >= bounds.Y+bounds.H {
			return
		}
		// Find next newline
		nlIdx := strings.IndexByte(remaining, '\n')
		var para string
		if nlIdx < 0 {
			para = remaining
			remaining = ""
		} else {
			para = remaining[:nlIdx]
			remaining = remaining[nlIdx+1:]
		}
		if para == "" {
			y++
			continue
		}
		// Fast path: single word or fits on one line — skip strings.Fields
		if utf8.RuneCountInString(para) <= contentW {
			if m.streaming && len(remaining) == 0 {
				para += "▊"
			}
			m.drawLineLocked(buf, padX, y, contentW, para, contentStyle)
			y++
			continue
		}
		// Word-wrap within the paragraph using zero-alloc scanner
		curLen := 0
		current := ""
		firstWord := true
		wordStart := -1
		for i := 0; i <= len(para); i++ {
			isEnd := i == len(para)
			isSpace := !isEnd && (para[i] == ' ' || para[i] == '\t')
			if isEnd || isSpace {
				if wordStart >= 0 {
					word := para[wordStart:i]
					wdLen := utf8.RuneCountInString(word)
					if firstWord {
						current = word
						curLen = wdLen
						firstWord = false
					} else if curLen+1+wdLen <= contentW {
						current += " " + word
						curLen += 1 + wdLen
					} else {
						m.drawLineLocked(buf, padX, y, contentW, current, contentStyle)
						y++
						if y >= bounds.Y+bounds.H {
							return
						}
						current = word
						curLen = wdLen
					}
					wordStart = -1
				}
			} else {
				if wordStart < 0 {
					wordStart = i
				}
			}
		}
		if firstWord {
			y++
			continue
		}
		if m.streaming && len(remaining) == 0 {
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
		lr := line
		if utf8.RuneCountInString(lr) > contentW {
			lr = truncateStr(lr, contentW)
		}
		lineLen := utf8.RuneCountInString(lr)
		spaces := (w - lineLen) / 2
		buf.DrawText(bounds.X+spaces, y, lr, sysStyle)
		y++
	}
}

// drawLineLocked draws a single content line, right-aligned for user messages.
func (m *MessageBubble) drawLineLocked(buf *buffer.Buffer, x, y, maxW int, text string, style buffer.Style) {
	if utf8.RuneCountInString(text) > maxW {
		text = truncateStr(text, maxW)
	}
	if m.role == RoleUser {
		rl := utf8.RuneCountInString(text)
		if rl < maxW {
			x += maxW - rl
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
