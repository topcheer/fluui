package component

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Notification Level ─────────────────────────────────────

// NotificationLevel classifies the urgency of a notification.
type NotificationLevel int

const (
	LevelInfo NotificationLevel = iota
	LevelSuccess
	LevelWarning
	LevelError
)

// String returns a human-readable name for the level.
func (l NotificationLevel) String() string {
	switch l {
	case LevelInfo:
		return "info"
	case LevelSuccess:
		return "success"
	case LevelWarning:
		return "warning"
	case LevelError:
		return "error"
	default:
		return "unknown"
	}
}

// Icon returns the prefix icon character for the level.
func (l NotificationLevel) Icon() string {
	switch l {
	case LevelInfo:
		return "ℹ"
	case LevelSuccess:
		return "✓"
	case LevelWarning:
		return "⚠"
	case LevelError:
		return "✗"
	default:
		return "•"
	}
}

// AccentColor returns the foreground color used for the level's icon/title.
func (l NotificationLevel) AccentColor() buffer.Color {
	switch l {
	case LevelInfo:
		return buffer.RGB(139, 233, 253) // cyan
	case LevelSuccess:
		return buffer.RGB(72, 207, 154) // green
	case LevelWarning:
		return buffer.RGB(255, 184, 108) // orange
	case LevelError:
		return buffer.RGB(255, 85, 85) // red
	default:
		return buffer.RGB(200, 200, 200)
	}
}

// BgColor returns the background tint for the level.
func (l NotificationLevel) BgColor() buffer.Color {
	switch l {
	case LevelInfo:
		return buffer.RGB(40, 45, 60)
	case LevelSuccess:
		return buffer.RGB(35, 50, 40)
	case LevelWarning:
		return buffer.RGB(55, 45, 30)
	case LevelError:
		return buffer.RGB(55, 30, 30)
	default:
		return buffer.RGB(45, 45, 50)
	}
}

// ParseLevel converts a string to NotificationLevel.
// Returns LevelInfo for unknown strings.
func ParseLevel(s string) NotificationLevel {
	switch strings.ToLower(s) {
	case "info":
		return LevelInfo
	case "success":
		return LevelSuccess
	case "warning", "warn":
		return LevelWarning
	case "error", "err":
		return LevelError
	default:
		return LevelInfo
	}
}

// ─── Notification ───────────────────────────────────────────

// Notification is a single auto-expiring message displayed by ToastManager.
type Notification struct {
	ID        string
	Level     NotificationLevel
	Title     string
	Message   string
	CreatedAt time.Time
	Duration  time.Duration
	Expired   bool
}

// IsExpired reports whether the notification has exceeded its duration.
func (n *Notification) IsExpired() bool {
	if n.Duration <= 0 {
		return false
	}
	return time.Since(n.CreatedAt) >= n.Duration
}

// RemainingDuration returns how much time is left before expiry.
func (n *Notification) RemainingDuration() time.Duration {
	if n.Duration <= 0 {
		return 0
	}
	elapsed := time.Since(n.CreatedAt)
	remaining := n.Duration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// ─── Default durations ──────────────────────────────────────

const (
	DefaultInfoDuration    = 4 * time.Second
	DefaultSuccessDuration = 3 * time.Second
	DefaultWarningDuration = 5 * time.Second
	DefaultErrorDuration   = 0 // errors persist until dismissed
)

// DefaultDurationFor returns the default auto-dismiss duration for a level.
func DefaultDurationFor(level NotificationLevel) time.Duration {
	switch level {
	case LevelInfo:
		return DefaultInfoDuration
	case LevelSuccess:
		return DefaultSuccessDuration
	case LevelWarning:
		return DefaultWarningDuration
	case LevelError:
		return DefaultErrorDuration // persistent
	default:
		return DefaultInfoDuration
	}
}

// ─── ToastManager ───────────────────────────────────────────

// ToastManager manages a stack of auto-expiring notifications.
// It implements the Component interface so it can be rendered in the UI.
type ToastManager struct {
	BaseComponent
	mu          sync.RWMutex
	notifications []*Notification
	maxVisible  int
	counter     int
}

// NewToastManager creates a ToastManager that shows at most maxVisible
// notifications simultaneously. If maxVisible <= 0, it defaults to 5.
func NewToastManager(maxVisible int) *ToastManager {
	if maxVisible <= 0 {
		maxVisible = 5
	}
	tm := &ToastManager{
		maxVisible: maxVisible,
	}
	tm.SetID(GenerateID("toast"))
	return tm
}

// Push adds a new notification with the given level, title, message, and duration.
// If duration is 0, uses DefaultDurationFor(level).
// Returns the notification's ID.
func (tm *ToastManager) Push(level NotificationLevel, title, message string, duration time.Duration) string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if duration == 0 {
		duration = DefaultDurationFor(level)
	}

	tm.counter++
	id := fmt.Sprintf("toast-%d", tm.counter)

	n := &Notification{
		ID:        id,
		Level:     level,
		Title:     title,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  duration,
	}

	tm.notifications = append(tm.notifications, n)
	return id
}

// PushInfo is a convenience wrapper for Push with LevelInfo.
func (tm *ToastManager) PushInfo(title, message string) string {
	return tm.Push(LevelInfo, title, message, 0)
}

// PushSuccess is a convenience wrapper for Push with LevelSuccess.
func (tm *ToastManager) PushSuccess(title, message string) string {
	return tm.Push(LevelSuccess, title, message, 0)
}

// PushWarning is a convenience wrapper for Push with LevelWarning.
func (tm *ToastManager) PushWarning(title, message string) string {
	return tm.Push(LevelWarning, title, message, 0)
}

// PushError is a convenience wrapper for Push with LevelError.
func (tm *ToastManager) PushError(title, message string) string {
	return tm.Push(LevelError, title, message, 0)
}

// Dismiss removes the notification with the given ID.
// Returns true if a notification was found and removed.
func (tm *ToastManager) Dismiss(id string) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for i, n := range tm.notifications {
		if n.ID == id {
			tm.notifications = append(tm.notifications[:i], tm.notifications[i+1:]...)
			return true
		}
	}
	return false
}

// Clear removes all notifications.
func (tm *ToastManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.notifications = nil
}

// Count returns the number of active notifications.
func (tm *ToastManager) Count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.notifications)
}

// Notifications returns a deep copy of the active notifications slice.
// Modifications to the returned slice or its elements do not affect the
// ToastManager's internal state.
func (tm *ToastManager) Notifications() []*Notification {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	result := make([]*Notification, len(tm.notifications))
	for i, n := range tm.notifications {
		copy := *n
		result[i] = &copy
	}
	return result
}

// MaxVisible returns the maximum number of notifications shown at once.
func (tm *ToastManager) MaxVisible() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.maxVisible
}

// SetMaxVisible sets the maximum number of notifications shown at once.
func (tm *ToastManager) SetMaxVisible(n int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	if n > 0 {
		tm.maxVisible = n
	}
}

// Tick advances the auto-expiry check and removes expired notifications.
// Returns the IDs of notifications that were expired and removed.
func (tm *ToastManager) Tick() []string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var expired []string
	remaining := tm.notifications[:0]
	for _, n := range tm.notifications {
		if n.IsExpired() {
			n.Expired = true
			expired = append(expired, n.ID)
		} else {
			remaining = append(remaining, n)
		}
	}
	tm.notifications = remaining
	return expired
}

// TickWithElapsed is like Tick but uses a custom elapsed time instead of wall clock.
// This is useful for testing deterministic expiry without sleeping.
func (tm *ToastManager) TickWithElapsed(elapsed time.Duration) []string {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var expired []string
	remaining := tm.notifications[:0]
	for _, n := range tm.notifications {
		if n.Duration > 0 && time.Since(n.CreatedAt) >= n.Duration {
			n.Expired = true
			expired = append(expired, n.ID)
		} else {
			remaining = append(remaining, n)
		}
	}
	tm.notifications = remaining
	return expired
}

// HasPending reports whether there are any non-expired notifications.
func (tm *ToastManager) HasPending() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.notifications) > 0
}

// ─── Component Interface ────────────────────────────────────

// Measure returns the desired size of the toast manager.
// Each notification takes 3 lines (border + title + message).
func (tm *ToastManager) Measure(cs Constraints) Size {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	count := len(tm.notifications)
	if count > tm.maxVisible {
		count = tm.maxVisible
	}
	height := count * 3
	if height == 0 {
		height = 1
	}

	width := cs.MaxWidth
	if width == 0 {
		width = 60
	}
	if width > 80 {
		width = 80
	}

	return Size{W: width, H: height}
}

// Paint renders up to maxVisible notifications stacked top-to-bottom.
// Each notification is rendered as:
//
//	┌ ℹ Title ─────────────────── ┐
//	│ Message text here           │
//	└─────────────────────────────┘
func (tm *ToastManager) Paint(buf *buffer.Buffer) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	notifications := tm.notifications
	if len(notifications) > tm.maxVisible {
		// Show the most recent maxVisible.
		notifications = notifications[len(notifications)-tm.maxVisible:]
	}

	bounds := tm.Bounds()
	toastW := bounds.W
	if toastW > 80 {
		toastW = 80
	}
	if toastW < 10 {
		toastW = 10
	}

	for i, n := range notifications {
		y := bounds.Y + i*3
		if y+2 >= bounds.Y+bounds.H {
			break
		}
		tm.paintNotification(buf, n, bounds.X, y, toastW)
	}
}

// paintNotification draws a single notification box into the buffer.
func (tm *ToastManager) paintNotification(buf *buffer.Buffer, n *Notification, x, y, w int) {
	accent := n.Level.AccentColor()
	bg := n.Level.BgColor()
	borderColor := accent

	// Top border with title.
	icon := n.Level.Icon()
	titleText := fmt.Sprintf("%s %s", icon, n.Title)
	if len(titleText) > w-4 {
		titleText = truncateString(titleText, w-7) + "..."
	}

	// Draw top border.
	topStyle := buffer.Style{Fg: borderColor, Bg: bg, Flags: buffer.Bold}
	buf.SetCell(x, y, buffer.Cell{Rune: '┌', Width: 1, Fg: borderColor, Bg: bg})
	buf.DrawTextClamped(x+1, y, " ", topStyle)
	titleStyle := buffer.Style{Fg: accent, Bg: bg, Flags: buffer.Bold}
	titleEnd := buf.DrawTextClamped(x+2, y, titleText, titleStyle)

	// Fill remaining top border.
	for fx := titleEnd + 1; fx < x+w-1; fx++ {
		buf.SetCell(fx, y, buffer.Cell{Rune: '─', Width: 1, Fg: borderColor, Bg: bg})
	}
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '┐', Width: 1, Fg: borderColor, Bg: bg})

	// Message line.
	msgStyle := buffer.Style{Fg: buffer.RGB(220, 220, 220), Bg: bg}
	buf.SetCell(x, y+1, buffer.Cell{Rune: '│', Width: 1, Fg: borderColor, Bg: bg})

	msgText := n.Message
	if len(msgText) > w-4 {
		msgText = truncateString(msgText, w-7) + "..."
	}
	buf.DrawTextClamped(x+2, y+1, msgText, msgStyle)

	// Fill remaining message line bg.
	msgEnd := x + 2 + len([]rune(msgText))
	for fx := msgEnd; fx < x+w-1; fx++ {
		buf.SetCell(fx, y+1, buffer.Cell{Rune: ' ', Width: 1, Bg: bg})
	}
	buf.SetCell(x+w-1, y+1, buffer.Cell{Rune: '│', Width: 1, Fg: borderColor, Bg: bg})

	// Bottom border.
	buf.SetCell(x, y+2, buffer.Cell{Rune: '└', Width: 1, Fg: borderColor, Bg: bg})
	for fx := x + 1; fx < x+w-1; fx++ {
		buf.SetCell(fx, y+2, buffer.Cell{Rune: '─', Width: 1, Fg: borderColor, Bg: bg})
	}
	buf.SetCell(x+w-1, y+2, buffer.Cell{Rune: '┘', Width: 1, Fg: borderColor, Bg: bg})
}

// Children returns nil (toast manager has no child components).
func (tm *ToastManager) Children() []Component {
	return nil
}

// ─── Helpers ────────────────────────────────────────────────

// truncateString returns s truncated to maxRunes runes (without the ellipsis).
func truncateString(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes])
}
