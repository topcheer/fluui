// Package theme provides centralized color management for fluui.
//
// A Theme holds all colors used by the TUI: backgrounds, foregrounds,
// borders, status indicators, diff colors, and per-block colors.
// The active theme is a package-level variable that all blocks reference.
// ChatApp.SetTheme() updates it at runtime.
package theme

import (
	"github.com/topcheer/fluui/internal/buffer"
)

// Color is an alias for buffer.Color so theme consumers don't need
// to import the buffer package just for the type.
type Color = buffer.Color

// Theme holds all colors used by the fluui TUI.
type Theme struct {
	Name string

	// Base
	Bg     Color // terminal background
	Fg     Color // terminal foreground
	Accent Color // accent/highlight color

	// Borders
	Border       Color // normal border
	BorderActive Color // focused border
	BorderMuted  Color // muted/secondary border

	// Status
	Success Color
	Error   Color
	Warning Color
	Muted   Color // dimmed text (comments, hints)

	// Code
	CodeBg Color
	CodeFg Color

	// Diff
	DiffAdd  Color // added lines (+)
	DiffDel  Color // removed lines (-)
	DiffMeta Color // metadata (index, ---, +++ )
	DiffHunk Color // hunk headers (@@)
	DiffFile Color // file headers (diff --git)

	// Blocks
	UserMsgBg     Color
	UserMsgFg     Color
	ThinkingBg    Color
	ThinkingFg    Color
	ToolCallBg    Color
	ToolResultBg  Color
	ToolResultFg  Color
	AssistantFg   Color

	// Input
	PromptFg  Color // input prompt color
	Separator Color // separator lines

	// Overlay
	MaskBg Color // modal mask background

	// Search
	SearchBarBg  Color // search bar background
	SearchBarFg  Color // search bar text
	SearchMatch  Color // match count / current match
	SearchNoMatch Color // no-match indicator
}

// Active is the current theme used by all fluui components.
// Update via SetActive(). Defaults to Dracula.
var Active = Dracula()

// currentIndex tracks the position in Builtin() list.
var currentIndex = 0

// SetActive sets the global active theme. All components that
// reference theme.Active will pick up the new colors.
func SetActive(t *Theme) {
	if t != nil {
		Active = t
		// Update currentIndex to match
		for i, b := range Builtin() {
			if b.Name == t.Name {
				currentIndex = i
				break
			}
		}
	}
}

// Get returns the active theme. Convenience function.
func Get() *Theme { return Active }

// Default returns the default theme (Dracula).
func Default() *Theme { return Dracula() }

// Cycle advances to the next built-in theme and returns it.
// Wraps around to the first theme after the last.
func Cycle() *Theme {
	all := Builtin()
	currentIndex = (currentIndex + 1) % len(all)
	Active = all[currentIndex]
	return Active
}

// CycleBack goes to the previous built-in theme and returns it.
// Wraps around to the last theme after the first.
func CycleBack() *Theme {
	all := Builtin()
	currentIndex = (currentIndex - 1 + len(all)) % len(all)
	Active = all[currentIndex]
	return Active
}

// CurrentIndex returns the current position in the Builtin() list.
func CurrentIndex() int { return currentIndex }

// SetByIndex sets the active theme by index in the Builtin() list.
func SetByIndex(i int) {
	all := Builtin()
	if i >= 0 && i < len(all) {
		currentIndex = i
		Active = all[i]
	}
}
