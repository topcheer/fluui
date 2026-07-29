package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownInlineCodeBasic(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("Use `fmt.Println` to print.")
	if mic.InlineCodeCount() != 1 {
		t.Errorf("InlineCodeCount = %d, want 1", mic.InlineCodeCount())
	}
}

func TestMarkdownInlineCodeFencedBlock(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("```go\nfunc main() {}\n```")
	if mic.CodeBlockCount() != 2 {
		t.Errorf("CodeBlockCount = %d, want 2 (label+content)", mic.CodeBlockCount())
	}
}

func TestMarkdownInlineCodeMixed(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("Use `vim` to edit.\n```bash\necho hello\n```\nDone `now`.")
	if mic.InlineCodeCount() != 2 {
		t.Errorf("InlineCodeCount = %d, want 2", mic.InlineCodeCount())
	}
}

func TestMarkdownInlineCodeEmpty(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("")
	if mic.InlineCodeCount() != 0 {
		t.Errorf("InlineCodeCount = %d, want 0", mic.InlineCodeCount())
	}
}

func TestMarkdownInlineCodeCounts(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("Inline `a` and `b`.\n```py\nprint(1)\n```")
	if mic.InlineCodeCount() != 2 {
		t.Errorf("InlineCodeCount = %d, want 2", mic.InlineCodeCount())
	}
	if mic.CodeBlockCount() != 2 {
		t.Errorf("CodeBlockCount = %d, want 2 (label+content)", mic.CodeBlockCount())
	}
}

func TestMarkdownInlineCodeMeasure(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("Text `code` text")
	s := mic.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 3 {
		t.Errorf("H = %d, want >= 3", s.H)
	}
}

func TestMarkdownInlineCodePaint(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetMarkdown("Run `make test` now.\n```go\nfmt.Println()\n```")
	mic.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})

	buf := buffer.NewBuffer(50, 8)
	mic.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	// Check text content
	foundText := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == 'R' {
			foundText = true
			break
		}
	}
	if !foundText {
		t.Error("text 'Run' not found")
	}
}

func TestMarkdownInlineCodePaintEmpty(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 3})
	buf := buffer.NewBuffer(50, 3)
	mic.Paint(buf) // should not panic
}

func TestMarkdownInlineCodeChildren(t *testing.T) {
	mic := NewMarkdownInlineCode()
	if mic.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownInlineCodeStyle(t *testing.T) {
	mic := NewMarkdownInlineCode()
	mic.SetStyle(InlineCodeStyle{
		Text:       buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		InlineCode: buffer.Style{Fg: buffer.RGB(255, 165, 0), Bg: buffer.RGB(30, 30, 30)},
		CodeBlock:  buffer.Style{Fg: buffer.RGB(0, 255, 0), Bg: buffer.RGB(15, 15, 15)},
		BlockLabel: buffer.Style{Fg: buffer.RGB(0, 0, 255), Flags: buffer.Bold},
		Border:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	mic.SetMarkdown("`code`\n```go\nx\n```")
	mic.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	mic.Paint(buf)
}
