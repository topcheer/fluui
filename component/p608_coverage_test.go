package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P608: Coverage gap fill for Measure and SetWidth methods at 0%.
// These are called by layout managers but not directly in unit tests.

func TestP608MeasureCoverage(t *testing.T) {
	cs := Constraints{MaxWidth: 200, MaxHeight: 50}

	components := []struct {
		name string
		fn   func() Component
	}{
		{"ActivityRing", func() Component { c := NewActivityRing(); c.SetCount(50, 100); return c }},
		{"AICitationList", func() Component { c := NewAICitationList(); c.AddCitation("a", "b"); return c }},
		{"AIConfidenceGauge", func() Component { c := NewAIConfidenceGauge(); c.SetValue(75); return c }},
		{"AIContextBadge", func() Component { c := NewAIContextBadge(); c.SetSource(ContextRAG, "x"); return c }},
		{"AIEmojiReaction", func() Component { c := NewAIEmojiReaction(); c.SetSentiment(ERSentimentPositive, 85); return c }},
		{"AIModelRankList", func() Component { c := NewAIModelRankList(); c.AddModel("M", 90, 1); return c }},
		{"AIPolishIndicator", func() Component { c := NewAIPolishIndicator(); c.SetScores(80, 85, 90); return c }},
		{"AIResponseScore", func() Component { c := NewAIResponseScore(); c.SetScores(80, 70, 90, 60, 85); return c }},
		{"AITokenCounter", func() Component { c := NewAITokenCounter(); c.SetCounts(500, 350); return c }},
		{"AsciiArtBox", func() Component { c := NewAsciiArtBox(); c.SetText("42"); return c }},
		{"BatteryGauge", func() Component { c := NewBatteryGauge(); c.SetLevel(75); return c }},
		{"ColorWheel", func() Component { c := NewColorWheel(); c.SetHue(180); return c }},
		{"CompactStatCard", func() Component { c := NewCompactStatCard(); c.SetLabel("QPS"); return c }},
		{"CompactTimeline", func() Component { c := NewCompactTimeline(); c.AddEvent(50, 0); return c }},
		{"CompassDial", func() Component { c := NewCompassDial(); c.SetHeading(90); return c }},
		{"ContextTrimmer", func() Component { c := NewContextTrimmer(); c.SetSegments(1000, 2000, 1000); return c }},
		{"CountdownTimer", func() Component { c := NewCountdownTimer(); c.SetRemaining(30000); return c }},
		{"EdgeLabel", func() Component { c := NewEdgeLabel(); c.SetLabel("data"); return c }},
		{"FileSizeBar", func() Component { c := NewFileSizeBar(); c.SetSize(1024, 4096); return c }},
		{"HallucinationIndicator", func() Component { c := NewHallucinationIndicator(); c.SetScores(85, 90); return c }},
		{"LogScale", func() Component { c := NewLogScale(); c.SetValue(1000, 1, 1000000); return c }},
		{"LogViewer", func() Component { c := NewLogViewer(); c.AddEntry(LVInfo, "msg"); return c }},
		{"Marquee", func() Component { c := NewMarquee(); c.SetText("scrolling"); return c }},
		{"MarkdownAlert", func() Component { c := NewMarkdownAlert(); c.SetLevel(AlertWarning); c.SetText("test"); return c }},
		{"MarkdownCodeBlock", func() Component { c := NewMarkdownCodeBlock(); c.SetLines([]string{"x"}); return c }},
		{"MarkdownMath", func() Component { c := NewMarkdownMath(); c.SetExpression("E=mc^2"); return c }},
		{"MiniCalendar", func() Component { c := NewMiniCalendar(); c.SetMonth(2024, 3); return c }},
		{"MiniGantt", func() Component { c := NewMiniGantt(); c.AddTask("A", 0, 10, buffer.RGB(0, 255, 0)); return c }},
		{"MiniMap", func() Component { c := NewMiniMap(); c.SetContent(200, 50, 50); return c }},
		{"NetworkLatencyMap", func() Component { c := NewNetworkLatencyMap(); c.SetRegions("A", "B"); return c }},
		{"PromptChain", func() Component { c := NewPromptChain(); c.AddStep("S", ChainDone); return c }},
		{"RadialGauge", func() Component { c := NewRadialGauge(); c.SetValue(50); return c }},
		{"SignalBars", func() Component { c := NewSignalBars(); c.SetLevel(3); return c }},
		{"StepDots", func() Component { c := NewStepDots(); c.SetTotal(5); c.SetCurrent(2); return c }},
		{"Stopwatch", func() Component { c := NewStopwatch(); c.SetElapsed(5000); return c }},
		{"StreamingCursor", func() Component { c := NewStreamingCursor(); return c }},
		{"StreamingDiff", func() Component { c := NewStreamingDiff(); c.SetOldText("a").SetNewText("b"); return c }},
		{"StreamingWord", func() Component { c := NewStreamingWord(); c.SetText("hi"); return c }},
		{"Thermometer", func() Component { c := NewThermometer(); c.SetTemperature(25, 0, 100); return c }},
		{"ThinkingBudget", func() Component { c := NewThinkingBudget(); c.SetBudget(500, 1000); return c }},
		{"ThinkingTrace", func() Component { c := NewThinkingTrace(); c.Start(); return c }},
		{"TokenBudgetBar", func() Component {
			c := NewTokenBudgetBar()
			c.SetZones(TokenZone{Name: "x", Tokens: 100, Color: buffer.RGB(255, 0, 0)})
			return c
		}},
		{"TokenCostChart", func() Component { c := NewTokenCostChart(); return c }},
		{"TokenEfficiencyBar", func() Component { c := NewTokenEfficiencyBar(); c.SetTokens(500, 1000, 500); return c }},
		{"TokenRing", func() Component { c := NewTokenRing(); c.SetUsage(500, 1000); return c }},
		{"ToolCallBadge", func() Component { c := NewToolCallBadge(); c.SetCalls(5, 2); return c }},
		{"VolumeMeter", func() Component { c := NewVolumeMeter(); c.SetLevel(75); return c }},
		{"WaveformDisplay", func() Component { c := NewWaveformDisplay(); c.SetSamples([]int{50, 80, 30}); return c }},
	}

	for _, c := range components {
		t.Run(c.name, func(t *testing.T) {
			comp := c.fn()
			s := comp.Measure(cs)
			if s.W < 0 || s.H < 0 {
				t.Errorf("%s.Measure returned negative size: %v", c.name, s)
			}
		})
	}
}

func TestP608SetWidthCoverage(t *testing.T) {
	// Exercise SetWidth methods that were at 0%
	t.Run("AICitationList", func(t *testing.T) {
		cl := NewAICitationList()
		cl.SetWidth(40)
		if cl.width != 40 {
			t.Errorf("width = %d, want 40", cl.width)
		}
	})
	t.Run("AIModelRankList", func(t *testing.T) {
		rl := NewAIModelRankList()
		rl.SetWidth(40)
		if rl.width != 40 {
			t.Errorf("width = %d, want 40", rl.width)
		}
	})
	t.Run("BatteryGauge", func(t *testing.T) {
		bg := NewBatteryGauge()
		bg.SetWidth(20)
		if bg.width != 20 {
			t.Errorf("width = %d, want 20", bg.width)
		}
	})
	t.Run("CompactTimeline", func(t *testing.T) {
		ct := NewCompactTimeline()
		ct.SetWidth(50)
		if ct.width != 50 {
			t.Errorf("width = %d, want 50", ct.width)
		}
	})
}
