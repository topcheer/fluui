package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P413: Coverage for Dropdown, ComboBox, Popup, InfoCard zero-coverage methods

// === Dropdown uncovered getters ===

func TestP413_Dropdown_MaxHeight(t *testing.T) {
	d := NewDropdown([]DropdownItem{{Label: "A"}})
	if d.MaxHeight() != 8 { t.Errorf("MaxHeight = %d, want 8", d.MaxHeight()) }
	d.SetMaxHeight(5)
	if d.MaxHeight() != 5 { t.Errorf("MaxHeight = %d", d.MaxHeight()) }
	d.SetMaxHeight(0)
	if d.MaxHeight() != 1 { t.Errorf("MaxHeight = %d, want 1", d.MaxHeight()) }
}

func TestP413_Dropdown_SelectedItem_Empty(t *testing.T) {
	d := NewDropdown(nil)
	item := d.SelectedItem()
	if item.Label != "" { t.Error("should be empty") }
}

// === ComboBox uncovered methods ===

func TestP413_ComboBox_SetItems(t *testing.T) {
	cb := NewComboBox([]string{"old"})
	cb.SetItems([]string{"new1", "new2"})
	if len(cb.Items()) != 2 { t.Errorf("Items = %d", len(cb.Items())) }
}

func TestP413_ComboBox_SelectedItem(t *testing.T) {
	cb := NewComboBox([]string{"a", "b"})
	cb.MoveDown()
	if cb.SelectedItem() != "b" { t.Errorf("SelectedItem = %q", cb.SelectedItem()) }
	// Out of range
	cb.MoveDown(); cb.MoveDown(); cb.MoveDown()
	_ = cb.SelectedItem() // should not panic
}

func TestP413_ComboBox_SetExpanded(t *testing.T) {
	cb := NewComboBox([]string{"a"})
	cb.SetExpanded(true)
	if !cb.Expanded() { t.Error("should be expanded") }
	cb.Collapse()
	if cb.Expanded() { t.Error("should be collapsed") }
}

func TestP413_ComboBox_Selected(t *testing.T) {
	cb := NewComboBox([]string{"a", "b"})
	if cb.Selected() != 0 { t.Errorf("Selected = %d", cb.Selected()) }
}

func TestP413_ComboBox_Measure(t *testing.T) {
	cb := NewComboBox([]string{"short", "very long item"})
	s := cb.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d", s.H) }
}

func TestP413_ComboBox_Measure_Expanded(t *testing.T) {
	cb := NewComboBox([]string{"a", "b", "c"})
	cb.SetExpanded(true)
	s := cb.Measure(Constraints{MaxWidth: 80, MaxHeight: 20})
	if s.H != 4 { t.Errorf("H = %d, want 4 (input + 3 items)", s.H) }
}

func TestP413_ComboBox_Paint_EmptyQueryCollapsed(t *testing.T) {
	cb := NewComboBox([]string{"a"})
	cb.Collapse()
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	cb.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != 't' { t.Errorf("cell = %q, want 't'", string(c.Rune)) }
}

func TestP413_ComboBox_Paint_NoMatchExpanded(t *testing.T) {
	cb := NewComboBox([]string{"a"})
	cb.SetExpanded(true)
	cb.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	cb.Paint(buf) // should not panic with empty filtered
}

// === Popup uncovered methods ===

func TestP413_Popup_Measure(t *testing.T) {
	p := NewPopup(NewText("x"))
	s := p.Measure(Constraints{MaxWidth: 30, MaxHeight: 10})
	if s.W != 30 { t.Errorf("W = %d", s.W) }
	if s.H != 10 { t.Errorf("H = %d", s.H) }
}

func TestP413_Popup_Measure_Defaults(t *testing.T) {
	p := NewPopup(NewText("x"))
	s := p.Measure(Constraints{}) // no constraints → defaults
	if s.W != 20 { t.Errorf("W = %d, want 20", s.W) }
	if s.H != 5 { t.Errorf("H = %d, want 5", s.H) }
}

func TestP413_Popup_Paint_NoContent(t *testing.T) {
	p := NewPopup(nil)
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	p.Paint(buf) // should not panic
}

func TestP413_Popup_Paint_TooSmall(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1}) // too small for border+content
	buf := buffer.NewBuffer(10, 1)
	p.Paint(buf)
}

// === InfoCard uncovered methods ===

func TestP413_InfoCard_SetFields(t *testing.T) {
	c := NewInfoCard("", "", "")
	c.SetIcon("⚠")
	c.SetTitle("Warn")
	c.SetBody("text")
	c.SetVariant(InfoCardError)
	if c.Icon() != "⚠" { t.Error("icon") }
	if c.Title() != "Warn" { t.Error("title") }
	if c.Body() != "text" { t.Error("body") }
	if c.Variant() != InfoCardError { t.Error("variant") }
}

func TestP413_InfoCard_Measure_NoBody(t *testing.T) {
	c := NewInfoCard("●", "Title", "")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H != 1 { t.Errorf("H = %d, want 1 (no body)", s.H) }
}

func TestP413_InfoCard_Measure_NoIcon(t *testing.T) {
	c := NewInfoCard("", "Title", "Body")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.W < 4 { t.Errorf("W = %d", s.W) }
}

func TestP413_InfoCard_Paint_NoIcon(t *testing.T) {
	c := NewInfoCard("", "Title", "Body text")
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)
	c.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'T' { t.Errorf("cell = %q, want 'T'", string(cell.Rune)) }
}

func TestP413_InfoCard_Paint_LongBody(t *testing.T) {
	c := NewInfoCard("●", "Title", "This is a very long body text that exceeds width")
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 2})
	buf := buffer.NewBuffer(10, 2)
	c.Paint(buf) // should truncate body
}

func TestP413_InfoCard_Paint_NonZeroOffset(t *testing.T) {
	c := NewInfoCard("●", "Title", "Body")
	c.SetBounds(Rect{X: 5, Y: 3, W: 20, H: 2})
	buf := buffer.NewBuffer(30, 10)
	c.Paint(buf)
	cell := buf.GetCell(5, 3)
	if cell.Rune != '●' { t.Errorf("offset cell = %q", string(cell.Rune)) }
}

// === CircularProgress uncovered ===

func TestP413_CircularProgress_SetBarWidth(t *testing.T) {
	c := NewCircularProgress(0.5)
	c.SetBarWidth(10)
	if c.BarWidth() != 10 { t.Errorf("BarWidth = %d", c.BarWidth()) }
	c.SetBarWidth(0)
	if c.BarWidth() != 1 { t.Errorf("BarWidth = %d, want 1", c.BarWidth()) }
}
