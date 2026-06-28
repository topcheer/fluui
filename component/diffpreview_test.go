package component

import (
	"strings"
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// --- Construction ---

func TestNewDiffPreview(t *testing.T) {
	dp := NewDiffPreview()
	if dp == nil {
		t.Fatal("NewDiffPreview returned nil")
	}
	if !dp.IsEmpty() {
		t.Error("new DiffPreview should be empty")
	}
	if dp.LineCount() != 0 {
		t.Errorf("expected 0 lines, got %d", dp.LineCount())
	}
}

func TestNewDiffPreview_HasID(t *testing.T) {
	dp := NewDiffPreview()
	if dp.ID() == "" {
		t.Error("DiffPreview should have a non-empty ID")
	}
}

func TestNewDiffPreview_UniqueIDs(t *testing.T) {
	dp1 := NewDiffPreview()
	dp2 := NewDiffPreview()
	if dp1.ID() == dp2.ID() {
		t.Error("two DiffPreviews should have unique IDs")
	}
}

func TestDiffPreview_ImplementsComponent(t *testing.T) {
	dp := NewDiffPreview()
	var _ Component = dp
}

// --- SetDiff ---

const sampleDiff = `diff --git a/f.go b/f.go
index abc..def
--- a/f.go
+++ b/f.go
@@ -1,3 +1,4 @@
 ctx
-old
+new
+added`

func TestDiffPreview_SetDiff(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(sampleDiff)

	if dp.IsEmpty() {
		t.Fatal("expected non-empty after SetDiff")
	}
	if dp.LineCount() != 9 {
		t.Errorf("expected 9 lines, got %d", dp.LineCount())
	}

	stats := dp.Stats()
	if stats.Additions != 2 {
		t.Errorf("Additions = %d, want 2", stats.Additions)
	}
	if stats.Deletions != 1 {
		t.Errorf("Deletions = %d, want 1", stats.Deletions)
	}
	if stats.Files != 1 {
		t.Errorf("Files = %d, want 1", stats.Files)
	}
	if stats.Hunks != 1 {
		t.Errorf("Hunks = %d, want 1", stats.Hunks)
	}
}

func TestDiffPreview_SetDiff_Empty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("")
	if !dp.IsEmpty() {
		t.Error("expected empty after SetDiff(\"\")")
	}
}

func TestDiffPreview_SetDiff_MultipleFiles(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(`diff --git a/file1.go b/file1.go
@@ -1,1 +1,1 @@
-a
+b
diff --git a/file2.go b/file2.go
@@ -1,1 +1,1 @@
-c
+d`)
	stats := dp.Stats()
	if stats.Files != 2 {
		t.Errorf("expected 2 files, got %d", stats.Files)
	}
}

// --- Lines ---

func TestDiffPreview_Lines(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+add\n-del")
	lines := dp.Lines()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0].Type != DiffAdd {
		t.Errorf("line[0].Type = %v, want DiffAdd", lines[0].Type)
	}
	if lines[1].Type != DiffDel {
		t.Errorf("line[1].Type = %v, want DiffDel", lines[1].Type)
	}
}

func TestDiffPreview_Lines_ReturnsCopy(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added")
	lines := dp.Lines()
	lines[0].Content = "modified"
	again := dp.Lines()
	if again[0].Content == "modified" {
		t.Error("Lines() should return a copy")
	}
}

// --- IsEmpty / HasChanges ---

func TestDiffPreview_IsEmpty(t *testing.T) {
	dp := NewDiffPreview()
	if !dp.IsEmpty() {
		t.Error("should be empty initially")
	}
	dp.SetDiff("+test")
	if dp.IsEmpty() {
		t.Error("should not be empty after SetDiff")
	}
}

func TestDiffPreview_HasChanges(t *testing.T) {
	dp := NewDiffPreview()
	if dp.HasChanges() {
		t.Error("should have no changes initially")
	}
	dp.SetDiff("+added")
	if !dp.HasChanges() {
		t.Error("should have changes after adding")
	}
}

func TestDiffPreview_DiffSummary(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+add\n-del")
	summary := dp.DiffSummary()
	if !strings.Contains(summary, "+1") || !strings.Contains(summary, "-1") {
		t.Errorf("summary should contain +1 and -1, got %q", summary)
	}
}

// --- Style ---

func TestDiffPreview_SetStyle(t *testing.T) {
	dp := NewDiffPreview()
	custom := DiffPreviewStyle{
		AddLine: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
	}
	dp.SetStyle(custom)
	s := dp.Style()
	if s.AddLine.Fg.R() != 0 || s.AddLine.Fg.G() != 255 {
		t.Error("SetStyle did not work")
	}
}

func TestDefaultDiffPreviewStyle(t *testing.T) {
	s := DefaultDiffPreviewStyle()
	if s.AddLine.Fg == s.DelLine.Fg {
		t.Error("AddLine and DelLine should differ")
	}
}

// --- Title ---

func TestDiffPreview_SetTitle(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetTitle("My Changes")
	if dp.Title() != "My Changes" {
		t.Errorf("Title = %q, want %q", dp.Title(), "My Changes")
	}
}

// --- Scrolling ---

func TestDiffPreview_ScrollY_Default(t *testing.T) {
	dp := NewDiffPreview()
	if dp.ScrollY() != 0 {
		t.Errorf("ScrollY should default to 0, got %d", dp.ScrollY())
	}
}

func genDiff(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteString("+line\n")
	}
	return sb.String()
}

func TestDiffPreview_ScrollDown(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(genDiff(30))
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollDown(1)
	if dp.ScrollY() != 1 {
		t.Errorf("ScrollY = %d, want 1", dp.ScrollY())
	}
}

func TestDiffPreview_ScrollUp(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(genDiff(30))
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollDown(5)
	dp.ScrollUp(2)
	if dp.ScrollY() != 3 {
		t.Errorf("ScrollY = %d, want 3", dp.ScrollY())
	}
}

func TestDiffPreview_ScrollUp_ClampedAtZero(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollUp(10)
	if dp.ScrollY() != 0 {
		t.Errorf("ScrollY should be clamped at 0, got %d", dp.ScrollY())
	}
}

func TestDiffPreview_ScrollTo(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(genDiff(30))
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollTo(2)
	if dp.ScrollY() != 2 {
		t.Errorf("ScrollY = %d, want 2", dp.ScrollY())
	}
}

func TestDiffPreview_ScrollTo_Negative(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollTo(-5)
	if dp.ScrollY() != 0 {
		t.Errorf("ScrollY should be 0 for negative, got %d", dp.ScrollY())
	}
}

func TestDiffPreview_ScrollPageDown(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(genDiff(50))
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollPageDown(3)
	if dp.ScrollY() != 3 {
		t.Errorf("ScrollY after PageDown(3) = %d, want 3", dp.ScrollY())
	}
}

func TestDiffPreview_ScrollPageUp(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(genDiff(50))
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	dp.ScrollTo(10)
	dp.ScrollPageUp(4)
	if dp.ScrollY() != 6 {
		t.Errorf("ScrollY after PageUp(4) = %d, want 6", dp.ScrollY())
	}
}

func TestDiffPreview_VisibleRange(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(sampleDiff)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 6})
	start, end := dp.VisibleRange()
	if start != 0 {
		t.Errorf("start = %d, want 0", start)
	}
	if end <= start {
		t.Errorf("end (%d) should be > start (%d)", end, start)
	}
}

// --- Measure ---

func TestDiffPreview_Measure_Empty(t *testing.T) {
	dp := NewDiffPreview()
	s := dp.Measure(Constraints{})
	if s.H < 3 {
		t.Errorf("H should be >= 3, got %d", s.H)
	}
}

func TestDiffPreview_Measure_WithDiff(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added line\n-removed line\n context")
	s := dp.Measure(Constraints{})
	if s.H < 5 {
		t.Errorf("H should be >= 5, got %d", s.H)
	}
}

func TestDiffPreview_Measure_ClampedToMax(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+a\n+b\nc\nd\ne\nf")
	s := dp.Measure(Constraints{MaxWidth: 20, MaxHeight: 4})
	if s.W > 20 {
		t.Errorf("W should be <= 20, got %d", s.W)
	}
	if s.H > 4 {
		t.Errorf("H should be <= 4, got %d", s.H)
	}
}

// --- SetBounds ---

func TestDiffPreview_SetBounds(t *testing.T) {
	dp := NewDiffPreview()
	r := Rect{X: 5, Y: 3, W: 60, H: 20}
	dp.SetBounds(r)
	b := dp.Bounds()
	if b.X != 5 || b.Y != 3 || b.W != 60 || b.H != 20 {
		t.Errorf("Bounds = %+v, want %+v", b, r)
	}
}

// --- Paint ---

func TestDiffPreview_Paint_NoPanic(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+added\n-removed\ncontext")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

func TestDiffPreview_Paint_Empty(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.Paint(buf)
}

func TestDiffPreview_Paint_ZeroBounds(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+test")
	buf := buffer.NewBuffer(10, 5)
	dp.Paint(buf)
}

func TestDiffPreview_Paint_SmallBounds(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+test line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 2, H: 2})
	buf := buffer.NewBuffer(2, 2)
	dp.Paint(buf)
}

func TestDiffPreview_Paint_RendersBorder(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+a")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 6})
	buf := buffer.NewBuffer(20, 6)
	dp.Paint(buf)

	tl := buf.GetCell(0, 0)
	if tl.Rune != '\u250C' {
		t.Errorf("top-left = %q, want %q", string(tl.Rune), "\u250C")
	}
	br := buf.GetCell(19, 5)
	if br.Rune != '\u2518' {
		t.Errorf("bottom-right = %q, want %q", string(br.Rune), "\u2518")
	}
}

func TestDiffPreview_Paint_RendersContent(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+new line")
	dp.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 8})
	buf := buffer.NewBuffer(30, 8)
	dp.Paint(buf)

	found := false
	for x := 1; x < 29; x++ {
		for y := 1; y < 7; y++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != ' ' && cell.Rune != 0 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("expected diff content in buffer")
	}
}

// --- Children / String ---

func TestDiffPreview_Children(t *testing.T) {
	dp := NewDiffPreview()
	if dp.Children() != nil {
		t.Error("Children() should return nil")
	}
}

func TestDiffPreview_String(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff("+a\n-b")
	s := dp.String()
	if !strings.Contains(s, "DiffPreview") {
		t.Errorf("String should contain 'DiffPreview', got %q", s)
	}
}

// --- DiffStats.String ---

func TestDiffStats_String(t *testing.T) {
	s := DiffStats{Additions: 5, Deletions: 3, Files: 2, Hunks: 1}
	str := s.String()
	if !strings.Contains(str, "+5") {
		t.Errorf("should contain '+5', got %q", str)
	}
	if !strings.Contains(str, "-3") {
		t.Errorf("should contain '-3', got %q", str)
	}
}

// --- ShowLineNumbers ---

func TestDiffPreview_SetShowLineNumbers(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetShowLineNumbers(false)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	dp.SetDiff("+test")
	dp.Paint(buf)
}

// --- SetLines ---

func TestDiffPreview_SetLines(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetLines([]DiffLine{
		{Type: DiffAdd, Content: "+new"},
		{Type: DiffDel, Content: "-old"},
	})
	if dp.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", dp.LineCount())
	}
}

// --- classifyDiffType ---

func TestClassifyDiffType_All(t *testing.T) {
	tests := []struct {
		input string
		want  DiffType
	}{
		{"diff --git a/file b/file", DiffFile},
		{"@@ -1,3 +1,3 @@", DiffHunk},
		{"+added", DiffAdd},
		{"-removed", DiffDel},
		{"--- a/file", DiffMeta},
		{"+++ b/file", DiffMeta},
		{"index abc..def", DiffMeta},
		{"new file mode 100644", DiffMeta},
		{"deleted file mode 100644", DiffMeta},
		{"context line", DiffContext},
	}
	for _, tt := range tests {
		got := classifyDiffType(tt.input)
		if got != tt.want {
			t.Errorf("classifyDiffType(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// --- ParseDiff ---

func TestParseDiff_Basic(t *testing.T) {
	text := `diff --git a/f b/f
@@ -1,3 +1,3 @@
-old
+new
 ctx`
	lines := ParseDiff(text)
	if len(lines) == 0 {
		t.Fatal("expected non-empty lines")
	}
	typeCounts := map[DiffType]int{}
	for _, l := range lines {
		typeCounts[l.Type]++
	}
	if typeCounts[DiffFile] == 0 {
		t.Error("expected DiffFile")
	}
	if typeCounts[DiffAdd] == 0 {
		t.Error("expected DiffAdd")
	}
	if typeCounts[DiffDel] == 0 {
		t.Error("expected DiffDel")
	}
}

// --- extractDiffFilename ---

func TestExtractDiffFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"diff --git a/file.go b/file.go", "file.go"},
		{"diff --git a/dir/sub.go b/dir/sub.go", "dir/sub.go"},
		{"diff --git a/no_b_part", ""},
	}
	for _, tt := range tests {
		got := extractDiffFilename(tt.input)
		if got != tt.want {
			t.Errorf("extractDiffFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Concurrency ---

func TestDiffPreview_ConcurrentAccess(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(sampleDiff)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				dp.Lines()
				dp.LineCount()
				dp.IsEmpty()
				dp.HasChanges()
				dp.Stats()
				dp.ScrollY()
				dp.ScrollDown(1)
				dp.ScrollUp(1)
			}
		}()
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				buf := buffer.NewBuffer(60, 10)
				dp.Paint(buf)
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 10; j++ {
			dp.SetDiff("+new\n-old")
		}
	}()
	wg.Wait()
}

func TestDiffPreview_ConcurrentPaint(t *testing.T) {
	dp := NewDiffPreview()
	dp.SetDiff(sampleDiff)
	dp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				buf := buffer.NewBuffer(40, 8)
				dp.Paint(buf)
			}
		}()
	}
	wg.Wait()
}
