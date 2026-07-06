package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/term"
)

// ─── MouseHandler.Handle scrollbar interactions (44.8% → 70%+) ───

func TestP78_MouseHandle_ScrollbarDown(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddUserMessage("Line 1")
	asst := app.AddAssistantText()
	asst.AppendDelta("Long assistant response that fills some space for scrolling")

	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.scrollView.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	// Click on scrollbar column (rightmost column)
	scrollbarX := app.scrollView.ScrollbarColumn()
	if scrollbarX < 0 {
		return // no scrollbar when content fits
	}

	consumed := handler.Handle(&term.MouseEvent{
		Action: term.MouseDown,
		X:      scrollbarX,
		Y:      5,
	})
	_ = consumed
}

func TestP78_MouseHandle_ScrollbarDrag(t *testing.T) {
	app := NewChatApp(80, 24)
	// Add enough content to create a scrollbar
	for i := 0; i < 40; i++ {
		app.AddUserMessage("Line " + string(rune('A'+i%26)))
	}
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.scrollView.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	scrollbarX := app.scrollView.ScrollbarColumn()
	if scrollbarX < 0 {
		return
	}

	// Start drag
	handler.Handle(&term.MouseEvent{
		Action: term.MouseDown,
		X:      scrollbarX,
		Y:      3,
	})
	// Drag
	handler.Handle(&term.MouseEvent{
		Action: term.MouseDrag,
		X:      scrollbarX,
		Y:      10,
	})
}

func TestP78_MouseHandle_ScrollbarUp(t *testing.T) {
	app := NewChatApp(80, 24)
	for i := 0; i < 40; i++ {
		app.AddUserMessage("Line " + string(rune('A'+i%26)))
	}
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.scrollView.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	scrollbarX := app.scrollView.ScrollbarColumn()
	if scrollbarX < 0 {
		return
	}

	// Down, Drag, Up
	handler.Handle(&term.MouseEvent{Action: term.MouseDown, X: scrollbarX, Y: 3})
	handler.Handle(&term.MouseEvent{Action: term.MouseDrag, X: scrollbarX, Y: 10})
	handler.Handle(&term.MouseEvent{Action: term.MouseUp, X: scrollbarX, Y: 10})
}

func TestP78_MouseHandle_DragOutsideScrollbar(t *testing.T) {
	app := NewChatApp(80, 24)
	for i := 0; i < 40; i++ {
		app.AddUserMessage("Line " + string(rune('A'+i%26)))
	}
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.scrollView.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	scrollbarX := app.scrollView.ScrollbarColumn()
	if scrollbarX < 0 {
		return
	}

	// Start drag on scrollbar
	handler.Handle(&term.MouseEvent{Action: term.MouseDown, X: scrollbarX, Y: 3})
	// Drag OUTSIDE scrollbar column (should still be consumed by scrollbar)
	consumed := handler.Handle(&term.MouseEvent{
		Action: term.MouseDrag,
		X:      10, // outside scrollbar
		Y:      10,
	})
	if !consumed {
		t.Error("drag while scrollbar is active should be consumed")
	}
}

func TestP78_MouseHandle_MouseUpOutsideScrollbar(t *testing.T) {
	app := NewChatApp(80, 24)
	for i := 0; i < 40; i++ {
		app.AddUserMessage("Line")
	}
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.scrollView.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	scrollbarX := app.scrollView.ScrollbarColumn()
	if scrollbarX < 0 {
		return
	}

	// Start scrollbar drag
	handler.Handle(&term.MouseEvent{Action: term.MouseDown, X: scrollbarX, Y: 3})
	// Release outside scrollbar column
	consumed := handler.Handle(&term.MouseEvent{Action: term.MouseUp, X: 10, Y: 10})
	if !consumed {
		t.Error("mouse up while dragging should be consumed")
	}
}

func TestP78_MouseHandle_WheelUp(t *testing.T) {
	app := NewChatApp(80, 24)
	for i := 0; i < 10; i++ {
		app.AddUserMessage("Line")
	}
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	app.scrollView.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	consumed := handler.Handle(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: term.MouseWheelUp,
		X:      10, Y: 10,
	})
	if !consumed {
		t.Error("wheel up should be consumed")
	}
}

func TestP78_MouseHandle_ClickNoRegion(t *testing.T) {
	app := NewChatApp(80, 24)
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	// Click on empty area (no blocks registered)
	consumed := handler.Handle(&term.MouseEvent{
		Action: term.MouseDown,
		X:      40, Y: 12,
	})
	if consumed {
		t.Error("click on empty area should not be consumed")
	}
}

func TestP78_MouseHandle_UnknownWheelButton(t *testing.T) {
	app := NewChatApp(80, 24)
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	// Unknown wheel button should not be consumed
	consumed := handler.Handle(&term.MouseEvent{
		Action: term.MouseWheel,
		Button: 99, // unknown
		X:      10, Y: 10,
	})
	if consumed {
		t.Error("unknown wheel button should not be consumed")
	}
}
