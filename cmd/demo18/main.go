// Demo18: ImageBlock showcase — inline image display in AI chat.
//
// This demo shows how to use ImageBlock to display images inline in a Fluui
// AI chat interface. ImageBlock detects the terminal's image protocol (iTerm2,
// Kitty Graphics, or Sixel) and generates the appropriate escape sequence.
// On terminals without image support, a metadata placeholder is shown.
package main

import (
	"fmt"
	"os"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/termcompat"
	"github.com/topcheer/fluui/theme"
)

func main() {
	// Detect available image protocol
	caps := termcompat.DetectImageProtocol()
	fmt.Printf("Image Protocol: %s (CanDisplay: %v)\n\n",
		termcompat.ImageProtocolName(caps.Protocol),
		caps.CanDisplay)

	// Create sample image data
	// A small "PNG" file (just header bytes for demo)
	pngHeader := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	pngData := append(pngHeader, make([]byte, 10240)...)

	// Create image blocks with different metadata
	images := []*block.ImageBlock{
		block.NewImageBlockWithDims("img1", "screenshot.png", pngData, 1920, 1080),
		block.NewImageBlockWithDims("img2", "chart.jpeg", make([]byte, 456789), 800, 600),
		block.NewImageBlock("img3", "avatar.gif", make([]byte, 2048)),
		block.NewRGBAImageBlock("img4", "heatmap.rgba", make([]byte, 4*100*100), 100, 100),
	}

	// Configure display sizes
	images[0].SetDisplaySize(60, 30)
	images[1].SetDisplaySize(40, 30)
	images[2].SetDisplaySize(20, 10)

	// Render each block to a buffer and print
	for i, img := range images {
		fmt.Printf("─── Image %d ───────────────────────────────────\n", i+1)

		// Show metadata
		fmt.Printf("Filename: %s\n", img.Filename())
		fmt.Printf("Format:   %s\n", img.Format())
		fmt.Printf("Size:     %s\n", img.FileSize())
		if img.Width() > 0 {
			fmt.Printf("Pixels:   %dx%d\n", img.Width(), img.Height())
		}
		fmt.Printf("Protocol: %s\n", img.ProtocolName())

		// Show the escape sequence status
		seq := img.Sequence()
		if seq != "" {
			fmt.Printf("Sequence: %d bytes (will display inline on supported terminals)\n", len(seq))
		} else {
			fmt.Printf("Sequence: (none — no image protocol support)\n")
		}
		fmt.Println()

		// Render the ASCII placeholder
		_ = img.Measure(component.Bounded(60, 20))
		img.SetBounds(component.Rect{X: 0, Y: 0, W: 50, H: 6})
		buf := buffer.NewBuffer(50, 6)

		t := theme.Get()
		_ = t
		img.Paint(buf)

		// Print the buffer as text
		for y := 0; y < 6; y++ {
			line := ""
			for x := 0; x < 50; x++ {
				cell := buf.GetCell(x, y)
				if cell.Rune != 0 {
					line += string(cell.Rune)
				} else {
					line += " "
				}
			}
			fmt.Println(line)
		}
		fmt.Println()
	}

	// Demonstrate serialize/deserialize round-trip
	fmt.Println("─── Serialization Round-Trip ──────────────────")
	original := images[0]
	data, _ := original.SerializeState()
	fmt.Printf("Serialized: %d bytes\n", len(data))

	restored := block.NewImageBlock("restored", "", nil)
	_ = restored.DeserializeState(data)
	fmt.Printf("Restored filename: %s\n", restored.Filename())
	fmt.Printf("Restored format:   %s\n", restored.Format())
	fmt.Printf("Restored dims:     %dx%d\n", restored.Width(), restored.Height())

	fmt.Println("\n✓ ImageBlock demo complete")

	os.Exit(0)
}
