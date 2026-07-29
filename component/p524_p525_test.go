package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestActivityFeedBasic(t *testing.T) {
	af := NewActivityFeed()
	af.AddEntry("alice", "pushed commit", time.Now().Add(-5*time.Minute))
	if af.EntryCount() != 1 { t.Errorf("EntryCount = %d, want 1", af.EntryCount()) }
}

func TestActivityFeedMultiple(t *testing.T) {
	af := NewActivityFeed()
	af.AddEntry("alice", "push", time.Now())
	af.AddEntry("bob", "review", time.Now())
	af.AddEntry("carol", "merge", time.Now())
	if af.EntryCount() != 3 { t.Errorf("EntryCount = %d, want 3", af.EntryCount()) }
}

func TestActivityFeedClear(t *testing.T) {
	af := NewActivityFeed()
	af.AddEntry("a", "b", time.Now())
	af.Clear()
	if af.EntryCount() != 0 { t.Errorf("EntryCount = %d, want 0", af.EntryCount()) }
}

func TestActivityFeedMaxEntries(t *testing.T) {
	af := NewActivityFeed()
	af.SetMaxEntries(2)
	af.AddEntry("a", "x", time.Now())
	af.AddEntry("b", "x", time.Now())
	af.AddEntry("c", "x", time.Now())
	s := af.Measure(Constraints{})
	if s.H > 4 { t.Errorf("H = %d, should be clamped by maxEntries", s.H) }
}

func TestActivityFeedRelativeTime(t *testing.T) {
	if formatRelativeTime(time.Now(), time.Now()) != "now" { t.Error("now should be 'now'") }
	if formatRelativeTime(time.Now().Add(-5*time.Minute), time.Now()) != "5m" { t.Error("5m expected") }
	if formatRelativeTime(time.Now().Add(-2*time.Hour), time.Now()) != "2h" { t.Error("2h expected") }
}

func TestActivityFeedMeasure(t *testing.T) {
	af := NewActivityFeed()
	af.AddEntry("a", "b", time.Now())
	s := af.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestActivityFeedPaint(t *testing.T) {
	af := NewActivityFeed()
	af.AddEntry("alice", "pushed code", time.Now().Add(-5*time.Minute))
	af.AddEntry("bob", "merged PR", time.Now().Add(-1*time.Hour))
	af.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 6})
	buf := buffer.NewBuffer(50, 6)
	af.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundDot := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '●' { foundDot = true; break }
	}
	if !foundDot { t.Error("timeline dot not found") }
}

func TestActivityFeedPaintEmpty(t *testing.T) {
	af := NewActivityFeed()
	af.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	af.Paint(buf)
}

func TestActivityFeedChildren(t *testing.T) {
	af := NewActivityFeed()
	if af.Children() != nil { t.Error("Children should be nil") }
}

func TestActivityFeedStyle(t *testing.T) {
	af := NewActivityFeed()
	af.SetStyle(ActivityFeedStyle{Actor: buffer.Style{Fg: buffer.RGB(0,255,0)}, Action: buffer.Style{Fg: buffer.RGB(200,200,200)}, Time: buffer.Style{Fg: buffer.RGB(150,150,150)}, Dot: buffer.Style{Fg: buffer.RGB(255,0,0)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	af.AddEntry("x", "y", time.Now())
	af.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	af.Paint(buf)
}

// ─── AIPanelHeader tests ───

func TestAIPanelHeaderBasic(t *testing.T) {
	h := NewAIPanelHeader()
	h.SetModel("Claude-3.5")
	h.SetProvider("Anthropic")
	h.SetTokenUsage(5000, 128000)
	h.SetStatus(AIStatusStreaming)
	if h.Model() != "Claude-3.5" { t.Errorf("Model = %q", h.Model()) }
	if h.Provider() != "Anthropic" { t.Errorf("Provider = %q", h.Provider()) }
	used, limit := h.TokenUsage()
	if used != 5000 || limit != 128000 { t.Errorf("Tokens = (%d,%d)", used, limit) }
	if h.Status() != AIStatusStreaming { t.Errorf("Status = %d", h.Status()) }
}

func TestAIPanelHeaderStatusIcons(t *testing.T) {
	if aiStatusIcon(AIStatusIdle) != '●' { t.Error("idle icon") }
	if aiStatusIcon(AIStatusThinking) != '◐' { t.Error("thinking icon") }
	if aiStatusIcon(AIStatusStreaming) != '◉' { t.Error("streaming icon") }
	if aiStatusIcon(AIStatusError) != '✗' { t.Error("error icon") }
}

func TestAIPanelHeaderStatusText(t *testing.T) {
	if aiStatusText(AIStatusIdle) != "idle" { t.Error("idle text") }
	if aiStatusText(AIStatusThinking) != "thinking" { t.Error("thinking text") }
	if aiStatusText(AIStatusStreaming) != "streaming" { t.Error("streaming text") }
}

func TestAIPanelHeaderMeasure(t *testing.T) {
	h := NewAIPanelHeader()
	s := h.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestAIPanelHeaderPaint(t *testing.T) {
	h := NewAIPanelHeader()
	h.SetModel("GPT-4o")
	h.SetProvider("OpenAI")
	h.SetTokenUsage(5000, 128000)
	h.SetStatus(AIStatusStreaming)
	h.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	h.Paint(buf)
	foundModel := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 0).Rune == 'G' { foundModel = true; break }
	}
	if !foundModel { t.Error("model name not found") }
}

func TestAIPanelHeaderPaintIdle(t *testing.T) {
	h := NewAIPanelHeader()
	h.SetStatus(AIStatusIdle)
	h.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	h.Paint(buf)
}

func TestAIPanelHeaderChildren(t *testing.T) {
	h := NewAIPanelHeader()
	if h.Children() != nil { t.Error("Children should be nil") }
}

func TestAIPanelHeaderStyle(t *testing.T) {
	h := NewAIPanelHeader()
	h.SetStyle(AIPanelHeaderStyle{Model: buffer.Style{Fg: buffer.RGB(255,255,255), Flags: buffer.Bold}, Provider: buffer.Style{Fg: buffer.RGB(0,255,0)}, Tokens: buffer.Style{Fg: buffer.RGB(150,150,150)}, Status: [4]buffer.Style{{}, {}, {}, {}}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	h.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	h.Paint(buf)
}
