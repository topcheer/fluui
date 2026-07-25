package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// FunnelSlice represents one stage of a funnel.
type FunnelSlice struct {
	Label string
	Value float64
}

// FunnelChart renders a horizontal funnel (trapezoid stack) showing
// decreasing values through stages. Useful for conversion funnels,
// token usage breakdown, or pipeline throughput visualization.
//
// Thread-safe.
type FunnelChart struct {
	BaseComponent
	mu     sync.Mutex
	slices []FunnelSlice
}

// NewFunnelChart creates a funnel chart.
func NewFunnelChart(slices []FunnelSlice) *FunnelChart {
	return &FunnelChart{
		BaseComponent: BaseComponent{id: GenerateID("funnel")},
		slices:        slices,
	}
}

// SetSlices replaces all slices.
func (f *FunnelChart) SetSlices(s []FunnelSlice) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.slices = s
}

// Slices returns a copy.
func (f *FunnelChart) Slices() []FunnelSlice {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]FunnelSlice, len(f.slices))
	copy(out, f.slices)
	return out
}

// SliceCount returns the number of stages.
func (f *FunnelChart) SliceCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.slices)
}

// TotalValue returns the first slice value (funnel top).
func (f *FunnelChart) TotalValue() float64 {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.slices) == 0 {
		return 0
	}
	return f.slices[0].Value
}

// Measure returns the desired size.
func (f *FunnelChart) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 40
	}
	h := f.SliceCount()
	if h < 1 {
		h = 1
	}
	return Size{W: maxW, H: h}
}

// Paint renders the funnel chart.
func (f *FunnelChart) Paint(buf *buffer.Buffer) {
	f.mu.Lock()
	slices := f.slices
	f.mu.Unlock()

	if len(slices) == 0 {
		return
	}

	b := f.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	colors := []buffer.Color{th.Accent, th.Success, th.Warning, th.Error, th.BorderMuted, th.Muted}

	maxVal := slices[0].Value
	if maxVal <= 0 {
		return
	}

	for i, slice := range slices {
		if i >= b.H {
			break
		}

		y := b.Y + i
		// Width proportional to value
		ratio := slice.Value / maxVal
		if ratio < 0 {
			ratio = 0
		}
		barW := int(ratio * float64(b.W))
		if barW < 1 {
			barW = 1
		}

		// Center the bar
		offset := (b.W - barW) / 2

		color := colors[i%len(colors)]

		// Fill bar
		for x := 0; x < barW; x++ {
			buf.SetCell(b.X+offset+x, y, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    color,
			})
		}

		// Draw label inside bar if space allows
		labelW := utf8.RuneCountInString(slice.Label)
		if labelW > 0 && labelW < barW-4 {
			lx := b.X + offset + (barW-labelW)/2
			buf.DrawText(lx, y, slice.Label, buffer.Style{Fg: th.Bg, Bg: color})
		}
	}
}
