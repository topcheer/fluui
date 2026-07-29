package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// p561: Coverage tests for 0% functions in component package.
// Each sub-test exercises a previously untested method to improve coverage.

func TestCoverageP561MeasureMethods(t *testing.T) {
	cs := Constraints{MaxWidth: 100, MaxHeight: 20}

	t.Run("AIMemoryBar", func(t *testing.T) {
		c := NewAIMemoryBar()
		c.SetWidth(40)
		s := c.Measure(cs)
		if s.W <= 0 || s.H <= 0 {
			t.Errorf("AIMemoryBar.Measure = %v, want positive", s)
		}
	})

	t.Run("AISafetyBadge", func(t *testing.T) {
		b := NewAISafetyBadge()
		b.SetClassification(SafetySafe, "ok")
		s := b.Measure(cs)
		if s.W <= 0 {
			t.Errorf("AISafetyBadge.Measure = %v, want positive", s)
		}
	})

	t.Run("BufferHealthBar", func(t *testing.T) {
		b := NewBufferHealthBar()
		s := b.Measure(cs)
		if s.W <= 0 {
			t.Errorf("BufferHealthBar.Measure = %v, want positive", s)
		}
	})

	t.Run("CacheHitRatioBar", func(t *testing.T) {
		ch := NewCacheHitRatioBar()
		ch.SetWidth(40)
		s := ch.Measure(cs)
		if s.W <= 0 {
			t.Errorf("CacheHitRatioBar.Measure = %v, want positive", s)
		}
	})

	t.Run("ConversationDepthBar", func(t *testing.T) {
		c := NewConversationDepthBar()
		s := c.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ConversationDepthBar.Measure = %v, want positive", s)
		}
	})

	t.Run("DebugOverlay", func(t *testing.T) {
		d := NewDebugOverlay()
		s := d.Measure(cs)
		if s.W <= 0 {
			t.Errorf("DebugOverlay.Measure = %v, want positive", s)
		}
	})

	t.Run("ErrorBoundary", func(t *testing.T) {
		eb := NewErrorBoundary()
		eb.SetError("test", "detail")
		s := eb.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ErrorBoundary.Measure = %v, want positive", s)
		}
	})

	t.Run("GradientBar", func(t *testing.T) {
		gb := NewGradientBar()
		s := gb.Measure(cs)
		if s.W <= 0 {
			t.Errorf("GradientBar.Measure = %v, want positive", s)
		}
	})

	t.Run("ModelLatencyGraph", func(t *testing.T) {
		g := NewModelLatencyGraph()
		g.SetSize(40, 5)
		s := g.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ModelLatencyGraph.Measure = %v, want positive", s)
		}
	})

	t.Run("ModelParameterBar", func(t *testing.T) {
		mp := NewModelParameterBar()
		s := mp.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ModelParameterBar.Measure = %v, want positive", s)
		}
	})

	t.Run("ModelSwitcher", func(t *testing.T) {
		ms := NewModelSwitcher()
		ms.SetModels("a", "b")
		s := ms.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ModelSwitcher.Measure = %v, want positive", s)
		}
	})

	t.Run("OutputFormatSelector", func(t *testing.T) {
		of := NewOutputFormatSelector()
		s := of.Measure(cs)
		if s.W <= 0 {
			t.Errorf("OutputFormatSelector.Measure = %v, want positive", s)
		}
	})

	t.Run("PipelineFlow", func(t *testing.T) {
		pf := NewPipelineFlow()
		pf.AddStage("A", StageDone)
		s := pf.Measure(cs)
		if s.W <= 0 {
			t.Errorf("PipelineFlow.Measure = %v, want positive", s)
		}
	})

	t.Run("PromptVariant", func(t *testing.T) {
		pv := NewPromptVariant()
		pv.SetVariantA("A", 85, 1200, false)
		s := pv.Measure(cs)
		if s.W <= 0 {
			t.Errorf("PromptVariant.Measure = %v, want positive", s)
		}
	})

	t.Run("ResponseTimer", func(t *testing.T) {
		rt := NewResponseTimer()
		s := rt.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ResponseTimer.Measure = %v, want positive", s)
		}
	})

	t.Run("ShortcutList", func(t *testing.T) {
		h := NewShortcutList()
		h.AddBinding("Q", "Quit")
		s := h.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ShortcutList.Measure = %v, want positive", s)
		}
	})

	t.Run("StreamingCursor", func(t *testing.T) {
		c := NewStreamingCursor()
		s := c.Measure(cs)
		if s.W <= 0 {
			t.Errorf("StreamingCursor.Measure = %v, want positive", s)
		}
	})

	t.Run("ThinkingBudget", func(t *testing.T) {
		tb := NewThinkingBudget()
		s := tb.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ThinkingBudget.Measure = %v, want positive", s)
		}
	})

	t.Run("TokenCostChart", func(t *testing.T) {
		c := NewTokenCostChart()
		c.SetSize(30, 5)
		s := c.Measure(cs)
		if s.W <= 0 {
			t.Errorf("TokenCostChart.Measure = %v, want positive", s)
		}
	})

	t.Run("TokenEfficiencyBar", func(t *testing.T) {
		bar := NewTokenEfficiencyBar()
		s := bar.Measure(cs)
		if s.W <= 0 {
			t.Errorf("TokenEfficiencyBar.Measure = %v, want positive", s)
		}
	})

	t.Run("TokenRing", func(t *testing.T) {
		r := NewTokenRing()
		s := r.Measure(cs)
		if s.W <= 0 {
			t.Errorf("TokenRing.Measure = %v, want positive", s)
		}
	})

	t.Run("ToolCallBadge", func(t *testing.T) {
		b := NewToolCallBadge()
		s := b.Measure(cs)
		if s.W <= 0 {
			t.Errorf("ToolCallBadge.Measure = %v, want positive", s)
		}
	})
}

func TestCoverageP561SetterMethods(t *testing.T) {
	t.Run("AISTreamRenderer_SetStyle", func(t *testing.T) {
		r := NewAIStreamRenderer()
		r.SetStyle(AIStreamRendererStyle{
			Text:      buffer.Style{Fg: buffer.RGB(255, 255, 255)},
			Cursor:    buffer.Style{Fg: buffer.RGB(100, 200, 255)},
			Thinking:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
			Status:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
			TokenInfo: buffer.Style{Fg: buffer.RGB(100, 149, 237)},
			Error:     buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		})
	})

	t.Run("LatencyHeatmap_SetCellWidth", func(t *testing.T) {
		h := NewLatencyHeatmap()
		h.SetCellWidth(5)
	})

	t.Run("DiffPreview_SetShowLineNumbers", func(t *testing.T) {
		d := NewDiffPreview()
		d.SetShowLineNumbers(true)
		d.SetShowStats(false)
	})

	t.Run("MultiSelect_SetCursor", func(t *testing.T) {
		ms := NewMultiSelect()
		ms.SetCursor(0)
	})

	t.Run("RateLimitIndicator_SetStyle", func(t *testing.T) {
		r := NewRateLimitIndicator()
		r.SetStyle(RateLimitStyle{
			Normal:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
			Warning:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
			Critical: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
			Label:    buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		})
	})

	t.Run("RateLimitIndicator_ResetTime", func(t *testing.T) {
		r := NewRateLimitIndicator()
		r.SetLimit(100)
		rt := r.ResetTime()
		_ = rt // time.Time, just exercise
	})

	t.Run("TextArea_SetPrompt", func(t *testing.T) {
		ta := NewTextArea()
		ta.SetPrompt("> ")
		ta.SetPlaceholder("Enter text...")
		ta.Focus()
		ta.Blur()
		ta.SetCharLimit(100)
	})

	t.Run("TokenMeter_SetShowAbs", func(t *testing.T) {
		tm := NewTokenMeter(1000)
		tm.SetShowAbs(true)
	})
}

func TestCoverageP561UtilityMethods(t *testing.T) {
	t.Run("BaseComponent_Paint", func(t *testing.T) {
		var bc BaseComponent
		bc.SetID(GenerateID("test"))
		buf := buffer.NewBuffer(10, 1)
		bc.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
		bc.Paint(buf) // should be no-op
	})

	t.Run("ContextWindowBar_thresholdLevel", func(t *testing.T) {
		c := NewContextWindowBar()
		c.SetContextLimit(10000)
		c.SetUsed(9000) // 90%
		// Exercise Paint to cover thresholdLevel internally
		c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
		buf := buffer.NewBuffer(40, 3)
		c.Paint(buf)
	})

	t.Run("MarkdownTable_Markdown", func(t *testing.T) {
		mt := NewMarkdownTable()
		mt.SetMarkdown("| Col1 | Col2 |\n| --- | --- |\n| A | B |")
		_ = mt.Markdown()
	})

	t.Run("PasswordStrength_Password", func(t *testing.T) {
		ps := NewPasswordStrength()
		ps.SetPassword("MyStr0ng!Pass")
		if pw := ps.Password(); pw != "MyStr0ng!Pass" {
			t.Errorf("Password = %q, want 'MyStr0ng!Pass'", pw)
		}
	})

	t.Run("ThinkingTrace_Elapsed", func(t *testing.T) {
		tt := NewThinkingTrace()
		tt.Start()
		e := tt.Elapsed()
		if e < 0 {
			t.Errorf("Elapsed = %v, want >= 0", e)
		}
	})
}
