package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIPanelHeader: AI Chat Panel Header Bar ───
//
// AIPanelHeader renders a compact header for AI chat panels showing model name,
// provider badge, token usage (used/limit), and a live status indicator.
//
// Usage:
//
//	h := NewAIPanelHeader()
//	h.SetModel("GPT-4o")
//	h.SetProvider("OpenAI")
//	h.SetTokenUsage(5000, 128000)
//	h.SetStatus("streaming")
//	h.Paint(buf)

// AIHeaderStatus describes the AI panel status.
type AIHeaderStatus int

const (
	AIStatusIdle      AIHeaderStatus = iota
	AIStatusThinking
	AIStatusStreaming
	AIStatusError
)

// AIPanelHeaderStyle holds styling.
type AIPanelHeaderStyle struct {
	Model     buffer.Style
	Provider  buffer.Style
	Tokens    buffer.Style
	Status    [4]buffer.Style // [idle, thinking, streaming, error]
	Border    buffer.Style
}

// DefaultAIPanelHeaderStyle returns defaults.
func DefaultAIPanelHeaderStyle() AIPanelHeaderStyle {
	model := buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold}
	prov := buffer.Style{Fg: buffer.RGB(96, 165, 250)}
	tokens := buffer.Style{Fg: buffer.RGB(148, 163, 184)}
	idle := buffer.Style{Fg: buffer.RGB(100, 116, 139)}
	thinking := buffer.Style{Fg: buffer.RGB(234, 179, 8)}
	streaming := buffer.Style{Fg: buffer.RGB(34, 197, 94)}
	errS := buffer.Style{Fg: buffer.RGB(239, 68, 68)}
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}
	return AIPanelHeaderStyle{Model: model, Provider: prov, Tokens: tokens, Status: [4]buffer.Style{idle, thinking, streaming, errS}, Border: border}
}

// aiStatusIcon returns the icon rune for a status.
func aiStatusIcon(s AIHeaderStatus) rune {
	switch s {
	case AIStatusIdle: return '●'
	case AIStatusThinking: return '◐'
	case AIStatusStreaming: return '◉'
	case AIStatusError: return '✗'
	default: return '●'
	}
}

// aiStatusText returns text for a status.
func aiStatusText(s AIHeaderStatus) string {
	switch s {
	case AIStatusIdle: return "idle"
	case AIStatusThinking: return "thinking"
	case AIStatusStreaming: return "streaming"
	case AIStatusError: return "error"
	default: return "idle"
	}
}

// AIPanelHeader renders an AI chat panel header.
type AIPanelHeader struct {
	BaseComponent
	mu sync.Mutex

	model    string
	provider string
	used     int
	limit    int
	status   AIHeaderStatus
	style    AIPanelHeaderStyle
	// cached display strings
	tokenStr string
}

// NewAIPanelHeader creates an AIPanelHeader.
func NewAIPanelHeader() *AIPanelHeader {
	h := &AIPanelHeader{model: "GPT-4", provider: "OpenAI", limit: 8192, style: DefaultAIPanelHeaderStyle()}
	h.SetID(GenerateID("aiheader"))
	h.tokenStr = "0/8192"
	return h
}

// SetModel sets the model name.
func (h *AIPanelHeader) SetModel(m string) *AIPanelHeader {
	h.mu.Lock()
	h.model = m
	h.mu.Unlock()
	return h
}

// Model returns the model name.
func (h *AIPanelHeader) Model() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.model
}

// SetProvider sets the provider name.
func (h *AIPanelHeader) SetProvider(p string) *AIPanelHeader {
	h.mu.Lock()
	h.provider = p
	h.mu.Unlock()
	return h
}

// Provider returns the provider name.
func (h *AIPanelHeader) Provider() string {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.provider
}

// SetTokenUsage sets token used/limit and caches display string.
func (h *AIPanelHeader) SetTokenUsage(used, limit int) *AIPanelHeader {
	h.mu.Lock()
	h.used = used
	h.limit = limit
	h.tokenStr = itoa(used) + "/" + itoa(limit)
	h.mu.Unlock()
	return h
}

// TokenUsage returns used and limit tokens.
func (h *AIPanelHeader) TokenUsage() (int, int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.used, h.limit
}

// SetStatus sets the AI status.
func (h *AIPanelHeader) SetStatus(s AIHeaderStatus) *AIPanelHeader {
	h.mu.Lock()
	h.status = s
	h.mu.Unlock()
	return h
}

// Status returns the current status.
func (h *AIPanelHeader) Status() AIHeaderStatus {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.status
}

// SetStyle sets custom style.
func (h *AIPanelHeader) SetStyle(s AIPanelHeaderStyle) *AIPanelHeader {
	h.mu.Lock()
	h.style = s
	h.mu.Unlock()
	return h
}

// Measure returns the preferred size.
func (h *AIPanelHeader) Measure(cs Constraints) Size {
	w := 50
	h2 := 1
	if cs.MaxWidth > 0 && w > cs.MaxWidth { w = cs.MaxWidth }
	return Size{W: w, H: h2}
}

// Paint renders the AI panel header into the buffer.
func (h *AIPanelHeader) Paint(buf *buffer.Buffer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	b := h.Bounds()
	x, y := b.X, b.Y

	statusIdx := int(h.status)
	if statusIdx < 0 || statusIdx > 3 { statusIdx = 0 }
	statusStyle := h.style.Status[statusIdx]
	modelStyle := h.style.Model
	provStyle := h.style.Provider
	tokenStyle := h.style.Tokens

	col := x

	// Status icon
	if col >= buf.Width { return }
	buf.SetCell(col, y, buffer.Cell{Rune: aiStatusIcon(h.status), Fg: statusStyle.Fg, Bg: statusStyle.Bg, Flags: statusStyle.Flags, Width: 1})
	col++
	if col >= buf.Width { return }
	buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: statusStyle.Fg, Bg: statusStyle.Bg, Flags: statusStyle.Flags, Width: 1})
	col++

	// Model name
	for _, r := range h.model {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: modelStyle.Fg, Bg: modelStyle.Bg, Flags: modelStyle.Flags, Width: 1})
		col++
	}
	if col >= buf.Width { return }
	buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: modelStyle.Fg, Bg: modelStyle.Bg, Flags: modelStyle.Flags, Width: 1})
	col++

	// Provider badge
	for _, r := range h.provider {
		if col >= buf.Width { return }
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: provStyle.Fg, Bg: provStyle.Bg, Flags: provStyle.Flags, Width: 1})
		col++
	}
	if col >= buf.Width { return }
	buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: provStyle.Fg, Bg: provStyle.Bg, Flags: provStyle.Flags, Width: 1})
	col++

	// Token usage (right-aligned)
	tokenLen := len(h.tokenStr)
	tokenStart := b.X + b.W - tokenLen
	if tokenStart < col { tokenStart = col }
	// Fill gap with spaces
	for c := col; c < tokenStart; c++ {
		if c >= buf.Width { break }
		buf.SetCell(c, y, buffer.Cell{Rune: ' ', Fg: tokenStyle.Fg, Bg: tokenStyle.Bg, Flags: tokenStyle.Flags, Width: 1})
	}
	for i, r := range h.tokenStr {
		cx := tokenStart + i
		if cx >= buf.Width { break }
		buf.SetCell(cx, y, buffer.Cell{Rune: r, Fg: tokenStyle.Fg, Bg: tokenStyle.Bg, Flags: tokenStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (h *AIPanelHeader) Children() []Component { return nil }
