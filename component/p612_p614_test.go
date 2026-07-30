package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── HeatCell Tests ───

func TestHeatCellBasic(t *testing.T) {
	hc := NewHeatCell()
	hc.SetValue(75)
	if v := hc.Value(); v != 75 {
		t.Errorf("Value = %d, want 75", v)
	}
}

func TestHeatCellZero(t *testing.T) {
	hc := NewHeatCell()
	if v := hc.Value(); v != 0 {
		t.Errorf("Value = %d, want 0", v)
	}
}

func TestHeatCellClamp(t *testing.T) {
	hc := NewHeatCell()
	hc.SetValue(-10)
	if v := hc.Value(); v != 0 {
		t.Errorf("Value = %d, want 0 (clamped)", v)
	}
	hc.SetValue(200)
	if v := hc.Value(); v != 100 {
		t.Errorf("Value = %d, want 100 (clamped)", v)
	}
}

func TestHeatCellColorZones(t *testing.T) {
	hc := NewHeatCell()
	hc.SetValue(10)
	if hc.curStyle.Fg != hc.style.Cold.Fg {
		t.Error("Expected Cold for level <=2")
	}
	hc.SetValue(45)
	if hc.curStyle.Fg != hc.style.Cool.Fg {
		t.Error("Expected Cool for level 3-4")
	}
	hc.SetValue(65)
	if hc.curStyle.Fg != hc.style.Warm.Fg {
		t.Error("Expected Warm for level 5-6")
	}
	hc.SetValue(90)
	if hc.curStyle.Fg != hc.style.Hot.Fg {
		t.Error("Expected Hot for level >=7")
	}
}

func TestHeatCellPaint(t *testing.T) {
	hc := NewHeatCell()
	hc.SetValue(50)
	hc.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 1})
	buf := buffer.NewBuffer(4, 1)
	hc.Paint(buf)
	if r := buf.GetCell(0, 0).Rune; r == 0 || r == ' ' {
		t.Error("Paint should show heat char")
	}
}

func TestHeatCellChildren(t *testing.T) {
	hc := NewHeatCell()
	if c := hc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestHeatCellStyle(t *testing.T) {
	hc := NewHeatCell()
	hc.SetStyle(HeatCellStyle{
		Cold:  buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Cool:  buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Warm:  buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Hot:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Empty: buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Label: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	hc.SetValue(85)
	buf := buffer.NewBuffer(4, 1)
	hc.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 1})
	hc.Paint(buf)
}

// ─── AIFactCheck Tests ───

func TestAIFactCheckBasic(t *testing.T) {
	fc := NewAIFactCheck()
	fc.SetResult(FactVerified, 3)
	if r := fc.Result(); r != FactVerified {
		t.Errorf("Result = %d, want FactVerified", r)
	}
}

func TestAIFactCheckAllResults(t *testing.T) {
	results := []FactCheckResult{FactUnverified, FactVerified, FactDisputed, FactPartial}
	for _, r := range results {
		fc := NewAIFactCheck()
		fc.SetResult(r, 1)
		if fc.Result() != r {
			t.Errorf("Result = %d, want %d", fc.Result(), r)
		}
	}
}

func TestAIFactCheckInvalid(t *testing.T) {
	fc := NewAIFactCheck()
	fc.SetResult(FactCheckResult(99), 1)
	if r := fc.Result(); r != FactUnverified {
		t.Errorf("Result = %d, want FactUnverified (clamped)", r)
	}
}

func TestAIFactCheckNegativeSources(t *testing.T) {
	fc := NewAIFactCheck()
	fc.SetResult(FactVerified, -5)
	if fc.sources != 0 {
		t.Errorf("sources = %d, want 0 (clamped)", fc.sources)
	}
}

func TestAIFactCheckPaint(t *testing.T) {
	fc := NewAIFactCheck()
	fc.SetResult(FactVerified, 2)
	fc.SetBounds(Rect{X: 0, Y: 0, W: 18, H: 1})
	buf := buffer.NewBuffer(18, 1)
	fc.Paint(buf)
	hasContent := false
	for i := 0; i < 18; i++ {
		if buf.GetCell(i, 0).Rune != ' ' && buf.GetCell(i, 0).Rune != 0 {
			hasContent = true
			break
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestAIFactCheckChildren(t *testing.T) {
	fc := NewAIFactCheck()
	if c := fc.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIFactCheckStyle(t *testing.T) {
	fc := NewAIFactCheck()
	fc.SetStyle(AIFactCheckStyle{
		Unverified: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Verified:   buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Disputed:   buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Partial:    buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		Source:     buffer.Style{Fg: buffer.RGB(0, 0, 255)},
		Bracket:    buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	fc.SetResult(FactDisputed, 1)
	buf := buffer.NewBuffer(18, 1)
	fc.SetBounds(Rect{X: 0, Y: 0, W: 18, H: 1})
	fc.Paint(buf)
}

// ─── ScoreDonut Tests ───

func TestScoreDonutBasic(t *testing.T) {
	sd := NewScoreDonut()
	sd.SetValue(72)
	if v := sd.Value(); v != 72 {
		t.Errorf("Value = %d, want 72", v)
	}
}

func TestScoreDonutZero(t *testing.T) {
	sd := NewScoreDonut()
	if v := sd.Value(); v != 0 {
		t.Errorf("Value = %d, want 0", v)
	}
}

func TestScoreDonutFull(t *testing.T) {
	sd := NewScoreDonut()
	sd.SetValue(100)
	if v := sd.Value(); v != 100 {
		t.Errorf("Value = %d, want 100", v)
	}
}

func TestScoreDonutClamp(t *testing.T) {
	sd := NewScoreDonut()
	sd.SetValue(-10)
	if v := sd.Value(); v != 0 {
		t.Errorf("Value = %d, want 0 (clamped)", v)
	}
	sd.SetValue(200)
	if v := sd.Value(); v != 100 {
		t.Errorf("Value = %d, want 100 (clamped)", v)
	}
}

func TestScoreDonutPaint(t *testing.T) {
	sd := NewScoreDonut()
	sd.SetValue(75)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	buf := buffer.NewBuffer(8, 3)
	sd.Paint(buf)
	// Should have ring chars or % sign
	hasContent := false
	for row := 0; row < 3; row++ {
		for col := 0; col < 8; col++ {
			r := buf.GetCell(col, row).Rune
			if r != ' ' && r != 0 {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Error("Paint produced no content")
	}
}

func TestScoreDonutChildren(t *testing.T) {
	sd := NewScoreDonut()
	if c := sd.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestScoreDonutStyle(t *testing.T) {
	sd := NewScoreDonut()
	sd.SetStyle(ScoreDonutStyle{
		Filled: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Empty:  buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Center: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Label:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	sd.SetValue(50)
	buf := buffer.NewBuffer(8, 3)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	sd.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintHeatCell(b *testing.B) {
	hc := NewHeatCell()
	hc.SetValue(75)
	hc.SetBounds(Rect{X: 0, Y: 0, W: 4, H: 1})
	buf := buffer.NewBuffer(4, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hc.Paint(buf)
	}
}

func BenchmarkPaintAIFactCheck(b *testing.B) {
	fc := NewAIFactCheck()
	fc.SetResult(FactVerified, 3)
	fc.SetBounds(Rect{X: 0, Y: 0, W: 18, H: 1})
	buf := buffer.NewBuffer(18, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fc.Paint(buf)
	}
}

func BenchmarkPaintScoreDonut(b *testing.B) {
	sd := NewScoreDonut()
	sd.SetValue(72)
	sd.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 3})
	buf := buffer.NewBuffer(8, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd.Paint(buf)
	}
}
