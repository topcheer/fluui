package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP342_Accordion_Create(t *testing.T) {
	items := []AccordionItem{
		{Title: "General", Content: "Settings"},
		{Title: "Advanced", Content: "Advanced settings"},
	}
	a := NewAccordion(items)
	if a.ItemCount() != 2 {
		t.Errorf("count = %d, want 2", a.ItemCount())
	}
}

func TestP342_Accordion_Toggle(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "content a"},
		{Title: "B", Content: "content b"},
	})
	if a.IsExpanded(0) {
		t.Error("should start collapsed")
	}
	a.Toggle(0)
	if !a.IsExpanded(0) {
		t.Error("should be expanded after toggle")
	}
	a.Toggle(0)
	if a.IsExpanded(0) {
		t.Error("should be collapsed after second toggle")
	}
}

func TestP342_Accordion_SingleOpen(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a"},
		{Title: "B", Content: "b", Expanded: true},
		{Title: "C", Content: "c", Expanded: true},
	})
	a.SetSingleOpen(true)
	// Only one should be expanded
	expanded := 0
	for _, item := range a.Items() {
		if item.Expanded {
			expanded++
		}
	}
	if expanded != 1 {
		t.Errorf("singleOpen should allow 1 expanded, got %d", expanded)
	}
}

func TestP342_Accordion_SingleOpen_Toggle(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a"},
		{Title: "B", Content: "b"},
	})
	a.SetSingleOpen(true)
	a.Toggle(0) // open A
	if !a.IsExpanded(0) {
		t.Error("A should be expanded")
	}
	a.Toggle(1) // open B
	if a.IsExpanded(0) {
		t.Error("A should collapse when B opens (singleOpen)")
	}
	if !a.IsExpanded(1) {
		t.Error("B should be expanded")
	}
}

func TestP342_Accordion_ExpandCollapse(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a"},
		{Title: "B", Content: "b"},
	})
	a.Expand(0)
	if !a.IsExpanded(0) {
		t.Error("should be expanded")
	}
	a.Collapse(0)
	if a.IsExpanded(0) {
		t.Error("should be collapsed")
	}
}

func TestP342_Accordion_ExpandAll(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a"},
		{Title: "B", Content: "b"},
	})
	a.ExpandAll()
	if !a.IsExpanded(0) || !a.IsExpanded(1) {
		t.Error("all should be expanded")
	}
}

func TestP342_Accordion_ExpandAll_SingleOpen(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a"},
	})
	a.SetSingleOpen(true)
	a.ExpandAll() // should be no-op
	if a.IsExpanded(0) {
		t.Error("ExpandAll should not work with singleOpen")
	}
}

func TestP342_Accordion_CollapseAll(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "A", Content: "a", Expanded: true},
		{Title: "B", Content: "b", Expanded: true},
	})
	a.CollapseAll()
	if a.IsExpanded(0) || a.IsExpanded(1) {
		t.Error("all should be collapsed")
	}
}

func TestP342_Accordion_InvalidIndex(t *testing.T) {
	a := NewAccordion([]AccordionItem{{Title: "A", Content: "a"}})
	a.Toggle(-1)  // no panic
	a.Toggle(99)  // no panic
	a.Expand(-1)  // no panic
	a.Collapse(99) // no panic
	if a.IsExpanded(-1) {
		t.Error("invalid index should return false")
	}
}

func TestP342_Accordion_SetItems(t *testing.T) {
	a := NewAccordion(nil)
	a.SetItems([]AccordionItem{{Title: "New", Content: "content"}})
	if a.ItemCount() != 1 {
		t.Errorf("count = %d", a.ItemCount())
	}
}

func TestP342_Accordion_Measure(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "Collapsed", Content: "hidden"},
		{Title: "Open", Content: "line1\nline2", Expanded: true},
	})
	s := a.Measure(Constraints{MaxWidth: 40, MaxHeight: 20})
	// Collapsed = 1 line, Open header + 2 content lines = 3 lines
	if s.H < 4 {
		t.Errorf("height = %d, expected at least 4", s.H)
	}
}

func TestP342_Accordion_Paint(t *testing.T) {
	a := NewAccordion([]AccordionItem{
		{Title: "Collapsed", Content: "hidden content"},
		{Title: "Expanded", Content: "visible\nmulti\nline", Expanded: true},
	})
	a.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	a.Paint(buf)

	// First header should have ▸ marker
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected non-empty cell at header")
	}
}

func TestP342_Accordion_Paint_Empty(t *testing.T) {
	a := NewAccordion(nil)
	a.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	a.Paint(buf) // should not panic
}

func TestP342_Accordion_Paint_Truncation(t *testing.T) {
	longTitle := "This is a very long title that should be truncated with ellipsis"
	longContent := "This is very long content that exceeds the width and needs truncation"
	a := NewAccordion([]AccordionItem{
		{Title: longTitle, Content: longContent, Expanded: true},
	})
	a.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	a.Paint(buf) // should not panic
}

func BenchmarkAccordion_Paint(b *testing.B) {
	items := make([]AccordionItem, 5)
	for i := range items {
		items[i] = AccordionItem{
			Title:   "Section " + string(rune('A'+i)),
			Content: "Line one\nLine two\nLine three",
			Expanded: i%2 == 0,
		}
	}
	a := NewAccordion(items)
	a.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	buf := buffer.NewBuffer(40, 20)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a.Paint(buf)
	}
}
