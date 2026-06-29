package markdown

import (
	"testing"
)

// FuzzRendererRender tests the markdown renderer with random input.
// The renderer must never panic on malformed or extreme markdown.
func FuzzRendererRender(f *testing.F) {
	seeds := []string{
		"",                           // empty
		"# Hello",                    // heading
		"**bold** *italic*",          // emphasis
		"- a\n- b\n- c",             // list
		"```go\nfmt.Println()\n```", // code block
		"| a | b |\n|---|---|\n| 1 | 2 |", // table
		"[link](http://example.com)", // link
		"![image](img.png)",          // image
		"> blockquote",               // blockquote
		"---",                        // hr
		"# H1\n## H2\n### H3\n#### H4\n##### H5\n###### H6", // all headings
		"```\n",                      // unclosed code block
		"[unclosed",                  // unclosed link
		"*\n-\n`\n",                  // mixed unclosed markers
		string(make([]byte, 10000)),  // large input (all nulls)
		"🎉🎊✨",                      // emoji
		"    \t  \n  \t  ",           // whitespace only
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 50000 {
			t.Skip()
		}

		r := NewMarkdownRenderer(DefaultTheme(), 80)

		// Render must never panic on any input
		blocks, err := r.Render(input)

		// If no error, blocks should be valid
		if err == nil {
			for _, blk := range blocks {
				// Each block should have a valid type
				if blk.Type < BlockHeading || blk.Type > BlockTable {
					t.Errorf("invalid block type: %d", blk.Type)
				}
			}
		}
	})
}
