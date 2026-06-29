// Package main implements a keyboard-driven calculator using Fluui.
//
// This example showcases:
//   - Custom rendering with buffer for calculator UI
//   - Full keyboard input handling (digits, operators, Enter, Esc)
//   - State machine for calculator operations
//   - History panel showing previous calculations
//   - StatusBar with key hints
//   - Error handling (division by zero, overflow)
//
// Keys:
//   0-9         — enter digits
//   .           — decimal point
//   + - * /     — arithmetic operators
//   %           — percentage
//   Enter, =    — calculate result
//   Backspace   — delete last digit
//   c, C        — clear all
//   Esc         — clear / quit (double-tap)
//   h           — toggle history panel
//   q, Ctrl+C   — quit
package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	fluui "github.com/topcheer/fluui"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
	"github.com/topcheer/fluui/internal/term"
)

// --- State ---

type op string

const (
	opNone op = ""
	opAdd  op = "+"
	opSub  op = "−"
	opMul  op = "×"
	opDiv  op = "÷"
)

type calcState struct {
	display   string // current display value
	stored    float64 // stored operand
	operation op     // pending operation
	fresh     bool   // true if display should be cleared on next digit
	error     bool   // true if error state
	history   []string
	showHist  bool
}

func newCalcState() *calcState {
	return &calcState{
		display:   "0",
		stored:    0,
		operation: opNone,
		fresh:     true,
	}
}

func (s *calcState) inputDigit(d string) {
	if s.error {
		s.clearAll()
	}
	if s.fresh {
		s.display = d
		s.fresh = false
	} else {
		if s.display == "0" {
			s.display = d
		} else {
			s.display += d
		}
	}
}

func (s *calcState) inputDecimal() {
	if s.error {
		s.clearAll()
	}
	if s.fresh {
		s.display = "0."
		s.fresh = false
		return
	}
	if !strings.Contains(s.display, ".") {
		s.display += "."
	}
}

func (s *calcState) inputOperator(o op) {
	if s.error {
		return
	}
	if s.operation != opNone && !s.fresh {
		s.calculate()
	}
	val, _ := strconv.ParseFloat(s.display, 64)
	s.stored = val
	s.operation = o
	s.fresh = true
}

func (s *calcState) calculate() {
	if s.operation == opNone || s.error {
		return
	}
	val, _ := strconv.ParseFloat(s.display, 64)
	var result float64
	expr := fmt.Sprintf("%g %s %g =", s.stored, s.operation, val)

	switch s.operation {
	case opAdd:
		result = s.stored + val
	case opSub:
		result = s.stored - val
	case opMul:
		result = s.stored * val
	case opDiv:
		if val == 0 {
			s.display = "Error: Division by zero"
			s.error = true
			s.operation = opNone
			s.fresh = true
			return
		}
		result = s.stored / val
	}

	// Check for overflow/NaN
	if math.IsInf(result, 0) || math.IsNaN(result) {
		s.display = "Error: Overflow"
		s.error = true
		s.operation = opNone
		s.fresh = true
		return
	}

	// Format result
	s.display = formatFloat(result)
	s.stored = result
	s.operation = opNone
	s.fresh = true

	// Add to history
	s.history = append(s.history, fmt.Sprintf("%s %s", expr, s.display))
	if len(s.history) > 20 {
		s.history = s.history[1:]
	}
}

func (s *calcState) clearAll() {
	s.display = "0"
	s.stored = 0
	s.operation = opNone
	s.fresh = true
	s.error = false
}

func (s *calcState) backspace() {
	if s.error {
		s.clearAll()
		return
	}
	if s.fresh {
		return
	}
	if len(s.display) > 1 {
		s.display = s.display[:len(s.display)-1]
		// Remove trailing decimal point
		if strings.HasSuffix(s.display, ".") && !strings.Contains(s.display[:len(s.display)-1], ".") {
			// keep it - user might want to add digits after decimal
		}
	} else {
		s.display = "0"
		s.fresh = true
	}
}

func (s *calcState) negate() {
	if s.error {
		return
	}
	val, _ := strconv.ParseFloat(s.display, 64)
	val = -val
	s.display = formatFloat(val)
}

func (s *calcState) percent() {
	if s.error {
		return
	}
	val, _ := strconv.ParseFloat(s.display, 64)
	val = val / 100
	s.display = formatFloat(val)
	s.fresh = true
}

func formatFloat(f float64) string {
	// Use %g for compact representation, but fall back for very large/small
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if math.Abs(f) >= 1e15 || (f != 0 && math.Abs(f) < 1e-10) {
		s = strconv.FormatFloat(f, 'e', 6, 64)
	}
	return s
}

// --- Colors ---

var (
	calcBorder    = buffer.RGB(0x44, 0x47, 0x5a)
	calcAccent    = buffer.RGB(0x7d, 0xd3, 0xfc)
	calcDigit     = buffer.RGB(0x50, 0xfa, 0x7b)
	calcOp        = buffer.RGB(0xff, 0x79, 0xc6)
	calcEquals    = buffer.RGB(0xf1, 0xfa, 0x8c)
	calcDim       = buffer.RGB(0x62, 0x72, 0xA4)
	calcError     = buffer.RGB(0xff, 0x55, 0x55)
	calcHistText  = buffer.RGB(0x8b, 0xe9, 0xfd)
	calcDisplayBg = buffer.RGB(0x1a, 0x1b, 0x26)
)

func main() {
	base, err := fluui.New()
	if err != nil {
		panic(err)
	}
	defer base.Close()

	state := newCalcState()

	statusBar := component.NewStatusBar()
	statusBar.AddLeft("app", " Fluui Calculator")
	statusBar.AddRight("hint", " [h]istory [c]lear [q]uit ")

	// --- Key handling ---

	base.OnKey(func(k *term.KeyEvent) {
		switch {
		// Quit
		case k.Rune == 'q':
			base.Quit()
		case k.Rune == 'c' && k.Modifiers&term.ModCtrl != 0:
			base.Quit()

		// Digits
		case k.Rune >= '0' && k.Rune <= '9':
			state.inputDigit(string(k.Rune))

		// Decimal
		case k.Rune == '.':
			state.inputDecimal()

		// Operators
		case k.Rune == '+':
			state.inputOperator(opAdd)
		case k.Rune == '-':
			state.inputOperator(opSub)
		case k.Rune == '*':
			state.inputOperator(opMul)
		case k.Rune == '/':
			state.inputOperator(opDiv)

		// Calculate
		case k.Key == term.KeyEnter || k.Rune == '=':
			state.calculate()

		// Backspace
		case k.Key == term.KeyBackspace:
			state.backspace()

		// Clear
		case k.Rune == 'c' || k.Rune == 'C':
			state.clearAll()

		// Escape: clear or quit
		case k.Key == term.KeyEscape:
			if state.error || state.display != "0" {
				state.clearAll()
			} else {
				base.Quit()
			}

		// Negate
		case k.Rune == 'n':
			state.negate()

		// Percent
		case k.Rune == '%':
			state.percent()

		// Toggle history
		case k.Rune == 'h':
			state.showHist = !state.showHist
		}

		base.MarkDirty()
	})

	// --- Rendering ---

	base.OnPaint(func(buf *buffer.Buffer) {
		w, h := base.Size()

		// Status bar
		statusBar.SetBounds(component.Rect{X: 0, Y: h - 1, W: w, H: 1})
		statusBar.Paint(buf)

		// History panel (right side, toggleable)
		histW := 0
		if state.showHist && w >= 50 {
			histW = w / 3
			if histW > 30 {
				histW = 30
			}
			if histW < 20 {
				histW = 20
			}
			drawHistoryPanel(buf, w-histW, 0, histW, h-1, state.history)
		}

		// Calculator area
		calcW := w - histW

		// Display area (top 4 lines)
		displayH := 4
		if displayH > h/3 {
			displayH = h / 3
		}

		// Draw display background
		displayBg := buffer.Style{Bg: calcDisplayBg}
		for y := 0; y < displayH; y++ {
			for x := 0; x < calcW; x++ {
				buf.SetCell(x, y, buffer.NewCell(' ', displayBg))
			}
		}

		// Draw operation indicator
		if state.operation != opNone {
			opStyle := buffer.Style{Fg: calcOp, Flags: buffer.Bold}
			opText := fmt.Sprintf(" %s", state.operation)
			buf.DrawText(calcW-len(opText)-2, 1, opText, opStyle)
		}

		// Draw stored value (if any)
		if state.operation != opNone {
			storedStyle := buffer.Style{Fg: calcDim}
			storedText := formatFloat(state.stored)
			if len(storedText) > calcW-4 {
				storedText = storedText[:calcW-4]
			}
			buf.DrawText(2, 0, storedText, storedStyle)
		}

		// Draw main display value
		var displayStyle buffer.Style
		if state.error {
			displayStyle = buffer.Style{Fg: calcError, Flags: buffer.Bold}
		} else {
			displayStyle = buffer.Style{Fg: calcAccent, Flags: buffer.Bold}
		}

		displayText := state.display
		// Truncate if too long
		maxW := calcW - 4
		if len([]rune(displayText)) > maxW {
			displayText = string([]rune(displayText)[:maxW])
		}
		// Right-align the display value
		displayX := calcW - len([]rune(displayText)) - 2
		if displayX < 2 {
			displayX = 2
		}
		buf.DrawText(displayX, displayH-2, displayText, displayStyle)

		// Separator
		sepStyle := buffer.Style{Fg: calcBorder}
		for x := 0; x < calcW; x++ {
			buf.SetCell(x, displayH, buffer.NewCell('─', sepStyle))
		}

		// Button grid area
		gridY := displayH + 1
		gridH := h - 1 - gridY
		if gridH < 3 {
			gridH = 3
		}

		// Define button layout
		// Each button: width = calcW/4, height = gridH/4
		btnW := calcW / 4
		btnH := gridH / 4
		if btnH < 2 {
			btnH = 2
		}
		if btnW < 8 {
			btnW = 8
		}

		buttons := []struct {
			label string
			color buffer.Color
		}{
			{"C", calcError},
			{"±", calcDim},
			{"%", calcDim},
			{"÷", calcOp},
			{"7", calcDigit},
			{"8", calcDigit},
			{"9", calcDigit},
			{"×", calcOp},
			{"4", calcDigit},
			{"5", calcDigit},
			{"6", calcDigit},
			{"−", calcOp},
			{"1", calcDigit},
			{"2", calcDigit},
			{"3", calcDigit},
			{"+", calcOp},
		}

		for i, btn := range buttons {
			row := i / 4
			col := i % 4
			bx := col * btnW
			by := gridY + row*btnH
			if by+btnH > gridY+gridH {
				break
			}
			drawCalcButton(buf, bx, by, btnW, btnH, btn.label, btn.color)
		}

		// Last row: 0 (wide) + . + =
		lastRow := 3
		lastY := gridY + lastRow*btnH
		if lastY+btnH <= gridY+gridH {
			// 0 button takes 2 columns
			drawCalcButton(buf, 0, lastY, btnW*2, btnH, "0", calcDigit)
			// Decimal
			drawCalcButton(buf, btnW*2, lastY, btnW, btnH, ".", calcDigit)
			// Equals
			drawCalcButton(buf, btnW*3, lastY, btnW, btnH, "=", calcEquals)
		}

		// Draw history toggle indicator
		if !state.showHist {
			toggleStyle := buffer.Style{Fg: calcDim}
			buf.DrawText(calcW-12, h-2, "[h] hist→", toggleStyle)
		} else {
			toggleStyle := buffer.Style{Fg: calcDim}
			buf.DrawText(calcW-6, h-2, "←hist", toggleStyle)
		}
	})

	base.Run()
}

func drawCalcButton(buf *buffer.Buffer, x, y, w, h int, label string, color buffer.Color) {
	style := buffer.Style{Fg: color}

	// Border
	borderStyle := buffer.Style{Fg: calcBorder}
	buf.SetCell(x, y, buffer.NewCell('┌', borderStyle))
	buf.SetCell(x+w-1, y, buffer.NewCell('┐', borderStyle))
	buf.SetCell(x, y+h-1, buffer.NewCell('└', borderStyle))
	buf.SetCell(x+w-1, y+h-1, buffer.NewCell('┘', borderStyle))

	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.NewCell('─', borderStyle))
		buf.SetCell(x+i, y+h-1, buffer.NewCell('─', borderStyle))
	}
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.NewCell('│', borderStyle))
		buf.SetCell(x+w-1, y+i, buffer.NewCell('│', borderStyle))
	}

	// Label centered
	labelX := x + w/2 - len(label)/2
	labelY := y + h/2
	if labelY >= y+h-1 {
		labelY = y + h - 2
	}
	if labelY < y+1 {
		labelY = y + 1
	}
	for i, r := range label {
		if labelX+i < x+w-1 {
			buf.SetCell(labelX+i, labelY, buffer.NewCell(r, style))
		}
	}
}

func drawHistoryPanel(buf *buffer.Buffer, x, y, w, h int, history []string) {
	// Border
	borderStyle := buffer.Style{Fg: calcBorder}
	buf.SetCell(x, y, buffer.NewCell('┌', borderStyle))
	buf.SetCell(x+w-1, y, buffer.NewCell('┐', borderStyle))
	buf.SetCell(x, y+h-1, buffer.NewCell('└', borderStyle))
	buf.SetCell(x+w-1, y+h-1, buffer.NewCell('┘', borderStyle))
	for i := 1; i < w-1; i++ {
		buf.SetCell(x+i, y, buffer.NewCell('─', borderStyle))
		buf.SetCell(x+i, y+h-1, buffer.NewCell('─', borderStyle))
	}
	for i := 1; i < h-1; i++ {
		buf.SetCell(x, y+i, buffer.NewCell('│', borderStyle))
		buf.SetCell(x+w-1, y+i, buffer.NewCell('│', borderStyle))
	}

	// Title
	titleStyle := buffer.Style{Fg: calcAccent, Flags: buffer.Bold}
	buf.DrawText(x+2, y, " History ", titleStyle)

	// Entries (most recent at bottom)
	entryStyle := buffer.Style{Fg: calcHistText}
	maxEntries := h - 3
	startIdx := 0
	if len(history) > maxEntries {
		startIdx = len(history) - maxEntries
	}
	for i, entry := range history[startIdx:] {
		entryY := y + 1 + i
		if entryY >= y+h-1 {
			break
		}
		text := entry
		if len(text) > w-4 {
			text = text[:w-4]
		}
		buf.DrawText(x+1, entryY, text, entryStyle)
	}

	if len(history) == 0 {
		emptyStyle := buffer.Style{Fg: calcDim}
		buf.DrawText(x+2, y+h/2, "(no history)", emptyStyle)
	}
}
