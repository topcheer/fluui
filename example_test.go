package fluui_test

import (
	"fmt"
	"io"

	"github.com/topcheer/fluui"
	"github.com/topcheer/fluui/internal/buffer"
)

// ExampleNewWithWriter demonstrates creating a headless App for CI/testing.
func ExampleNewWithWriter() {
	app := fluui.NewWithWriter(io.Discard, 80, 24)
	app.DrawText(0, 0, "Hello!", buffer.DefaultStyle)
	w, h := app.Size()
	fmt.Printf("%dx%d\n", w, h)
	// Output: 80x24
}

// ExampleApp_DrawText demonstrates drawing text at coordinates.
func ExampleApp_DrawText() {
	app := fluui.NewWithWriter(io.Discard, 80, 24)
	app.DrawText(10, 5, "Hello", buffer.DefaultStyle)
	cell := app.Renderer().Back().GetCell(10, 5)
	fmt.Printf("Cell rune: %c\n", cell.Rune)
	// Output: Cell rune: H
}

// ExampleApp_SetTitle demonstrates setting the terminal window title.
func ExampleApp_SetTitle() {
	app := fluui.NewWithWriter(io.Discard, 80, 24)
	app.SetTitle("My App")
	fmt.Printf("Title: %s\n", app.Title())
	// Output: Title: My App
}

// ExampleApp_Size demonstrates getting terminal dimensions.
func ExampleApp_Size() {
	app := fluui.NewWithWriter(io.Discard, 120, 40)
	w, h := app.Size()
	fmt.Printf("%dx%d\n", w, h)
	// Output: 120x40
}
