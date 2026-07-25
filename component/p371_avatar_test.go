package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P371: Avatar component tests

func TestP371_NewAvatar(t *testing.T) {
	a := NewAvatar("Alice Brown")
	if a.Name() != "Alice Brown" {
		t.Errorf("Name = %q", a.Name())
	}
	if a.ID() == "" {
		t.Error("ID should not be empty")
	}
}

func TestP371_SetName(t *testing.T) {
	a := NewAvatar("Alice")
	a.SetName("Bob")
	if a.Name() != "Bob" {
		t.Errorf("Name = %q", a.Name())
	}
}

func TestP371_Initials(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Alice Brown", "AB"},
		{"alice brown", "AB"},
		{"Alice", "A"},
		{"A", "A"},
		{"", ""},
		{"John Doe Smith", "JD"},
		{"  leading", "L"},
		{"trailing  ", "T"},
		{"Mary-Jane Watson", "MW"},
		{"O'Brien", "O"},
		{"中文 Name", "N"},
	}
	for _, tt := range tests {
		a := NewAvatar(tt.name)
		got := a.Initials()
		if got != tt.want {
			t.Errorf("Initials(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestP371_SetInitials(t *testing.T) {
	a := NewAvatar("Alice Brown")
	a.SetInitials("XY")
	if a.Initials() != "XY" {
		t.Errorf("Initials = %q, want XY", a.Initials())
	}
	// Revert to auto
	a.SetInitials("")
	if a.Initials() != "AB" {
		t.Errorf("Initials = %q, want AB (auto)", a.Initials())
	}
}

func TestP371_Icon(t *testing.T) {
	a := NewAvatar("Alice")
	a.SetIcon("🤖")
	if a.Icon() != "🤖" {
		t.Errorf("Icon = %q", a.Icon())
	}
	a.SetIcon("")
	if a.Icon() != "" {
		t.Error("Icon should be empty")
	}
}

func TestP371_Size(t *testing.T) {
	a := NewAvatar("Alice")
	if a.AvatarSize() != AvatarMedium {
		t.Errorf("default size = %v, want AvatarMedium", a.AvatarSize())
	}
	a.SetSize(AvatarSmall)
	if a.AvatarSize() != AvatarSmall {
		t.Errorf("size = %v, want AvatarSmall", a.AvatarSize())
	}
	a.SetSize(AvatarLarge)
	if a.AvatarSize() != AvatarLarge {
		t.Errorf("size = %v, want AvatarLarge", a.AvatarSize())
	}
}

func TestP371_Bg(t *testing.T) {
	a := NewAvatar("Alice")
	// Zero value = auto
	c := a.Bg()
	if c.Type != buffer.ColorNone {
		t.Errorf("default Bg type = %v, want ColorNone", c.Type)
	}
	override := buffer.RGB(0xFF, 0x00, 0x00)
	a.SetBg(override)
	if a.Bg() != override {
		t.Error("Bg override not applied")
	}
}

func TestP371_Measure(t *testing.T) {
	// Small: 1x1
	a := NewAvatar("X")
	a.SetSize(AvatarSmall)
	s := a.Measure(Constraints{MaxWidth: 10, MaxHeight: 5})
	if s.W != 1 || s.H != 1 {
		t.Errorf("Small Measure = %v, want {1,1}", s)
	}

	// Medium: 3x1
	a.SetSize(AvatarMedium)
	s = a.Measure(Constraints{MaxWidth: 10, MaxHeight: 5})
	if s.W != 3 || s.H != 1 {
		t.Errorf("Medium Measure = %v, want {3,1}", s)
	}

	// Large: 3x1
	a.SetSize(AvatarLarge)
	s = a.Measure(Constraints{MaxWidth: 10, MaxHeight: 5})
	if s.W != 3 || s.H != 1 {
		t.Errorf("Large Measure = %v, want {3,1}", s)
	}

	// Clamped by constraints
	a.SetSize(AvatarMedium)
	s = a.Measure(Constraints{MaxWidth: 2, MaxHeight: 1})
	if s.W != 2 || s.H != 1 {
		t.Errorf("Clamped Measure = %v, want {2,1}", s)
	}
}

func TestP371_Paint_Small(t *testing.T) {
	a := NewAvatar("Alice")
	a.SetSize(AvatarSmall)
	a.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	a.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'A' {
		t.Errorf("Small cell rune = %q, want 'A'", string(cell.Rune))
	}
}

func TestP371_Paint_Medium(t *testing.T) {
	a := NewAvatar("Alice Brown")
	a.SetSize(AvatarMedium)
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	// "AB" drawn at 0,1; cell 2 should be padding space
	c0 := buf.GetCell(0, 0)
	if c0.Rune != 'A' {
		t.Errorf("Medium cell[0] = %q, want 'A'", string(c0.Rune))
	}
	c1 := buf.GetCell(1, 0)
	if c1.Rune != 'B' {
		t.Errorf("Medium cell[1] = %q, want 'B'", string(c1.Rune))
	}
	c2 := buf.GetCell(2, 0)
	if c2.Rune != ' ' {
		t.Errorf("Medium cell[2] = %q, want ' '", string(c2.Rune))
	}
}

func TestP371_Paint_Large(t *testing.T) {
	a := NewAvatar("Bob")
	a.SetSize(AvatarLarge)
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'B' {
		t.Errorf("Large cell[0] = %q, want 'B'", string(cell.Rune))
	}
	if cell.Flags&buffer.Bold == 0 {
		t.Error("Large should have Bold flag")
	}
}

func TestP371_Paint_Icon(t *testing.T) {
	a := NewAvatar("AI")
	a.SetIcon("★")
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	// Icon drawn, rest padded with spaces
	c0 := buf.GetCell(0, 0)
	if c0.Rune != '★' {
		t.Errorf("Icon cell[0] = %q, want '★'", string(c0.Rune))
	}
}

func TestP371_Paint_ZeroBounds(t *testing.T) {
	a := NewAvatar("Alice")
	a.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	a.Paint(buf) // should not panic
}

func TestP371_Paint_NarrowWidth(t *testing.T) {
	a := NewAvatar("Alice Brown")
	a.SetBounds(Rect{X: 0, Y: 0, W: 1, H: 1})
	buf := buffer.NewBuffer(1, 1)
	a.Paint(buf) // should not panic, should clip
}

func TestP371_Paint_NonZeroOffset(t *testing.T) {
	a := NewAvatar("Alice Brown")
	a.SetBounds(Rect{X: 5, Y: 3, W: 3, H: 1})
	buf := buffer.NewBuffer(10, 5)
	a.Paint(buf)
	c := buf.GetCell(5, 3)
	if c.Rune != 'A' {
		t.Errorf("offset cell = %q, want 'A'", string(c.Rune))
	}
}

func TestP371_Paint_BgOverride(t *testing.T) {
	a := NewAvatar("Alice")
	override := buffer.RGB(0xFF, 0x00, 0x00)
	a.SetBg(override)
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	a.Paint(buf)
	cell := buf.GetCell(0, 0)
	if cell.Bg != override {
		t.Error("Bg override not applied in Paint")
	}
}

func TestP371_ColorConsistency(t *testing.T) {
	// Same name → same color
	c1 := avatarColorForName("Alice")
	c2 := avatarColorForName("Alice")
	if c1 != c2 {
		t.Error("same name should produce same color")
	}
	// Different names → likely different colors
	c3 := avatarColorForName("Bob")
	// Not guaranteed but overwhelmingly likely with 12 palette entries
	if c1 == c3 {
		t.Log("note: Alice and Bob happen to map to same color (low probability)")
	}
}

func TestP371_ExtractInitials_EdgeCases(t *testing.T) {
	// Single non-letter
	if got := extractInitials("123"); got != "" {
		t.Errorf("extractInitials('123') = %q, want ''", got)
	}
	// Numbers with letter
	if got := extractInitials("42x"); got != "X" {
		t.Errorf("extractInitials('42x') = %q, want 'X'", got)
	}
}

func TestP371_Concurrent(t *testing.T) {
	a := NewAvatar("Alice")
	done := make(chan struct{})
	go func() {
		for i := 0; i < 500; i++ {
			a.SetName("Bob")
			a.SetIcon("★")
			_ = a.Initials()
		}
		close(done)
	}()
	for i := 0; i < 500; i++ {
		_ = a.Name()
		_ = a.Icon()
	}
	<-done
}

func TestP371_SatisfiesComponent(t *testing.T) {
	var _ Component = (*Avatar)(nil)
}

// P371: Avatar benchmark — zero alloc
func BenchmarkP371_Avatar_Paint(b *testing.B) {
	a := NewAvatar("Alice Brown")
	a.SetSize(AvatarMedium)
	a.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 1})
	buf := buffer.NewBuffer(3, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Paint(buf)
	}
}
