package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P221: codeblock paintStreamingCursorLocked + diffpreview paintBorderLocked

func TestCodeBlock_StreamingCursorEmptyLines_P221(t *testing.T) {
	cb := NewCodeBlock("", "go", "")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(40, 10)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursorWithLines_P221(t *testing.T) {
	cb := NewCodeBlock("test", "go", "func main() {\n\tprintln(\"hello\")\n}")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(60, 10)
	cb.Paint(buf)
}

func TestCodeBlock_StreamingCursorShowTitle_P221(t *testing.T) {
	cb := NewCodeBlock("title", "go", "code")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	cb.SetStreaming(true)
	buf := buffer.NewBuffer(40, 5)
	cb.Paint(buf)
}

func TestDiffPreview_PaintWithStats_P221(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetTitle("test.go")
	dp.SetLines([]DiffLine{
		{Type: DiffAdd, Content: "added line"},
		{Type: DiffDel, Content: "removed line"},
		{Type: DiffAdd, Content: "another add"},
	})
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

func TestDiffPreview_PaintNoTitle_P221(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	dp.Paint(buf)
}

func TestDiffPreview_PaintWithTitleNoStats_P221(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetTitle("file.rs")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 6})
	buf := buffer.NewBuffer(30, 6)
	dp.Paint(buf)
}