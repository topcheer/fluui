package component

import (
	"math"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// CanvasCell represents a single drawable cell on the canvas grid.
type CanvasCell struct {
	Char rune
	Fg   buffer.Color
	Bg   buffer.Color
	Used bool // true if this cell has been drawn to
}

// CanvasLayer is a named drawing layer with z-order.
type CanvasLayer struct {
	Name   string
	Cells  [][]CanvasCell
	Width  int
	Height int
}

// NewCanvasLayer creates a blank layer of the given dimensions.
func NewCanvasLayer(name string, w, h int) *CanvasLayer {
	cells := make([][]CanvasCell, h)
	for y := 0; y < h; y++ {
		cells[y] = make([]CanvasCell, w)
	}
	return &CanvasLayer{
		Name:   name,
		Cells:  cells,
		Width:  w,
		Height: h,
	}
}

// Canvas is a free-form 2D drawing surface that maps to the terminal cell grid.
// It supports layers, lines (Bresenham), circles (midpoint algorithm),
// rectangles, filled shapes, and text placement.
//
// This component is inspired by Ratatui's Canvas and enables arbitrary
// visualizations including charts, diagrams, ASCII art, and simple games.
type Canvas struct {
	BaseComponent
	mu sync.RWMutex

	layers  []*CanvasLayer
	active  *CanvasLayer // currently active drawing layer
	bgColor buffer.Color

	// world coordinate bounds for chart-like usage
	xMin, xMax float64
	yMin, yMax float64

	bounds Rect
}

// NewCanvas creates a Canvas with a single default layer.
func NewCanvas() *Canvas {
	c := &Canvas{
		bgColor: buffer.Color{},
		xMin:    0,
		xMax:    1,
		yMin:    0,
		yMax:    1,
	}
	c.AddLayer("default")
	return c
}

// AddLayer creates and adds a new layer, making it the active layer.
func (c *Canvas) AddLayer(name string) *CanvasLayer {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Use bounds dimensions if available, otherwise default 80x24
	w, h := 80, 24
	if c.bounds.W > 0 {
		w = c.bounds.W
	}
	if c.bounds.H > 0 {
		h = c.bounds.H
	}

	layer := NewCanvasLayer(name, w, h)
	c.layers = append(c.layers, layer)
	c.active = layer
	return layer
}

// SetActiveLayer makes the named layer the active drawing target.
func (c *Canvas) SetActiveLayer(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, l := range c.layers {
		if l.Name == name {
			c.active = l
			return true
		}
	}
	return false
}

// ActiveLayer returns the currently active layer.
func (c *Canvas) ActiveLayer() *CanvasLayer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.active
}

// Layers returns all layers.
func (c *Canvas) Layers() []*CanvasLayer {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*CanvasLayer, len(c.layers))
	copy(out, c.layers)
	return out
}

// RemoveLayer removes the named layer. The default layer cannot be removed.
func (c *Canvas) RemoveLayer(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, l := range c.layers {
		if l.Name == name && name != "default" {
			c.layers = append(c.layers[:i], c.layers[i+1:]...)
			if c.active == l {
				c.active = c.layers[0]
			}
			return true
		}
	}
	return false
}

// ClearLayer clears all cells in the named layer.
func (c *Canvas) ClearLayer(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, l := range c.layers {
		if l.Name == name {
			for y := 0; y < l.Height; y++ {
				for x := 0; x < l.Width; x++ {
					l.Cells[y][x] = CanvasCell{}
				}
			}
			return
		}
	}
}

// ClearAll clears all layers.
func (c *Canvas) ClearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, l := range c.layers {
		for y := 0; y < l.Height; y++ {
			for x := 0; x < l.Width; x++ {
				l.Cells[y][x] = CanvasCell{}
			}
		}
	}
}

// SetBackgroundColor sets the canvas background color.
func (c *Canvas) SetBackgroundColor(col buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bgColor = col
}

// SetWorldBounds sets the world coordinate bounds for chart-like mapping.
func (c *Canvas) SetWorldBounds(xMin, xMax, yMin, yMax float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.xMin = xMin
	c.xMax = xMax
	c.yMin = yMin
	c.yMax = yMax
}

// WorldToGrid converts world coordinates to grid cell coordinates.
func (c *Canvas) WorldToGrid(x, y float64) (int, int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	w, h := 80, 24
	if c.bounds.W > 0 {
		w = c.bounds.W
	}
	if c.bounds.H > 0 {
		h = c.bounds.H
	}

	xRange := c.xMax - c.xMin
	yRange := c.yMax - c.yMin
	if xRange <= 0 {
		xRange = 1
	}
	if yRange <= 0 {
		yRange = 1
	}

	gx := int((x - c.xMin) / xRange * float64(w))
	gy := int((c.yMax - y) / yRange * float64(h)) // flip Y

	return gx, gy
}

// --- Drawing primitives on the active layer ---

// SetCell draws a single cell on the active layer.
func (c *Canvas) SetCell(x, y int, char rune, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil {
		return
	}
	c.setCellLocked(x, y, char, fg, c.bgColor)
}

// SetCellBG draws a single cell with explicit background on the active layer.
func (c *Canvas) SetCellBG(x, y int, char rune, fg, bg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil {
		return
	}
	c.setCellLocked(x, y, char, fg, bg)
}

func (c *Canvas) setCellLocked(x, y int, char rune, fg, bg buffer.Color) {
	l := c.active
	if l == nil || x < 0 || x >= l.Width || y < 0 || y >= l.Height {
		return
	}
	l.Cells[y][x] = CanvasCell{Char: char, Fg: fg, Bg: bg, Used: true}
}

// DrawLine draws a line from (x1,y1) to (x2,y2) using Bresenham's algorithm.
func (c *Canvas) DrawLine(x1, y1, x2, y2 int, char rune, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil {
		return
	}

	dx := x2 - x1
	if dx < 0 {
		dx = -dx
	}
	dy := y2 - y1
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x1 >= x2 {
		sx = -1
	}
	sy := 1
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy
	x, y := x1, y1

	for {
		c.setCellLocked(x, y, char, fg, c.bgColor)
		if x == x2 && y == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// DrawLineWorld draws a line using world coordinates.
func (c *Canvas) DrawLineWorld(x1, y1, x2, y2 float64, char rune, fg buffer.Color) {
	gx1, gy1 := c.WorldToGrid(x1, y1)
	gx2, gy2 := c.WorldToGrid(x2, y2)
	c.DrawLine(gx1, gy1, gx2, gy2, char, fg)
}

// DrawRect draws an unfilled rectangle outline.
func (c *Canvas) DrawRect(x, y, w, h int, char rune, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil || w <= 0 || h <= 0 {
		return
	}

	// Top and bottom edges
	for i := x; i < x+w; i++ {
		c.setCellLocked(i, y, char, fg, c.bgColor)
		c.setCellLocked(i, y+h-1, char, fg, c.bgColor)
	}
	// Left and right edges
	for j := y; j < y+h; j++ {
		c.setCellLocked(x, j, char, fg, c.bgColor)
		c.setCellLocked(x+w-1, j, char, fg, c.bgColor)
	}
}

// FillRect draws a filled rectangle.
func (c *Canvas) FillRect(x, y, w, h int, char rune, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil || w <= 0 || h <= 0 {
		return
	}
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.setCellLocked(i, j, char, fg, c.bgColor)
		}
	}
}

// DrawCircle draws an unfilled circle using the midpoint algorithm.
func (c *Canvas) DrawCircle(cx, cy, r int, char rune, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil || r <= 0 {
		return
	}

	x := r
	y := 0
	err := 0

	for x >= y {
		c.setCellLocked(cx+x, cy+y, char, fg, c.bgColor)
		c.setCellLocked(cx+y, cy+x, char, fg, c.bgColor)
		c.setCellLocked(cx-y, cy+x, char, fg, c.bgColor)
		c.setCellLocked(cx-x, cy+y, char, fg, c.bgColor)
		c.setCellLocked(cx-x, cy-y, char, fg, c.bgColor)
		c.setCellLocked(cx-y, cy-x, char, fg, c.bgColor)
		c.setCellLocked(cx+y, cy-x, char, fg, c.bgColor)
		c.setCellLocked(cx+x, cy-y, char, fg, c.bgColor)

		y++
		if err <= 0 {
			err += 2*y + 1
		}
		if err > 0 {
			x--
			err -= 2*x + 1
		}
	}
}

// FillCircle draws a filled circle.
func (c *Canvas) FillCircle(cx, cy, r int, char rune, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil || r <= 0 {
		return
	}

	for dy := -r; dy <= r; dy++ {
		dx := int(math.Sqrt(float64(r*r - dy*dy)))
		for i := cx - dx; i <= cx+dx; i++ {
			c.setCellLocked(i, cy+dy, char, fg, c.bgColor)
		}
	}
}

// DrawCircleWorld draws a circle using world coordinates.
func (c *Canvas) DrawCircleWorld(cx, cy, r float64, char rune, fg buffer.Color) {
	gx, gy := c.WorldToGrid(cx, cy)
	// Convert radius: average of x and y pixel radius
	_, gridR := c.WorldToGrid(cx+r, cy)
	gridRadius := gridR - gy
	if gridRadius < 0 {
		gridRadius = -gridRadius
	}
	if gridRadius < 1 {
		gridRadius = 1
	}
	c.DrawCircle(gx, gy, gridRadius, char, fg)
}

// Print draws a text string starting at (x, y).
func (c *Canvas) Print(x, y int, text string, fg buffer.Color) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil {
		return
	}
	runes := []rune(text)
	for i, r := range runes {
		c.setCellLocked(x+i, y, r, fg, c.bgColor)
	}
}

// --- Component interface ---

// Measure returns the desired size for the canvas.
func (c *Canvas) Measure(cs Constraints) Size {
	w := 80
	h := 24
	if cs.HasWidth() && cs.MaxWidth > 0 {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && cs.MaxHeight > 0 {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// SetBounds sets the canvas position and size, and resizes the active layer.
func (c *Canvas) SetBounds(r Rect) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bounds = r
	// Resize all layers to match new bounds
	for _, l := range c.layers {
		if l.Width != r.W || l.Height != r.H {
			c.resizeLayerLocked(l, r.W, r.H)
		}
	}
}

// Bounds returns the current bounds.
func (c *Canvas) Bounds() Rect {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.bounds
}

func (c *Canvas) resizeLayerLocked(l *CanvasLayer, w, h int) {
	newCells := make([][]CanvasCell, h)
	for y := 0; y < h; y++ {
		newCells[y] = make([]CanvasCell, w)
		// Copy existing content
		if y < l.Height {
			copyCount := l.Width
			if copyCount > w {
				copyCount = w
			}
			copy(newCells[y], l.Cells[y][:copyCount])
		}
	}
	l.Cells = newCells
	l.Width = w
	l.Height = h
}

// Paint renders all layers into the buffer, composited in z-order.
func (c *Canvas) Paint(buf *buffer.Buffer) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	bounds := c.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	// Composite layers bottom-to-top
	for _, l := range c.layers {
		for y := 0; y < l.Height && y < bounds.H; y++ {
			for x := 0; x < l.Width && x < bounds.W; x++ {
				cell := l.Cells[y][x]
				if !cell.Used {
					continue
				}
				px := bounds.X + x
				py := bounds.Y + y
				if px >= 0 && py >= 0 && px < buf.Width && py < buf.Height {
					buf.SetCell(px, py, buffer.NewCell(cell.Char, buffer.Style{
						Fg: cell.Fg,
						Bg: cell.Bg,
					}))
				}
			}
		}
	}
}

// --- helpers ---

