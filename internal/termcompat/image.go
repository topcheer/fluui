package termcompat

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// ImageProtocol represents a terminal image display protocol.
type ImageProtocol int

const (
	// ImageNone means no image support detected.
	ImageNone ImageProtocol = iota
	// ImageSixel is the Sixel protocol (xterm + VT340, mlterm, mintty, foot).
	ImageSixel
	// ImageIterm2 is the iTerm2 inline image protocol (also supported by WezTerm).
	ImageIterm2
	// ImageKitty is the Kitty graphics protocol.
	ImageKitty
)

// ImageCapabilities describes what image features the terminal supports.
type ImageCapabilities struct {
	// Protocol is the best available image protocol.
	Protocol ImageProtocol

	// CanDisplay indicates whether any image protocol is available.
	CanDisplay bool

	// MaxWidth is the maximum image width in pixels (0 = unknown/unlimited).
	MaxWidth int

	// MaxHeight is the maximum image height in pixels (0 = unknown/unlimited).
	MaxHeight int

	// CanAnimate indicates animated image support (GIF, APNG).
	CanAnimate bool
}

// sixelTerminals is the set of terminal names known to support Sixel.
var sixelTerminals = map[string]bool{
	"mlterm":          true,
	"foot":            true,
	"mintty":          true,
	"xterm":           true, // xterm compiled with Sixel support
	"XTerm":           true,
	"rxvt-unicode":    true, // urxvt with Sixel patch
	"wezterm":         true, // WezTerm also supports Sixel
	"Gnome Terminal":  true, // GNOME Terminal with experimental Sixel
}

// sixelTermPrefixes matches $TERM values for Sixel-supporting terminals.
var sixelTermPrefixes = []string{
	"mlterm",
	"foot",
	"mintty",
}

// DetectImageProtocol determines the best available image protocol from the
// current environment variables.
func DetectImageProtocol() ImageCapabilities {
	return DetectImageProtocolFromEnv(
		os.Getenv("TERM_PROGRAM"),
		os.Getenv("TERM"),
		os.Getenv("COLORTERM"),
	)
}

// DetectImageProtocolFromEnv detects image protocol support from explicit
// environment variable values. This allows tests to inject mock values.
//
// Detection priority: Kitty > iTerm2 > Sixel > None.
// Kitty is checked first because it has the most features and is the most
// performant. iTerm2 protocol is checked second (WezTerm also supports it).
// Sixel is checked third from known terminal names and TERM prefixes.
func DetectImageProtocolFromEnv(termProgram, term, colorTerm string) ImageCapabilities {
	caps := ImageCapabilities{Protocol: ImageNone}

	// --- Kitty detection ---
	// TERM_PROGRAM=kitty or TERM=xterm-kitty
	if isKitty(termProgram, term) {
		caps.Protocol = ImageKitty
		caps.CanDisplay = true
		caps.CanAnimate = true
		return caps
	}

	// --- iTerm2 detection ---
	// TERM_PROGRAM=iTerm.app, or WezTerm (supports iTerm2 protocol)
	if isIterm2(termProgram, term) {
		caps.Protocol = ImageIterm2
		caps.CanDisplay = true
		caps.CanAnimate = true
		return caps
	}

	// --- Sixel detection ---
	// Check TERM_PROGRAM for known sixel terminals.
	if isSixel(termProgram, term) {
		caps.Protocol = ImageSixel
		caps.CanDisplay = true
		caps.CanAnimate = false // Sixel doesn't natively support animation
		return caps
	}

	return caps
}

// isKitty returns true if the terminal is Kitty.
func isKitty(termProgram, term string) bool {
	if termProgram == "kitty" {
		return true
	}
	if strings.HasPrefix(term, "xterm-kitty") {
		return true
	}
	return false
}

// isIterm2 returns true if the terminal supports the iTerm2 inline image protocol.
// WezTerm also supports this protocol.
func isIterm2(termProgram, term string) bool {
	if strings.Contains(termProgram, "iTerm") {
		return true
	}
	if strings.Contains(termProgram, "WezTerm") {
		return true
	}
	if strings.HasPrefix(term, "wezterm") {
		return true
	}
	return false
}

// isSixel returns true if the terminal likely supports Sixel graphics.
func isSixel(termProgram, term string) bool {
	// Check TERM_PROGRAM against known sixel terminals.
	if termProgram != "" && sixelTerminals[termProgram] {
		return true
	}

	// Check TERM prefixes for known sixel terminals.
	for _, prefix := range sixelTermPrefixes {
		if strings.HasPrefix(term, prefix) {
			return true
		}
	}

	// Check TERM for explicit sixel markers.
	if strings.Contains(term, "sixel") {
		return true
	}

	return false
}

// ImageProtocolName returns a human-readable name for the protocol.
func ImageProtocolName(p ImageProtocol) string {
	switch p {
	case ImageSixel:
		return "Sixel"
	case ImageIterm2:
		return "iTerm2"
	case ImageKitty:
		return "Kitty"
	default:
		return "None"
	}
}

// --- Escape sequence formatters ---

// FormatSixelPlaceholder returns a Sixel escape sequence for image data.
// The data should be raw Sixel-encoded bytes (DCS + ... + ST).
// This wraps the data in the proper DCS (Device Control String) prefix
// and ST (String Terminator) suffix.
func FormatSixelPlaceholder(data []byte) string {
	// DCS (Device Control String): ESC P ... ST
	// Sixel data typically starts with "8;1;q" (DPR=8, background=1, sixel)
	return "\x1bP" + string(data) + "\x1b\\"
}

// FormatIterm2Image returns the iTerm2 inline image escape sequence.
// The image data should be base64-encoded. The name is shown in the
// terminal's image proxy. Width and height are in pixels (0 = auto).
//
// Format: ESC ] 1337 ; File = inline=1 ; name=<base64> ; width=<W> ;
//         height=<H> : <base64-data> ST
func FormatIterm2Image(base64Data, name string, width, height int) string {
	// Base64-encode the filename.
	encodedName := base64.StdEncoding.EncodeToString([]byte(name))

	// Build the arguments.
	var args []string
	args = append(args, "inline=1")
	args = append(args, "name="+encodedName)
	if width > 0 {
		args = append(args, fmt.Sprintf("width=%d", width))
	}
	if height > 0 {
		args = append(args, fmt.Sprintf("height=%d", height))
	}

	argStr := strings.Join(args, ";")
	// OSC 1337 with ST terminator.
	return "\x1b]1337;File=" + argStr + ":" + base64Data + "\x1b\\"
}

// FormatKittyImage returns the Kitty graphics protocol escape sequence.
// The image data should be base64-encoded. Width and height are in cells
// (0 = auto/1x1).
//
// Uses the "direct from stdin" transmission mode (t=d) for simplicity.
// Format: ESC _ G f=100 ; t=d ; s=<W> ; v=<H> ; q=1 ; a=T ;
//         m=1 ; <base64-chunk> ESC _ G m=0 ; <rest> ESC \
func FormatKittyImage(base64Data string, width, height int) string {
	// Kitty uses APC (Application Program Command): ESC _
	// We transmit the image in chunks of 4096 bytes of base64 data.
	const chunkSize = 4096

	var sb strings.Builder

	// Build the action key-value pairs for the first chunk.
	var kvs []string
	kvs = append(kvs, "a=T") // action: transmit
	kvs = append(kvs, "f=100") // format: PNG
	kvs = append(kvs, "t=d")   // transmission: direct
	if width > 0 {
		kvs = append(kvs, fmt.Sprintf("s=%d", width))
	}
	if height > 0 {
		kvs = append(kvs, fmt.Sprintf("v=%d", height))
	}
	kvs = append(kvs, "q=1") // suppress acknowledgments

	hdr := strings.Join(kvs, ",")

	// First chunk.
	if len(base64Data) <= chunkSize {
		// Single chunk.
		sb.WriteString("\x1b_G")
		sb.WriteString(hdr)
		sb.WriteString(",m=0;")
		sb.WriteString(base64Data)
		sb.WriteString("\x1b\\")
	} else {
		// Multi-chunk: first chunk with m=1 (more to follow).
		sb.WriteString("\x1b_G")
		sb.WriteString(hdr)
		sb.WriteString(",m=1;")
		sb.WriteString(base64Data[:chunkSize])
		sb.WriteString("\x1b\\")

		remaining := base64Data[chunkSize:]
		for len(remaining) > chunkSize {
			sb.WriteString("\x1b_Gm=1;")
			sb.WriteString(remaining[:chunkSize])
			sb.WriteString("\x1b\\")
			remaining = remaining[chunkSize:]
		}

		// Last chunk with m=0.
		sb.WriteString("\x1b_Gm=0;")
		sb.WriteString(remaining)
		sb.WriteString("\x1b\\")
	}

	return sb.String()
}
