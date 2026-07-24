package component

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// SkeletonBlock describes a single placeholder rectangle in a skeleton layout.
type SkeletonBlock struct {
	X, Y, W, H int // relative to component origin
}

// SkeletonLoader displays animated placeholder blocks ("skeletons") while
// content is loading. This is the standard loading state in modern web and
// mobile apps, and is increasingly used in AI TUIs to show "content is
// being generated" before the actual response arrives.
//
// The shimmer animation cycles through blocks, making them pulse between
// dim and accent colors. Thread-safe with zero-alloc Paint.
type SkeletonLoader struct {
	BaseComponent
	mu sync.Mutex

	blocks  []SkeletonBlock
	frames  uint64 // atomic, incremented each tick
	running atomic.Bool
	stopCh  chan struct{}
}

// NewSkeletonLoader creates a skeleton loader with the given blocks.
func NewSkeletonLoader(blocks []SkeletonBlock) *SkeletonLoader {
	return &SkeletonLoader{
		BaseComponent: BaseComponent{id: GenerateID("skeleton")},
		blocks:        blocks,
		stopCh:        make(chan struct{}),
	}
}

// NewSkeletonText creates a text-line skeleton with N lines of the given width.
// Each line is 1 row tall with 1-row spacing.
func NewSkeletonText(lines, width int) *SkeletonLoader {
	blocks := make([]SkeletonBlock, lines)
	for i := 0; i < lines; i++ {
		w := width
		if i == lines-1 {
			w = width * 2 / 3 // last line shorter
			if w < 4 {
				w = 4
			}
		}
		blocks[i] = SkeletonBlock{X: 0, Y: i * 2, W: w, H: 1}
	}
	return NewSkeletonLoader(blocks)
}

// SetBlocks replaces the skeleton layout.
func (s *SkeletonLoader) SetBlocks(blocks []SkeletonBlock) {
	s.mu.Lock()
	s.blocks = blocks
	s.mu.Unlock()
}

// Blocks returns the current skeleton blocks.
func (s *SkeletonLoader) Blocks() []SkeletonBlock {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.blocks
}

// Start begins the shimmer animation at the given interval.
func (s *SkeletonLoader) Start(interval time.Duration) {
	if s.running.Load() {
		return
	}
	s.running.Store(true)
	s.mu.Lock()
	s.stopCh = make(chan struct{})
	stopCh := s.stopCh
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				atomic.AddUint64(&s.frames, 1)
			}
		}
	}()
}

// Stop halts the animation.
func (s *SkeletonLoader) Stop() {
	if !s.running.Load() {
		return
	}
	s.running.Store(false)
	s.mu.Lock()
	if s.stopCh != nil {
		select {
		case <-s.stopCh:
		default:
			close(s.stopCh)
		}
	}
	s.mu.Unlock()
}

// IsRunning returns whether the animation is active.
func (s *SkeletonLoader) IsRunning() bool {
	return s.running.Load()
}

// FrameIndex returns the current shimmer frame index (0 or 1).
func (s *SkeletonLoader) FrameIndex() int {
	return int(atomic.LoadUint64(&s.frames) % 2)
}

// Measure returns the desired size based on block layout.
func (s *SkeletonLoader) Measure(cs Constraints) Size {
	s.mu.Lock()
	defer s.mu.Unlock()

	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 40
	}
	maxH := 0
	for _, blk := range s.blocks {
		bot := blk.Y + blk.H
		if bot > maxH {
			maxH = bot
		}
	}
	if maxH < 1 {
		maxH = 1
	}
	return Size{W: maxW, H: maxH}
}

// Paint renders the skeleton blocks with shimmer animation.
func (s *SkeletonLoader) Paint(buf *buffer.Buffer) {
	s.mu.Lock()
	blocks := s.blocks
	s.mu.Unlock()

	th := theme.Get()
	b := s.Bounds()

	frame := s.FrameIndex()

	for _, blk := range blocks {
		x := b.X + blk.X
		y := b.Y + blk.Y
		w := blk.W
		h := blk.H

		// Clamp to buffer bounds
		if x < b.X {
			x = b.X
		}
		if y < b.Y {
			y = b.Y
		}

		// Alternate shade on odd frames
		var bg theme.Color
		if frame == 0 {
			bg = th.Border
		} else {
			bg = th.Muted
		}
		cellStyle := buffer.Style{Bg: bg}

		for row := 0; row < h; row++ {
			for col := 0; col < w; col++ {
				bx := x + col
				by := y + row
				if bx >= b.X+b.W || by >= b.Y+b.H {
					break
				}
				buf.SetCell(bx, by, buffer.Cell{
					Rune:  ' ',
					Width: 1,
					Fg:    cellStyle.Fg,
					Bg:    cellStyle.Bg,
				})
			}
		}
	}
}
