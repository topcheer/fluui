package component

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── CSS-like Style Engine ───
//
// StyleSheet provides declarative styling for components, inspired by CSS
// but adapted for terminal UIs. It allows defining named style classes that
// can be applied to any component, with inheritance and priority resolution.
//
// Usage:
//
//	ss := NewStyleSheet()
//	ss.Add(".error", StyleDecl{
//	    Fg:    buffer.NamedColor(buffer.NamedRed),
//	    Bold:  true,
//	})
//	ss.Add(".success", StyleDecl{
//	    Fg:    buffer.NamedColor(buffer.NamedGreen),
//	})
//
//	// Apply to a component
//	style := ss.Resolve("error")  // → buffer.Style{Fg: red, Flags: Bold}
//
//	// Combine classes
//	style = ss.Resolve("error large")  // merges .error + .large

// StyleDecl is a declarative style definition where any field can be set
// (non-nil/non-zero) or left unset (to inherit from parent/defaults).
type StyleDecl struct {
	Fg            buffer.Color // foreground color (zero = inherit)
	Bg            buffer.Color // background color (zero = inherit)
	Bold          *bool        // nil = inherit
	Italic        *bool
	Underline     *bool
	Dim           *bool
	Blink         *bool
	Reverse       *bool
	Strikethrough *bool

	// Layout hints (optional, for future padding/margin support)
	PaddingTop    *int
	PaddingBottom *int
	PaddingLeft   *int
	PaddingRight  *int
}

// boolPtr is a helper to create a *bool from a literal.
func boolPtr(v bool) *bool { return &v }

// intPtr is a helper to create a *int from a literal.
func intPtr(v int) *int { return &v }

// ToStyle converts a StyleDecl to a buffer.Style, inheriting from the given
// base style for fields that are not set.
func (d StyleDecl) ToStyle(base buffer.Style) buffer.Style {
	s := base
	if d.Fg.Type != 0 {
		s.Fg = d.Fg
	}
	if d.Bg.Type != 0 {
		s.Bg = d.Bg
	}
	if d.Bold != nil {
		if *d.Bold {
			s.Flags |= buffer.Bold
		} else {
			s.Flags &^= buffer.Bold
		}
	}
	if d.Italic != nil {
		if *d.Italic {
			s.Flags |= buffer.Italic
		} else {
			s.Flags &^= buffer.Italic
		}
	}
	if d.Underline != nil {
		if *d.Underline {
			s.Flags |= buffer.Underline
		} else {
			s.Flags &^= buffer.Underline
		}
	}
	if d.Dim != nil {
		if *d.Dim {
			s.Flags |= buffer.Dim
		} else {
			s.Flags &^= buffer.Dim
		}
	}
	if d.Blink != nil {
		if *d.Blink {
			s.Flags |= buffer.Blink
		} else {
			s.Flags &^= buffer.Blink
		}
	}
	if d.Reverse != nil {
		if *d.Reverse {
			s.Flags |= buffer.Reverse
		} else {
			s.Flags &^= buffer.Reverse
		}
	}
	if d.Strikethrough != nil {
		if *d.Strikethrough {
			s.Flags |= buffer.Strikethrough
		} else {
			s.Flags &^= buffer.Strikethrough
		}
	}
	return s
}

// Padding returns the combined padding from this declaration.
// Returns (0,0,0,0) if no padding is set.
func (d StyleDecl) Padding() (top, bottom, left, right int) {
	if d.PaddingTop != nil {
		top = *d.PaddingTop
	}
	if d.PaddingBottom != nil {
		bottom = *d.PaddingBottom
	}
	if d.PaddingLeft != nil {
		left = *d.PaddingLeft
	}
	if d.PaddingRight != nil {
		right = *d.PaddingRight
	}
	return
}

// styleRule is an internal rule: a class name → declaration with priority.
type styleRule struct {
	class    string
	decl     StyleDecl
	priority int // higher = more specific (earlier in class list = higher)
}

// StyleSheet is a registry of style declarations indexed by class name.
// It supports class-based selectors, inheritance via class composition,
// and priority resolution (later classes override earlier ones).
type StyleSheet struct {
	mu    sync.RWMutex
	rules map[string]StyleDecl // class name → declaration
}

// NewStyleSheet creates an empty style sheet.
func NewStyleSheet() *StyleSheet {
	return &StyleSheet{
		rules: make(map[string]StyleDecl),
	}
}

// Add registers a style declaration for the given class name.
// The class name should include the leading dot (e.g., ".error", ".success").
// If the class already exists, it is overwritten.
func (ss *StyleSheet) Add(className string, decl StyleDecl) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	// Normalize: strip leading dot if present
	name := normalizeClassName(className)
	ss.rules[name] = decl
}

// Remove deletes a class from the style sheet.
func (ss *StyleSheet) Remove(className string) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	name := normalizeClassName(className)
	if _, ok := ss.rules[name]; ok {
		delete(ss.rules, name)
		return true
	}
	return false
}

// Has reports whether a class is defined.
func (ss *StyleSheet) Has(className string) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	name := normalizeClassName(className)
	_, ok := ss.rules[name]
	return ok
}

// Get returns the declaration for a class, or false if not found.
func (ss *StyleSheet) Get(className string) (StyleDecl, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	name := normalizeClassName(className)
	decl, ok := ss.rules[name]
	return decl, ok
}

// Classes returns all registered class names, sorted alphabetically.
func (ss *StyleSheet) Classes() []string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	result := make([]string, 0, len(ss.rules))
	for name := range ss.rules {
		result = append(result, "."+name)
	}
	sort.Strings(result)
	return result
}

// Count returns the number of registered classes.
func (ss *StyleSheet) Count() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return len(ss.rules)
}

// Resolve computes the final buffer.Style by merging declarations for all
// given classes. Classes are space-separated in the input string.
// Later classes override earlier ones (CSS specificity by order).
//
// Example: ss.Resolve("error bold") merges .error + .bold
// The base style is used as the starting point for inheritance.
func (ss *StyleSheet) Resolve(classes string) buffer.Style {
	return ss.ResolveWithBase(classes, buffer.Style{})
}

// ResolveWithBase computes the final style using an explicit base.
func (ss *StyleSheet) ResolveWithBase(classes string, base buffer.Style) buffer.Style {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	result := base
	for _, cls := range parseClasses(classes) {
		if decl, ok := ss.rules[cls]; ok {
			result = decl.ToStyle(result)
		}
	}
	return result
}

// ResolveDecl computes the merged StyleDecl for the given classes.
// This is useful when you need padding/layout hints, not just buffer.Style.
func (ss *StyleSheet) ResolveDecl(classes string) StyleDecl {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	result := StyleDecl{}
	for _, cls := range parseClasses(classes) {
		if decl, ok := ss.rules[cls]; ok {
			result = mergeDecls(result, decl)
		}
	}
	return result
}

// Apply applies resolved styles to a component that supports SetStyle.
// Returns the resolved style.
func (ss *StyleSheet) Apply(c Stylable, classes string) buffer.Style {
	style := ss.Resolve(classes)
	c.SetStyle(style)
	return style
}

// Stylable is an interface for components that expose a SetStyle method.
type Stylable interface {
	SetStyle(buffer.Style)
}

// ─── Theme Integration ───

// StyleClass is a named style class associated with a theme color slot.
// This allows styles to reference theme variables instead of hardcoded colors.
type StyleClass struct {
	Name       string
	ThemeSlot  string // e.g., "accent", "error", "success" (maps to theme.Theme)
	Decl       StyleDecl
}

// StyleSheetFromTheme creates a StyleSheet with default classes derived
// from theme color names. The classes are:
//
//	.primary, .secondary, .accent, .success, .warning, .error, .muted
//	.bold, .italic, .underline, .dim, .reverse
//
// Each maps to the corresponding theme color or text attribute.
func StyleSheetFromTheme(colorFn func(name string) buffer.Color) *StyleSheet {
	ss := NewStyleSheet()

	// Color classes from theme
	colorClasses := []struct {
		class    string
		themeKey string
	}{
		{".primary", "primary"},
		{".secondary", "secondary"},
		{".accent", "accent"},
		{".success", "success"},
		{".warning", "warning"},
		{".error", "error"},
		{".muted", "muted"},
	}
	for _, cc := range colorClasses {
		if colorFn != nil {
			c := colorFn(cc.themeKey)
			if c.Type != 0 {
				ss.Add(cc.class, StyleDecl{Fg: c})
			}
		}
	}

	// Text attribute classes
	ss.Add(".bold", StyleDecl{Bold: boolPtr(true)})
	ss.Add(".italic", StyleDecl{Italic: boolPtr(true)})
	ss.Add(".underline", StyleDecl{Underline: boolPtr(true)})
	ss.Add(".dim", StyleDecl{Dim: boolPtr(true)})
	ss.Add(".reverse", StyleDecl{Reverse: boolPtr(true)})
	ss.Add(".strikethrough", StyleDecl{Strikethrough: boolPtr(true)})

	return ss
}

// ─── Parsing ───

// ParseStyleSheet parses a simple CSS-like text format into a StyleSheet.
//
// Syntax:
//
//	.error {
//	    fg: red
//	    bold: true
//	}
//	.success {
//	    fg: green
//	}
//
// Color names: red, green, blue, yellow, cyan, magenta, white, black,
// bright_red, bright_green, etc.
// Also supports hex: #ff0000
func ParseStyleSheet(text string) (*StyleSheet, error) {
	ss := NewStyleSheet()
	lines := strings.Split(text, "\n")

	var currentClass string
	var decl StyleDecl

	flush := func() {
		if currentClass != "" {
			ss.Add(currentClass, decl)
		}
		currentClass = ""
		decl = StyleDecl{}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") && !strings.Contains(trimmed, ":") {
			// Skip comments and blank lines, but not property lines starting with hex colors
			if strings.HasPrefix(trimmed, "#") && strings.Contains(trimmed, ":") {
				// This is a property with hex color, don't skip
			} else {
				continue
			}
		}

		// Class start
		if strings.HasSuffix(trimmed, "{") {
			flush()
			classParts := strings.Fields(strings.TrimSuffix(trimmed, "{"))
			if len(classParts) > 0 {
				currentClass = classParts[0]
			}
			continue
		}

		// Class end
		if trimmed == "}" {
			flush()
			continue
		}

		// Property
		if currentClass == "" {
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		applyDecl(&decl, key, val)
	}
	flush()

	return ss, nil
}

// applyDecl sets a property on a StyleDecl from string key/value.
func applyDecl(decl *StyleDecl, key, val string) {
	val = strings.TrimSpace(val)
	switch strings.ToLower(key) {
	case "fg", "foreground", "color":
		if c := parseColor(val); c.Type != 0 {
			decl.Fg = c
		}
	case "bg", "background":
		if c := parseColor(val); c.Type != 0 {
			decl.Bg = c
		}
	case "bold":
		decl.Bold = parseBool(val)
	case "italic":
		decl.Italic = parseBool(val)
	case "underline":
		decl.Underline = parseBool(val)
	case "dim":
		decl.Dim = parseBool(val)
	case "blink":
		decl.Blink = parseBool(val)
	case "reverse":
		decl.Reverse = parseBool(val)
	case "strikethrough", "strike":
		decl.Strikethrough = parseBool(val)
	case "padding-top", "pt":
		if n := parseInt(val); n != nil {
			decl.PaddingTop = n
		}
	case "padding-bottom", "pb":
		if n := parseInt(val); n != nil {
			decl.PaddingBottom = n
		}
	case "padding-left", "pl":
		if n := parseInt(val); n != nil {
			decl.PaddingLeft = n
		}
	case "padding-right", "pr":
		if n := parseInt(val); n != nil {
			decl.PaddingRight = n
		}
	}
}

// parseColor parses a color name or hex value into a buffer.Color.
func parseColor(s string) buffer.Color {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" || s == "inherit" || s == "default" || s == "none" {
		return buffer.Color{}
	}
	// Hex color
	if strings.HasPrefix(s, "#") {
		return buffer.Hex(s)
	}
	// Named color
	return buffer.NamedColor(int(parseNamedColor(s)))
}

// parseNamedColor converts a string name to a buffer.NamedColor constant.
func parseNamedColor(s string) int {
	colorMap := map[string]int{
		"black":         buffer.NamedBlack,
		"red":           buffer.NamedRed,
		"green":         buffer.NamedGreen,
		"yellow":        buffer.NamedYellow,
		"blue":          buffer.NamedBlue,
		"magenta":       buffer.NamedMagenta,
		"cyan":          buffer.NamedCyan,
		"white":         buffer.NamedWhite,
		"bright_black":  buffer.NamedBrightBlack,
		"bright_red":    buffer.NamedBrightRed,
		"bright_green":  buffer.NamedBrightGreen,
		"bright_yellow": buffer.NamedBrightYellow,
		"bright_blue":   buffer.NamedBrightBlue,
		"bright_magenta": buffer.NamedBrightMagenta,
		"bright_cyan":   buffer.NamedBrightCyan,
		"bright_white":  buffer.NamedBrightWhite,
	}
	if c, ok := colorMap[s]; ok {
		return c
	}
	return buffer.NamedWhite // fallback (int constant)
}

// parseBool parses "true"/"false"/"yes"/"no"/"1"/"0" to *bool.
func parseBool(s string) *bool {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "yes", "1", "on":
		return boolPtr(true)
	case "false", "no", "0", "off":
		return boolPtr(false)
	}
	return nil
}

// parseInt parses a string to *int.
func parseInt(s string) *int {
	s = strings.TrimSpace(s)
	n := 0
	negative := false
	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if c < '0' || c > '9' {
			return nil
		}
		n = n*10 + int(c-'0')
	}
	if negative {
		n = -n
	}
	return intPtr(n)
}

// ─── Internal Helpers ───

func normalizeClassName(s string) string {
	return strings.TrimPrefix(strings.TrimSpace(strings.ToLower(s)), ".")
}

func parseClasses(s string) []string {
	parts := strings.Fields(s)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		result = append(result, normalizeClassName(p))
	}
	return result
}

// mergeDecls combines two declarations, with `overlay` taking priority.
func mergeDecls(base, overlay StyleDecl) StyleDecl {
	result := base
	if overlay.Fg.Type != 0 {
		result.Fg = overlay.Fg
	}
	if overlay.Bg.Type != 0 {
		result.Bg = overlay.Bg
	}
	if overlay.Bold != nil {
		result.Bold = overlay.Bold
	}
	if overlay.Italic != nil {
		result.Italic = overlay.Italic
	}
	if overlay.Underline != nil {
		result.Underline = overlay.Underline
	}
	if overlay.Dim != nil {
		result.Dim = overlay.Dim
	}
	if overlay.Blink != nil {
		result.Blink = overlay.Blink
	}
	if overlay.Reverse != nil {
		result.Reverse = overlay.Reverse
	}
	if overlay.Strikethrough != nil {
		result.Strikethrough = overlay.Strikethrough
	}
	if overlay.PaddingTop != nil {
		result.PaddingTop = overlay.PaddingTop
	}
	if overlay.PaddingBottom != nil {
		result.PaddingBottom = overlay.PaddingBottom
	}
	if overlay.PaddingLeft != nil {
		result.PaddingLeft = overlay.PaddingLeft
	}
	if overlay.PaddingRight != nil {
		result.PaddingRight = overlay.PaddingRight
	}
	return result
}

// String returns a human-readable representation of the StyleSheet.
func (ss *StyleSheet) String() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return fmt.Sprintf("StyleSheet(%d classes)", len(ss.rules))
}
