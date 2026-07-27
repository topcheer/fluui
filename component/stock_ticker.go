package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StockTicker: Real-time Stock Price Display ───
//
// StockTicker renders a compact stock quote with price, change, and
// percentage change. Green for gains, red for losses. Common in financial
// dashboards and AI-powered trading interfaces.
//
// Usage:
//
//	st := NewStockTicker("AAPL", 189.50, +1.25)
//	st.Paint(buf) // renders "AAPL  $189.50  ▲1.25 (+0.66%)"

// StockTickerStyle holds visual styles.
type StockTickerStyle struct {
	Symbol  buffer.Style
	Price   buffer.Style
	Up      buffer.Style
	Down    buffer.Style
	Neutral buffer.Style
}

// DefaultStockTickerStyle returns sensible defaults.
func DefaultStockTickerStyle() StockTickerStyle {
	return StockTickerStyle{
		Symbol:  buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Price:   buffer.Style{Fg: buffer.White},
		Up:      buffer.Style{Fg: buffer.RGB(16, 163, 127)},   // green
		Down:    buffer.Style{Fg: buffer.RGB(220, 80, 80)},    // red
		Neutral: buffer.Style{Fg: buffer.RGB(150, 150, 150)},  // gray
	}
}

// StockTicker displays a stock quote with price change.
type StockTicker struct {
	BaseComponent
	mu        sync.RWMutex
	symbol    string
	price     float64
	change    float64
	prevClose float64
	style     StockTickerStyle
	showPct   bool
}

// NewStockTicker creates a stock ticker with the given symbol, price, and change.
func NewStockTicker(symbol string, price, change float64) *StockTicker {
	st := &StockTicker{
		symbol:  symbol,
		price:   price,
		change:  change,
		style:   DefaultStockTickerStyle(),
		showPct: true,
	}
	st.prevClose = price - change
	st.SetID(GenerateID("stock"))
	return st
}

// Symbol returns the stock symbol.
func (st *StockTicker) Symbol() string {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.symbol
}

// SetSymbol sets the stock symbol.
func (st *StockTicker) SetSymbol(s string) *StockTicker {
	st.mu.Lock()
	st.symbol = s
	st.mu.Unlock()
	return st
}

// Price returns the current price.
func (st *StockTicker) Price() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.price
}

// SetPrice updates the price and recalculates change from prevClose.
func (st *StockTicker) SetPrice(p float64) *StockTicker {
	st.mu.Lock()
	if st.prevClose > 0 {
		st.change = p - st.prevClose
	}
	st.price = p
	st.mu.Unlock()
	return st
}

// Change returns the absolute price change.
func (st *StockTicker) Change() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.change
}

// PercentChange returns the percentage change from prevClose.
func (st *StockTicker) PercentChange() float64 {
	st.mu.RLock()
	defer st.mu.RUnlock()
	if st.prevClose == 0 {
		return 0
	}
	return st.change / st.prevClose * 100
}

// IsUp returns true if the stock gained value.
func (st *StockTicker) IsUp() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.change > 0
}

// IsDown returns true if the stock lost value.
func (st *StockTicker) IsDown() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.change < 0
}

// SetShowPct toggles percentage display.
func (st *StockTicker) SetShowPct(show bool) *StockTicker {
	st.mu.Lock()
	st.showPct = show
	st.mu.Unlock()
	return st
}

// ShowPct returns whether percentage is shown.
func (st *StockTicker) ShowPct() bool {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.showPct
}

// SetStyle sets the visual style.
func (st *StockTicker) SetStyle(s StockTickerStyle) *StockTicker {
	st.mu.Lock()
	st.style = s
	st.mu.Unlock()
	return st
}

// Style returns the current style.
func (st *StockTicker) Style() StockTickerStyle {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.style
}

// Measure computes the desired size.
func (st *StockTicker) Measure(cs Constraints) Size {
	w := 30 // "SYMBOL  $999.99  ▲99.99 (+99.99%)"
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the stock ticker.
func (st *StockTicker) Paint(buf *buffer.Buffer) {
	st.mu.Lock()
	defer st.mu.Unlock()

	b := st.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	x := b.X

	// Symbol
	for _, r := range st.symbol {
		if x >= b.X+b.W {
			break
		}
		buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: st.style.Symbol.Fg, Bg: st.style.Symbol.Bg, Flags: st.style.Symbol.Flags, Width: 1})
		x++
	}

	// Space
	if x < b.X+b.W {
		buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Width: 1})
		x++
	}

	// Price
	priceStr := formatPrice(st.price)
	for _, r := range priceStr {
		if x >= b.X+b.W {
			break
		}
		buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: st.style.Price.Fg, Bg: st.style.Price.Bg, Width: 1})
		x++
	}

	// Space
	if x < b.X+b.W {
		buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Width: 1})
		x++
	}

	// Change indicator
	var arrow rune
	var changeStyle buffer.Style
	if st.change > 0 {
		arrow = '▲'
		changeStyle = st.style.Up
	} else if st.change < 0 {
		arrow = '▼'
		changeStyle = st.style.Down
	} else {
		arrow = '─'
		changeStyle = st.style.Neutral
	}

	if x < b.X+b.W {
		buf.SetCell(x, b.Y, buffer.Cell{Rune: arrow, Fg: changeStyle.Fg, Bg: changeStyle.Bg, Flags: changeStyle.Flags, Width: 1})
		x++
	}

	// Change value
	changeStr := formatPrice(absFloat64(st.change))
	for _, r := range changeStr {
		if x >= b.X+b.W {
			break
		}
		buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: changeStyle.Fg, Bg: changeStyle.Bg, Flags: changeStyle.Flags, Width: 1})
		x++
	}

	// Percentage
	if st.showPct && x < b.X+b.W {
		// Compute inline to avoid deadlock (Paint holds Lock, PercentChange needs RLock)
		var pct float64
		if st.prevClose != 0 {
			pct = st.change / st.prevClose * 100
		}
		pctStr := formatPct(pct)
		// Prefix with space
		if x < b.X+b.W {
			buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Width: 1})
			x++
		}
		for _, r := range pctStr {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: changeStyle.Fg, Bg: changeStyle.Bg, Flags: changeStyle.Flags, Width: 1})
			x++
		}
	}
}

// formatPrice formats a price with 2 decimals.
func formatPrice(v float64) string {
	return strconv.FormatFloat(v, 'f', 2, 64)
}

// formatPct formats a percentage with sign.
func formatPct(v float64) string {
	sign := "+"
	if v < 0 {
		sign = ""
	}
	return sign + strconv.FormatFloat(v, 'f', 2, 64) + "%"
}

// Children returns nil.
func (st *StockTicker) Children() []Component { return nil }
