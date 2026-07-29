package component

import (
	"sync"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── AIEmojiReaction: AI Response Sentiment Emoji Badge ───
//
// AIEmojiReaction renders a compact badge showing the detected sentiment
// of an AI response as an emoji, with an optional confidence percentage.
// Useful for quick visual feedback in chat interfaces.
//
// Usage:
//
//	er := NewAIEmojiReaction()
//	er.SetSentiment(ERSentimentPositive, 85)
//	er.Paint(buf)

// ERSentiment represents detected sentiment.
type ERSentiment int

const (
	ERSentimentPositive ERSentiment = 0
	ERSentimentNeutral  ERSentiment = 1
	ERSentimentNegative ERSentiment = 2
	ERSentimentMixed    ERSentiment = 3
)

var sentimentEmojis = [...]rune{'😊', '😐', '😔', '🤔'}
var sentimentLabels = [...]string{"Positive", "Neutral", "Negative", "Mixed"}

// AIEmojiReactionStyle holds styling.
type AIEmojiReactionStyle struct {
	Emoji   buffer.Style
	Label   buffer.Style
	Pct     buffer.Style
	Bracket buffer.Style
}

// DefaultAIEmojiReactionStyle returns defaults.
func DefaultAIEmojiReactionStyle() AIEmojiReactionStyle {
	return AIEmojiReactionStyle{
		Emoji:   buffer.Style{Flags: buffer.Bold},
		Label:   buffer.Style{Fg: buffer.RGB(226, 232, 240)},
		Pct:     buffer.Style{Fg: buffer.RGB(148, 163, 184)},
		Bracket: buffer.Style{Fg: buffer.RGB(71, 85, 105)},
	}
}

// AIEmojiReaction renders a sentiment emoji badge.
type AIEmojiReaction struct {
	BaseComponent
	mu sync.Mutex

	sentiment  ERSentiment
	confidence int
	style      AIEmojiReactionStyle
	// cached
	pctStr   string
	labelStr string
}

// NewAIEmojiReaction creates an AIEmojiReaction.
func NewAIEmojiReaction() *AIEmojiReaction {
	er := &AIEmojiReaction{sentiment: ERSentimentNeutral, confidence: 50, style: DefaultAIEmojiReactionStyle()}
	er.SetID(GenerateID("emojireact"))
	er.recomputeLocked()
	return er
}

// SetSentiment sets the sentiment type and confidence (0-100).
func (er *AIEmojiReaction) SetSentiment(s ERSentiment, confidence int) *AIEmojiReaction {
	er.mu.Lock()
	if int(s) < 0 || int(s) >= len(sentimentLabels) {
		s = ERSentimentNeutral
	}
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}
	er.sentiment = s
	er.confidence = confidence
	er.recomputeLocked()
	er.mu.Unlock()
	return er
}

func (er *AIEmojiReaction) recomputeLocked() {
	er.pctStr = itoa(er.confidence) + "%"
	er.labelStr = sentimentLabels[er.sentiment]
}

// Sentiment returns the current sentiment.
func (er *AIEmojiReaction) Sentiment() ERSentiment {
	er.mu.Lock()
	defer er.mu.Unlock()
	return er.sentiment
}

// SetStyle sets custom style.
func (er *AIEmojiReaction) SetStyle(s AIEmojiReactionStyle) *AIEmojiReaction {
	er.mu.Lock()
	er.style = s
	er.mu.Unlock()
	return er
}

// Measure returns preferred size.
func (er *AIEmojiReaction) Measure(cs Constraints) Size {
	w := 20
	if cs.MaxWidth > 0 && w > cs.MaxWidth {
		w = cs.MaxWidth
	}
	return Size{W: w, H: 1}
}

// Paint renders the emoji reaction badge.
func (er *AIEmojiReaction) Paint(buf *buffer.Buffer) {
	er.mu.Lock()
	defer er.mu.Unlock()

	b := er.Bounds()
	x, y := b.X, b.Y

	emojiStyle := er.style.Emoji
	labelStyle := er.style.Label
	pctStyle := er.style.Pct
	bracketStyle := er.style.Bracket

	col := x

	// Emoji
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: sentimentEmojis[er.sentiment], Fg: emojiStyle.Fg, Bg: emojiStyle.Bg, Flags: emojiStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Label
	for _, r := range er.labelStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: labelStyle.Fg, Bg: labelStyle.Bg, Flags: labelStyle.Flags, Width: 1})
		col++
	}

	// Confidence in brackets
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ' ', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: '(', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
		col++
	}
	for _, r := range er.pctStr {
		if col >= buf.Width {
			break
		}
		buf.SetCell(col, y, buffer.Cell{Rune: r, Fg: pctStyle.Fg, Bg: pctStyle.Bg, Flags: pctStyle.Flags, Width: 1})
		col++
	}
	if col < buf.Width {
		buf.SetCell(col, y, buffer.Cell{Rune: ')', Fg: bracketStyle.Fg, Bg: bracketStyle.Bg, Flags: bracketStyle.Flags, Width: 1})
	}
}

// Children returns nil.
func (er *AIEmojiReaction) Children() []Component { return nil }
