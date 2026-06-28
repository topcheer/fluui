package component

import (
	"regexp"
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// LinkRange describes a detected URL within rendered text.
type LinkRange struct {
	URL     string
	Text    string
	StartX  int // column where the link text starts (0-based)
	EndX    int // column where the link text ends (exclusive)
	Y       int // row index
	LineIdx int // index of the source line
}

// LinkStyle holds the visual styling for links.
type LinkStyle struct {
	Normal    buffer.Style // default link appearance
	Underline buffer.Style // underlined variant for hover/active
}

// DefaultLinkStyle returns a sensible default link style (blue + underline).
func DefaultLinkStyle() LinkStyle {
	blue := buffer.RGB(0x44, 0x8A, 0xFF)
	return LinkStyle{
		Normal: buffer.Style{
			Fg:    blue,
			Flags: buffer.Underline,
		},
		Underline: buffer.Style{
			Fg:    blue,
			Flags: buffer.Underline | buffer.Bold,
		},
	}
}

// urlPattern matches common URL schemes.
// Matches: http://, https://, ftp://, www. (auto-promoted), git://, ssh://
var urlPattern = regexp.MustCompile(
	`(?:https?|ftp|git|ssh)://[^\s<>"'` + "`" + `)]+|www\.[^\s<>"'` + "`" + `)]+\.[^\s<>"'` + "`" + `)]+`,
)

// LinkManager detects URLs in text, renders them as clickable links in a buffer,
// and provides hit-testing for mouse clicks.
type LinkManager struct {
	mu      sync.RWMutex
	links   []LinkRange
	style   LinkStyle
	onClick func(url string) // callback when a link is clicked
	enabled bool
}

// NewLinkManager creates a LinkManager with default styling.
func NewLinkManager() *LinkManager {
	return &LinkManager{
		enabled: true,
		style: DefaultLinkStyle(),
	}
}

// SetStyle updates the link rendering style.
func (lm *LinkManager) SetStyle(s LinkStyle) {
	lm.mu.Lock()
	lm.style = s
	lm.mu.Unlock()
}

// Style returns the current link style.
func (lm *LinkManager) Style() LinkStyle {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.style
}

// SetOnClick sets the callback invoked when a link is clicked.
// The callback receives the clicked URL.
func (lm *LinkManager) SetOnClick(fn func(url string)) {
	lm.mu.Lock()
	lm.onClick = fn
	lm.mu.Unlock()
}

// Links returns a copy of all currently tracked link ranges.
func (lm *LinkManager) Links() []LinkRange {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	out := make([]LinkRange, len(lm.links))
	copy(out, lm.links)
	return out
}

// LinkCount returns the number of tracked links.
func (lm *LinkManager) LinkCount() int {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return len(lm.links)
}

// Clear removes all tracked links.
func (lm *LinkManager) Clear() {
	lm.mu.Lock()
	lm.links = lm.links[:0]
	lm.mu.Unlock()
}

// DetectLinks scans a line of text and returns all URL ranges found.
// The lineIdx parameter tags each range with its source line index.
// The yOffset parameter sets the Y coordinate for each range.
func DetectLinks(text string, lineIdx, yOffset int) []LinkRange {
	matches := urlPattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}
	ranges := make([]LinkRange, 0, len(matches))
	for _, m := range matches {
		start, end := m[0], m[1]
		url := text[start:end]
		// Promote www. to https://
		displayURL := url
		if len(url) >= 4 && url[:4] == "www." {
			displayURL = "https://" + url
		}
		ranges = append(ranges, LinkRange{
			URL:     displayURL,
			Text:    url,
			StartX:  start,
			EndX:    end,
			Y:       yOffset,
			LineIdx: lineIdx,
		})
	}
	return ranges
}

// ScanText scans multiple lines of text for URLs and stores the results.
// Previous links are cleared.
func (lm *LinkManager) ScanText(lines []string) {
	detected := make([]LinkRange, 0)
	for i, line := range lines {
		detected = append(detected, DetectLinks(line, i, i)...)
	}
	lm.mu.Lock()
	lm.links = detected
	lm.mu.Unlock()
}

// ScanLine scans a single line of text for URLs and adds them to the tracked set.
func (lm *LinkManager) ScanLine(text string, lineIdx, yOffset int) {
	detected := DetectLinks(text, lineIdx, yOffset)
	if len(detected) == 0 {
		return
	}
	lm.mu.Lock()
	lm.links = append(lm.links, detected...)
	lm.mu.Unlock()
}

// AddLink manually adds a link range.
func (lm *LinkManager) AddLink(lr LinkRange) {
	lm.mu.Lock()
	lm.links = append(lm.links, lr)
	lm.mu.Unlock()
}

// LinkAt performs a hit test: returns the link at the given (x, y) position,
// or nil if no link is present there.
func (lm *LinkManager) LinkAt(x, y int) *LinkRange {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	for i := range lm.links {
		lr := &lm.links[i]
		if lr.Y == y && x >= lr.StartX && x < lr.EndX {
			return lr
		}
	}
	return nil
}

// ClickLink attempts to click the link at (x, y). Returns true if a link was
// found and the OnClick callback (if set) was invoked.
func (lm *LinkManager) ClickLink(x, y int) bool {
	lr := lm.LinkAt(x, y)
	if lr == nil {
		return false
	}
	lm.mu.RLock()
	fn := lm.onClick
	lm.mu.RUnlock()
	if fn != nil {
		fn(lr.URL)
	}
	return true
}

// AnnotateBuffer marks cells in the given buffer that correspond to tracked links.
// Cells within link ranges get their Link pointer set and their style updated.
// The startX/startY parameters specify the buffer offset where line 0 begins.
func (lm *LinkManager) AnnotateBuffer(buf *buffer.Buffer, startX, startY int) {
	if buf == nil {
		return
	}
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	style := lm.style.Normal
	for i := range lm.links {
		lr := &lm.links[i]
		y := startY + lr.Y
		if y < 0 || y >= buf.Height {
			continue
		}
		for x := lr.StartX; x < lr.EndX; x++ {
			bx := startX + x
			if bx < 0 || bx >= buf.Width {
				continue
			}
			cell := buf.GetCell(bx, y)
			cell.Link = &buffer.Link{
				URL:  lr.URL,
				Text: lr.Text,
			}
			cell.Fg = style.Fg
			cell.Flags |= style.Flags
			buf.SetCell(bx, y, cell)
		}
	}
}

// HasLinks reports whether any links are currently tracked.
func (lm *LinkManager) HasLinks() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return len(lm.links) > 0
}

// Enabled reports whether link scanning/clicking is active.
func (lm *LinkManager) Enabled() bool {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.enabled
}

// SetEnabled toggles link scanning/clicking.
func (lm *LinkManager) SetEnabled(v bool) {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.enabled = v
}

// FindByURL returns all link ranges matching the given URL.
func (lm *LinkManager) FindByURL(url string) []LinkRange {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	var result []LinkRange
	for _, lr := range lm.links {
		if lr.URL == url {
			result = append(result, lr)
		}
	}
	return result
}

// String returns a debug representation.
func (lm *LinkManager) String() string {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return "LinkManager(links=" + itoa(len(lm.links)) + ")"
}
