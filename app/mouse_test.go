package app

import (
	"testing"

	"github.com/topcheer/fluui/component"
)

func TestMouseHandlerCreation(t *testing.T) {
	app := NewChatApp(80, 24)
	handler := NewMouseHandler(app)

	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if handler.RegionTree() == nil {
		t.Fatal("expected non-nil region tree")
	}
	if handler.RegionTree().Len() != 0 {
		t.Fatalf("expected 0 regions, got %d", handler.RegionTree().Len())
	}
}

func TestMouseHandlerRebuild(t *testing.T) {
	app := NewChatApp(80, 24)
	app.AddThinking()
	app.AddToolResult()

	// Layout the container so blocks get bounds.
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	if handler.RegionTree().Len() != 2 {
		t.Fatalf("expected 2 regions (thinking + tool_result), got %d", handler.RegionTree().Len())
	}
}

func TestMouseHandlerClickThinking(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := app.AddThinking()

	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	// ThinkingBlock defaults to collapsed=true.
	if !tb.Collapsed() {
		t.Fatal("expected ThinkingBlock to start collapsed")
	}

	// Click on the header row (Y=0, X=5).
	consumed := handler.HandleClick(5, 0)
	if !consumed {
		t.Fatal("expected click to be consumed")
	}
	if tb.Collapsed() {
		t.Fatal("expected ThinkingBlock to be expanded after click")
	}

	// Click again to collapse.
	handler.RebuildRegions()
	consumed = handler.HandleClick(5, 0)
	if !consumed {
		t.Fatal("expected second click to be consumed")
	}
	if !tb.Collapsed() {
		t.Fatal("expected ThinkingBlock to be collapsed again")
	}
}

func TestMouseHandlerClickToolResult(t *testing.T) {
	app := NewChatApp(80, 24)
	tr := app.AddToolResult()

	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	// ToolResultBlock defaults to collapsed=false.
	if tr.Collapsed() {
		t.Fatal("expected ToolResultBlock to start expanded")
	}

	// Click on the top border row (Y=0, X=3).
	consumed := handler.HandleClick(3, 0)
	if !consumed {
		t.Fatal("expected click to be consumed")
	}
	if !tr.Collapsed() {
		t.Fatal("expected ToolResultBlock to be collapsed after click")
	}
}

func TestMouseHandlerClickMiss(t *testing.T) {
	app := NewChatApp(80, 24)
	tb := app.AddThinking()

	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	// Click far below the block (Y=20, X=5).
	consumed := handler.HandleClick(5, 20)
	if consumed {
		t.Fatal("expected click to miss (return false)")
	}
	if !tb.Collapsed() {
		t.Fatal("ThinkingBlock should still be collapsed (no toggle)")
	}
}

func TestMouseHandlerMultipleBlocks(t *testing.T) {
	app := NewChatApp(80, 24)
	tb1 := app.AddThinking()
	tr := app.AddToolResult()
	tb2 := app.AddThinking()

	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})

	handler := NewMouseHandler(app)
	handler.RebuildRegions()

	if handler.RegionTree().Len() != 3 {
		t.Fatalf("expected 3 regions, got %d", handler.RegionTree().Len())
	}

	// All blocks should start collapsed (thinking) or expanded (tool_result).
	if !tb1.Collapsed() {
		t.Fatal("tb1 should be collapsed")
	}
	if tr.Collapsed() {
		t.Fatal("tr should be expanded")
	}
	if !tb2.Collapsed() {
		t.Fatal("tb2 should be collapsed")
	}

	// Click tb1 header (Y=0).
	handler.HandleClick(5, 0)
	if tb1.Collapsed() {
		t.Fatal("tb1 should be expanded after click")
	}

	// Click tr header. tr is at Y=1 (tb1 collapsed H=1 + spacing 1 → no wait,
	// let's just check it toggles somewhere).
	// After tb1 is expanded, layout changed, so rebuild.
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	handler.RebuildRegions()

	// Find tr's bounds to click correctly.
	trBounds := tr.Bounds()
	handler.HandleClick(3, trBounds.Y)
	if !tr.Collapsed() {
		t.Fatal("tr should be collapsed after click")
	}

	// Click tb2 header.
	app.container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 24})
	handler.RebuildRegions()
	tb2Bounds := tb2.Bounds()
	handler.HandleClick(5, tb2Bounds.Y)
	if tb2.Collapsed() {
		t.Fatal("tb2 should be expanded after click")
	}
}
