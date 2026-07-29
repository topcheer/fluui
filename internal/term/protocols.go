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
// Cursor Save/Restore — DECSC/DECRC and ANSI.SYS
// ---------------------------------------------------------------------------

// SaveCursor saves the current cursor position and attributes (DECSC).
const SaveCursor = "\x1b7"

// RestoreCursor restores the previously saved cursor position and attributes (DECRC).
const RestoreCursor = "\x1b8"

// SaveCursorANSI is the ANSI.SYS variant of cursor save (ESC 7 is DEC, CSI s is ANSI).
const SaveCursorANSI = "\x1b[s"

// RestoreCursorANSI is the ANSI.SYS variant of cursor restore (CSI u).
const RestoreCursorANSI = "\x1b[u"

// ---------------------------------------------------------------------------
// Scroll Region — DECSTBM
// ---------------------------------------------------------------------------

// SetScrollRegion sets the top and bottom margins for scrolling (DECSTBM).
// top and bottom are 1-based line numbers. Pass 0,0 to reset to full screen.
func SetScrollRegion(top, bottom int) string {
	if top < 0 {
		top = 0
	}
	if bottom < 0 {
		bottom = 0
	}
	return "\x1b[" + intToStr(top) + ";" + intToStr(bottom) + "r"
}

// ResetScrollRegion resets the scroll region to the full screen.
const ResetScrollRegion = "\x1b[r"

// ---------------------------------------------------------------------------
// Device Status Report — DSR
// ---------------------------------------------------------------------------

// QueryCursorPosition asks the terminal to report the cursor position.
// The response is CSI row;col R (1-based).
const QueryCursorPosition = "\x1b[6n"

// QueryTerminalSize asks the terminal to report its size via DSR (CSI 14 t).
// The response is CSI 4 ; height ; width t (pixels).
const QueryTerminalSize = "\x1b[14t"

// QueryCellSize asks the terminal to report the cell size in pixels (CSI 16 t).
// The response is CSI 6 ; height ; width t.
const QueryCellSize = "\x1b[16t"

// ParseCursorPositionResponse parses a DSR cursor position response (CSI row;col R).
// Returns 1-based row, col and ok=true on success.
func ParseCursorPositionResponse(s string) (row, col int, ok bool) {
	// Expected format: ESC [ row ; col R
	if len(s) < 6 || s[0] != 0x1b || s[1] != '[' {
		return 0, 0, false
	}
	rest := s[2:]
	rEnd := 0
	for rEnd < len(rest) && rest[rEnd] != ';' {
		rEnd++
	}
	if rEnd == len(rest) {
		return 0, 0, false
	}
	cPart := rest[rEnd+1:]
	cEnd := 0
	for cEnd < len(cPart) && cPart[cEnd] != 'R' {
		cEnd++
	}
	if cEnd == 0 || cEnd == len(cPart) {
		return 0, 0, false
	}
	row = atoiDef(rest[:rEnd])
	if row == 0 {
		row = 1
	}
	col = atoiDef(cPart[:cEnd])
	if col == 0 {
		col = 1
	}
	return row, col, true
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

// enhancement flag 1). After enabling, the Parser will receive CSI u forms.


// ---------------------------------------------------------------------------
// Notification bell —BEL byte.
// ---------------------------------------------------------------------------

// Bell is the BEL control character (0x07).
const Bell = "\x07"

// ---------------------------------------------------------------------------
// Cursor Shape — DECSCUSR (CSI Ps SP q)
// ---------------------------------------------------------------------------

// CursorShape specifies the visual style of the text cursor.
type CursorShape int

const (
	// CursorShapeDefault lets the terminal decide (typically blinking block).
	CursorShapeDefault CursorShape = 0
	// CursorShapeBlinkingBlock is a blinking filled rectangle (DECSCUSR 0/1).
	CursorShapeBlinkingBlock CursorShape = 1
	// CursorShapeSteadyBlock is a non-blinking filled rectangle (DECSCUSR 2).
	CursorShapeSteadyBlock CursorShape = 2
	// CursorShapeBlinkingUnderline is a blinking underline (DECSCUSR 3).
	CursorShapeBlinkingUnderline CursorShape = 3
	// CursorShapeSteadyUnderline is a non-blinking underline (DECSCUSR 4).
	CursorShapeSteadyUnderline CursorShape = 4
	// CursorShapeBlinkingBar is a blinking vertical bar (xterm DECSCUSR 5).
	CursorShapeBlinkingBar CursorShape = 5
	// CursorShapeSteadyBar is a non-blinking vertical bar (xterm DECSCUSR 6).
	CursorShapeSteadyBar CursorShape = 6
)

// SetCursorShape returns the DECSCUSR escape sequence to change the cursor
// shape. Use CursorShapeSteadyBar for text editing mode and CursorShapeSteadyBlock
// for normal mode.
//
// Format: CSI Ps SP q   (where SP is a literal space 0x20)
// Supported by: xterm, iTerm2, Kitty, WezTerm, Alacritty, GNOME, Windows Terminal.
func SetCursorShape(shape CursorShape) string {
	if shape < 0 || shape > 6 {
		shape = 0
	}
	return "\x1b[" + string(rune('0'+shape)) + " q"
}

// ResetCursorShape restores the cursor to the terminal's default shape.
func ResetCursorShape() string {
	return "\x1b[0 q"
}

// ---------------------------------------------------------------------------
// Desktop Notification — OSC 9 (iTerm2 / WezTerm)
// ---------------------------------------------------------------------------

// DesktopNotification sends a desktop notification via OSC 9.
// iTerm2 and WezTerm display this as a system notification.
// The message appears in the OS notification center.
//
// Format: ESC ] 9 ; <message> BEL
func DesktopNotification(message string) string {
	return "\x1b]9;" + escapeOSCString(message) + "\x07"
}

// escapeOSCString replaces BEL bytes inside OSC payloads.
func escapeOSCString(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0x07 {
			return strings.ReplaceAll(s, "\x07", "\\x07")
		}
	}
	return s
}

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

// itoa converts a non-negative int to its ASCII string representation.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
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

// ---------------------------------------------------------------------------
// iTerm2 Inline Images (OSC 1337)
// ---------------------------------------------------------------------------
//
// The iTerm2 inline image protocol allows displaying images directly in
// the terminal. Supported by: iTerm2, WezTerm (partial), and other terminals.
//
// Format: ESC ] 1337 ; File=<args> : <base64-encoded-data> BEL
//
// Arguments:
//   name=base64(filename)
//   size=<bytes>
//   width=<N|Npx|N%|auto>
//   height=<N|Npx|N%|auto>
//   preserveAspectRatio=0|1
//   inline=0|1

// ImageSize specifies how an image dimension should be interpreted.
type ImageSize struct {
	Value int    // N (for cells or pixels) or percentage
	Unit  string // "", "px", "%", or "auto"
}

// AutoSize represents an automatic image dimension.
var AutoSize = ImageSize{Unit: "auto"}

// ImageOptions controls how an inline image is displayed.
type ImageOptions struct {
	Name                 string    // filename (will be base64 encoded)
	Width                ImageSize // display width
	Height               ImageSize // display height
	PreserveAspectRatio  bool      // default true
	Inline               bool      // default true (display inline vs download)
}

// DefaultImageOptions returns sensible defaults for inline image display.
func DefaultImageOptions() ImageOptions {
	return ImageOptions{
		Width:               AutoSize,
		Height:              AutoSize,
		PreserveAspectRatio: true,
		Inline:              true,
	}
}

// InlineImage generates an iTerm2 OSC 1337 escape sequence for displaying
// a base64-encoded image inline in the terminal.
// The data should be raw image bytes (PNG, JPEG, GIF, etc.).
func InlineImage(data []byte, opts ImageOptions) string {
	var args strings.Builder
	args.Grow(128)

	// name (base64 encoded filename)
	if opts.Name != "" {
		args.WriteString("name=")
		args.WriteString(base64.StdEncoding.EncodeToString([]byte(opts.Name)))
		args.WriteString(";")
	}

	// size
	args.WriteString("size=")
	args.WriteString(colorItoa(len(data)))
	args.WriteString(";")

	// width
	if opts.Width.Unit == "auto" {
		args.WriteString("width=auto;")
	} else if opts.Width.Value > 0 {
		args.WriteString("width=")
		args.WriteString(colorItoa(opts.Width.Value))
		args.WriteString(opts.Width.Unit)
		args.WriteString(";")
	}

	// height
	if opts.Height.Unit == "auto" {
		args.WriteString("height=auto;")
	} else if opts.Height.Value > 0 {
		args.WriteString("height=")
		args.WriteString(colorItoa(opts.Height.Value))
		args.WriteString(opts.Height.Unit)
		args.WriteString(";")
	}

	// preserveAspectRatio
	if opts.PreserveAspectRatio {
		args.WriteString("preserveAspectRatio=1;")
	} else {
		args.WriteString("preserveAspectRatio=0;")
	}

	// inline
	if opts.Inline {
		args.WriteString("inline=1;")
	} else {
		args.WriteString("inline=0;")
	}

	// Build the full sequence
	b64 := base64.StdEncoding.EncodeToString(data)
	return "\x1b]1337;File=" + args.String() + ":" + b64 + "\x07"
}

// InlineImageBase64 generates an iTerm2 OSC 1337 escape sequence using
// pre-base64-encoded image data. This avoids re-encoding if the data is
// already base64.
func InlineImageBase64(b64Data string, opts ImageOptions) string {
	var args strings.Builder
	args.Grow(128)

	if opts.Name != "" {
		args.WriteString("name=")
		args.WriteString(base64.StdEncoding.EncodeToString([]byte(opts.Name)))
		args.WriteString(";")
	}

	if opts.Width.Unit == "auto" {
		args.WriteString("width=auto;")
	} else if opts.Width.Value > 0 {
		args.WriteString("width=")
		args.WriteString(colorItoa(opts.Width.Value))
		args.WriteString(opts.Width.Unit)
		args.WriteString(";")
	}

	if opts.Height.Unit == "auto" {
		args.WriteString("height=auto;")
	} else if opts.Height.Value > 0 {
		args.WriteString("height=")
		args.WriteString(colorItoa(opts.Height.Value))
		args.WriteString(opts.Height.Unit)
		args.WriteString(";")
	}

	if opts.PreserveAspectRatio {
		args.WriteString("preserveAspectRatio=1;")
	} else {
		args.WriteString("preserveAspectRatio=0;")
	}

	if opts.Inline {
		args.WriteString("inline=1;")
	} else {
		args.WriteString("inline=0;")
	}

	return "\x1b]1337;File=" + args.String() + ":" + b64Data + "\x07"
}

// KittyImageBase64 generates a Kitty Graphics Protocol escape sequence for
// displaying a base64-encoded image. Kitty uses a different format from iTerm2.
//
// Format: ESC _ Gi=31;<base64-data> ESC \
//
// This is the "transmit pixel data" action (a=t, i=31 for new image).
func KittyImageBase64(b64Data string, width, height int) string {
	// Kitty graphics: ESC _ G <key=value pairs>;<base64 data> ESC \
	params := "a=t,f=100,s=" + colorItoa(width) + ",v=" + colorItoa(height)
	return "\x1b_G" + params + ";" + b64Data + "\x1b\\"
}

// KittyDeleteAllImages generates a Kitty Graphics Protocol sequence that
// removes all placed images from the terminal.
func KittyDeleteAllImages() string {
	return "\x1b_Ga=d,d=a\x1b\\"
}

// KittyQueryCell generates a Kitty Graphics Protocol sequence that queries
// whether there is an image at the specified cell position.
func KittyQueryCell(col, row int) string {
	return "\x1b_Ga=q,s=" + colorItoa(col) + ",v=" + colorItoa(row) + "\x1b\\"
}

// ---------------------------------------------------------------------------
// Terminal Capability Detection — DA1, DA2, XTVERSION, XTGETTCAP
// ---------------------------------------------------------------------------

// QueryDA1 sends the Primary Device Attributes request (CSI c).
// The terminal responds with CSI ? ... c listing its capabilities
// (e.g. CSI ? 62 ; 1 ; 2 ; 4 ; 6 ; 9 ; 15 ; 16 ; 29 c for VT220).
const QueryDA1 = "\x1b[c"

// QueryDA2 sends the Secondary Device Attributes request (CSI > c).
// The terminal responds with CSI > Pn ; Pv ; Pc c where:
//   Pn = terminal type (0=VT100, 1=VT220, 2=VT240, 18=VT330, 19=VT340, 24=VT320)
//   Pv = firmware version (encoded)
//   Pc = number of ROM cartridges
const QueryDA2 = "\x1b[>c"

// QueryXTVersion sends the terminal identification request (CSI > q).
// The terminal responds with CSI > Ps ; ... q where Ps is the terminal
// name and version, e.g. "xterm(372)" or "tmux 3.3a".
// Uses the XTVERSION extension (xterm patch 282+, supported by many terminals).
const QueryXTVersion = "\x1b[>q"

// QueryXTGetTCAP sends a terminfo capability query (CSI + q).
// Pass the hex-encoded capability name(s). The terminal responds with
// CSI 1 + r Pt = Pv ST or CSI 0 + r if unsupported.
// Multiple capabilities can be queried by separating hex names with ";".
func QueryXTGetTCAP(hexCapNames string) string {
	return "\x1b[+q" + hexCapNames
}

// EncodeTCapName encodes a terminfo capability name to the hex format
// expected by XTGETTCAP. For example "TN" → "544e", "Co" → "436f".
func EncodeTCapName(name string) string {
	const hexChars = "0123456789abcdef"
	var buf [256]byte
	out := buf[:0]
	for i := 0; i < len(name) && len(out)+2 <= cap(buf); i++ {
		out = append(out, hexChars[name[i]>>4], hexChars[name[i]&0xf])
	}
	return string(out)
}

// ParseDA1Response parses a Primary Device Attributes response.
// Expected format: ESC [ ? attr1 ; attr2 ; ... c
// Returns the list of attribute codes and ok=true on success.
func ParseDA1Response(s string) (attrs []int, ok bool) {
	// Must start with ESC [ ?
	if len(s) < 4 || s[0] != 0x1b || s[1] != '[' || s[2] != '?' {
		return nil, false
	}
	// Must end with 'c'
	if s[len(s)-1] != 'c' {
		return nil, false
	}
	body := s[3 : len(s)-1]
	if len(body) == 0 {
		return nil, false
	}
	// Split by ';'
	start := 0
	for start <= len(body) {
		end := start
		for end < len(body) && body[end] != ';' {
			end++
		}
		if end > start {
			attrs = append(attrs, atoiDef(body[start:end]))
		}
		start = end + 1
	}
	if len(attrs) == 0 {
		return nil, false
	}
	return attrs, true
}

// DA2Response holds the parsed result of a Secondary Device Attributes response.
type DA2Response struct {
	TerminalType int // Pn: terminal type code
	Version      int // Pv: firmware version
	ROMCartridges int // Pc: ROM cartridge count (usually 0)
}

// ParseDA2Response parses a Secondary Device Attributes response.
// Expected format: ESC [ > Pn ; Pv ; Pc c
func ParseDA2Response(s string) (resp DA2Response, ok bool) {
	if len(s) < 5 || s[0] != 0x1b || s[1] != '[' || s[2] != '>' {
		return DA2Response{}, false
	}
	if s[len(s)-1] != 'c' {
		return DA2Response{}, false
	}
	body := s[3 : len(s)-1]
	fields := splitSemi(body)
	if len(fields) == 0 {
		return DA2Response{}, false
	}
	resp.TerminalType = atoiDef(fields[0])
	if len(fields) > 1 {
		resp.Version = atoiDef(fields[1])
	}
	if len(fields) > 2 {
		resp.ROMCartridges = atoiDef(fields[2])
	}
	return resp, true
}

// XTVersionResponse holds the parsed terminal name and version.
type XTVersionResponse struct {
	Name    string // e.g. "xterm", "tmux", "contour"
	Version string // e.g. "372", "3.3a"
}

// ParseXTVersionResponse parses a terminal version response.
// Expected format: ESC [ > name ( version ST
// or legacy format: ESC [ > name ; version q
// The ST terminator is ESC \ (0x1b 0x5c) or BEL (0x07).
func ParseXTVersionResponse(s string) (resp XTVersionResponse, ok bool) {
	// Must start with ESC [ >
	if len(s) < 4 || s[0] != 0x1b || s[1] != '[' || s[2] != '>' {
		return XTVersionResponse{}, false
	}
	rest := s[3:]
	// Strip trailing ST (ESC \ ) or BEL
	if len(rest) >= 2 && rest[len(rest)-2] == 0x1b && rest[len(rest)-1] == '\\' {
		rest = rest[:len(rest)-2]
	} else if len(rest) >= 1 && rest[len(rest)-1] == 0x07 {
		rest = rest[:len(rest)-1]
	} else if len(rest) >= 1 && rest[len(rest)-1] == 'q' {
		// Legacy CSI > ... q format
		rest = rest[:len(rest)-1]
	} else {
		return XTVersionResponse{}, false
	}
	// Split on '(' — name(version) format
	openParen := indexOfByte(rest, '(')
	if openParen >= 0 {
		resp.Name = rest[:openParen]
		closeParen := indexOfByte(rest[openParen+1:], ')')
		if closeParen >= 0 {
			resp.Version = rest[openParen+1 : openParen+1+closeParen]
		} else {
			resp.Version = rest[openParen+1:]
		}
		return resp, true
	}
	// Split on ';' — name;version format
	semiPos := indexOfByte(rest, ';')
	if semiPos >= 0 {
		resp.Name = rest[:semiPos]
		resp.Version = rest[semiPos+1:]
	} else {
		resp.Name = rest
	}
	return resp, true
}

// ParseXTGetTCapResponse parses an XTGETTCAP response.
// Expected format for success: ESC P 1 + r hexName = hexValue ESC \
// Expected format for failure: ESC P 0 + r ESC \
// Returns the hex-encoded name, hex-encoded value, and ok=true on success.
func ParseXTGetTCapResponse(s string) (hexName, hexValue string, ok bool) {
	// Must start with ESC P (DCS)
	if len(s) < 4 || s[0] != 0x1b || s[1] != 'P' {
		return "", "", false
	}
	rest := s[2:]
	// Strip trailing ST (ESC \ )
	if len(rest) >= 2 && rest[len(rest)-2] == 0x1b && rest[len(rest)-1] == '\\' {
		rest = rest[:len(rest)-2]
	} else if len(rest) >= 1 && rest[len(rest)-1] == 0x07 {
		rest = rest[:len(rest)-1]
	} else {
		return "", "", false
	}
	// Check for failure: starts with "0+r"
	if len(rest) >= 3 && rest[0] == '0' && rest[1] == '+' && rest[2] == 'r' {
		return "", "", false
	}
	// Check for success: starts with "1+r"
	if len(rest) < 3 || rest[0] != '1' || rest[1] != '+' || rest[2] != 'r' {
		return "", "", false
	}
	body := rest[3:]
	// Split on '='
	eqPos := indexOfByte(body, '=')
	if eqPos < 0 {
		return body, "", true
	}
	return body[:eqPos], body[eqPos+1:], true
}

// --- internal helpers ---

// splitSemi splits a string by ';' and returns non-empty fields.
func splitSemi(s string) []string {
	if len(s) == 0 {
		return nil
	}
	var fields []string
	start := 0
	for start <= len(s) {
		end := start
		for end < len(s) && s[end] != ';' {
			end++
		}
		if end > start {
			fields = append(fields, s[start:end])
		}
		start = end + 1
	}
	return fields
}

// indexOfByte returns the index of the first occurrence of b in s, or -1.
func indexOfByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// OSC 133 — Shell Integration / Prompt Marking (Semantic Prompts)
// ---------------------------------------------------------------------------

// OSC133Mark identifies the type of shell integration marker.
type OSC133Mark uint8

const (
	OSC133Unknown OSC133Mark = iota
	OSC133PromptStart   // A — marks the start of a prompt
	OSC133PromptEnd     // A with optional metadata — end of prompt text
	OSC133CommandStart  // B — marks the start of user input / command
	OSC133OutputStart   // C — marks the start of command output
	OSC133CommandEnd    // D — marks the end of command with exit code
)

// osc133Bel builds an OSC 133 sequence with BEL terminator.
func osc133Bel(payload string) string {
	return "\x1b]133;" + payload + "\x07"
}

// OSC133PromptStartSeq generates the shell integration marker for the start
// of a prompt region. Supported by iTerm2, Kitty, WezTerm, GNOME Terminal,
// Windows Terminal, and others.
//
// Format: ESC ] 133 ; A ST
func OSC133PromptStartSeq() string {
	return osc133Bel("A")
}

// OSC133PromptStartMeta is like OSC133PromptStartSeq but includes optional
// metadata (e.g. claude=1, tmux=prompt). The meta string is embedded as-is.
//
// Format: ESC ] 133 ; A ; <meta> ST
func OSC133PromptStartMeta(meta string) string {
	if meta == "" {
		return osc133Bel("A")
	}
	return osc133Bel("A;" + meta)
}

// OSC133CommandStartSeq generates the marker for the start of user input.
// This marks where the command begins (after the prompt).
//
// Format: ESC ] 133 ; B ST
func OSC133CommandStartSeq() string {
	return osc133Bel("B")
}

// OSC133OutputStartSeq generates the marker for the start of command output.
//
// Format: ESC ] 133 ; C ST
func OSC133OutputStartSeq() string {
	return osc133Bel("C")
}

// OSC133CommandEndSeq generates the marker for the end of a command with
// an exit code. The exitCode is the process return code (0 = success).
//
// Format: ESC ] 133 ; D ; <exit_code> ST
func OSC133CommandEndSeq(exitCode int) string {
	return osc133Bel("D;" + intToStr(exitCode))
}

// OSC133Result holds a parsed OSC 133 marker.
type OSC133Result struct {
	Mark     OSC133Mark
	ExitCode int    // only valid when Mark == OSC133CommandEnd
	Meta     string // optional metadata (e.g. from A;<meta>)
}

// ParseOSC133 attempts to parse an OSC 133 sequence from the given string.
// The input should be a complete OSC 133 sequence (with BEL or ST terminator).
// Returns the parsed result and ok=true on success.
//
// Accepted formats:
//
//	ESC ] 133 ; A ST          → PromptStart
//	ESC ] 133 ; A ; meta ST   → PromptStart with metadata
//	ESC ] 133 ; B ST          → CommandStart
//	ESC ] 133 ; C ST          → OutputStart
//	ESC ] 133 ; D ST          → CommandEnd (exit code omitted)
//	ESC ] 133 ; D ; 0 ST      → CommandEnd with exit code
func ParseOSC133(s string) (OSC133Result, bool) {
	// Must start with ESC ] 133 ;
	if len(s) < 7 || s[0] != 0x1b || s[1] != ']' {
		return OSC133Result{}, false
	}
	// Check "133;" prefix
	if s[2] != '1' || s[3] != '3' || s[4] != '3' || s[5] != ';' {
		return OSC133Result{}, false
	}

	// Find terminator: BEL (0x07) or ST (ESC \)
	body := s[6:]
	stripped, hadTerminator := stripTerminatorChecked(body)
	if !hadTerminator || len(stripped) == 0 {
		return OSC133Result{}, false
	}
	body = stripped

	// First char is the mark type
	markChar := body[0]
	result := OSC133Result{}

	switch markChar {
	case 'A':
		result.Mark = OSC133PromptStart
		// Check for optional metadata: A;meta
		if len(body) > 1 && body[1] == ';' {
			result.Meta = body[2:]
		}
	case 'B':
		result.Mark = OSC133CommandStart
	case 'C':
		result.Mark = OSC133OutputStart
	case 'D':
		result.Mark = OSC133CommandEnd
		// Check for optional exit code: D;<code>
		if len(body) > 1 && body[1] == ';' {
			result.ExitCode = atoiDef(body[2:])
		} else {
			result.ExitCode = -1 // exit code omitted
		}
	default:
		return OSC133Result{}, false
	}

	return result, true
}

// stripTerminator removes a trailing BEL or ST from the OSC body.
func stripTerminator(s string) string {
	if len(s) >= 1 && s[len(s)-1] == 0x07 {
		return s[:len(s)-1]
	}
	if len(s) >= 2 && s[len(s)-2] == 0x1b && s[len(s)-1] == '\\' {
		return s[:len(s)-2]
	}
	return s
}

// stripTerminatorChecked removes a trailing BEL or ST and reports whether one was found.
func stripTerminatorChecked(s string) (string, bool) {
	if len(s) >= 1 && s[len(s)-1] == 0x07 {
		return s[:len(s)-1], true
	}
	if len(s) >= 2 && s[len(s)-2] == 0x1b && s[len(s)-1] == '\\' {
		return s[:len(s)-2], true
	}
	return s, false
}

// ─── OSC 133: Shell Integration (Semantic Prompt Marking) ───
//
// OSC 133 is supported by iTerm2, WezTerm, Kitty, GNOME Terminal, and others.
// It marks command boundaries so terminals can identify prompts, commands,
// and their output — essential for AI terminal assistants to parse sessions.

// ShellPromptStart marks the beginning of a shell prompt (OSC 133;A).
func ShellPromptStart() string {
	return "\x1b]133;A\x07"
}

// ShellCommandStart marks the end of the prompt / start of the command (OSC 133;B).
func ShellCommandStart() string {
	return "\x1b]133;B\x07"
}

// ShellOutputStart marks the start of command output (OSC 133;C).
func ShellOutputStart() string {
	return "\x1b]133;C\x07"
}

// ShellOutputEnd marks the end of command output with an optional exit code (OSC 133;D).
// If exitCode is -1, no exit code is included.
func ShellOutputEnd(exitCode int) string {
	if exitCode < 0 {
		return "\x1b]133;D\x07"
	}
	return "\x1b]133;D;" + intToStr(exitCode) + "\x07"
}

// ShellIntegration wraps a command execution with all four OSC 133 markers:
// prompt start, command start, output start, and output end.
// Usage: fmt.Print(ShellIntegration(0)) at the end of a command.
// The caller is responsible for emitting each marker at the right point.
// This helper emits a complete set for testing/demonstration.
func ShellIntegration(exitCode int) string {
	return ShellPromptStart() + ShellCommandStart() + ShellOutputStart() + ShellOutputEnd(exitCode)
}

// ─── Bracketed Paste Mode (DECSET/DECRST 2004) ───
//
// When enabled, pasted text is wrapped in ESC[200~ ... ESC[201~.
// This allows TUI apps to distinguish paste from typed input.
// Enable/Disable constants are already defined above.

// IsBracketedPaste checks if data is a bracketed paste sequence.
// Returns the unwrapped content and true if it is, or data and false otherwise.
func IsBracketedPaste(data string) (string, bool) {
	const prefix = "\x1b[200~"
	const suffix = "\x1b[201~"
	if len(data) >= len(prefix)+len(suffix) &&
		data[:len(prefix)] == prefix &&
		data[len(data)-len(suffix):] == suffix {
		return data[len(prefix) : len(data)-len(suffix)], true
	}
	return data, false
}

// ─── OSC 4: Query/Set Palette Color ───
//
// OSC 4;n;? BEL queries the color of palette index n.
// OSC 4;n;spec BEL sets the color of palette index n.

// QueryPaletteColorIndex queries a specific palette color index via OSC 4.
func QueryPaletteColorIndex(index int) string {
	return "\x1b]4;" + intToStr(index) + ";?\x07"
}

// SetPaletteColor sets a palette color index to an RGB value via OSC 4.
func SetPaletteColor(index int, r, g, b uint8) string {
	return "\x1b]4;" + intToStr(index) + ";rgb:" +
		colorItoa(int(r)) + "/" + colorItoa(int(g)) + "/" + colorItoa(int(b)) + "\x07"
}

// ─── OSC 7: Current Working Directory ───
//
// OSC 7 tells the terminal the current working directory, enabling:
// - Cmd+Click on paths in iTerm2/WezTerm to open them
// - New tabs/windows inherit the CWD
// - Terminal remembers CWD per-pane

// ReportWorkingDir reports the current directory to the terminal via OSC 7.
// The path should be an absolute filesystem path (e.g., "/home/user/project").
func ReportWorkingDir(path string) string {
	return "\x1b]7;file://" + escapeOSCString(path) + "\x07"
}

// ReportWorkingDirHost reports the current directory with an explicit hostname.
// This is the full OSC 7 format used by shells: file://hostname/path
func ReportWorkingDirHost(host, path string) string {
	return "\x1b]7;file://" + escapeOSCString(host) + escapeOSCString(path) + "\x07"
}

// ─── OSC 777: URxvt-style Notification ───
//
// OSC 777 is supported by rxvt-unicode and some derived terminals.
// It displays a desktop notification with a title and message.
// Simpler than OSC 9 — doesn't support HTML or images.

// URxvtNotification sends a desktop notification via OSC 777.
// Format: ESC ] 777 ; notify ; TITLE ; MESSAGE BEL
func URxvtNotification(title, message string) string {
	return "\x1b]777;notify;" + escapeOSCString(title) + ";" + escapeOSCString(message) + "\x07"
}

// ─── OSC 9;4: iTerm2 Progress Bar ───
//
// OSC 9;4 controls the progress bar displayed in the terminal tab/dock icon
// on macOS (iTerm2, WezTerm). Useful for AI streaming progress indicators.

// ProgressBarStyle controls how the progress bar appears.
type ProgressBarStyle int

const (
	// ProgressBarSet sets the progress percentage (0-100).
	ProgressBarSet ProgressBarStyle = iota
	// ProgressBarError sets the progress bar to error state.
	ProgressBarError
	// ProgressBarWarning sets the progress bar to warning state.
	ProgressBarWarning
	// ProgressBarIndeterminate shows an indeterminate (spinning) state.
	ProgressBarIndeterminate
	// ProgressBarClear removes the progress bar.
	ProgressBarClear
)

// SetTabProgressBar sets the terminal tab progress bar via OSC 9;4.
// pct should be 0-100. style controls the bar appearance.
func SetTabProgressBar(pct int, style ProgressBarStyle) string {
	var state string
	switch style {
	case ProgressBarError:
		state = "2"
	case ProgressBarWarning:
		state = "3"
	case ProgressBarIndeterminate:
		state = "4"
	case ProgressBarClear:
		state = "1"
	default:
		state = ""
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return "\x1b]9;4;" + state + ";" + intToStr(pct) + "\x07"
}

// ClearTabProgressBar removes the tab progress bar.
func ClearTabProgressBar() string {
	return "\x1b]9;4;0;\x07"
}

// ─── Focus Reporting (CSI ?1004) ───
//
// When enabled, the terminal sends FocusIn/FocusOut events to the application.
// This is already partially supported; these provide the enable/disable helpers.

// EnableFocusReporting enables focus in/out event reporting (CSI ?1004 h).
const EnableFocusReporting = "\x1b[?1004h"

// DisableFocusReporting disables focus in/out event reporting (CSI ?1004 l).
const DisableFocusReporting = "\x1b[?1004l"

// ─── OSC 633: VSCode Shell Integration ───
//
// OSC 633 is used by VSCode's integrated terminal for shell integration.
// It marks command boundaries for prompt detection and command output tracking.

// VSCodePromptStart marks the start of a command prompt via OSC 633.
const VSCodePromptStart = "\x1b]633;A\x07"

// VSCodeCommandStart marks the start of a command after the prompt via OSC 633.
const VSCodeCommandStart = "\x1b]633;C\x07"

// VSCodeCommandEnd marks the end of a command's output via OSC 633.
const VSCodeCommandEnd = "\x1b]633;D\x07"

// VSCodePromptEnd marks the end of a prompt (before command) via OSC 633.
const VSCodePromptEnd = "\x1b]633;B\x07"

// VSCodeCwd reports the current working directory to VSCode via OSC 633.
func VSCodeCwd(path string) string {
	return "\x1b]633;P=Cwd=" + escapeOSCString(path) + "\x07"
}

// ─── OSC 99: Kitty Desktop Notification ───
//
// OSC 99 is Kitty's enhanced notification system with rich formatting
// (icons, urgency levels, actions). More capable than OSC 777 or OSC 9.

// KittyNotification sends a desktop notification via OSC 99.
// Supports title, body, and optional icon.
func KittyNotification(title, body string) string {
	return "\x1b]99;i=1:d=0;" + escapeOSCString(title) + "\x1b\\" +
		"\x1b]99;i=1:d=1;" + escapeOSCString(body) + "\x1b\\"
}

// ─── OSC 8 with explicit ID ───
//
// Allows multiple distinct hyperlinks by assigning IDs.

// OSC8LinkWithID creates a hyperlink with an explicit ID for tracking.
func OSC8LinkWithID(id, text string, opts HyperlinkOptions) string {
	if id != "" {
		opts.ID = id
	}
	return OSC8Link(opts, text)
}

// EnableKittyKeyboard enables the Kitty keyboard enhancement protocol (CSI > 1 u).
const EnableKittyKeyboard = "\x1b[>1u"

// DisableKittyKeyboard disables the Kitty keyboard enhancement protocol.
const DisableKittyKeyboard = "\x1b[<u"

// ─── Kitty Keyboard Enhancement Protocol ───
//
// CSI > 1 u enables Kitty keyboard protocol for enhanced key reporting.
// CSI < u disables it.



// ─── Sixel Graphics Support (DCS q) ───
//
// Sixel is a DEC graphics protocol that encodes pixel images using a compact
// character-based format. Supported by xterm, mlterm, mintty, RLogin, and
// many VT340-compatible terminals. The terminal reports Sixel capability
// in the DA1 response (attribute bit 4 of the first parameter).

// QuerySixelCapability generates a DA1 request that can be used to determine
// if the terminal supports Sixel graphics. The caller should parse the
// response with ParseDA1Response and check if attribute 4 is set.
func QuerySixelCapability() string {
	return "\x1b[c" // DA1 request — same as primary device attributes
}

// SixelStart begins a Sixel graphics sequence.
// DCS = integer parameter(s) q ... ST
// The DCS introducer is ESC P (0x1b 0x50), and the string terminator is ESC \ (0x1b 0x5c).
func SixelStart(palette int) string {
	if palette > 0 {
		return "\x1bP" + intToStr(palette) + "q"
	}
	return "\x1bPq"
}

// SixelEnd terminates a Sixel graphics sequence.
func SixelEnd() string {
	return "\x1b\\"
}

// ─── DECRQM: Request Mode ───
//
// DECRQM (Request Mode) queries the current state of a DEC private or ANSI
// mode. The terminal responds with DECRPM (Report Mode).
// Format: CSI ? Pn $ p (DEC private) or CSI Pn $ p (ANSI)

// RequestDECMode generates a DECRQM request for a DEC private mode number.
// Example: RequestDECMode(2026) queries synchronized output state.
func RequestDECMode(mode int) string {
	return "\x1b[?" + intToStr(mode) + "$p"
}

// RequestANSIMode generates a DECRQM request for an ANSI mode number.
// Example: RequestANSIMode(4) queries insert/replace mode.
func RequestANSIMode(mode int) string {
	return "\x1b[" + intToStr(mode) + "$p"
}

// ─── OSC 52: Clipboard Access ───
//
// OSC 52 provides clipboard read/write access. Supported by xterm, mintty,
// and many modern terminals. The data must be base64-encoded.

// OSC52Copy writes base64-encoded text to the system clipboard via OSC 52.
// The b64 parameter should be base64-encoded clipboard content.
func OSC52Copy(b64 string) string {
	return "\x1b]52;c;" + b64 + "\x07"
}

// OSC52CopySelection writes base64-encoded text to the primary selection.
func OSC52CopySelection(b64 string) string {
	return "\x1b]52;p;" + b64 + "\x07"
}

// OSC52Query requests the current clipboard content via OSC 52.
// The terminal responds with the base64-encoded clipboard data.
func OSC52Query() string {
	return "\x1b]52;c;?\x07"
}

// ─── DECSED: Selective Erase in Display ───
//
// DECSED (CSI ? Pn J) erases lines in the display without affecting
// protected characters. Useful for TUI redraws that preserve borders.

// SelectiveEraseDisplay erases all unprotected content in the display.
const SelectiveEraseDisplay = "\x1b[?2J"

// SelectiveEraseToEnd erases unprotected content from cursor to end of display.
const SelectiveEraseToEnd = "\x1b[?0J"

// SelectiveEraseToStart erases unprotected content from start to cursor.
const SelectiveEraseToStart = "\x1b[?1J"
