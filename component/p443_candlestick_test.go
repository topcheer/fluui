package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestCandlestickChart_New_P443(t *testing.T) {
	cc := NewCandlestickChart()
	if cc.CandleCount() != 0 {
		t.Errorf("CandleCount = %d, want 0", cc.CandleCount())
	}
}

func TestCandlestickChart_AddCandle_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.AddCandle(Candle{Open: 100, High: 105, Low: 98, Close: 103})
	cc.AddCandle(Candle{Open: 103, High: 108, Low: 101, Close: 99})
	if cc.CandleCount() != 2 {
		t.Errorf("CandleCount = %d, want 2", cc.CandleCount())
	}
}

func TestCandlestickChart_SetCandles_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.SetCandles([]Candle{
		{Open: 10, High: 15, Low: 8, Close: 12},
		{Open: 12, High: 14, Low: 10, Close: 11},
	})
	if cc.CandleCount() != 2 {
		t.Errorf("CandleCount = %d, want 2", cc.CandleCount())
	}
}

func TestCandlestickChart_Candles_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.AddCandle(Candle{Open: 10, High: 15, Low: 8, Close: 12})
	c := cc.Candles()
	if len(c) != 1 || c[0].Open != 10 {
		t.Errorf("Candles mismatch: %v", c)
	}
}

func TestCandlestickChart_Clear_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.AddCandle(Candle{Open: 10, High: 15, Low: 8, Close: 12})
	cc.Clear()
	if cc.CandleCount() != 0 {
		t.Error("should have 0 candles after Clear")
	}
}

func TestCandlestickChart_Style_P443(t *testing.T) {
	cc := NewCandlestickChart()
	st := DefaultCandlestickStyle()
	cc.SetStyle(st)
	if cc.Style().Bullish.Fg != st.Bullish.Fg {
		t.Error("style mismatch")
	}
}

func TestCandlestickChart_Measure_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.AddCandle(Candle{Open: 10, High: 15, Low: 8, Close: 12})
	sz := cc.Measure(Constraints{})
	if sz.H < 10 {
		t.Errorf("H = %d, want >= 10", sz.H)
	}
}

func TestCandlestickChart_Paint_NoPanic_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.AddCandle(Candle{Open: 100, High: 105, Low: 98, Close: 103})
	cc.AddCandle(Candle{Open: 103, High: 108, Low: 101, Close: 99})
	cc.AddCandle(Candle{Open: 99, High: 102, Low: 95, Close: 101})
	cc.AddCandle(Candle{Open: 101, High: 110, Low: 100, Close: 108})
	cc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 15})
	buf := buffer.NewBuffer(40, 15)
	cc.Paint(buf)
}

func TestCandlestickChart_Paint_BullishBearish_P443(t *testing.T) {
	cc := NewCandlestickChart()
	// Bullish: close > open
	cc.AddCandle(Candle{Open: 100, High: 105, Low: 98, Close: 103})
	// Bearish: close < open
	cc.AddCandle(Candle{Open: 103, High: 108, Low: 101, Close: 99})
	cc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	cc.Paint(buf)
}

func TestCandlestickChart_Paint_ZeroBounds_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cc.Paint(buf)
}

func TestCandlestickChart_Paint_Empty_P443(t *testing.T) {
	cc := NewCandlestickChart()
	cc.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	cc.Paint(buf) // no candles, no-op
}

func TestCandlestickChart_Children_P443(t *testing.T) {
	if NewCandlestickChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestPriceToY_P443(t *testing.T) {
	y := priceToY(50, 0, 100, 0, 10)
	if y != 5 {
		t.Errorf("priceToY(50,0,100,0,10) = %d, want 5", y)
	}
	y = priceToY(0, 0, 100, 0, 10)
	if y != 10 {
		t.Errorf("priceToY(0,...) = %d, want 10", y)
	}
	y = priceToY(100, 0, 100, 0, 10)
	if y != 0 {
		t.Errorf("priceToY(100,...) = %d, want 0", y)
	}
}

func BenchmarkCandlestickChart_Paint_P443(b *testing.B) {
	cc := NewCandlestickChart()
	for i := 0; i < 20; i++ {
		cc.AddCandle(Candle{Open: 100, High: 105, Low: 98, Close: 103})
	}
	cc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cc.Paint(buf)
	}
}
