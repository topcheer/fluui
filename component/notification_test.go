package component

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Level tests ────────────────────────────────────────────

func TestNotificationLevel_String(t *testing.T) {
	tests := []struct {
		level NotificationLevel
		want  string
	}{
		{LevelInfo, "info"},
		{LevelSuccess, "success"},
		{LevelWarning, "warning"},
		{LevelError, "error"},
		{NotificationLevel(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("level %d String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestNotificationLevel_Icon(t *testing.T) {
	icons := map[NotificationLevel]string{
		LevelInfo:    "ℹ",
		LevelSuccess: "✓",
		LevelWarning: "⚠",
		LevelError:   "✗",
	}
	for level, want := range icons {
		if got := level.Icon(); got != want {
			t.Errorf("level %d Icon() = %q, want %q", level, got, want)
		}
	}
}

func TestNotificationLevel_Colors(t *testing.T) {
	// Verify each level has distinct accent and bg colors.
	levels := []NotificationLevel{LevelInfo, LevelSuccess, LevelWarning, LevelError}
	for _, l := range levels {
		accent := l.AccentColor()
		bg := l.BgColor()
		if accent == (buffer.Color{}) {
			t.Errorf("level %d has zero accent color", l)
		}
		if bg == (buffer.Color{}) {
			t.Errorf("level %d has zero bg color", l)
		}
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  NotificationLevel
	}{
		{"info", LevelInfo},
		{"success", LevelSuccess},
		{"warning", LevelWarning},
		{"warn", LevelWarning},
		{"error", LevelError},
		{"err", LevelError},
		{"unknown", LevelInfo}, // default
		{"", LevelInfo},
	}
	for _, tt := range tests {
		got := ParseLevel(tt.input)
		if got != tt.want {
			t.Errorf("ParseLevel(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestDefaultDurationFor(t *testing.T) {
	if d := DefaultDurationFor(LevelInfo); d != DefaultInfoDuration {
		t.Errorf("info duration = %v, want %v", d, DefaultInfoDuration)
	}
	if d := DefaultDurationFor(LevelSuccess); d != DefaultSuccessDuration {
		t.Errorf("success duration = %v, want %v", d, DefaultSuccessDuration)
	}
	if d := DefaultDurationFor(LevelWarning); d != DefaultWarningDuration {
		t.Errorf("warning duration = %v, want %v", d, DefaultWarningDuration)
	}
	// Error should persist (0 duration).
	if d := DefaultDurationFor(LevelError); d != 0 {
		t.Errorf("error duration = %v, want 0", d)
	}
}

// ─── Notification tests ─────────────────────────────────────

func TestNotification_IsExpired(t *testing.T) {
	// Fresh notification should not be expired.
	n := &Notification{
		Duration:  5 * time.Second,
		CreatedAt: time.Now(),
	}
	if n.IsExpired() {
		t.Error("fresh notification should not be expired")
	}

	// Old notification should be expired.
	n.CreatedAt = time.Now().Add(-10 * time.Second)
	if !n.IsExpired() {
		t.Error("old notification should be expired")
	}

	// Duration 0 means persistent — never expires.
	n2 := &Notification{
		Duration:  0,
		CreatedAt: time.Now().Add(-100 * time.Hour),
	}
	if n2.IsExpired() {
		t.Error("persistent notification (duration=0) should never expire")
	}
}

func TestNotification_RemainingDuration(t *testing.T) {
	n := &Notification{
		Duration:  10 * time.Second,
		CreatedAt: time.Now(),
	}
	rem := n.RemainingDuration()
	if rem <= 0 || rem > 10*time.Second {
		t.Errorf("remaining = %v, expected (0, 10s]", rem)
	}

	// Persistent notification has 0 remaining.
	n2 := &Notification{Duration: 0}
	if rem := n2.RemainingDuration(); rem != 0 {
		t.Errorf("persistent remaining = %v, want 0", rem)
	}

	// Already expired notification has 0 remaining.
	n3 := &Notification{
		Duration:  1 * time.Second,
		CreatedAt: time.Now().Add(-5 * time.Second),
	}
	if rem := n3.RemainingDuration(); rem != 0 {
		t.Errorf("expired remaining = %v, want 0", rem)
	}
}

// ─── ToastManager tests ─────────────────────────────────────

func TestNewToastManager_Defaults(t *testing.T) {
	tm := NewToastManager(0) // 0 should default to 5
	if tm.MaxVisible() != 5 {
		t.Errorf("MaxVisible() = %d, want 5 (default)", tm.MaxVisible())
	}
	if tm.Count() != 0 {
		t.Errorf("Count() = %d, want 0", tm.Count())
	}
	if tm.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestNewToastManager_CustomMaxVisible(t *testing.T) {
	tm := NewToastManager(3)
	if tm.MaxVisible() != 3 {
		t.Errorf("MaxVisible() = %d, want 3", tm.MaxVisible())
	}
}

func TestToastManager_Push(t *testing.T) {
	tm := NewToastManager(10)
	id := tm.Push(LevelInfo, "Title", "Message", 5*time.Second)
	if id == "" {
		t.Error("Push should return non-empty ID")
	}
	if tm.Count() != 1 {
		t.Errorf("Count() = %d, want 1", tm.Count())
	}
	notifications := tm.Notifications()
	if len(notifications) != 1 {
		t.Fatalf("Notifications() len = %d, want 1", len(notifications))
	}
	n := notifications[0]
	if n.Title != "Title" {
		t.Errorf("Title = %q, want %q", n.Title, "Title")
	}
	if n.Message != "Message" {
		t.Errorf("Message = %q, want %q", n.Message, "Message")
	}
	if n.Level != LevelInfo {
		t.Errorf("Level = %d, want %d", n.Level, LevelInfo)
	}
}

func TestToastManager_PushHelpers(t *testing.T) {
	tm := NewToastManager(10)
	tm.PushInfo("info", "msg")
	tm.PushSuccess("success", "msg")
	tm.PushWarning("warning", "msg")
	tm.PushError("error", "msg")
	if tm.Count() != 4 {
		t.Errorf("Count() = %d, want 4", tm.Count())
	}
	notifications := tm.Notifications()
	if notifications[0].Level != LevelInfo {
		t.Error("first should be info")
	}
	if notifications[1].Level != LevelSuccess {
		t.Error("second should be success")
	}
	if notifications[2].Level != LevelWarning {
		t.Error("third should be warning")
	}
	if notifications[3].Level != LevelError {
		t.Error("fourth should be error")
	}
}

func TestToastManager_PushDefaultDuration(t *testing.T) {
	tm := NewToastManager(10)
	id := tm.PushInfo("title", "msg")
	notifications := tm.Notifications()
	for _, n := range notifications {
		if n.ID == id {
			if n.Duration != DefaultInfoDuration {
				t.Errorf("duration = %v, want %v", n.Duration, DefaultInfoDuration)
			}
		}
	}
}

func TestToastManager_Dismiss(t *testing.T) {
	tm := NewToastManager(10)
	id1 := tm.PushInfo("t1", "m1")
	id2 := tm.PushInfo("t2", "m2")
	if tm.Count() != 2 {
		t.Fatalf("Count() = %d, want 2", tm.Count())
	}

	// Dismiss existing.
	if !tm.Dismiss(id1) {
		t.Error("Dismiss should return true for existing ID")
	}
	if tm.Count() != 1 {
		t.Errorf("Count() = %d, want 1", tm.Count())
	}

	// Remaining should be id2.
	remaining := tm.Notifications()
	if remaining[0].ID != id2 {
		t.Errorf("remaining ID = %q, want %q", remaining[0].ID, id2)
	}

	// Dismiss non-existent.
	if tm.Dismiss("nonexistent") {
		t.Error("Dismiss should return false for non-existent ID")
	}
}

func TestToastManager_Clear(t *testing.T) {
	tm := NewToastManager(10)
	tm.PushInfo("t1", "m1")
	tm.PushInfo("t2", "m2")
	tm.PushInfo("t3", "m3")
	if tm.Count() != 3 {
		t.Fatalf("Count() = %d, want 3", tm.Count())
	}
	tm.Clear()
	if tm.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after Clear", tm.Count())
	}
}

func TestToastManager_NotificationsCopy(t *testing.T) {
	tm := NewToastManager(10)
	tm.PushInfo("t1", "m1")
	n1 := tm.Notifications()
	// Verify it's a slice copy — modifying the slice itself (not the pointed-to
	// struct) should not affect the internal state.
	n1 = append(n1, &Notification{ID: "fake", Title: "injected"})
	n2 := tm.Notifications()
	if len(n2) != 1 {
		t.Errorf("internal notifications count = %d, want 1 (slice copy should not leak)", len(n2))
	}
}

func TestToastManager_SetMaxVisible(t *testing.T) {
	tm := NewToastManager(10)
	tm.PushInfo("t1", "m1")
	tm.PushInfo("t2", "m2")
	tm.PushInfo("t3", "m3")

	tm.SetMaxVisible(2)
	if tm.MaxVisible() != 2 {
		t.Errorf("MaxVisible() = %d, want 2", tm.MaxVisible())
	}
}

func TestToastManager_SetMaxVisible_Invalid(t *testing.T) {
	tm := NewToastManager(5)
	tm.SetMaxVisible(0) // should be ignored
	tm.SetMaxVisible(-1)
	if tm.MaxVisible() != 5 {
		t.Errorf("MaxVisible() = %d, want 5 (unchanged)", tm.MaxVisible())
	}
}

func TestToastManager_Tick_Expired(t *testing.T) {
	tm := NewToastManager(10)
	// Push a notification that expires immediately (1ns duration).
	tm.Push(LevelInfo, "fast", "expires now", 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond) // let it expire

	expired := tm.Tick()
	if len(expired) != 1 {
		t.Fatalf("expired count = %d, want 1", len(expired))
	}
	if tm.Count() != 0 {
		t.Errorf("Count() = %d, want 0 after expiry", tm.Count())
	}
}

func TestToastManager_Tick_NotExpired(t *testing.T) {
	tm := NewToastManager(10)
	tm.PushInfo("title", "msg")
	expired := tm.Tick()
	if len(expired) != 0 {
		t.Errorf("expired count = %d, want 0", len(expired))
	}
	if tm.Count() != 1 {
		t.Errorf("Count() = %d, want 1", tm.Count())
	}
}

func TestToastManager_Tick_ErrorPersistent(t *testing.T) {
	tm := NewToastManager(10)
	// Error notifications persist (duration=0).
	tm.PushError("error", "won't expire")
	expired := tm.Tick()
	if len(expired) != 0 {
		t.Errorf("error should not expire, got %d expired", len(expired))
	}
	if tm.Count() != 1 {
		t.Errorf("Count() = %d, want 1", tm.Count())
	}
}

func TestToastManager_Tick_Mixed(t *testing.T) {
	tm := NewToastManager(10)
	tm.PushError("persistent", "won't expire")
	tm.Push(LevelInfo, "fast", "expires", 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond)

	expired := tm.Tick()
	if len(expired) != 1 {
		t.Fatalf("expired = %d, want 1", len(expired))
	}
	if tm.Count() != 1 {
		t.Errorf("Count() = %d, want 1 (error persists)", tm.Count())
	}
	notifications := tm.Notifications()
	if notifications[0].Title != "persistent" {
		t.Errorf("remaining title = %q, want %q", notifications[0].Title, "persistent")
	}
}

func TestToastManager_HasPending(t *testing.T) {
	tm := NewToastManager(10)
	if tm.HasPending() {
		t.Error("HasPending() = true, want false (empty)")
	}
	tm.PushInfo("title", "msg")
	if !tm.HasPending() {
		t.Error("HasPending() = false, want true")
	}
	tm.Clear()
	if tm.HasPending() {
		t.Error("HasPending() = true, want false after Clear")
	}
}

func TestToastManager_Measure_Empty(t *testing.T) {
	tm := NewToastManager(5)
	size := tm.Measure(Constraints{})
	if size.H == 0 {
		t.Error("Measure with empty should have non-zero height (at least 1)")
	}
}

func TestToastManager_Measure_WithNotifications(t *testing.T) {
	tm := NewToastManager(5)
	tm.PushInfo("t1", "m1")
	tm.PushInfo("t2", "m2")
	size := tm.Measure(Constraints{MaxWidth: 60})
	// Each notification is 3 lines.
	if size.H != 6 {
		t.Errorf("H = %d, want 6 (2 notifications * 3 lines)", size.H)
	}
	if size.W <= 0 {
		t.Errorf("W = %d, should be positive", size.W)
	}
}

func TestToastManager_Measure_RespectsMaxVisible(t *testing.T) {
	tm := NewToastManager(2) // max 2 visible
	tm.PushInfo("t1", "m1")
	tm.PushInfo("t2", "m2")
	tm.PushInfo("t3", "m3")
	tm.PushInfo("t4", "m4")
	size := tm.Measure(Constraints{MaxWidth: 60})
	// Only 2 visible despite 4 notifications.
	if size.H != 6 {
		t.Errorf("H = %d, want 6 (2 maxVisible * 3 lines)", size.H)
	}
}

func TestToastManager_Paint_Empty(t *testing.T) {
	tm := NewToastManager(5)
	buf := buffer.NewBuffer(60, 20)
	// Should not panic or crash.
	tm.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	tm.Paint(buf)
}

func TestToastManager_Paint_WithNotifications(t *testing.T) {
	tm := NewToastManager(5)
	tm.PushInfo("Info Title", "Info message")
	tm.PushError("Error Title", "Error message")

	tm.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(60, 20)
	tm.Paint(buf)

	// Check that the title text appears in the buffer.
	found := false
	for y := 0; y < 20; y++ {
		line := collectLine(buf, y, 60)
		if strings.Contains(line, "Info Title") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Paint should render 'Info Title' in the buffer")
	}
}

func TestToastManager_Children(t *testing.T) {
	tm := NewToastManager(5)
	if tm.Children() != nil {
		t.Error("Children() should return nil")
	}
}

func TestToastManager_PushUniqueIDs(t *testing.T) {
	tm := NewToastManager(10)
	id1 := tm.PushInfo("t1", "m1")
	id2 := tm.PushInfo("t2", "m2")
	id3 := tm.PushInfo("t3", "m3")
	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Error("Push should return unique IDs")
	}
}

// ─── Concurrency tests ──────────────────────────────────────

func TestToastManager_ConcurrentAccess(t *testing.T) {
	tm := NewToastManager(100)
	var wg sync.WaitGroup

	// Concurrent pushers.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				tm.PushInfo("title", "msg")
			}
		}(i)
	}

	// Concurrent readers.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_ = tm.Count()
				_ = tm.Notifications()
				_ = tm.HasPending()
			}
		}()
	}

	// Concurrent dismissers.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				notifications := tm.Notifications()
				if len(notifications) > 0 {
					tm.Dismiss(notifications[0].ID)
				}
			}
		}()
	}

	// Concurrent tickers.
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				tm.Tick()
			}
		}()
	}

	wg.Wait()
	// Just verify it doesn't deadlock or panic.
	_ = tm.Count()
}

// ─── Helper ─────────────────────────────────────────────────

func collectLine(buf *buffer.Buffer, y, w int) string {
	var sb strings.Builder
	for x := 0; x < w; x++ {
		c := buf.GetCell(x, y)
		if c.Rune != 0 {
			sb.WriteRune(c.Rune)
		} else {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
