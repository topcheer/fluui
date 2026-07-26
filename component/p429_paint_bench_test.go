package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// Paint benchmarks for components that previously lacked zero-alloc verification.

func BenchmarkAIProgress_Paint_P429(b *testing.B) {
	c := NewAIProgress()
	c.SetPhase(AIPhaseGenerating)
	c.SetProgress(0.5)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkConfidenceMeter_Paint_P429(b *testing.B) {
	c := NewConfidenceMeter(0.75)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkStreamingText_Paint_P429(b *testing.B) {
	c := NewStreamingText()
	c.SetText("The quick brown fox jumps over the lazy dog repeatedly")
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkSparkline_Paint_P429(b *testing.B) {
	c := NewSparkline()
	c.SetData([]float64{1, 3, 2, 5, 4, 6, 3, 7, 5, 8, 4, 9, 6, 3, 7, 5, 4, 8, 6, 3})
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkBarChart_Paint_P429(b *testing.B) {
	c := NewBarChart()
	c.AddSeries(BarSeries{Name: "A", Data: []BarData{{Label: "M", Value: 10}, {Label: "T", Value: 20}, {Label: "W", Value: 30}}})
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkLineChart_Paint_P429(b *testing.B) {
	c := NewLineChart()
	c.AddSeries(ChartSeries{
		Name: "CPU",
		Data: []ChartPoint{{X: 0, Y: 10}, {X: 1, Y: 30}, {X: 2, Y: 50}, {X: 3, Y: 70}, {X: 4, Y: 50}},
	})
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkAutoComplete_Paint_P429(b *testing.B) {
	c := NewAutoComplete()
	c.SetItems([]CompletionItem{
		{Label: "apple"}, {Label: "banana"}, {Label: "cherry"},
		{Label: "date"}, {Label: "elderberry"}, {Label: "fig"}, {Label: "grape"},
	})
	c.SetQuery("a")
	c.Show(0, 1)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkCommandPalette_Paint_P429(b *testing.B) {
	c := NewCommandPalette()
	for _, label := range []string{"Open File", "Save", "Save As", "Close", "Quit", "Find", "Replace"} {
		c.AddCommand(Command{ID: label, Label: label})
	}
	c.SetQuery("sa")
	c.Show(0, 0)
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkCodeBlock_Paint_P429(b *testing.B) {
	c := NewCodeBlock("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n", "go")
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 6})
	buf := buffer.NewBuffer(40, 6)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkMarkdownStream_Paint_P429(b *testing.B) {
	c := NewMarkdownStream()
	c.SetSource("# Hello\n\nThis is **bold** and *italic*.\n\n- Item 1\n- Item 2\n")
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkTable_Paint_P429(b *testing.B) {
	c := NewTable([]string{"Name", "Age", "City"},
		[]string{"Alice", "30", "NYC"},
		[]string{"Bob", "25", "LA"},
		[]string{"Charlie", "35", "SF"},
		[]string{"Diana", "28", "Boston"},
		[]string{"Eve", "32", "Seattle"})
	c.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkTree_Paint_P429(b *testing.B) {
	root := NewTreeNode("root", "Project")
	src := NewTreeNode("src", "src")
	src.AddChild(NewTreeNode("main", "main.go"))
	src.AddChild(NewTreeNode("util", "util.go"))
	root.AddChild(src)
	root.AddChild(NewTreeNode("test", "test"))

	c := NewTree()
	c.SetRoot(root)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkRichLog_Paint_P429(b *testing.B) {
	c := NewRichLog()
	c.SetMaxSize(50)
	for i := 0; i < 20; i++ {
		c.Info("log entry that is somewhat long for testing purposes")
	}
	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})
	buf := buffer.NewBuffer(60, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkCalendar_Paint_P429(b *testing.B) {
	c := NewCalendar()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 10})
	buf := buffer.NewBuffer(30, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkViewport_Paint_P429(b *testing.B) {
	child := NewText("line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10")
	c := NewViewport(child)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkSplitPane_Paint_P429(b *testing.B) {
	c := NewSplitPane(NewText("Left"), NewText("Right"))
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkQRCode_Paint_P429(b *testing.B) {
	c := NewQRCode("https://github.com/topcheer/fluui")
	c.SetBounds(Rect{X: 0, Y: 0, W: 25, H: 25})
	buf := buffer.NewBuffer(25, 25)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkTooltip_Paint_P429(b *testing.B) {
	c := NewTooltip("This is a helpful tooltip with some guidance text.")
	c.Show()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 3})
	buf := buffer.NewBuffer(40, 3)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkSelectionList_Paint_P429(b *testing.B) {
	c := NewSelectionList([]string{"Apple", "Banana", "Cherry", "Date", "Elderberry", "Fig", "Grape"})
	c.Toggle(0)
	c.Toggle(2)
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 8})
	buf := buffer.NewBuffer(30, 8)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkFilteredList_Paint_P429(b *testing.B) {
	c := NewFilteredList([]string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honeydew"})
	c.SetQuery("a")
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 5})
	buf := buffer.NewBuffer(30, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}
