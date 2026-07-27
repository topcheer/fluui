package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CandlestickChart: Financial OHLC Stock Chart ───
//
// CandlestickChart renders Open-High-Low-Close (OHLC) candlesticks for
// financial data visualization. Green candles indicate price increases
// (close ≥ open), red candles indicate decreases (close < open).
//
// Usage:
//
//	cc := NewCandlestickChart()
//	cc.AddCandle(Candle{Open: 100, High: 105, Low: 98, Close: 103})
//	cc.AddCandle(Candle{Open: 103, High: 108, Low: 101, Close: 99})
//	cc.SetBounds(Rect{X:0, Y:0, W:60, H:15})
//	cc.Paint(buf)

// Candle represents a single OHLC data point.
type Candle struct {
	Open  float64
	High  float64
	Low   float64
	Close float64
}

// CandlestickStyle holds visual styles.
type CandlestickStyle struct {
	Bullish  buffer.Style // green (close >= open)
	Bearish  buffer.Style // red (close < open)
	Wick     buffer.Style // high-low line
	Axis     buffer.Style
}

// DefaultCandlestickStyle returns sensible defaults.
func DefaultCandlestickStyle() CandlestickStyle {
	return CandlestickStyle{
		Bullish: buffer.Style{Fg: buffer.RGB(16, 163, 127)},
		Bearish: buffer.Style{Fg: buffer.RGB(220, 80, 80)},
		Wick:    buffer.Style{Fg: buffer.RGB(120, 120, 120)},
		Axis:    buffer.Style{Fg: buffer.RGB(80, 80, 80)},
	}
}

// CandlestickChart renders OHLC candlesticks.
type CandlestickChart struct {
	BaseComponent
	mu       sync.RWMutex
	candles  []Candle
	style    CandlestickStyle
}

// NewCandlestickChart creates an empty candlestick chart.
func NewCandlestickChart() *CandlestickChart {
	cc := &CandlestickChart{
		style: DefaultCandlestickStyle(),
	}
	cc.SetID(GenerateID("candle"))
	return cc
}

// AddCandle adds an OHLC data point.
func (cc *CandlestickChart) AddCandle(c Candle) *CandlestickChart {
	cc.mu.Lock()
	cc.candles = append(cc.candles, c)
	cc.mu.Unlock()
	return cc
}

// SetCandles replaces all candles.
func (cc *CandlestickChart) SetCandles(candles []Candle) *CandlestickChart {
	cc.mu.Lock()
	cc.candles = candles
	cc.mu.Unlock()
	return cc
}

// Candles returns the current data.
func (cc *CandlestickChart) Candles() []Candle {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.candles
}

// CandleCount returns the number of candles.
func (cc *CandlestickChart) CandleCount() int {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return len(cc.candles)
}

// Clear removes all candles.
func (cc *CandlestickChart) Clear() *CandlestickChart {
	cc.mu.Lock()
	cc.candles = cc.candles[:0]
	cc.mu.Unlock()
	return cc
}

// SetStyle sets the visual style.
func (cc *CandlestickChart) SetStyle(s CandlestickStyle) *CandlestickChart {
	cc.mu.Lock()
	cc.style = s
	cc.mu.Unlock()
	return cc
}

// Style returns the current style.
func (cc *CandlestickChart) Style() CandlestickStyle {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	return cc.style
}

// priceRange returns min low and max high (caller holds lock).
func (cc *CandlestickChart) priceRangeLocked() (min, max float64) {
	if len(cc.candles) == 0 {
		return 0, 1
	}
	min = cc.candles[0].Low
	max = cc.candles[0].High
	for _, c := range cc.candles {
		if c.Low < min {
			min = c.Low
		}
		if c.High > max {
			max = c.High
		}
	}
	if min == max {
		max = min + 1
	}
	return min, max
}

// priceToY converts a price to a y-coordinate.
func priceToY(price, min, max float64, y0, height int) int {
	if max == min {
		return y0
	}
	ratio := (max - price) / (max - min)
	return y0 + int(ratio*float64(height))
}

// Measure computes the desired size.
func (cc *CandlestickChart) Measure(cs Constraints) Size {
	cc.mu.RLock()
	defer cc.mu.RUnlock()
	w := len(cc.candles)*2 + 4
	if w < 20 {
		w = 20
	}
	h := 15
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the candlestick chart.
func (cc *CandlestickChart) Paint(buf *buffer.Buffer) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	b := cc.bounds
	if b.W < 4 || b.H < 3 || len(cc.candles) == 0 {
		return
	}

	min, max := cc.priceRangeLocked()
	chartH := b.H - 1 // reserve 1 for axis
	if chartH < 2 {
		chartH = 2
	}

	candleSpacing := 2
	if b.W/len(cc.candles) < 2 {
		candleSpacing = 1
	}

	for i, candle := range cc.candles {
		x := b.X + i*candleSpacing
		if x >= b.X+b.W {
			break
		}

		// Determine bullish/bearish
		isBullish := candle.Close >= candle.Open
		var bodyStyle buffer.Style
		if isBullish {
			bodyStyle = cc.style.Bullish
		} else {
			bodyStyle = cc.style.Bearish
		}

		// Wick: from Low to High
		wickTopY := priceToY(candle.High, min, max, b.Y, chartH)
		wickBotY := priceToY(candle.Low, min, max, b.Y, chartH)
		for y := wickTopY; y <= wickBotY; y++ {
			if y >= b.Y && y < b.Y+chartH {
				buf.SetCell(x, y, buffer.Cell{Rune: '│', Fg: cc.style.Wick.Fg, Bg: cc.style.Wick.Bg, Width: 1})
			}
		}

		// Body: from Open to Close
		bodyTopY := priceToY(candle.High, min, max, b.Y, chartH)
		bodyOpenY := priceToY(candle.Open, min, max, b.Y, chartH)
		bodyCloseY := priceToY(candle.Close, min, max, b.Y, chartH)
		_ = bodyTopY // suppress unused

		bodyLo := bodyOpenY
		bodyHi := bodyCloseY
		if bodyLo > bodyHi {
			bodyLo, bodyHi = bodyHi, bodyLo
		}

		for y := bodyLo; y <= bodyHi; y++ {
			if y >= b.Y && y < b.Y+chartH {
				buf.SetCell(x, y, buffer.Cell{Rune: '█', Fg: bodyStyle.Fg, Bg: bodyStyle.Bg, Flags: bodyStyle.Flags, Width: 1})
			}
		}
	}

	// Axis line at bottom
	axisY := b.Y + chartH
	for x := 0; x < b.W; x++ {
		buf.SetCell(b.X+x, axisY, buffer.Cell{Rune: '─', Fg: cc.style.Axis.Fg, Bg: cc.style.Axis.Bg, Width: 1})
	}
}

// Children returns nil.
func (cc *CandlestickChart) Children() []Component { return nil }
