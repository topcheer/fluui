package markdown

import (
	"regexp"
	"strings"
)

// OSC8 escape sequences for terminal hyperlinks.
// Format: ESC ]8;;URL ESC \ TEXT ESC ]8;; ESC \
const (
	osc8Start = "\x1b]8;;"
	osc8Sep   = "\x1b\\"
	osc8End   = "\x1b]8;;\x1b\\"
)

// LinkRenderer formats markdown links as either OSC8 hyperlinks (when the
// terminal supports them) or plain-text "text (url)" fallback.
type LinkRenderer struct {
	enabled bool
}

// NewLinkRenderer creates a LinkRenderer. If hasOSC8 is true, links are
// rendered as clickable OSC8 hyperlinks; otherwise a plain-text fallback
// is used.
func NewLinkRenderer(hasOSC8 bool) *LinkRenderer {
	return &LinkRenderer{enabled: hasOSC8}
}

// Enabled reports whether OSC8 hyperlink formatting is active.
func (lr *LinkRenderer) Enabled() bool {
	return lr.enabled
}

// FormatLink returns the formatted link string.
//
// When OSC8 is enabled, the result wraps the text in an OSC8 escape sequence
// that makes it clickable in supporting terminals:
//
//	ESC]8;;URL ESC\ TEXT ESC]8;; ESC\
//
// When disabled, the result is "text (url)" — the URL is shown in
// parentheses so the reader can copy-paste it.
//
// If url is empty, the text is returned as-is regardless of the enabled flag.
func (lr *LinkRenderer) FormatLink(text, url string) string {
	if url == "" {
		return text
	}
	if lr.enabled {
		return osc8Start + url + osc8Sep + text + osc8End
	}
	return text + " (" + url + ")"
}

// osc8Pattern matches a complete OSC8 hyperlink sequence, capturing the
// URL (group 1) and the display text (group 2).
var osc8Pattern = regexp.MustCompile(
	`\x1b\]8;;([^\x1b]*)\x1b\\([^\x1b]*)\x1b\]8;;\x1b\\`,
)

// StripOSC8 removes all OSC8 hyperlink escape sequences from s, leaving
// only the display text. This is useful for testing, plain-text export,
// or terminals that don't understand OSC8.
func StripOSC8(s string) string {
	// Replace each OSC8 sequence with just the display text
	return osc8Pattern.ReplaceAllString(s, "$2")
}

// ExtractURLs returns all URLs found in OSC8 sequences within s.
// This is useful for link extraction or validation.
func ExtractURLs(s string) []string {
	matches := osc8Pattern.FindAllStringSubmatch(s, -1)
	urls := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 && m[1] != "" {
			urls = append(urls, m[1])
		}
	}
	return urls
}

// FormatOSC8 is a package-level convenience that formats a link using OSC8
// escape sequences. Callers should prefer LinkRenderer.FormatLink for
// configurable behavior.
func FormatOSC8(text, url string) string {
	if url == "" {
		return text
	}
	return osc8Start + url + osc8Sep + text + osc8End
}

// Ensure the strings import is used even if only StripOSC8 references it.
var _ = strings.TrimSpace
