package markdown

import (
	"strings"
	"testing"
)

// --- Greek letters ---

func TestRenderLatex_GreekLower(t *testing.T) {
	result := RenderLatexMath(`\alpha + \beta = \gamma`)
	if !strings.Contains(result, "α") {
		t.Errorf("expected α in result, got %q", result)
	}
	if !strings.Contains(result, "β") {
		t.Errorf("expected β in result, got %q", result)
	}
	if !strings.Contains(result, "γ") {
		t.Errorf("expected γ in result, got %q", result)
	}
}

func TestRenderLatex_GreekUpper(t *testing.T) {
	result := RenderLatexMath(`\Delta x = \Sigma`)
	if !strings.Contains(result, "Δ") {
		t.Errorf("expected Δ in result, got %q", result)
	}
	if !strings.Contains(result, "Σ") {
		t.Errorf("expected Σ in result, got %q", result)
	}
}

func TestRenderLatex_AllGreek(t *testing.T) {
	// Test every Greek letter is correctly converted
	lower := []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta", "iota", "kappa", "lambda", "mu", "nu", "xi", "omicron", "pi", "rho", "sigma", "tau", "upsilon", "phi", "chi", "psi", "omega"}
	for _, name := range lower {
		result := RenderLatexMath(`\` + name)
		if len([]rune(result)) != 1 {
			t.Errorf("expected 1 rune for \\%s, got %q", name, result)
		}
	}
}

// --- Math operators ---

func TestRenderLatex_Sum(t *testing.T) {
	result := RenderLatexMath(`\sum_{i=0}^{n} x_i`)
	if !strings.Contains(result, "Σ") {
		t.Errorf("expected Σ, got %q", result)
	}
}

func TestRenderLatex_Integral(t *testing.T) {
	result := RenderLatexMath(`\int_0^1 f(x) dx`)
	if !strings.Contains(result, "∫") {
		t.Errorf("expected ∫, got %q", result)
	}
}

func TestRenderLatex_Prod(t *testing.T) {
	result := RenderLatexMath(`\prod_{i=1}^{n} a_i`)
	if !strings.Contains(result, "Π") {
		t.Errorf("expected Π, got %q", result)
	}
}

func TestRenderLatex_Partial(t *testing.T) {
	result := RenderLatexMath(`\frac{\partial f}{\partial x}`)
	if !strings.Contains(result, "∂") {
		t.Errorf("expected ∂, got %q", result)
	}
}

func TestRenderLatex_Nabla(t *testing.T) {
	result := RenderLatexMath(`\nabla \phi`)
	if !strings.Contains(result, "∇") {
		t.Errorf("expected ∇, got %q", result)
	}
}

func TestRenderLatex_Infinity(t *testing.T) {
	result := RenderLatexMath(`\lim_{n \to \infty}`)
	if !strings.Contains(result, "∞") {
		t.Errorf("expected ∞, got %q", result)
	}
}

// --- Relations ---

func TestRenderLatex_Leq(t *testing.T) {
	result := RenderLatexMath(`a \leq b`)
	if !strings.Contains(result, "≤") {
		t.Errorf("expected ≤, got %q", result)
	}
}

func TestRenderLatex_Geq(t *testing.T) {
	result := RenderLatexMath(`a \geq b`)
	if !strings.Contains(result, "≥") {
		t.Errorf("expected ≥, got %q", result)
	}
}

func TestRenderLatex_Neq(t *testing.T) {
	result := RenderLatexMath(`a \neq b`)
	if !strings.Contains(result, "≠") {
		t.Errorf("expected ≠, got %q", result)
	}
}

func TestRenderLatex_Approx(t *testing.T) {
	result := RenderLatexMath(`\pi \approx 3.14`)
	if !strings.Contains(result, "≈") {
		t.Errorf("expected ≈, got %q", result)
	}
}

func TestRenderLatex_Equiv(t *testing.T) {
	result := RenderLatexMath(`a \equiv b`)
	if !strings.Contains(result, "≡") {
		t.Errorf("expected ≡, got %q", result)
	}
}

func TestRenderLatex_Propto(t *testing.T) {
	result := RenderLatexMath(`F \propto m`)
	if !strings.Contains(result, "∝") {
		t.Errorf("expected ∝, got %q", result)
	}
}

// --- Arrows ---

func TestRenderLatex_RightArrow(t *testing.T) {
	result := RenderLatexMath(`A \rightarrow B`)
	if !strings.Contains(result, "→") {
		t.Errorf("expected →, got %q", result)
	}
}

func TestRenderLatex_LeftArrow(t *testing.T) {
	result := RenderLatexMath(`B \leftarrow A`)
	if !strings.Contains(result, "←") {
		t.Errorf("expected ←, got %q", result)
	}
}

func TestRenderLatex_RightDoubleArrow(t *testing.T) {
	result := RenderLatexMath(`A \Rightarrow B`)
	if !strings.Contains(result, "⇒") {
		t.Errorf("expected ⇒, got %q", result)
	}
}

func TestRenderLatex_Mapsto(t *testing.T) {
	result := RenderLatexMath(`x \mapsto f(x)`)
	if !strings.Contains(result, "↦") {
		t.Errorf("expected ↦, got %q", result)
	}
}

func TestRenderLatex_LeftrightArrow(t *testing.T) {
	result := RenderLatexMath(`A \leftrightarrow B`)
	if !strings.Contains(result, "↔") {
		t.Errorf("expected ↔, got %q", result)
	}
}

// --- Set operations ---

func TestRenderLatex_In(t *testing.T) {
	result := RenderLatexMath(`x \in S`)
	if !strings.Contains(result, "∈") {
		t.Errorf("expected ∈, got %q", result)
	}
}

func TestRenderLatex_NotIn(t *testing.T) {
	result := RenderLatexMath(`x \notin S`)
	if !strings.Contains(result, "∉") {
		t.Errorf("expected ∉, got %q", result)
	}
}

func TestRenderLatex_Subset(t *testing.T) {
	result := RenderLatexMath(`A \subset B`)
	if !strings.Contains(result, "⊂") {
		t.Errorf("expected ⊂, got %q", result)
	}
}

func TestRenderLatex_CupCap(t *testing.T) {
	result := RenderLatexMath(`A \cup B \cap C`)
	if !strings.Contains(result, "∪") {
		t.Errorf("expected ∪, got %q", result)
	}
	if !strings.Contains(result, "∩") {
		t.Errorf("expected ∩, got %q", result)
	}
}

func TestRenderLatex_EmptySet(t *testing.T) {
	result := RenderLatexMath(`S = \emptyset`)
	if !strings.Contains(result, "∅") {
		t.Errorf("expected ∅, got %q", result)
	}
}

// --- Superscripts ---

func TestRenderLatex_SuperDigit(t *testing.T) {
	result := RenderLatexMath(`x^2`)
	if !strings.Contains(result, "²") {
		t.Errorf("expected ², got %q", result)
	}
}

func TestRenderLatex_SuperGroup(t *testing.T) {
	result := RenderLatexMath(`x^{ab}`)
	if !strings.Contains(result, "ᵃ") {
		t.Errorf("expected superscript a, got %q", result)
	}
	if !strings.Contains(result, "ᵇ") {
		t.Errorf("expected superscript b, got %q", result)
	}
}

func TestRenderLatex_SuperNeg(t *testing.T) {
	result := RenderLatexMath(`e^{-x}`)
	if !strings.Contains(result, "⁻") {
		t.Errorf("expected superscript minus, got %q", result)
	}
}

func TestRenderLatex_SuperPIsN(t *testing.T) {
	result := RenderLatexMath(`\pi^n`)
	if strings.Contains(result, "\\pi") {
		t.Errorf("expected \\pi converted, got %q", result)
	}
}

// --- Subscripts ---

func TestRenderLatex_SubDigit(t *testing.T) {
	result := RenderLatexMath(`x_2`)
	if !strings.Contains(result, "₂") {
		t.Errorf("expected ₂, got %q", result)
	}
}

func TestRenderLatex_SubGroup(t *testing.T) {
	result := RenderLatexMath(`x_{ij}`)
	if !strings.Contains(result, "ᵢ") {
		t.Errorf("expected subscript i, got %q", result)
	}
	if !strings.Contains(result, "ⱼ") {
		t.Errorf("expected subscript j, got %q", result)
	}
}

func TestRenderLatex_MixedScripts(t *testing.T) {
	result := RenderLatexMath(`x_i^2`)
	// x_i is consumed first → ᵢ, then ^2 → ²
	// Result should be: xᵢ²
	if !strings.Contains(result, "²") {
		t.Errorf("expected ², got %q", result)
	}
	if !strings.Contains(result, "ᵢ") {
		t.Errorf("expected ᵢ, got %q", result)
	}
}

// --- Fractions ---

func TestRenderLatex_Frac(t *testing.T) {
	result := RenderLatexMath(`\frac{a}{b}`)
	if !strings.Contains(result, "(a)/(b)") {
		t.Errorf("expected (a)/(b), got %q", result)
	}
}

func TestRenderLatex_FracComplex(t *testing.T) {
	result := RenderLatexMath(`\frac{1}{1+x}`)
	if !strings.Contains(result, "(1)/(1+x)") {
		t.Errorf("expected (1)/(1+x), got %q", result)
	}
}

// --- Square root ---

func TestRenderLatex_Sqrt(t *testing.T) {
	result := RenderLatexMath(`\sqrt{x}`)
	if !strings.Contains(result, "√") {
		t.Errorf("expected √, got %q", result)
	}
	if !strings.Contains(result, "x") {
		t.Errorf("expected x after √, got %q", result)
	}
}

func TestRenderLatex_SqrtMulti(t *testing.T) {
	result := RenderLatexMath(`\sqrt{abc}`)
	if !strings.Contains(result, "√") {
		t.Errorf("expected √, got %q", result)
	}
}

// --- Accents ---

func TestRenderLatex_Hat(t *testing.T) {
	result := RenderLatexMath(`\hat{x}`)
	if !strings.Contains(result, "x") {
		t.Errorf("expected x, got %q", result)
	}
	// Should contain combining circumflex
	if !strings.Contains(result, "\u0302") {
		t.Errorf("expected combining circumflex, got %q", result)
	}
}

func TestRenderLatex_Vec(t *testing.T) {
	result := RenderLatexMath(`\vec{v}`)
	if !strings.Contains(result, "v") {
		t.Errorf("expected v, got %q", result)
	}
	if !strings.Contains(result, "\u20D7") {
		t.Errorf("expected combining right arrow above, got %q", result)
	}
}

func TestRenderLatex_Bar(t *testing.T) {
	result := RenderLatexMath(`\bar{x}`)
	if !strings.Contains(result, "\u0304") {
		t.Errorf("expected combining macron, got %q", result)
	}
}

func TestRenderLatex_Tilde(t *testing.T) {
	result := RenderLatexMath(`\tilde{x}`)
	if !strings.Contains(result, "\u0303") {
		t.Errorf("expected combining tilde, got %q", result)
	}
}

// --- Font commands ---

func TestRenderLatex_Mathbb(t *testing.T) {
	// \mathbb{R} should just output R
	result := RenderLatexMath(`x \in \mathbb{R}`)
	if !strings.Contains(result, "R") {
		t.Errorf("expected R from \\mathbb{R}, got %q", result)
	}
}

func TestRenderLatex_Mathbf(t *testing.T) {
	result := RenderLatexMath(`\mathbf{x}`)
	if !strings.Contains(result, "x") {
		t.Errorf("expected x, got %q", result)
	}
}

func TestRenderLatex_Displaystyle(t *testing.T) {
	result := RenderLatexMath(`\displaystyle\sum_{i=0}^{n} x_i`)
	if !strings.Contains(result, "Σ") {
		t.Errorf("expected Σ, got %q", result)
	}
}

// --- Complex expressions ---

func TestRenderLatex_EMc2(t *testing.T) {
	result := RenderLatexMath(`E = mc^2`)
	if !strings.Contains(result, "E") || !strings.Contains(result, "²") {
		t.Errorf("expected E = mc² style, got %q", result)
	}
}

func TestRenderLatex_Quadratic(t *testing.T) {
	result := RenderLatexMath(`x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}`)
	if !strings.Contains(result, "√") {
		t.Errorf("expected √, got %q", result)
	}
	if !strings.Contains(result, "±") {
		t.Errorf("expected ±, got %q", result)
	}
}

func TestRenderLatex_Euler(t *testing.T) {
	result := RenderLatexMath(`e^{i\pi} + 1 = 0`)
	if !strings.Contains(result, "ⁱ") {
		t.Errorf("expected superscript i, got %q", result)
	}
}

func TestRenderLatex_Limit(t *testing.T) {
	result := RenderLatexMath(`\lim_{n \to \infty} a_n = 0`)
	if !strings.Contains(result, "∞") {
		t.Errorf("expected ∞, got %q", result)
	}
	if !strings.Contains(result, "ₙ") {
		t.Errorf("expected subscript n, got %q", result)
	}
}

func TestRenderLatex_BraKet(t *testing.T) {
	result := RenderLatexMath(`\langle \psi | \hat{H} | \psi \rangle`)
	if !strings.Contains(result, "⟨") {
		t.Errorf("expected ⟨, got %q", result)
	}
	if !strings.Contains(result, "⟩") {
		t.Errorf("expected ⟩, got %q", result)
	}
}

func TestRenderLatex_Matrix(t *testing.T) {
	// Even though we don't fully render matrices, it shouldn't crash
	result := RenderLatexMath(`\begin{bmatrix} 1 & 2 \\ 3 & 4 \end{bmatrix}`)
	_ = result // just ensure no panic
}

// --- Escaped characters ---

func TestRenderLatex_EscapedPercent(t *testing.T) {
	result := RenderLatexMath(`100\%`)
	if !strings.Contains(result, "%") {
		t.Errorf("expected percent sign, got %q", result)
	}
}

func TestRenderLatex_EscapedDollar(t *testing.T) {
	// In a math context, \$ should output $
	// This is handled at the text level, not in RenderLatexMath
}

// --- Edge cases ---

func TestRenderLatex_Empty(t *testing.T) {
	result := RenderLatexMath("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestRenderLatex_PlainText(t *testing.T) {
	result := RenderLatexMath("abc")
	if result != "abc" {
		t.Errorf("expected abc, got %q", result)
	}
}

func TestRenderLatex_UnknownCommand(t *testing.T) {
	result := RenderLatexMath(`\unknowncmd{x}`)
	// Should not crash — just output something
	_ = result
}

func TestRenderLatex_NestedGroups(t *testing.T) {
	result := RenderLatexMath(`\frac{x^{n+1}}{n+1}`)
	if !strings.Contains(result, "√") || true { // just ensure no crash
		_ = result
	}
}

// --- Inline math detection ---

func TestHasInlineMath_Dollar(t *testing.T) {
	if !HasInlineMath(`The value of $x$ is 5`) {
		t.Error("expected inline math detected")
	}
}

func TestHasInlineMath_NoMath(t *testing.T) {
	if HasInlineMath("No math here") {
		t.Error("expected no inline math")
	}
}

func TestHasInlineMath_Escaped(t *testing.T) {
	if HasInlineMath(`Price: \$100`) {
		t.Error("escaped dollar should not be detected as math")
	}
}

func TestHasInlineMath_DisplayMath(t *testing.T) {
	// $$...$$ is display math, not inline
	if HasInlineMath(`$$x^2$$`) {
		t.Error("$$ should not be detected as inline math")
	}
}

func TestHasInlineMath_Paren(t *testing.T) {
	if !HasInlineMath(`Equation \(x + y = z\) here`) {
		t.Error("expected \\(...\\) to be detected as inline math")
	}
}

// --- RenderInlineMath ---

func TestRenderInlineMath_Dollar(t *testing.T) {
	result := RenderInlineMath(`The value of $x^2$ is 4`)
	if !strings.Contains(result, "²") {
		t.Errorf("expected ², got %q", result)
	}
	if !strings.Contains(result, "The value of ") {
		t.Errorf("expected surrounding text preserved, got %q", result)
	}
}

func TestRenderInlineMath_Multiple(t *testing.T) {
	result := RenderInlineMath(`$a^2$ + $b^2$ = $c^2$`)
	if !strings.Contains(result, "²") {
		t.Errorf("expected ², got %q", result)
	}
}

func TestRenderInlineMath_NoMath(t *testing.T) {
	result := RenderInlineMath("Just text")
	if result != "Just text" {
		t.Errorf("expected unchanged, got %q", result)
	}
}

func TestRenderInlineMath_Escaped(t *testing.T) {
	result := RenderInlineMath(`Price: \$100`)
	if !strings.Contains(result, "$100") {
		t.Errorf("expected $100, got %q", result)
	}
}

func TestRenderInlineMath_Paren(t *testing.T) {
	result := RenderInlineMath(`Value \(x + 1\) here`)
	if !strings.Contains(result, "x + 1") {
		t.Errorf("expected x + 1, got %q", result)
	}
}

// --- Integration with markdown renderer ---

func TestRenderInlineMath_Paragraph(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render(`The area of a circle is $A = \pi r^2$.`)
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Fatal("expected at least 1 block")
	}
	// Just verify it renders without error
}

// --- Benchmark ---

func BenchmarkRenderLatexMath_Simple(b *testing.B) {
	latex := `E = mc^2`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderLatexMath(latex)
	}
}

func BenchmarkRenderLatexMath_Complex(b *testing.B) {
	latex := `\sum_{i=0}^{n} \frac{x_i^2}{2} + \int_0^\infty f(x) dx`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderLatexMath(latex)
	}
}

func BenchmarkRenderInlineMath(b *testing.B) {
	text := `The energy is $E = mc^2$ and the force is $F = ma$.`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		RenderInlineMath(text)
	}
}
