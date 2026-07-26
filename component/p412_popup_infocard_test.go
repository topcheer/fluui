package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// === Popup tests ===

func TestP412_Popup_New(t *testing.T) {
	p := NewPopup(NewText("content"))
	if !p.Visible() { t.Error("should be visible by default") }
	if p.ID() == "" { t.Error("ID empty") }
}

func TestP412_Popup_ShowHide(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.Hide()
	if p.Visible() { t.Error("should be hidden") }
	p.Show()
	if !p.Visible() { t.Error("should be visible") }
	p.SetVisible(false)
	if p.Visible() { t.Error("should be hidden") }
}

func TestP412_Popup_SetTitle(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.SetTitle("Options")
	if p.Title() != "Options" { t.Errorf("Title = %q", p.Title()) }
}

func TestP412_Popup_SetContent(t *testing.T) {
	p := NewPopup(nil)
	p.SetContent(NewText("new"))
	if p.Content() == nil { t.Error("Content should not be nil") }
}

func TestP412_Popup_SetShadow(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.SetShadow(true)
	if !p.Shadow() { t.Error("should have shadow") }
}

func TestP412_Popup_Paint(t *testing.T) {
	p := NewPopup(NewText("Hello"))
	p.SetTitle("Test")
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	p.Paint(buf)
	c := buf.GetCell(0, 0)
	if c.Rune != '\u256d' { t.Errorf("cell = %q, want ╭", string(c.Rune)) } // top-left corner
}

func TestP412_Popup_Paint_Hidden(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.Hide()
	p.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	buf := buffer.NewBuffer(10, 5)
	p.Paint(buf)
	// Nothing should be drawn — check no border corner
	if buf.GetCell(0, 0).Rune == '\u256d' { t.Error("hidden popup should not draw border") }
}

func TestP412_Popup_Paint_Shadow(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.SetShadow(true)
	p.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(10, 5)
	p.Paint(buf)
	// Shadow at bottom-right
	if buf.GetCell(1, 3).Rune != '\u2592' { t.Error("should have shadow char") }
}

func TestP412_Popup_Paint_ZeroBounds(t *testing.T) {
	p := NewPopup(NewText("x"))
	p.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	p.Paint(buf)
}

func TestP412_Popup_Concurrent(t *testing.T) {
	p := NewPopup(NewText("x"))
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { p.Show() }; close(done) }()
	for i := 0; i < 500; i++ { _ = p.Visible() }
	<-done
}

func BenchmarkP412_Popup_Paint(b *testing.B) {
	p := NewPopup(NewText("Hello World"))
	p.SetTitle("Info")
	p.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	buf := buffer.NewBuffer(20, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { p.Paint(buf) }
}

// === InfoCard tests ===

func TestP412_InfoCard_New(t *testing.T) {
	c := NewInfoCard("ℹ", "Status", "All systems operational")
	if c.Icon() != "ℹ" { t.Errorf("Icon = %q", c.Icon()) }
	if c.Title() != "Status" { t.Errorf("Title = %q", c.Title()) }
	if c.Body() != "All systems operational" { t.Errorf("Body = %q", c.Body()) }
	if c.Variant() != InfoCardDefault { t.Errorf("Variant = %v", c.Variant()) }
	if c.ID() == "" { t.Error("ID empty") }
}

func TestP412_InfoCard_SetFields(t *testing.T) {
	c := NewInfoCard("", "", "")
	c.SetIcon("⚠")
	c.SetTitle("Warning")
	c.SetBody("Disk almost full")
	c.SetVariant(InfoCardWarning)
	if c.Icon() != "⚠" { t.Error("Icon") }
	if c.Title() != "Warning" { t.Error("Title") }
	if c.Body() != "Disk almost full" { t.Error("Body") }
	if c.Variant() != InfoCardWarning { t.Error("Variant") }
}

func TestP412_InfoCard_Measure(t *testing.T) {
	c := NewInfoCard("✓", "OK", "All good")
	s := c.Measure(Constraints{MaxWidth: 80, MaxHeight: 10})
	if s.H < 2 { t.Errorf("H = %d, want >= 2", s.H) }
}

func TestP412_InfoCard_Paint(t *testing.T) {
	c := NewInfoCard("✓", "Success", "Operation completed")
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	c.Paint(buf)
	// Icon at start
	cell := buf.GetCell(0, 0)
	if cell.Rune != '✓' { t.Errorf("icon cell = %q", string(cell.Rune)) }
}

func TestP412_InfoCard_Paint_Variants(t *testing.T) {
	for _, v := range []InfoCardVariant{InfoCardDefault, InfoCardSuccess, InfoCardWarning, InfoCardError, InfoCardAccent} {
		c := NewInfoCard("●", "Test", "Body")
		c.SetVariant(v)
		c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
		buf := buffer.NewBuffer(20, 3)
		c.Paint(buf)
	}
}

func TestP412_InfoCard_Paint_NoBody(t *testing.T) {
	c := NewInfoCard("●", "Title only", "")
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	c.Paint(buf)
}

func TestP412_InfoCard_Paint_ZeroBounds(t *testing.T) {
	c := NewInfoCard("●", "T", "B")
	c.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	c.Paint(buf)
}

func TestP412_InfoCard_Concurrent(t *testing.T) {
	c := NewInfoCard("●", "T", "B")
	done := make(chan struct{})
	go func() { for i := 0; i < 500; i++ { c.SetTitle("concurrent") }; close(done) }()
	for i := 0; i < 500; i++ { _ = c.Title() }
	<-done
}

func TestP412_Popup_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Popup)(nil)
}

func TestP412_InfoCard_SatisfiesComponent(t *testing.T) {
	var _ Component = (*InfoCard)(nil)
}

func BenchmarkP412_InfoCard_Paint(b *testing.B) {
	c := NewInfoCard("✓", "Status", "All systems operational and running smoothly")
	c.SetVariant(InfoCardSuccess)
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 2})
	buf := buffer.NewBuffer(40, 2)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ { c.Paint(buf) }
}
