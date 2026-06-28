package component

import (
	"fmt"
	"strings"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ============================================================
// P16-B: Benchmarks for Phase 15 Components
// FilePicker, StatusBar, TabBar, DiffPreview, LinkManager
// Compared against P12 widget baselines (Table, Tree, ProgressBar)
// ============================================================

// --- Helpers ---

func makeMockEntries(n int) []FileEntry {
	entries := make([]FileEntry, n)
	for i := 0; i < n; i++ {
		isDir := i%5 == 0
		name := fmt.Sprintf("file_%04d.go", i)
		if isDir {
			name = fmt.Sprintf("dir_%04d", i)
		}
		entries[i] = FileEntry{
			Name:    name,
			Path:    "/mock/" + name,
			IsDir:   isDir,
			Size:    int64(i * 1024),
			ModTime: int64(i),
		}
	}
	return entries
}

func makeLargeDiff(numFiles, linesPerFile int) string {
	var sb strings.Builder
	for f := 0; f < numFiles; f++ {
		sb.WriteString(fmt.Sprintf("--- a/file%d.go\n", f))
		sb.WriteString(fmt.Sprintf("+++ b/file%d.go\n", f))
		sb.WriteString(fmt.Sprintf("@@ -1,%d +1,%d @@\n", linesPerFile, linesPerFile))
		for l := 0; l < linesPerFile; l++ {
			if l%3 == 0 {
				sb.WriteString(fmt.Sprintf("+added line %d\n", l))
			} else if l%3 == 1 {
				sb.WriteString(fmt.Sprintf("-removed line %d\n", l))
			} else {
				sb.WriteString(fmt.Sprintf(" context line %d\n", l))
			}
		}
	}
	return sb.String()
}

func makeURLText(n int) []string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = fmt.Sprintf("Check https://example.com/page/%d and http://test.org/%d for details", i, i+1)
	}
	return lines
}

func newBenchBuffer(w, h int) *buffer.Buffer {
	return buffer.NewBuffer(w, h)
}

// ============================================================
// FilePicker Benchmarks
// ============================================================

func BenchmarkFilePicker_Measure(b *testing.B) {
	fp := NewFilePicker("/mock")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return makeMockEntries(100), nil
	})
	fp.loadDir("/mock")

	cs := Constraints{MaxWidth: 80, MaxHeight: 24, Has: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp.Measure(cs)
	}
}

func BenchmarkFilePicker_Paint(b *testing.B) {
	fp := NewFilePicker("/mock")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return makeMockEntries(100), nil
	})
	fp.loadDir("/mock")
	fp.SetBounds(Rect{0, 0, 80, 24})

	buf := newBenchBuffer(80, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp.Paint(buf)
	}
}

func BenchmarkFilePicker_Navigation(b *testing.B) {
	fp := NewFilePicker("/mock")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return makeMockEntries(500), nil
	})
	fp.loadDir("/mock")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp.MoveDown()
		if i%500 == 0 {
			fp.SetCursor(0)
		}
	}
}

func BenchmarkFilePicker_Filter(b *testing.B) {
	fp := NewFilePicker("/mock")
	fp.SetDirReader(func(dir string) ([]FileEntry, error) {
		return makeMockEntries(200), nil
	})
	fp.loadDir("/mock")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp.SetFilter("file_00")
		fp.SetFilter("")
	}
}

func BenchmarkFilePicker_LoadDir_100(b *testing.B) {
	entries := makeMockEntries(100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp := NewFilePicker("/mock")
		fp.SetDirReader(func(dir string) ([]FileEntry, error) {
			return entries, nil
		})
		fp.loadDir("/mock")
	}
}

func BenchmarkFilePicker_LoadDir_1000(b *testing.B) {
	entries := makeMockEntries(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fp := NewFilePicker("/mock")
		fp.SetDirReader(func(dir string) ([]FileEntry, error) {
			return entries, nil
		})
		fp.loadDir("/mock")
	}
}

// ============================================================
// StatusBar Benchmarks
// ============================================================

func BenchmarkStatusBar_Measure(b *testing.B) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddCenter("ctx", "8k/128k")
	sb.AddRight("clock", "14:32")

	cs := Constraints{MaxWidth: 120, Has: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb.Measure(cs)
	}
}

func BenchmarkStatusBar_Paint(b *testing.B) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddLeft("tokens", "1.5k tok/s")
	sb.AddCenter("ctx", "8k/128k")
	sb.AddRight("clock", "14:32")
	sb.AddRight("mode", "INSERT")
	sb.SetBounds(Rect{0, 0, 120, 1})

	buf := newBenchBuffer(120, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb.Paint(buf)
	}
}

func BenchmarkStatusBar_UpdateItems(b *testing.B) {
	sb := NewStatusBar()
	sb.AddLeft("model", "GPT-4")
	sb.AddLeft("tokens", "1000")
	sb.AddCenter("ctx", "8k/128k")
	sb.AddRight("clock", "14:32")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sb.SetItemText("tokens", fmt.Sprintf("%d tok/s", 1000+i%1000))
		sb.SetItemText("clock", fmt.Sprintf("%02d:%02d", (i/60)%24, i%60))
	}
}

// ============================================================
// TabBar Benchmarks
// ============================================================

func BenchmarkTabBar_AddTabs(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb := NewTabBar()
		for j := 0; j < 10; j++ {
			tb.AddTab(fmt.Sprintf("tab-%d", j), fmt.Sprintf("Session %d", j))
		}
	}
}

func BenchmarkTabBar_Measure(b *testing.B) {
	tb := NewTabBar()
	for j := 0; j < 10; j++ {
		tb.AddTab(fmt.Sprintf("tab-%d", j), fmt.Sprintf("Session %d", j))
	}

	cs := Constraints{MaxWidth: 120, Has: true}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Measure(cs)
	}
}

func BenchmarkTabBar_Paint(b *testing.B) {
	tb := NewTabBar()
	for j := 0; j < 10; j++ {
		tb.AddTab(fmt.Sprintf("tab-%d", j), fmt.Sprintf("Session %d", j))
	}
	tb.SetActive(3)
	tb.SetBounds(Rect{0, 0, 120, 1})

	buf := newBenchBuffer(120, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.Paint(buf)
	}
}

func BenchmarkTabBar_Navigation(b *testing.B) {
	tb := NewTabBar()
	for j := 0; j < 20; j++ {
		tb.AddTab(fmt.Sprintf("tab-%d", j), fmt.Sprintf("Session %d", j))
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tb.NextTab()
		if i%20 == 0 {
			tb.SetActive(0)
		}
	}
}

// ============================================================
// DiffPreview Benchmarks
// ============================================================

func BenchmarkParseDiff_100(b *testing.B) {
	diff := makeLargeDiff(5, 20)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseDiff(diff)
	}
}

func BenchmarkParseDiff_500(b *testing.B) {
	diff := makeLargeDiff(10, 50)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseDiff(diff)
	}
}

func BenchmarkDiffPreview_SetDiff(b *testing.B) {
	diff := makeLargeDiff(5, 20)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp := NewDiffPreview()
		dp.SetDiff(diff)
	}
}

func BenchmarkDiffPreview_Paint(b *testing.B) {
	dp := NewDiffPreview()
	dp.SetDiff(makeLargeDiff(5, 20))
	dp.SetBounds(Rect{0, 0, 80, 24})

	buf := newBenchBuffer(80, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.Paint(buf)
	}
}

func BenchmarkDiffPreview_Scroll(b *testing.B) {
	dp := NewDiffPreview()
	dp.SetDiff(makeLargeDiff(20, 50))
	dp.SetBounds(Rect{0, 0, 80, 24})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.ScrollDown(1)
		if i%100 == 0 {
			dp.ScrollUp(100)
		}
	}
}

// ============================================================
// LinkManager Benchmarks
// ============================================================

func BenchmarkDetectLinks_10(b *testing.B) {
	text := strings.Repeat("visit https://example.com/test and http://foo.org/bar ", 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkDetectLinks_100(b *testing.B) {
	text := strings.Repeat("visit https://example.com/test and http://foo.org/bar ", 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectLinks(text, 0, 0)
	}
}

func BenchmarkLinkManager_ScanText(b *testing.B) {
	lines := makeURLText(100)
	lm := NewLinkManager()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.ScanText(lines)
	}
}

func BenchmarkLinkManager_AnnotateBuffer(b *testing.B) {
	lm := NewLinkManager()
	lines := makeURLText(10)
	lm.ScanText(lines)

	buf := newBenchBuffer(120, 10)
	for y, line := range lines {
		buf.DrawText(0, y, line, buffer.DefaultStyle)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lm.AnnotateBuffer(buf, 0, 0)
	}
}

// ============================================================
// P12 Baseline Comparison Benchmarks
// ============================================================

func BenchmarkP12Baseline_TablePaint(b *testing.B) {
	t := NewTable([]string{"ID", "Name", "Value"})
	for i := 0; i < 50; i++ {
		t.AddRow([]string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("item-%d", i),
			fmt.Sprintf("val-%d", i),
		})
	}
	t.SetBounds(Rect{0, 0, 80, 24})

	buf := newBenchBuffer(80, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Paint(buf)
	}
}

func BenchmarkP12Baseline_TreePaint(b *testing.B) {
	tr := NewTree()
	root := NewTreeNode("root", "root")
	for i := 0; i < 10; i++ {
		child := NewTreeNode(fmt.Sprintf("child-%d", i), fmt.Sprintf("Child %d", i))
		for j := 0; j < 5; j++ {
			child.AddChild(NewTreeNode(fmt.Sprintf("leaf-%d-%d", i, j), fmt.Sprintf("Leaf %d.%d", i, j)))
		}
		root.AddChild(child)
	}
	tr.SetRoot(root)
	tr.ExpandAll()
	tr.SetBounds(Rect{0, 0, 80, 24})

	buf := newBenchBuffer(80, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Paint(buf)
	}
}

func BenchmarkP12Baseline_ProgressBarPaint(b *testing.B) {
	pb := NewProgressBar()
	pb.SetProgress(65.0)
	pb.SetBounds(Rect{0, 0, 60, 1})

	buf := newBenchBuffer(60, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pb.Paint(buf)
	}
}
