package component

import (
	"sync"
	"unicode/utf8"

	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/theme"
)

// BannerVariant controls the visual style of a Banner.
type BannerVariant int

const (
	BannerNeutral  BannerVariant = iota
	BannerInfo
	BannerSuccess
	BannerWarning
	BannerDanger
)

// Banner renders a full-width inline banner with an icon and message.
// Unlike Callout (which has a colored left bar), Banner uses a full
// background tint for stronger visual emphasis. Useful for important
// announcements, AI model status changes, or connection state.
//
// Thread-safe.
type Banner struct {
	BaseComponent
	mu       sync.Mutex
	variant  BannerVariant
	message  string
	action   string // optional action hint (e.g., "Press R to retry")
	dismissed bool
}

// NewBanner creates a banner with the given variant and message.
func NewBanner(variant BannerVariant, message string) *Banner {
	return &Banner{
		BaseComponent: BaseComponent{id: GenerateID("banner")},
		variant:       variant,
		message:       message,
	}
}

// SetMessage changes the banner text.
func (b *Banner) SetMessage(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.message = s
}

// SetVariant changes the banner style.
func (b *Banner) SetVariant(v BannerVariant) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.variant = v
}

// SetAction sets an optional action hint shown at the right.
func (b *Banner) SetAction(s string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.action = s
}

// Message returns the current message.
func (b *Banner) Message() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.message
}

// IsDismissed returns whether the banner has been dismissed.
func (b *Banner) IsDismissed() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dismissed
}

// Dismiss hides the banner.
func (b *Banner) Dismiss() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dismissed = true
}

// Show un-dismisses the banner.
func (b *Banner) Show() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.dismissed = false
}

// Measure returns the desired size (1 row, full width).
func (b *Banner) Measure(cs Constraints) Size {
	maxW := cs.MaxWidth
	if maxW <= 0 {
		maxW = 60
	}
	return Size{W: maxW, H: 1}
}

// bannerStyle returns icon, fg, bg for the variant.
func bannerStyle(v BannerVariant, th *theme.Theme) (string, buffer.Color, buffer.Color) {
	switch v {
	case BannerInfo:
		return "\u2139", th.Accent, th.Accent // ℹ
	case BannerSuccess:
		return "\u2714", th.Success, th.Success // ✔
	case BannerWarning:
		return "\u26a0", th.Warning, th.Warning // ⚠
	case BannerDanger:
		return "\u2716", th.Error, th.Error // ✖
	default:
		return "\u25cf", th.Muted, th.Muted // ●
	}
}

// Paint renders the banner.
func (b *Banner) Paint(buf *buffer.Buffer) {
	b.mu.Lock()
	dismissed := b.dismissed
	variant := b.variant
	msg := b.message
	action := b.action
	b.mu.Unlock()

	if dismissed {
		return
	}

	bd := b.Bounds()
	if bd.W <= 0 || bd.H <= 0 {
		return
	}

	th := theme.Get()
	icon, fg, bg := bannerStyle(variant, th)

	// Full background tint
	for x := bd.X; x < bd.X+bd.W; x++ {
		buf.SetCell(x, bd.Y, buffer.Cell{
			Rune:  ' ',
			Width: 1,
			Bg:    bg,
		})
	}

	x := bd.X
	iconStyle := buffer.Style{Fg: th.Bg, Bg: bg}
	msgStyle := buffer.Style{Fg: th.Bg, Bg: bg}
	actionStyle := buffer.Style{Fg: fg, Bg: bg}

	// Icon
	x += buf.DrawText(x, bd.Y, " "+icon+" ", iconStyle)

	// Message — draw directly if it fits, truncate only if needed
	msgW := utf8.RuneCountInString(msg)
	availW := bd.W - (x - bd.X)
	if action != "" {
		availW -= utf8.RuneCountInString(action) + 2
	}
	if msgW <= availW {
		buf.DrawText(x, bd.Y, msg, msgStyle)
	} else if availW > 2 {
		// Truncation path (rare) — still 1 alloc
		buf.DrawText(x, bd.Y, truncateRunes(msg, availW-1)+"\u2026", msgStyle)
	}

	// Action hint right-aligned
	if action != "" {
		actW := utf8.RuneCountInString(action)
		ax := bd.X + bd.W - actW - 1
		if ax > x {
			buf.DrawText(ax, bd.Y, " "+action, actionStyle)
		}
	}
}
