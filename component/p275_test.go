package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestMaskedInput_Measure_ClampWidth_P275(t *testing.T) {
	mi := NewMaskedInput("###-###")
	s := mi.Measure(Constraints{MaxWidth: 3})
	if s.W > 3 {
		t.Errorf("width should be clamped to 3, got %d", s.W)
	}
}

func TestMaskedInput_HandleKey_PrintableMatch_P275(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.HandleKey(&term.KeyEvent{Rune: '5'})
}

func TestMaskedInput_HandleKey_NonMatch_P275(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.HandleKey(&term.KeyEvent{Rune: 'A'})
}

func TestMaskedInput_HandleKey_NilEvent_P275(t *testing.T) {
	mi := NewMaskedInput("###")
	mi.HandleKey(nil)
}

func TestMenuBar_FirstSelectable_P275(t *testing.T) {
	mb := NewMenuBar([]Menu{
		{ID: "file", Title: "File", Items: []MenuEntry{
			{Label: "New"},
			{Separator: true},
			{Label: "Open", Disabled: true},
			{Label: "Save"},
		}},
	})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	mb.Paint(buf)
}

func TestMenuBar_InvalidMenu_P275(t *testing.T) {
	mb := NewMenuBar([]Menu{})
	mb.SetBounds(Rect{X: 0, Y: 0, W: 80, H: 1})
	buf := buffer.NewBuffer(80, 1)
	mb.Paint(buf)
}

func TestNotification_DefaultDurationFor_P275(t *testing.T) {
	d := DefaultDurationFor(LevelInfo)
	if d <= 0 {
		t.Error("info duration should be positive")
	}
	d = DefaultDurationFor(NotificationLevel(99))
	if d <= 0 {
		t.Error("unknown level should still return positive duration")
	}
}

func TestNotification_Paint_MultiLevel_P275(t *testing.T) {
	tm := NewToastManager(5)
	tm.PushInfo("Info", "Info message")
	tm.PushSuccess("OK", "Success message")
	tm.PushWarning("Warn", "Warning message")
	tm.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	tm.Paint(buf)
}

func TestLink_ScanLine_EmptyText_P275(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("", 0, 0)
	if len(lm.Links()) != 0 {
		t.Error("empty text should produce no links")
	}
}

func TestLink_ScanLine_WithURL_P275(t *testing.T) {
	lm := NewLinkManager()
	lm.ScanLine("Check https://example.com out", 0, 0)
	links := lm.Links()
	if len(links) == 0 {
		t.Error("should detect URL in text")
	}
}

func TestLineChart_Paint_WithData_P275(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{
		Name: "data",
		Data: []ChartPoint{
			{X: 0, Y: 10}, {X: 1, Y: 20}, {X: 2, Y: 15},
			{X: 3, Y: 30}, {X: 4, Y: 25}, {X: 5, Y: 40},
		},
	})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	lc.Paint(buf)
}

func TestLineChart_Paint_MultiSeries_P275(t *testing.T) {
	lc := NewLineChart()
	lc.AddSeries(ChartSeries{Name: "A", Data: []ChartPoint{{X: 0, Y: 1}, {X: 1, Y: 5}, {X: 2, Y: 3}}})
	lc.AddSeries(ChartSeries{Name: "B", Data: []ChartPoint{{X: 0, Y: 2}, {X: 1, Y: 4}, {X: 2, Y: 6}}})
	lc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 15})
	buf := buffer.NewBuffer(60, 15)
	lc.Paint(buf)
}
