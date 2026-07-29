package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestMarkdownLinkBasic(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("Visit [Fluui](https://fluui.dev) now.")
	if ml.LinkCount() != 1 {
		t.Errorf("LinkCount = %d, want 1", ml.LinkCount())
	}
}

func TestMarkdownLinkMultiple(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("[A](url1) and [B](url2) and [C](url3)")
	if ml.LinkCount() != 3 {
		t.Errorf("LinkCount = %d, want 3", ml.LinkCount())
	}
}

func TestMarkdownLinkRefStyle(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("[Fluui][1]\n\n[1]: https://fluui.dev")
	if ml.LinkCount() != 1 {
		t.Errorf("LinkCount = %d, want 1", ml.LinkCount())
	}
}

func TestMarkdownLinkAutolink(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("Visit <https://example.com> today")
	if ml.LinkCount() != 1 {
		t.Errorf("LinkCount = %d, want 1", ml.LinkCount())
	}
}

func TestMarkdownLinkNoLinks(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("Just plain text here.")
	if ml.LinkCount() != 0 {
		t.Errorf("LinkCount = %d, want 0", ml.LinkCount())
	}
}

func TestMarkdownLinkEmpty(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("")
	if ml.LinkCount() != 0 {
		t.Errorf("LinkCount = %d, want 0", ml.LinkCount())
	}
}

func TestMarkdownLinkMeasure(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("[link](url)")
	s := ml.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestMarkdownLinkPaint(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetMarkdown("Click [here](https://example.com) now.")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 4})
	buf := buffer.NewBuffer(50, 4)
	ml.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
}

func TestMarkdownLinkChildren(t *testing.T) {
	ml := NewMarkdownLink()
	if ml.Children() != nil { t.Error("Children should be nil") }
}

func TestMarkdownLinkStyle(t *testing.T) {
	ml := NewMarkdownLink()
	ml.SetStyle(MarkdownLinkStyle{
		Text:   buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Link:   buffer.Style{Fg: buffer.RGB(0, 255, 0), Flags: buffer.Underline},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	ml.SetMarkdown("[x](y)")
	ml.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 4})
	buf := buffer.NewBuffer(40, 4)
	ml.Paint(buf)
}

// ─── TagCloud tests ───

func TestTagCloudBasic(t *testing.T) {
	tc := NewTagCloud()
	tc.AddTag("go", 50)
	tc.AddTag("tui", 30)
	if tc.TagCount() != 2 { t.Errorf("TagCount = %d, want 2", tc.TagCount()) }
}

func TestTagCloudMultiple(t *testing.T) {
	tc := NewTagCloud()
	tc.AddTag("alpha", 10)
	tc.AddTag("beta", 20)
	tc.AddTag("gamma", 30)
	tc.AddTag("delta", 40)
	if tc.TagCount() != 4 { t.Errorf("TagCount = %d, want 4", tc.TagCount()) }
}

func TestTagCloudSetTags(t *testing.T) {
	tc := NewTagCloud()
	tc.SetTags([]TagItem{{Name: "x", Weight: 1}, {Name: "y", Weight: 2}})
	if tc.TagCount() != 2 { t.Errorf("TagCount = %d, want 2", tc.TagCount()) }
}

func TestTagCloudWeightScaling(t *testing.T) {
	tc := NewTagCloud()
	tc.AddTag("small", 1)
	tc.AddTag("large", 100)
	tc.mu.Lock()
	tc.computeSortedLocked()
	maxW := tc.maxWeightLocked()
	if maxW != 100 { t.Errorf("maxWeight = %d, want 100", maxW) }
	tc.mu.Unlock()
}

func TestTagCloudEmpty(t *testing.T) {
	tc := NewTagCloud()
	if tc.TagCount() != 0 { t.Errorf("TagCount = %d, want 0", tc.TagCount()) }
}

func TestTagCloudSortedAlphabetically(t *testing.T) {
	tc := NewTagCloud()
	tc.AddTag("zebra", 10)
	tc.AddTag("apple", 20)
	tc.AddTag("mango", 30)
	tc.mu.Lock()
	tc.computeSortedLocked()
	if tc.cachedSorted[0].Name != "apple" { t.Errorf("first = %s, want apple", tc.cachedSorted[0].Name) }
	if tc.cachedSorted[2].Name != "zebra" { t.Errorf("last = %s, want zebra", tc.cachedSorted[2].Name) }
	tc.mu.Unlock()
}

func TestTagCloudMeasure(t *testing.T) {
	tc := NewTagCloud()
	tc.AddTag("a", 1)
	s := tc.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
}

func TestTagCloudPaint(t *testing.T) {
	tc := NewTagCloud()
	tc.AddTag("go", 50)
	tc.AddTag("tui", 30)
	tc.AddTag("ai", 80)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 6})
	buf := buffer.NewBuffer(50, 6)
	tc.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
}

func TestTagCloudPaintEmpty(t *testing.T) {
	tc := NewTagCloud()
	tc.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	tc.Paint(buf)
}

func TestTagCloudChildren(t *testing.T) {
	tc := NewTagCloud()
	if tc.Children() != nil { t.Error("Children should be nil") }
}

func TestTagCloudStyle(t *testing.T) {
	tc := NewTagCloud()
	tc.SetStyle(TagCloudStyle{
		Small:  buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Medium: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Large:  buffer.Style{Fg: buffer.RGB(255, 0, 0), Flags: buffer.Bold},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	tc.AddTag("test", 100)
	tc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	tc.Paint(buf)
}
