package app

import (
	"errors"
	"io"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/internal/term"
	"github.com/topcheer/fluui/internal/termcompat"
)

// --- ClipboardConfig ---

// ErrClipboardNotSupported is returned when the terminal does not support
// OSC52 clipboard operations.
var ErrClipboardNotSupported = errors.New("clipboard not supported: OSC52 unavailable on this terminal")

// ClipboardConfig wraps OSC52 clipboard operations with terminal capability
// awareness. It checks termcompat.Capabilities to decide whether OSC52
// sequences should be used, and wraps them in tmux passthrough when needed.
type ClipboardConfig struct {
	caps termcompat.Capabilities
}

// NewClipboardConfig creates a ClipboardConfig with no capabilities.
// OSC52 will not be usable until SetCapabilities is called.
func NewClipboardConfig() *ClipboardConfig {
	return &ClipboardConfig{}
}

// NewClipboardWithCapabilities creates a ClipboardConfig pre-configured
// from detected terminal capabilities.
func NewClipboardWithCapabilities(caps termcompat.Capabilities) *ClipboardConfig {
	return &ClipboardConfig{caps: caps}
}

// Capabilities returns the current terminal capabilities.
func (c *ClipboardConfig) Capabilities() termcompat.Capabilities {
	return c.caps
}

// SetCapabilities updates the terminal capabilities.
func (c *ClipboardConfig) SetCapabilities(caps termcompat.Capabilities) {
	c.caps = caps
}

// CanCopy returns true if OSC52 clipboard operations are available.
func (c *ClipboardConfig) CanCopy() bool {
	return c.caps.ShouldUseOSC52()
}

// Copy generates the appropriate OSC52 escape sequence for the given text.
// If inside tmux, wraps the sequence in tmux passthrough.
// Returns ErrClipboardNotSupported if OSC52 is not available.
func (c *ClipboardConfig) Copy(text string) (string, error) {
	if !c.CanCopy() {
		return "", ErrClipboardNotSupported
	}
	seq := term.CopyOSC52(text)
	if c.caps.InsideTmux {
		return wrapTmuxPassthrough(seq), nil
	}
	return seq, nil
}

// CopyToWriter generates the OSC52 sequence and writes it to w.
// Returns ErrClipboardNotSupported if OSC52 is not available.
func (c *ClipboardConfig) CopyToWriter(w io.Writer, text string) error {
	seq, err := c.Copy(text)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(seq))
	return err
}

// wrapTmuxPassthrough wraps a string in tmux DCS passthrough sequences.
// tmux uses: ESC P tmux ; ESC <content with ESC doubled> ESC \ ESC \
func wrapTmuxPassthrough(s string) string {
	// Double all ESC bytes for tmux passthrough.
	doubled := make([]byte, 0, len(s)*2)
	for i := 0; i < len(s); i++ {
		if s[i] == 0x1b { // ESC
			doubled = append(doubled, 0x1b, 0x1b)
		} else {
			doubled = append(doubled, s[i])
		}
	}
	return "\x1bPtmux;" + string(doubled) + "\x1b\\\x1b\\"
}

// --- ChatApp clipboard integration ---

// SetClipboardCapabilities configures clipboard operations from terminal capabilities.
func (a *ChatApp) SetClipboardCapabilities(caps termcompat.Capabilities) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clipboardConfig = NewClipboardWithCapabilities(caps)
}

// ClipboardConfig returns the current clipboard configuration.
// Returns nil if no capabilities have been set.
func (a *ChatApp) ClipboardConfig() *ClipboardConfig {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clipboardConfig
}

// CopyLastBlockClipboard finds the last content block, generates a
// capability-aware OSC52 sequence, and writes it to w.
// Uses ClipboardConfig if available, otherwise falls back to raw OSC52.
func (a *ChatApp) CopyLastBlockClipboard(w io.Writer) error {
	text, ok := a.LastBlockText()
	if !ok || text == "" {
		return errors.New("no content to copy")
	}

	a.mu.Lock()
	cc := a.clipboardConfig
	a.mu.Unlock()

	if cc != nil {
		return cc.CopyToWriter(w, text)
	}

	// Fallback: raw OSC52 without capability checking.
	_, _ = w.Write([]byte(term.CopyOSC52(text)))
	return nil
}

// CopyLastBlockOSC52 is the legacy entry point for copy keybindings.
// It finds the last content block, generates a raw OSC52 sequence, and
// writes it to w without capability checking. Prefer CopyLastBlockClipboard.
// Returns true if a block was found and copied, false if no content.
func (a *ChatApp) CopyLastBlockOSC52(w io.Writer) bool {
	text, ok := a.LastBlockText()
	if !ok || text == "" {
		return false
	}
	_, _ = w.Write([]byte(term.CopyOSC52(text)))
	return true
}

// extractBlockText attempts to extract meaningful text from a block
// using type assertions, since the Block interface does not expose Content().
func extractBlockText(b block.Block) (string, bool) {
	switch v := b.(type) {
	case *block.AssistantTextBlock:
		return v.Content(), true
	case *block.UserMessageBlock:
		return v.Content(), true
	case *block.ThinkingBlock:
		return v.Content(), true
	case *block.ToolResultBlock:
		return v.Output(), true
	case *block.ToolCallBlock:
		return v.ToolName() + "(" + v.RawArgs() + ")", true
	default:
		return "", false
	}
}
