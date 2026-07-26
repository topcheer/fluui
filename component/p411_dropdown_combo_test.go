package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === Dropdown tests ===

func TestP411_Dropdown_New(t *testing.T) {
	d := NewDropdown([]DropdownItem{
		{Label: "Option A", Value: "a"},
		{Label: "Option B", Value: "b"},
	})
	if len(d.Items()) != 2 { t.Errorf("Items = %d", len(d.Items())) }
	if d.Selected() != 0 { t.Errorf("Selected = %d", d.Selected()) }
	if d.ID() == "" { t.Error("ID empty") }
}

func TestP411_Dropdown_SetSelected(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}, {Label: "B"}, {Label: "C"}})
	d.SetSelected(2)
	if d.Selected() != 2 { t.Errorf("Selected = %d", d.Selected()) }
	if d.SelectedItem().Label != "C" { t.Errorf("Label = %q", d.SelectedItem().Label) }
}

func TestP411_Dropdown_SetSelected_OutOfRange(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}})
	d.SetSelected(5)
	if d.Selected() != 0 { t.Errorf("Should stay 0") }
	d.SetSelected(-1)
	if d.Selected() != 0 { t.Errorf("Should stay 0") }
}

func TestP411_Dropdown_Expand(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}})
	if d.Expanded() { t.Error("should be collapsed") }
	d.Toggle()
	if !d.Expanded() { t.Error("should be expanded") }
	d.SetExpanded(false)
	if d.Expanded() { t.Error("should be collapsed") }
}

func TestP411_Dropdown_MoveUp(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}, {Label: "B"}})
	d.MoveDown()
	d.MoveUp()
	if d.Selected() != 0 { t.Errorf("Selected = %d", d.Selected()) }
	d.MoveUp() // can't go above 0
	if d.Selected() != 0 { t.Errorf("Selected = %d", d.Selected()) }
}

func TestP411_Dropdown_MoveDown(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}, {Label: "B"}})
	d.MoveDown()
	if d.Selected() != 1 { t.Errorf("Selected = %d", d.Selected()) }
	d.MoveDown()
	if d.Selected() != 1 { t.Errorf("Can't exceed: %d", d.Selected()) }
}

func TestP411_Dropdown_SetLabel(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}})
	d.SetLabel("Model")
	if d.Label() != "Model" { t.Errorf("Label = %q", d.Label()) }
}

func TestP411_Dropdown_SetItems(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "old"}})
	d.SetItems([]DropdownItem{{Label: "new1"}, {Label: "new2"}})
	if len(d.Items()) != 2 { t.Errorf("Items = %d", len(d.Items())) }
}

func TestP411_Dropdown_Measure(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "Short"}, {Label: "Very long label"}})
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d", s.H) }
}

func TestP411_Dropdown_Measure_Expanded(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}, {Label: "B"}, {Label: "C"}})
	d.SetExpanded(true)
	s := d.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	if s.H != 4 { t.Errorf("H = %d, want 4 (header + 3 items)", s.H) }
}

func TestP411_Dropdown_Paint_Collapsed(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "Test"}})
	d.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	d.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '[' { t.Errorf("cell = %q, want '['", string(c.Rune)) }
}

func TestP411_Dropdown_Paint_Expanded(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}, {Label: "B"}})
	d.SetExpanded(true)
	d.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 3})
	buf := buffer.NewBuffer(10, 3)
	d.Paint(buf)
	c := buf.GetCell(0, 1) // first item
	if c.Rune != 'A' { t.Errorf("item cell = %q, want 'A'", string(c.Rune)) }
}

func TestP411_Dropdown_Paint_ZeroBounds(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}})
	d.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	d.Paint(buf)
}

func TestP411_Dropdown_Concurrent(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}, {Label: "B"}})
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { d.MoveDown() }; close(done) }()
	for i := 0; i < 500; i++ { _ = d.Selected() }
	<-done
}

func BenchmarkP411_Dropdown_Paint(b *testing.B) {
	d := NewDropdown([]DropdownItem{{Label: "GPT-4"}, {Label: "Claude-3"}, {Label: "Gemini"}})
	d.SetExpanded(true)
	d.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 4})
	buf := buffer.NewBuffer(20, 4)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { d.Paint(buf) }
}

// === ComboBox tests ===

func TestP411_ComboBox_New(t *testing.T) {
	cb := NewComboBox([]string{"apple", "banana"})
	if len(cb.Items()) != 2 { t.Errorf("Items = %d", len(cb.Items())) }
	if cb.ID() == "" { t.Error("ID empty") }
}

func TestP411_ComboBox_SetQuery(t *testing.T) {
	cb := NewComboBox([]string{"apple", "apricot", "banana"})
	cb.SetQuery("ap")
	if len(cb.Filtered()) != 2 { t.Errorf("Filtered = %d, want 2", len(cb.Filtered())) }
	if !cb.Expanded() { t.Error("should be expanded when query matches") }
}

func TestP411_ComboBox_SetQuery_NoMatch(t *testing.T) {
	cb := NewComboBox([]string{"a", "b"})
	cb.SetQuery("xyz")
	if len(cb.Filtered()) != 0 { t.Errorf("Filtered = %d", len(cb.Filtered())) }
	if cb.Expanded() { t.Error("should not expand with no matches") }
}

func TestP411_ComboBox_SelectCurrent(t *testing.T) {
	cb := NewComboBox([]string{"apple", "banana"})
	cb.SetQuery("ap")
	cb.MoveDown()
	cb.SelectCurrent()
	if cb.Query() != "apricot" && cb.Query() != "apple" { t.Errorf("Query = %q", cb.Query()) }
	if cb.Expanded() { t.Error("should collapse after select") }
}

func TestP411_ComboBox_MoveUp(t *testing.T) {
	cb := NewComboBox([]string{"a", "b"})
	cb.MoveDown()
	cb.MoveUp()
	if cb.Selected() != 0 { t.Errorf("Selected = %d", cb.Selected()) }
}

func TestP411_ComboBox_Paint(t *testing.T) {
	cb := NewComboBox([]string{"item1", "item2"})
	cb.SetQuery("item")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 'i' { t.Errorf("cell = %q", string(c.Rune)) }
}

func TestP411_ComboBox_Paint_EmptyQuery(t *testing.T) {
	cb := NewComboBox([]string{"a"})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	cb.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 't' { t.Errorf("cell = %q, want 't' (type to search)", string(c.Rune)) }
}

func TestP411_ComboBox_Paint_ZeroBounds(t *testing.T) {
	cb := NewComboBox([]string{"a"})
	cb.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cb.Paint(buf)
}

func TestP411_ComboBox_Concurrent(t *testing.T) {
	cb := NewComboBox([]string{"a", "b"})
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { cb.SetQuery("a") }; close(done) }()
	for i := 0; i < 500; i++ { _ = cb.Filtered() }
	<-done
}

func TestP411_Dropdown_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Dropdown)(nil)
}

func TestP411_ComboBox_SatisfiesComponent(t *testing.T) {
	var _ Component = (*ComboBox)(nil)
}

func BenchmarkP411_ComboBox_Paint(b *testing.B) {
	cb := NewComboBox([]string{"gpt-4", "gpt-3.5", "claude-3", "gemini"})
	cb.SetQuery("g")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { cb.Paint(buf) }
}
