package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ─── Construction & defaults ───

func TestColorPicker_New(t *testing.T) {
	cp := NewColorPicker()
	if cp.Mode() != PickerPalette {
		t.Errorf("default mode = %v, want PickerPalette", cp.Mode())
	}
	c := cp.Color()
	if c.Type != buffer.ColorTrue || c.Val != 0 {
		t.Errorf("default color = %v, want RGB(0,0,0)", c)
	}
}

func TestColorPicker_SetMode(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	if cp.Mode() != PickerRGB {
		t.Errorf("mode = %v, want PickerRGB", cp.Mode())
	}
	cp.SetMode(PickerHex)
	if cp.Mode() != PickerHex {
		t.Errorf("mode = %v, want PickerHex", cp.Mode())
	}
}

func TestColorPicker_NextMode(t *testing.T) {
	cp := NewColorPicker()
	cp.NextMode() // Palette → RGB
	if cp.Mode() != PickerRGB {
		t.Errorf("after NextMode: %v, want PickerRGB", cp.Mode())
	}
	cp.NextMode() // RGB → Hex
	if cp.Mode() != PickerHex {
		t.Errorf("after NextMode: %v, want PickerHex", cp.Mode())
	}
	cp.NextMode() // Hex → Palette (wraps)
	if cp.Mode() != PickerPalette {
		t.Errorf("after NextMode wrap: %v, want PickerPalette", cp.Mode())
	}
}

func TestColorPicker_PrevMode(t *testing.T) {
	cp := NewColorPicker()
	cp.PrevMode() // Palette → Hex (wraps backward)
	if cp.Mode() != PickerHex {
		t.Errorf("after PrevMode wrap: %v, want PickerHex", cp.Mode())
	}
}

// ─── Color setting ───

func TestColorPicker_SetColor_TrueColor(t *testing.T) {
	cp := NewColorPicker()
	cp.SetColor(buffer.RGB(128, 64, 200))
	c := cp.Color()
	if c.R() != 128 || c.G() != 64 || c.B() != 200 {
		t.Errorf("SetColor RGB mismatch: R=%d G=%d B=%d", c.R(), c.G(), c.B())
	}
	r, g, b := cp.RGBValues()
	if r != 128 || g != 64 || b != 200 {
		t.Errorf("RGBValues mismatch: r=%d g=%d b=%d", r, g, b)
	}
}

func TestColorPicker_SetRGB(t *testing.T) {
	cp := NewColorPicker()
	cp.SetRGB(255, 128, 0)
	c := cp.Color()
	if c.Type != buffer.ColorTrue {
		t.Errorf("color type = %v, want ColorTrue", c.Type)
	}
	if c.R() != 255 || c.G() != 128 || c.B() != 0 {
		t.Errorf("RGB mismatch")
	}
	hex := cp.HexString()
	if hex != "#ff8000" {
		t.Errorf("HexString = %s, want #ff8000", hex)
	}
}

func TestColorPicker_SetPaletteIndex(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(42)
	if cp.PaletteIndex() != 42 {
		t.Errorf("PaletteIndex = %d, want 42", cp.PaletteIndex())
	}
	c := cp.Color()
	if c.Type != buffer.Color256 || c.Val != 42 {
		t.Errorf("color = %v, want Color256(42)", c)
	}
}

func TestColorPicker_SetPaletteIndex_Clamped(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(-5)
	if cp.PaletteIndex() != 0 {
		t.Errorf("negative clamped = %d, want 0", cp.PaletteIndex())
	}
	cp.SetPaletteIndex(300)
	if cp.PaletteIndex() != 255 {
		t.Errorf("overflow clamped = %d, want 255", cp.PaletteIndex())
	}
}

func TestColorPicker_SetActiveChannel(t *testing.T) {
	cp := NewColorPicker()
	cp.SetActiveChannel(1)
	if cp.ActiveChannel() != 1 {
		t.Errorf("channel = %d, want 1", cp.ActiveChannel())
	}
	cp.SetActiveChannel(-1)
	if cp.ActiveChannel() != 0 {
		t.Errorf("negative channel = %d, want 0", cp.ActiveChannel())
	}
	cp.SetActiveChannel(5)
	if cp.ActiveChannel() != 2 {
		t.Errorf("overflow channel = %d, want 2", cp.ActiveChannel())
	}
}

// ─── OnChange callback ───

func TestColorPicker_OnChange(t *testing.T) {
	cp := NewColorPicker()
	var fired bool
	var received buffer.Color
	cp.OnChange = func(c buffer.Color) {
		fired = true
		received = c
	}
	cp.SetRGB(100, 200, 50)
	if !fired {
		t.Error("OnChange not fired")
	}
	if received.R() != 100 || received.G() != 200 || received.B() != 50 {
		t.Errorf("received = %v, want RGB(100,200,50)", received)
	}
}

func TestColorPicker_OnChange_Palette(t *testing.T) {
	cp := NewColorPicker()
	var fired bool
	cp.OnChange = func(c buffer.Color) { fired = true }
	cp.SetPaletteIndex(10)
	if !fired {
		t.Error("OnChange not fired on palette")
	}
}

// ─── Palette keyboard ───

func TestColorPicker_PaletteKey_Right(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(0)
	consumed := cp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if !consumed {
		t.Error("right key not consumed")
	}
	if cp.PaletteIndex() != 1 {
		t.Errorf("after right: idx=%d, want 1", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Left(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(5)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if cp.PaletteIndex() != 4 {
		t.Errorf("after left: idx=%d, want 4", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Left_AtZero(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(0)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if cp.PaletteIndex() != 0 {
		t.Errorf("left at 0: idx=%d, want 0", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Right_AtMax(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(255)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if cp.PaletteIndex() != 255 {
		t.Errorf("right at max: idx=%d, want 255", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Down(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(0)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	if cp.PaletteIndex() != 16 {
		t.Errorf("after down: idx=%d, want 16", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Up(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(50)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	if cp.PaletteIndex() != 34 {
		t.Errorf("after up: idx=%d, want 34", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Home(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(100)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyHome})
	if cp.PaletteIndex() != 0 {
		t.Errorf("after home: idx=%d, want 0", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_End(t *testing.T) {
	cp := NewColorPicker()
	cp.SetPaletteIndex(0)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	if cp.PaletteIndex() != 255 {
		t.Errorf("after end: idx=%d, want 255", cp.PaletteIndex())
	}
}

func TestColorPicker_PaletteKey_Enter(t *testing.T) {
	cp := NewColorPicker()
	var confirmed buffer.Color
	cp.OnConfirm = func(c buffer.Color) { confirmed = c }
	consumed := cp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !consumed {
		t.Error("enter not consumed")
	}
	if confirmed.Type == buffer.ColorNone {
		t.Error("OnConfirm not fired")
	}
}

func TestColorPicker_PaletteKey_Unknown(t *testing.T) {
	cp := NewColorPicker()
	consumed := cp.HandleKey(&term.KeyEvent{Key: term.KeyTab})
	if consumed {
		t.Error("unknown key should not be consumed in palette mode")
	}
}

// ─── RGB keyboard ───

func TestColorPicker_RGBKey_Up(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(0) // Red
	cp.SetRGB(100, 100, 100)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	r, _, _ := cp.RGBValues()
	if r != 101 {
		t.Errorf("after up on red: r=%d, want 101", r)
	}
}

func TestColorPicker_RGBKey_Down(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(1) // Green
	cp.SetRGB(100, 100, 100)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	_, g, _ := cp.RGBValues()
	if g != 99 {
		t.Errorf("after down on green: g=%d, want 99", g)
	}
}

func TestColorPicker_RGBKey_Left_SwitchChannel(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(2) // Blue
	cp.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	if cp.ActiveChannel() != 1 {
		t.Errorf("after left: channel=%d, want 1", cp.ActiveChannel())
	}
}

func TestColorPicker_RGBKey_Right_SwitchChannel(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(0) // Red
	cp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	if cp.ActiveChannel() != 1 {
		t.Errorf("after right: channel=%d, want 1", cp.ActiveChannel())
	}
}

func TestColorPicker_RGBKey_VimKeys(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(0)
	cp.SetRGB(50, 50, 50)

	cp.HandleKey(&term.KeyEvent{Rune: 'k'}) // +1
	r, _, _ := cp.RGBValues()
	if r != 51 {
		t.Errorf("after k: r=%d, want 51", r)
	}

	cp.HandleKey(&term.KeyEvent{Rune: 'j'}) // -1
	r, _, _ = cp.RGBValues()
	if r != 50 {
		t.Errorf("after j: r=%d, want 50", r)
	}

	cp.HandleKey(&term.KeyEvent{Rune: 'H'}) // +10
	r, _, _ = cp.RGBValues()
	if r != 60 {
		t.Errorf("after H: r=%d, want 60", r)
	}

	cp.HandleKey(&term.KeyEvent{Rune: 'L'}) // -10
	r, _, _ = cp.RGBValues()
	if r != 50 {
		t.Errorf("after L: r=%d, want 50", r)
	}
}

func TestColorPicker_RGBKey_ClampMax(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(0)
	cp.SetRGB(255, 0, 0)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyUp})
	r, _, _ := cp.RGBValues()
	if r != 255 {
		t.Errorf("clamp max: r=%d, want 255", r)
	}
}

func TestColorPicker_RGBKey_ClampMin(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetActiveChannel(0)
	cp.SetRGB(0, 0, 0)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyDown})
	r, _, _ := cp.RGBValues()
	if r != 0 {
		t.Errorf("clamp min: r=%d, want 0", r)
	}
}

func TestColorPicker_RGBKey_Enter(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	var confirmed bool
	cp.OnConfirm = func(c buffer.Color) { confirmed = true }
	cp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !confirmed {
		t.Error("OnConfirm not fired in RGB mode")
	}
}

// ─── Hex keyboard ───

func TestColorPicker_HexKey_TypeDigits(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	cp.HandleKey(&term.KeyEvent{Rune: 'f'})
	cp.HandleKey(&term.KeyEvent{Rune: 'f'})
	cp.HandleKey(&term.KeyEvent{Rune: '0'})
	cp.HandleKey(&term.KeyEvent{Rune: '0'})
	cp.HandleKey(&term.KeyEvent{Rune: '0'})
	cp.HandleKey(&term.KeyEvent{Rune: '0'})
	hex := cp.HexString()
	if hex != "#ff0000" {
		t.Errorf("hex = %s, want #ff0000", hex)
	}
	c := cp.Color()
	if c.R() != 255 || c.G() != 0 || c.B() != 0 {
		t.Errorf("color from hex: R=%d G=%d B=%d", c.R(), c.G(), c.B())
	}
}

func TestColorPicker_HexKey_Backspace(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	cp.SetColor(buffer.RGB(0xFF, 0x80, 0x40)) // hexBuf = "ff8040"
	// Cursor at end
	cp.HandleKey(&term.KeyEvent{Key: term.KeyEnd})
	cp.HandleKey(&term.KeyEvent{Key: term.KeyBackspace})
	hex := cp.HexString()
	if len(hex) != 6 { // # + 5 chars
		t.Errorf("after backspace: hex=%s (len %d), want len 6", hex, 6)
	}
}

func TestColorPicker_HexKey_LeftRight(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	cp.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	cp.HandleKey(&term.KeyEvent{Key: term.KeyLeft})
	cp.HandleKey(&term.KeyEvent{Key: term.KeyRight})
	// Should not panic
}

func TestColorPicker_HexKey_Enter(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	var confirmed bool
	cp.OnConfirm = func(c buffer.Color) { confirmed = true }
	cp.HandleKey(&term.KeyEvent{Key: term.KeyEnter})
	if !confirmed {
		t.Error("OnConfirm not fired in hex mode")
	}
}

func TestColorPicker_HexKey_InvalidDigit(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	consumed := cp.HandleKey(&term.KeyEvent{Rune: 'z'})
	if consumed {
		t.Error("invalid hex digit should not be consumed")
	}
}

func TestColorPicker_HexKey_UppercaseDigits(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	cp.HandleKey(&term.KeyEvent{Rune: 'A'})
	cp.HandleKey(&term.KeyEvent{Rune: 'B'})
	cp.HandleKey(&term.KeyEvent{Rune: 'C'})
	cp.HandleKey(&term.KeyEvent{Rune: 'D'})
	cp.HandleKey(&term.KeyEvent{Rune: 'E'})
	cp.HandleKey(&term.KeyEvent{Rune: 'F'})
	c := cp.Color()
	if c.R() != 0xAB || c.G() != 0xCD || c.B() != 0xEF {
		t.Errorf("uppercase hex: R=%d G=%d B=%d, want AB CD EF", c.R(), c.G(), c.B())
	}
}

// ─── Measure & Paint ───

func TestColorPicker_Measure(t *testing.T) {
	cp := NewColorPicker()
	s := cp.Measure(Unbounded())
	if s.W <= 0 || s.H <= 0 {
		t.Errorf("measure = %v, want positive dims", s)
	}
}

func TestColorPicker_Measure_Constrained(t *testing.T) {
	cp := NewColorPicker()
	s := cp.Measure(Bounded(20, 10))
	if s.W > 20 {
		t.Errorf("constrained W = %d, want <= 20", s.W)
	}
	if s.H > 10 {
		t.Errorf("constrained H = %d, want <= 10", s.H)
	}
}

func TestColorPicker_Paint_Palette(t *testing.T) {
	cp := NewColorPicker()
	cp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	buf := buffer.NewBuffer(60, 20)
	cp.Paint(buf)
	// Check that something was painted
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("nothing painted at (0,0)")
	}
}

func TestColorPicker_Paint_RGB(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerRGB)
	cp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 16})
	buf := buffer.NewBuffer(50, 16)
	cp.Paint(buf)
}

func TestColorPicker_Paint_Hex(t *testing.T) {
	cp := NewColorPicker()
	cp.SetMode(PickerHex)
	cp.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 16})
	buf := buffer.NewBuffer(50, 16)
	cp.Paint(buf)
}

func TestColorPicker_Paint_ZeroBounds(t *testing.T) {
	cp := NewColorPicker()
	cp.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cp.Paint(buf) // should not panic
}

// ─── ColorName helper ───

func TestColorName_NamedColors(t *testing.T) {
	tests := []struct {
		val  int
		name string
	}{
		{0, "Black"},
		{1, "Red"},
		{7, "White"},
		{12, "Bright Blue"},
		{15, "Bright White"},
	}
	for _, tt := range tests {
		name := ColorName(buffer.NamedColor(tt.val))
		if name != tt.name {
			t.Errorf("ColorName(NamedColor(%d)) = %s, want %s", tt.val, name, tt.name)
		}
	}
}

func TestColorName_TrueColor(t *testing.T) {
	name := ColorName(buffer.RGB(255, 128, 0))
	if name != "#FF8000" {
		t.Errorf("ColorName(RGB(255,128,0)) = %s, want #FF8000", name)
	}
}

func TestColorName_256(t *testing.T) {
	name := ColorName(buffer.Color256Val(42))
	if name != "256 #42" {
		t.Errorf("ColorName(256(42)) = %s, want '256 #42'", name)
	}
}

func TestColorName_Default(t *testing.T) {
	name := ColorName(buffer.NoColor())
	if name != "Default" {
		t.Errorf("ColorName(NoColor) = %s, want 'Default'", name)
	}
}

// ─── Style ───

func TestColorPicker_SetStyle(t *testing.T) {
	cp := NewColorPicker()
	s := DefaultColorPickerStyle()
	s.Title = buffer.Style{Fg: buffer.Green, Flags: buffer.Bold}
	cp.SetStyle(s)
	got := cp.Style()
	if got.Title.Fg != buffer.Green {
		t.Error("style not set correctly")
	}
}

// ─── Concurrent safety ───

func TestColorPicker_ConcurrentAccess(t *testing.T) {
	cp := NewColorPicker()
	// Set bounds ONCE before concurrent operations (SetBounds on BaseComponent is not locked)
	cp.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			cp.SetRGB(uint8(n*10), uint8(n*20), uint8(n*5))
			cp.SetPaletteIndex(n * 10)
			cp.SetMode(ColorPickerMode(n % pickerModeCount))
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cp.Color()
			cp.Mode()
			cp.PaletteIndex()
			cp.RGBValues()
			cp.HexString()
		}()
	}

	// Concurrent painters (each uses its own buffer, bounds already set)
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := buffer.NewBuffer(60, 20)
			cp.Paint(buf)
		}()
	}

	wg.Wait()
}
