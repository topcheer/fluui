package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ModelBadge: AI Model Identifier Badge ───
//
// ModelBadge displays which AI model generated a response, with provider-specific
// color coding. It's the AI-native equivalent of a "verified" badge — users can
// instantly see which model produced the content.
//
// Usage:
//
//	mb := NewModelBadge("gpt-4o")
//	mb.SetContextWindow(128000)
//	mb.Paint(buf) // renders "◆ GPT-4o" in OpenAI green
//
//	mb2 := NewModelBadge("claude-sonnet-4-20250514")
//	// auto-detects Anthropic provider, renders in orange

// AIProvider identifies the AI model provider.
type AIProvider int

const (
	ProviderUnknown AIProvider = iota
	ProviderOpenAI
	ProviderAnthropic
	ProviderGoogle
	ProviderMistral
	ProviderMeta
	ProviderCohere
	ProviderLocal
)

// providerInfo holds display metadata for each provider.
type providerInfo struct {
	Name   string
	Icon   string
	Fg     buffer.Color
	Bg     buffer.Color
}

// providerRegistry maps provider to display info (stack-allocated lookups).
var providerRegistry = [...]providerInfo{
	ProviderUnknown:   {Name: "Unknown", Icon: "?", Fg: buffer.RGB(150, 150, 150), Bg: buffer.RGB(40, 40, 40)},
	ProviderOpenAI:    {Name: "OpenAI", Icon: "◆", Fg: buffer.RGB(16, 163, 127), Bg: buffer.RGB(10, 30, 25)},
	ProviderAnthropic: {Name: "Anthropic", Icon: "✦", Fg: buffer.RGB(217, 119, 87), Bg: buffer.RGB(40, 25, 15)},
	ProviderGoogle:    {Name: "Google", Icon: "✧", Fg: buffer.RGB(66, 133, 244), Bg: buffer.RGB(15, 25, 45)},
	ProviderMistral:   {Name: "Mistral", Icon: "▲", Fg: buffer.RGB(255, 175, 64), Bg: buffer.RGB(40, 30, 15)},
	ProviderMeta:      {Name: "Meta", Icon: "∞", Fg: buffer.RGB(8, 143, 214), Bg: buffer.RGB(10, 25, 40)},
	ProviderCohere:    {Name: "Cohere", Icon: "◈", Fg: buffer.RGB(57, 192, 237), Bg: buffer.RGB(10, 30, 40)},
	ProviderLocal:     {Name: "Local", Icon: "⬡", Fg: buffer.RGB(180, 180, 180), Bg: buffer.RGB(30, 30, 30)},
}

// modelPrefixMap maps model name prefixes to providers (checked in order).
var modelPrefixMap = []struct {
	prefix  string
	prov    AIProvider
	display string
}{
	{"gpt-", ProviderOpenAI, "GPT"},
	{"o1", ProviderOpenAI, "o1"},
	{"o3", ProviderOpenAI, "o3"},
	{"o4", ProviderOpenAI, "o4"},
	{"text-embedding", ProviderOpenAI, "Embed"},
	{"claude-", ProviderAnthropic, "Claude"},
	{"gemini-", ProviderGoogle, "Gemini"},
	{"mistral-", ProviderMistral, "Mistral"},
	{"mixtral-", ProviderMistral, "Mixtral"},
	{"codestral", ProviderMistral, "Codestral"},
	{"llama-", ProviderMeta, "Llama"},
	{"llama3", ProviderMeta, "Llama3"},
	{"command-r", ProviderCohere, "Command"},
	{"qwen", ProviderLocal, "Qwen"},
	{"deepseek", ProviderLocal, "DeepSeek"},
	{"phi-", ProviderLocal, "Phi"},
	{"ollama/", ProviderLocal, ""},
	{"local", ProviderLocal, "Local"},
}

// detectProvider determines the provider and short display name from a model ID.
func detectProvider(modelID string) (AIProvider, string) {
	id := toLowerASCII(modelID)
	for _, entry := range modelPrefixMap {
		if startsWith(id, entry.prefix) {
			display := entry.display
			if display == "" {
				display = providerRegistry[entry.prov].Name
			}
			return entry.prov, display
		}
	}
	return ProviderUnknown, modelID
}

// toLowerASCII converts a string to lowercase (ASCII only, zero-alloc).
func toLowerASCII(s string) string {
	b := []byte(s)
	for i := 0; i < len(b); i++ {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

// startsWith checks if s starts with prefix (zero-alloc).
func startsWith(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

// ModelBadge displays an AI model identifier with provider-specific styling.
type ModelBadge struct {
	BaseComponent
	mu            sync.RWMutex
	modelID       string
	provider      AIProvider
	displayName   string
	contextWindow int    // tokens, 0 = unknown
	customLabel   string // overrides computed label if non-empty
	style         buffer.Style
}

// NewModelBadge creates a model badge for the given model ID.
// The provider is auto-detected from the model name.
func NewModelBadge(modelID string) *ModelBadge {
	prov, display := detectProvider(modelID)
	mb := &ModelBadge{
		modelID:     modelID,
		provider:    prov,
		displayName: display,
		style: buffer.Style{
			Fg: providerRegistry[prov].Fg,
			Bg: providerRegistry[prov].Bg,
			Flags: buffer.Bold,
		},
	}
	mb.SetID(GenerateID("modelbadge"))
	return mb
}

// ModelID returns the full model identifier.
func (mb *ModelBadge) ModelID() string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.modelID
}

// SetModelID changes the displayed model.
func (mb *ModelBadge) SetModelID(id string) *ModelBadge {
	mb.mu.Lock()
	mb.modelID = id
	mb.provider, mb.displayName = detectProvider(id)
	mb.style.Fg = providerRegistry[mb.provider].Fg
	mb.style.Bg = providerRegistry[mb.provider].Bg
	mb.mu.Unlock()
	return mb
}

// Provider returns the detected provider.
func (mb *ModelBadge) Provider() AIProvider {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.provider
}

// DisplayName returns the short display name (e.g., "GPT", "Claude").
func (mb *ModelBadge) DisplayName() string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.displayName
}

// SetCustomLabel overrides the computed display label.
func (mb *ModelBadge) SetCustomLabel(label string) *ModelBadge {
	mb.mu.Lock()
	mb.customLabel = label
	mb.mu.Unlock()
	return mb
}

// ContextWindow returns the context window size in tokens (0 = unknown).
func (mb *ModelBadge) ContextWindow() int {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.contextWindow
}

// SetContextWindow sets the context window size in tokens.
func (mb *ModelBadge) SetContextWindow(tokens int) *ModelBadge {
	mb.mu.Lock()
	mb.contextWindow = tokens
	mb.mu.Unlock()
	return mb
}

// SetStyle overrides the default style.
func (mb *ModelBadge) SetStyle(s buffer.Style) *ModelBadge {
	mb.mu.Lock()
	mb.style = s
	mb.mu.Unlock()
	return mb
}

// Style returns the current style.
func (mb *ModelBadge) Style() buffer.Style {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return mb.style
}

// ProviderName returns the human-readable provider name.
func (mb *ModelBadge) ProviderName() string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return providerRegistry[mb.provider].Name
}

// ProviderIcon returns the provider icon character.
func (mb *ModelBadge) ProviderIcon() string {
	mb.mu.RLock()
	defer mb.mu.RUnlock()
	return providerRegistry[mb.provider].Icon
}

// labelLocked computes the display label (caller holds lock).
func (mb *ModelBadge) labelLocked() string {
	if mb.customLabel != "" {
		return mb.customLabel
	}
	return mb.displayName
}

// Measure computes the desired size.
func (mb *ModelBadge) Measure(cs Constraints) Size {
	mb.mu.RLock()
	label := mb.labelLocked()
	mb.mu.RUnlock()
	w := buffer.StringWidth(label) + 4 // icon + space + label + padding
	if w < 6 {
		w = 6
	}
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the model badge into the buffer.
func (mb *ModelBadge) Paint(buf *buffer.Buffer) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	b := mb.bounds
	if b.W < 4 || b.H < 1 {
		return
	}

	icon := providerRegistry[mb.provider].Icon
	label := mb.labelLocked()

	// Draw: " icon label "
	x := b.X
	// Padding space
	buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Bg: mb.style.Bg})
	x++
	// Icon
	buf.SetCell(x, b.Y, buffer.Cell{Rune: []rune(icon)[0], Fg: mb.style.Fg, Bg: mb.style.Bg, Flags: mb.style.Flags})
	x++
	// Space
	buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Bg: mb.style.Bg})
	x++
	// Label
	maxW := b.X + b.W - x - 1 // leave 1 for trailing padding
	labelW := buffer.StringWidth(label)
	if labelW > maxW {
		// Truncate label to fit
		runed := []rune(label)
		if maxW > 1 {
			label = string(runed[:maxW-1]) + "…"
		} else if maxW == 1 {
			label = "…"
		} else {
			label = ""
		}
	}
	if label != "" {
		for _, r := range label {
			if x >= b.X+b.W-1 {
				break
			}
			buf.SetCell(x, b.Y, buffer.Cell{Rune: r, Fg: mb.style.Fg, Bg: mb.style.Bg, Flags: mb.style.Flags})
			x++
		}
	}
	// Fill remaining with bg
	for ; x < b.X+b.W; x++ {
		buf.SetCell(x, b.Y, buffer.Cell{Rune: ' ', Bg: mb.style.Bg})
	}
}

// Children returns nil (leaf component).
func (mb *ModelBadge) Children() []Component {
	return nil
}
