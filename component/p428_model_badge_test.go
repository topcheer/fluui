package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestModelBadge_New_P428(t *testing.T) {
	tests := []struct {
		modelID       string
		wantProvider  AIProvider
		wantDisplay   string
		wantProvName  string
	}{
		{"gpt-4o", ProviderOpenAI, "GPT", "OpenAI"},
		{"gpt-3.5-turbo", ProviderOpenAI, "GPT", "OpenAI"},
		{"o1-preview", ProviderOpenAI, "o1", "OpenAI"},
		{"o3-mini", ProviderOpenAI, "o3", "OpenAI"},
		{"claude-sonnet-4-20250514", ProviderAnthropic, "Claude", "Anthropic"},
		{"claude-opus-4-20250514", ProviderAnthropic, "Claude", "Anthropic"},
		{"gemini-1.5-pro", ProviderGoogle, "Gemini", "Google"},
		{"mistral-large-latest", ProviderMistral, "Mistral", "Mistral"},
		{"mixtral-8x7b", ProviderMistral, "Mixtral", "Mistral"},
		{"llama-3.1-70b", ProviderMeta, "Llama", "Meta"},
		{"llama3-8b", ProviderMeta, "Llama3", "Meta"},
		{"command-r-plus", ProviderCohere, "Command", "Cohere"},
		{"qwen-2.5", ProviderLocal, "Qwen", "Local"},
		{"deepseek-v3", ProviderLocal, "DeepSeek", "Local"},
		{"phi-3-mini", ProviderLocal, "Phi", "Local"},
		{"unknown-model", ProviderUnknown, "unknown-model", "Unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.modelID, func(t *testing.T) {
			mb := NewModelBadge(tc.modelID)
			if mb.Provider() != tc.wantProvider {
				t.Errorf("Provider = %v, want %v", mb.Provider(), tc.wantProvider)
			}
			if mb.DisplayName() != tc.wantDisplay {
				t.Errorf("DisplayName = %q, want %q", mb.DisplayName(), tc.wantDisplay)
			}
			if mb.ProviderName() != tc.wantProvName {
				t.Errorf("ProviderName = %q, want %q", mb.ProviderName(), tc.wantProvName)
			}
			if mb.ModelID() != tc.modelID {
				t.Errorf("ModelID = %q, want %q", mb.ModelID(), tc.modelID)
			}
		})
	}
}

func TestModelBadge_SetModelID_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	if mb.Provider() != ProviderOpenAI {
		t.Fatalf("expected OpenAI, got %v", mb.Provider())
	}
	mb.SetModelID("claude-sonnet-4-20250514")
	if mb.Provider() != ProviderAnthropic {
		t.Errorf("expected Anthropic, got %v", mb.Provider())
	}
	if mb.DisplayName() != "Claude" {
		t.Errorf("DisplayName = %q, want %q", mb.DisplayName(), "Claude")
	}
}

func TestModelBadge_ContextWindow_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	if mb.ContextWindow() != 0 {
		t.Error("default context window should be 0")
	}
	mb.SetContextWindow(128000)
	if mb.ContextWindow() != 128000 {
		t.Errorf("ContextWindow = %d, want 128000", mb.ContextWindow())
	}
}

func TestModelBadge_CustomLabel_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	mb.SetCustomLabel("GPT-4o")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	mb.Paint(buf) // should use custom label
}

func TestModelBadge_Style_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	st := buffer.Style{Fg: buffer.Red, Bg: buffer.Black, Flags: buffer.Bold}
	mb.SetStyle(st)
	got := mb.Style()
	if got.Fg != buffer.Red {
		t.Error("style Fg not set")
	}
}

func TestModelBadge_ProviderIcon_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	icon := mb.ProviderIcon()
	if icon == "" {
		t.Error("provider icon should not be empty")
	}
}

func TestModelBadge_Measure_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	sz := mb.Measure(Constraints{})
	if sz.H != 1 {
		t.Errorf("H = %v, want 1", sz.H)
	}
	if sz.W < 6 {
		t.Errorf("W = %v, want >= 6", sz.W)
	}
	// With max width constraint
	sz = mb.Measure(Constraints{MaxWidth: 5})
	if sz.W != 5 {
		t.Errorf("W = %v, want 5 (constrained)", sz.W)
	}
}

func TestModelBadge_Paint_NoPanic_P428(t *testing.T) {
	mb := NewModelBadge("claude-sonnet-4-20250514")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	mb.Paint(buf)
}

func TestModelBadge_Paint_TinyBounds_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	mb.Paint(buf) // should not panic with tiny bounds
}

func TestModelBadge_Paint_ZeroBounds_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	mb.Paint(buf) // should be no-op
}

func TestModelBadge_Paint_LongLabel_P428(t *testing.T) {
	mb := NewModelBadge("unknown-very-long-model-name-that-exceeds-bounds")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	mb.Paint(buf) // should truncate without panic
}

func TestModelBadge_Children_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o")
	if mb.Children() != nil {
		t.Error("Children should return nil")
	}
}

func TestModelBadge_FluentChain_P428(t *testing.T) {
	mb := NewModelBadge("gpt-4o").
		SetContextWindow(128000).
		SetCustomLabel("GPT-4o")
	if mb.ContextWindow() != 128000 {
		t.Error("fluent SetContextWindow failed")
	}
}

func TestDetectProvider_Lowercase_P428(t *testing.T) {
	// Test case-insensitive detection
	prov, _ := detectProvider("GPT-4o")
	if prov != ProviderOpenAI {
		t.Errorf("expected OpenAI for GPT-4o (uppercase), got %v", prov)
	}
	prov, _ = detectProvider("CLAUDE-3-OPUS")
	if prov != ProviderAnthropic {
		t.Errorf("expected Anthropic for CLAUDE-3-OPUS (uppercase), got %v", prov)
	}
}

// --- Benchmarks ---

func BenchmarkModelBadge_Paint_P428(b *testing.B) {
	mb := NewModelBadge("claude-sonnet-4-20250514")
	mb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Paint(buf)
	}
}

func BenchmarkModelBadge_Measure_P428(b *testing.B) {
	mb := NewModelBadge("gpt-4o")
	cs := Constraints{MaxWidth: 40, MaxHeight: 1}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mb.Measure(cs)
	}
}

func BenchmarkDetectProvider_P428(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detectProvider("claude-sonnet-4-20250514")
	}
}
