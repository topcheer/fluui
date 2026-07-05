package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewCanvas_Defaults(t *testing.T) {
	c := NewCanvas()
	if c.ActiveLayer() == nil {
		t.Fatal("expected non-nil active layer")
	}
	if c.ActiveLayer().Name != "default" {
		t.Errorf("expected 'default' layer, got %q", c.ActiveLayer().Name)
	}
	if len(c.Layers()) != 1 {
		t.Errorf("expected 1 layer, got %d", len(c.Layers()))
	}
}

func TestCanvas_AddLayer(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	l := c.AddLayer("overlay")
	if l == nil {
		t.Fatal("expected non-nil layer")
	}
	if l.Name != "overlay" {
		t.Errorf("expected 'overlay', got %q", l.Name)
	}
	if c.ActiveLayer().Name != "overlay" {
		t.Error("new layer should be active")
	}
	if len(c.Layers()) != 2 {
		t.Errorf("expected 2 layers, got %d", len(c.Layers()))
	}
}

func TestCanvas_SetActiveLayer(t *testing.T) {
	c := NewCanvas()
	c.AddLayer("overlay")

	if !c.SetActiveLayer("default") {
		t.Error("expected to find 'default' layer")
	}
	if c.ActiveLayer().Name != "default" {
		t.Error("expected default to be active")
	}

	if c.SetActiveLayer("nonexistent") {
		t.Error("should not find nonexistent layer")
	}
}

func TestCanvas_RemoveLayer(t *testing.T) {
	c := NewCanvas()
	c.AddLayer("overlay")
	c.AddLayer("debug")

	// Cannot remove default
	if c.RemoveLayer("default") {
		t.Error("should not remove default layer")
	}

	// Remove overlay
	if !c.RemoveLayer("overlay") {
		t.Error("expected to remove overlay")
	}
	if len(c.Layers()) != 2 {
		t.Errorf("expected 2 layers after removal, got %d", len(c.Layers()))
	}
}

func TestCanvas_RemoveLayer_SwitchesActive(t *testing.T) {
	c := NewCanvas()
	c.AddLayer("temp")
	c.SetActiveLayer("temp")

	c.RemoveLayer("temp")
	if c.ActiveLayer().Name != "default" {
		t.Errorf("expected active to switch to default, got %q", c.ActiveLayer().Name)
	}
}

func TestCanvas_LayersReturnsCopy(t *testing.T) {
	c := NewCanvas()
	l1 := c.Layers()
	l1[0] = nil
	l2 := c.Layers()
	if l2[0] == nil {
		t.Error("Layers() should return a copy")
	}
}

func TestCanvas_ClearLayer(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.SetCell(5, 5, 'X', buffer.NamedColor(buffer.NamedRed))
	layer := c.ActiveLayer()
	if !layer.Cells[5][5].Used {
		t.Fatal("expected cell to be set before clear")
	}

	c.ClearLayer("default")
	layer = c.ActiveLayer()
	if layer.Cells[5][5].Used {
		t.Error("expected cell to be cleared")
	}
}

func TestCanvas_ClearAll(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	c.AddLayer("layer2")

	c.SetCell(1, 1, 'A', buffer.NamedColor(buffer.NamedRed))
	c.SetActiveLayer("layer2")
	c.SetCell(2, 2, 'B', buffer.NamedColor(buffer.NamedBlue))

	c.ClearAll()

	c.SetActiveLayer("default")
	if c.ActiveLayer().Cells[1][1].Used {
		t.Error("expected layer 1 cleared")
	}
	c.SetActiveLayer("layer2")
	if c.ActiveLayer().Cells[2][2].Used {
		t.Error("expected layer 2 cleared")
	}
}

func TestCanvas_SetBackgroundColor(t *testing.T) {
	c := NewCanvas()
	col := buffer.NamedColor(buffer.NamedBlue)
	c.SetBackgroundColor(col)
	// Just verify no panic
}

func TestCanvas_SetCell(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.SetCell(5, 3, 'X', buffer.NamedColor(buffer.NamedRed))
	cell := c.ActiveLayer().Cells[3][5]
	if cell.Char != 'X' {
		t.Errorf("expected 'X', got %q", string(cell.Char))
	}
	if !cell.Used {
		t.Error("expected Used=true")
	}
}

func TestCanvas_SetCell_OutOfBounds(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	c.SetCell(-1, 0, 'X', buffer.NamedColor(buffer.NamedRed))
	c.SetCell(0, -1, 'X', buffer.NamedColor(buffer.NamedRed))
	c.SetCell(10, 0, 'X', buffer.NamedColor(buffer.NamedRed))
	c.SetCell(0, 5, 'X', buffer.NamedColor(buffer.NamedRed))
	// Should not panic
}

func TestCanvas_SetCellBG(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	fg := buffer.NamedColor(buffer.NamedWhite)
	bg := buffer.NamedColor(buffer.NamedBlue)
	c.SetCellBG(3, 4, '#', fg, bg)
	cell := c.ActiveLayer().Cells[4][3]
	if cell.Char != '#' {
		t.Errorf("expected '#', got %q", string(cell.Char))
	}
}

func TestCanvas_DrawLine_Horizontal(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.DrawLine(0, 0, 9, 0, '-', buffer.NamedColor(buffer.NamedWhite))
	layer := c.ActiveLayer()
	for x := 0; x < 10; x++ {
		if !layer.Cells[0][x].Used {
			t.Errorf("expected cell (%d,0) to be drawn", x)
		}
		if layer.Cells[0][x].Char != '-' {
			t.Errorf("expected '-' at (%d,0), got %q", x, string(layer.Cells[0][x].Char))
		}
	}
}

func TestCanvas_DrawLine_Vertical(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.DrawLine(0, 0, 0, 9, '|', buffer.NamedColor(buffer.NamedWhite))
	layer := c.ActiveLayer()
	for y := 0; y < 10; y++ {
		if !layer.Cells[y][0].Used {
			t.Errorf("expected cell (0,%d) to be drawn", y)
		}
	}
}

func TestCanvas_DrawLine_Diagonal(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.DrawLine(0, 0, 9, 9, '\\', buffer.NamedColor(buffer.NamedWhite))
	layer := c.ActiveLayer()
	for i := 0; i < 10; i++ {
		if !layer.Cells[i][i].Used {
			t.Errorf("expected diagonal cell (%d,%d) to be drawn", i, i)
		}
	}
}

func TestCanvas_DrawLine_Single(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.DrawLine(5, 5, 5, 5, '*', buffer.NamedColor(buffer.NamedYellow))
	if !c.ActiveLayer().Cells[5][5].Used {
		t.Error("expected single point to be drawn")
	}
}

func TestCanvas_DrawRect(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	c.DrawRect(2, 2, 6, 4, '#', buffer.NamedColor(buffer.NamedGreen))
	layer := c.ActiveLayer()

	// Corners should be set
	if !layer.Cells[2][2].Used {
		t.Error("expected top-left corner")
	}
	if !layer.Cells[5][7].Used {
		t.Error("expected bottom-right corner")
	}

	// Interior should NOT be set (unfilled)
	if layer.Cells[3][3].Used {
		t.Error("interior should not be drawn for unfilled rect")
	}
}

func TestCanvas_FillRect(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	c.FillRect(2, 2, 5, 3, '@', buffer.NamedColor(buffer.NamedBlue))
	layer := c.ActiveLayer()

	// All interior cells should be set
	for y := 2; y < 5; y++ {
		for x := 2; x < 7; x++ {
			if !layer.Cells[y][x].Used {
				t.Errorf("expected cell (%d,%d) filled", x, y)
			}
		}
	}

	// Outside should not
	if layer.Cells[0][0].Used {
		t.Error("cell (0,0) should not be set")
	}
}

func TestCanvas_FillRect_Invalid(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	c.FillRect(5, 5, 0, 0, '@', buffer.NamedColor(buffer.NamedBlue))
	c.FillRect(5, 5, -1, 3, '@', buffer.NamedColor(buffer.NamedBlue))
	// Should not panic, nothing drawn
}

func TestCanvas_DrawCircle(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})

	c.DrawCircle(15, 10, 5, 'o', buffer.NamedColor(buffer.NamedCyan))
	layer := c.ActiveLayer()

	// Center should NOT be set (unfilled)
	if layer.Cells[10][15].Used {
		t.Error("center should not be drawn for unfilled circle")
	}

	// Points on the circle should be set
	// At angle 0 (rightmost point): (20, 10)
	if !layer.Cells[10][20].Used {
		t.Error("expected rightmost point of circle")
	}
	// At angle 90 (top): (15, 5)
	if !layer.Cells[5][15].Used {
		t.Error("expected top point of circle")
	}
}

func TestCanvas_FillCircle(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})

	c.FillCircle(15, 10, 4, '#', buffer.NamedColor(buffer.NamedGreen))
	layer := c.ActiveLayer()

	// Center should be filled
	if !layer.Cells[10][15].Used {
		t.Error("center should be filled")
	}
}

func TestCanvas_DrawCircle_RadiusZero(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})

	c.DrawCircle(10, 10, 0, 'o', buffer.NamedColor(buffer.NamedWhite))
	// Should not panic
}

func TestCanvas_Print(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	c.Print(2, 3, "Hello", buffer.NamedColor(buffer.NamedYellow))
	layer := c.ActiveLayer()

	expected := "Hello"
	for i, r := range expected {
		if layer.Cells[3][2+i].Char != r {
			t.Errorf("expected %q at (%d,3), got %q", string(r), 2+i, string(layer.Cells[3][2+i].Char))
		}
	}
}

func TestCanvas_Print_LongerThanWidth(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 3})

	c.Print(0, 0, "ABCDEFGHIJ", buffer.NamedColor(buffer.NamedWhite))
	layer := c.ActiveLayer()

	// Only first 5 chars should fit
	if layer.Cells[0][4].Char != 'E' {
		t.Errorf("expected 'E' at (4,0), got %q", string(layer.Cells[0][4].Char))
	}
}

func TestCanvas_Measure(t *testing.T) {
	c := NewCanvas()
	sz := c.Measure(Bounded(60, 15))
	if sz.W != 60 || sz.H != 15 {
		t.Errorf("expected 60x15, got %dx%d", sz.W, sz.H)
	}
}

func TestCanvas_Measure_Defaults(t *testing.T) {
	c := NewCanvas()
	sz := c.Measure(Unbounded())
	if sz.W != 80 || sz.H != 24 {
		t.Errorf("expected default 80x24, got %dx%d", sz.W, sz.H)
	}
}

func TestCanvas_SetBounds_ResizesLayer(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	c.SetCell(5, 5, 'X', buffer.NamedColor(buffer.NamedRed))

	c.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 20})
	layer := c.ActiveLayer()
	if layer.Width != 60 || layer.Height != 20 {
		t.Errorf("expected layer resized to 60x20, got %dx%d", layer.Width, layer.Height)
	}
	// Content should be preserved
	if !layer.Cells[5][5].Used {
		t.Error("expected original content preserved after resize")
	}
}

func TestCanvas_SetBounds_ShrinksLayer(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	c.SetCell(5, 5, 'X', buffer.NamedColor(buffer.NamedRed))

	c.SetBounds(Rect{X: 0, Y: 0, W: 3, H: 3})
	layer := c.ActiveLayer()
	if layer.Width != 3 || layer.Height != 3 {
		t.Errorf("expected layer shrunk to 3x3, got %dx%d", layer.Width, layer.Height)
	}
}

func TestCanvas_Paint_Basic(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 10})

	c.SetCell(5, 5, 'X', buffer.NamedColor(buffer.NamedRed))
	c.DrawLine(0, 0, 19, 0, '-', buffer.NamedColor(buffer.NamedWhite))

	buf := buffer.NewBuffer(20, 10)
	c.Paint(buf)

	// Check cell at (5,5)
	cell := buf.Cells[5*20+5]
	if cell.Rune != 'X' {
		t.Errorf("expected 'X' at (5,5), got %q", string(cell.Rune))
	}
	// Check line
	cell = buf.Cells[0*20+0]
	if cell.Rune != '-' {
		t.Errorf("expected '-' at (0,0), got %q", string(cell.Rune))
	}
}

func TestCanvas_Paint_MultipleLayers(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	c.SetCell(0, 0, 'A', buffer.NamedColor(buffer.NamedRed))

	c.AddLayer("overlay")
	c.SetCell(0, 0, 'B', buffer.NamedColor(buffer.NamedBlue))

	buf := buffer.NewBuffer(10, 5)
	c.Paint(buf)

	// Top layer should overwrite bottom
	cell := buf.Cells[0]
	if cell.Rune != 'B' {
		t.Errorf("expected 'B' (top layer), got %q", string(cell.Rune))
	}
}

func TestCanvas_Paint_BoundsOffset(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 5, Y: 3, W: 20, H: 10})

	c.SetCell(0, 0, 'X', buffer.NamedColor(buffer.NamedRed))

	buf := buffer.NewBuffer(30, 15)
	c.Paint(buf)

	cell := buf.Cells[3*30+5]
	if cell.Rune != 'X' {
		t.Errorf("expected 'X' at offset (5,3), got %q", string(cell.Rune))
	}
}

func TestCanvas_Paint_ZeroBounds_NoOp(t *testing.T) {
	c := NewCanvas()
	// Don't set bounds (stays zero)
	buf := buffer.NewBuffer(80, 24)
	c.Paint(buf) // should not panic
}

func TestCanvas_Paint_UnusedCellsSkipped(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	c.SetCell(0, 0, 'X', buffer.NamedColor(buffer.NamedRed))
	// Only 1 cell set, rest should be blank

	buf := buffer.NewBuffer(10, 5)
	c.Paint(buf)

	// Cell (1,0) should be blank (space, not 'X')
	cell := buf.Cells[1]
	if cell.Rune == 'X' {
		t.Errorf("expected blank cell at (1,0), got %q", string(cell.Rune))
	}
}

func TestCanvas_SetWorldBounds(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 100, H: 100})
	c.SetWorldBounds(0, 100, 0, 100)

	// World (0,100) -> top-left grid (0,0)
	gx, gy := c.WorldToGrid(0, 100)
	if gx != 0 || gy != 0 {
		t.Errorf("expected (0,0), got (%d,%d)", gx, gy)
	}

	// World (100,0) -> bottom-right grid (100,100)
	gx, gy = c.WorldToGrid(100, 0)
	if gx != 100 || gy != 100 {
		t.Errorf("expected (100,100), got (%d,%d)", gx, gy)
	}
}

func TestCanvas_WorldToGrid_FlipY(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 10})
	c.SetWorldBounds(0, 10, 0, 10)

	// World Y=10 (top) should map to grid Y=0 (top row)
	_, gy := c.WorldToGrid(5, 10)
	if gy != 0 {
		t.Errorf("expected Y-flip: world y=10 -> grid y=0, got %d", gy)
	}

	// World Y=0 (bottom) should map to grid Y=10 (bottom row)
	_, gy = c.WorldToGrid(5, 0)
	if gy != 10 {
		t.Errorf("expected Y-flip: world y=0 -> grid y=10, got %d", gy)
	}
}

func TestCanvas_WorldToGrid_DefaultBounds(t *testing.T) {
	c := NewCanvas()
	// Default world bounds: 0..1, 0..1
	gx, gy := c.WorldToGrid(0.5, 0.5)
	// Should be roughly center
	if gx < 0 || gy < 0 {
		t.Errorf("expected non-negative grid coords, got (%d,%d)", gx, gy)
	}
}

func TestCanvas_DrawLineWorld(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	c.SetWorldBounds(0, 10, 0, 10)

	c.DrawLineWorld(0, 0, 10, 10, '*', buffer.NamedColor(buffer.NamedRed))
	layer := c.ActiveLayer()

	// At least something should be drawn
	drawn := 0
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			if layer.Cells[y][x].Used {
				drawn++
			}
		}
	}
	if drawn == 0 {
		t.Error("expected some cells to be drawn for world line")
	}
}

func TestCanvas_DrawCircleWorld(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})
	c.SetWorldBounds(0, 10, 0, 10)

	c.DrawCircleWorld(5, 5, 3, 'o', buffer.NamedColor(buffer.NamedCyan))
	// Should not panic, should draw something
	layer := c.ActiveLayer()
	drawn := 0
	for y := 0; y < 20; y++ {
		for x := 0; x < 40; x++ {
			if layer.Cells[y][x].Used {
				drawn++
			}
		}
	}
	if drawn == 0 {
		t.Error("expected some cells drawn for world circle")
	}
}

func TestCanvas_Concurrent(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})

	var wg sync.WaitGroup

	// Concurrent writers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.SetCell(j%40, n*4, 'X', buffer.NamedColor(buffer.NamedRed))
				c.DrawLine(0, n, 39, n, '-', buffer.NamedColor(buffer.NamedWhite))
			}
		}(i)
	}

	// Concurrent readers
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				c.ActiveLayer()
				c.Layers()
				buf := buffer.NewBuffer(40, 20)
				c.Paint(buf)
			}
		}()
	}

	wg.Wait()
}

func TestCanvas_Children(t *testing.T) {
	c := NewCanvas()
	if c.Children() != nil {
		t.Error("expected nil children for leaf component")
	}
}

func TestCanvas_Bounds(t *testing.T) {
	c := NewCanvas()
	r := Rect{X: 1, Y: 2, W: 50, H: 20}
	c.SetBounds(r)
	if c.Bounds() != r {
		t.Errorf("expected %+v, got %+v", r, c.Bounds())
	}
}

func TestCanvas_LayerZOrder(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})

	// Default layer draws 'A'
	c.SetCell(0, 0, 'A', buffer.NamedColor(buffer.NamedRed))

	// Add second layer draws 'B' at same position
	c.AddLayer("layer2")
	c.SetCell(0, 0, 'B', buffer.NamedColor(buffer.NamedBlue))

	// Add third layer draws 'C' at same position
	c.AddLayer("layer3")
	c.SetCell(0, 0, 'C', buffer.NamedColor(buffer.NamedGreen))

	buf := buffer.NewBuffer(10, 5)
	c.Paint(buf)

	// Top-most layer should win
	cell := buf.Cells[0]
	if cell.Rune != 'C' {
		t.Errorf("expected 'C' (top layer), got %q", string(cell.Rune))
	}
}

func TestCanvas_DrawLine_NegativeDirection(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})

	// Line from (9,0) to (0,0) — reverse direction
	c.DrawLine(9, 0, 0, 0, '-', buffer.NamedColor(buffer.NamedWhite))
	layer := c.ActiveLayer()
	for x := 0; x < 10; x++ {
		if !layer.Cells[0][x].Used {
			t.Errorf("expected cell (%d,0) drawn in reverse line", x)
		}
	}
}

func TestCanvas_DrawLine_Steep(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 20})

	// Steep line: more vertical than horizontal
	c.DrawLine(0, 0, 3, 9, '|', buffer.NamedColor(buffer.NamedWhite))
	layer := c.ActiveLayer()

	// All 10 y-rows should have at least one cell drawn
	for y := 0; y < 10; y++ {
		found := false
		for x := 0; x < 4; x++ {
			if layer.Cells[y][x].Used {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected at least one cell in row %d for steep line", y)
		}
	}
}

func TestCanvas_FillCircle_SmallRadius(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 20})

	c.FillCircle(10, 10, 1, '#', buffer.NamedColor(buffer.NamedGreen))
	layer := c.ActiveLayer()

	// Radius 1: center + immediate neighbors should be filled
	if !layer.Cells[10][10].Used {
		t.Error("center should be filled for r=1")
	}
}
