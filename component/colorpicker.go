package component

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// ColorPickerMode specifies the current selection mode.
type ColorPickerMode int

const (
	// PickerPalette shows a 16x16 grid of 256-color palette colors.
	PickerPalette ColorPickerMode = iota
	// PickerRGB shows three RGB sliders for true-color selection.
	PickerRGB
	// PickerHex shows a hex code input field.
	PickerHex
)

// pickerModeCount is the number of modes for cycling.
const pickerModeCount = 3

// ColorPickerStyle holds visual styles for the ColorPicker.
type ColorPickerStyle struct {
	Title     buffer.Style
	Label     buffer.Style
	Value     buffer.Style
	Box       buffer.Style
	Cursor    buffer.Style
	SwatchBorder buffer.Style
}

// DefaultColorPickerStyle returns a sensible default style.
func DefaultColorPickerStyle() ColorPickerStyle {
	return ColorPickerStyle{
		Title:        buffer.Style{Fg: buffer.White, Flags: buffer.Bold},
		Label:        buffer.Style{Fg: buffer.White},
		Value:        buffer.Style{Fg: buffer.Cyan, Flags: buffer.Bold},
		Box:          buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Cursor:       buffer.Style{Fg: buffer.Yellow, Flags: buffer.Bold | buffer.Reverse},
		SwatchBorder: buffer.Style{Fg: buffer.RGB(80, 80, 80)},
	}
}

// ColorPicker is an interactive color selection component with three modes:
// 256-color palette grid, RGB sliders, and hex code input.
//
// It supports keyboard navigation (arrows, Tab to switch modes, Enter to confirm)
// and fires an OnChange callback whenever the selected color changes.
type ColorPicker struct {
	BaseComponent
	mu     sync.RWMutex
	mode   ColorPickerMode
	color  buffer.Color
	style  ColorPickerStyle

	// Palette cursor position (0-15 grid, but 0-255 linearly)
	paletteIdx int

	// RGB values (0-255 each)
	r, g, b uint8

	// Active RGB channel (0=R, 1=G, 2=B) for slider adjustment
	activeChannel int

	// Hex input buffer
	hexBuf    []rune
	hexCursor int

	// OnChange fires whenever the selected color changes.
	OnChange func(c buffer.Color)
	// OnConfirm fires when Enter is pressed.
	OnConfirm func(c buffer.Color)
}

// NewColorPicker creates a color picker starting in palette mode with color #000000.
func NewColorPicker() *ColorPicker {
	return &ColorPicker{
		mode:   PickerPalette,
		color:  buffer.RGB(0, 0, 0),
		style:  DefaultColorPickerStyle(),
		hexBuf: []rune("000000"),
	}
}

// SetMode switches the selection mode.
func (cp *ColorPicker) SetMode(mode ColorPickerMode) {
	cp.mu.Lock()
	cp.mode = mode
	cp.mu.Unlock()
}

// Mode returns the current selection mode.
func (cp *ColorPicker) Mode() ColorPickerMode {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.mode
}

// Color returns the currently selected color.
func (cp *ColorPicker) Color() buffer.Color {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.color
}

// SetColor sets the selected color directly.
func (cp *ColorPicker) SetColor(c buffer.Color) {
	cp.mu.Lock()
	cp.color = c
	if c.Type == buffer.ColorTrue {
		cp.r = c.R()
		cp.g = c.G()
		cp.b = c.B()
		cp.hexBuf = []rune(fmt.Sprintf("%02x%02x%02x", cp.r, cp.g, cp.b))
	}
	cp.mu.Unlock()
	cp.fireChange()
}

// SetStyle configures the visual style.
func (cp *ColorPicker) SetStyle(s ColorPickerStyle) {
	cp.mu.Lock()
	cp.style = s
	cp.mu.Unlock()
}

// Style returns the current visual style.
func (cp *ColorPicker) Style() ColorPickerStyle {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.style
}

// NextMode cycles to the next mode (palette → RGB → hex → palette).
func (cp *ColorPicker) NextMode() {
	cp.mu.Lock()
	cp.mode = (cp.mode + 1) % pickerModeCount
	cp.mu.Unlock()
}

// PrevMode cycles to the previous mode.
func (cp *ColorPicker) PrevMode() {
	cp.mu.Lock()
	cp.mode = (cp.mode - 1 + pickerModeCount) % pickerModeCount
	cp.mu.Unlock()
}

// PaletteIndex returns the current palette cursor index (0-255).
func (cp *ColorPicker) PaletteIndex() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.paletteIdx
}

// SetPaletteIndex sets the palette cursor position, clamped to [0, 255].
func (cp *ColorPicker) SetPaletteIndex(idx int) {
	cp.mu.Lock()
	if idx < 0 {
		idx = 0
	}
	if idx > 255 {
		idx = 255
	}
	cp.paletteIdx = idx
	cp.color = buffer.Color256Val(uint8(idx))
	cp.mu.Unlock()
	cp.fireChange()
}

// RGBValues returns the current R, G, B values.
func (cp *ColorPicker) RGBValues() (r, g, b uint8) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.r, cp.g, cp.b
}

// SetRGB sets the RGB values directly.
func (cp *ColorPicker) SetRGB(r, g, b uint8) {
	cp.mu.Lock()
	cp.r = r
	cp.g = g
	cp.b = b
	cp.color = buffer.RGB(r, g, b)
	cp.hexBuf = []rune(fmt.Sprintf("%02x%02x%02x", r, g, b))
	cp.mu.Unlock()
	cp.fireChange()
}

// ActiveChannel returns the current active RGB channel (0=R, 1=G, 2=B).
func (cp *ColorPicker) ActiveChannel() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return cp.activeChannel
}

// SetActiveChannel sets the active RGB channel.
func (cp *ColorPicker) SetActiveChannel(ch int) {
	cp.mu.Lock()
	if ch < 0 {
		ch = 0
	}
	if ch > 2 {
		ch = 2
	}
	cp.activeChannel = ch
	cp.mu.Unlock()
}

// HexString returns the current color as a hex string.
func (cp *ColorPicker) HexString() string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	return "#" + string(cp.hexBuf)
}

// HandleKey processes keyboard input.
func (cp *ColorPicker) HandleKey(k *term.KeyEvent) bool {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	switch cp.mode {
	case PickerPalette:
		return cp.handlePaletteKey(k)
	case PickerRGB:
		return cp.handleRGBKey(k)
	case PickerHex:
		return cp.handleHexKey(k)
	}
	return false
}

func (cp *ColorPicker) handlePaletteKey(k *term.KeyEvent) bool {
	switch k.Key {
	case term.KeyLeft:
		if cp.paletteIdx > 0 {
			cp.paletteIdx--
			cp.color = buffer.Color256Val(uint8(cp.paletteIdx))
			cp.fireChangeLocked()
		}
		return true
	case term.KeyRight:
		if cp.paletteIdx < 255 {
			cp.paletteIdx++
			cp.color = buffer.Color256Val(uint8(cp.paletteIdx))
			cp.fireChangeLocked()
		}
		return true
	case term.KeyUp:
		if cp.paletteIdx >= 16 {
			cp.paletteIdx -= 16
			cp.color = buffer.Color256Val(uint8(cp.paletteIdx))
			cp.fireChangeLocked()
		}
		return true
	case term.KeyDown:
		if cp.paletteIdx <= 239 {
			cp.paletteIdx += 16
			cp.color = buffer.Color256Val(uint8(cp.paletteIdx))
			cp.fireChangeLocked()
		}
		return true
	case term.KeyHome:
		cp.paletteIdx = 0
		cp.color = buffer.Color256Val(0)
		cp.fireChangeLocked()
		return true
	case term.KeyEnd:
		cp.paletteIdx = 255
		cp.color = buffer.Color256Val(255)
		cp.fireChangeLocked()
		return true
	case term.KeyEnter:
		cb := cp.OnConfirm
		if cb != nil {
			cb(cp.color)
		}
		return true
	}
	return false
}

func (cp *ColorPicker) handleRGBKey(k *term.KeyEvent) bool {
	switch k.Key {
	case term.KeyLeft:
		if cp.activeChannel > 0 {
			cp.activeChannel--
		}
		return true
	case term.KeyRight:
		if cp.activeChannel < 2 {
			cp.activeChannel++
		}
		return true
	case term.KeyUp:
		cp.adjustChannel(1)
		return true
	case term.KeyDown:
		cp.adjustChannel(-1)
		return true
	case term.KeyEnter:
		cb := cp.OnConfirm
		if cb != nil {
			cb(cp.color)
		}
		return true
	}

	// Vim-style: h/l for channel, j/k for value
	if k.Rune != 0 {
		switch k.Rune {
		case 'h':
			if cp.activeChannel > 0 {
				cp.activeChannel--
			}
			return true
		case 'l':
			if cp.activeChannel < 2 {
				cp.activeChannel++
			}
			return true
		case 'k':
			cp.adjustChannel(1)
			return true
		case 'j':
			cp.adjustChannel(-1)
			return true
		case 'H':
			cp.adjustChannel(10)
			return true
		case 'L':
			cp.adjustChannel(-10)
			return true
		}
	}
	return false
}

func (cp *ColorPicker) handleHexKey(k *term.KeyEvent) bool {
	switch k.Key {
	case term.KeyLeft:
		if cp.hexCursor > 0 {
			cp.hexCursor--
		}
		return true
	case term.KeyRight:
		if cp.hexCursor < len(cp.hexBuf) {
			cp.hexCursor++
		}
		return true
	case term.KeyHome:
		cp.hexCursor = 0
		return true
	case term.KeyEnd:
		cp.hexCursor = len(cp.hexBuf)
		return true
	case term.KeyBackspace:
		if cp.hexCursor > 0 {
			cp.hexCursor--
			cp.hexBuf = append(cp.hexBuf[:cp.hexCursor], cp.hexBuf[cp.hexCursor+1:]...)
			cp.updateColorFromHex()
		}
		return true
	case term.KeyEnter:
		cb := cp.OnConfirm
		if cb != nil {
			cb(cp.color)
		}
		return true
	}

	// Handle hex digit input (0-9, a-f, A-F)
	if k.Rune != 0 && cp.hexCursor < 6 {
		c := k.Rune
		if isHexDigit(c) {
			if cp.hexCursor < len(cp.hexBuf) {
				cp.hexBuf[cp.hexCursor] = c
			} else {
				cp.hexBuf = append(cp.hexBuf, c)
			}
			cp.hexCursor++
			cp.updateColorFromHex()
			return true
		}
	}
	return false
}

func (cp *ColorPicker) adjustChannel(delta int) {
	switch cp.activeChannel {
	case 0: // Red
		newVal := int(cp.r) + delta
		if newVal < 0 {
			newVal = 0
		}
		if newVal > 255 {
			newVal = 255
		}
		cp.r = uint8(newVal)
	case 1: // Green
		newVal := int(cp.g) + delta
		if newVal < 0 {
			newVal = 0
		}
		if newVal > 255 {
			newVal = 255
		}
		cp.g = uint8(newVal)
	case 2: // Blue
		newVal := int(cp.b) + delta
		if newVal < 0 {
			newVal = 0
		}
		if newVal > 255 {
			newVal = 255
		}
		cp.b = uint8(newVal)
	}
	cp.color = buffer.RGB(cp.r, cp.g, cp.b)
	cp.hexBuf = []rune(fmt.Sprintf("%02x%02x%02x", cp.r, cp.g, cp.b))
	cp.fireChangeLocked()
}

func (cp *ColorPicker) updateColorFromHex() {
	hexStr := string(cp.hexBuf)
	if len(hexStr) == 6 {
		c := buffer.Hex("#" + hexStr)
		if c.Type == buffer.ColorTrue {
			cp.color = c
			cp.r = c.R()
			cp.g = c.G()
			cp.b = c.B()
			cp.fireChangeLocked()
		}
	}
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func (cp *ColorPicker) fireChangeLocked() {
	if cp.OnChange != nil {
		cp.OnChange(cp.color)
	}
}

func (cp *ColorPicker) fireChange() {
	cp.mu.RLock()
	cb := cp.OnChange
	c := cp.color
	cp.mu.RUnlock()
	if cb != nil {
		cb(c)
	}
}

// Measure returns the desired size of the color picker.
func (cp *ColorPicker) Measure(cs Constraints) Size {
	w := 40
	h := 16
	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.HasHeight() && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// SetBounds sets the component's position and size.
func (cp *ColorPicker) SetBounds(r Rect) {
	cp.BaseComponent.SetBounds(r)
}

// Paint renders the color picker into the buffer.
func (cp *ColorPicker) Paint(buf *buffer.Buffer) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	bounds := cp.Bounds()
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	// Title line
	title := " Color Picker "
	cp.drawString(buf, bounds.X, bounds.Y, title, cp.style.Title)

	// Mode tabs
	modeNames := []string{"[1] Palette", "[2] RGB", "[3] Hex"}
	modeX := bounds.X + 20
	for i, name := range modeNames {
		style := cp.style.Label
		if i == int(cp.mode) {
			style = cp.style.Value
		}
		cp.drawString(buf, modeX, bounds.Y, name, style)
		modeX += len(name) + 2
	}

	// Swatch preview (right side)
	swatchX := bounds.X + bounds.W - 12
	cp.drawSwatch(buf, swatchX, bounds.Y, 10, 3)

	// Content area starts at y+2
	contentY := bounds.Y + 2

	switch cp.mode {
	case PickerPalette:
		cp.paintPalette(buf, bounds.X, contentY, bounds.W)
	case PickerRGB:
		cp.paintRGB(buf, bounds.X, contentY, bounds.W)
	case PickerHex:
		cp.paintHex(buf, bounds.X, contentY, bounds.W)
	}

	// Bottom: current value
	valY := bounds.Y + bounds.H - 2
	valStr := cp.color.String()
	cp.drawString(buf, bounds.X, valY, "Selected: ", cp.style.Label)
	cp.drawString(buf, bounds.X+10, valY, valStr, cp.style.Value)

	// Help line
	helpY := bounds.Y + bounds.H - 1
	helpText := "Tab: switch mode  Arrows: navigate  Enter: confirm  Esc: quit"
	cp.drawString(buf, bounds.X, helpY, helpText, cp.style.Box)
}

func (cp *ColorPicker) paintPalette(buf *buffer.Buffer, x, y, maxW int) {
	// 16x16 grid of 256 colors
	// First 16 are named colors, rest are 256-color palette
	cellW := 2
	gridW := 16 * cellW

	for row := 0; row < 16; row++ {
		for col := 0; col < 16; col++ {
			idx := row*16 + col
			cx := x + col*cellW
			cy := y + row

			var color buffer.Color
			if idx < 16 {
				color = buffer.NamedColor(idx)
			} else {
				color = buffer.Color256Val(uint8(idx))
			}

			// Draw colored cell (2 chars wide)
			for dx := 0; dx < cellW && cx+dx < x+maxW; dx++ {
				buf.SetCell(cx+dx, cy, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Bg:    color,
				})
			}

			// Draw cursor border
			if idx == cp.paletteIdx {
				for dx := 0; dx < cellW && cx+dx < x+maxW; dx++ {
					cell := buf.GetCell(cx+dx, cy)
					cell.Fg = buffer.Yellow
					cell.Flags |= buffer.Reverse
					buf.SetCell(cx+dx, cy, cell)
				}
			}
		}
	}

	// Show palette index
	idxStr := fmt.Sprintf("Index: %d/255", cp.paletteIdx)
	cp.drawString(buf, x+gridW+2, y, idxStr, cp.style.Label)
}

func (cp *ColorPicker) paintRGB(buf *buffer.Buffer, x, y, maxW int) {
	labels := []string{"Red  ", "Green", "Blue "}
	values := []uint8{cp.r, cp.g, cp.b}
	sliderW := maxW - 20
	if sliderW > 30 {
		sliderW = 30
	}
	if sliderW < 10 {
		sliderW = 10
	}

	for i := 0; i < 3; i++ {
		rowY := y + i*2

		// Channel label
		labelStyle := cp.style.Label
		if i == cp.activeChannel {
			labelStyle = cp.style.Value
		}
		cp.drawString(buf, x, rowY, labels[i], labelStyle)

		// Slider bar
		sliderX := x + 8
		filled := int(float64(sliderW) * float64(values[i]) / 255.0)

		for sx := 0; sx < sliderW; sx++ {
			var style buffer.Style
			if sx < filled {
				switch i {
				case 0:
					style = buffer.Style{Fg: buffer.RGB(uint8(255), 0, 0), Flags: buffer.Reverse}
				case 1:
					style = buffer.Style{Fg: buffer.RGB(0, uint8(255), 0), Flags: buffer.Reverse}
				case 2:
					style = buffer.Style{Fg: buffer.RGB(0, 0, uint8(255)), Flags: buffer.Reverse}
				}
			} else {
				style = cp.style.Box
			}
			buf.SetCell(sliderX+sx, rowY, buffer.Cell{
				Rune:  '─',
				Width: 1,
				Fg:    style.Fg,
				Flags: style.Flags,
			})
		}

		// Value text
		valStr := fmt.Sprintf("%3d", values[i])
		valStyle := cp.style.Value
		if i != cp.activeChannel {
			valStyle = cp.style.Label
		}
		cp.drawString(buf, sliderX+sliderW+2, rowY, valStr, valStyle)
	}

	// Active channel hint
	hint := "←→: switch channel  ↑↓: ±1  H/L: ±10"
	cp.drawString(buf, x, y+7, hint, cp.style.Box)
}

func (cp *ColorPicker) paintHex(buf *buffer.Buffer, x, y, maxW int) {
	// Label
	cp.drawString(buf, x, y, "Hex: #", cp.style.Label)

	// Hex input display
	hexX := x + 6
	for i, r := range cp.hexBuf {
		style := cp.style.Value
		if i == cp.hexCursor {
			style = cp.style.Cursor
		}
		buf.SetCell(hexX+i, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    style.Fg,
			Flags: style.Flags,
		})
	}

	// Cursor indicator (block at cursor position)
	if cp.hexCursor < 6 {
		buf.SetCell(hexX+cp.hexCursor, y, buffer.Cell{
			Rune:  cp.hexBuf[cp.hexCursor],
			Width: 1,
			Fg:    buffer.Black,
			Bg:    buffer.White,
		})
	}

	// Color name / description
	hexStr := "#" + string(cp.hexBuf)
	cp.drawString(buf, x, y+2, "Preview: "+hexStr, cp.style.Label)

	// Swatch under hex
	swatchColor := buffer.Hex(hexStr)
	for dx := 0; dx < 20 && x+dx < x+maxW; dx++ {
		buf.SetCell(x+dx, y+4, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Bg:    swatchColor,
		})
	}

	// Hint
	cp.drawString(buf, x, y+6, "Type hex digits (0-9, a-f). Backspace to delete.", cp.style.Box)
}

func (cp *ColorPicker) drawSwatch(buf *buffer.Buffer, x, y, w, h int) {
	// Border
	for dx := 0; dx < w; dx++ {
		buf.SetCell(x+dx, y, buffer.Cell{Rune: '─', Width: 1, Fg: cp.style.SwatchBorder.Fg})
		buf.SetCell(x+dx, y+h-1, buffer.Cell{Rune: '─', Width: 1, Fg: cp.style.SwatchBorder.Fg})
	}
	for dy := 0; dy < h; dy++ {
		buf.SetCell(x, y+dy, buffer.Cell{Rune: '│', Width: 1, Fg: cp.style.SwatchBorder.Fg})
		buf.SetCell(x+w-1, y+dy, buffer.Cell{Rune: '│', Width: 1, Fg: cp.style.SwatchBorder.Fg})
	}
	buf.SetCell(x, y, buffer.Cell{Rune: '┌', Width: 1, Fg: cp.style.SwatchBorder.Fg})
	buf.SetCell(x+w-1, y, buffer.Cell{Rune: '┐', Width: 1, Fg: cp.style.SwatchBorder.Fg})
	buf.SetCell(x, y+h-1, buffer.Cell{Rune: '└', Width: 1, Fg: cp.style.SwatchBorder.Fg})
	buf.SetCell(x+w-1, y+h-1, buffer.Cell{Rune: '┘', Width: 1, Fg: cp.style.SwatchBorder.Fg})

	// Fill with color
	for dy := 1; dy < h-1; dy++ {
		for dx := 1; dx < w-1; dx++ {
			buf.SetCell(x+dx, y+dy, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    cp.color,
			})
		}
	}
}

func (cp *ColorPicker) drawString(buf *buffer.Buffer, x, y int, s string, style buffer.Style) {
	cols := []rune(s)
	for i, r := range cols {
		buf.SetCell(x+i, y, buffer.Cell{
			Rune:  r,
			Width: 1,
			Fg:    style.Fg,
			Bg:    style.Bg,
			Flags: style.Flags,
		})
	}
}

// ColorName returns a human-readable name for common colors.
func ColorName(c buffer.Color) string {
	if c.Type == buffer.ColorNamed {
		switch c.Val {
		case 0:
			return "Black"
		case 1:
			return "Red"
		case 2:
			return "Green"
		case 3:
			return "Yellow"
		case 4:
			return "Blue"
		case 5:
			return "Magenta"
		case 6:
			return "Cyan"
		case 7:
			return "White"
		case 8:
			return "Bright Black"
		case 9:
			return "Bright Red"
		case 10:
			return "Bright Green"
		case 11:
			return "Bright Yellow"
		case 12:
			return "Bright Blue"
		case 13:
			return "Bright Magenta"
		case 14:
			return "Bright Cyan"
		case 15:
			return "Bright White"
		}
		return "Named " + strconv.Itoa(int(c.Val))
	}
	if c.Type == buffer.Color256 {
		return "256 #" + strconv.Itoa(int(c.Val))
	}
	if c.Type == buffer.ColorTrue {
		return "#" + strings.ToUpper(fmt.Sprintf("%06x", c.Val))
	}
	return "Default"
}
