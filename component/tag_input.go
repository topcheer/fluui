package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// TagInput is an input for managing a collection of tags/chips.
// Users type text and press Enter or comma to add a tag; Backspace on
// empty input removes the last tag. Common in AI prompt builders,
// metadata editors, and filter UIs.
//
// Thread-safe. Zero-alloc Paint.
type TagInput struct {
	BaseComponent
	mu sync.Mutex

	tags     []string
	input    string
	placeholder string
	maxTags  int
}

// NewTagInput creates a tag input.
func NewTagInput(placeholder string) *TagInput {
	return &TagInput{
		BaseComponent: BaseComponent{id: GenerateID("taginput")},
		placeholder:   placeholder,
		maxTags:       0, // 0 = unlimited
	}
}

// Tags returns a copy of the current tags.
func (t *TagInput) Tags() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, len(t.tags))
	copy(out, t.tags)
	return out
}

// SetTags replaces all tags.
func (t *TagInput) SetTags(tags []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tags = tags
}

// AddTag adds a tag if it's non-empty and not a duplicate.
func (t *TagInput) AddTag(tag string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if tag == "" {
		return false
	}
	if t.maxTags > 0 && len(t.tags) >= t.maxTags {
		return false
	}
	for _, existing := range t.tags {
		if existing == tag {
			return false // duplicate
		}
	}
	t.tags = append(t.tags, tag)
	return true
}

// RemoveTag removes the tag at the given index.
func (t *TagInput) RemoveTag(idx int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if idx < 0 || idx >= len(t.tags) {
		return
	}
	t.tags = append(t.tags[:idx], t.tags[idx+1:]...)
}

// RemoveLast removes the last tag. Returns the removed tag or "".
func (t *TagInput) RemoveLast() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.tags) == 0 {
		return ""
	}
	last := t.tags[len(t.tags)-1]
	t.tags = t.tags[:len(t.tags)-1]
	return last
}

// TagCount returns the number of tags.
func (t *TagInput) TagCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.tags)
}

// SetMaxTags limits the maximum number of tags (0 = unlimited).
func (t *TagInput) SetMaxTags(n int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxTags = n
}

// Input returns the current input text.
func (t *TagInput) Input() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.input
}

// SetInput replaces the input text.
func (t *TagInput) SetInput(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.input = s
}

// CommitInput adds the current input as a tag and clears the input.
// Returns true if a tag was added.
func (t *TagInput) CommitInput() bool {
	t.mu.Lock()
	input := t.input
	t.input = ""
	t.mu.Unlock()
	return t.AddTag(input)
}

// Clear removes all tags.
func (t *TagInput) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tags = nil
	t.input = ""
}

// Measure returns the desired size (always 1 row tall).
func (t *TagInput) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 40
	}
	return Size{W: maxW, H: 1}
}

// Paint renders the tag input.
func (t *TagInput) Paint(buf *buffer.Buffer) {
	t.mu.Lock()
	tags := t.tags
	input := t.input
	placeholder := t.placeholder
	t.mu.Unlock()

	b := t.Bounds()
	if b.W <= 0 || b.H <= 0 {
		return
	}

	th := theme.Get()
	tagStyle := buffer.Style{Fg: th.Bg, Bg: th.Accent}
	inputStyle := buffer.Style{Fg: th.Fg}
	mutedStyle := buffer.Style{Fg: th.Muted}
	cursorStyle := buffer.Style{Fg: th.Accent}

	x := b.X
	y := b.Y
	maxX := b.X + b.W

	// Draw tags
	for _, tag := range tags {
		tagText := tag + " "
		tagW := utf8.RuneCountInString(tagText)
		if x+tagW+1 > maxX {
			break // no room
		}
		// Draw tag background
		for i := 0; i < tagW; i++ {
			buf.SetCell(x+i, y, buffer.Cell{
				Rune:  ' ',
				Width: 1,
				Bg:    th.Accent,
			})
		}
		// Draw tag text on top
		x = buf.DrawText(x, y, tagText, tagStyle)
	}

	// Draw input or placeholder
	if input != "" {
		x = buf.DrawText(x, y, input, inputStyle)
	} else if len(tags) == 0 && placeholder != "" {
		buf.DrawText(x, y, placeholder, mutedStyle)
	}

	// Draw cursor at end of input
	if x < maxX {
		buf.SetCell(x, y, buffer.Cell{
			Rune:  '\u2588', // █ block cursor
			Width: 1,
			Fg:    th.Accent,
			Bg:    cursorStyle.Bg,
		})
	}
}
