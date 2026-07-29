package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownListUnordered(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("- Apple\n- Banana\n- Cherry")
	if ml.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", ml.ItemCount())
	}
	if ml.ListType() != "unordered" {
		t.Errorf("ListType = %q, want unordered", ml.ListType())
	}
}

func TestMarkdownListOrdered(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("1. First\n2. Second\n3. Third")
	if ml.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", ml.ItemCount())
	}
	if ml.ListType() != "ordered" {
		t.Errorf("ListType = %q, want ordered", ml.ListType())
	}
}

func TestMarkdownListNested(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("- Top\n  - Nested\n    - Deep")
	if ml.ItemCount() != 3 {
		t.Errorf("ItemCount = %d, want 3", ml.ItemCount())
	}
	ml.mu.Lock()
	if ml.cachedLines[1].Indent != 1 {
		t.Errorf("nested indent = %d, want 1", ml.cachedLines[1].Indent)
	}
	if ml.cachedLines[2].Indent != 2 {
		t.Errorf("deep indent = %d, want 2", ml.cachedLines[2].Indent)
	}
	ml.mu.Unlock()
}

func TestMarkdownListMixed(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("- Unordered\n1. Ordered")
	if ml.ListType() != "mixed" {
		t.Errorf("ListType = %q, want mixed", ml.ListType())
	}
}

func TestMarkdownListEmpty(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("")
	if ml.ItemCount() != 0 {
		t.Errorf("ItemCount = %d, want 0", ml.ItemCount())
	}
	if ml.ListType() != "" {
		t.Errorf("ListType = %q, want empty", ml.ListType())
	}
}

func TestMarkdownListMeasure(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("- A\n- B")
	s := ml.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H < 4 {
		t.Errorf("H = %d, want >= 4", s.H)
	}
}

func TestMarkdownListPaint(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("- Apple\n  - Red\n- Banana")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 6})
	buf := buffer.NewBuffer(50, 6)
	ml.Paint(buf)

	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}
	// Check bullet char
	foundBullet := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '•' {
			foundBullet = true
			break
		}
	}
	if !foundBullet {
		t.Error("bullet char not found")
	}
}

func TestMarkdownListPaintOrdered(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetMarkdown("1. First\n2. Second")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	ml.Paint(buf)

	// Check number "1" rendered
	foundNum := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '1' {
			foundNum = true
			break
		}
	}
	if !foundNum {
		t.Error("ordered number not found")
	}
}

func TestMarkdownListChildren(t *testing.T) {
	ml := NewMarkdownList()
	if ml.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestMarkdownListStyle(t *testing.T) {
	ml := NewMarkdownList()
	ml.SetStyle(MarkdownListStyle{
		Text:   buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Bullet: [3]buffer.Style{{Fg: buffer.RGB(255, 0, 0)}, {Fg: buffer.RGB(0, 255, 0)}, {Fg: buffer.RGB(0, 0, 255)}},
		Number: buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	ml.SetMarkdown("- Test\n  - Nested")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ml.Paint(buf)
}

// ─── BreadcrumbTrail tests ───

func TestBreadcrumbTrailBasic(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.AddCrumb("Home")
	bt.AddCrumb("Files")
	if bt.CrumbCount() != 2 {
		t.Errorf("CrumbCount = %d, want 2", bt.CrumbCount())
	}
}

func TestBreadcrumbTrailSetCrumbs(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.SetCrumbs([]string{"A", "B", "C", "D"})
	if bt.CrumbCount() != 4 {
		t.Errorf("CrumbCount = %d, want 4", bt.CrumbCount())
	}
}

func TestBreadcrumbTrailSeparator(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.SetSeparator('/')
	bt.AddCrumb("a")
	bt.AddCrumb("b")
	bt.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	bt.Paint(buf)
	// Check separator rendered
	foundSep := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == '/' {
			foundSep = true
			break
		}
	}
	if !foundSep {
		t.Error("separator not found")
	}
}

func TestBreadcrumbTrailShort(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.AddCrumb("A")
	bt.AddCrumb("B")
	bt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	bt.Paint(buf)
	// Should find both crumbs, no ellipsis
	foundEllipsis := false
	for x := 0; x < 30; x++ {
		if buf.GetCell(x, 0).Rune == '…' {
			foundEllipsis = true
			break
		}
	}
	if foundEllipsis {
		t.Error("should not have ellipsis for short path")
	}
}

func TestBreadcrumbTrailTruncation(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.AddCrumb("VeryLongSegmentOne")
	bt.AddCrumb("VeryLongSegmentTwo")
	bt.AddCrumb("VeryLongSegmentThree")
	bt.AddCrumb("VeryLongSegmentFour")
	bt.AddCrumb("Current")
	bt.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	bt.Paint(buf)
	// Should find ellipsis due to truncation
	foundEllipsis := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 0).Rune == '…' {
			foundEllipsis = true
			break
		}
	}
	if !foundEllipsis {
		t.Error("expected ellipsis for truncated long path")
	}
}

func TestBreadcrumbTrailEmpty(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	bt.Paint(buf) // should not panic
}

func TestBreadcrumbTrailMeasure(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.AddCrumb("A")
	s := bt.Measure(Constraints{})
	if s.W < 10 {
		t.Errorf("W = %d, want >= 10", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d, want 1", s.H)
	}
}

func TestBreadcrumbTrailPaint(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.AddCrumb("Home")
	bt.AddCrumb("Projects")
	bt.AddCrumb("Fluui")
	bt.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	bt.Paint(buf)

	// Check "Home" is rendered
	foundHome := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == 'H' {
			foundHome = true
			break
		}
	}
	if !foundHome {
		t.Error("'Home' crumb not found")
	}
}

func TestBreadcrumbTrailChildren(t *testing.T) {
	bt := NewBreadcrumbTrail()
	if bt.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestBreadcrumbTrailStyle(t *testing.T) {
	bt := NewBreadcrumbTrail()
	bt.SetStyle(BreadcrumbStyle{
		Crumb:     buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Active:    buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold},
		Separator: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Ellipsis:  buffer.Style{Fg: buffer.RGB(120, 120, 120)},
		Border:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	bt.AddCrumb("A")
	bt.AddCrumb("B")
	bt.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	bt.Paint(buf)
}
