package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestPasswordStrengthBasic(t *testing.T) {
	ps := NewPasswordStrength()
	ps.SetPassword("MyStr0ng!Pass")
	if ps.Level() == PWVeryWeak { t.Error("strong password should not be very weak") }
}

func TestPasswordStrengthEmpty(t *testing.T) {
	ps := NewPasswordStrength()
	ps.SetPassword("")
	if ps.Level() != PWVeryWeak { t.Errorf("Level = %d, want PWVeryWeak", ps.Level()) }
}

func TestPasswordStrengthWeak(t *testing.T) {
	ps := NewPasswordStrength()
	ps.SetPassword("abc")
	if ps.Level() != PWVeryWeak { t.Errorf("Level = %d, want PWVeryWeak for 'abc'", ps.Level()) }
}

func TestPasswordStrengthLevels(t *testing.T) {
	ps := NewPasswordStrength()
	ps.SetPassword("Abcdef1!")
	lvl1 := ps.Level()
	ps.SetPassword("Abcdefgh1!xyz")
	lvl2 := ps.Level()
	if lvl2 <= lvl1 { t.Error("stronger password should have higher level") }
}

func TestPasswordStrengthLevelText(t *testing.T) {
	if pwLevelText(PWVeryWeak) != "Very Weak" { t.Error("text mismatch") }
	if pwLevelText(PWStrong) != "Strong" { t.Error("text mismatch") }
}

func TestPasswordStrengthMeasure(t *testing.T) {
	ps := NewPasswordStrength()
	s := ps.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestPasswordStrengthPaint(t *testing.T) {
	ps := NewPasswordStrength()
	ps.SetPassword("MyStr0ng!Pass")
	ps.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	ps.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundBar := false
	for x := 0; x < 30; x++ {
		if buf.GetCell(x, 1).Rune == '█' { foundBar = true; break }
	}
	if !foundBar { t.Error("filled bar not found") }
}

func TestPasswordStrengthChildren(t *testing.T) {
	ps := NewPasswordStrength()
	if ps.Children() != nil { t.Error("Children should be nil") }
}

func TestPasswordStrengthStyle(t *testing.T) {
	ps := NewPasswordStrength()
	ps.SetStyle(PasswordStrengthStyle{Bar: [5]buffer.Style{{Fg: buffer.RGB(255,0,0)}, {Fg: buffer.RGB(255,165,0)}, {Fg: buffer.RGB(255,255,0)}, {Fg: buffer.RGB(0,255,0)}, {Fg: buffer.RGB(0,200,0)}}, Label: buffer.Style{Fg: buffer.RGB(150,150,150)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	ps.SetPassword("Test123!")
	ps.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	ps.Paint(buf)
}

// ─── AITokenFlow tests ───

func TestAITokenFlowBasic(t *testing.T) {
	tf := NewAITokenFlow()
	tf.AddStage("Input", 500, buffer.RGB(96, 165, 250))
	if tf.StageCount() != 1 { t.Errorf("StageCount = %d, want 1", tf.StageCount()) }
}

func TestAITokenFlowMultiple(t *testing.T) {
	tf := NewAITokenFlow()
	tf.AddStage("Input", 500, buffer.RGB(96, 165, 250))
	tf.AddStage("Embedding", 480, buffer.RGB(167, 139, 250))
	tf.AddStage("Output", 200, buffer.RGB(34, 197, 94))
	if tf.StageCount() != 3 { t.Errorf("StageCount = %d, want 3", tf.StageCount()) }
}

func TestAITokenFlowClear(t *testing.T) {
	tf := NewAITokenFlow()
	tf.AddStage("A", 100, buffer.RGB(0, 0, 0))
	tf.Clear()
	if tf.StageCount() != 0 { t.Errorf("StageCount = %d, want 0", tf.StageCount()) }
}

func TestAITokenFlowEmpty(t *testing.T) {
	tf := NewAITokenFlow()
	if tf.StageCount() != 0 { t.Errorf("StageCount = %d, want 0", tf.StageCount()) }
}

func TestAITokenFlowBarWidth(t *testing.T) {
	tf := NewAITokenFlow()
	tf.SetMaxBarWidth(10)
	tf.AddStage("A", 100, buffer.RGB(0, 0, 0))
	tf.AddStage("B", 50, buffer.RGB(0, 0, 0))
	tf.mu.Lock()
	if tf.stages[0].BarWidth != 10 { t.Errorf("stage 0 width = %d, want 10", tf.stages[0].BarWidth) }
	if tf.stages[1].BarWidth != 5 { t.Errorf("stage 1 width = %d, want 5", tf.stages[1].BarWidth) }
	tf.mu.Unlock()
}

func TestAITokenFlowMeasure(t *testing.T) {
	tf := NewAITokenFlow()
	tf.AddStage("X", 100, buffer.RGB(0, 0, 0))
	s := tf.Measure(Constraints{})
	if s.W < 10 { t.Errorf("W = %d", s.W) }
	if s.H < 3 { t.Errorf("H = %d", s.H) }
}

func TestAITokenFlowPaint(t *testing.T) {
	tf := NewAITokenFlow()
	tf.AddStage("Input", 500, buffer.RGB(96, 165, 250))
	tf.AddStage("Output", 200, buffer.RGB(34, 197, 94))
	tf.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 8})
	buf := buffer.NewBuffer(50, 8)
	tf.Paint(buf)
	if buf.GetCell(0, 0).Rune != '┌' { t.Error("border missing") }
	foundBar := false
	for x := 0; x < 50; x++ {
		if buf.GetCell(x, 1).Rune == '█' { foundBar = true; break }
	}
	if !foundBar { t.Error("bar not found") }
}

func TestAITokenFlowPaintEmpty(t *testing.T) {
	tf := NewAITokenFlow()
	tf.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 5})
	buf := buffer.NewBuffer(50, 5)
	tf.Paint(buf)
}

func TestAITokenFlowChildren(t *testing.T) {
	tf := NewAITokenFlow()
	if tf.Children() != nil { t.Error("Children should be nil") }
}

func TestAITokenFlowStyle(t *testing.T) {
	tf := NewAITokenFlow()
	tf.SetStyle(AITokenFlowStyle{Connector: buffer.Style{Fg: buffer.RGB(100,100,100)}, Label: buffer.Style{Fg: buffer.RGB(200,200,200)}, Border: buffer.Style{Fg: buffer.RGB(64,64,64)}})
	tf.AddStage("X", 100, buffer.RGB(255, 0, 0))
	tf.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	tf.Paint(buf)
}
