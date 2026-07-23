package component

import (
	"strconv"
	"sync"
	"unicode/utf8"

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

	// Build: "model  ↑1.2k ↓800  $0.024  ▓▓▓░░░░ 45%"
	text := w.buildLineLocked(bounds.W)
	muted := buffer.Style{Fg: theme.Get().Muted}
	buf.DrawText(bounds.X, bounds.Y, text, muted)
}

// buildLineLocked constructs the display string.
func (w *TokenUsageWidget) buildLineLocked(maxW int) string {
	model := w.model
	if model == "" {
		model = "unknown"
	}

	inStr := formatTokenCount(w.inputTok)
	outStr := formatTokenCount(w.outputTok)
	costStr := formatCost(w.costLocked())

	// Context bar
	ctxBar := ""
	if w.ctxTotal > 0 {
		pct := w.ctxPercentLocked()
		ctxBar = "  " + buildProgressBar(pct, 8) + " " + strconv.FormatFloat(pct, 'f', 0, 64) + "%"
	}

	line := model + "  ↑" + inStr + " ↓" + outStr + "  " + costStr + ctxBar

	// Clamp to width using utf8.RuneCountInString (avoids []rune allocation)
	if utf8.RuneCountInString(line) > maxW {
		if maxW > 1 {
			return truncateStr(line, maxW-1) + "…"
		}
		return "…"
	}
	return line
}

// formatTokenCount formats a token count for compact display.
// Uses strconv.AppendFloat to avoid fmt.Sprintf allocations.
func formatTokenCount(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}
	var buf [16]byte
	if n < 1_000_000 {
		b := strconv.AppendFloat(buf[:0], float64(n)/1000, 'f', 1, 64)
		return string(b) + "k"
	}
	b := strconv.AppendFloat(buf[:0], float64(n)/1_000_000, 'f', 1, 64)
	return string(b) + "M"
}

// formatCost formats a USD cost for display.
func formatCost(c float64) string {
	var buf [32]byte
	if c < 0.01 {
		b := strconv.AppendFloat(buf[:0], c, 'f', 4, 64)
		return "$" + string(b)
	}
	b := strconv.AppendFloat(buf[:0], c, 'f', 2, 64)
	return "$" + string(b)
}

// buildProgressBar creates a text progress bar of given width (0-100 pct).
func buildProgressBar(pct float64, width int) string {
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
	var buf [64]byte
	b := buf[:0]
	for i := 0; i < filled; i++ {
		b = append(b, "▓"...)
	}
	for i := filled; i < width; i++ {
		b = append(b, "░"...)
	}
	return string(b)
}
