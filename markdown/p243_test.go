package markdown

import (
	"strings"
	"testing"
)

func TestLatexConsumeCommand_SingleChar_P243(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{`\,`, " "}, {`\;`, " "}, {`\:`, " "},
		{`\%`, "%"}, {`\#`, "#"}, {`\&`, "&"},
		{`\_`, "_"}, {`\{`, "{"}, {`\}`, "}"},
		{`\ `, " "}, {`\x`, "x"},
	}
	for _, tc := range cases {
		p := &latexParser{input: tc.input}
		var sb strings.Builder
		p.consumeCommand(&sb)
		if sb.String() != tc.want {
			t.Errorf("consumeCommand(%q) = %q, want %q", tc.input, sb.String(), tc.want)
		}
	}
}

func TestLatexConsumeCommand_LineBreak_P243(t *testing.T) {
	p := &latexParser{input: `\\hello`}
	var sb strings.Builder
	p.consumeCommand(&sb)
	if sb.String() != "  " {
		t.Errorf("got %q", sb.String())
	}
}

func TestLatexSkipBracket_NoBracket_P243(t *testing.T) {
	p := &latexParser{input: "hello"}
	p.skipBracket()
	if p.pos != 0 {
		t.Error("should not advance")
	}
}

func TestLatexSkipBracket_WithContent_P243(t *testing.T) {
	p := &latexParser{input: "[optional]rest"}
	p.skipBracket()
	if p.pos < 10 {
		t.Errorf("pos=%d", p.pos)
	}
}

func TestLatexSkipBracket_Nested_P243(t *testing.T) {
	p := &latexParser{input: "[a[b]c]rest"}
	p.skipBracket()
	if p.pos < 7 { // [a[b]c] = 7 chars consumed
		t.Errorf("pos=%d", p.pos)
	}
}

func TestRenderList_TaskList_P243(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("- [ ] unchecked\n- [x] checked\n- normal")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Error("no blocks")
	}
}

func TestRenderList_OrderedList_P243(t *testing.T) {
	r := NewMarkdownRenderer(DefaultTheme(), 80)
	blocks, err := r.Render("1. first\n2. second\n3. third")
	if err != nil {
		t.Fatal(err)
	}
	if len(blocks) == 0 {
		t.Error("no blocks")
	}
}
