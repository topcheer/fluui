package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CompactStatCard Tests ───

func TestCompactStatCardBasic(t *testing.T) {
	sc := NewCompactStatCard()
	sc.SetLabel("QPS")
	sc.SetValue("1234")
	if l := sc.Label(); l != "QPS" {
		t.Errorf("Label = %q, want 'QPS'", l)
	}
}

func TestCompactStatCardTrendUp(t *testing.T) {
	sc := NewCompactStatCard()
	sc.SetTrend(TrendUp, 15)
	if sc.trend != TrendUp {
		t.Error("Expected TrendUp")
	}
}

func TestCompactStatCardTrendDown(t *testing.T) {
	sc := NewCompactStatCard()
	sc.SetTrend(TrendDown, 5)
	if sc.trend != TrendDown {
		t.Error("Expected TrendDown")
	}
}

func TestCompactStatCardTrendFlat(t *testing.T) {
	sc := NewCompactStatCard()
	sc.SetTrend(TrendFlat, 0)
	if sc.trend != TrendFlat {
		t.Error("Expected TrendFlat")
	}
}

func TestCompactStatCardPaint(t *testing.T) {
	sc := NewCompactStatCard()
	sc.SetLabel("Latency")
	sc.SetValue("42ms")
	sc.SetTrend(TrendUp, 12)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 3})
	buf := buffer.NewBuffer(14, 3)
	sc.Paint(buf)
	// Top row should have border
	if r := buf.GetCell(0, 0).Rune; r != '─' {
		t.Errorf("First rune = %q, want '─'", r)
	}
}

func TestCompactStatCardChildren(t *testing.T) {
	sc := NewCompactStatCard()
	if c := sc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestCompactStatCardStyle(t *testing.T) {
	sc := NewCompactStatCard()
	sc.SetStyle(CompactStatCardStyle{
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Value:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Up:     buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Down:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Flat:   buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	sc.SetLabel("CPU").SetValue("85%").SetTrend(TrendDown, 3)
	buf := buffer.NewBuffer(14, 3)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 3})
	sc.Paint(buf)
}

// ─── AIEmojiReaction Tests ───

func TestAIEmojiReactionBasic(t *testing.T) {
	er := NewAIEmojiReaction()
	er.SetSentiment(ERSentimentPositive, 85)
	if s := er.Sentiment(); s != ERSentimentPositive {
		t.Errorf("Sentiment = %d, want ERSentimentPositive", s)
	}
}

func TestAIEmojiReactionAllTypes(t *testing.T) {
	types := []ERSentiment{ERSentimentPositive, ERSentimentNeutral, ERSentimentNegative, ERSentimentMixed}
	for _, s := range types {
		er := NewAIEmojiReaction()
		er.SetSentiment(s, 50)
		if er.Sentiment() != s {
			t.Errorf("Sentiment = %d, want %d", er.Sentiment(), s)
		}
	}
}

func TestAIEmojiReactionInvalidType(t *testing.T) {
	er := NewAIEmojiReaction()
	er.SetSentiment(ERSentiment(99), 50)
	if s := er.Sentiment(); s != ERSentimentNeutral {
		t.Errorf("Sentiment = %d, want ERSentimentNeutral (clamped)", s)
	}
}

func TestAIEmojiReactionClamp(t *testing.T) {
	er := NewAIEmojiReaction()
	er.SetSentiment(ERSentimentPositive, -10)
	if er.confidence != 0 {
		t.Errorf("confidence = %d, want 0 (clamped)", er.confidence)
	}
	er.SetSentiment(ERSentimentPositive, 200)
	if er.confidence != 100 {
		t.Errorf("confidence = %d, want 100 (clamped)", er.confidence)
	}
}

func TestAIEmojiReactionPaint(t *testing.T) {
	er := NewAIEmojiReaction()
	er.SetSentiment(ERSentimentPositive, 85)
	er.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	er.Paint(buf)
	hasContent := false
	for i := 0; i < 20; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestAIEmojiReactionChildren(t *testing.T) {
	er := NewAIEmojiReaction()
	if c := er.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIEmojiReactionStyle(t *testing.T) {
	er := NewAIEmojiReaction()
	er.SetStyle(AIEmojiReactionStyle{
		Emoji:   buffer.Style{Fg: buffer.RGB(255, 255, 0)},
		Label:   buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Pct:     buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Bracket: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	er.SetSentiment(ERSentimentNegative, 30)
	buf := buffer.NewBuffer(20, 1)
	er.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	er.Paint(buf)
}

// ─── TokenBudgetBar Tests ───

func TestTokenBudgetBarBasic(t *testing.T) {
	tb := NewTokenBudgetBar()
	tb.SetZones(
		TokenZone{Name: "sys", Tokens: 2000, Color: buffer.RGB(168, 85, 247)},
		TokenZone{Name: "conv", Tokens: 6000, Color: buffer.RGB(59, 130, 246)},
		TokenZone{Name: "out", Tokens: 2000, Color: buffer.RGB(34, 197, 94)},
	)
	if total := tb.TotalTokens(); total != 10000 {
		t.Errorf("TotalTokens = %d, want 10000", total)
	}
}

func TestTokenBudgetBarEmpty(t *testing.T) {
	tb := NewTokenBudgetBar()
	if total := tb.TotalTokens(); total != 0 {
		t.Errorf("TotalTokens = %d, want 0", total)
	}
}

func TestTokenBudgetBarOverflow(t *testing.T) {
	tb := NewTokenBudgetBar()
	zones := make([]TokenZone, tokenBudgetMaxZones+5)
	for i := range zones {
		zones[i] = TokenZone{Name: "z", Tokens: 100, Color: buffer.RGB(255, 0, 0)}
	}
	tb.SetZones(zones...)
	if n := tb.ZoneCount(); n != tokenBudgetMaxZones {
		t.Errorf("ZoneCount = %d, want %d (capped)", n, tokenBudgetMaxZones)
	}
}

func TestTokenBudgetBarNegative(t *testing.T) {
	tb := NewTokenBudgetBar()
	tb.SetZones(TokenZone{Name: "a", Tokens: -100, Color: buffer.RGB(0, 0, 0)})
	if total := tb.TotalTokens(); total != 0 {
		t.Errorf("TotalTokens = %d, want 0 (clamped)", total)
	}
}

func TestTokenBudgetBarPaint(t *testing.T) {
	tb := NewTokenBudgetBar()
	tb.SetZones(
		TokenZone{Name: "a", Tokens: 500, Color: buffer.RGB(59, 130, 246)},
		TokenZone{Name: "b", Tokens: 500, Color: buffer.RGB(34, 197, 94)},
	)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	buf := buffer.NewBuffer(30, 2)
	tb.Paint(buf)
	hasBar := false
	for i := 0; i < 20; i++ {
		r := buf.GetCell(i, 0).Rune
		if r == '█' || r == '░' {
			hasBar = true
			break
		}
	}
	if !hasBar {
		t.Error("Paint should show bar")
	}
}

func TestTokenBudgetBarChildren(t *testing.T) {
	tb := NewTokenBudgetBar()
	if c := tb.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestTokenBudgetBarStyle(t *testing.T) {
	tb := NewTokenBudgetBar()
	tb.SetStyle(TokenBudgetBarStyle{
		Empty: buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	tb.SetZones(TokenZone{Name: "x", Tokens: 1000, Color: buffer.RGB(255, 0, 0)})
	buf := buffer.NewBuffer(30, 2)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	tb.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintCompactStatCard(b *testing.B) {
	sc := NewCompactStatCard()
	sc.SetLabel("QPS").SetValue("1234").SetTrend(TrendUp, 15)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 14, H: 3})
	buf := buffer.NewBuffer(14, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sc.Paint(buf)
	}
}

func BenchmarkPaintAIEmojiReaction(b *testing.B) {
	er := NewAIEmojiReaction()
	er.SetSentiment(ERSentimentPositive, 85)
	er.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		er.Paint(buf)
	}
}

func BenchmarkPaintTokenBudgetBar(b *testing.B) {
	tb := NewTokenBudgetBar()
	tb.SetZones(
		TokenZone{Name: "sys", Tokens: 2000, Color: buffer.RGB(168, 85, 247)},
		TokenZone{Name: "conv", Tokens: 6000, Color: buffer.RGB(59, 130, 246)},
		TokenZone{Name: "out", Tokens: 2000, Color: buffer.RGB(34, 197, 94)},
	)
	tb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 2})
	buf := buffer.NewBuffer(30, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Paint(buf)
	}
}
