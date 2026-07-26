package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP396_NewToast(t *testing.T) {
	tt := NewToast("Saved", ToastSuccess)
	if tt.Message() != "Saved" {
		t.Errorf("Message = %q", tt.Message())
	}
	if tt.Level() != ToastSuccess {
		t.Errorf("Level = %v", tt.Level())
	}
	if tt.Duration() != 3*time.Second {
		t.Errorf("Duration = %v", tt.Duration())
	}
	if tt.Position() != ToastBottomRight {
		t.Errorf("Position = %v", tt.Position())
	}
	if tt.ID() == "" {
		t.Error("ID empty")
	}
}

func TestP396_SetMessage(t *testing.T) {
	tt := NewToast("old", ToastInfo)
	tt.SetMessage("new")
	if tt.Message() != "new" {
		t.Errorf("Message = %q", tt.Message())
	}
}

func TestP396_SetLevel(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	tt.SetLevel(ToastError)
	if tt.Level() != ToastError {
		t.Errorf("Level = %v", tt.Level())
	}
}

func TestP396_SetDuration(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	tt.SetDuration(5 * time.Second)
	if tt.Duration() != 5*time.Second {
		t.Errorf("Duration = %v", tt.Duration())
	}
}

func TestP396_SetPosition(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	for _, p := range []ToastPosition{ToastTopLeft, ToastTopRight, ToastBottomLeft, ToastTopCenter, ToastBottomCenter} {
		tt.SetPosition(p)
		if tt.Position() != p {
			t.Errorf("Position = %v, want %v", tt.Position(), p)
		}
	}
}

func TestP396_Dismiss(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	if tt.Dismissed() {
		t.Error("should not be dismissed initially")
	}
	tt.Dismiss()
	if !tt.Dismissed() {
		t.Error("should be dismissed after Dismiss()")
	}
}

func TestP396_ResolvePosition(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	tt.mu.Lock()
	defer tt.mu.Unlock()
	tests := []struct {
		pos              ToastPosition
		screenW, screenH int
		toastW, toastH   int
		wantX, wantY     int
	}{
		{ToastTopLeft, 80, 24, 10, 1, 0, 0},
		{ToastTopRight, 80, 24, 10, 1, 70, 0},
		{ToastBottomLeft, 80, 24, 10, 1, 0, 23},
		{ToastBottomRight, 80, 24, 10, 1, 70, 23},
		{ToastTopCenter, 80, 24, 10, 1, 35, 0},
		{ToastBottomCenter, 80, 24, 10, 1, 35, 23},
	}
	for _, tc := range tests {
		tt.pos = tc.pos
		gotX, gotY := tt.resolvePositionLocked(tc.screenW, tc.screenH, tc.toastW, tc.toastH)
		if gotX != tc.wantX || gotY != tc.wantY {
			t.Errorf("pos %v: got (%d,%d), want (%d,%d)", tc.pos, gotX, gotY, tc.wantX, tc.wantY)
		}
	}
}

func TestP396_Measure(t *testing.T) {
	tt := NewToast("Hello World", ToastInfo)
	s := tt.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// " ℹ Hello World " = 4 + 11 = 15
	if s.W != 15 {
		t.Errorf("W = %d, want 15", s.W)
	}
	if s.H != 1 {
		t.Errorf("H = %d", s.H)
	}
}

func TestP396_Measure_ShortMessage(t *testing.T) {
	tt := NewToast("Hi", ToastInfo)
	s := tt.Measure(Constraints{MaxWidth: 80, MaxHeight: 5})
	// minimum width 6
	if s.W != 6 {
		t.Errorf("W = %d, want 6", s.W)
	}
}

func TestP396_Paint_Success(t *testing.T) {
	tt := NewToast("Done", ToastSuccess)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	tt.Paint(buf)
	// Cell 0 = padding space
	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("cell[0] = %q, want ' '", string(c0.Rune))
	}
}

func TestP396_Paint_Error(t *testing.T) {
	tt := NewToast("Failed", ToastError)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	tt.Paint(buf)
	c0 := buf.GetCell(0, 0)
	if c0.Rune != ' ' {
		t.Errorf("cell[0] = %q, want ' '", string(c0.Rune))
	}
}

func TestP396_Paint_Warning(t *testing.T) {
	tt := NewToast("Careful", ToastWarning)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 1})
	buf := buffer.NewBuffer(15, 1)
	tt.Paint(buf)
}

func TestP396_Paint_Info(t *testing.T) {
	tt := NewToast("Note", ToastInfo)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	tt.Paint(buf)
}

func TestP396_Paint_Dismissed(t *testing.T) {
	tt := NewToast("hidden", ToastInfo)
	tt.Dismiss()
	tt.SetBounds(Rect{X: 0, Y: 0, W: 12, H: 1})
	buf := buffer.NewBuffer(12, 1)
	tt.Paint(buf)
	// Should not draw anything
	c := buf.GetCell(0, 0)
	if c.Rune != ' ' {
		t.Error("dismissed toast should not render")
	}
}

func TestP396_Paint_LongMessage(t *testing.T) {
	tt := NewToast("This is a very long message that exceeds bounds width", ToastInfo)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 1})
	buf := buffer.NewBuffer(10, 1)
	tt.Paint(buf) // should truncate with …
}

func TestP396_Paint_ZeroBounds(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	tt.Paint(buf)
}

func TestP396_Paint_NonZeroOffset(t *testing.T) {
	tt := NewToast("OK", ToastSuccess)
	tt.SetBounds(Rect{X: 70, Y: 23, W: 10, H: 1})
	buf := buffer.NewBuffer(80, 24)
	tt.Paint(buf)
	c := buf.GetCell(70, 23)
	if c.Rune != ' ' {
		t.Errorf("offset cell = %q, want ' '", string(c.Rune))
	}
}

func TestP396_FormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{3 * time.Second, "3.0s"},
		{1500 * time.Millisecond, "1.5s"},
	}
	for _, tc := range tests {
		got := formatToastDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatToastDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestP396_Concurrent(t *testing.T) {
	tt := NewToast("test", ToastInfo)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			tt.SetMessage("concurrent")
			tt.SetLevel(ToastError)
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = tt.Message()
		_ = tt.Level()
	}
	<-done
}

func TestP396_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Toast)(nil)
}

func BenchmarkP396_Toast_Paint(b *testing.B) {
	tt := NewToast("Operation completed successfully", ToastSuccess)
	tt.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 1})
	buf := buffer.NewBuffer(40, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tt.Paint(buf)
	}
}
