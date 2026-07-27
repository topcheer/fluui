package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestAIStreamRenderer_New_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	if r.Status() != AIStreamIdle {
		t.Errorf("Status = %v, want Idle", r.Status())
	}
	if r.Text() != "" {
		t.Errorf("Text = %q, want empty", r.Text())
	}
}

func TestAIStreamRenderer_Start_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	if r.Status() != AIStreamThinking {
		t.Errorf("Status = %v, want Thinking", r.Status())
	}
}

func TestAIStreamRenderer_StartWithModel_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.StartWithModel("gpt-4o")
	if r.Model() != "gpt-4o" {
		t.Errorf("Model = %q", r.Model())
	}
	if r.Status() != AIStreamThinking {
		t.Errorf("Status = %v", r.Status())
	}
}

func TestAIStreamRenderer_Append_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.Append("Hello ")
	r.Append("world")
	if r.Text() != "Hello world" {
		t.Errorf("Text = %q", r.Text())
	}
	if r.Status() != AIStreamStreaming {
		t.Errorf("Status = %v, want Streaming", r.Status())
	}
}

func TestAIStreamRenderer_SetText_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.SetText("override")
	if r.Text() != "override" {
		t.Errorf("Text = %q", r.Text())
	}
}

func TestAIStreamRenderer_SetTokens_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.SetTokens(42, 15.5)
	if r.TokenCount() != 42 {
		t.Errorf("TokenCount = %d", r.TokenCount())
	}
	if r.TPS() != 15.5 {
		t.Errorf("TPS = %v", r.TPS())
	}
}

func TestAIStreamRenderer_Complete_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.Append("text")
	r.Complete("stop")
	if r.Status() != AIStreamDone {
		t.Errorf("Status = %v, want Done", r.Status())
	}
}

func TestAIStreamRenderer_SetError_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.SetError("timeout")
	if r.Status() != AIStreamError {
		t.Errorf("Status = %v, want Error", r.Status())
	}
}

func TestAIStreamRenderer_Model_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.SetModel("claude-sonnet-4")
	if r.Model() != "claude-sonnet-4" {
		t.Errorf("Model = %q", r.Model())
	}
}

func TestAIStreamRenderer_Measure_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	sz := r.Measure(Constraints{})
	if sz.W < 10 || sz.H < 3 {
		t.Errorf("size too small: %v", sz)
	}
}

func TestAIStreamRenderer_Paint_Thinking_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	r.Paint(buf)
}

func TestAIStreamRenderer_Paint_Streaming_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.StartWithModel("gpt-4o")
	r.Append("Hello **world**!\n\nThis is streaming.")
	r.SetTokens(42, 15.5)
	r.SetCursor(true)
	r.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	r.Paint(buf)
}

func TestAIStreamRenderer_Paint_Done_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.Append("Done!")
	r.Complete("stop")
	r.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	r.Paint(buf)
}

func TestAIStreamRenderer_Paint_Error_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.Start()
	r.SetError("rate limit")
	r.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	r.Paint(buf)
}

func TestAIStreamRenderer_Paint_ZeroBounds_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	r.Paint(buf)
}

func TestAIStreamRenderer_Paint_NoStatus_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.SetShowStatus(false)
	r.SetText("text")
	r.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	r.Paint(buf)
}

func TestAIStreamRenderer_Paint_NoTokens_P451(t *testing.T) {
	r := NewAIStreamRenderer()
	r.SetShowTokens(false)
	r.Start()
	r.SetText("text")
	r.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	r.Paint(buf)
}

func TestAIStreamRenderer_Children_P451(t *testing.T) {
	if NewAIStreamRenderer().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkAIStreamRenderer_Paint_P451(b *testing.B) {
	r := NewAIStreamRenderer()
	r.StartWithModel("gpt-4o")
	r.Append("Hello world! This is a test of the AI streaming renderer. ")
	r.Append("It handles markdown **bold** and *italic* text.\n\nNew paragraph.")
	r.SetTokens(128, 25.3)
	r.SetCursor(true)
	r.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Paint(buf)
	}
}
