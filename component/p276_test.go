package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestColorPicker_Paint_RGB_P276(t *testing.T) {
	cp := NewColorPicker()
	cp.SetColor(buffer.RGB(128, 64, 200))
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	cp.Paint(buf)
}

func TestColorPicker_Paint_RGB_Narrow_P276(t *testing.T) {
	cp := NewColorPicker()
	cp.SetColor(buffer.RGB(255, 0, 0))
	cp.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 8})
	buf := buffer.NewBuffer(15, 8)
	cp.Paint(buf)
}

func TestColorPicker_Paint_RGB_ActiveChannel_P276(t *testing.T) {
	cp := NewColorPicker()
	cp.SetColor(buffer.RGB(100, 150, 200))
	cp.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	buf := buffer.NewBuffer(40, 8)
	cp.Paint(buf)
}

func TestPages_Measure_NilPage_P276(t *testing.T) {
	p := NewPages()
	s := p.Measure(Constraints{MaxWidth: 100, MaxHeight: 50})
	if s.W != 0 || s.H != 0 {
		t.Error("page with no content should return 0x0")
	}
}

func TestPages_Measure_WithContent_P276(t *testing.T) {
	p := NewPages()
	child := &fixedSize{w: 40, h: 20}
	p.AddPage("page1", child)
	p.SwitchTo("page1")
	s := p.Measure(Constraints{MaxWidth: 100, MaxHeight: 50})
	if s.W != 40 || s.H != 20 {
		t.Errorf("expected 40x20, got %dx%d", s.W, s.H)
	}
}

func TestWizard_ButtonLabel_AllButtons_P276(t *testing.T) {
	buttons := []WizardButton{WizardBtnBack, WizardBtnNext, WizardBtnFinish, WizardBtnCancel, WizardButton(99)}
	for _, b := range buttons {
		label := b.ButtonLabel()
		if label == "" {
			t.Errorf("button %d should have non-empty label", b)
		}
	}
	if WizardButton(99).ButtonLabel() != "?" {
		t.Error("unknown button should return '?'")
	}
}

func TestTable_EnsureVisible_ScrollDown_P276(t *testing.T) {
	rows := make([][]string, 20)
	for i := range rows {
		rows[i] = []string{"val1", "val2"}
	}
	tbl := NewTable([]string{"A", "B"}, rows...)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	tbl.SetSelectedRow(15)
	buf := buffer.NewBuffer(40, 5)
	tbl.Paint(buf)
}

func TestTable_EnsureVisible_ScrollUp_P276(t *testing.T) {
	rows := make([][]string, 20)
	for i := range rows {
		rows[i] = []string{"x", "y"}
	}
	tbl := NewTable([]string{"A", "B"}, rows...)
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	tbl.SetSelectedRow(15)
	tbl.SetSelectedRow(2)
	buf := buffer.NewBuffer(40, 5)
	tbl.Paint(buf)
}

func TestTable_clampScrollY_P276(t *testing.T) {
	tbl := NewTable([]string{"A"}, []string{"1"}, []string{"2"}, []string{"3"})
	tbl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	tbl.Paint(buf)
}

func TestMarkdownViewer_PaintToc_WithHeaders_P276(t *testing.T) {
	mv := NewMarkdownViewer("# Title\n\n## Sub A\n\ntext\n\n## Sub B\n\nmore")
	mv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	mv.Paint(buf)
}

func TestMarkdownViewer_PaintToc_DeepHeaders_P276(t *testing.T) {
	mv := NewMarkdownViewer("# H1\n\n## H2\n\n### H3\n\ntext\n\n## H2b\n\n### H3b")
	mv.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	mv.Paint(buf)
}
