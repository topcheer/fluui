package component

import (
	"fmt"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// PaginationStyle controls visual appearance of the pagination bar.
type PaginationStyle struct {
	Normal    buffer.Style
	Selected  buffer.Style
	Disabled  buffer.Style
	Separator buffer.Style
	Arrow     buffer.Style
}

// DefaultPaginationStyle returns a default style set.
func DefaultPaginationStyle() PaginationStyle {
	return PaginationStyle{
		Normal:    buffer.Style{Fg: buffer.RGB(248, 248, 242)},
		Selected:  buffer.Style{Fg: buffer.RGB(248, 248, 242), Bg: buffer.RGB(68, 71, 90), Flags: buffer.Bold},
		Disabled:  buffer.Style{Fg: buffer.RGB(98, 114, 164)},
		Separator: buffer.Style{Fg: buffer.RGB(98, 114, 164)},
		Arrow:     buffer.Style{Fg: buffer.RGB(189, 147, 249)},
	}
}

// Pagination is a horizontal page navigation component.
// It implements the Component interface and is typically rendered as a 1-line bar.
type Pagination struct {
	BaseComponent
	mu sync.RWMutex

	totalPages int
	currentPage int
	itemsPerPage int
	totalItems  int
	style       PaginationStyle

	// How many page numbers to show on each side of current
	pageRange int

	// Callbacks
	OnPageChange func(page int)
}

// NewPagination creates a new pagination with defaults.
func NewPagination() *Pagination {
	return &Pagination{
		BaseComponent: BaseComponent{id: GenerateID("pagination")},
		style:         DefaultPaginationStyle(),
		itemsPerPage:  20,
		pageRange:     2,
	}
}

// --- Page Management ---

// SetTotalItems sets the total number of items and recomputes page count.
func (p *Pagination) SetTotalItems(n int) {
	p.mu.Lock()
	p.totalItems = n
	p.recomputePagesLocked()
	p.clampPageLocked()
	p.mu.Unlock()
}

// SetItemsPerPage sets items per page and recomputes page count.
func (p *Pagination) SetItemsPerPage(n int) {
	p.mu.Lock()
	if n > 0 {
		p.itemsPerPage = n
	}
	p.recomputePagesLocked()
	p.clampPageLocked()
	p.mu.Unlock()
}

// SetTotalPages directly sets the page count (useful when totalItems is unknown).
func (p *Pagination) SetTotalPages(n int) {
	p.mu.Lock()
	p.totalPages = n
	p.clampPageLocked()
	p.mu.Unlock()
}

// TotalPages returns the total number of pages.
func (p *Pagination) TotalPages() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.totalPages
}

// CurrentPage returns the current page (0-indexed).
func (p *Pagination) CurrentPage() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPage
}

// ItemsPerPage returns items per page setting.
func (p *Pagination) ItemsPerPage() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.itemsPerPage
}

// TotalItems returns the total item count.
func (p *Pagination) TotalItems() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.totalItems
}

// SetPage sets the current page, clamping to valid range.
func (p *Pagination) SetPage(page int) {
	p.mu.Lock()
	p.clampPageToLocked(page)
	p.mu.Unlock()
}

// NextPage advances to the next page if not on the last.
func (p *Pagination) NextPage() bool {
	p.mu.Lock()
	if p.currentPage >= p.totalPages-1 {
		p.mu.Unlock()
		return false
	}
	p.currentPage++
	cb := p.OnPageChange
	page := p.currentPage
	p.mu.Unlock()
	if cb != nil {
		cb(page)
	}
	return true
}

// PrevPage moves to the previous page if not on the first.
func (p *Pagination) PrevPage() bool {
	p.mu.Lock()
	if p.currentPage <= 0 {
		p.mu.Unlock()
		return false
	}
	p.currentPage--
	cb := p.OnPageChange
	page := p.currentPage
	p.mu.Unlock()
	if cb != nil {
		cb(page)
	}
	return true
}

// FirstPage jumps to the first page.
func (p *Pagination) FirstPage() {
	p.mu.Lock()
	p.currentPage = 0
	cb := p.OnPageChange
	p.mu.Unlock()
	if cb != nil {
		cb(0)
	}
}

// LastPage jumps to the last page.
func (p *Pagination) LastPage() {
	p.mu.Lock()
	if p.totalPages > 0 {
		p.currentPage = p.totalPages - 1
	}
	cb := p.OnPageChange
	page := p.currentPage
	p.mu.Unlock()
	if cb != nil {
		cb(page)
	}
}

// HasNext returns true if there's a next page.
func (p *Pagination) HasNext() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPage < p.totalPages-1
}

// HasPrev returns true if there's a previous page.
func (p *Pagination) HasPrev() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPage > 0
}

// IsEmpty returns true if there are no pages.
func (p *Pagination) IsEmpty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.totalPages == 0
}

// --- Range helpers ---

// PageStartIndex returns the starting item index for the current page.
func (p *Pagination) PageStartIndex() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentPage * p.itemsPerPage
}

// PageEndIndex returns the ending item index (exclusive) for the current page.
func (p *Pagination) PageEndIndex() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	end := (p.currentPage + 1) * p.itemsPerPage
	if end > p.totalItems {
		end = p.totalItems
	}
	return end
}

// CurrentPageItems returns indices for items on the current page.
func (p *Pagination) CurrentPageItems() []int {
	start := p.PageStartIndex()
	end := p.PageEndIndex()
	var result []int
	for i := start; i < end; i++ {
		result = append(result, i)
	}
	return result
}

// --- Configuration ---

// SetStyle sets the visual style.
func (p *Pagination) SetStyle(s PaginationStyle) {
	p.mu.Lock()
	p.style = s
	p.mu.Unlock()
}

// Style returns the current style.
func (p *Pagination) Style() PaginationStyle {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.style
}

// SetPageRange sets how many page numbers to show on each side of current.
func (p *Pagination) SetPageRange(r int) {
	p.mu.Lock()
	if r >= 0 {
		p.pageRange = r
	}
	p.mu.Unlock()
}

// PageRange returns the current page range setting.
func (p *Pagination) PageRange() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pageRange
}

// --- Component Interface ---

// Measure returns the desired size (fills width, height=1).
func (p *Pagination) Measure(cs Constraints) Size {
	w := 40
	if cs.HasWidth() {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the pagination bar.
func (p *Pagination) Paint(buf *buffer.Buffer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	b := p.bounds
	if b.W <= 0 || b.H <= 0 {
		return
	}

	if p.totalPages == 0 {
		return
	}

	// Build the visible page list with ellipsis
	pages := p.buildVisiblePagesLocked()

	x := b.X
	for _, pg := range pages {
		if x >= b.X+b.W {
			break
		}
		if pg == -1 {
			// Ellipsis
			for _, r := range "…" {
				if x < b.X+b.W {
					buf.SetCell(x, b.Y, buffer.NewCell(r, p.style.Separator))
					x++
				}
			}
			x++ // space after ellipsis
			continue
		}

		label := fmt.Sprintf(" %d ", pg+1) // 1-indexed display
		style := p.style.Normal
		if pg == p.currentPage {
			style = p.style.Selected
		}

		for _, r := range label {
			if x >= b.X+b.W {
				break
			}
			buf.SetCell(x, b.Y, buffer.NewCell(r, style))
			x++
		}
	}

	// Right-align page info
	info := fmt.Sprintf("%d/%d", p.currentPage+1, p.totalPages)
	infoX := b.X + b.W - len([]rune(info)) - 1
	if infoX > x {
		for _, r := range info {
			if infoX >= b.X+b.W {
				break
			}
			buf.SetCell(infoX, b.Y, buffer.NewCell(r, p.style.Arrow))
			infoX++
		}
	}
}

// String returns a debug description.
func (p *Pagination) String() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return fmt.Sprintf("Pagination{pages:%d current:%d items:%d perPage:%d}",
		p.totalPages, p.currentPage, p.totalItems, p.itemsPerPage)
}

// --- Internal helpers ---

func (p *Pagination) recomputePagesLocked() {
	if p.itemsPerPage <= 0 {
		p.totalPages = 0
		return
	}
	p.totalPages = (p.totalItems + p.itemsPerPage - 1) / p.itemsPerPage
	if p.totalPages == 0 && p.totalItems > 0 {
		p.totalPages = 1
	}
}

func (p *Pagination) clampPageLocked() {
	p.clampPageToLocked(p.currentPage)
}

func (p *Pagination) clampPageToLocked(page int) {
	if page < 0 {
		page = 0
	}
	if p.totalPages == 0 {
		p.currentPage = 0
		return
	}
	if page >= p.totalPages {
		page = p.totalPages - 1
	}
	p.currentPage = page
}

// buildVisiblePagesLocked returns a list of page indices to display,
// with -1 representing an ellipsis.
func (p *Pagination) buildVisiblePagesLocked() []int {
	var pages []int
	r := p.pageRange
	start := p.currentPage - r
	end := p.currentPage + r

	if start > 0 {
		pages = append(pages, 0)
		if start > 1 {
			pages = append(pages, -1) // ellipsis
		}
	}

	for i := start; i <= end; i++ {
		if i >= 0 && i < p.totalPages {
			pages = append(pages, i)
		}
	}

	if end < p.totalPages-1 {
		if end < p.totalPages-2 {
			pages = append(pages, -1) // ellipsis
		}
		pages = append(pages, p.totalPages-1)
	}

	return pages
}
