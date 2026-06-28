package component

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- Construction ---

func TestNewStatusBar(t *testing.T) {
	sb := NewStatusBar()
	if sb == nil {
		t.Fatal("NewStatusBar returned nil")
	}
	if sb.ID() == "" {
		t.Error("ID should not be empty")
	}
	if sb.ItemCount() != 0 {
		t.Errorf("new bar should have 0 items, got %d", sb.ItemCount())
	}
}

func TestStatusBar_ID(t *testing.T) {
	sb := NewStatusBar()
	sb.SetID("test-statusbar")
	if sb.ID() != "test-statusbar" {
		t.Errorf("expected 'test-statusbar', got %q", sb.ID())
	}
}

func TestStatusBar_UniqueIDs(t *testing.T) {
	sb1 := NewStatusBar()
	sb2 := NewStatusBar()
	if sb1.ID() == sb2.ID() {
		t.Error("two status bars should have different IDs")
	}
}

func TestStatusBar_ImplementsComponent(t *testing.T) {
	var _ Component = NewStatusBar()
}

// --- Add items ---

func TestStatusBar_AddItem(t *testing.T) {
	sb := NewStatusBar()
	sb.AddItem(StatusItem{ID: "model", Text: "GPT-4", Align: StatusAlignLeft})
	sb.AddItem(StatusItem{ID: "time", Text: "12:00", Align: StatusAlignRight})
	if sb.ItemCount() != 2 {
		t.Errorf("expected 2 items, got %d", sb.ItemCount())
	}
}

func TestStatusBar_AddLeft(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	items := sb.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Align != StatusAlignLeft {
		t.Errorf("expected AlignLeft, got %v", items[0].Align)
	}
	if items[0].Text != "GPT-4" {
		t.Errorf("expected 'GPT-4', got %q", items[0].Text)
	}
}

func TestStatusBar_AddCenter(t *testing.T) {
	sb := NewStatusBar()
	sb.AddCenter("info", "Processing...")
	items := sb.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Align != StatusAlignCenter {
		t.Errorf("expected AlignCenter, got %v", items[0].Align)
	}
}

func TestStatusBar_AddRight(t *testing.T) {
	sb := NewStatusBar()
	sb.AddRight("time", "12:00")
	items := sb.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Align != StatusAlignRight {
		t.Errorf("expected AlignRight, got %v", items[0].Align)
	}
}

// --- Remove items ---

func TestStatusBar_RemoveItem(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddRight("time", "12:00")
	if !sb.RemoveItem("model") {
		t.Error("RemoveItem should return true for existing item")
	}
	if sb.ItemCount() != 1 {
		t.Errorf("expected 1 item after removal, got %d", sb.ItemCount())
	}
	if sb.RemoveItem("nonexistent") {
		t.Error("RemoveItem should return false for non-existent item")
	}
}

func TestStatusBar_Clear(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("a", "1")
	sb.AddLeft("b", "2")
	sb.AddRight("c", "3")
	sb.Clear()
	if sb.ItemCount() != 0 {
		t.Errorf("expected 0 items after Clear, got %d", sb.ItemCount())
	}
}

// --- Update items ---

func TestStatusBar_SetItemText(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.SetItemText("model", "Claude")
	items := sb.Items()
	if items[0].Text != "Claude" {
		t.Errorf("expected 'Claude', got %q", items[0].Text)
	}
}

func TestStatusBar_SetItemText_NonExistent(t *testing.T) {
	sb := NewStatusBar()
	sb.SetItemText("new-item", "text")
	if sb.ItemCount() != 1 {
		t.Errorf("expected 1 item, got %d", sb.ItemCount())
	}
	items := sb.Items()
	if items[0].ID != "new-item" {
		t.Errorf("expected 'new-item', got %q", items[0].ID)
	}
	if items[0].Align != StatusAlignLeft {
		t.Errorf("expected AlignLeft for auto-added item, got %v", items[0].Align)
	}
}

func TestStatusBar_SetItemStyle(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	customStyle := buffer.Style{Fg: buffer.RGB(255, 0, 0)}
	sb.SetItemStyle("model", customStyle)
	items := sb.Items()
	if !items[0].Style.Fg.Equal(customStyle.Fg) {
		t.Error("style was not set correctly")
	}
}

// --- Items() returns copies ---

func TestStatusBar_ItemsReturnsCopy(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	items := sb.Items()
	items[0].Text = "modified"
	original := sb.Items()
	if original[0].Text != "GPT-4" {
		t.Error("Items() should return a copy, not a reference")
	}
}

// --- Measure ---

func TestStatusBar_Measure(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddRight("time", "12:00")
	size := sb.Measure(Constraints{MaxWidth: 100, MaxHeight: 10})
	if size.H != 1 {
		t.Errorf("expected height 1, got %d", size.H)
	}
	if size.W <= 0 {
		t.Error("width should be > 0")
	}
}

func TestStatusBar_Measure_Empty(t *testing.T) {
	sb := NewStatusBar()
	size := sb.Measure(Constraints{MaxWidth: 80, MaxHeight: 1})
	if size.H != 1 {
		t.Errorf("expected height 1, got %d", size.H)
	}
}

func TestStatusBar_Measure_ClampedToMaxWidth(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "A very long text that exceeds width")
	size := sb.Measure(Constraints{MaxWidth: 10, MaxHeight: 1})
	if size.W > 10 {
		t.Errorf("width should not exceed MaxWidth: got %d", size.W)
	}
}

// --- SetBounds / Bounds ---

func TestStatusBar_SetBounds(t *testing.T) {
	sb := NewStatusBar()
	r := Rect{X: 0, Y: 24, W: 80, H: 1}
	sb.SetBounds(r)
	b := sb.Bounds()
	if b.W != 80 || b.H != 1 || b.Y != 24 {
		t.Errorf("bounds not set correctly: %+v", b)
	}
}

// --- Paint ---

func TestStatusBar_Paint_NoPanic(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddCenter("status", "Ready")
	sb.AddRight("time", "12:00")
	buf := buffer.NewBuffer(80, 1)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	sb.Paint(buf)
}

func TestStatusBar_Paint_ZeroBounds(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	buf := buffer.NewBuffer(80, 1)
	sb.Paint(buf) // should not panic
}

func TestStatusBar_Paint_LeftText(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	buf := newTestBuffer(80, 1)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	sb.Paint(buf)
	text := cellRunes(buf, 0, 0, 10)
	if text == "" {
		t.Error("expected non-empty left text in buffer")
	}
}

func TestStatusBar_Paint_RightText(t *testing.T) {
	sb := NewStatusBar()
	sb.AddRight("time", "12:00:00")
	buf := newTestBuffer(80, 1)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	sb.Paint(buf)
	text := cellRunes(buf, 66, 0, 14)
	if text == "" {
		t.Error("expected non-empty right text in buffer")
	}
}

func TestStatusBar_Paint_CenterText(t *testing.T) {
	sb := NewStatusBar()
	sb.AddCenter("status", "READY")
	buf := newTestBuffer(80, 1)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	sb.Paint(buf)
	text := cellRunes(buf, 36, 0, 10)
	if text == "" {
		t.Error("expected non-empty center text in buffer")
	}
}

func TestStatusBar_Paint_AllSegments(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "Claude-3")
	sb.AddCenter("status", "Thinking")
	sb.AddRight("time", "09:30")
	buf := newTestBuffer(80, 1)
	sb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	sb.Paint(buf)

	// All three segments should have content
	leftText := cellRunes(buf, 0, 0, 10)
	if leftText == "" {
		t.Error("expected left text")
	}
	rightText := cellRunes(buf, 68, 0, 12)
	if rightText == "" {
		t.Error("expected right text")
	}
}

// --- Style ---

func TestStatusBar_SetStyle(t *testing.T) {
	sb := NewStatusBar()
	custom := StatusBarStyle{
		Background: buffer.Style{Fg: buffer.RGB(255, 255, 255), Bg: buffer.RGB(0, 0, 0)},
	}
	sb.SetStyle(custom)
	s := sb.Style()
	if s.Background.Fg.IsDefault() {
		t.Error("style Fg should not be default")
	}
}

func TestStatusBar_DefaultStatusBarStyle(t *testing.T) {
	s := DefaultStatusBarStyle()
	if s.Background.Fg.IsDefault() {
		t.Error("Background.Fg should not be default")
	}
}

// --- Separator ---

func TestStatusBar_SetSeparator(t *testing.T) {
	sb := NewStatusBar()
	sb.SetSeparator(" :: ")
	if sb.Separator() != " :: " {
		t.Errorf("expected ' :: ', got %q", sb.Separator())
	}
}

func TestStatusBar_DefaultSeparator(t *testing.T) {
	sb := NewStatusBar()
	if sb.Separator() != " │ " {
		t.Errorf("expected ' │ ', got %q", sb.Separator())
	}
}

// --- Children ---

func TestStatusBar_Children(t *testing.T) {
	sb := NewStatusBar()
	if sb.Children() != nil {
		t.Error("StatusBar should have no children")
	}
}

// --- Convenience methods ---

func TestStatusBar_SetModel(t *testing.T) {
	sb := NewStatusBar()
	sb.SetModel("GPT-4")
	items := sb.Items()
	found := false
	for _, it := range items {
		if it.ID == "model" && it.Text == "GPT-4" {
			found = true
		}
	}
	if !found {
		t.Error("SetModel did not create/update 'model' item correctly")
	}
}

func TestStatusBar_SetTokenRate(t *testing.T) {
	sb := NewStatusBar()
	sb.SetTokenRate(42)
	items := sb.Items()
	found := false
	for _, it := range items {
		if it.ID == "tokenrate" && it.Text == "42 tok/s" {
			found = true
		}
	}
	if !found {
		t.Error("SetTokenRate did not create 'tokenrate' item correctly")
	}
}

func TestStatusBar_SetTokenRate_K(t *testing.T) {
	sb := NewStatusBar()
	sb.SetTokenRate(1500)
	items := sb.Items()
	for _, it := range items {
		if it.ID == "tokenrate" {
			if it.Text != "1.5k tok/s" {
				t.Errorf("expected '1.5k tok/s', got %q", it.Text)
			}
		}
	}
}

func TestStatusBar_SetContextWindow(t *testing.T) {
	sb := NewStatusBar()
	sb.SetContextWindow(4000, 8000)
	items := sb.Items()
	found := false
	for _, it := range items {
		if it.ID == "context" && it.Text == "4000/8000 (50%)" {
			found = true
		}
	}
	if !found {
		t.Error("SetContextWindow did not create 'context' item correctly")
	}
}

func TestStatusBar_SetClock(t *testing.T) {
	sb := NewStatusBar()
	sb.SetClock(time.Date(2024, 1, 1, 12, 30, 45, 0, time.UTC))
	items := sb.Items()
	found := false
	for _, it := range items {
		if it.ID == "clock" && it.Text == "12:30:45" {
			found = true
		}
	}
	if !found {
		t.Error("SetClock did not create 'clock' item correctly")
	}
}

// --- Format helpers ---

func TestStatusBar_FormatTokenRate(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0 tok/s"},
		{-5, "0 tok/s"},
		{42, "42 tok/s"},
		{999, "999 tok/s"},
		{1000, "1k tok/s"},
		{1500, "1.5k tok/s"},
		{12500, "12.5k tok/s"},
	}
	for _, tt := range tests {
		got := formatTokenRate(tt.input)
		if got != tt.want {
			t.Errorf("formatTokenRate(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStatusBar_FormatContextWindow(t *testing.T) {
	got := formatContextWindow(4000, 8000)
	want := "4000/8000 (50%)"
	if got != want {
		t.Errorf("formatContextWindow(4000, 8000) = %q, want %q", got, want)
	}

	got = formatContextWindow(0, 0)
	want = "0/0 (0%)"
	if got != want {
		t.Errorf("formatContextWindow(0, 0) = %q, want %q", got, want)
	}
}

func TestStatusBar_Itoa(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{42, "42"},
		{-42, "-42"},
		{999999, "999999"},
	}
	for _, tt := range tests {
		got := itoa(tt.input)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- SetHeight ---

func TestStatusBar_SetHeight(t *testing.T) {
	sb := NewStatusBar()
	sb.SetHeight(2)
	size := sb.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if size.H != 2 {
		t.Errorf("expected height 2, got %d", size.H)
	}
}

func TestStatusBar_SetHeight_ClampedToMin1(t *testing.T) {
	sb := NewStatusBar()
	sb.SetHeight(0)
	size := sb.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	if size.H != 1 {
		t.Errorf("expected height 1 (clamped), got %d", size.H)
	}
}

// --- Multiple updates ---

func TestStatusBar_MultipleUpdates(t *testing.T) {
	sb := NewStatusBar()
	sb.SetModel("GPT-4")
	sb.SetTokenRate(100)
	sb.SetContextWindow(1000, 4000)
	sb.SetClock(time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC))

	if sb.ItemCount() != 4 {
		t.Errorf("expected 4 items, got %d", sb.ItemCount())
	}

	// Update model again
	sb.SetModel("Claude-3")
	if sb.ItemCount() != 4 {
		t.Errorf("expected 4 items after update, got %d", sb.ItemCount())
	}
}

// --- Concurrency ---

func TestStatusBar_ConcurrentAccess(t *testing.T) {
	sb := NewStatusBar()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				sb.AddLeft(fmt.Sprintf("key-%d-%d", n, j), "val")
				sb.SetItemText(fmt.Sprintf("key-%d-%d", n, j), "updated")
			}
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = sb.Items()
				_ = sb.ItemCount()
				_ = sb.Style()
				_ = sb.Separator()
			}
		}()
	}

	wg.Wait()
}

func TestStatusBar_ConcurrentPaint(t *testing.T) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddCenter("status", "Thinking")
	sb.AddRight("time", "12:00")
	sb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})

	var wg sync.WaitGroup

	// Painters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := newTestBuffer(60, 1)
				sb.Paint(buf)
			}
		}()
	}

	// Writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				sb.SetItemText("model", fmt.Sprintf("Model-%d", j))
			}
		}(i)
	}

	wg.Wait()
}
