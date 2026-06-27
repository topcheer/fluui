// Package hit implements the mouse hit-testing system.
//
// A Region describes a rectangular area on screen that can respond to mouse
// input. Regions are collected into a RegionTree which is queried by the event
// loop to determine which region (if any) is under the cursor at a given
// coordinate. Regions are rebuilt every render frame: components register their
// interactive areas during Paint, the tree is queried on each mouse event, and
// the tree is cleared before the next Paint cycle.
package hit

// Rect is an axis-aligned integer rectangle in terminal cell coordinates.
// X and Y are the top-left origin; W and H are the width and height in cells.
type Rect struct {
	X, Y, W, H int
}

// Contains reports whether the cell coordinate (x, y) falls inside r.
// A zero-area rectangle (W == 0 or H == 0) contains no points.
func (r Rect) Contains(x, y int) bool {
	if r.W <= 0 || r.H <= 0 {
		return false
	}
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// CursorStyle hints how the mouse cursor should appear over a region.
type CursorStyle uint8

const (
	CursorDefault CursorStyle = iota // Normal arrow cursor
	CursorPointer                    // Hand pointer (over links/buttons)
	CursorText                       // I-beam (over selectable text)
)

// Action describes what should happen when a region is clicked.
type Action struct {
	Type ActionType
	URL  string // For ActionOpenURL: the URL to open
	Text string // For ActionCopy: the text to copy to clipboard
	Fn   func() // For ActionCustom: the callback to invoke
}

// ActionType identifies the kind of interaction a region triggers.
type ActionType uint8

const (
	ActionToggle  ActionType = iota // Toggle expand/collapse (e.g. thinking block)
	ActionOpenURL                   // Open an external URL
	ActionCopy                      // Copy text to the clipboard
	ActionCustom                    // Invoke a custom callback (Action.Fn)
)

// Region describes a single clickable area on screen.
type Region struct {
	ID      string      // Unique identifier within a render frame
	Bounds  Rect        // Screen-space bounding rectangle
	Action  Action      // What happens when clicked
	Cursor  CursorStyle // Cursor hint for hover feedback
	BlockID string      // Owning block ID (for debugging and routing)
}
