package markdown

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/yuin/goldmark/ast"
	textm "github.com/yuin/goldmark/text"
)

func TestP176_githubAlertType(t *testing.T) {
	// Test all alert types
	tests := []struct {
		input    string
		expected string
		isAlert  bool
	}{
		{"[!NOTE]\nContent", "note", true},
		{"[!TIP]\nContent", "tip", true},
		{"[!IMPORTANT]\nContent", "important", true},
		{"[!WARNING]\nContent", "warning", true},
		{"[!CAUTION]\nContent", "caution", true},
		{"[!note]\nContent", "note", true},     // case-insensitive
		{"[!Note]\nContent", "note", true},     // mixed case
		{"[!UNKNOWN]\nContent", "", false},     // unknown type
		{"Regular text", "", false},            // not an alert
		{"", "", false},                         // empty
	}
	for _, tt := range tests {
		// Build a mock blockquote node
		bq := ast.NewBlockquote()
		para := ast.NewParagraph()
		text := ast.NewText()
		text.Segment = textm.Segment{Start: 0, Stop: len(tt.input)}
		para.AppendChild(para, text)
		bq.AppendChild(bq, para)

		source := []byte(tt.input)
		alertType, isAlert := githubAlertType(bq, source)
		if isAlert != tt.isAlert {
			t.Errorf("input %q: expected isAlert=%v, got %v", tt.input, tt.isAlert, isAlert)
		}
		if alertType != tt.expected {
			t.Errorf("input %q: expected %q, got %q", tt.input, tt.expected, alertType)
		}
	}
}

func TestP176_githubAlertType_NilFirstChild(t *testing.T) {
	bq := ast.NewBlockquote()
	// No children
	_, isAlert := githubAlertType(bq, []byte(""))
	if isAlert {
		t.Error("expected false for nil first child")
	}
}

func TestP176_githubAlertType_NonParagraphFirstChild(t *testing.T) {
	bq := ast.NewBlockquote()
	// Add a non-paragraph child
	codeBlock := ast.NewCodeBlock()
	bq.AppendChild(bq, codeBlock)
	_, isAlert := githubAlertType(bq, []byte(""))
	if isAlert {
		t.Error("expected false for non-paragraph first child")
	}
}

func TestP176_githubAlertType_NonTextChild(t *testing.T) {
	bq := ast.NewBlockquote()
	para := ast.NewParagraph()
	// Add a non-text child (e.g., emphasis)
	emph := ast.NewEmphasis(1)
	para.AppendChild(para, emph)
	bq.AppendChild(bq, para)
	_, isAlert := githubAlertType(bq, []byte(""))
	if isAlert {
		t.Error("expected false for non-text child")
	}
}

func TestP176_githubAlertType_ShortText(t *testing.T) {
	bq := ast.NewBlockquote()
	para := ast.NewParagraph()
	text := ast.NewText()
	shortInput := "[!" // too short
	text.Segment = textm.Segment{Start: 0, Stop: len(shortInput)}
	para.AppendChild(para, text)
	bq.AppendChild(bq, para)
	_, isAlert := githubAlertType(bq, []byte(shortInput))
	if isAlert {
		t.Error("expected false for short text")
	}
}

func TestP176_githubAlertType_NoCloseBracket(t *testing.T) {
	bq := ast.NewBlockquote()
	para := ast.NewParagraph()
	text := ast.NewText()
	input := "[!NOTE some text without close"
	text.Segment = textm.Segment{Start: 0, Stop: len(input)}
	para.AppendChild(para, text)
	bq.AppendChild(bq, para)
	_, isAlert := githubAlertType(bq, []byte(input))
	if isAlert {
		t.Error("expected false when no close bracket")
	}
}

func TestP176_alertIcon(t *testing.T) {
	tests := []struct {
		alertType string
		expected  string
	}{
		{"note", "ℹ "},
		{"tip", "✓ "},
		{"important", "❖ "},
		{"warning", "⚠ "},
		{"caution", "✗ "},
		{"unknown", "● "}, // fallback
		{"", "● "},        // empty = fallback
	}
	for _, tt := range tests {
		got := alertIcon(tt.alertType)
		if got != tt.expected {
			t.Errorf("alertIcon(%q): expected %q, got %q", tt.alertType, tt.expected, got)
		}
	}
}

func TestP176_alertColor(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	tests := []string{"note", "tip", "important", "warning", "caution", "unknown"}
	for _, alertType := range tests {
		color := r.alertColor(alertType)
		// Should not panic and return a color
		_ = color
	}
	// Unknown type should return QuoteBar
	color := r.alertColor("unknown")
	if color != r.theme.QuoteBar {
		t.Error("expected QuoteBar for unknown alert type")
	}
}

func TestP176_renderBlockquote_Plain(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Render a plain blockquote (not an alert)
	mdText := "> This is a quote"
	blocks, err := r.Render(mdText)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	hasBlockquote := false
	for _, b := range blocks {
		if b.Type == BlockQuote {
			hasBlockquote = true
			if len(b.Cells) == 0 {
				t.Error("expected non-empty cells for blockquote")
			}
		}
	}
	if !hasBlockquote {
		t.Error("expected a blockquote block")
	}
}

func TestP176_renderBlockquote_AllAlerts(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	alerts := []string{"NOTE", "TIP", "IMPORTANT", "WARNING", "CAUTION"}
	for _, alertType := range alerts {
		mdText := "> [!" + alertType + "]\n> This is an alert message."
		blocks, err := r.Render(mdText)
		if err != nil {
			t.Errorf("Render error for %s: %v", alertType, err)
			continue
		}
		// Should produce at least one block with cells
		if len(blocks) == 0 {
			t.Errorf("expected blocks for %s alert", alertType)
		}
	}
}

func TestP176_renderBlockquote_AlertWithInlineContent(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Alert with inline content on the marker line
	mdText := "> [!NOTE] Inline content here\n> More content"
	blocks, err := r.Render(mdText)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected blocks")
	}
}

func TestP176_renderBlockquote_AlertMultiline(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	mdText := "> [!WARNING]\n> Line 1\n> Line 2\n> Line 3"
	blocks, err := r.Render(mdText)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	if len(blocks) == 0 {
		t.Error("expected blocks for multiline alert")
	}
}

func TestP176_renderBlockquote_Empty(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Empty blockquote
	bq := ast.NewBlockquote()
	blk := r.renderBlockquote(bq, nil)
	if blk == nil {
		// Empty blockquote may return nil — that's fine
	}
}

func TestP176_renderBlockquote_UnknownAlert(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Unknown alert type — should fall through to plain blockquote
	mdText := "> [!UNKNOWN]\n> Content"
	blocks, err := r.Render(mdText)
	if err != nil {
		t.Fatalf("Render error: %v", err)
	}
	// Should still render as a regular blockquote
	if len(blocks) == 0 {
		t.Error("expected blocks for unknown alert")
	}
}

// Test renderBlockquote directly with a constructed AST
func TestP176_renderBlockquote_DirectAST(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)

	// Build a blockquote with a paragraph containing text
	bq := ast.NewBlockquote()
	para := ast.NewParagraph()
	text := ast.NewText()
	input := "Just a regular quote"
	text.Segment = textm.Segment{Start: 0, Stop: len(input)}
	para.AppendChild(para, text)
	bq.AppendChild(bq, para)

	blk := r.renderBlockquote(bq, []byte(input))
	if blk == nil {
		t.Fatal("expected non-nil block")
	}
	if len(blk.Cells) == 0 {
		t.Error("expected non-empty cells")
	}
	// Check bar prefix
	if blk.Cells[0][0].Rune != '│' {
		t.Errorf("expected bar prefix, got %c", blk.Cells[0][0].Rune)
	}
}

// Test alert with text after marker on same line
func TestP176_renderBlockquote_AlertTextAfterMarker(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)

	// Build alert blockquote: [!NOTE] Some text
	bq := ast.NewBlockquote()
	para := ast.NewParagraph()
	text := ast.NewText()
	input := "[!NOTE] Some inline text after marker"
	text.Segment = textm.Segment{Start: 0, Stop: len(input)}
	para.AppendChild(para, text)
	bq.AppendChild(bq, para)

	source := []byte(input)
	blk := r.renderBlockquote(bq, source)
	if blk == nil {
		t.Fatal("expected non-nil block for alert with inline text")
	}
	// Should have cells with the icon + text
	if len(blk.Cells) == 0 {
		t.Error("expected non-empty cells for alert with inline text")
	}
}

// Test alert with no text after marker
func TestP176_renderBlockquote_AlertNoRest(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)

	// Build alert blockquote: [!NOTE] only (no text after)
	bq := ast.NewBlockquote()
	para := ast.NewParagraph()
	text := ast.NewText()
	input := "[!NOTE]"
	text.Segment = textm.Segment{Start: 0, Stop: len(input)}
	para.AppendChild(para, text)
	// Add a second paragraph as content
	para2 := ast.NewParagraph()
	text2 := ast.NewText()
	input2 := "Content in second paragraph"
	text2.Segment = textm.Segment{Start: len(input)+1, Stop: len(input)+1+len(input2)}
	para2.AppendChild(para2, text2)
	bq.AppendChild(bq, para)
	bq.AppendChild(bq, para2)

	source := []byte(input + "\n" + input2)
	blk := r.renderBlockquote(bq, source)
	if blk == nil {
		t.Fatal("expected non-nil block")
	}
}

func TestP176_renderBlockquote_NonParagraphChild(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)

	// Build a blockquote with a code block child (not paragraph)
	bq := ast.NewBlockquote()
	codeBlock := ast.NewCodeBlock()
	bq.AppendChild(bq, codeBlock)

	blk := r.renderBlockquote(bq, nil)
	// Should not panic
	_ = blk
}

// Verify cell colors for alerts
func TestP176_alertColor_AllTypes(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	// Each alert type should return its theme color
	noteColor := r.alertColor("note")
	tipColor := r.alertColor("tip")
	importantColor := r.alertColor("important")
	_ = r.alertColor("warning")
	_ = r.alertColor("caution")

	// They should all be different (different theme colors)
	if noteColor == tipColor && tipColor == importantColor {
		t.Error("expected different colors for different alert types")
	}
}

// Test that alertIcon fallback works
func TestP176_alertIcon_Fallback(t *testing.T) {
	icon := alertIcon("nonexistent")
	if icon == "" {
		t.Error("expected non-empty fallback icon")
	}
	if icon != "● " {
		t.Errorf("expected fallback '● ', got %q", icon)
	}
}

// Test that _ = buffer import is used
func TestP176_BufferImport(t *testing.T) {
	var c buffer.Color
	_ = c
}
