// Package main demonstrates creating a custom Block type.
//
// Shows how to implement the block.Block interface with a custom DividerBlock,
// register it with the registry, and use serialization.
//
// Keys:
//   Esc / Ctrl+C — quit
package main

import (
	"encoding/json"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/app"
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// DividerBlock is a custom block that renders a horizontal divider line.
type DividerBlock struct {
	block.BaseBlock
	char rune
}

// NewDividerBlock creates a custom divider block.
func NewDividerBlock(id string) *DividerBlock {
	b := &DividerBlock{
		BaseBlock: block.NewBaseBlock(id, block.BlockType(99)), // custom type
		char:      '\u2550',                                    // double horizontal line
	}
	return b
}

// Measure returns the block's desired size.
func (b *DividerBlock) Measure(cs component.Constraints) component.Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	return component.Size{W: w, H: 1}
}

// SetBounds stores the allocated rectangle.
func (b *DividerBlock) SetBounds(r component.Rect) {
	b.BaseBlock.SetBounds(r)
}

// Paint renders the divider line.
func (b *DividerBlock) Paint(buf *buffer.Buffer) {
	r := b.Bounds()
	style := buffer.Style{Fg: buffer.RGB(0xBD, 0x93, 0xF9)}
	for x := r.X; x < r.X+r.W; x++ {
		buf.SetCell(x, r.Y, buffer.Cell{Rune: b.char, Width: 1, Fg: style.Fg})
	}
}

// SerializeState implements block.Serializer.
func (b *DividerBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"char": b.char,
	})
}

// DeserializeState implements block.Deserializer.
func (b *DividerBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Char rune `json:"char"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Char != 0 {
		b.char = s.Char
	}
	return nil
}

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	w, h := base.Size()
	chat := app.NewChatApp(w, h)

	// Add standard blocks
	chat.AddUserMessage("Here's a custom divider block:")

	// Add custom block to the container directly
	chat.Container().AddBlock(NewDividerBlock("div-1"))

	// Add more content
	text := chat.AddAssistantText()
	text.AppendDelta("The divider above is a **custom block type**.")
	text.Complete()

	// Serialize and show in terminal
	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()
		chat.SetSize(w, h)
		chat.Render(buf)
	})

	base.OnKey(func(k *term.KeyEvent) {
		if k.Key == term.KeyEscape {
			base.Quit()
		}
		chat.HandleKey(k)
		base.MarkDirty()
	})

	base.Run()
}
