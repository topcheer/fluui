package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestNewDiffViewer_Defaults(t *testing.T) {
	dv := NewDiffViewer()
	if !dv.ShowLineNumbers() {
		t.Error("expected line numbers on by default")
	}
	if dv.LineCount() != 0 {
		t.Errorf("expected 0 lines, got %d", dv.LineCount())
	}
	if dv.Title() != "" {
		t.Errorf("expected empty title, got %q", dv.Title())
	}
}

func TestDiffViewer_SetContent(t *testing.T) {
	diff := `diff --git a/file.go b/file.go
index abc..def 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 unchanged
-old line
+new line
+added line
 unchanged2
`
	dv := NewDiffViewer()
	dv.SetContent(diff)

	if dv.LineCount() < 8 {
		t.Errorf("expected at least 8 lines, got %d", dv.LineCount())
	}

	lines := dv.Lines()
	// Check line types
	hasAdd := false
	hasDel := false
	hasHunk := false
	for _, l := range lines {
		switch l.Type {
		case DiffAdd:
			hasAdd = true
		case DiffDel:
			hasDel = true
		case DiffHunk:
			hasHunk = true
		}
	}
	if !hasAdd {
		t.Error("expected at least one DiffAdd line")
	}
	if !hasDel {
		t.Error("expected at least one DiffDel line")
	}
	if !hasHunk {
		t.Error("expected at least one DiffHunk line")
	}
}

func TestDiffViewer_LineNumbers(t *testing.T) {
	diff := `--- a/f.go
+++ b/f.go
@@ -1,3 +1,3 @@
 context1
-old line
+new line
`
	dv := NewDiffViewer()
	dv.SetContent(diff)
	lines := dv.Lines()

	for _, l := range lines {
		if l.Type == DiffDel {
			if l.OldNo != 2 {
				t.Errorf("expected OldNo=2 for deleted line, got %d", l.OldNo)
			}
		}
		if l.Type == DiffAdd {
			if l.NewNo != 2 {
				t.Errorf("expected NewNo=2 for added line, got %d", l.NewNo)
			}
		}
		if l.Type == DiffContext && l.Content == "context1" {
			if l.OldNo != 1 || l.NewNo != 1 {
				t.Errorf("expected OldNo=1,NewNo=1 for context line, got %d,%d", l.OldNo, l.NewNo)
			}
		}
	}
}

func TestDiffViewer_SetLines(t *testing.T) {
	lines := []DiffLine{
		{Type: DiffContext, Content: "hello", OldNo: 1, NewNo: 1},
		{Type: DiffAdd, Content: "world", NewNo: 2},
		{Type: DiffDel, Content: "old", OldNo: 2},
	}
	dv := NewDiffViewer()
	dv.SetLines(lines)

	if dv.LineCount() != 3 {
		t.Fatalf("expected 3 lines, got %d", dv.LineCount())
	}

	got := dv.Lines()
	if got[0].Content != "hello" {
		t.Errorf("expected 'hello', got %q", got[0].Content)
	}
	if got[1].Type != DiffAdd {
		t.Errorf("expected DiffAdd, got %v", got[1].Type)
	}
}

func TestDiffViewer_LinesReturnsCopy(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines([]DiffLine{{Type: DiffAdd, Content: "x"}})
	l1 := dv.Lines()
	l1[0].Content = "modified"
	l2 := dv.Lines()
	if l2[0].Content == "modified" {
		t.Error("Lines() should return a defensive copy")
	}
}

func TestDiffViewer_ScrollDown(t *testing.T) {
	dv := NewDiffViewer()
	lines := make([]DiffLine, 100)
	for i := range lines {
		lines[i] = DiffLine{Type: DiffContext, Content: "line"}
	}
	dv.SetLines(lines)
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollDown(5)
	if dv.ScrollOffset() != 5 {
		t.Errorf("expected offset 5, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_ScrollDown_Clamped(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 20))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollDown(100)
	// maxScroll = 20 - 10 = 10
	if dv.ScrollOffset() != 10 {
		t.Errorf("expected clamped offset 10, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_ScrollUp(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollTo(10)
	dv.ScrollUp(3)
	if dv.ScrollOffset() != 7 {
		t.Errorf("expected offset 7, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_ScrollUp_ClampedToZero(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollUp(50)
	if dv.ScrollOffset() != 0 {
		t.Errorf("expected clamped offset 0, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_ScrollTo(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollTo(15)
	if dv.ScrollOffset() != 15 {
		t.Errorf("expected offset 15, got %d", dv.ScrollOffset())
	}

	dv.ScrollTo(-5)
	if dv.ScrollOffset() != 0 {
		t.Errorf("expected clamped offset 0, got %d", dv.ScrollOffset())
	}

	dv.ScrollTo(1000)
	// maxScroll = 100 - 10 = 90
	if dv.ScrollOffset() != 90 {
		t.Errorf("expected clamped offset 90, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_ScrollDown_NegativeNoOp(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollDown(-1)
	if dv.ScrollOffset() != 0 {
		t.Errorf("expected offset 0 for negative scroll, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_SetShowLineNumbers(t *testing.T) {
	dv := NewDiffViewer()
	if !dv.ShowLineNumbers() {
		t.Fatal("expected default true")
	}
	dv.SetShowLineNumbers(false)
	if dv.ShowLineNumbers() {
		t.Error("expected false after SetShowLineNumbers(false)")
	}
	dv.SetShowLineNumbers(true)
	if !dv.ShowLineNumbers() {
		t.Error("expected true after SetShowLineNumbers(true)")
	}
}

func TestDiffViewer_SetTitle(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetTitle("Changes in main.go")
	if dv.Title() != "Changes in main.go" {
		t.Errorf("expected title set, got %q", dv.Title())
	}
}

func TestDiffViewer_SetStyles(t *testing.T) {
	dv := NewDiffViewer()
	s := buffer.Style{Fg: buffer.NamedColor(buffer.NamedBlue)}

	dv.SetStyleAdd(s)
	dv.SetStyleDel(s)
	dv.SetStyleContext(s)
	dv.SetStyleHunk(s)
	dv.SetStyleMeta(s)
	dv.SetStyleLineNum(s)
	// Just verify no panic
}

func TestDiffViewer_Measure(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines([]DiffLine{
		{Type: DiffContext, Content: "short"},
		{Type: DiffAdd, Content: "a longer line of content here"},
	})

	sz := dv.Measure(Bounded(100, 100))
	if sz.H != 2 {
		t.Errorf("expected height 2, got %d", sz.H)
	}
}

func TestDiffViewer_Measure_WithConstraints(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 50))

	sz := dv.Measure(Bounded(40, 10))
	if sz.H > 10 {
		t.Errorf("expected height <= 10, got %d", sz.H)
	}
}

func TestDiffViewer_SetBounds_GetBounds(t *testing.T) {
	dv := NewDiffViewer()
	r := Rect{X: 5, Y: 10, W: 80, H: 24}
	dv.SetBounds(r)
	got := dv.Bounds()
	if got != r {
		t.Errorf("expected bounds %+v, got %+v", r, got)
	}
}

func TestDiffViewer_Paint_Basic(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines([]DiffLine{
		{Type: DiffAdd, Content: "added", NewNo: 1},
		{Type: DiffDel, Content: "removed", OldNo: 1},
		{Type: DiffContext, Content: "same", OldNo: 2, NewNo: 2},
	})
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	buf := buffer.NewBuffer(80, 10)
	dv.Paint(buf)

	// Check that something was painted
	hasContent := false
	for _, cell := range buf.Cells {
		if cell.Rune != 0 && cell.Rune != ' ' {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint should render content into the buffer")
	}
}

func TestDiffViewer_Paint_WithTitle(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetTitle("diff.go")
	dv.SetLines([]DiffLine{{Type: DiffContext, Content: "line"}})
	dv.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})

	buf := buffer.NewBuffer(20, 5)
	dv.Paint(buf)

	// First row should contain the title
	found := false
	titleRunes := []rune("diff.go")
	for i, r := range titleRunes {
		cell := buf.Cells[i] // row 0
		if cell.Rune == r {
			found = true
		}
	}
	if !found {
		t.Error("expected title to be painted on first row")
	}
}

func TestDiffViewer_Paint_WithScroll(t *testing.T) {
	dv := NewDiffViewer()
	lines := make([]DiffLine, 20)
	for i := range lines {
		lines[i] = DiffLine{Type: DiffContext, Content: "line" + itoa(i)}
	}
	dv.SetLines(lines)
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 5})

	dv.ScrollTo(10)
	buf := buffer.NewBuffer(80, 5)
	dv.Paint(buf)

	// Should paint lines 10-14 (5 visible)
	// Just verify content was painted
	hasContent := false
	for _, cell := range buf.Cells {
		if cell.Rune == 'l' {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("expected scrolled content to be painted")
	}
}

func TestDiffViewer_Paint_ZeroBounds_NoOp(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines([]DiffLine{{Type: DiffContext, Content: "x"}})
	dv.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})

	buf := buffer.NewBuffer(80, 24)
	dv.Paint(buf)
	// Should not panic
}

func TestDiffViewer_Paint_EmptyLines(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 24})
	buf := buffer.NewBuffer(80, 24)
	dv.Paint(buf)
	// Should not panic
}

func TestDiffViewer_HandleKey_ArrowDown(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	handled := dv.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if !handled {
		t.Error("expected KeyDown to be handled")
	}
	if dv.ScrollOffset() != 1 {
		t.Errorf("expected offset 1, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_ArrowUp(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollTo(5)
	handled := dv.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if !handled {
		t.Error("expected KeyUp to be handled")
	}
	if dv.ScrollOffset() != 4 {
		t.Errorf("expected offset 4, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_PageDown(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	handled := dv.HandleKey(&term.KeyEvent{Key: term.KeyPageDown})
	if !handled {
		t.Error("expected PageDown to be handled")
	}
	// visibleHeight = 10, scroll by 10
	if dv.ScrollOffset() != 10 {
		t.Errorf("expected offset 10, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_PageUp(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollTo(15)
	handled := dv.HandleKey(&term.KeyEvent{Key: term.KeyPageUp})
	if !handled {
		t.Error("expected PageUp to be handled")
	}
	if dv.ScrollOffset() != 5 {
		t.Errorf("expected offset 5, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_VimJK(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	handled := dv.HandleKey(&term.KeyEvent{Rune: 'j'})
	if !handled || dv.ScrollOffset() != 1 {
		t.Errorf("expected 'j' to scroll down, handled=%v offset=%d", handled, dv.ScrollOffset())
	}

	handled = dv.HandleKey(&term.KeyEvent{Rune: 'k'})
	if !handled || dv.ScrollOffset() != 0 {
		t.Errorf("expected 'k' to scroll up, handled=%v offset=%d", handled, dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_VimG_GotoTop(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	dv.ScrollTo(20)
	handled := dv.HandleKey(&term.KeyEvent{Rune: 'g'})
	if !handled || dv.ScrollOffset() != 0 {
		t.Errorf("expected 'g' to goto top, handled=%v offset=%d", handled, dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_VimG_GotoBottom(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	handled := dv.HandleKey(&term.KeyEvent{Rune: 'G'})
	if !handled {
		t.Error("expected 'G' to be handled")
	}
	// maxScroll = 100 - 10 = 90
	if dv.ScrollOffset() != 90 {
		t.Errorf("expected offset 90, got %d", dv.ScrollOffset())
	}
}

func TestDiffViewer_HandleKey_Unhandled(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	handled := dv.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if handled {
		t.Error("expected Enter to not be handled")
	}
}

func TestDiffViewer_HandleKey_Nil(t *testing.T) {
	dv := NewDiffViewer()
	handled := dv.HandleKey(nil)
	if handled {
		t.Error("expected nil key to not be handled")
	}
}

func TestDiffViewer_Children(t *testing.T) {
	dv := NewDiffViewer()
	if dv.Children() != nil {
		t.Error("expected nil children for leaf component")
	}
}

func TestDiffViewer_Concurrent(t *testing.T) {
	dv := NewDiffViewer()
	dv.SetLines(make([]DiffLine, 100))
	dv.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 10})

	done := make(chan struct{})
	// Concurrent reader
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			dv.Lines()
			dv.ScrollOffset()
			dv.LineCount()
			buf := buffer.NewBuffer(80, 10)
			dv.Paint(buf)
		}
	}()

	// Concurrent writer
	for i := 0; i < 100; i++ {
		dv.ScrollDown(1)
		dv.ScrollUp(1)
	}

	<-done
}

func TestParseDiffWithLineNums_FullDiff(t *testing.T) {
	diff := `diff --git a/f.go b/f.go
index abc..def 100644
--- a/f.go
+++ b/f.go
@@ -1,3 +1,4 @@
 line1
-removed
+added
+new
 line2
`
	lines := ParseDiffWithLineNums(diff)

	// Verify line types
	fileCount := 0
	metaCount := 0
	hunkCount := 0
	addCount := 0
	delCount := 0
	ctxCount := 0

	for _, l := range lines {
		switch l.Type {
		case DiffFile:
			fileCount++
		case DiffMeta:
			metaCount++
		case DiffHunk:
			hunkCount++
		case DiffAdd:
			addCount++
		case DiffDel:
			delCount++
		case DiffContext:
			ctxCount++
		}
	}

	if fileCount != 2 {
		t.Errorf("expected 2 DiffFile, got %d", fileCount)
	}
	if metaCount != 2 {
		t.Errorf("expected 2 DiffMeta, got %d", metaCount)
	}
	if hunkCount != 1 {
		t.Errorf("expected 1 DiffHunk, got %d", hunkCount)
	}
	if addCount != 2 {
		t.Errorf("expected 2 DiffAdd, got %d", addCount)
	}
	if delCount != 1 {
		t.Errorf("expected 1 DiffDel, got %d", delCount)
	}
	if ctxCount != 2 {
		t.Errorf("expected 2 DiffContext, got %d", ctxCount)
	}
}

func TestParseDiffWithLineNums_Empty(t *testing.T) {
	lines := ParseDiffWithLineNums("")
	if len(lines) != 0 {
		t.Errorf("expected 0 lines for empty input, got %d", len(lines))
	}
}

func TestParseDiffWithLineNums_LineNumberTracking(t *testing.T) {
	diff := `--- a/f.go
+++ b/f.go
@@ -1,5 +1,5 @@
 a
 b
-c
+d
 e
`
	lines := ParseDiffWithLineNums(diff)

	for _, l := range lines {
		if l.Type == DiffDel && l.Content == "c" {
			if l.OldNo != 3 {
				t.Errorf("expected OldNo=3 for deleted 'c', got %d", l.OldNo)
			}
		}
		if l.Type == DiffAdd && l.Content == "d" {
			if l.NewNo != 3 {
				t.Errorf("expected NewNo=3 for added 'd', got %d", l.NewNo)
			}
		}
	}
}

func TestParseDiffWithLineNums_HunkHeaderReset(t *testing.T) {
	diff := `--- a/f.go
+++ b/f.go
@@ -1,2 +1,2 @@
 a
-b
+c
@@ -10,2 +10,2 @@
 d
-e
+f
`
	lines := ParseDiffWithLineNums(diff)

	// After second hunk, line numbers should reset to 10
	for _, l := range lines {
		if l.Type == DiffDel && l.Content == "e" {
			if l.OldNo != 11 {
				t.Errorf("expected OldNo=11 for 'e' after hunk reset, got %d", l.OldNo)
			}
		}
		if l.Type == DiffAdd && l.Content == "f" {
			if l.NewNo != 11 {
				t.Errorf("expected NewNo=11 for 'f' after hunk reset, got %d", l.NewNo)
			}
		}
	}
}

func TestFormatLineNum(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "      "},
		{1, "     1"},
		{5, "     5"},
		{50, "    50"},
		{500, "   500"},
		{5000, "  5000"},
		{50000, " 50000"},
	}
	for _, tc := range tests {
		got := formatLineNum(tc.input)
		if len(got) != 6 {
			t.Errorf("formatLineNum(%d) = %q (len %d), want length 6", tc.input, got, len(got))
		}
	}
}

func TestParseHunkHeader(t *testing.T) {
	tests := []struct {
		input   string
		oldWant int
		newWant int
	}{
		{"@@ -1,5 +1,5 @@", 1, 1},
		{"@@ -10,3 +10,4 @@", 10, 10},
		{"@@ -1 +1 @@", 1, 1},
	}
	for _, tc := range tests {
		old, nw := parseHunkHeader(tc.input)
		if old != tc.oldWant || nw != tc.newWant {
			t.Errorf("parseHunkHeader(%q) = (%d, %d), want (%d, %d)", tc.input, old, nw, tc.oldWant, tc.newWant)
		}
	}
}

func TestParseHunkHeader_Invalid(t *testing.T) {
	old, nw := parseHunkHeader("no hunk here")
	if old != 0 || nw != 0 {
		t.Errorf("expected (0, 0) for invalid header, got (%d, %d)", old, nw)
	}
}
