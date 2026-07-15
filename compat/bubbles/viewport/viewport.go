// Package viewport provides a drop-in compatibility layer for
// charm.land/bubbles/v2/viewport.
//
// Unlike fluui's component.Viewport (which is a rendering component),
// this compat viewport is a self-contained string-based scrollable buffer
// that mirrors the bubbles v2 API: New(options...), SetContent(string),
// View() string, SetWidth/SetHeight/Height, Update(Msg) (Model, Cmd).
package viewport

import (
	"strings"

	tea "github.com/topcheer/fluui/compat/bubbletea"
)

// Model is a string-based scrollable viewport (bubbles v2 compatible).
type Model struct {
	// vp is the underlying content holder (accessible for compat).
	vp *vpInner

	// Dimensions
	width  int
	height int

	// Scroll state
	yOffset int

	// Options
	autoFollow          bool
	scrollIndicatorChar string
}

// vpInner holds the actual content (accessible as m.vp in ggcode).
type vpInner struct {
	content    string
	totalLines int
}

// SetYOffset sets the scroll offset (compat for m.vp.SetYOffset).
func (v *vpInner) SetYOffset(n int) {}

// Option is a viewport configuration option.
type Option func(*Model)

// WithWidth sets the viewport width.
func WithWidth(w int) Option {
	return func(m *Model) { m.width = w }
}

// WithHeight sets the viewport height.
func WithHeight(h int) Option {
	return func(m *Model) { m.height = h }
}

// ScrollIndicatorStyle returns a scroll indicator style option.
func ScrollIndicatorStyle(s string) Option {
	return func(m *Model) { m.scrollIndicatorChar = s }
}

// New creates a new viewport Model with the given options.
func New(opts ...Option) Model {
	m := Model{
		width:      80,
		height:     24,
		autoFollow: true,
		vp:         &vpInner{},
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// SetContent sets the string content of the viewport.
func (m *Model) SetContent(content string) {
	m.vp.content = content
	if content == "" {
		m.vp.totalLines = 0
	} else {
		m.vp.totalLines = strings.Count(content, "\n") + 1
	}
	if m.autoFollow {
		m.GotoBottom()
	}
}

// Content returns the current content string.
func (m *Model) Content() string {
	return m.vp.content
}

// View renders the visible portion of the viewport as a string.
func (m *Model) View() string {
	if m.vp.content == "" || m.height <= 0 {
		return ""
	}
	lines := strings.Split(m.vp.content, "\n")
	start := m.yOffset
	if start > len(lines) {
		start = len(lines)
	}
	end := start + m.height
	if end > len(lines) {
		end = len(lines)
	}
	visible := lines[start:end]
	result := strings.Join(visible, "\n")
	// Pad to fill height
	for len(strings.Split(result, "\n")) < m.height {
		result += "\n"
	}
	return result
}

// SetSize updates the viewport dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.clampYOffset()
}

// SetWidth sets the viewport width.
func (m *Model) SetWidth(w int) {
	m.width = w
}

// SetHeight sets the viewport height.
func (m *Model) SetHeight(h int) {
	m.height = h
	m.clampYOffset()
}

// Width returns the viewport width.
func (m *Model) Width() int {
	return m.width
}

// Height returns the viewport height.
func (m *Model) Height() int {
	return m.height
}

// ScrollUp scrolls up by n lines.
func (m *Model) ScrollUp(n int) {
	m.autoFollow = false
	m.yOffset -= n
	if m.yOffset < 0 {
		m.yOffset = 0
	}
}

// ScrollDown scrolls down by n lines.
func (m *Model) ScrollDown(n int) {
	m.yOffset += n
	max := m.maxYOffset()
	if m.yOffset > max {
		m.yOffset = max
	}
	if m.yOffset >= max {
		m.autoFollow = true
	}
}

// GotoBottom scrolls to the bottom and enables auto-follow.
func (m *Model) GotoBottom() {
	m.yOffset = m.maxYOffset()
	m.autoFollow = true
}

// GotoTop scrolls to the top.
func (m *Model) GotoTop() {
	m.yOffset = 0
	m.autoFollow = false
}

// AtBottom returns true if the viewport is at the bottom.
func (m *Model) AtBottom() bool {
	return m.yOffset >= m.maxYOffset()
}

// YOffset returns the current vertical scroll offset.
func (m *Model) YOffset() int {
	return m.yOffset
}

// SetYOffset sets the vertical scroll offset.
func (m *Model) SetYOffset(n int) {
	m.yOffset = n
	m.clampYOffset()
}

// TotalLineCount returns the total number of content lines.
func (m *Model) TotalLineCount() int {
	return m.vp.totalLines
}

// VisibleLineCount returns the number of visible lines.
func (m *Model) VisibleLineCount() int {
	return m.height
}

// AutoFollow returns whether auto-follow is enabled.
func (m *Model) AutoFollow() bool {
	return m.autoFollow
}

// SetAutoFollow enables/disables auto-follow.
func (m *Model) SetAutoFollow(b bool) {
	m.autoFollow = b
}

// ScrollIndicatorStyle returns a scroll indicator string.
// Returns empty string if at bottom with auto-follow.
func (m *Model) ScrollIndicatorStyle() string {
	if m.AtBottom() && m.autoFollow {
		return ""
	}
	total := m.TotalLineCount()
	if total <= m.height {
		return ""
	}
	return "▼"
}

// Update handles a bubbletea message and returns the updated model + cmd.
// Handles KeyPressMsg for scroll navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyUp:
			m.ScrollUp(1)
		case tea.KeyDown:
			m.ScrollDown(1)
		case tea.KeyPgUp:
			m.ScrollUp(m.height)
		case tea.KeyPgDn:
			m.ScrollDown(m.height)
		case tea.KeyHome:
			m.GotoTop()
		case tea.KeyEnd:
			m.GotoBottom()
		}
	case tea.MouseWheelMsg:
		if msg.Up {
			m.ScrollUp(1)
		} else if msg.Down {
			m.ScrollDown(1)
		}
	}
	return m, nil
}

// Init returns the initial command (nil for viewport).
func (m Model) Init() tea.Cmd {
	return nil
}

// ─── Internal helpers ───

func (m *Model) maxYOffset() int {
	max := m.vp.totalLines - m.height
	if max < 0 {
		max = 0
	}
	return max
}

func (m *Model) clampYOffset() {
	max := m.maxYOffset()
	if m.yOffset > max {
		m.yOffset = max
	}
	if m.yOffset < 0 {
		m.yOffset = 0
	}
}