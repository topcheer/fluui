package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── FileSizeBar: Human-Readable File Size Display ───
//
// FileSizeBar renders a horizontal bar showing file/disk usage with
// human-readable size labels (B, KB, MB, GB, TB). Useful for storage
// and bandwidth monitoring dashboards.
//
// Usage:
//
//	fb := NewFileSizeBar()
//	fb.SetSize(8589934592, 107374182400) // 8GB of 100GB
//	fb.Paint(buf)

// FileSizeBarStyle holds styling.
type FileSizeBarStyle struct {
	Fill   buffer.Style
	Empty  buffer.Style
	Label  buffer.Style
	Value  buffer.Style
	Suffix buffer.Style
}

// DefaultFileSizeBarStyle returns defaults.
func DefaultFileSizeBarStyle() FileSizeBarStyle {
	return FileSizeBarStyle{
		Fill:   buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Empty:  buffer.Style{Fg: buffer.RGB(51, 65, 85)},
		Label:  buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Value:  buffer.Style{Fg: buffer.RGB(226, 232, 240), Flags: buffer.Bold},
		Suffix: buffer.Style{Fg: buffer.RGB(100, 116, 139)},
	}
}

var sizeSuffixes = [...]string{"B", "KB", "MB", "GB", "TB", "PB"}

// FileSizeBar renders a human-readable file size bar.
type FileSizeBar struct {
	BaseComponent
	mu sync.Mutex

	used  int64
	total int64
	width int
	style FileSizeBarStyle
	// cached
	usedStr  string
	totalStr string
	usedIdx  int
	totalIdx int
	fillW    int
}

// NewFileSizeBar creates a FileSizeBar.
func NewFileSizeBar() *FileSizeBar {
	fb := &FileSizeBar{total: 100, width: 24, style: DefaultFileSizeBarStyle()}
	fb.SetID(GenerateID("filesize"))
	fb.recomputeLocked()
	return fb
}

// SetSize sets used and total bytes.
func (fb *FileSizeBar) SetSize(used, total int64) *FileSizeBar {
	fb.mu.Lock()
	if total < 1 {
		total = 1
	}
	if used < 0 {
		used = 0
	}
	if used > total {
		used = total
	}
	fb.used = used
	fb.total = total
	fb.recomputeLocked()
	fb.mu.Unlock()
	return fb
}

func humanizeSize(bytes int64) (int, string) {
	val := bytes
	idx := 0
	for val >= 1024 && idx < len(sizeSuffixes)-1 {
		val >>= 10
		idx++
	}
	return int(val), sizeSuffixes[idx]
}

func (fb *FileSizeBar) recomputeLocked() {
	usedVal, usedSuf := humanizeSize(fb.used)
	totalVal, totalSuf := humanizeSize(fb.total)
	fb.usedStr = itoa(usedVal) + usedSuf
	fb.totalStr = itoa(totalVal) + totalSuf
	const barW = 16
	fb.fillW = int(fb.used) * barW / int(fb.total)
}

// Used returns the used bytes.
func (fb *FileSizeBar) Used() int64 {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	return fb.used
}

// SetWidth sets the bar width.
func (fb *FileSizeBar) SetWidth(w int) *FileSizeBar {
	fb.mu.Lock()
	if w < 10 {
		w = 10
	}
	fb.width = w
	fb.mu.Unlock()
	return fb
}

// SetStyle sets custom style.
func (fb *FileSizeBar) SetStyle(s FileSizeBarStyle) *FileSizeBar {
	fb.mu.Lock()
	fb.style = s
	fb.mu.Unlock()
	return fb
}

// Measure returns preferred size.
func (fb *FileSizeBar) Measure(cs Constraints) Size {
	w := fb.width + 12
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the file size bar.
func (fb *FileSizeBar) Paint(buf *buffer.Buffer) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	b := fb.Bounds()
	x, y := b.X, b.Y

	fillStyle := fb.style.Fill
	emptyStyle := fb.style.Empty
	valueStyle := fb.style.Value
	labelStyle := fb.style.Label

	col := x
	const barW = 16
	for i := 0; i < fb.fillW; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '█', Fg: fillStyle.Fg, Bg: fillStyle.Bg, Flags: fillStyle.Flags, Width: 1})
		col++
	}
	for i := fb.fillW; i < barW; i++ {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: '░', Fg: emptyStyle.Fg, Bg: emptyStyle.Bg, Flags: emptyStyle.Flags, Width: 1})
		col++
	}

	// Value labels
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range fb.usedStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '/', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}
	for _, r := range fb.totalStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: valueStyle.Fg, Bg: valueStyle.Bg, Flags: valueStyle.Flags, Width: 1})
		col++
	}
}

// Children returns nil.
func (fb *FileSizeBar) Children() []Component { return nil }
