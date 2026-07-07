package markdown

import (
	"strings"
	"unicode"

	"github.com/topcheer/fluui/internal/buffer"
)

// ---------------------------------------------------------------------------
// LaTeX Math to Unicode Renderer
// ---------------------------------------------------------------------------
//
// Converts LaTeX math expressions to Unicode text for terminal display.
// Supports:
//   - Greek letters (\alpha → α, \beta → β, ...)
//   - Math operators (\sum → Σ, \int → ∫, \prod → Π, ...)
//   - Relations (\leq → ≤, \geq → ≥, \neq → ≠, \approx → ≈)
//   - Arrows (\rightarrow → →, \Rightarrow → ⇒, \mapsto → ↦)
//   - Sets (\in → ∈, \subset → ⊂, \cup → ∪, \cap → ∩)
//   - Superscripts (x^2 → x², x^{ab} → x^ab)
//   - Subscripts (x_2 → x₂, x_{ij} → x_ij)
//   - Fractions (\frac{a}{b} → a/b)
//   - Square root (\sqrt{x} → √x̄)
//   - Accents (\hat{x} → x̂)
//
// This is a pure Go implementation — no external dependencies.

// latexGreek maps LaTeX Greek letter commands to Unicode runes.
var latexGreek = map[string]rune{
	// Lowercase
	"alpha":   'α',
	"beta":    'β',
	"gamma":   'γ',
	"delta":   'δ',
	"epsilon": 'ε',
	"varepsilon": 'ε',
	"zeta":    'ζ',
	"eta":     'η',
	"theta":   'θ',
	"vartheta": 'ϑ',
	"iota":    'ι',
	"kappa":   'κ',
	"lambda":  'λ',
	"mu":      'μ',
	"nu":      'ν',
	"xi":      'ξ',
	"omicron": 'ο',
	"pi":      'π',
	"varpi":   'ϖ',
	"rho":     'ρ',
	"varrho":  'ϱ',
	"sigma":   'σ',
	"varsigma": 'ς',
	"tau":     'τ',
	"upsilon": 'υ',
	"phi":     'φ',
	"varphi":  'ϕ',
	"chi":     'χ',
	"psi":     'ψ',
	"omega":   'ω',
	// Uppercase
	"Alpha":   'Α',
	"Beta":    'Β',
	"Gamma":   'Γ',
	"Delta":   'Δ',
	"Epsilon": 'Ε',
	"Zeta":    'Ζ',
	"Eta":     'Η',
	"Theta":   'Θ',
	"Iota":    'Ι',
	"Kappa":   'Κ',
	"Lambda":  'Λ',
	"Mu":      'Μ',
	"Nu":      'Ν',
	"Xi":      'Ξ',
	"Omicron": 'Ο',
	"Pi":      'Π',
	"Rho":     'Ρ',
	"Sigma":   'Σ',
	"Tau":     'Τ',
	"Upsilon": 'Υ',
	"Phi":     'Φ',
	"Chi":     'Χ',
	"Psi":     'Ψ',
	"Omega":   'Ω',
}

// latexSymbol maps LaTeX symbol commands to Unicode runes.
var latexSymbol = map[string]rune{
	// Operators
	"sum":      'Σ',
	"int":      '∫',
	"oint":     '∮',
	"prod":     'Π',
	"coprod":   '∐',
	"bigcap":   '⋂',
	"bigcup":   '⋃',
	"bigvee":   '⋁',
	"bigwedge": '⋀',
	"bigoplus": '⨁',
	"bigotimes": '⨂',
	// Relations
	"leq":      '≤',
	"le":       '≤',
	"geq":      '≥',
	"ge":       '≥',
	"neq":      '≠',
	"ne":       '≠',
	"approx":   '≈',
	"equiv":    '≡',
	"sim":      '∼',
	"simeq":    '≃',
	"cong":     '≅',
	"propto":   '∝',
	// Arrows
	"rightarrow": '→',
	"to":         '→',
	"leftarrow":  '←',
	"gets":       '←',
	"leftrightarrow": '↔',
	"Rightarrow":    '⇒',
	"Leftarrow":     '⇐',
	"Leftrightarrow": '⇔',
	"mapsto":        '↦',
	"uparrow":       '↑',
	"downarrow":     '↓',
	"updownarrow":   '↕',
	"rightharpoonup": '⇀',
	"leftharpoonup":  '⇁',
	// Sets
	"in":        '∈',
	"notin":     '∉',
	"ni":        '∋',
	"subset":    '⊂',
	"supset":    '⊃',
	"subseteq":  '⊆',
	"supseteq":  '⊇',
	"cup":       '∪',
	"cap":       '∩',
	"emptyset":  '∅',
	"varnothing": '∅',
	// Misc symbols
	"infty":     '∞',
	"partial":   '∂',
	"nabla":     '∇',
	"forall":    '∀',
	"exists":    '∃',
	"nexists":   '∄',
	"neg":       '¬',
	"lnot":      '¬',
	"angle":     '∠',
	"perp":      '⊥',
	"parallel":  '∥',
	"circ":      '∘',
	"bullet":    '•',
	"dagger":    '†',
	"ddagger":   '‡',
	"ldots":     '…',
	"cdots":     '⋯',
	"vdots":     '⋮',
	"ddots":     '⋱',
	"prime":     '′',
	"backslash": '\\',
	"vert":      '|',
	"vertvert":  '‖',
	"ast":       '∗',
	"star":      '⋆',
	"times":     '×',
	"div":       '÷',
	"pm":        '±',
	"mp":        '∓',
	"cdot":      '·',
	"otimes":    '⊗',
	"oplus":     '⊕',
	"ominus":    '⊖',
	"odot":      '⊙',
	"lceil":     '⌈',
	"rceil":     '⌉',
	"lfloor":    '⌊',
	"rfloor":    '⌋',
	"langle":    '⟨',
	"rangle":    '⟩',
	"hbar":      'ℏ',
	"ell":       'ℓ',
	"Re":        'ℜ',
	"Im":        'ℑ',
	"aleph":     'ℵ',
	"imath":     'ı',
	"jmath":     'ȷ',
	// spacing — handled specially, not as single runes
	"quad":  ' ',
}

// latexAccent maps accent commands to combining characters.
var latexAccent = map[string]rune{
	"hat":    '\u0302', // ̂
	"check":  '\u030C', // ̌
	"breve":  '\u0306', // ̆
	"acute":  '\u0301', // ́
	"grave":  '\u0300', // ̀
	"tilde":  '\u0303', // ̃
	"bar":    '\u0304', // ̄
	"vec":    '\u20D7', // ⃗
	"dot":    '\u0307', // ̇
	"ddot":   '\u0308', // ̈
}

// latexFontCmds are font commands that are silently consumed (output their argument).
var latexFontCmds = map[string]bool{
	"textbf": true, "textit": true, "textrm": true, "textsf": true,
	"texttt": true, "mathbf": true, "mathit": true, "mathrm": true,
	"mathsf": true, "mathtt": true, "mathbb": true, "mathcal": true,
	"mathfrak": true, "boldsymbol": true, "bm": true, "pmb": true,
	"text": true, "operatorname": true, "displaystyle": true,
	"limits": true, "nolimits": true, "left": true, "right": true,
	"big": true, "Big": true, "bigg": true, "Bigg": true,
	"bigl": true, "bigr": true, "Bigl": true, "Bigr": true,
}

// latexSuperscript maps digits/letters to Unicode superscript equivalents.
var latexSuperscript = map[rune]rune{
	'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
	'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
	'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
	'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ',
	'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ', 'o': 'ᵒ',
	'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ',
	'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
	'A': 'ᴬ', 'B': 'ᴮ', 'D': 'ᴰ', 'E': 'ᴱ', 'G': 'ᴳ',
	'H': 'ᴴ', 'I': 'ᴵ', 'J': 'ᴶ', 'K': 'ᴷ', 'L': 'ᴸ',
	'M': 'ᴹ', 'N': 'ᴺ', 'O': 'ᴼ', 'P': 'ᴾ', 'R': 'ᴿ',
	'T': 'ᵀ', 'U': 'ᵁ', 'V': 'ⱽ', 'W': 'ᵂ',
}

// latexSubscript maps digits/letters to Unicode subscript equivalents.
var latexSubscript = map[rune]rune{
	'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
	'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
	'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
	'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
	'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
	'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
	'v': 'ᵥ', 'x': 'ₓ',
}

// superscriptRune converts a rune to its Unicode superscript if available.
func superscriptRune(r rune) (rune, bool) {
	sr, ok := latexSuperscript[r]
	return sr, ok
}

// subscriptRune converts a rune to its Unicode subscript if available.
func subscriptRune(r rune) (rune, bool) {
	sr, ok := latexSubscript[r]
	return sr, ok
}

// RenderLatexMath converts a LaTeX math expression to Unicode text.
// Returns the Unicode representation of the math expression.
func RenderLatexMath(latex string) string {
	var result strings.Builder
	result.Grow(len(latex) * 2)
	renderLatexToBuilder(latex, &result)
	return result.String()
}

// renderLatexToBuilder parses a LaTeX math expression and writes the Unicode
// result directly into the provided builder. This avoids the intermediate
// string allocation when called from RenderInlineMath.
func renderLatexToBuilder(latex string, result *strings.Builder) {
	p := &latexParser{input: latex}

	for !p.atEnd() {
		p.consume(result)
	}
}

// latexParser is a simple state machine for parsing LaTeX math.
type latexParser struct {
	input string
	pos   int
}

func (p *latexParser) atEnd() bool {
	return p.pos >= len(p.input)
}

func (p *latexParser) peek() byte {
	if p.pos < len(p.input) {
		return p.input[p.pos]
	}
	return 0
}

func (p *latexParser) advance() byte {
	if p.pos < len(p.input) {
		b := p.input[p.pos]
		p.pos++
		return b
	}
	return 0
}

// consume processes the next token and writes to the result.
func (p *latexParser) consume(result *strings.Builder) {
	ch := p.peek()

	switch ch {
	case '\\':
		p.consumeCommand(result)
	case '^':
		p.advance() // skip ^
		p.consumeScript(result, true) // superscript
	case '_':
		p.advance() // skip _
		p.consumeScript(result, false) // subscript
	case '{':
		p.advance() // skip {
		// Groups are transparent — just consume their content
		p.consumeGroup(result)
	case '}':
		p.advance() // skip stray }
	case '$':
		p.advance() // skip $ markers (handled by caller)
	default:
		// Regular character
		p.advance()
		result.WriteByte(ch)
	}
}

// consumeCommand processes a \command and writes the Unicode equivalent.
func (p *latexParser) consumeCommand(result *strings.Builder) {
	p.advance() // skip backslash

	// Check for \\ (line break)
	if p.peek() == '\\' {
		p.advance()
		result.WriteString("  ") // line break → spaces
		return
	}

	// Check for single-char commands like \, \%
	ch := p.peek()
	if !unicode.IsLetter(rune(ch)) {
		p.advance()
		switch ch {
		case ',':
			result.WriteByte(' ')
		case ';':
			result.WriteByte(' ')
		case ':':
			result.WriteByte(' ')
		case '%':
			result.WriteByte('%')
		case '#':
			result.WriteByte('#')
		case '&':
			result.WriteByte('&')
		case '_':
			result.WriteByte('_')
		case '{':
			result.WriteByte('{')
		case '}':
			result.WriteByte('}')
		case ' ':
			result.WriteByte(' ')
		default:
			result.WriteByte(ch)
		}
		return
	}

	// Read command name (letters only)
	var cmdName strings.Builder
	for !p.atEnd() && unicode.IsLetter(rune(p.peek())) {
		cmdName.WriteByte(p.advance())
	}
	cmd := cmdName.String()

	// Skip optional star (e.g., \sum*)
	if p.peek() == '*' {
		p.advance()
	}

	// Skip optional [...] arguments
	if p.peek() == '[' {
		p.skipBracket()
	}

	// Look up in maps
	if r, ok := latexGreek[cmd]; ok {
		result.WriteRune(r)
		return
	}
	if r, ok := latexSymbol[cmd]; ok {
		if r != 0 {
			result.WriteRune(r)
		}
		return
	}
	if accent, ok := latexAccent[cmd]; ok {
		p.consumeAccent(result, accent)
		return
	}
	if cmd == "frac" {
		p.consumeFrac(result)
		return
	}
	if cmd == "sqrt" {
		p.consumeSqrt(result)
		return
	}
	if cmd == "not" {
		// \not\in → ∉ — simplified: just output a strikethrough
		// Skip the next command and output negated version
		if p.peek() == '\\' {
			// Read next command
			p.advance()
			var next strings.Builder
			for !p.atEnd() && unicode.IsLetter(rune(p.peek())) {
				next.WriteByte(p.advance())
			}
			nc := next.String()
			switch nc {
			case "in":
				result.WriteRune('∉')
			case "subset":
				result.WriteRune('⊄')
			case "supset":
				result.WriteRune('⊅')
			default:
				result.WriteString("∉") // best effort
			}
		}
		return
	}
	if latexFontCmds[cmd] {
		// Font commands: consume argument and output content
		p.skipSpaces()
		if p.peek() == '{' {
			p.advance()
			p.consumeGroup(result)
		}
		return
	}

	// Unknown command — output as-is with backslash
	result.WriteByte('\\')
	result.WriteString(cmd)
}

// consumeGroup reads until matching }, consuming content recursively.
func (p *latexParser) consumeGroup(result *strings.Builder) {
	depth := 1
	for !p.atEnd() && depth > 0 {
		ch := p.peek()
		if ch == '{' {
			depth++
			p.advance()
		} else if ch == '}' {
			depth--
			p.advance()
			if depth == 0 {
				return
			}
		} else {
			p.consume(result)
			continue
		}
	}
}

// consumeScript handles ^{...} or _{...} and single-char scripts.
func (p *latexParser) consumeScript(result *strings.Builder, isSuper bool) {
	p.skipSpaces()

	if p.peek() == '{' {
		p.advance()
		// Read group content as subscript/superscript
		var group strings.Builder
		p.consumeGroup(&group)
		text := group.String()

		for _, r := range text {
			if isSuper {
				if sr, ok := superscriptRune(r); ok {
					result.WriteRune(sr)
				} else {
					// No unicode superscript — use ^ notation
					result.WriteByte('^')
					result.WriteRune(r)
				}
			} else {
				if sr, ok := subscriptRune(r); ok {
					result.WriteRune(sr)
				} else {
					result.WriteByte('_')
					result.WriteRune(r)
				}
			}
		}
		return
	}

	// Single character
	ch := rune(p.advance())
	if isSuper {
		if sr, ok := superscriptRune(ch); ok {
			result.WriteRune(sr)
		} else {
			result.WriteByte('^')
			result.WriteRune(ch)
		}
	} else {
		if sr, ok := subscriptRune(ch); ok {
			result.WriteRune(sr)
		} else {
			result.WriteByte('_')
			result.WriteRune(ch)
		}
	}
}

// consumeFrac handles \frac{num}{den} → num/den
func (p *latexParser) consumeFrac(result *strings.Builder) {
	p.skipSpaces()

	var num, den strings.Builder
	if p.peek() == '{' {
		p.advance()
		p.consumeGroup(&num)
	}
	p.skipSpaces()
	if p.peek() == '{' {
		p.advance()
		p.consumeGroup(&den)
	}

	result.WriteString("(")
	result.WriteString(num.String())
	result.WriteString(")/(")
	result.WriteString(den.String())
	result.WriteString(")")
}

// consumeSqrt handles \sqrt{x} → √x
func (p *latexParser) consumeSqrt(result *strings.Builder) {
	p.skipSpaces()
	result.WriteRune('√')

	// Check for optional [n] (nth root) — skip it
	if p.peek() == '[' {
		p.skipBracket()
	}

	if p.peek() == '{' {
		p.advance()
		var inner strings.Builder
		p.consumeGroup(&inner)

		// Add overline if content is more than 1 char
		text := inner.String()
		if len([]rune(text)) > 1 {
			result.WriteString(text)
			// Add combining overline for each char
			for range text {
				result.WriteRune('\u0305') // combining overline
			}
		} else {
			result.WriteString(text)
		}
		return
	}

	// Single char
	ch := rune(p.advance())
	result.WriteRune(ch)
}

// consumeAccent handles \hat{x} → x̂
func (p *latexParser) consumeAccent(result *strings.Builder, accent rune) {
	p.skipSpaces()
	if p.peek() == '{' {
		p.advance()
		var inner strings.Builder
		p.consumeGroup(&inner)
		result.WriteString(inner.String())
		result.WriteRune(accent)
		return
	}
	ch := rune(p.advance())
	result.WriteRune(ch)
	result.WriteRune(accent)
}

// skipSpaces consumes whitespace.
func (p *latexParser) skipSpaces() {
	for !p.atEnd() && (p.peek() == ' ' || p.peek() == '\t') {
		p.advance()
	}
}

// skipBracket consumes a [...] group.
func (p *latexParser) skipBracket() {
	if p.peek() != '[' {
		return
	}
	p.advance() // skip [
	depth := 1
	for !p.atEnd() && depth > 0 {
		ch := p.peek()
		if ch == '[' {
			depth++
		} else if ch == ']' {
			depth--
		}
		if depth > 0 {
			p.advance()
		} else {
			p.advance() // skip ]
		}
	}
}

// HasInlineMath checks if a text string contains inline math ($...$ or \(...\)).
func HasInlineMath(text string) bool {
	return findInlineMath(text) >= 0
}

// findInlineMath returns the index of the first $ that starts inline math, or -1.
func findInlineMath(text string) int {
	// Fast rejection: if text contains neither '$' nor '\(' (actually just '\'),
	// there can't be any inline math. strings.IndexByte uses SIMD/optimized search.
	dollarIdx := strings.IndexByte(text, '$')
	backslashIdx := strings.IndexByte(text, '\\')
	if dollarIdx < 0 && backslashIdx < 0 {
		return -1
	}
	for i := 0; i < len(text); i++ {
		if text[i] == '$' {
			// Check it's not \$ (escaped)
			if i > 0 && text[i-1] == '\\' {
				continue
			}
			// Check it's not $$ (display math)
			if i+1 < len(text) && text[i+1] == '$' {
				i++ // skip both $ characters
				continue
			}
			// Must have a closing $
			for j := i + 1; j < len(text); j++ {
				if text[j] == '$' && text[j-1] != '\\' {
					return i
				}
			}
		}
		// Check \(...\) pattern
		if i+1 < len(text) && text[i] == '\\' && text[i+1] == '(' {
			return i
		}
	}
	return -1
}

// RenderInlineMath replaces inline math ($...$, \(...\)) in text with Unicode.
func RenderInlineMath(text string) string {
	var result strings.Builder
	result.Grow(len(text) * 2)
	renderInlineMathToBuilder(text, &result)
	return result.String()
}

// renderInlineMathToBuilder writes inline math result into a caller-provided builder.
func renderInlineMathToBuilder(text string, result *strings.Builder) {
	i := 0
	for i < len(text) {
		// Check for escaped \$
		if i+1 < len(text) && text[i] == '\\' && text[i+1] == '$' {
			result.WriteByte('$')
			i += 2
			continue
		}

		// Check for inline math $...$
		if text[i] == '$' && i+1 < len(text) && text[i+1] != '$' {
			end := -1
			for j := i + 1; j < len(text); j++ {
				if text[j] == '$' && text[j-1] != '\\' {
					end = j
					break
				}
			}
			if end > 0 {
				latex := text[i+1 : end]
				renderLatexToBuilder(latex, result)
				i = end + 1
				continue
			}
		}

		// Check for \(...\) pattern
		if i+1 < len(text) && text[i] == '\\' && text[i+1] == '(' {
			end := strings.Index(text[i+2:], "\\)")
			if end >= 0 {
				latex := text[i+2 : i+2+end]
				renderLatexToBuilder(latex, result)
				i = i + 2 + end + 2
				continue
			}
		}

		result.WriteByte(text[i])
		i++
	}
}

// RenderMathToCells converts a LaTeX math expression to styled cells.
func RenderMathToCells(latex string, fg buffer.Color) []buffer.Cell {
	unicode := RenderLatexMath(latex)
	cells := make([]buffer.Cell, 0, len(unicode))
	for _, ch := range unicode {
		w := buffer.RuneWidth(ch)
		cells = append(cells, buffer.Cell{
			Rune:  ch,
			Width: uint8(w),
			Fg:    fg,
		})
	}
	return cells
}
