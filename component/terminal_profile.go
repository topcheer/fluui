package component

import (
	"os"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// TerminalProfile displays detected terminal capabilities (TERM, TERM_PROGRAM,
// COLORTERM, color support, protocol support). Useful as a debug panel
// or a first-run diagnostics screen to verify terminal compatibility.
//
// Thread-safe.
type TerminalProfile struct {
	BaseComponent
	mu sync.Mutex
}

// NewTerminalProfile creates a terminal profile panel.
func NewTerminalProfile() *TerminalProfile {
	return &TerminalProfile{
		BaseComponent: BaseComponent{id: GenerateID("termprofile")},
	}
}

// envOr returns the env var or "-" if unset.
func envOr(key string) string {
	v := os.Getenv(key)
	if v == "" {
		return "-"
	}
	return v
}

// hasColorSupport detects color capability from COLORTERM.
func hasColorSupport() string {
	ct := os.Getenv("COLORTERM")
	if ct == "truecolor" || ct == "24bit" {
		return "24-bit (TrueColor)"
	}
	term := os.Getenv("TERM")
	if strings.Contains(term, "256color") || strings.Contains(term, "256") {
		return "256-color"
	}
	return "16-color"
}

// Measure returns the desired size.
func (t *TerminalProfile) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 50
	}
	return Size{W: maxW, H: 8}
}

// Paint renders the terminal profile.
func (t *TerminalProfile) Paint(buf *buffer.Buffer) {
	t.mu.Lock()
	defer t.mu.Unlock()

	b := t.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	labelStyle := buffer.Style{Fg: th.Muted}
	valueStyle := buffer.Style{Fg: th.Fg}
	okStyle := buffer.Style{Fg: th.Success}

	entries := []struct {
		label string
		value string
	}{
		{"TERM", envOr("TERM")},
		{"TERM_PROGRAM", envOr("TERM_PROGRAM")},
		{"COLORTERM", envOr("COLORTERM")},
		{"Color Support", hasColorSupport()},
		{"Shell", envOr("SHELL")},
		{"LANG", envOr("LANG")},
	}

	y := b.Y
	for _, e := range entries {
		if y >= b.Y+b.H {
			break
		}

		x := b.X
		// Label
		buf.DrawText(x, y, e.label+":", labelStyle)
		x += 20
		if x > b.X+b.W {
			x = b.X + b.W - 1
		}
		// Value
		val := e.value
		if val == "-" {
			buf.DrawText(x, y, val, labelStyle)
		} else {
			buf.DrawText(x, y, val, valueStyle)
		}
		y++
	}

	// Status line
	if y < b.Y+b.H {
		buf.DrawText(b.X, y, "\u2714 Terminal profile detected", okStyle)
	}
}
