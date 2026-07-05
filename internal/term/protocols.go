package term

import (
	"encoding/base64"
	"math"
	"strings"
)

// Terminal protocol escape sequence helpers.
//
// This file provides generator functions for four modern terminal protocols
// that are widely supported by contemporary terminal emulators but were not
// yet exposed by the fluui/term package:
//
//   - OSC 8  — clickable hyperlinks (xterm, iTerm2, Kitty, WezTerm, GNOME, Windows Terminal)
//   - Sync    — synchronized output / BPS / DCS to reduce tearing on large updates (Kitty, WezTerm, Alacritty)
//   - Focus   — FocusIn / FocusOut reporting (xterm CSI ?1004 h/l)
//   - Title   — SetWindowTitle via OSC 0/1/2
//
// Each helper returns a plain string. Callers write the string directly to the
// terminal via Terminal.WriteRaw or through a renderer's output path.

// ---------------------------------------------------------------------------
// OSC 8 — Hyperlinks
// ---------------------------------------------------------------------------

// HyperlinkOptions configures an OSC 8 hyperlink.
type HyperlinkOptions struct {
	// URL is the destination URI (https://, file://, mailto:, etc.).
	// Required for the link to be clickable.
	URL string

	// ID is an optional stable identifier. Cells with the same ID are
	// treated as a single link by some terminals (hover highlights all).
	ID string

	// Params is an optional key=value list appended to the OSC 8 params
	// field (e.g. "icon=emoji"). Rarely used.
	Params string
}

// OSC8Start returns the escape sequence that begins an OSC 8 hyperlink.
// Write the visible text immediately after, then terminate with OSC8End.
func OSC8Start(opts HyperlinkOptions) string {
	var sb strings.Builder
	sb.Grow(32 + len(opts.Params) + len(opts.ID) + len(opts.URL))
	sb.WriteString("\x1b]8;")
	// params field: may contain id=... or key=value pairs
	if opts.Params != "" || opts.ID != "" {
		first := true
		if opts.ID != "" {
			sb.WriteString("id=")
			sb.WriteString(opts.ID)
			first = false
		}
		if opts.Params != "" {
			if !first {
				sb.WriteByte(':')
			}
			sb.WriteString(opts.Params)
		}
	}
	sb.WriteByte(';')
	sb.WriteString(opts.URL)
	sb.WriteString("\x1b\\") // ST
	return sb.String()
}

// OSC8End returns the escape sequence that terminates an OSC 8 hyperlink.
// The URL and params fields are empty to close the link.
func OSC8End() string {
	return "\x1b]8;;\x1b\\"
}

// OSC8Link returns a complete hyperlinked string: the opening escape,
// the visible text, and the closing escape.
func OSC8Link(opts HyperlinkOptions, text string) string {
	return OSC8Start(opts) + text + OSC8End()
}

// ---------------------------------------------------------------------------
// Synchronized Output — BSU/ESU (Begin/End Synchronized Update)
// ---------------------------------------------------------------------------

// Synchronized Output (also known as BPS — Batched Presentation State) groups
// a sequence of screen updates so the terminal applies them atomically. This
// eliminates flicker and tearing during large redraws.
//
// Format (DCS wrapper):
//
//	Begin: DCS $ 2026 t   (ESC P = $ 2026 t)
//	End:   DCS $ 2026 u   (ESC P = $ 2026 u)
//
// Supported by Kitty, WezTerm, Alacritty, foot, Konsole, ghostty.

const (
	// SyncBegin starts a synchronized-update region.
	SyncBegin = "\x1bP=1s\x1b\\"
	// SyncEnd terminates a synchronized-update region.
	SyncEnd = "\x1bP=2s\x1b\\"
)

// Sync wraps the given output string in synchronized-update markers so the
// terminal renders it atomically. If the output is empty, returns an empty
// string without wrapping (avoids unnecessary control sequences).
func Sync(output string) string {
	if output == "" {
		return ""
	}
	var sb strings.Builder
	sb.Grow(len(SyncBegin) + len(output) + len(SyncEnd))
	sb.WriteString(SyncBegin)
	sb.WriteString(output)
	sb.WriteString(SyncEnd)
	return sb.String()
}

// ---------------------------------------------------------------------------
// Focus Tracking — CSI ?1004 h / l
// ---------------------------------------------------------------------------

// Focus reporting (DEC private mode 1004) causes the terminal to emit
// ESC [ I when the window gains focus and ESC [ O when it loses focus.
// The Parser converts these to EventFocus (see input.go).

const (
	// EnableFocus enables focus tracking: the terminal sends focus in/out events.
	EnableFocus = "\x1b[?1004h"
	// DisableFocus disables focus tracking.
	DisableFocus = "\x1b[?1004l"
)

// ---------------------------------------------------------------------------
// Window Title — OSC 0 / 1 / 2
// ---------------------------------------------------------------------------

// SetWindowTitle returns an OSC 2 escape that sets both the window title and
// the icon name. Most terminals treat OSC 0 and OSC 2 identically.
func SetWindowTitle(title string) string {
	return setTitleOSC("2", title)
}

// SetIconName returns an OSC 1 escape that sets only the icon name.
// Most modern terminals ignore the distinction and also update the title.
func SetIconName(title string) string {
	return setTitleOSC("1", title)
}

// SetWindowTitleAndIcon returns an OSC 0 escape that sets both the window
// title and the icon name simultaneously (the most common form).
func SetWindowTitleAndIcon(title string) string {
	return setTitleOSC("0", title)
}

func setTitleOSC(kind, title string) string {
	// Use BEL terminator — universally supported and shorter than ST.
	// OSC payloads must not contain a raw BEL (0x07) byte; escape it if present.
	escaped := title
	if strings.ContainsRune(escaped, '\x07') {
		escaped = strings.ReplaceAll(escaped, "\x07", "")
	}
	var sb strings.Builder
	sb.Grow(6 + len(escaped) + 1)
	sb.WriteString("\x1b]")
	sb.WriteString(kind)
	sb.WriteByte(';')
	sb.WriteString(escaped)
	sb.WriteString("\x07")
	return sb.String()
}

// ---------------------------------------------------------------------------
// Cursor Visibility — DECTCEM
// ---------------------------------------------------------------------------

// HideCursor disables the cursor (DECTCEM reset).
const HideCursor = "\x1b[?25l"

// ShowCursor enables the cursor (DECTCEM set).
const ShowCursor = "\x1b[?25h"

// ---------------------------------------------------------------------------
// Alternate Screen Buffer
// ---------------------------------------------------------------------------

// EnterAltScreen switches to the alternate screen buffer.
const EnterAltScreen = "\x1b[?1049h"

// LeaveAltScreen switches back to the primary screen buffer.
const LeaveAltScreen = "\x1b[?1049l"

// ---------------------------------------------------------------------------
// Bracketed Paste — already parsed by the Parser; expose enable/disable here.
// ---------------------------------------------------------------------------

// EnableBracketedPaste enables bracketed paste mode (CSI ?2004 h).
const EnableBracketedPaste = "\x1b[?2004h"

// DisableBracketedPaste disables bracketed paste mode (CSI ?2004 l).
const DisableBracketedPaste = "\x1b[?2004l"

// ---------------------------------------------------------------------------
// Mouse Tracking — common modes
// ---------------------------------------------------------------------------

// EnableMouseSGR enables SGR mouse mode (CSI ?1006 h). Usually combined with
// button-event (1002) or any-event (1003) tracking.
const EnableMouseSGR = "\x1b[?1006h"

// DisableMouseSGR disables SGR mouse mode.
const DisableMouseSGR = "\x1b[?1006l"

// ---------------------------------------------------------------------------
// 24-bit color enable (not a toggle — always on in capable terminals)
// ---------------------------------------------------------------------------

// EnableTrueColor is a no-op on most terminals that already detected true
// color, but is provided for explicitness. Format: CSI ? 4 ; 1 : rgb m.
// Most applications should rely on the Terminal's ColorProfile instead.
const EnableTrueColor = "\x1b[?4;1$pc"

// ---------------------------------------------------------------------------
// Kitty Keyboard Protocol (CSI > 1 u / CSI < u) — capability query helpers.
// Full Kitty keyboard support is complex; here we expose only the enable and
// disable escape sequences so callers can opt in for advanced key reporting.
// ---------------------------------------------------------------------------

// EnableKittyKeyboard enables the Kitty keyboard protocol (progressive
// enhancement flag 1). After enabling, the Parser will receive CSI u forms.
const EnableKittyKeyboard = "\x1b[>1u"

// DisableKittyKeyboard disables the Kitty keyboard protocol.
const DisableKittyKeyboard = "\x1b[<u"

// ---------------------------------------------------------------------------
// Notification bell —BEL byte.
// ---------------------------------------------------------------------------

// Bell is the BEL control character (0x07).
const Bell = "\x07"

// ---------------------------------------------------------------------------
// QueryWindowTitle — OSC 21 / report response via input stream.
// ---------------------------------------------------------------------------

// QueryWindowTitle returns the OSC 21 escape that asks the terminal to report
// the current window title. The response arrives as ESC ] l <title> ESC \ on
// terminals that support it (xterm, rxvt). Most other terminals ignore it.
func QueryWindowTitle() string {
	return "\x1b]21\x1b\\"
}

// ---------------------------------------------------------------------------
// CopyToClipboard convenience (alias for ClipboardSystem).
// ---------------------------------------------------------------------------

// CopyClipboard is a convenience wrapper around CopyOSC52 for the system
// clipboard. Provided here so callers can import a single protocols file.
func CopyClipboard(text string) string {
	return CopyOSC52Source(text, ClipboardSystem)
}

// CopyPrimary is a convenience wrapper that targets the X11 primary selection.
func CopyPrimary(text string) string {
	return CopyOSC52Source(text, ClipboardPrimary)
}

// ensure base64 import is used even if future refactors drop helpers above.
var _ = base64.StdEncoding

// ---------------------------------------------------------------------------
// OSC 4 / 10 / 11 / 12 — Color Query
// ---------------------------------------------------------------------------
//
// OSC 4 queries a specific palette index (0-255).
// OSC 10 queries the default foreground color.
// OSC 11 queries the default background color.
// OSC 12 queries the text cursor color.
//
// Query format:  ESC ] 4 ; <index> ; ?          ST
//                ESC ] 10 ; ?                   ST
//                ESC ] 11 ; ?                   ST
//
// Response format: ESC ] 4 ; <index> ; rgb:RRRR/GGGG/BBBB  ST
//                  ESC ] 10 ; rgb:RRRR/GGGG/BBBB           ST
//                  ESC ] 11 ; rgb:RRRR/GGGG/BBBB           ST
//
// The R/G/B components may be 1-4 hex digits. They are normalized to 8-bit.

// QueryPaletteColor generates an OSC 4 query for palette index (0-255).
// The terminal responds with the RGB value of that palette slot.
func QueryPaletteColor(index int) string {
	return "\x1b]4;" + colorItoa(index) + ";?\x1b\\"
}

// QueryDefaultFG generates an OSC 10 query for the default foreground color.
func QueryDefaultFG() string {
	return "\x1b]10;?\x1b\\"
}

// QueryDefaultBG generates an OSC 11 query for the default background color.
func QueryDefaultBG() string {
	return "\x1b]11;?\x1b\\"
}

// QueryCursorColor generates an OSC 12 query for the cursor color.
func QueryCursorColor() string {
	return "\x1b]12;?\x1b\\"
}

// ColorResponse holds a parsed color query response.
type ColorResponse struct {
	// Index is the palette index (0-255). -1 for OSC 10/11/12 queries.
	Index int
	// R, G, B are the 8-bit color components (0-255).
	R, G, B uint8
	// Valid is true if parsing succeeded.
	Valid bool
}

// ParseColorResponse parses a terminal's response to an OSC 4/10/11/12 query.
// The input should be the full escape sequence received from the terminal,
// e.g. "\x1b]4;0;rgb:0000/0000/0000\x1b\\" or "\x1b]10;rgb:cccc/cccc/cccc\x1b\\".
// Returns ColorResponse{Valid: false} if parsing fails.
func ParseColorResponse(s string) ColorResponse {
	// Must start with ESC ] (OSC introducer)
	if len(s) < 4 || s[0] != 0x1b || s[1] != ']' {
		return ColorResponse{Index: -1}
	}

	// Find the string terminator (ESC \ or BEL)
	body := s[2:] // skip ESC ]
	for i := 0; i < len(body); i++ {
		if body[i] == 0x1b && i+1 < len(body) && body[i+1] == '\\' {
			body = body[:i]
			break
		}
		if body[i] == 0x07 { // BEL terminator
			body = body[:i]
			break
		}
	}

	// Split by ";" — expected: ["4", "0", "rgb:RRRR/GGGG/BBBB"]
	// or for OSC 10/11/12: ["10", "rgb:RRRR/GGGG/BBBB"]
	parts := strings.Split(body, ";")
	if len(parts) < 2 {
		return ColorResponse{Index: -1}
	}

	cr := ColorResponse{Index: -1}

	var colorStr string
	if strings.HasPrefix(parts[0], "4") && len(parts) >= 3 {
		// OSC 4 response: "4 ; <index> ; rgb:..."
		cr.Index = atoiDef(parts[1])
		colorStr = parts[2]
	} else {
		// OSC 10/11/12 response: "10 ; rgb:..." or "11 ; rgb:..."
		colorStr = parts[1]
	}

	// Parse rgb:RRRR/GGGG/BBBB format
	if !strings.HasPrefix(colorStr, "rgb:") {
		return cr
	}
	colorStr = colorStr[4:] // strip "rgb:"
	rgb := strings.Split(colorStr, "/")
	if len(rgb) != 3 {
		return cr
	}

	cr.R = parseHexComponent(rgb[0])
	cr.G = parseHexComponent(rgb[1])
	cr.B = parseHexComponent(rgb[2])
	cr.Valid = true
	return cr
}

// IsDarkBackground heuristically determines whether a terminal with the
// given background RGB is "dark" (true) or "light" (false).
// Uses the relative luminance formula from WCAG 2.0.
func IsDarkBackground(r, g, b uint8) bool {
	// Convert to 0-1 range
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	// Apply gamma correction
	rf = gammaCorrect(rf)
	gf = gammaCorrect(gf)
	bf = gammaCorrect(bf)

	// Relative luminance (WCAG formula)
	luminance := 0.2126*rf + 0.7152*gf + 0.0722*bf

	// Dark if luminance < 0.5
	return luminance < 0.5
}

// gammaCorrect applies sRGB gamma correction for luminance calculation.
func gammaCorrect(v float64) float64 {
	if v <= 0.03928 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

// parseHexComponent parses a hex color component (1-4 digits) to uint8.
// Examples: "ff" → 255, "ffff" → 255, "0" → 0, "80" → 128.
func parseHexComponent(s string) uint8 {
	if len(s) == 0 {
		return 0
	}
	val := uint32(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		var d uint32
		switch {
		case c >= '0' && c <= '9':
			d = uint32(c - '0')
		case c >= 'a' && c <= 'f':
			d = uint32(c - 'a' + 10)
		case c >= 'A' && c <= 'F':
			d = uint32(c - 'A' + 10)
		default:
			return 0
		}
		val = val*16 + d
	}
	// Scale to 8-bit based on number of hex digits
	switch len(s) {
	case 1:
		// 4-bit: scale to 8-bit (0x0→0, 0xf→255)
		return uint8(val*0x11)
	case 2:
		return uint8(val)
	case 3:
		// 12-bit: scale down
		return uint8(val >> 4)
	case 4:
		// 16-bit: scale down to 8-bit
		return uint8(val >> 8)
	default:
		return uint8(val)
	}
}

// atoiDef parses a decimal string to int. Returns 0 on error.
func atoiDef(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// colorItoa converts an int to its decimal string representation.
func colorItoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
