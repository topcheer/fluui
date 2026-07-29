package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestKeyHintBarBasic(t *testing.T) {
	kb := NewKeyHintBar()
	kb.AddHint("Q", "Quit")
	kb.AddHint("S", "Save")
	if kb.HintCount() != 2 { t.Errorf("HintCount = %d, want 2", kb.HintCount()) }
}

func TestKeyHintBarSetHints(t *testing.T) {
	kb := NewKeyHintBar()
	kb.SetHints([]KeyHint{{Key: "A", Description: "a"}, {Key: "B", Description: "b"}})
	if kb.HintCount() != 2 { t.Errorf("HintCount = %d, want 2", kb.HintCount()) }
}

func TestKeyHintBarClear(t *testing.T) {
	kb := NewKeyHintBar()
	kb.AddHint("Q", "Quit")
	kb.Clear()
	if kb.HintCount() != 0 { t.Errorf("HintCount = %d, want 0", kb.HintCount()) }
}

func TestKeyHintBarEmpty(t *testing.T) {
	kb := NewKeyHintBar()
	if kb.HintCount() != 0 { t.Errorf("HintCount = %d, want 0", kb.HintCount()) }
}

func TestKeyHintBarMeasure(t *testing.T) {
	kb := NewKeyHintBar()
	kb.AddHint("Q", "Quit")
	s := kb.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H != 1 { t.Errorf("H = %d, want 1", s.H) }
}

func TestKeyHintBarPaint(t *testing.T) {
	kb := NewKeyHintBar()
	kb.AddHint("Q", "Quit")
	kb.AddHint("S", "Save")
	kb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	kb.Paint(buf)
	// Check bracket exists
	foundBracket := false
	for x := 0; x < 40; x++ {
		if buf.GetCell(x, 0).Rune == '[' { foundBracket = true; break }
	}
	if !foundBracket { t.Error("key bracket not found") }
}

func TestKeyHintBarPaintEmpty(t *testing.T) {
	kb := NewKeyHintBar()
	kb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	kb.Paint(buf)
}

func TestKeyHintBarChildren(t *testing.T) {
	kb := NewKeyHintBar()
	if kb.Children() != nil { t.Error("Children should be nil") }
}

func TestKeyHintBarStyle(t *testing.T) {
	kb := NewKeyHintBar()
	kb.SetStyle(KeyHintBarStyle{
		Key: buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold},
		KeyBg: buffer.Style{Fg: buffer.RGB(50, 50, 50), Bg: buffer.RGB(30, 30, 30)},
		Desc: buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Separator: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	kb.AddHint("X", "Test")
	kb.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	kb.Paint(buf)
}

// ─── DataLabel tests ───

func TestDataLabelBasic(t *testing.T) {
	dl := NewDataLabel()
	dl.SetLabel("Revenue")
	dl.SetValue(42.5)
	dl.SetUnit("K")
	dl.SetTrend(DataTrendUp)
	if dl.Label() != "Revenue" { t.Errorf("Label = %q", dl.Label()) }
	if dl.Value() != 42.5 { t.Errorf("Value = %f", dl.Value()) }
	if dl.Unit() != "K" { t.Errorf("Unit = %q", dl.Unit()) }
	if dl.Trend() != DataTrendUp { t.Errorf("Trend = %d", dl.Trend()) }
}

func TestDataLabelTrends(t *testing.T) {
	dl := NewDataLabel()
	dl.SetTrend(DataTrendDown)
	if dl.Trend() != DataTrendDown { t.Error("should be down") }
	dl.SetTrend(DataTrendFlat)
	if dl.Trend() != DataTrendFlat { t.Error("should be flat") }
}

func TestDataLabelMeasure(t *testing.T) {
	dl := NewDataLabel()
	s := dl.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestDataLabelPaint(t *testing.T) {
	dl := NewDataLabel()
	dl.SetLabel("Users")
	dl.SetValue(1234.0)
	dl.SetUnit(" total")
	dl.SetTrend(DataTrendUp)
	dl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 4})
	buf := buffer.NewBuffer(20, 4)
	dl.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	// Check up arrow exists
	foundArrow := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 2).Rune == '↑' { foundArrow = true; break }
	}
	if !foundArrow { t.Error("up trend arrow not found") }
}

func TestDataLabelPaintDown(t *testing.T) {
	dl := NewDataLabel()
	dl.SetLabel("Errors")
	dl.SetValue(5.0)
	dl.SetTrend(DataTrendDown)
	dl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 4})
	buf := buffer.NewBuffer(20, 4)
	dl.Paint(buf)
	foundDown := false
	for x := 0; x < 20; x++ {
		if buf.GetCell(x, 2).Rune == '↓' { foundDown = true; break }
	}
	if !foundDown { t.Error("down trend arrow not found") }
}

func TestDataLabelChildren(t *testing.T) {
	dl := NewDataLabel()
	if dl.Children() != nil { t.Error("Children should be nil") }
}

func TestDataLabelStyle(t *testing.T) {
	dl := NewDataLabel()
	dl.SetStyle(DataLabelStyle{
		Label: buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Value: buffer.Style{Fg: buffer.RGB(255, 255, 255), Flags: buffer.Bold},
		Unit: buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Up: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Down: buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Flat: buffer.Style{Fg: buffer.RGB(150, 150, 150)},
		Border: buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	dl.SetLabel("Test")
	dl.SetValue(1.0)
	dl.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 4})
	buf := buffer.NewBuffer(20, 4)
	dl.Paint(buf)
}
