package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP410_FilteredList_New(t *testing.T) {
	fl := NewFilteredList([]string{"apple", "banana", "cherry"})
	if len(fl.Items()) != 3 { t.Errorf("Items = %d", len(fl.Items())) }
	if len(fl.Filtered()) != 3 { t.Errorf("Filtered = %d", len(fl.Filtered())) }
	if fl.ID() == "" { t.Error("ID empty") }
}

func TestP410_FilteredList_SetQuery(t *testing.T) {
	fl := NewFilteredList([]string{"apple", "banana", "cherry", "apricot"})
	fl.SetQuery("ap")
	got := fl.Filtered()
	if len(got) != 2 { t.Errorf("Filtered = %d, want 2", len(got)) }
	if got[0] != "apple" { t.Errorf("got[0] = %q", got[0]) }
	if got[1] != "apricot" { t.Errorf("got[1] = %q", got[1]) }
}

func TestP410_FilteredList_SetQuery_CaseInsensitive(t *testing.T) {
	fl := NewFilteredList([]string{"Hello", "world", "HELP"})
	fl.SetQuery("he")
	got := fl.Filtered()
	if len(got) != 2 { t.Errorf("Filtered = %d, want 2", len(got)) }
}

func TestP410_FilteredList_SetQuery_NoMatch(t *testing.T) {
	fl := NewFilteredList([]string{"a", "b"})
	fl.SetQuery("xyz")
	if len(fl.Filtered()) != 0 { t.Errorf("Filtered = %d, want 0", len(fl.Filtered())) }
}

func TestP410_FilteredList_SetQuery_Empty(t *testing.T) {
	fl := NewFilteredList([]string{"a", "b"})
	fl.SetQuery("a")
	fl.SetQuery("")
	if len(fl.Filtered()) != 2 { t.Errorf("Filtered = %d, want 2", len(fl.Filtered())) }
}

func TestP410_FilteredList_MoveUp(t *testing.T) {
	fl := NewFilteredList([]string{"a", "b", "c"})
	fl.MoveDown()
	fl.MoveDown()
	fl.MoveUp()
	if fl.Selected() != 1 { t.Errorf("Selected = %d, want 1", fl.Selected()) }
	// Can't go above 0
	fl.MoveUp()
	fl.MoveUp()
	if fl.Selected() != 0 { t.Errorf("Selected = %d, want 0", fl.Selected()) }
}

func TestP410_FilteredList_MoveDown(t *testing.T) {
	fl := NewFilteredList([]string{"a", "b"})
	fl.MoveDown()
	if fl.Selected() != 1 { t.Errorf("Selected = %d", fl.Selected()) }
	fl.MoveDown()
	if fl.Selected() != 1 { t.Errorf("Selected = %d, can't exceed", fl.Selected()) }
}

func TestP410_FilteredList_SelectedItem(t *testing.T) {
	fl := NewFilteredList([]string{"x", "y", "z"})
	fl.MoveDown()
	if fl.SelectedItem() != "y" { t.Errorf("SelectedItem = %q", fl.SelectedItem()) }
}

func TestP410_FilteredList_SetItems(t *testing.T) {
	fl := NewFilteredList([]string{"old"})
	fl.SetItems([]string{"new1", "new2"})
	if len(fl.Items()) != 2 { t.Errorf("Items = %d", len(fl.Items())) }
}

func TestP410_FilteredList_Measure(t *testing.T) {
	fl := NewFilteredList([]string{"short", "very long item name"})
	s := fl.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	if s.W < 10 { t.Errorf("W = %d, too small", s.W) }
	if s.H < 2 { t.Errorf("H = %d", s.H) }
}

func TestP410_FilteredList_Paint(t *testing.T) {
	fl := NewFilteredList([]string{"item1", "item2"})
	fl.MoveDown() // select item2
	fl.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	fl.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'i' { t.Errorf("cell[0,0] = %q", string(c.Rune)) }
}

func TestP410_FilteredList_Paint_NoMatches(t *testing.T) {
	fl := NewFilteredList([]string{"a", "b"})
	fl.SetQuery("xyz")
	fl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	fl.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'N' { t.Errorf("cell[0,0] = %q, want 'N' (No matches)", string(c.Rune)) }
}

func TestP410_FilteredList_Paint_ZeroBounds(t *testing.T) {
	fl := NewFilteredList([]string{"a"})
	fl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	fl.Paint(buf)
}

func TestP410_FilteredList_Concurrent(t *testing.T) {
	fl := NewFilteredList([]string{"a", "b"})
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ { fl.SetQuery("a") }
		close(done)
	}()
	for i := 0; i < 500; i++ { _ = fl.Filtered() }
	<-done
}

func TestP410_FilteredList_SatisfiesComponent(t *testing.T) {
	var _ Component = (*FilteredList)(nil)
}

func BenchmarkP410_FilteredList_Paint(b *testing.B) {
	items := make([]string, 20)
	for i := range items { items[i] = "item-" + string(rune('a'+i)) }
	fl := NewFilteredList(items)
	fl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})
	buf := buffer.NewBuffer(20, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { fl.Paint(buf) }
}
