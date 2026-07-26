package component

import (
	"strconv"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// Rating renders a star rating display (e.g., ★★★☆☆ for 3/5).
// Useful for AI model quality scores, feedback UIs, and review components.
//
// Thread-safe.
type Rating struct {
	BaseComponent
	mu      sync.RWMutex
	value   float64 // current rating (e.g., 3.5)
	max     int     // max stars (e.g., 5)
	star    rune    // filled star char
	half    rune    // half star char
	empty   rune    // empty star char
	showNum bool    // show numeric value after stars
}

// NewRating creates a rating with the given value and max stars.
func NewRating(value float64, max int) *Rating {
	if max < 1 {
		max = 1
	}
	return &Rating{
		BaseComponent: BaseComponent{id: GenerateID("rating")},
		value:         clampFloat(value, 0, float64(max)),
		max:           max,
		star:          '\u2605', // ★
		half:          '\u2606', // ☆ (simplified — no half-star unicode in most terminals)
		empty:         '\u2606', // ☆
		showNum:       false,
	}
}

// Value returns the current rating value.
func (r *Rating) Value() float64 { r.mu.RLock(); defer r.mu.RUnlock(); return r.value }

// SetValue sets the rating value. Clamped to [0, max].
func (r *Rating) SetValue(v float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = clampFloat(v, 0, float64(r.max))
}

// Max returns the maximum star count.
func (r *Rating) Max() int { r.mu.RLock(); defer r.mu.RUnlock(); return r.max }

// SetMax sets the maximum star count.
func (r *Rating) SetMax(m int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if m < 1 {
		m = 1
	}
	r.max = m
	r.value = clampFloat(r.value, 0, float64(m))
}

// ShowNumber returns whether the numeric value is displayed.
func (r *Rating) ShowNumber() bool { r.mu.RLock(); defer r.mu.RUnlock(); return r.showNum }

// SetShowNumber toggles numeric value display after stars.
func (r *Rating) SetShowNumber(b bool) { r.mu.Lock(); defer r.mu.Unlock(); r.showNum = b }

// SetStars overrides the star characters. Defaults: ★, ☆, ☆.
func (r *Rating) SetStars(filled, half, empty rune) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.star = filled
	r.half = half
	r.empty = empty
}

// Measure returns the preferred size.
func (r *Rating) Measure(cs Constraints) Size {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w := r.max // one char per star
	if r.showNum {
		w += 3 // " 4.5"
	}
	h := 1
	if cs.HasWidth() && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return Size{W: w, H: h}
}

// Paint draws the rating stars. Zero allocations.
func (r *Rating) Paint(buf *buffer.Buffer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bounds := r.bounds
	if bounds.W <= 0 || bounds.H <= 0 {
		return
	}

	tt := theme.Get()
	filledStyle := buffer.Style{Fg: tt.Warning, Flags: buffer.Bold}
	emptyStyle := buffer.Style{Fg: tt.Muted}

	x := bounds.X
	y := bounds.Y
	maxX := bounds.X + bounds.W

	for i := 0; i < r.max && x < maxX; i++ {
		starVal := r.value - float64(i)
		if starVal >= 0.75 {
			buf.SetCell(x, y, buffer.Cell{Rune: r.star, Width: 1, Fg: filledStyle.Fg, Flags: filledStyle.Flags})
		} else if starVal >= 0.25 {
			buf.SetCell(x, y, buffer.Cell{Rune: r.half, Width: 1, Fg: filledStyle.Fg, Flags: 0})
		} else {
			buf.SetCell(x, y, buffer.Cell{Rune: r.empty, Width: 1, Fg: emptyStyle.Fg})
		}
		x++
	}

	// Numeric value
	if r.showNum && x+1 < maxX {
		buf.SetCell(x, y, buffer.Cell{Rune: ' ', Width: 1, Fg: tt.Muted})
		x++
		var nb [8]byte
		obs := nb[:0]
		obs = strconv.AppendFloat(obs, r.value, 'f', 1, 64)
		numStyle := buffer.Style{Fg: tt.Muted}
		buf.DrawBytes(x, y, obs, numStyle)
	}
}
