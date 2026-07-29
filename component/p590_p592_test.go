package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── ActivityRing Tests ───

func TestActivityRingBasic(t *testing.T) {
	ar := NewActivityRing()
	ar.SetCount(75, 100)
	if c := ar.Count(); c != 75 {
		t.Errorf("Count = %d, want 75", c)
	}
}

func TestActivityRingZero(t *testing.T) {
	ar := NewActivityRing()
	ar.SetCount(0, 100)
	if c := ar.Count(); c != 0 {
		t.Errorf("Count = %d, want 0", c)
	}
}

func TestActivityRingClamp(t *testing.T) {
	ar := NewActivityRing()
	ar.SetCount(150, 100)
	if c := ar.Count(); c != 100 {
		t.Errorf("Count = %d, want 100 (clamped)", c)
	}
	ar.SetCount(-5, 0)
	if c := ar.Count(); c != 0 {
		t.Errorf("Count = %d, want 0 (clamped)", c)
	}
}

func TestActivityRingLabel(t *testing.T) {
	ar := NewActivityRing()
	ar.SetLabel("evt")
	if ar.label != "evt" {
		t.Errorf("label = %q, want 'evt'", ar.label)
	}
}

func TestActivityRingPaint(t *testing.T) {
	ar := NewActivityRing()
	ar.SetCount(50, 100)
	ar.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(8, 1)
	ar.Paint(buf)
	// Should have ring chars
	if r := buf.GetCell(0, 0).Rune; r == 0 || r == ' ' {
		t.Error("Paint should show ring character")
	}
}

func TestActivityRingChildren(t *testing.T) {
	ar := NewActivityRing()
	if c := ar.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestActivityRingStyle(t *testing.T) {
	ar := NewActivityRing()
	ar.SetStyle(ActivityRingStyle{
		Filled: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Empty:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
		Center: buffer.Style{Fg: buffer.RGB(255, 255, 255)},
	})
	ar.SetCount(80, 100)
	buf := buffer.NewBuffer(8, 1)
	ar.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	ar.Paint(buf)
}

// ─── AIModelRankList Tests ───

func TestAIModelRankBasic(t *testing.T) {
	rl := NewAIModelRankList()
	rl.AddModel("GPT-4o", 92, 1)
	rl.AddModel("Claude", 90, -1)
	if n := rl.Count(); n != 2 {
		t.Errorf("Count = %d, want 2", n)
	}
}

func TestAIModelRankOverflow(t *testing.T) {
	rl := NewAIModelRankList()
	for i := 0; i < modelRankMaxEntries+5; i++ {
		rl.AddModel("m", 50, 0)
	}
	if n := rl.Count(); n != modelRankMaxEntries {
		t.Errorf("Count = %d, want %d (capped)", n, modelRankMaxEntries)
	}
}

func TestAIModelRankClear(t *testing.T) {
	rl := NewAIModelRankList()
	rl.AddModel("a", 50, 0)
	rl.Clear()
	if n := rl.Count(); n != 0 {
		t.Errorf("Count after Clear = %d, want 0", n)
	}
}

func TestAIModelRankPaint(t *testing.T) {
	rl := NewAIModelRankList()
	rl.AddModel("GPT-4o", 92, 1)
	rl.AddModel("Claude", 90, -1)
	rl.AddModel("Llama", 85, 0)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	rl.Paint(buf)
	// Row 0 should start with "1."
	if r := buf.GetCell(0, 0).Rune; r != '1' {
		t.Errorf("First rune = %q, want '1'", r)
	}
}

func TestAIModelRankChildren(t *testing.T) {
	rl := NewAIModelRankList()
	if c := rl.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestAIModelRankStyle(t *testing.T) {
	rl := NewAIModelRankList()
	rl.SetStyle(ModelRankStyle{
		Rank:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Name:  buffer.Style{Fg: buffer.RGB(255, 255, 255)},
		Score: buffer.Style{Fg: buffer.RGB(255, 215, 0)},
		Up:    buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Down:  buffer.Style{Fg: buffer.RGB(255, 0, 0)},
		Same:  buffer.Style{Fg: buffer.RGB(128, 128, 128)},
	})
	rl.AddModel("Test", 88, 2)
	buf := buffer.NewBuffer(30, 3)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	rl.Paint(buf)
}

// ─── MiniMap Tests ───

func TestMiniMapBasic(t *testing.T) {
	mm := NewMiniMap()
	mm.SetContent(500, 100, 80)
	if s := mm.ViewStart(); s != 100 {
		t.Errorf("ViewStart = %d, want 100", s)
	}
}

func TestMiniMapClamp(t *testing.T) {
	mm := NewMiniMap()
	mm.SetContent(500, -10, 80)
	if s := mm.ViewStart(); s != 0 {
		t.Errorf("ViewStart = %d, want 0 (clamped)", s)
	}
	mm.SetContent(500, 450, 80) // start > total-height
	if s := mm.ViewStart(); s != 420 {
		t.Errorf("ViewStart = %d, want 420 (clamped to total-height)", s)
	}
}

func TestMiniMapSmall(t *testing.T) {
	mm := NewMiniMap()
	mm.SetContent(10, 0, 5)
	// Should not panic, visStartRow should be reasonable
	if mm.visStartRow < 0 {
		t.Errorf("visStartRow = %d, want >= 0", mm.visStartRow)
	}
}

func TestMiniMapPaint(t *testing.T) {
	mm := NewMiniMap()
	mm.SetContent(200, 50, 50)
	mm.SetBounds(Rect{X: 0, Y: 0, W: 3, H: miniMapHeight})
	buf := buffer.NewBuffer(3, miniMapHeight)
	mm.Paint(buf)
	// Should have visible block
	hasVisible := false
	for i := 0; i < miniMapHeight; i++ {
		if buf.GetCell(0, i).Rune == '█' {
			hasVisible = true
			break
		}
	}
	if !hasVisible {
		t.Error("Paint should show visible region")
	}
}

func TestMiniMapChildren(t *testing.T) {
	mm := NewMiniMap()
	if c := mm.Children(); c != nil {
		t.Errorf("Children = %v, want nil", c)
	}
}

func TestMiniMapStyle(t *testing.T) {
	mm := NewMiniMap()
	mm.SetStyle(MiniMapStyle{
		Full:    buffer.Style{Fg: buffer.RGB(30, 30, 30)},
		Visible: buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Border:  buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	mm.SetContent(300, 100, 80)
	buf := buffer.NewBuffer(3, miniMapHeight)
	mm.SetBounds(Rect{X: 0, Y: 0, W: 3, H: miniMapHeight})
	mm.Paint(buf)
}

// ─── Benchmarks ───

func BenchmarkPaintActivityRing(b *testing.B) {
	ar := NewActivityRing()
	ar.SetCount(75, 100)
	ar.SetBounds(Rect{X: 0, Y: 0, W: 8, H: 1})
	buf := buffer.NewBuffer(8, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ar.Paint(buf)
	}
}

func BenchmarkPaintAIModelRankList(b *testing.B) {
	rl := NewAIModelRankList()
	rl.AddModel("GPT-4o", 92, 1)
	rl.AddModel("Claude-3.5", 90, -1)
	rl.AddModel("Llama-3", 85, 0)
	rl.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	buf := buffer.NewBuffer(30, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Paint(buf)
	}
}

func BenchmarkPaintMiniMap(b *testing.B) {
	mm := NewMiniMap()
	mm.SetContent(500, 100, 80)
	mm.SetBounds(Rect{X: 0, Y: 0, W: 3, H: miniMapHeight})
	buf := buffer.NewBuffer(3, miniMapHeight)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mm.Paint(buf)
	}
}
