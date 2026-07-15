// Package ansi provides ANSI style configuration types for glamour compat.
//
// This mirrors the charm.land/glamour/v2/ansi package's StyleConfig
// and related types used by ggcode's markdown renderer.
package ansi

// StyleConfig holds the complete style configuration for glamour rendering.
type StyleConfig struct {
	Document     Document
	BlockQuote   StyleBlock
	List         StyleBlock
	Table        StyleBlock
	Heading      [6]StyleBlockH
	Hr           StyleBlock
	Code         StyleBlock
	CodeBlock    StyleBlock
	Emph         StyleBlock
	Strong       StyleBlock
	Strikethrough StyleBlock
	Link         StyleBlock
	LinkText     StyleBlock
	Image        StyleBlock
	Paragraph    StyleBlock
}

// Document is the root document style.
type Document struct {
	Margin      *uint
	BlockPrefix string
	BlockSuffix string
	Color       *string
	BackgroundColor *string
}

// StyleBlock is a generic style block with prefix/suffix and color attributes.
type StyleBlock struct {
	Prefix          string
	Suffix          string
	Color           *string
	BackgroundColor *string
	Underline       *bool
	Bold            *bool
	Upper           *bool
	Italic          *bool
	CrossedOut      *bool
	Faint           *bool
	Conceal         *bool
	Blink           *bool
	Reverse         *bool
}

// StyleBlockH extends StyleBlock for headings (adds level-specific styling).
type StyleBlockH struct {
	StyleBlock
	Level int
}

// StyleCodeBlockConfig holds code block configuration with optional Chroma syntax highlighting.
type StyleCodeBlockConfig = StyleBlock // alias for compat

// ChromaStyleConfig holds Chroma syntax highlighting styles.
type ChromaStyleConfig struct {
	Text              StyleBlock
	Error             StyleBlock
	Comment           StyleBlock
	CommentPreproc    StyleBlock
	Keyword           StyleBlock
	KeywordReserved   StyleBlock
	KeywordNamespace  StyleBlock
	Operator          StyleBlock
	Punctuation       StyleBlock
	Name              StyleBlock
	NameTag           StyleBlock
	NameAttribute     StyleBlock
	NameFunction      StyleBlock
	LiteralString     StyleBlock
	LiteralStringEscape StyleBlock
}

// AddChromaStyles returns a StyleConfig with code block Chroma styling.
func AddChromaStyles(sc *StyleConfig, chroma ChromaStyleConfig) {
	if sc == nil {
		return
	}
	// Attach chroma to CodeBlock via a custom field (not in original StyleBlock)
	// In real glamour, CodeBlock has a Chroma field. We use a package-level map.
	chromaStyles[sc] = chroma
}

// Package-level map to associate StyleConfig with ChromaStyleConfig.
var chromaStyles = map[*StyleConfig]ChromaStyleConfig{}

// Ptr returns a pointer to the given string (convenience helper).
func Ptr(s string) *string { return &s }

// PtrBool returns a pointer to the given bool.
func PtrBool(b bool) *bool { return &b }

// PtrUint returns a pointer to the given uint.
func PtrUint(u uint) *uint { return &u }