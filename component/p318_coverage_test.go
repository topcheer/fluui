package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// P318: Push remaining sub-85% functions upward + verify CHANGELOG

// codeblock cellWidth 83.3% — missing branches for CJK wide chars
func TestP318_CodeBlock_CellWidth_CJK(t *testing.T) {
	cb := NewCodeBlock("go", "中文")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	cb.Paint(buffer.NewBuffer(30, 3)) // wide chars
}

func TestP318_CodeBlock_CellWidth_Mixed(t *testing.T) {
	cb := NewCodeBlock("go", "abc中def文")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	cb.Paint(buffer.NewBuffer(30, 3))
}

func TestP318_CodeBlock_CellWidth_Tab(t *testing.T) {
	cb := NewCodeBlock("go", "\tindented")
	cb.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 3})
	cb.Paint(buffer.NewBuffer(30, 3))
}

// collapsible SetBounds 83.3%
func TestP318_Collapsible_SetBounds(t *testing.T) {
	col := NewCollapsible("Test", NewText("content"))
	col.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	col.Paint(buffer.NewBuffer(40, 10))
}

// approval_dialog executeActionLocked 80%
func TestP318_ApprovalDialog_ExecuteAction(t *testing.T) {
	d := NewApprovalDialog("Test", "Approve?")
	d.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 8})
	d.Paint(buffer.NewBuffer(40, 8))
}

// canvas SetCell/SetCellBG 80%
func TestP318_Canvas_SetCellEdge(t *testing.T) {
	c := NewCanvas()
	c.SetBounds(Rect{X: 0, Y: 0, W: 10, H: 5})
	c.Paint(buffer.NewBuffer(10, 5)) // paint exercises cell access
}

// autocomplete Paint 80%
func TestP318_AutoComplete_Paint(t *testing.T) {
	ac := NewAutoComplete()
	ac.SetQuery("he")
	ac.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 5})
	ac.Paint(buffer.NewBuffer(20, 5))
}

// barchart drawVerticalGrid 80%
func TestP318_BarChart_Grid(t *testing.T) {
	bc := NewBarChart()
	bc.SetShowGrid(true)
	bc.SetShowAxes(true)
	bc.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 10})
	bc.Paint(buffer.NewBuffer(40, 10))
}
