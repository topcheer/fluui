package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// DrawerSide specifies which edge the drawer slides from.
type DrawerSide int

const (
	DrawerLeft  DrawerSide = iota
	DrawerRight
)

// Drawer is a slide-in side panel that overlays content. Common for
// settings panels, navigation menus, filter controls, and context details.
// The drawer can be open or closed; when open it renders a bordered panel
// with a title and optional child content.
//
// Thread-safe.
type Drawer struct {
	BaseComponent
	mu     sync.Mutex
	side   DrawerSide
	title  string
	open   bool
	width  int
}

// NewDrawer creates a drawer with the given side and title.
func NewDrawer(side DrawerSide, title string) *Drawer {
	return &Drawer{
		BaseComponent: BaseComponent{id: GenerateID("drawer")},
		side:          side,
		title:         title,
		width:         30,
	}
}

// IsOpen returns whether the drawer is visible.
func (d *Drawer) IsOpen() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.open
}

// Open shows the drawer.
func (d *Drawer) Open() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.open = true
}

// Close hides the drawer.
func (d *Drawer) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.open = false
}

// Toggle flips the open state.
func (d *Drawer) Toggle() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.open = !d.open
}

// SetWidth sets the drawer width in columns.
func (d *Drawer) SetWidth(w int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if w < 5 {
		w = 5
	}
	d.width = w
}

// Width returns the current drawer width.
func (d *Drawer) Width() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.width
}

// SetTitle changes the drawer title.
func (d *Drawer) SetTitle(s string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.title = s
}

// Title returns the current title.
func (d *Drawer) Title() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.title
}

// SetSide changes which edge the drawer attaches to.
func (d *Drawer) SetSide(s DrawerSide) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.side = s
}

// Side returns the current drawer side.
func (d *Drawer) Side() DrawerSide {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.side
}

// Measure returns the desired size (0x0 when closed).
func (d *Drawer) Measure(cs Constraints) Size {
	d.mu.Lock()
	open := d.open
	w := d.width
	d.mu.Unlock()
	if !open {
		return Size{W: 0, H: 0}
	}
	maxH := cs.MaxHeight
	if maxH <= 0 {
		maxH = 20
	}
	return Size{W: w, H: maxH}
}

// Paint renders the drawer if open.
func (d *Drawer) Paint(buf *buffer.Buffer) {
	d.mu.Lock()
	open := d.open
	side := d.side
	title := d.title
	dw := d.width
	d.mu.Unlock()

	if !open {
		return
	}

	parent := d.Bounds()
	if parent.W <= 0 || parent.H <= 0 {
		return
	}

	th := theme.Get()
	borderStyle := buffer.Style{Fg: th.Border}
	titleStyle := buffer.Style{Fg: th.Accent}
	bgStyle := buffer.Style{Fg: th.Fg, Bg: th.CodeBg}

	// Position the drawer at left or right edge
	var b Rect
	if side == DrawerLeft {
		b = Rect{X: parent.X, Y: parent.Y, W: dw, H: parent.H}
	} else {
		b = Rect{X: parent.X + parent.W - dw, Y: parent.Y, W: dw, H: parent.H}
	}

	// Clamp width
	if b.W > parent.W {
		b.W = parent.W
	}

	// Fill background
	for y := b.Y; y < b.Y+b.H; y++ {
		for x := b.X; x < b.X+b.W; x++ {
			buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Bg: th.CodeBg})
		}
	}

	// Top border with title
	titleW := utf8.RuneCountInString(title)
	availTitleW := b.W - 4
	if titleW > availTitleW && availTitleW > 2 {
		title = truncateRunes(title, availTitleW-1) + "\u2026"
		titleW = utf8.RuneCountInString(title)
	}

	// Draw top border: ╭ title ───╮
	buf.DrawText(b.X, b.Y, "\u256d", borderStyle) // ╭
	if title != "" {
		buf.DrawText(b.X+1, b.Y, " "+title+" ", titleStyle)
	}
	// Right corner
	buf.DrawText(b.X+b.W-1, b.Y, "\u256e", borderStyle) // ╮

	// Side borders
	for y := b.Y + 1; y < b.Y+b.H-1; y++ {
		buf.DrawText(b.X, y, "\u2502", borderStyle)     // │
		buf.DrawText(b.X+b.W-1, y, "\u2502", borderStyle) // │
	}

	// Bottom border: ╰────────╯
	buf.DrawText(b.X, b.Y+b.H-1, "\u2570", borderStyle) // ╰
	buf.DrawText(b.X+b.W-1, b.Y+b.H-1, "\u256f", borderStyle) // ╯

	// Fill top border gaps with dashes
	for x := b.X + 1 + titleW + 2; x < b.X+b.W-1; x++ {
		buf.DrawText(x, b.Y, "\u2500", borderStyle)
	}
	for x := b.X + 1; x < b.X+b.W-1; x++ {
		buf.DrawText(x, b.Y+b.H-1, "\u2500", borderStyle)
	}

	_ = bgStyle
}
