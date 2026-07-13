package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── Pages: Named Multi-Page Navigation Container ───
//
// Pages is a simple named-page container — only one page visible at a time.
// Inspired by tview's Pages widget.
//
// Usage:
//	pages := NewPages()
//	pages.AddPage("home", homePanel)
//	pages.AddPage("settings", settingsPanel)
//	pages.SwitchTo("settings")
//	pages.HasPage("home") // true
//	pages.CurrentPage()    // "settings"

// Pages is a named multi-page container.
type Pages struct {
	mu      sync.RWMutex
	BaseComponent
	pages   map[string]Component
	order   []string
	current string
}

// NewPages creates an empty Pages container.
func NewPages() *Pages {
	return &Pages{
		pages: make(map[string]Component),
	}
}

// AddPage adds or replaces a page with the given name.
func (p *Pages) AddPage(name string, child Component) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.pages[name]; !exists {
		p.order = append(p.order, name)
	}
	p.pages[name] = child
	if p.current == "" {
		p.current = name
	}
}

// RemovePage removes a page by name.
func (p *Pages) RemovePage(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.pages, name)
	for i, n := range p.order {
		if n == name {
			p.order = append(p.order[:i], p.order[i+1:]...)
			break
		}
	}
	if p.current == name {
		if len(p.order) > 0 {
			p.current = p.order[0]
		} else {
			p.current = ""
		}
	}
}

// SwitchTo makes the named page visible. Returns false if page doesn't exist.
func (p *Pages) SwitchTo(name string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.pages[name]; !exists {
		return false
	}
	p.current = name
	return true
}

// CurrentPage returns the name of the visible page.
func (p *Pages) CurrentPage() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

// HasPage returns whether a page with the given name exists.
func (p *Pages) HasPage(name string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	_, exists := p.pages[name]
	return exists
}

// PageCount returns the number of pages.
func (p *Pages) PageCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.pages)
}

// PageNames returns all page names in insertion order.
func (p *Pages) PageNames() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]string, len(p.order))
	copy(result, p.order)
	return result
}

// CurrentComponent returns the visible page's component (or nil).
func (p *Pages) CurrentComponent() Component {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pages[p.current]
}

// NextPage switches to the next page (wraps around). Returns the new page name.
func (p *Pages) NextPage() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.order) == 0 {
		return ""
	}
	for i, name := range p.order {
		if name == p.current {
			p.current = p.order[(i+1)%len(p.order)]
			return p.current
		}
	}
	if len(p.order) > 0 {
		p.current = p.order[0]
	}
	return p.current
}

// PrevPage switches to the previous page (wraps around).
func (p *Pages) PrevPage() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.order) == 0 {
		return ""
	}
	for i, name := range p.order {
		if name == p.current {
			p.current = p.order[(i-1+len(p.order))%len(p.order)]
			return p.current
		}
	}
	if len(p.order) > 0 {
		p.current = p.order[0]
	}
	return p.current
}

func (p *Pages) Measure(constraints Constraints) Size {
	p.mu.RLock()
	current := p.pages[p.current]
	p.mu.RUnlock()
	if current == nil {
		return Size{W: 0, H: 0}
	}
	return current.Measure(constraints)
}

func (p *Pages) Paint(buf *buffer.Buffer) {
	p.mu.RLock()
	current := p.pages[p.current]
	p.mu.RUnlock()
	if current == nil {
		return
	}
	bounds := p.Bounds()
	current.SetBounds(bounds)
	current.Paint(buf)
}

func (p *Pages) Children() []Component {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]Component, 0, len(p.order))
	for _, name := range p.order {
		if c := p.pages[name]; c != nil {
			result = append(result, c)
		}
	}
	return result
}