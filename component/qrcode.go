package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"

	qrcode "github.com/skip2/go-qrcode"
)

// QRCode renders a QR code in a terminal buffer using Unicode block characters.
type QRCode struct {
	BaseComponent

	data       string
	module     int // 1 or 2 (chars per module)
	pixels     [][]bool
	matrixSize int // modules per side
	margin     int

	darkColor  buffer.Color
	lightColor buffer.Color
	bgColor    buffer.Color

	mu sync.RWMutex
}

// NewQRCode creates a QRCode component from the given data string.
func NewQRCode(data string) *QRCode {
	q := &QRCode{
		data:       data,
		module:     2,
		margin:     2,
		darkColor:  buffer.RGB(255, 255, 255),
		lightColor: buffer.RGB(0, 0, 0),
		bgColor:    buffer.RGB(0, 0, 0),
	}
	q.generate()
	return q
}

// SetData regenerates the QR code with new data.
func (q *QRCode) SetData(data string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.data = data
	return q.generateLocked()
}

// SetModuleSize sets the module size (1=compact, 2=standard).
func (q *QRCode) SetModuleSize(n int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if n < 1 {
		n = 1
	}
	if n > 2 {
		n = 2
	}
	q.module = n
}

// SetMargin sets the margin in modules.
func (q *QRCode) SetMargin(n int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if n < 0 {
		n = 0
	}
	q.margin = n
}

// SetColors sets the dark, light, and background colors.
func (q *QRCode) SetColors(dark, light, bg buffer.Color) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.darkColor = dark
	q.lightColor = light
	q.bgColor = bg
}

// Size returns the QR code matrix size in modules.
func (q *QRCode) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.matrixSize
}

// PixelSize returns the rendered pixel dimensions (w, h).
func (q *QRCode) PixelSize() (int, int) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	total := q.matrixSize + 2*q.margin
	w := total * q.module
	h := (total*q.module + 1) / 2 // rows are halved due to vertical block chars
	return w, h
}

func (q *QRCode) generate() {
	_ = q.generateLocked()
}

func (q *QRCode) generateLocked() error {
	qr, err := qrcode.New(q.data, qrcode.Medium)
	if err != nil {
		return err
	}
	bitmap := qr.Bitmap()
	q.matrixSize = len(bitmap)
	q.pixels = make([][]bool, q.matrixSize)
	for i := range q.pixels {
		q.pixels[i] = make([]bool, q.matrixSize)
		copy(q.pixels[i], bitmap[i])
	}
	return nil
}

// Measure returns the desired size of the QR code.
func (q *QRCode) Measure(cs Constraints) Size {
	q.mu.RLock()
	defer q.mu.RUnlock()

	total := q.matrixSize + 2*q.margin
	w := total * q.module
	h := (total + 1) / 2 // each terminal row = 2 module rows
	return Size{W: w, H: h}
}

// Paint renders the QR code into the buffer using Unicode block characters.
func (q *QRCode) Paint(buf *buffer.Buffer) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.pixels) == 0 || buf == nil {
		return
	}

	m := q.module
	bounds := q.Bounds()
	offX, offY := bounds.X, bounds.Y

	// Helper: get pixel value at (mx, my) in module coords, considering margin.
	getPixel := func(mx, my int) bool {
		// Apply margin offset
		ax := mx - q.margin
		ay := my - q.margin
		if ax < 0 || ay < 0 || ax >= q.matrixSize || ay >= q.matrixSize {
			return false // margin area is "light"
		}
		return q.pixels[ay][ax]
	}

	total := q.matrixSize + 2*q.margin

	// Render: each terminal row covers 2 module rows (vertical compression).
	for ty := 0; ty < (total+1)/2; ty++ {
		for tx := 0; tx < total; tx++ {
			// Top module row
			topY := ty * 2
			botY := ty*2 + 1

			topDark := false
			botDark := false

			// Expand module to m chars wide
			for dx := 0; dx < m; dx++ {
				_ = dx
			}

			// Check pixel at expanded coords
			topDark = getPixel(tx, topY)
			if botY < total {
				botDark = getPixel(tx, botY)
			} else {
				botDark = false
			}

			var r rune
			var fg buffer.Color
			if topDark && botDark {
				r = '█'
				fg = q.darkColor
			} else if topDark && !botDark {
				r = '▀'
				fg = q.darkColor
			} else if !topDark && botDark {
				r = '▄'
				fg = q.darkColor
			} else {
				r = ' '
				fg = q.lightColor
			}

			// Draw m chars wide
			for dx := 0; dx < m; dx++ {
				x := offX + tx*m + dx
				y := offY + ty
				if x >= 0 && x < buf.Width && y >= 0 && y < buf.Height {
					buf.SetCell(x, y, buffer.Cell{
						Rune:  r,
						Width: 1,
						Fg:    fg,
						Bg:    q.bgColor,
					})
				}
			}
		}
	}
}

// HandleKey is a no-op for QR codes (non-interactive).
func (q *QRCode) HandleKey(k *term.KeyEvent) bool {
	return false
}

// Children returns nil (QR codes have no children).
func (q *QRCode) Children() []Component {
	return nil
}
