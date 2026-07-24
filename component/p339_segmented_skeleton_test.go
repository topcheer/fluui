package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── SegmentedControl ─────────────────────────────────────

func TestP339_SegmentedControl_Create(t *testing.T) {
	sc := NewSegmentedControl([]string{"Chat", "Code", "Settings"})
	if sc.ActiveIndex() != 0 {
		t.Errorf("active = %d, want 0", sc.ActiveIndex())
	}
	if sc.ActiveLabel() != "Chat" {
		t.Errorf("label = %q", sc.ActiveLabel())
	}
	if sc.SegmentCount() != 3 {
		t.Errorf("count = %d, want 3", sc.SegmentCount())
	}
}

func TestP339_SegmentedControl_SetActive(t *testing.T) {
	sc := NewSegmentedControl([]string{"A", "B", "C"})
	sc.SetActive(2)
	if sc.ActiveIndex() != 2 {
		t.Errorf("active = %d, want 2", sc.ActiveIndex())
	}
	if sc.ActiveLabel() != "C" {
		t.Errorf("label = %q", sc.ActiveLabel())
	}
}

func TestP339_SegmentedControl_SetActive_Invalid(t *testing.T) {
	sc := NewSegmentedControl([]string{"A", "B"})
	sc.SetActive(5)  // out of range
	sc.SetActive(-1) // negative
	if sc.ActiveIndex() != 0 {
		t.Errorf("invalid set should not change: %d", sc.ActiveIndex())
	}
}

func TestP339_SegmentedControl_NextPrev(t *testing.T) {
	sc := NewSegmentedControl([]string{"A", "B", "C"})
	sc.SelectNext()
	if sc.ActiveIndex() != 1 {
		t.Errorf("after next: %d", sc.ActiveIndex())
	}
	sc.SelectNext()
	sc.SelectNext() // wraps to 0
	if sc.ActiveIndex() != 0 {
		t.Errorf("after wrap: %d", sc.ActiveIndex())
	}
	sc.SelectPrev() // wraps to 2
	if sc.ActiveIndex() != 2 {
		t.Errorf("after prev wrap: %d", sc.ActiveIndex())
	}
}

func TestP339_SegmentedControl_SetSegments(t *testing.T) {
	sc := NewSegmentedControl([]string{"A", "B", "C", "D"})
	sc.SetActive(3)
	sc.SetSegments([]string{"X", "Y"})
	if sc.ActiveIndex() != 1 {
		t.Errorf("active should clamp: %d", sc.ActiveIndex())
	}
}

func TestP339_SegmentedControl_Empty(t *testing.T) {
	sc := NewSegmentedControl(nil)
	sc.SelectNext() // should not panic
	sc.SelectPrev()
	if sc.ActiveLabel() != "" {
		t.Errorf("empty label should be %q", sc.ActiveLabel())
	}
}

func TestP339_SegmentedControl_Measure(t *testing.T) {
	sc := NewSegmentedControl([]string{"AB", "CD"})
	s := sc.Measure(Constraints{MaxWidth: 80, MaxHeight: 1})
	if s.H != 1 {
		t.Errorf("height = %d, want 1", s.H)
	}
}

func TestP339_SegmentedControl_Paint(t *testing.T) {
	sc := NewSegmentedControl([]string{"Chat", "Code"})
	sc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	sc.Paint(buf)

	// Should render something
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("expected non-empty cell at position 0")
	}
}

func TestP339_SegmentedControl_Paint_Empty(t *testing.T) {
	sc := NewSegmentedControl(nil)
	sc.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 1})
	buf := buffer.NewBuffer(30, 1)
	sc.Paint(buf) // should not panic
}

// ─── SkeletonLoader ───────────────────────────────────────

func TestP339_SkeletonLoader_Create(t *testing.T) {
	sk := NewSkeletonLoader([]SkeletonBlock{
		{X: 0, Y: 0, W: 20, H: 1},
		{X: 0, Y: 2, W: 15, H: 1},
	})
	if len(sk.Blocks()) != 2 {
		t.Errorf("blocks = %d, want 2", len(sk.Blocks()))
	}
	if sk.IsRunning() {
		t.Error("should not be running on creation")
	}
}

func TestP339_SkeletonLoader_Text(t *testing.T) {
	sk := NewSkeletonText(3, 40)
	blocks := sk.Blocks()
	if len(blocks) != 3 {
		t.Fatalf("blocks = %d, want 3", len(blocks))
	}
	// Last line should be shorter
	if blocks[2].W >= 40 {
		t.Error("last line should be shorter")
	}
}

func TestP339_SkeletonLoader_SetBlocks(t *testing.T) {
	sk := NewSkeletonText(2, 20)
	sk.SetBlocks([]SkeletonBlock{{X: 0, Y: 0, W: 10, H: 2}})
	if len(sk.Blocks()) != 1 {
		t.Errorf("blocks = %d, want 1", len(sk.Blocks()))
	}
}

func TestP339_SkeletonLoader_StartStop(t *testing.T) {
	sk := NewSkeletonText(3, 20)
	sk.Start(10 * time.Millisecond)
	if !sk.IsRunning() {
		t.Error("should be running")
	}
	time.Sleep(30 * time.Millisecond)
	sk.Stop()
	if sk.IsRunning() {
		t.Error("should not be running after stop")
	}
}

func TestP339_SkeletonLoader_StopIdempotent(t *testing.T) {
	sk := NewSkeletonText(2, 10)
	sk.Start(10 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	sk.Stop()
	sk.Stop()
	sk.Stop()
}

func TestP339_SkeletonLoader_FrameIndex(t *testing.T) {
	sk := NewSkeletonText(2, 10)
	if sk.FrameIndex() != 0 {
		t.Errorf("initial frame = %d, want 0", sk.FrameIndex())
	}
}

func TestP339_SkeletonLoader_Measure(t *testing.T) {
	sk := NewSkeletonLoader([]SkeletonBlock{
		{X: 0, Y: 0, W: 20, H: 1},
		{X: 0, Y: 2, W: 10, H: 3},
	})
	s := sk.Measure(Constraints{MaxWidth: 40, MaxHeight: 10})
	if s.H < 5 {
		t.Errorf("height = %d, should be at least 5", s.H)
	}
}

func TestP339_SkeletonLoader_Paint(t *testing.T) {
	sk := NewSkeletonText(3, 30)
	sk.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	buf := buffer.NewBuffer(40, 10)
	sk.Paint(buf)

	// Check some cells have non-default background
	filled := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			c := buf.GetCell(x, y)
			if c.Bg.Type != 0 || c.Bg.Val != 0 {
				filled++
			}
		}
	}
	if filled == 0 {
		t.Error("expected some filled cells")
	}
}

func TestP339_SkeletonLoader_Paint_BoundsClamped(t *testing.T) {
	sk := NewSkeletonLoader([]SkeletonBlock{
		{X: 0, Y: 0, W: 100, H: 100},
	})
	sk.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})
	buf := buffer.NewBuffer(5, 3)
	sk.Paint(buf) // should not panic with oversized blocks
}

func BenchmarkSegmentedControl_Paint(b *testing.B) {
	sc := NewSegmentedControl([]string{"Chat", "Code", "Settings", "Docs"})
	sc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 1})
	buf := buffer.NewBuffer(60, 1)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sc.Paint(buf)
	}
}

func BenchmarkSkeletonLoader_Paint(b *testing.B) {
	sk := NewSkeletonText(5, 40)
	sk.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})
	buf := buffer.NewBuffer(60, 12)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sk.Paint(buf)
	}
}
