package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Carousel: Paginated Content Carousel ───
//
// Carousel displays one "slide" at a time from a list, with navigation dots
// (●○○○) and prev/next indicators. Useful for onboarding flows, image
// galleries, and feature showcases.
//
// Usage:
//
//	c := NewCarousel()
//	c.AddSlide("Welcome", "Get started with Fluui!")
//	c.AddSlide("Features", "160+ components")
//	c.Next()
//	c.Paint(buf)

// CarouselStyle holds styling for Carousel.
type CarouselStyle struct {
	Title    buffer.Style
	Content  buffer.Style
	Dots     buffer.Style
	Active   buffer.Style // active dot
	Nav      buffer.Style // prev/next arrows
	Border   buffer.Style
}

// DefaultCarouselStyle returns sensible defaults.
func DefaultCarouselStyle() CarouselStyle {
	title := buffer.Style{Fg: buffer.RGB(167, 139, 250), Flags: buffer.Bold} // violet-400 bold
	content := buffer.Style{Fg: buffer.RGB(226, 232, 240)}                    // slate-200
	dots := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                         // slate-600
	active := buffer.Style{Fg: buffer.RGB(96, 165, 250)}                      // blue-400
	nav := buffer.Style{Fg: buffer.RGB(148, 163, 184)}                        // slate-400
	border := buffer.Style{Fg: buffer.RGB(71, 85, 105)}                       // slate-600
	return CarouselStyle{Title: title, Content: content, Dots: dots, Active: active, Nav: nav, Border: border}
}

// CarouselSlide represents a single slide in the carousel.
type CarouselSlide struct {
	Title   string
	Content string
}

// Carousel displays paginated content with navigation dots.
type Carousel struct {
	BaseComponent
	mu sync.Mutex

	slides []CarouselSlide
	current int
	style   CarouselStyle
}

// NewCarousel creates a Carousel with defaults.
func NewCarousel() *Carousel {
	c := &Carousel{
		style: DefaultCarouselStyle(),
	}
	c.SetID(GenerateID("carousel"))
	return c
}

// AddSlide adds a slide to the carousel.
func (c *Carousel) AddSlide(title, content string) *Carousel {
	c.mu.Lock()
	c.slides = append(c.slides, CarouselSlide{Title: title, Content: content})
	c.mu.Unlock()
	return c
}

// Next advances to the next slide (wraps around).
func (c *Carousel) Next() *Carousel {
	c.mu.Lock()
	if len(c.slides) > 0 {
		c.current = (c.current + 1) % len(c.slides)
	}
	c.mu.Unlock()
	return c
}

// Prev goes to the previous slide (wraps around).
func (c *Carousel) Prev() *Carousel {
	c.mu.Lock()
	if len(c.slides) > 0 {
		c.current = (c.current - 1 + len(c.slides)) % len(c.slides)
	}
	c.mu.Unlock()
	return c
}

// CurrentIndex returns the current slide index.
func (c *Carousel) CurrentIndex() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

// SlideCount returns the total number of slides.
func (c *Carousel) SlideCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.slides)
}

// SetCurrent sets the current slide index (clamped to valid range).
func (c *Carousel) SetCurrent(idx int) *Carousel {
	c.mu.Lock()
	if len(c.slides) > 0 {
		if idx < 0 {
			idx = 0
		}
		if idx >= len(c.slides) {
			idx = len(c.slides) - 1
		}
		c.current = idx
	}
	c.mu.Unlock()
	return c
}

// SetStyle sets the custom style.
func (c *Carousel) SetStyle(s CarouselStyle) *Carousel {
	c.mu.Lock()
	c.style = s
	c.mu.Unlock()
	return c
}

// Measure returns the preferred size.
func (c *Carousel) Measure(cs Constraints) Size {
	w := 40
	h := 8 // border + title + content(3) + dots + border
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	if cs.MaxHeight > 0 && h > cs.MaxHeight {
		h = cs.MaxHeight
	}
	return Size{W: w, H: h}
}

// Paint renders the carousel into the buffer.
func (c *Carousel) Paint(buf *buffer.Buffer) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := c.Bounds()
	x, y := b.X, b.Y
	w, h := b.W, b.H
	if w < 10 {
		w = 40
	}
	if h < 5 {
		h = 8
	}

	// Draw border
	bs := c.style.Border
	for row := 0; row < h && y+row < buf.Height; row++ {
		for col := 0; col < w && x+col < buf.Width; col++ {
			var ch rune
			if row == 0 && col == 0 {
				ch = '┌'
			} else if row == 0 && col == w-1 {
				ch = '┐'
			} else if row == h-1 && col == 0 {
				ch = '└'
			} else if row == h-1 && col == w-1 {
				ch = '┘'
			} else if row == 0 || row == h-1 {
				ch = '─'
			} else if col == 0 || col == w-1 {
				ch = '│'
			}
			if ch != 0 {
				buf.SetCell(x+col, y+row, buffer.Cell{Rune: ch, Fg: bs.Fg, Bg: bs.Bg, Flags: bs.Flags, Width: 1})
			}
		}
	}

	if len(c.slides) == 0 {
		return
	}

	// Navigation arrows
	navStyle := c.style.Nav
	if x+1 < buf.Width && y+h/2 < buf.Height {
		buf.SetCell(x+1, y+h/2, buffer.Cell{Rune: '◀', Fg: navStyle.Fg, Bg: navStyle.Bg, Flags: navStyle.Flags, Width: 1})
	}
	if x+w-2 < buf.Width && y+h/2 < buf.Height {
		buf.SetCell(x+w-2, y+h/2, buffer.Cell{Rune: '▶', Fg: navStyle.Fg, Bg: navStyle.Bg, Flags: navStyle.Flags, Width: 1})
	}

	// Current slide
	slide := c.slides[c.current]

	// Title (centered)
	titleStyle := c.style.Title
	titleY := y + 1
	titleLen := len(slide.Title)
	titleStart := x + (w-titleLen)/2
	if titleStart < x+3 {
		titleStart = x + 3
	}
	for i, r := range slide.Title {
		cx := titleStart + i
		if cx < x+w-1 && cx < buf.Width && titleY < buf.Height {
			buf.SetCell(cx, titleY, buffer.Cell{Rune: r, Fg: titleStyle.Fg, Bg: titleStyle.Bg, Flags: titleStyle.Flags, Width: 1})
		}
	}

	// Content (wrapped naively)
	contentStyle := c.style.Content
	contentStartX := x + 3
	contentEndX := x + w - 3
	contentY := y + 2
	contentCol := contentStartX
	for _, r := range slide.Content {
		if r == '\n' {
			contentY++
			contentCol = contentStartX
			continue
		}
		if contentCol >= contentEndX {
			contentY++
			contentCol = contentStartX
		}
		if contentY >= y+h-2 || contentY >= buf.Height {
			break
		}
		if contentCol < buf.Width {
			buf.SetCell(contentCol, contentY, buffer.Cell{Rune: r, Fg: contentStyle.Fg, Bg: contentStyle.Bg, Flags: contentStyle.Flags, Width: 1})
		}
		contentCol++
	}

	// Navigation dots (centered at bottom)
	dotY := y + h - 2
	dotCount := len(c.slides)
	dotTotalW := dotCount*2 - 1 // each dot + space
	dotStart := x + (w-dotTotalW)/2
	if dotStart < x+1 {
		dotStart = x + 1
	}
	dotStyle := c.style.Dots
	activeStyle := c.style.Active
	for i := 0; i < dotCount; i++ {
		dx := dotStart + i*2
		if dx >= x+w-1 || dx >= buf.Width || dotY >= buf.Height {
			break
		}
		if i == c.current {
			buf.SetCell(dx, dotY, buffer.Cell{Rune: '●', Fg: activeStyle.Fg, Bg: activeStyle.Bg, Flags: activeStyle.Flags, Width: 1})
		} else {
			buf.SetCell(dx, dotY, buffer.Cell{Rune: '○', Fg: dotStyle.Fg, Bg: dotStyle.Bg, Flags: dotStyle.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (c *Carousel) Children() []Component { return nil }
