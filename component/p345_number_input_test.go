package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

func TestP345_NumberInput_Create(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	if n.Value() != 50 {
		t.Errorf("value = %d, want 50", n.Value())
	}
}

func TestP345_NumberInput_ClampOnCreate(t *testing.T) {
	n := NewNumberInput(200, 0, 100)
	if n.Value() != 100 {
		t.Errorf("value = %d, want 100 (clamped)", n.Value())
	}
	n2 := NewNumberInput(-10, 0, 100)
	if n2.Value() != 0 {
		t.Errorf("value = %d, want 0 (clamped)", n2.Value())
	}
}

func TestP345_NumberInput_IncrementDecrement(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.Increment()
	if n.Value() != 51 {
		t.Errorf("after increment: %d", n.Value())
	}
	n.Decrement()
	n.Decrement()
	if n.Value() != 49 {
		t.Errorf("after 2 decrements: %d", n.Value())
	}
}

func TestP345_NumberInput_StepClamp(t *testing.T) {
	n := NewNumberInput(95, 0, 100)
	n.SetStep(10)
	n.Increment() // 105 → clamped to 100
	if n.Value() != 100 {
		t.Errorf("clamped increment: %d, want 100", n.Value())
	}
	n.Decrement()
	if n.Value() != 90 {
		t.Errorf("decrement with step 10: %d, want 90", n.Value())
	}
}

func TestP345_NumberInput_SetValue(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.SetValue(75)
	if n.Value() != 75 {
		t.Errorf("value = %d, want 75", n.Value())
	}
	n.SetValue(200)
	if n.Value() != 100 {
		t.Errorf("clamped: %d, want 100", n.Value())
	}
	n.SetValue(-5)
	if n.Value() != 0 {
		t.Errorf("clamped: %d, want 0", n.Value())
	}
}

func TestP345_NumberInput_SetRange(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.SetRange(10, 20)
	if n.Value() != 20 {
		t.Errorf("value should clamp to new range: %d", n.Value())
	}
}

func TestP345_NumberInput_SetRange_InvalidMax(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.SetRange(50, 10) // max < min → max = min = 50
	if n.Value() != 50 {
		t.Errorf("value = %d", n.Value())
	}
}

func TestP345_NumberInput_SetStep(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.SetStep(0) // should clamp to 1
	n.Increment()
	if n.Value() != 51 {
		t.Errorf("step 0 should clamp to 1: %d", n.Value())
	}
}

func TestP345_NumberInput_HandleKeyUp(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	handled := n.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if !handled {
		t.Error("Up should be handled")
	}
	if n.Value() != 51 {
		t.Errorf("value after Up: %d", n.Value())
	}
}

func TestP345_NumberInput_HandleKeyDown(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if n.Value() != 49 {
		t.Errorf("value after Down: %d", n.Value())
	}
}

func TestP345_NumberInput_HandleKeyDigit(t *testing.T) {
	n := NewNumberInput(5, 0, 100)
	n.HandleKey(&term.KeyEvent{Rune: '3'}) // 53
	if n.Value() != 53 {
		t.Errorf("value after typing 3: %d", n.Value())
	}
}

func TestP345_NumberInput_HandleKeyDigit_Clamp(t *testing.T) {
	n := NewNumberInput(99, 0, 100)
	n.HandleKey(&term.KeyEvent{Rune: '9'}) // 999 → clamp to 100
	if n.Value() != 100 {
		t.Errorf("clamped digit input: %d", n.Value())
	}
}

func TestP345_NumberInput_HandleKey_Nil(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	if n.HandleKey(nil) {
		t.Error("nil event should return false")
	}
}

func TestP345_NumberInput_Focus(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	if n.IsFocused() {
		t.Error("should start unfocused")
	}
	n.SetFocused(true)
	if !n.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}
}

func TestP345_NumberInput_PrefixSuffix(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	n.SetPrefix("$")
	n.SetSuffix("tokens")
	n.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	n.Paint(buf) // should render with prefix/suffix
}

func TestP345_NumberInput_Paint(t *testing.T) {
	n := NewNumberInput(42, 0, 100)
	n.SetFocused(true)
	n.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)
	n.Paint(buf)

	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP345_NumberInput_Paint_NegativeValue(t *testing.T) {
	n := NewNumberInput(-5, -100, 100)
	n.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 1})
	buf := buffer.NewBuffer(20, 1)
	n.Paint(buf) // should render negative value
}

func TestP345_NumberInput_Measure(t *testing.T) {
	n := NewNumberInput(50, 0, 100)
	s := n.Measure(Constraints{MaxWidth: 40, MaxHeight: 1})
	if s.H != 1 {
		t.Errorf("height = %d, want 1", s.H)
	}
	if s.W < 4 {
		t.Errorf("width = %d, too small", s.W)
	}
}

func TestP345_IntToBuf(t *testing.T) {
	tests := []struct {
		v    int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{-5, "-5"},
		{999, "999"},
		{-100, "-100"},
	}
	for _, tt := range tests {
		got := string(intToBuf(nil, tt.v))
		if got != tt.want {
			t.Errorf("intToBuf(%d) = %q, want %q", tt.v, got, tt.want)
		}
	}
}

func TestP345_ClampInt(t *testing.T) {
	if clampIntRange(5, 0, 10) != 5 {
		t.Error("5 should stay 5")
	}
	if clampIntRange(-1, 0, 10) != 0 {
		t.Error("-1 should clamp to 0")
	}
	if clampIntRange(11, 0, 10) != 10 {
		t.Error("11 should clamp to 10")
	}
}

func BenchmarkNumberInput_Paint(b *testing.B) {
	n := NewNumberInput(42, 0, 100)
	n.SetSuffix("tokens")
	n.SetFocused(true)
	n.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 2})
	buf := buffer.NewBuffer(20, 2)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		n.Paint(buf)
	}
}
