package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestStockTicker_New_P450(t *testing.T) {
	st := NewStockTicker("AAPL", 189.50, 1.25)
	if st.Symbol() != "AAPL" {
		t.Errorf("Symbol = %q", st.Symbol())
	}
	if st.Price() != 189.50 {
		t.Errorf("Price = %v", st.Price())
	}
	if st.Change() != 1.25 {
		t.Errorf("Change = %v", st.Change())
	}
}

func TestStockTicker_IsUpDown_P450(t *testing.T) {
	st := NewStockTicker("AAPL", 100, 1.0)
	if !st.IsUp() {
		t.Error("should be up")
	}
	if st.IsDown() {
		t.Error("should not be down")
	}
	st.SetPrice(98) // prevClose=99, price=98, change=-1
	if !st.IsDown() {
		t.Error("should be down after price drop")
	}
}

func TestStockTicker_PercentChange_P450(t *testing.T) {
	st := NewStockTicker("X", 110, 10) // prevClose=100
	pct := st.PercentChange()
	if pct < 9.9 || pct > 10.1 {
		t.Errorf("PercentChange = %v, want ~10", pct)
	}
}

func TestStockTicker_SetPrice_P450(t *testing.T) {
	st := NewStockTicker("X", 100, 0)
	st.SetPrice(105)
	if st.Change() != 5 {
		t.Errorf("Change = %v, want 5", st.Change())
	}
}

func TestStockTicker_SetSymbol_P450(t *testing.T) {
	st := NewStockTicker("X", 100, 0)
	st.SetSymbol("GOOG")
	if st.Symbol() != "GOOG" {
		t.Errorf("Symbol = %q", st.Symbol())
	}
}

func TestStockTicker_ShowPct_P450(t *testing.T) {
	st := NewStockTicker("X", 100, 1)
	if !st.ShowPct() {
		t.Error("default should show pct")
	}
	st.SetShowPct(false)
	if st.ShowPct() {
		t.Error("should be false")
	}
}

func TestStockTicker_Style_P450(t *testing.T) {
	st := NewStockTicker("X", 100, 1)
	s := DefaultStockTickerStyle()
	st.SetStyle(s)
	if st.Style().Up.Fg != s.Up.Fg {
		t.Error("style mismatch")
	}
}

func TestStockTicker_Measure_P450(t *testing.T) {
	st := NewStockTicker("X", 100, 1)
	sz := st.Measure(Constraints{})
	if sz.H != 1 {
		t.Errorf("H = %d, want 1", sz.H)
	}
}

func TestStockTicker_Paint_NoPanic_P450(t *testing.T) {
	st := NewStockTicker("AAPL", 189.50, 1.25)
	st.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	st.Paint(buf)
}

func TestStockTicker_Paint_Down_P450(t *testing.T) {
	st := NewStockTicker("TSLA", 240.50, -3.75)
	st.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	st.Paint(buf)
}

func TestStockTicker_Paint_ZeroBounds_P450(t *testing.T) {
	st := NewStockTicker("X", 100, 1)
	st.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	st.Paint(buf)
}

func TestStockTicker_Children_P450(t *testing.T) {
	if NewStockTicker("X", 1, 0).Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkStockTicker_Paint_P450(b *testing.B) {
	st := NewStockTicker("AAPL", 189.50, 1.25)
	st.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		st.Paint(buf)
	}
}
