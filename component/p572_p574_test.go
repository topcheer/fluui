package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ContextTrimmer Tests ───

func TestContextTrimmerBasic(t *testing.T) {
	ct := NewContextTrimmer()
	ct.SetSegments(2000, 6000, 2000)
	if total := ct.TotalTokens(); total != 10000 {
		t.Errorf("TotalTokens = %d, want 10000", total)
	}
}

func TestContextTrimmerZero(t *testing.T) {
	ct := NewContextTrimmer()
	if total := ct.TotalTokens(); total != 0 {
		t.Errorf("TotalTokens = %d, want 0", total)
	}
}

func TestContextTrimmerNegative(t *testing.T) {
	ct := NewContextTrimmer()
	ct.SetSegments(-10, -20, -30)
	if total := ct.TotalTokens(); total != 0 {
		t.Errorf("TotalTokens = %d, want 0 (clamped)", total)
	}
}

func TestContextTrimmerWidth(t *testing.T) {
	ct := NewContextTrimmer()
	ct.SetWidth(50)
	if ct.width != 50 {
		t.Errorf("width = %d, want 50", ct.width)
	}
	ct.SetWidth(5)
	if ct.width != 20 {
		t.Errorf("width = %d, want 20 (clamped)", ct.width)
	}
}

func TestContextTrimmerPaint(t *testing.T) {
	ct := NewContextTrimmer()
	ct.SetSegments(3000, 4000, 3000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	ct.Paint(buf)
	// Should have bar content
	hasContent := false
	for i := 0; i < 40; i++ {
		r := buf.GetCell(i, 0).Rune
		if r == '█' || r == '▓' || r == '░' {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint should show segmented bar")
	}
}

func TestContextTrimmerChildren(t *testing.T) {
	ct := NewContextTrimmer()
	if c := ct.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestContextTrimmerStyle(t *testing.T) {
	ct := NewContextTrimmer()
	ct.SetStyle(ContextTrimmerStyle{
		Keep:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Retrim:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Discard: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Label:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	ct.SetSegments(1000, 2000, 1000)
	buf := buffer.NewBuffer(40, 3)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	ct.Paint(buf)
}

// ─── PromptChain Tests ───

func TestPromptChainBasic(t *testing.T) {
	pc := NewPromptChain()
	pc.AddStep("Step 1", ChainDone)
	pc.AddStep("Step 2", ChainActive)
	if n := pc.Count(); n != 2 {
		t.Errorf("Count = %d, want 2", n)
	}
}

func TestPromptChainOverflow(t *testing.T) {
	pc := NewPromptChain()
	for i := 0; i < promptChainMaxSteps+5; i++ {
		pc.AddStep("S", ChainPending)
	}
	if n := pc.Count(); n != promptChainMaxSteps {
		t.Errorf("Count = %d, want %d (capped)", n, promptChainMaxSteps)
	}
}

func TestPromptChainClear(t *testing.T) {
	pc := NewPromptChain()
	pc.AddStep("A", ChainDone)
	pc.Clear()
	if n := pc.Count(); n != 0 {
		t.Errorf("Count after Clear = %d, want 0", n)
	}
}

func TestPromptChainSetStatus(t *testing.T) {
	pc := NewPromptChain()
	pc.AddStep("A", ChainPending)
	pc.AddStep("B", ChainPending)
	pc.SetStepStatus(0, ChainDone)
	if pc.steps[0].status != ChainDone {
		t.Errorf("Step 0 status = %d, want ChainDone", pc.steps[0].status)
	}
}

func TestPromptChainSetStatusInvalid(t *testing.T) {
	pc := NewPromptChain()
	pc.AddStep("A", ChainPending)
	pc.SetStepStatus(-1, ChainDone)
	pc.SetStepStatus(99, ChainDone)
}

func TestPromptChainPaint(t *testing.T) {
	pc := NewPromptChain()
	pc.AddStep("Analyze", ChainDone)
	pc.AddStep("Process", ChainActive)
	pc.AddStep("Output", ChainPending)
	pc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	pc.Paint(buf)
	// First row should have ✓ (ChainDone)
	if r := buf.GetCell(0, 0).Rune; r != '✓' {
		t.Errorf("First rune = %q, want '✓'", r)
	}
}

func TestPromptChainChildren(t *testing.T) {
	pc := NewPromptChain()
	if c := pc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestPromptChainStyle(t *testing.T) {
	pc := NewPromptChain()
	pc.SetStyle(PromptChainStyle{
		Done:      buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Active:    buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Pending:   buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Error:     buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Connector: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Name:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	pc.AddStep("A", ChainError)
	buf := buffer.NewBuffer(30, 6)
	pc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	pc.Paint(buf)
}

// ─── StreamingWord Tests ───

func TestStreamingWordBasic(t *testing.T) {
	sw := NewStreamingWord()
	sw.SetText("The quick brown fox")
	sw.SetCursor(2)
	if n := sw.WordCount(); n != 2 {
		t.Errorf("WordCount = %d, want 2", n)
	}
}

func TestStreamingWordTotalWords(t *testing.T) {
	sw := NewStreamingWord()
	sw.SetText("one two three four")
	if n := sw.TotalWords(); n != 4 {
		t.Errorf("TotalWords = %d, want 4", n)
	}
}

func TestStreamingWordEmpty(t *testing.T) {
	sw := NewStreamingWord()
	if n := sw.TotalWords(); n != 0 {
		t.Errorf("TotalWords = %d, want 0", n)
	}
}

func TestStreamingWordCursorOverflow(t *testing.T) {
	sw := NewStreamingWord()
	sw.SetText("a b c")
	sw.SetCursor(100)
	if n := sw.WordCount(); n != 3 {
		t.Errorf("WordCount = %d, want 3 (clamped to total)", n)
	}
}

func TestStreamingWordNegativeCursor(t *testing.T) {
	sw := NewStreamingWord()
	sw.SetText("a b c")
	sw.SetCursor(-5)
	if n := sw.WordCount(); n != 0 {
		t.Errorf("WordCount = %d, want 0 (clamped)", n)
	}
}

func TestStreamingWordPaint(t *testing.T) {
	sw := NewStreamingWord()
	sw.SetText("Hello World Foo")
	sw.SetCursor(2)
	sw.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	sw.Paint(buf)
	// Should have cursor character
	hasCursor := false
	for i := 0; i < 30; i++ {
		if buf.GetCell(i, 0).Rune == '▋' {
			hasCursor = true
			break
		}
	}
	if !hasCursor {
		t.Error("Paint should show cursor character")
	}
}

func TestStreamingWordChildren(t *testing.T) {
	sw := NewStreamingWord()
	if c := sw.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestStreamingWordStyle(t *testing.T) {
	sw := NewStreamingWord()
	sw.SetStyle(StreamingWordStyle{
		Text:     buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Cursor:   buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Unstyled: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	sw.SetText("test words here")
	sw.SetCursor(1)
	buf := buffer.NewBuffer(30, 1)
	sw.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	sw.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintContextTrimmer(b *testing.B) {
	ct := NewContextTrimmer()
	ct.SetSegments(3000, 4000, 3000)
	ct.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ct.Paint(buf)
	}
}

func BenchmarkPaintPromptChain(b *testing.B) {
	pc := NewPromptChain()
	pc.AddStep("Analyze", ChainDone)
	pc.AddStep("Retrieve", ChainActive)
	pc.AddStep("Generate", ChainPending)
	pc.AddStep("Validate", ChainPending)
	pc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 8})
	buf := buffer.NewBuffer(30, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pc.Paint(buf)
	}
}

func BenchmarkPaintStreamingWord(b *testing.B) {
	sw := NewStreamingWord()
	sw.SetText("The quick brown fox jumps over the lazy dog in the park")
	sw.SetCursor(5)
	sw.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 1})
	buf := buffer.NewBuffer(50, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sw.Paint(buf)
	}
}
