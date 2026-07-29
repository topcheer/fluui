package component_test

import (
	"fmt"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// p575_examples.go — Batch supplement for missing Example functions (Direction E).
// Each example demonstrates the primary constructor and key setter usage.

// ExampleAICitationList demonstrates citation list usage.
func ExampleAICitationList() {
	cl := component.NewAICitationList()
	cl.AddCitation("Smith 2024", "https://arxiv.org/abs/1234")
	cl.AddCitation("Jones 2023", "")
	fmt.Printf("Count:%d\n", cl.Count())
	// Output: Count:2
}

// ExampleAIMemoryBar demonstrates conversation memory usage bar.
func ExampleAIMemoryBar() {
	m := component.NewAIMemoryBar()
	m.SetUsage(4000, 8000, 16000)
	fmt.Printf("Used:%d%%\n", m.UsedPercent())
	// Output: Used:75%
}

// ExampleAISafetyBadge demonstrates content safety badge.
func ExampleAISafetyBadge() {
	b := component.NewAISafetyBadge()
	b.SetClassification(component.SafetySafe, "clean")
	fmt.Printf("Level:%d\n", b.Level())
	// Output: Level:0
}

// ExampleAccessibilityFocusRing demonstrates keyboard focus ring.
func ExampleAccessibilityFocusRing() {
	r := component.NewAccessibilityFocusRing()
	r.SetFocused(true)
	fmt.Printf("Focused:%v\n", r.Focused())
	// Output: Focused:true
}

// ExampleBufferHealthBar demonstrates render buffer utilization.
func ExampleBufferHealthBar() {
	bh := component.NewBufferHealthBar()
	bh.SetUtilization(65, 100)
	fmt.Printf("Pct:%d\n", bh.Percent())
	// Output: Pct:65
}

// ExampleCacheHitRatioBar demonstrates cache hit ratio display.
func ExampleCacheHitRatioBar() {
	ch := component.NewCacheHitRatioBar()
	ch.SetHits(80)
	ch.SetMisses(20)
	fmt.Printf("HitPct:%d\n", ch.HitPercent())
	// Output: HitPct:80
}

// ExampleConfidenceIntervalChart demonstrates confidence band chart.
func ExampleConfidenceIntervalChart() {
	c := component.NewConfidenceIntervalChart()
	c.SetBounds(10.0, 20.0, 30.0)
	fmt.Printf("Label:%s\n", c.Label())
	// Output: Label:
}

// ExampleContextTrimmer demonstrates context window trimming zones.
func ExampleContextTrimmer() {
	ct := component.NewContextTrimmer()
	ct.SetSegments(2000, 6000, 2000)
	fmt.Printf("Total:%d\n", ct.TotalTokens())
	// Output: Total:10000
}

// ExampleConversationDepthBar demonstrates conversation turn depth.
func ExampleConversationDepthBar() {
	c := component.NewConversationDepthBar()
	c.SetDepth(8, 20)
	fmt.Printf("Depth:%d\n", c.Depth())
	// Output: Depth:8
}

// ExampleDatasetDistributionHistogram demonstrates distribution histogram.
func ExampleDatasetDistributionHistogram() {
	h := component.NewDatasetDistributionHistogram()
	h.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	h.Paint(buf)
	fmt.Println("painted")
	// Output: painted
}

// ExampleDebugOverlay demonstrates debug info panel.
func ExampleDebugOverlay() {
	d := component.NewDebugOverlay()
	d.SetMetrics(60, 142, 8192, 2500)
	fmt.Printf("FPS:%d\n", d.FPS())
	// Output: FPS:60
}

// ExampleDigitalClock demonstrates digital clock display.
func ExampleDigitalClock() {
	dc := component.NewDigitalClock()
	dc.SetTime(14, 30, 45)
	fmt.Printf("Hour:%d\n", dc.Hour())
	// Output: Hour:14
}

// ExampleErrorBoundary demonstrates error display panel.
func ExampleErrorBoundary() {
	eb := component.NewErrorBoundary()
	eb.SetError("test error", "file.go:42")
	fmt.Printf("HasError:%v\n", eb.HasError())
	// Output: HasError:true
}

// ExampleGradientBar demonstrates gradient progress bar.
func ExampleGradientBar() {
	gb := component.NewGradientBar()
	gb.SetValue(50, 100)
	fmt.Printf("Pct:%d\n", gb.Percent())
	// Output: Pct:50
}

// ExampleHallucinationIndicator demonstrates hallucination risk meter.
func ExampleHallucinationIndicator() {
	hi := component.NewHallucinationIndicator()
	hi.SetScores(90, 85)
	fmt.Printf("Risk:%d\n", hi.Risk())
	// Output: Risk:0
}

// ExampleLatencyHeatmap demonstrates latency heatmap grid.
func ExampleLatencyHeatmap() {
	h := component.NewLatencyHeatmap()
	h.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	h.Paint(buf)
	fmt.Println("painted")
	// Output: painted
}

// ExampleMarkdownCodeBlock demonstrates fenced code block rendering.
func ExampleMarkdownCodeBlock() {
	cb := component.NewMarkdownCodeBlock()
	cb.SetLanguage("go")
	cb.SetLines([]string{"func main() {}"})
	fmt.Printf("Lines:%d\n", cb.LineCount())
	// Output: Lines:1
}

// ExampleMarquee demonstrates scrolling text.
func ExampleMarquee() {
	m := component.NewMarquee()
	m.SetText("Scrolling text")
	m.SetOffset(5)
	fmt.Printf("Offset:%d\n", m.Offset())
	// Output: Offset:5
}

// ExampleModelComparisonMatrix demonstrates multi-model comparison.
func ExampleModelComparisonMatrix() {
	m := component.NewModelComparisonMatrix()
	m.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	m.Paint(buf)
	fmt.Println("painted")
	// Output: painted
}

// ExampleModelLatencyGraph demonstrates latency sparkline.
func ExampleModelLatencyGraph() {
	g := component.NewModelLatencyGraph()
	g.AddLatency(100)
	g.AddLatency(200)
	fmt.Printf("Cur:%d\n", g.CurrentLatency())
	// Output: Cur:200
}

// ExampleModelParameterBar demonstrates model parameter display.
func ExampleModelParameterBar() {
	mp := component.NewModelParameterBar()
	mp.SetParams(70, 90, 4096)
	fmt.Printf("Temp:%d\n", mp.Temperature())
	// Output: Temp:70
}

// ExampleModelSwitcher demonstrates model selector.
func ExampleModelSwitcher() {
	ms := component.NewModelSwitcher()
	ms.SetModels("gpt-4", "claude-3")
	fmt.Printf("Active:%d\n", ms.ActiveIndex())
	// Output: Active:0
}

// ExampleOutputFormatSelector demonstrates format picker.
func ExampleOutputFormatSelector() {
	of := component.NewOutputFormatSelector()
	of.SetActive(component.FormatJSON)
	fmt.Printf("Active:%d\n", of.Active())
	// Output: Active:1
}

// ExamplePipelineFlow demonstrates pipeline stages.
func ExamplePipelineFlow() {
	pf := component.NewPipelineFlow()
	pf.AddStage("Input", component.StageDone)
	pf.AddStage("Output", component.StagePending)
	fmt.Printf("Count:%d\n", pf.Count())
	// Output: Count:2
}

// ExamplePromptChain demonstrates prompt execution chain.
func ExamplePromptChain() {
	pc := component.NewPromptChain()
	pc.AddStep("Analyze", component.ChainDone)
	pc.AddStep("Generate", component.ChainActive)
	fmt.Printf("Count:%d\n", pc.Count())
	// Output: Count:2
}

// ExamplePromptVariant demonstrates A/B prompt comparison.
func ExamplePromptVariant() {
	pv := component.NewPromptVariant()
	pv.SetVariantA("A", 85, 1200, false)
	pv.SetVariantB("B", 92, 980, true)
	fmt.Println("ok")
	// Output: ok
}

// ExampleRadialGauge demonstrates circular progress gauge.
func ExampleRadialGauge() {
	rg := component.NewRadialGauge()
	rg.SetValue(65)
	fmt.Printf("Value:%d\n", rg.Value())
	// Output: Value:65
}

// ExampleResponseTimer demonstrates AI response timing.
func ExampleResponseTimer() {
	rt := component.NewResponseTimer()
	rt.SetDurations(120, 3500)
	fmt.Printf("Total:%d\n", rt.TotalDuration())
	// Output: Total:3500
}

// ExampleShortcutList demonstrates keyboard shortcut display.
func ExampleShortcutList() {
	h := component.NewShortcutList()
	h.AddBinding("Q", "Quit")
	h.AddBinding("S", "Save")
	fmt.Printf("Count:%d\n", h.Count())
	// Output: Count:2
}

// ExampleSignalBars demonstrates signal strength indicator.
func ExampleSignalBars() {
	sb := component.NewSignalBars()
	sb.SetLevel(4)
	fmt.Printf("Level:%d\n", sb.Level())
	// Output: Level:4
}

// ExampleStreamingCursor demonstrates animated streaming indicator.
func ExampleStreamingCursor() {
	c := component.NewStreamingCursor()
	c.SetActive(true)
	fmt.Printf("Active:%v\n", c.IsActive())
	// Output: Active:true
}

// ExampleStreamingWord demonstrates word-by-word streaming.
func ExampleStreamingWord() {
	sw := component.NewStreamingWord()
	sw.SetText("hello world foo")
	sw.SetCursor(2)
	fmt.Printf("Words:%d\n", sw.WordCount())
	// Output: Words:2
}

// ExampleThinkingBudget demonstrates reasoning token budget.
func ExampleThinkingBudget() {
	tb := component.NewThinkingBudget()
	tb.SetBudget(500, 1000)
	fmt.Printf("Pct:%d\n", tb.Percent())
	// Output: Pct:50
}

// ExampleThermometer demonstrates vertical temperature gauge.
func ExampleThermometer() {
	th := component.NewThermometer()
	th.SetTemperature(25, 0, 100)
	fmt.Printf("Temp:%d\n", th.Temperature())
	// Output: Temp:25
}

// ExampleTokenCostChart demonstrates cost mini bar chart.
func ExampleTokenCostChart() {
	c := component.NewTokenCostChart()
	c.AddCost(10)
	c.AddCost(20)
	fmt.Printf("Total:%d\n", c.TotalCost())
	// Output: Total:30
}

// ExampleTokenEfficiencyBar demonstrates AI token efficiency.
func ExampleTokenEfficiencyBar() {
	bar := component.NewTokenEfficiencyBar()
	bar.SetTokens(500, 1000, 500)
	fmt.Printf("Eff:%d\n", bar.EfficiencyPercent())
	// Output: Eff:25
}

// ExampleTokenRing demonstrates circular token usage ring.
func ExampleTokenRing() {
	r := component.NewTokenRing()
	r.SetUsage(750, 1000)
	fmt.Printf("Pct:%d\n", r.Percent())
	// Output: Pct:75
}

// ExampleToolCallBadge demonstrates tool call status badge.
func ExampleToolCallBadge() {
	b := component.NewToolCallBadge()
	b.SetCalls(12, 3)
	fmt.Printf("Total:%d\n", b.TotalCalls())
	// Output: Total:15
}

// ExampleVolumeMeter demonstrates audio volume meter.
func ExampleVolumeMeter() {
	vm := component.NewVolumeMeter()
	vm.SetLevel(75)
	fmt.Printf("Level:%d\n", vm.Level())
	// Output: Level:75
}
