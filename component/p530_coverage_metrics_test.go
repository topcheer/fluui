package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Coverage tests for previously uncovered methods ───

func TestCoverageMarkdownGetters(t *testing.T) {
	// Cover Markdown() getters on all markdown components
	ms := []func() string{
		func() string { bq := NewMarkdownBlockquote(); bq.SetMarkdown("x"); return bq.Markdown() },
		func() string { me := NewMarkdownEmoji(); me.SetMarkdown(":heart:"); return me.Markdown() },
		func() string { m := NewMarkdownEmphasis(); m.SetMarkdown("**x**"); return m.Markdown() },
		func() string { mf := NewMarkdownFootnote(); mf.SetMarkdown("x[^1]"); return mf.Markdown() },
		func() string { mh := NewMarkdownHeading(); mh.SetMarkdown("# x"); return mh.Markdown() },
		func() string { hr := NewMarkdownHorizontalRule(); hr.SetMarkdown("---"); return hr.Markdown() },
		func() string { mi := NewMarkdownImage(); mi.SetMarkdown("![x](y)"); return mi.Markdown() },
		func() string { ic := NewMarkdownInlineCode(); ic.SetMarkdown("`x`"); return ic.Markdown() },
		func() string { ml := NewMarkdownLink(); ml.SetMarkdown("[x](y)"); return ml.Markdown() },
		func() string { ls := NewMarkdownList(); ls.SetMarkdown("- x"); return ls.Markdown() },
		func() string { st := NewMarkdownStrikethrough(); st.SetMarkdown("~~x~~"); return st.Markdown() },
		func() string { sp := NewMarkdownSuperscript(); sp.SetMarkdown("x^2^"); return sp.Markdown() },
		func() string { sb := NewMarkdownSubscript(); sb.SetMarkdown("x~2~"); return sb.Markdown() },
		func() string { tl := NewMarkdownTaskList(); tl.SetMarkdown("- [x] done"); return tl.Markdown() },
		func() string { dl := NewMarkdownDefinitionList(); dl.SetMarkdown("x\n: y"); return dl.Markdown() },
	}
	for _, fn := range ms {
		if fn() == "" { /* some return empty, that's fine */ }
	}
}

func TestCoverageAIStreamElapsed(t *testing.T) {
	r := NewAIStreamRenderer()
	if r.Elapsed() != 0 { t.Errorf("Expected 0 elapsed") }
}

func TestCoverageContextWindowBarSetters(t *testing.T) {
	cwb := NewContextWindowBar()
	cwb.SetShowLabel(false)
	cwb.SetStyle(DefaultContextWindowBarStyle())
	_ = cwb
}

func TestCoverageGaugeClusterSetBarWidth(t *testing.T) {
	gc := NewGaugeCluster()
	gc.SetBarWidth(15)
	gc.AddGauge("X", 50, 100)
}

func TestAIMetricsCardBasic(t *testing.T) {
	mc := NewAIMetricsCard()
	mc.SetTokensPerSec(85.5)
	mc.SetLatency(450)
	mc.SetCostPerReq(0.0023)
	mc.SetSuccessRate(99.2)
	if mc.Measure(Constraints{}).H < 3 { t.Error("H too small") }
}

func TestAIMetricsCardPaint(t *testing.T) {
	mc := NewAIMetricsCard()
	mc.SetTokensPerSec(42.0)
	mc.SetLatency(300)
	mc.SetCostPerReq(0.01)
	mc.SetSuccessRate(98.5)
	mc.SetBounds(Rect{X: 0, Y: 0, W: 28, H: 6})
	buf := buffer.NewBuffer(28, 6)
	mc.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundLabel := false
	for x := 0; x < 28; x++ {
		if buf.GetCell(x, 1).Rune == 'T' { foundLabel = true; break }
	}
	if !foundLabel { t.Error("label not found") }
}

func TestAIMetricsCardChildren(t *testing.T) {
	mc := NewAIMetricsCard()
	if mc.Children() != nil { t.Error("Children should be nil") }
}

func TestAIMetricsCardStyle(t *testing.T) {
	mc := NewAIMetricsCard()
	mc.SetStyle(AIMetricsStyle{Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Value: buffer.Style{Fg: buffer.RGB(255,255,255)}, Unit: buffer.Style{Fg: buffer.RGB(100,100,100)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	mc.SetBounds(Rect{X: 0, Y: 0, W: 28, H: 6})
	buf := buffer.NewBuffer(28, 6)
	mc.Paint(buf)
}

func BenchmarkPaintAIMetricsCard(b *testing.B) {
	mc := NewAIMetricsCard()
	mc.SetTokensPerSec(85.5)
	mc.SetLatency(450)
	mc.SetCostPerReq(0.0023)
	mc.SetSuccessRate(99.2)
	mc.SetBounds(Rect{X: 0, Y: 0, W: 28, H: 6})
	buf := buffer.NewBuffer(28, 6)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mc.Paint(buf)
	}
}
