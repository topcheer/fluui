package app

import (
	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/hit"
	"github.com/topcheer/fluui/internal/term"
)

// MouseHandler maps mouse clicks to hit regions and block interactions.
// It maintains a RegionTree built from the ChatApp's container blocks.
//
// Interactive blocks register their clickable areas during RebuildRegions.
// HandleClick performs hit-testing and dispatches the appropriate action
// (e.g. Toggle for ThinkingBlock and ToolResultBlock headers).
//
// Scrollbar interaction: clicks/drags in the scrollbar column are routed
// directly to the ScrollView before any hit-region testing.
type MouseHandler struct {
	chatApp *ChatApp
	tree    *hit.RegionTree
}

// NewMouseHandler creates a MouseHandler bound to the given ChatApp.
func NewMouseHandler(chatApp *ChatApp) *MouseHandler {
	return &MouseHandler{
		chatApp: chatApp,
		tree:    hit.NewRegionTree(),
	}
}

// RegionTree returns the underlying hit region tree.
func (h *MouseHandler) RegionTree() *hit.RegionTree {
	return h.tree
}

// RebuildRegions clears the region tree and re-registers clickable areas
// for all interactive blocks.
//
// ThinkingBlock: the header row (row 0) is registered as a toggle region.
// ToolResultBlock: the top border row is registered as a toggle region.
//
// This should be called after layout changes (SetBounds) and before
// processing mouse events.
func (h *MouseHandler) RebuildRegions() {
	h.tree.Clear()

	for _, b := range h.chatApp.container.Blocks() {
		bounds := b.Bounds()
		if bounds.W <= 0 || bounds.H <= 0 {
			continue
		}

		switch b.Type() {
		case block.TypeThinking, block.TypeToolResult:
			h.tree.Add(hit.Region{
				ID:      b.ID(),
				BlockID: b.ID(),
				Bounds:  hit.Rect{X: bounds.X, Y: bounds.Y, W: bounds.W, H: 1},
				Action:  hit.Action{Type: hit.ActionToggle},
				Cursor:  hit.CursorPointer,
			})
		}
	}
}

// HandleClick processes a mouse click at screen coordinates (x, y).
// If the click hits a registered region, the corresponding block action
// is performed (e.g. Toggle for ThinkingBlock / ToolResultBlock).
// Returns true if the click was consumed by a region.
func (h *MouseHandler) HandleClick(x, y int) bool {
	region := h.tree.Hit(x, y)
	if region == nil {
		return false
	}

	// Find the owning block and dispatch the action.
	for _, b := range h.chatApp.container.Blocks() {
		if b.ID() != region.BlockID {
			continue
		}
		switch region.Action.Type {
		case hit.ActionToggle:
			toggleBlock(b)
		case hit.ActionCustom:
			if region.Action.Fn != nil {
				region.Action.Fn()
			}
		}
		return true
	}

	return false
}

// Handle processes a full term.MouseEvent, routing scrollbar, click and wheel events.
// Returns true if the event was consumed.
//
// Scrollbar events (click, drag, release in the scrollbar column) are
// handled first, before hit-region testing.
func (h *MouseHandler) Handle(mouse *term.MouseEvent) bool {
	// --- Scrollbar interaction ---
	sv := h.chatApp.scrollView
	barX := sv.ScrollbarColumn()

	if barX >= 0 && mouse.X == barX {
		svBounds := sv.Bounds()
		// Check Y is within scrollbar vertical range.
		if mouse.Y >= svBounds.Y && mouse.Y < svBounds.Y+svBounds.H {
			relY := mouse.Y - svBounds.Y

			switch mouse.Action {
			case term.MouseDown:
				sv.HandleScrollbarDown(relY)
				return true
			case term.MouseDrag:
				sv.HandleScrollbarDrag(relY)
				return true
			case term.MouseUp:
				sv.HandleScrollbarUp()
				return true
			}
		}
	}

	// If we're dragging the scrollbar, consume all drag events even outside the column.
	if sv.IsDragging() && mouse.Action == term.MouseDrag {
		svBounds := sv.Bounds()
		relY := mouse.Y - svBounds.Y
		sv.HandleScrollbarDrag(relY)
		return true
	}

	// Release drag if mouse-up happens anywhere.
	if sv.IsDragging() && mouse.Action == term.MouseUp {
		sv.HandleScrollbarUp()
		return true
	}

	switch mouse.Action {
	case term.MouseDown:
		return h.HandleClick(mouse.X, mouse.Y)
	case term.MouseWheel:
		switch mouse.Button {
		case term.MouseWheelUp:
			h.chatApp.ScrollUp()
			return true
		case term.MouseWheelDown:
			h.chatApp.ScrollDown()
			return true
		}
	}
	return false
}

// toggleBlock type-switches the block and calls Toggle() if the concrete
// type supports it (ThinkingBlock, ToolResultBlock).
func toggleBlock(b block.Block) {
	switch v := b.(type) {
	case *block.ThinkingBlock:
		v.Toggle()
	case *block.ToolResultBlock:
		v.Toggle()
	}
}
