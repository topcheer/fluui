package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── P474: Zero-alloc Paint benchmarks for P465-P473 components ───

func BenchmarkPaintResponseInspector(b *testing.B) {
	c := NewResponseInspector()
	c.SetModel("gpt-4o-2024-08-06")
	c.SetLatency(450 * time.Millisecond)
	c.SetTokens(1200, 3500)
	c.SetFinishReason(FinishStop)
	c.SetTemperature(0.7)
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintContextWindowBar(b *testing.B) {
	c := NewContextWindowBar()
	c.SetContextLimit(128000)
	c.SetUsed(95000)
	c.SetBarWidth(40)
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintRateLimitIndicator(b *testing.B) {
	c := NewRateLimitIndicator()
	c.SetLimit(5000)
	c.SetRemaining(3200)
	c.SetResetTime(time.Now().Add(30 * time.Minute))
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintSankeyChart(b *testing.B) {
	c := NewSankeyChart()
	c.AddFlow("Revenue", "Marketing", 500)
	c.AddFlow("Revenue", "Engineering", 800)
	c.AddFlow("Revenue", "Sales", 300)
	c.AddFlow("Marketing", "Ads", 200)
	c.AddFlow("Marketing", "Content", 300)
	c.AddFlow("Engineering", "DevOps", 400)
	c.AddFlow("Engineering", "Backend", 400)
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintScatterPlot(b *testing.B) {
	c := NewScatterPlot()
	c.SetXRange(0, 100)
	c.SetYRange(0, 100)
	for i := 0; i < 50; i++ {
		c.AddPoint(float64(i*2), float64(i)*1.5)
	}
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintMergeView(b *testing.B) {
	c := NewMergeView()
	c.SetLeft("ours", "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8")
	c.SetRight("theirs", "line1\nchanged\nline3\nadded\nline5\ndifferent\nline7\nline8")
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintFunctionCallVisualizer(b *testing.B) {
	c := NewFunctionCallVisualizer()
	c.AddCall("search_web", `{"q":"go tui library"}`, 120*time.Millisecond, CallSuccess)
	c.AddCall("fetch_url", `{"url":"https://example.com"}`, 45*time.Millisecond, CallSuccess)
	c.AddNestedCall("parse_html", `{"raw":"<html>..."}`, 15*time.Millisecond, CallSuccess, 2)
	c.AddNestedCall("extract_text", `{"node":"body"}`, 8*time.Millisecond, CallError, 2)
	c.AddCall("summarize", `{"text":"long text..."}`, 200*time.Millisecond, CallRunning)
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintCodeEditor(b *testing.B) {
	c := NewCodeEditor()
	c.SetLanguage("go")
	c.SetCode(`package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    for i := 0; i < 10; i++ {
        fmt.Printf("Count: %d\n", i)
    }
}
`)
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}

func BenchmarkPaintStreamProgressIndicator(b *testing.B) {
	c := NewStreamProgressIndicator()
	c.SetExpected(500)
	c.Start()
	c.AddTokens(250)
	c.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c.Paint(buf)
	}
}
