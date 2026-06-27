// Package main demonstrates the Fluui plugin system.
//
// Shows how to:
//   1. Define a custom Block type with serialization
//   2. Register it via a Plugin
//   3. Add it to a BlockContainer
//   4. Serialize/deserialize the container
//
// This example runs without a terminal — it just prints the serialized JSON.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// --- Custom DividerBlock ---

// DividerBlock is a custom block that renders a horizontal divider.
type DividerBlock struct {
	block.BaseBlock
	char rune
}

func NewDividerBlock(id string) *DividerBlock {
	return &DividerBlock{
		BaseBlock: block.NewBaseBlock(id, block.BlockType(99)),
		char:      '\u2550',
	}
}

func (b *DividerBlock) Measure(cs component.Constraints) component.Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	return component.Size{W: w, H: 1}
}

func (b *DividerBlock) Paint(buf *buffer.Buffer) {
	r := b.Bounds()
	for x := r.X; x < r.X+r.W; x++ {
		buf.SetCell(x, r.Y, buffer.Cell{Rune: b.char, Width: 1, Fg: buffer.RGB(0xBD, 0x93, 0xF9)})
	}
}

func (b *DividerBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{"char": b.char})
}

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

// TypeName returns the registry type name for serialization.
func (b *DividerBlock) TypeName() string { return "divider" }

// --- Plugin ---

type dividerPlugin struct{}

func (p *dividerPlugin) Name() string { return "divider-plugin" }
func (p *dividerPlugin) Init(r *block.Registry) error {
	r.Register("divider", func(id string) block.Block {
		return NewDividerBlock(id)
	})
	return nil
}

// --- Custom QuoteBlock ---

type QuoteBlock struct {
	block.BaseBlock
	author string
	quote  string
}

func NewQuoteBlock(id string) *QuoteBlock {
	return &QuoteBlock{
		BaseBlock: block.NewBaseBlock(id, block.BlockType(100)),
	}
}

func (b *QuoteBlock) Measure(cs component.Constraints) component.Size {
	return component.Size{W: cs.MaxWidth, H: 2}
}

func (b *QuoteBlock) Paint(buf *buffer.Buffer) {
	r := b.Bounds()
	buf.DrawText(r.X, r.Y, "\""+b.quote+"\"", buffer.Style{
		Fg: buffer.RGB(0x8B, 0xE9, 0xFD),
	})
	buf.DrawText(r.X, r.Y+1, "  — "+b.author, buffer.Style{
		Fg: buffer.RGB(0x62, 0x72, 0xA4),
	})
}

func (b *QuoteBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{"author": b.author, "quote": b.quote})
}

func (b *QuoteBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Author string `json:"author"`
		Quote  string `json:"quote"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.author = s.Author
	b.quote = s.Quote
	return nil
}

// TypeName returns the registry type name for serialization.
func (b *QuoteBlock) TypeName() string { return "quote" }

type quotePlugin struct{}

func (p *quotePlugin) Name() string { return "quote-plugin" }
func (p *quotePlugin) Init(r *block.Registry) error {
	r.Register("quote", func(id string) block.Block {
		return NewQuoteBlock(id)
	})
	return nil
}

// --- Main ---

func main() {
	// 1. Create registry with default block types.
	registry := block.NewDefaultRegistry()

	// 2. Create plugin manager and load custom plugins.
	pm := block.NewPluginManager(registry)
	if err := pm.LoadAll([]block.Plugin{
		&dividerPlugin{},
		&quotePlugin{},
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load plugins: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded %d plugins:\n", pm.Count())
	for _, p := range pm.Plugins() {
		fmt.Printf("  - %s\n", p.Name())
	}

	// 3. Build a container with custom blocks.
	container := block.NewBlockContainer()

	// Add standard block
	container.AddBlock(block.NewUserMessageBlock("user-1", "Tell me a quote."))

	// Add custom divider
	div, _ := registry.Create("divider", "div-1")
	div.Complete()
	container.AddBlock(div)

	// Add custom quote block
	quote, _ := registry.Create("quote", "quote-1")
	qb := quote.(*QuoteBlock)
	qb.author = "Alan Kay"
	qb.quote = "The best way to predict the future is to invent it."
	qb.Complete()
	container.AddBlock(qb)

	// Add another divider
	div2, _ := registry.Create("divider", "div-2")
	div2.Complete()
	container.AddBlock(div2)

	// Add standard assistant text
	text := block.NewAssistantTextBlock("asst-1")
	text.AppendDelta("That's a great quote!")
	text.Complete()
	container.AddBlock(text)

	fmt.Printf("\nContainer has %d blocks:\n", container.Len())
	for _, b := range container.Blocks() {
		fmt.Printf("  - [%s] %s\n", b.Type(), b.ID())
	}

	// 4. Serialize the container.
	data, err := block.SaveContainer(container, registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Save error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n--- Serialized JSON (%d bytes) ---\n", len(data))
	fmt.Println(string(data))

	// 5. Deserialize into a new container.
	loaded, err := block.LoadContainer(data, registry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Load error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n--- Loaded container has %d blocks ---\n", loaded.Len())

	// Verify custom blocks survived round-trip.
	for _, b := range loaded.Blocks() {
		switch v := b.(type) {
		case *QuoteBlock:
			fmt.Printf("  Quote: %q — %s\n", v.quote, v.author)
		case *DividerBlock:
			fmt.Printf("  Divider (char=%q)\n", v.char)
		case *block.UserMessageBlock:
			fmt.Printf("  UserMessage: %q\n", v.Content())
		case *block.AssistantTextBlock:
			fmt.Printf("  AssistantText: %q\n", v.Content())
		}
	}

	fmt.Println("\nPlugin system demo complete!")
}
