// Package main demonstrates markdown rendering with Fluui.
//
// Renders headings, bold/italic, code blocks, lists, tables, links,
// and blockquotes with the Dracula color theme and syntax highlighting.
//
// Keys:
//   Up/Down       — scroll
//   Esc / Ctrl+C  — quit
package main

import (
	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

const sampleMarkdown = `# Markdown Rendering Demo

Fluui renders **bold**, *italic*, ~~strikethrough~~, and ` + "`inline code`" + `.

## Features

- Headings with distinct colors (H1-H6)
- Ordered and unordered lists
- Syntax-highlighted code blocks
- Tables with alignment
- Blockquotes
- Links (clickable in supporting terminals)

### Code Block Example

` + "```go" + `
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
` + "```" + `

### Table Example

| Language | Type System | Year |
|:---------|:-----------:|----:|
| Go       | Static      | 2009 |
| Rust     | Static      | 2010 |
| Python   | Dynamic     | 1991 |

### Blockquote

> The best way to predict the future is to invent it.
> — Alan Kay

### Inline Formatting

You can use ` + "`code`" + `, **bold**, and *italic* in the same line.
Links like [Fluui](https://github.com/topcheer/fluui) are also supported.
`

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	w, h := base.Size()
	chat := app.NewChatApp(w, h)

	// Render markdown into an AssistantTextBlock
	md := chat.AddAssistantText()
	md.AppendDelta(sampleMarkdown)
	md.Complete()

	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()
		chat.SetSize(w, h)
		chat.Render(buf)
	})

	base.OnKey(func(k *term.KeyEvent) {
		if k.Key == term.KeyEscape {
			base.Quit()
			return
		}
		chat.HandleKey(k)
		base.MarkDirty()
	})

	base.OnMouse(func(m *term.MouseEvent) {
		chat.HandleMouse(m)
		base.MarkDirty()
	})

	base.Run()
}
