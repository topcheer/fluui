package app

import (
	"sync"

	"github.com/topcheer/fluui/internal/term"
)

// maxUndoStates limits how many undo snapshots are kept.
const maxUndoStates = 100

// undoState is a snapshot of InputLine text + cursor at a moment.
type undoState struct {
	buf    []rune
	cursor int
}

// undoStack holds undo/redo stacks for an InputLine.
// Before any mutation, saveUndo() pushes the current state.
// Undo pops from undo → redo; redo reverses.
type undoStack struct {
	mu        sync.Mutex
	undoList  []undoState
	redoList  []undoState
}

// newUndoStack creates an undo/redo stack with default max history.
func newUndoStack() *undoStack {
	return &undoStack{}
}

// snapshot captures a defensive copy of buf + cursor.
func (u *undoStack) snapshot(i *InputLine) undoState {
	buf := make([]rune, len(i.buf))
	copy(buf, i.buf)
	return undoState{buf: buf, cursor: i.cursor}
}

// saveUndo pushes the current InputLine state onto the undo stack
// and clears the redo stack. Call BEFORE any text mutation.
func (u *undoStack) saveUndo(i *InputLine) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.undoList = append(u.undoList, u.snapshot(i))
	if len(u.undoList) > maxUndoStates {
		u.undoList = u.undoList[1:]
	}
	u.redoList = u.redoList[:0]
}

// performUndo restores the previous state, pushes current to redo.
func (u *undoStack) performUndo(i *InputLine) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.undoList) == 0 {
		return false
	}
	u.redoList = append(u.redoList, u.snapshot(i))

	prev := u.undoList[len(u.undoList)-1]
	u.undoList = u.undoList[:len(u.undoList)-1]

	i.buf = make([]rune, len(prev.buf))
	copy(i.buf, prev.buf)
	i.cursor = prev.cursor
	i.historyIdx = -1
	return true
}

// performRedo restores the next state, pushes current to undo.
func (u *undoStack) performRedo(i *InputLine) bool {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.redoList) == 0 {
		return false
	}
	u.undoList = append(u.undoList, u.snapshot(i))

	next := u.redoList[len(u.redoList)-1]
	u.redoList = u.redoList[:len(u.redoList)-1]

	i.buf = make([]rune, len(next.buf))
	copy(i.buf, next.buf)
	i.cursor = next.cursor
	i.historyIdx = -1
	return true
}

func (u *undoStack) canUndo() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.undoList) > 0
}

func (u *undoStack) canRedo() bool {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.redoList) > 0
}

func (u *undoStack) clearAll() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.undoList = u.undoList[:0]
	u.redoList = u.redoList[:0]
}

func (u *undoStack) undoDepth() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.undoList)
}

func (u *undoStack) redoDepth() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.redoList)
}

// --- InputLine Undo/Redo public API ---

// saveUndo snapshots the current state before a mutation.
func (i *InputLine) saveUndo() {
	if i.undo != nil {
		i.undo.saveUndo(i)
	}
}

// Undo restores the previous text state. Returns true if performed.
func (i *InputLine) Undo() bool {
	if i.undo == nil {
		return false
	}
	return i.undo.performUndo(i)
}

// Redo re-applies a previously undone state. Returns true if performed.
func (i *InputLine) Redo() bool {
	if i.undo == nil {
		return false
	}
	return i.undo.performRedo(i)
}

// CanUndo returns true if undo is available.
func (i *InputLine) CanUndo() bool {
	if i.undo == nil {
		return false
	}
	return i.undo.canUndo()
}

// CanRedo returns true if redo is available.
func (i *InputLine) CanRedo() bool {
	if i.undo == nil {
		return false
	}
	return i.undo.canRedo()
}

// UndoCount returns the number of undo states.
func (i *InputLine) UndoCount() int {
	if i.undo == nil {
		return 0
	}
	return i.undo.undoDepth()
}

// RedoCount returns the number of redo states.
func (i *InputLine) RedoCount() int {
	if i.undo == nil {
		return 0
	}
	return i.undo.redoDepth()
}

// ClearUndoHistory empties all undo/redo state.
func (i *InputLine) ClearUndoHistory() {
	if i.undo != nil {
		i.undo.clearAll()
	}
}

// handleUndoKey checks for Ctrl+Z (undo) / Ctrl+Shift+Z or Ctrl+Y (redo).
// Returns true if the key was consumed.
func (i *InputLine) handleUndoKey(key *term.KeyEvent) bool {
	if key.Modifiers&term.ModCtrl == 0 || key.Rune == 0 {
		return false
	}

	// Ctrl+Z = undo
	if key.Rune == 'z' {
		i.Undo()
		return true
	}

	// Ctrl+Shift+Z = redo (uppercase Z with Ctrl+Shift)
	// Ctrl+Y = redo (alternative, vim-style)
	if key.Rune == 'Z' || key.Rune == 'y' {
		i.Redo()
		return true
	}

	return false
}
