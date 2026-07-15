// Package glamour provides a compatibility layer for charm.land/glamour/v2.
// It wraps fluui's markdown renderer to provide a glamour-compatible API.
package glamour

import (
	"strings"
	"sync"

	"github.com/topcheer/fluui/compat/glamour/ansi"
	"github.com/topcheer/fluui/markdown"
)

// TermRenderer renders markdown to terminal-styled text.
type TermRenderer struct {
	mu       sync.Mutex
	renderer *markdown.MarkdownRenderer
	width    int
	styles   ansi.StyleConfig
}

// StyleConfig is an alias for ansi.StyleConfig (glamour v2 compatible).
// ggcode passes ansi.StyleConfig to WithStyles.
type StyleConfig = ansi.StyleConfig

// Style is a passthrough for styling (lipgloss.Style).
type Style = struct {
	Foreground string
	Background string
	Bold       bool
	Italic     bool
}

// TermRendererOption configures a TermRenderer.
type TermRendererOption func(*TermRenderer)

// WithStyles sets the style configuration.
func WithStyles(styles ansi.StyleConfig) TermRendererOption {
	return func(r *TermRenderer) {
		r.styles = styles
	}
}

// WithWordWrap sets the word wrap width.
func WithWordWrap(width int) TermRendererOption {
	return func(r *TermRenderer) {
		r.width = width
	}
}

// NewTermRenderer creates a new markdown renderer.
func NewTermRenderer(opts ...TermRendererOption) (*TermRenderer, error) {
	r := &TermRenderer{
		width: 80,
	}
	for _, opt := range opts {
		opt(r)
	}
	r.renderer = markdown.NewMarkdownRenderer(markdown.DefaultTheme(), r.width)
	return r, nil
}

// Render renders markdown text to styled terminal output.
func (r *TermRenderer) Render(md string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	blocks, err := r.renderer.Render(md)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	for _, blk := range blocks {
		for _, cellLine := range blk.Cells {
			for _, cell := range cellLine {
				sb.WriteRune(cell.Rune)
			}
			sb.WriteByte('\n')
		}
	}
	return sb.String(), nil
}

// DefaultStyleConfig returns the default (dark) style configuration.
func DefaultStyleConfig() ansi.StyleConfig {
	return ansi.StyleConfig{}
}