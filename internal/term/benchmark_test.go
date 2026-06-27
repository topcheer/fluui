package term

import (
	"bytes"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// byteWriter is a simple io.Writer that discards output.
type byteWriter struct{ data []byte }

func (w *byteWriter) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

// BenchmarkWriterBatchStyleChanges benchmarks batching 100 style changes.
// Each SetStyle call produces SGR escape sequences. The writer batches
// them in an internal buffer before flushing.
func BenchmarkWriterBatchStyleChanges(b *testing.B) {
	styles := []buffer.Style{
		buffer.DefaultStyle.WithFg(buffer.RGB(255, 0, 0)),
		buffer.DefaultStyle.WithFg(buffer.RGB(0, 255, 0)).AddFlags(buffer.Bold),
		buffer.DefaultStyle.WithFg(buffer.RGB(0, 0, 255)).AddFlags(buffer.Italic),
		buffer.DefaultStyle.WithFg(buffer.RGB(255, 255, 0)).AddFlags(buffer.Underline),
		buffer.DefaultStyle.WithFg(buffer.RGB(255, 0, 255)).AddFlags(buffer.Reverse),
		buffer.DefaultStyle.WithFg(buffer.RGB(0, 255, 255)),
		buffer.DefaultStyle.WithFg(buffer.RGB(128, 128, 128)).AddFlags(buffer.Dim),
		buffer.DefaultStyle.WithFg(buffer.RGB(64, 64, 64)).AddFlags(buffer.Strikethrough),
	}

	bw := &byteWriter{}
	tw := NewWriter(bw, ProfileTrue)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tw.buf.Reset()
		for j := 0; j < 100; j++ {
			s := styles[j%len(styles)]
			tw.SetStyle(s)
			tw.MoveTo(j%80, j%24)
			tw.WriteString("X")
		}
		tw.ResetStyle()
		_ = tw.Flush()
	}
}

// BenchmarkWriterFlush benchmarks flushing a ~1KB buffer to the underlying writer.
func BenchmarkWriterFlush(b *testing.B) {
	// Pre-build a 1KB payload.
	var payload bytes.Buffer
	for payload.Len() < 1024 {
		payload.WriteString("\x1b[38;2;255;128;0mHello, World! \x1b[0m")
	}
	data := payload.Bytes()[:1024]

	bw := &byteWriter{}
	tw := NewWriter(bw, ProfileTrue)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tw.buf.Reset()
		tw.buf.Write(data)
		_ = tw.Flush()
	}
}
