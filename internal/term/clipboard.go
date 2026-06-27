package term

import (
	"encoding/base64"
	"strings"
)

// OSC52 escape sequence constants.
// OSC52 is a terminal escape sequence for clipboard access.
// Format: ESC ] 52 ; <clipboard> ; <base64-data> ESC \
// or with BEL terminator: ESC ] 52 ; <clipboard> ; <base64-data> BEL

const (
	// osc52Prefix is the start of an OSC52 sequence for the system clipboard.
	osc52Prefix = "\x1b]52;c;"
	// osc52SuffixBEL terminates the sequence with BEL.
	osc52SuffixBEL = "\x07"
	// osc52SuffixST terminates the sequence with String Terminator (ESC \).
	osc52SuffixST = "\x1b\\"
)

// ClipboardSource identifies which clipboard to target.
type ClipboardSource string

const (
	// ClipboardSystem targets the system clipboard (default).
	ClipboardSystem ClipboardSource = "c"
	// ClipboardPrimary targets the primary selection (X11).
	ClipboardPrimary ClipboardSource = "p"
	// ClipboardClipboard targets the CLIPBOARD selection (X11).
	ClipboardClipboard ClipboardSource = "c"
)

// CopyOSC52 generates an OSC52 escape sequence to set the clipboard content.
// The text is base64-encoded and wrapped in the proper escape sequence.
// Use the returned string by writing it directly to the terminal output.
func CopyOSC52(text string) string {
	return CopyOSC52Source(text, ClipboardSystem)
}

// CopyOSC52Source generates an OSC52 with a specific clipboard source.
func CopyOSC52Source(text string, source ClipboardSource) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	var sb strings.Builder
	sb.Grow(len(osc52Prefix) + len(source) + 1 + len(encoded) + len(osc52SuffixST))
	sb.WriteString("\x1b]52;")
	sb.WriteString(string(source))
	sb.WriteByte(';')
	sb.WriteString(encoded)
	sb.WriteString("\x1b\\")
	return sb.String()
}

// ParseOSC52Response extracts clipboard text from a terminal response to a
// paste query. The response has the format: ESC ] 52 ; <source> ; <base64> ESC \
// Returns the decoded text and true if parsing succeeded.
func ParseOSC52Response(response string) (string, bool) {
	// Strip leading ESC ] 52 ;
	idx := strings.Index(response, "\x1b]52;")
	if idx < 0 {
		return "", false
	}
	rest := response[idx+5:] // skip "\x1b]52;"

	// Find the second semicolon (after source)
	semi := strings.IndexByte(rest, ';')
	if semi < 0 {
		return "", false
	}
	encoded := rest[semi+1:]

	// Strip trailing ESC \ or BEL
	encoded = strings.TrimSuffix(encoded, "\x1b\\")
	encoded = strings.TrimSuffix(encoded, "\x07")
	encoded = strings.TrimSpace(encoded)

	// Check if this is a paste query echo (encoded == "?")
	// Return empty string and true — it's a valid OSC52 response, just a query.
	if encoded == "?" {
		return "", true
	}

	if encoded == "" {
		return "", true
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// PasteQuery generates the OSC52 query sequence to request clipboard content
// from the terminal. The terminal will respond with the clipboard data.
func PasteQuery() string {
	return "\x1b]52;c;?\x1b\\"
}

// IsOSC52Response checks whether a response string looks like an OSC52 reply.
func IsOSC52Response(s string) bool {
	return strings.HasPrefix(s, "\x1b]52;")
}
