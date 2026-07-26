package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// TokenPricing holds per-million-token rates for cost estimation.
type TokenPricing struct {
	InputPerMillion  float64 // USD per 1M input tokens
	OutputPerMillion float64 // USD per 1M output tokens
}

// DefaultPricing returns common model pricing. Caller should override.
var DefaultPricing = TokenPricing{
	InputPerMillion:  3.00,
	OutputPerMillion: 15.00,
}

// TokenUsageWidget renders token consumption and context window usage
// for AI chat interfaces. Shows input/output token counts, estimated cost,
// model name, and a context window progress bar.
//
// Thread-safe.
type TokenUsageWidget struct {
	BaseComponent
	mu sync.Mutex

	model     string
	inputTok  int
	outputTok int
	ctxUsed   int
	ctxTotal  int
	pricing   TokenPricing
}

// NewTokenUsageWidget creates a token usage widget with default pricing.
func NewTokenUsageWidget(model string) *TokenUsageWidget {
	return &TokenUsageWidget{
		BaseComponent: BaseComponent{id: GenerateID("tokenusage")},
		model:         model,
		pricing:       DefaultPricing,
	}
}

// Model returns the model name.
func (w *TokenUsageWidget) Model() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.model
}

// SetModel sets the model name.
func (w *TokenUsageWidget) SetModel(m string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.model = m
}

// AddTokens adds input and output token counts.
func (w *TokenUsageWidget) AddTokens(input, output int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.inputTok += input
	w.outputTok += output
}

// SetTokens sets the exact input and output token counts.
func (w *TokenUsageWidget) SetTokens(input, output int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.inputTok = input
	w.outputTok = output
}

// InputTokens returns the input token count.
func (w *TokenUsageWidget) InputTokens() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.inputTok
}

// OutputTokens returns the output token count.
func (w *TokenUsageWidget) OutputTokens() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.outputTok
}

// TotalTokens returns input + output tokens.
func (w *TokenUsageWidget) TotalTokens() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.inputTok + w.outputTok
}

// SetContextUsage sets the context window used and total.
func (w *TokenUsageWidget) SetContextUsage(used, total int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ctxUsed = used
	w.ctxTotal = total
}

// ContextUsed returns the context window tokens used.
func (w *TokenUsageWidget) ContextUsed() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ctxUsed
}

// ContextTotal returns the context window total capacity.
func (w *TokenUsageWidget) ContextTotal() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ctxTotal
}

// ContextPercent returns the context window usage as a percentage (0-100).
func (w *TokenUsageWidget) ContextPercent() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ctxPercentLocked()
}

func (w *TokenUsageWidget) ctxPercentLocked() float64 {
	if w.ctxTotal <= 0 {
		return 0
	}
	pct := float64(w.ctxUsed) / float64(w.ctxTotal) * 100
	if pct > 100 {
		pct = 100
	}
	if pct < 0 {
		pct = 0
	}
	return pct
}

// SetPricing sets the token pricing for cost estimation.
func (w *TokenUsageWidget) SetPricing(p TokenPricing) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pricing = p
}

// EstimatedCost returns the estimated cost in USD.
func (w *TokenUsageWidget) EstimatedCost() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.costLocked()
}

func (w *TokenUsageWidget) costLocked() float64 {
	inputCost := float64(w.inputTok) / 1_000_000 * w.pricing.InputPerMillion
	outputCost := float64(w.outputTok) / 1_000_000 * w.pricing.OutputPerMillion
	return inputCost + outputCost
}

// Reset zeroes all token counts and context usage.
func (w *TokenUsageWidget) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.inputTok = 0
	w.outputTok = 0
	w.ctxUsed = 0
}

// Measure returns the desired size. Always 1 line tall, full width.
func (w *TokenUsageWidget) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 80
	}
	return Size{W: maxW, H: 1}
}

// Paint renders the token usage widget as a single-line status.
func (w *TokenUsageWidget) Paint(buf *buffer.Buffer) {
	w.mu.Lock()
	defer w.mu.Unlock()

	bounds := w.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	muted := buffer.Style{Fg: theme.Get().Muted}

	// Draw directly to buffer piece by piece (zero allocation)
	model := w.model
	if model == "" {
		model = "unknown"
	}
	x := bounds.X
	maxX := bounds.X + bounds.W

	x = buf.DrawText(x, bounds.Y, model, muted)
	if x >= maxX {
		return
	}
	x = buf.DrawText(x, bounds.Y, "  \u2191", muted) // ↑
	if x >= maxX {
		return
	}

	// Token counts via stack buffer → string (1 alloc total, shared)
	var tb [32]byte
	tbs := tb[:0]
	tbs = appendTokenCount(tbs, w.inputTok)
	x = buf.DrawText(x, bounds.Y, string(tbs), muted)
	if x >= maxX {
		return
	}

	x = buf.DrawText(x, bounds.Y, " \u2193", muted) // ↓
	if x >= maxX {
		return
	}

	var tb2 [32]byte
	tbs2 := tb2[:0]
	tbs2 = appendTokenCount(tbs2, w.outputTok)
	x = buf.DrawText(x, bounds.Y, string(tbs2), muted)
	if x >= maxX {
		return
	}

	x = buf.DrawText(x, bounds.Y, "  $", muted)
	if x >= maxX {
		return
	}

	var cb [16]byte
	cbs := cb[:0]
	cbs = appendCost(cbs, w.costLocked())
	x = buf.DrawText(x, bounds.Y, string(cbs), muted)

	// Context bar
	if w.ctxTotal > 0 {
		if x+2 < maxX {
			x = buf.DrawText(x, bounds.Y, "  ", muted)
		}
		pct := w.ctxPercentLocked()
		// Draw progress bar via direct SetCell
		barW := 8
		filled := int(pct / 100 * float64(barW))
		if filled > barW {
			filled = barW
		}
		for i := 0; i < barW && x < maxX; i++ {
			if i < filled {
				buf.DrawText(x, bounds.Y, "\u2593", muted) // ▓
			} else {
				buf.DrawText(x, bounds.Y, "\u2591", muted) // ░
			}
			x++
		}
		if x < maxX {
			buf.DrawText(x, bounds.Y, " ", muted)
			x++
		}
		if x < maxX {
			var pb [8]byte
			pbs := pb[:0]
			pbs = strconv.AppendFloat(pbs, pct, 'f', 0, 64)
			pbs = append(pbs, '%')
			buf.DrawText(x, bounds.Y, string(pbs), muted)
		}
	}
}

// appendTokenCount appends a compact token count to b (e.g., "1.2k", "3M").
func appendTokenCount(b []byte, n int) []byte {
	if n < 1000 {
		return strconv.AppendInt(b, int64(n), 10)
	}
	if n < 1_000_000 {
		b = strconv.AppendFloat(b, float64(n)/1000, 'f', 1, 64)
		b = append(b, 'k')
		return b
	}
	b = strconv.AppendFloat(b, float64(n)/1_000_000, 'f', 1, 64)
	b = append(b, 'M')
	return b
}

// appendCost appends a USD cost to b.
func appendCost(b []byte, c float64) []byte {
	if c < 0.01 {
		return strconv.AppendFloat(b, c, 'f', 4, 64)
	}
	return strconv.AppendFloat(b, c, 'f', 2, 64)
}

// appendProgressBar appends a text progress bar to b.
func appendProgressBar(b []byte, pct float64, width int) []byte {
	if width < 1 {
		width = 1
	}
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	for i := 0; i < filled; i++ {
		b = append(b, "\u2593"...) // ▓
	}
	for i := filled; i < width; i++ {
		b = append(b, "\u2591"...) // ░
	}
	return b
}
