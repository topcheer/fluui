package block

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestDetectDiff(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "git diff header",
			input: "diff --git a/foo.go b/foo.go\n@@ -1,3 +1,4 @@\n context\n+added\n-removed\n",
			want:  true,
		},
		{
			name:  "hunk with add/del",
			input: "@@ -1,3 +1,4 @@\n old line\n+new line\n-old line\n",
			want:  true,
		},
		{
			name:  "plain text",
			input: "Hello world\nThis is just text\n",
			want:  false,
		},
		{
			name:  "empty",
			input: "",
			want:  false,
		},
		{
			name:  "only additions",
			input: "+just add\n+more add\n",
			want:  false, // no hunk header
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectDiff(tc.input)
			if got != tc.want {
				t.Errorf("DetectDiff() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParseDiffContext(t *testing.T) {
	lines := ParseDiff("just a normal line")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Type != DiffContext {
		t.Errorf("expected DiffContext, got %v", lines[0].Type)
	}
	if lines[0].Content != "just a normal line" {
		t.Errorf("content: got %q", lines[0].Content)
	}
}

func TestParseDiffAdd(t *testing.T) {
	lines := ParseDiff("+added code")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Type != DiffAdd {
		t.Errorf("expected DiffAdd, got %v", lines[0].Type)
	}
	if lines[0].Content != "+added code" {
		t.Errorf("content: got %q", lines[0].Content)
	}
}

func TestParseDiffDel(t *testing.T) {
	lines := ParseDiff("-removed code")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Type != DiffDel {
		t.Errorf("expected DiffDel, got %v", lines[0].Type)
	}
	if lines[0].Content != "-removed code" {
		t.Errorf("content: got %q", lines[0].Content)
	}
}

func TestParseDiffHunk(t *testing.T) {
	lines := ParseDiff("@@ -1,3 +1,4 @@")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Type != DiffHunk {
		t.Errorf("expected DiffHunk, got %v", lines[0].Type)
	}
	if lines[0].Content != "@@ -1,3 +1,4 @@" {
		t.Errorf("content: got %q", lines[0].Content)
	}
}

func TestParseDiffFile(t *testing.T) {
	lines := ParseDiff("diff --git a/foo.go b/foo.go")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0].Type != DiffFile {
		t.Errorf("expected DiffFile, got %v", lines[0].Type)
	}
	if lines[0].Content != "diff --git a/foo.go b/foo.go" {
		t.Errorf("content: got %q", lines[0].Content)
	}
}

func TestParseDiffMeta(t *testing.T) {
	tests := []struct {
		line string
	}{
		{"index abc..def 100644"},
		{"--- a/foo.go"},
		{"+++ b/foo.go"},
	}
	for _, tc := range tests {
		lines := ParseDiff(tc.line)
		if len(lines) != 1 {
			t.Fatalf("expected 1 line for %q, got %d", tc.line, len(lines))
		}
		if lines[0].Type != DiffMeta {
			t.Errorf("for %q: expected DiffMeta, got %v", tc.line, lines[0].Type)
		}
	}
}

func TestParseDiffMixed(t *testing.T) {
	// Build diff with strings.Join to have precise control over lines
	input := strings.Join([]string{
		"diff --git a/main.go b/main.go", // DiffFile
		"index abc..def 100644",          // DiffMeta
		"--- a/main.go",                   // DiffMeta
		"+++ b/main.go",                   // DiffMeta
		"@@ -1,5 +1,6 @@",                 // DiffHunk
		" package main",                   // DiffContext
		" func main() {",                  // DiffContext
		"-old line",                       // DiffDel
		"+new line",                       // DiffAdd
		"+another new",                    // DiffAdd
		" }",                              // DiffContext
		"",                                // trailing empty = DiffContext
	}, "\n")

	lines := ParseDiff(input)

	expected := []DiffType{
		DiffFile,
		DiffMeta,
		DiffMeta,
		DiffMeta,
		DiffHunk,
		DiffContext,
		DiffContext,
		DiffDel,
		DiffAdd,
		DiffAdd,
		DiffContext,
		DiffContext,
	}

	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d", len(expected), len(lines))
	}

	for i, dt := range expected {
		if lines[i].Type != dt {
			t.Errorf("line %d: expected %v, got %v (content: %q)",
				i, dt, lines[i].Type, lines[i].Content)
		}
	}
}

func TestDiffStyleAdd(t *testing.T) {
	style := DiffStyle(DiffAdd)
	expected := buffer.RGB(0x50, 0xFA, 0x7B) // green
	if !style.Fg.Equal(expected) {
		t.Errorf("Add Fg: got %v, want %v", style.Fg, expected)
	}
	// Add should not have Bold
	if style.Flags&buffer.Bold != 0 {
		t.Error("Add should not have Bold flag")
	}
}

func TestDiffStyleDel(t *testing.T) {
	style := DiffStyle(DiffDel)
	expected := buffer.RGB(0xFF, 0x55, 0x55) // red
	if !style.Fg.Equal(expected) {
		t.Errorf("Del Fg: got %v, want %v", style.Fg, expected)
	}
}

func TestDiffStyleHunk(t *testing.T) {
	style := DiffStyle(DiffHunk)
	expected := buffer.RGB(0x8B, 0xE9, 0xFD) // cyan
	if !style.Fg.Equal(expected) {
		t.Errorf("Hunk Fg: got %v, want %v", style.Fg, expected)
	}
}

func TestDiffStyleFile(t *testing.T) {
	style := DiffStyle(DiffFile)
	expected := buffer.RGB(0xBD, 0x93, 0xF9) // purple
	if !style.Fg.Equal(expected) {
		t.Errorf("File Fg: got %v, want %v", style.Fg, expected)
	}
	if style.Flags&buffer.Bold == 0 {
		t.Error("File should have Bold flag")
	}
}

func TestDiffStyleContext(t *testing.T) {
	style := DiffStyle(DiffContext)
	// Context should have default (no special) style
	if !style.Equal(buffer.DefaultStyle) {
		t.Errorf("Context style: got %v, want DefaultStyle", style)
	}
}

func TestDiffStyleMeta(t *testing.T) {
	style := DiffStyle(DiffMeta)
	expected := buffer.RGB(0x62, 0x72, 0xA4) // dim gray-blue
	if !style.Fg.Equal(expected) {
		t.Errorf("Meta Fg: got %v, want %v", style.Fg, expected)
	}
	if style.Flags&buffer.Dim == 0 {
		t.Error("Meta should have Dim flag")
	}
}

func TestParseDiffEmptyString(t *testing.T) {
	lines := ParseDiff("")
	// strings.Split("", "\n") returns [""], so 1 line
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for empty string, got %d", len(lines))
	}
	if lines[0].Type != DiffContext {
		t.Errorf("expected DiffContext, got %v", lines[0].Type)
	}
}
