// Package main implements a keyboard-driven todo app using Fluui.
//
// This example showcases:
//   - Table for listing todos with status, priority, and description
//   - TabBar for filtering (All / Active / Done)
//   - StatusBar with item counts and key hints
//   - Inline text entry for adding and editing todos
//   - Checkbox-style toggling with Space
//   - Priority cycling with p
//   - Full keyboard navigation (vim-style + arrows)
//
// Keys:
//   j/k, Up/Down  — navigate todos
//   Space         — toggle done/pending
//   p             — cycle priority (low → medium → high)
//   a             — add new todo (enter text, then Enter to confirm)
//   e             — edit selected todo (enter text, then Enter to confirm)
//   d             — delete selected todo
//   x             — clear all done todos
//   1/2/3         — filter: all / active / done
//   Enter         — confirm add/edit
//   Esc           — cancel add/edit / quit
//   q, Ctrl+C     — quit
package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- Domain ---

type priority int

const (
	prioLow    priority = 0
	prioMedium priority = 1
	prioHigh   priority = 2
)

func (p priority) String() string {
	switch p {
	case prioLow:
		return "Low"
	case prioMedium:
		return "Medium"
	case prioHigh:
		return "High"
	}
	return "?"
}

func (p priority) Symbol() string {
	switch p {
	case prioLow:
		return "▼"
	case prioMedium:
		return "◆"
	case prioHigh:
		return "▲"
	}
	return "?"
}

type todo struct {
	done     bool
	prio     priority
	text     string
	created  time.Time
}

// --- State ---

type appState struct {
	todos      []todo
	cursor     int
	filter     int // 0=all, 1=active, 2=done
	editMode   bool // true when adding/editing text
	editText   []rune
	editingIdx int  // -1 = adding new, >=0 = editing existing
}

func (s *appState) filteredIndices() []int {
	var idxs []int
	for i, t := range s.todos {
		switch s.filter {
		case 0: // all
			idxs = append(idxs, i)
		case 1: // active
			if !t.done {
				idxs = append(idxs, i)
			}
		case 2: // done
			if t.done {
				idxs = append(idxs, i)
			}
		}
	}
	return idxs
}

func (s *appState) activeCount() int {
	count := 0
	for _, t := range s.todos {
		if !t.done {
			count++
		}
	}
	return count
}

func (s *appState) doneCount() int {
	count := 0
	for _, t := range s.todos {
		if t.done {
			count++
		}
	}
	return count
}

func (s *appState) selectedGlobalIdx() int {
	idxs := s.filteredIndices()
	if len(idxs) == 0 || s.cursor < 0 || s.cursor >= len(idxs) {
		return -1
	}
	return idxs[s.cursor]
}

func (s *appState) clampCursor() {
	idxs := s.filteredIndices()
	if len(idxs) == 0 {
		s.cursor = 0
		return
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(idxs) {
		s.cursor = len(idxs) - 1
	}
}

// --- Colors ---

var (
	colorBorder  = buffer.RGB(0x55, 0x55, 0x55)
	colorAccent  = buffer.RGB(0x7d, 0xd3, 0xfc)
	colorGreen   = buffer.RGB(0x50, 0xfa, 0x7b)
	colorYellow  = buffer.RGB(0xf1, 0xfa, 0x8c)
	colorRed     = buffer.RGB(0xff, 0x55, 0x55)
	colorDim     = buffer.RGB(0x62, 0x72, 0xA4)
	colorDone    = buffer.RGB(0x6b, 0x77, 0x8b) // gray for completed
	colorActive  = buffer.RGB(0xe6, 0xe6, 0xe6) // white for active
	colorPrioLow = buffer.RGB(0x6a, 0x9f, 0xd7)
	colorPrioMed = buffer.RGB(0xf1, 0xfa, 0x8c)
	colorPrioHi  = buffer.RGB(0xff, 0x6e, 0x6e)
)

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	state := &appState{
		todos: []todo{
			{done: false, prio: prioHigh, text: "Learn Fluui framework", created: time.Now()},
			{done: false, prio: prioMedium, text: "Build a todo app", created: time.Now()},
			{done: false, prio: prioLow, text: "Write documentation", created: time.Now()},
			{done: true, prio: prioLow, text: "Set up project structure", created: time.Now()},
		},
		cursor:     0,
		filter:     0,
		editingIdx: -1,
	}

	// --- Components ---

	tabBar := component.NewTabBar()
	tabBar.AddTab("all", "1:All")
	tabBar.AddTab("active", "2:Active")
	tabBar.AddTab("done", "3:Done")
	tabBar.SetActive(0)

	statusBar := component.NewStatusBar()
	statusBar.AddLeft("app", " Fluui Todo")
	statusBar.AddCenter("count", "")
	statusBar.AddRight("hint", " [a]dd [e]dit [d]el [Space]toggle [p]rio [x]clear [q]uit ")

	// --- Key handling ---

	base.OnKey(func(k *term.KeyEvent) {
		if state.editMode {
			// Edit mode key handling
			switch {
			case k.Key == term.KeyEscape:
				state.editMode = false
				state.editText = nil
				state.editingIdx = -1

			case k.Key == term.KeyEnter:
				text := strings.TrimSpace(string(state.editText))
				state.editMode = false
				state.editText = nil
				if text != "" {
					if state.editingIdx >= 0 && state.editingIdx < len(state.todos) {
						state.todos[state.editingIdx].text = text
					} else {
						state.todos = append(state.todos, todo{
							done:    false,
							prio:    prioMedium,
							text:    text,
							created: time.Now(),
						})
					}
				}
				state.editingIdx = -1

			case k.Key == term.KeyBackspace:
				if len(state.editText) > 0 {
					state.editText = state.editText[:len(state.editText)-1]
				}

			case k.Rune == 'u' && k.Modifiers&term.ModCtrl != 0:
				state.editText = nil

			case k.Key == term.KeyUnknown && k.Rune != 0:
				// Printable character
				state.editText = append(state.editText, k.Rune)
			}
			base.MarkDirty()
			return
		}

		// Normal mode key handling
		switch {
		case k.Key == term.KeyEscape || k.Rune == 'q':
			base.Quit()

		case k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0:
			base.Quit()

		case k.Rune == '1':
			state.filter = 0
			tabBar.SetActive(0)
			state.clampCursor()

		case k.Rune == '2':
			state.filter = 1
			tabBar.SetActive(1)
			state.clampCursor()

		case k.Rune == '3':
			state.filter = 2
			tabBar.SetActive(2)
			state.clampCursor()

		case k.Rune == 'a':
			state.editMode = true
			state.editText = nil
			state.editingIdx = -1

		case k.Rune == 'e':
			gidx := state.selectedGlobalIdx()
			if gidx >= 0 {
				state.editMode = true
				state.editText = []rune(state.todos[gidx].text)
				state.editingIdx = gidx
			}

		case k.Rune == 'd':
			gidx := state.selectedGlobalIdx()
			if gidx >= 0 {
				state.todos = append(state.todos[:gidx], state.todos[gidx+1:]...)
				state.clampCursor()
			}

		case k.Rune == 'x':
			var remaining []todo
			for _, t := range state.todos {
				if !t.done {
					remaining = append(remaining, t)
				}
			}
			state.todos = remaining
			state.clampCursor()

		case k.Rune == ' ':
			gidx := state.selectedGlobalIdx()
			if gidx >= 0 {
				state.todos[gidx].done = !state.todos[gidx].done
			}

		case k.Rune == 'p':
			gidx := state.selectedGlobalIdx()
			if gidx >= 0 {
				state.todos[gidx].prio = (state.todos[gidx].prio + 1) % 3
			}

		case k.Key == term.KeyUp || k.Rune == 'k':
			idxs := state.filteredIndices()
			if len(idxs) > 0 && state.cursor > 0 {
				state.cursor--
			}

		case k.Key == term.KeyDown || k.Rune == 'j':
			idxs := state.filteredIndices()
			if state.cursor < len(idxs)-1 {
				state.cursor++
			}
		}

		base.MarkDirty()
	})

	// --- Rendering ---

	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()

		// Layout: TabBar (1) + separator (1) + content (h-4) + separator (1) + StatusBar (1)
		// If editing, reserve 1 line for edit input at bottom of content

		tabBar.SetBounds(component.Rect{X: 0, Y: 0, W: w, H: 1})
		tabBar.Paint(buf)

		// Update status bar counts
		statusBar.SetItemText("count", fmt.Sprintf(" %d total / %d active / %d done ",
			len(state.todos), state.activeCount(), state.doneCount()))
		statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})
		statusBar.Paint(buf)

		// Content area
		contentY := 2
		contentH := h - 3

		// Separator under tab bar
		sepStyle := buffer.Style{Fg: colorBorder}
		for x := 0; x < w; x++ {
			buf.SetCell(x, 1, buffer.NewCell('─', sepStyle))
		}
		for x := 0; x < w; x++ {
			buf.SetCell(x, h-2, buffer.NewCell('─', sepStyle))
		}

		if state.editMode {
			// Edit mode: show input line at bottom
			prompt := "> Add: "
			if state.editingIdx >= 0 {
				prompt = "> Edit: "
			}
			editY := contentY + contentH - 1
			if editY < contentY {
				editY = contentY
			}
			editStyle := buffer.Style{Fg: colorAccent, Flags: buffer.Bold}
			buf.DrawText(1, editY, prompt, editStyle)
			textX := 1 + len(prompt)
			for i, r := range state.editText {
				if textX+i < w-1 {
					buf.SetCell(textX+i, editY, buffer.NewCell(r, editStyle))
				}
			}
			// Cursor
			cursorX := textX + len(state.editText)
			if cursorX < w-1 {
				buf.SetCell(cursorX, editY, buffer.NewCell(' ', buffer.Style{
					Fg:    buffer.RGB(0x00, 0x00, 0x00),
					Bg:    colorAccent,
					Flags: buffer.Reverse,
				}))
			}
			contentH--
		}

		// Draw todo list
		idxs := state.filteredIndices()
		if len(idxs) == 0 {
			emptyStyle := buffer.Style{Fg: colorDim}
			msg := "No todos. Press 'a' to add one."
			buf.DrawText(w/2-len(msg)/2, contentY+contentH/2, msg, emptyStyle)
			return
		}

		for displayIdx, globalIdx := range idxs {
			if displayIdx >= contentH {
				break
			}
			t := state.todos[globalIdx]
			y := contentY + displayIdx

			isSelected := displayIdx == state.cursor

			// Background highlight for selected
			if isSelected {
				bgStyle := buffer.Style{Bg: buffer.RGB(0x2a, 0x2d, 0x3e), Fg: colorAccent}
				for x := 0; x < w; x++ {
					buf.SetCell(x, y, buffer.NewCell(' ', bgStyle))
				}
			}

			// Checkbox
			var checkRune rune
			var checkStyle buffer.Style
			if t.done {
				checkRune = '✓'
				checkStyle = buffer.Style{Fg: colorGreen, Flags: buffer.Bold}
			} else {
				checkRune = '○'
				checkStyle = buffer.Style{Fg: colorDim}
			}
			if isSelected {
				checkStyle.Fg = colorAccent
				checkStyle.Flags |= buffer.Bold
			}
			buf.SetCell(1, y, buffer.NewCell(checkRune, checkStyle))

			// Priority
			var prioStyle buffer.Style
			switch t.prio {
			case prioLow:
				prioStyle = buffer.Style{Fg: colorPrioLow}
			case prioMedium:
				prioStyle = buffer.Style{Fg: colorPrioMed}
			case prioHigh:
				prioStyle = buffer.Style{Fg: colorPrioHi, Flags: buffer.Bold}
			}
			buf.SetCell(4, y, buffer.NewCell([]rune(t.prio.Symbol())[0], prioStyle))

			// Text
			var textStyle buffer.Style
			if t.done {
				textStyle = buffer.Style{Fg: colorDone, Flags: buffer.Dim}
			} else {
				textStyle = buffer.Style{Fg: colorActive}
			}
			if isSelected {
				textStyle.Flags |= buffer.Bold
				textStyle.Fg = colorAccent
			}

			// Truncate text to fit
			maxTextW := w - 8
			if maxTextW < 10 {
				maxTextW = 10
			}
			text := t.text
			if len([]rune(text)) > maxTextW {
				text = string([]rune(text)[:maxTextW-3]) + "..."
			}
			buf.DrawText(7, y, text, textStyle)

			// Done items get strikethrough (use overline chars)
			if t.done {
				for x := 7; x < 7+len([]rune(text)); x++ {
					if x < w-1 {
						cell := buf.GetCell(x, y)
						cell.Fg = colorDone
						buf.SetCell(x, y, cell)
					}
				}
			}
		}

		// Show scroll indicator if needed
		if len(idxs) > contentH {
			arrowStyle := buffer.Style{Fg: colorDim}
			buf.SetCell(w-1, contentY, buffer.NewCell('▲', arrowStyle))
			buf.SetCell(w-1, contentY+contentH-1, buffer.NewCell('▼', arrowStyle))
		}
	})

	base.Run()
}

// Unused but kept for potential future use
var _ = strconv.Itoa
