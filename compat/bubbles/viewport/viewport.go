// Package viewport provides a drop-in compatibility layer for
// charm.land/bubbles/v2/viewport.
//
// It wraps fluui's component.Viewport with the bubbles viewport API.
package viewport

import (
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// Model wraps component.Viewport with the bubbles.viewport API.
type Model struct {
	*component.Viewport
	Width      int
	Height     int
	AutoFollow bool
}

// New creates a new viewport Model (bubbles.viewport.New).
func New(width, height int) Model {
	vp := component.NewViewport(component.NewFill(' ', buffer.Style{}))
	vp.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: height})
	return Model{
		Viewport: vp,
		Width:    width,
		Height:   height,
	}
}

// SetSize sets the viewport dimensions.
func (m *Model) SetSize(width, height int) {
	m.Width = width
	m.Height = height
	m.Viewport.SetBounds(component.Rect{X: 0, Y: 0, W: width, H: height})
}

// SetContent sets the content component.
func (m *Model) SetContent(c component.Component) {
	m.Viewport.SetContent(c)
}

// ScrollUp scrolls up by n lines.
func (m *Model) ScrollUp(n int) {
	m.Viewport.ScrollUp(n)
}

// ScrollDown scrolls down by n lines.
func (m *Model) ScrollDown(n int) {
	m.Viewport.ScrollDown(n)
}

// GotoBottom scrolls to the bottom.
func (m *Model) GotoBottom() {
	m.Viewport.ScrollToBottom()
}

// GotoTop scrolls to the top.
func (m *Model) GotoTop() {
	m.Viewport.ScrollToTop()
}

// AtBottom returns whether the viewport is at the bottom.
func (m *Model) AtBottom() bool {
	return m.Viewport.OffsetY() >= m.Viewport.MaxOffsetY()
}

// YOffset returns the current vertical scroll offset.
func (m *Model) YOffset() int {
	return m.Viewport.OffsetY()
}

// XOffset returns the current horizontal scroll offset.
func (m *Model) XOffset() int {
	return m.Viewport.OffsetX()
}

// TotalLineCount returns the total number of lines in the content.
func (m *Model) TotalLineCount() int {
	return m.Viewport.ContentHeight()
}

// VisibleLineCount returns the number of visible lines.
func (m *Model) VisibleLineCount() int {
	return m.Height
}

// View returns the rendered string (bubbles viewport.View()).
func (m *Model) View() string {
	// In fluui, rendering is done via Paint(buf), not View().
	// This compat method returns a placeholder — actual rendering
	// goes through the fluui component pipeline.
	return ""
}

// Update handles key events for viewport navigation.
func (m *Model) Update(key *term.KeyEvent) {
	m.Viewport.HandleKey(key)
}

// SetAutoFollow enables/disables auto-follow (scroll to bottom on new content).
func (m *Model) SetAutoFollow(b bool) {
	m.AutoFollow = b
}