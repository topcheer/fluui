package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIModelRankList: AI Model Performance Ranking ───
//
// AIModelRankList renders a ranked list of AI models with scores and
// delta indicators. Each row shows rank number, model name, score,
// and change indicator. Useful for model comparison dashboards.
//
// Usage:
//
//	rl := NewAIModelRankList()
//	rl.AddModel("GPT-4o", 92, 1)  // name, score, rankChange (+1)
//	rl.AddModel("Claude-3.5", 90, -1)
//	rl.Paint(buf)

// ModelRankStyle holds styling.
type ModelRankStyle struct {
	Rank  buffer.Style
	Name  buffer.Style
	Score buffer.Style
	Up    buffer.Style
	Down  buffer.Style
	Same  buffer.Style
}

// DefaultModelRankStyle returns defaults.
func DefaultModelRankStyle() ModelRankStyle {
	return ModelRankStyle{
		Rank:  buffer.Style{Fg: buffer.RGB(148, 163, 184), Flags: buffer.Bold},
		Name:  buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Score: buffer.Style{Fg: buffer.RGB(234, 179, 8), Flags: buffer.Bold},
		Up:    buffer.Style{Fg: buffer.RGB(34, 197, 94)},
		Down:  buffer.Style{Fg: buffer.RGB(239, 68, 68)},
		Same:  buffer.Style{Fg: buffer.RGB(100, 116, 139)},
	}
}

const modelRankMaxEntries = 15

// modelRankEntry holds a single ranked model.
type modelRankEntry struct {
	name       string
	score      int
	rankChange int // +N=up, -N=down, 0=same
}

// AIModelRankList renders a ranked model list.
type AIModelRankList struct {
	BaseComponent
	mu sync.Mutex

	entries [modelRankMaxEntries]modelRankEntry
	count   int
	width   int
	style   ModelRankStyle
}

// NewAIModelRankList creates an AIModelRankList.
func NewAIModelRankList() *AIModelRankList {
	rl := &AIModelRankList{width: 30, style: DefaultModelRankStyle()}
	rl.SetID(GenerateID("modelrank"))
	return rl
}

// AddModel adds a ranked model entry.
func (rl *AIModelRankList) AddModel(name string, score, rankChange int) *AIModelRankList {
	rl.mu.Lock()
	if rl.count < modelRankMaxEntries {
		rl.entries[rl.count] = modelRankEntry{name: name, score: score, rankChange: rankChange}
		rl.count++
	}
	rl.mu.Unlock()
	return rl
}

// Clear removes all entries.
func (rl *AIModelRankList) Clear() *AIModelRankList {
	rl.mu.Lock()
	rl.count = 0
	rl.mu.Unlock()
	return rl
}

// Count returns the number of entries.
func (rl *AIModelRankList) Count() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.count
}

// SetWidth sets the display width.
func (rl *AIModelRankList) SetWidth(w int) *AIModelRankList {
	rl.mu.Lock()
	if w < 15 {
		w = 15
	}
	rl.width = w
	rl.mu.Unlock()
	return rl
}

// SetStyle sets custom style.
func (rl *AIModelRankList) SetStyle(s ModelRankStyle) *AIModelRankList {
	rl.mu.Lock()
	rl.style = s
	rl.mu.Unlock()
	return rl
}

// Measure returns preferred size.
func (rl *AIModelRankList) Measure(cs Constraints) Size {
	rl.mu.Lock()
	h := rl.count
	rl.mu.Unlock()
	if h < 1 {
		h = 1
	}
	w := rl.width
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: h}
}

// Paint renders the model rank list.
func (rl *AIModelRankList) Paint(buf *buffer.Buffer) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b := rl.Bounds()
	x, y := b.X, b.Y

	rankStyle := rl.style.Rank
	nameStyle := rl.style.Name
	scoreStyle := rl.style.Score
	upStyle := rl.style.Up
	downStyle := rl.style.Down
	sameStyle := rl.style.Same

	for i := 0; i < rl.count; i++ {
		entry := rl.entries[i]
		yy := y + i
		if yy >= buf.Height {
			break
		}
		col := x

		// Rank number
		rankStr := itoa(i+1) + "."
		for _, r := range rankStr {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: rankStyle.Fg, Bg: rankStyle.Bg, Flags: rankStyle.Flags, Width: 1})
			col++
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}

		// Model name
		for _, r := range entry.name {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}

		// Score (right-aligned-ish)
		for col < x+rl.width-8 && col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: nameStyle.Fg, Bg: nameStyle.Bg, Flags: nameStyle.Flags, Width: 1})
			col++
		}
		scoreStr := itoa(entry.score)
		for _, r := range scoreStr {
			if col >= buf.Width {
				break
			}
			buf.SetCell(col, yy, buffer.Cell{Rune: r, Fg: scoreStyle.Fg, Bg: scoreStyle.Bg, Flags: scoreStyle.Flags, Width: 1})
			col++
		}

		// Rank change indicator
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: ' ', Fg: sameStyle.Fg, Bg: sameStyle.Bg, Flags: sameStyle.Flags, Width: 1})
			col++
		}
		var changeRune rune
		var changeSt buffer.Style
		if entry.rankChange > 0 {
			changeRune = '↑'
			changeSt = upStyle
		} else if entry.rankChange < 0 {
			changeRune = '↓'
			changeSt = downStyle
		} else {
			changeRune = '→'
			changeSt = sameStyle
		}
		if col < buf.Width {
			buf.SetCell(col, yy, buffer.Cell{Rune: changeRune, Fg: changeSt.Fg, Bg: changeSt.Bg, Flags: changeSt.Flags, Width: 1})
		}
	}
}

// Children returns nil.
func (rl *AIModelRankList) Children() []Component { return nil }
