package block_test

import (
	"testing"

	"github.com/topcheer/fluui/block"
	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// === Integration Test 1: Full conversation flow ===

func TestIntegrationFullConversation(t *testing.T) {
	container := block.NewBlockContainer()

	// Simulate a full conversation flow
	userMsg := block.NewUserMessageBlock("u1", "Hello")
	assistant1 := block.NewAssistantTextBlock("a1")
	assistant1.AppendDelta("Hi there!")
	assistant1.Complete()

	thinking := block.NewThinkingBlock("t1")
	thinking.AppendDelta("Let me check the files.")
	thinking.Complete()

	toolCall := block.NewToolCallBlock("tc1", "read_file", `{"path":"main.go"}`)
	toolCall.Complete()

	toolResult := block.NewToolResultBlock("tr1")
	toolResult.AppendDelta("package main\nfunc main() {}")
	toolResult.Complete()

	assistant2 := block.NewAssistantTextBlock("a2")
	assistant2.AppendDelta("Done!")
	assistant2.Complete()

	container.AddBlock(userMsg)
	container.AddBlock(assistant1)
	container.AddBlock(thinking)
	container.AddBlock(toolCall)
	container.AddBlock(toolResult)
	container.AddBlock(assistant2)

	// Verify all blocks are in the container
	if container.Len() != 6 {
		t.Fatalf("expected 6 blocks, got %d", container.Len())
	}

	// Verify block types in order
	expectedTypes := []block.BlockType{
		block.TypeUserMessage,
		block.TypeAssistantText,
		block.TypeThinking,
		block.TypeToolCall,
		block.TypeToolResult,
		block.TypeAssistantText,
	}
	for i, expected := range expectedTypes {
		b := container.Blocks()[i]
		if b.Type() != expected {
			t.Errorf("block %d: expected type %s, got %s", i, expected, b.Type())
		}
	}

	// Measure the container
	size := container.Measure(component.Constraints{MaxWidth: 80})
	if size.W <= 0 || size.H <= 0 {
		t.Errorf("container measure: expected positive size, got %v", size)
	}

	// SetBounds and Paint into a buffer
	bounds := component.Rect{X: 0, Y: 0, W: 80, H: size.H}
	container.SetBounds(bounds)

	// Verify each block got its bounds
	for i, b := range container.Blocks() {
		bb := b.(component.Component)
		r := bb.Bounds()
		if r.W != 80 {
			t.Errorf("block %d: expected width 80, got %d", i, r.W)
		}
		if r.H <= 0 {
			t.Errorf("block %d: expected positive height", i)
		}
	}

	// Paint into a buffer — should not panic
	buf := buffer.NewBuffer(80, size.H)
	container.Paint(buf)

	// Verify the buffer has content (not all blank)
	hasContent := false
	for y := 0; y < size.H && !hasContent; y++ {
		for x := 0; x < 80; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != ' ' && cell.Rune != 0 {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Error("buffer should have rendered content after Paint")
	}

	// Verify specific content: "Hi there!" should be on some line
	foundHi := false
	foundDone := false
	for y := 0; y < size.H; y++ {
		lineText := ""
		for x := 0; x < 80; x++ {
			cell := buf.GetCell(x, y)
			lineText += string(cell.Rune)
		}
		if contains(lineText, "Hi") {
			foundHi = true
		}
		if contains(lineText, "Done") {
			foundDone = true
		}
	}
	if !foundHi {
		t.Error("expected to find 'Hi' in rendered output")
	}
	if !foundDone {
		t.Error("expected to find 'Done' in rendered output")
	}

	// Verify all blocks are complete
	for i, b := range container.Blocks() {
		if b.State() != block.BlockComplete {
			t.Errorf("block %d: expected complete state, got %s", i, b.State())
		}
	}
}

// === Integration Test 2: Container vertical layout ===

func TestIntegrationContainerLayout(t *testing.T) {
	container := block.NewBlockContainer()
	container.SetSpacing(1)

	// Add 3 blocks with known heights
	b1 := block.NewUserMessageBlock("u1", "Line one")
	b2 := block.NewUserMessageBlock("u2", "Line two")
	b3 := block.NewUserMessageBlock("u3", "Line three")

	container.AddBlock(b1)
	container.AddBlock(b2)
	container.AddBlock(b3)

	// Each "Line X" wraps to 1 line at width 80
	cs := component.Constraints{MaxWidth: 80}
	sz1 := b1.Measure(cs)
	sz2 := b2.Measure(cs)
	sz3 := b3.Measure(cs)

	expectedH := sz1.H + sz2.H + sz3.H + 2*1 // 2 spacings of 1
	containerSize := container.Measure(cs)
	if containerSize.H != expectedH {
		t.Errorf("container height: expected %d, got %d", expectedH, containerSize.H)
	}

	// SetBounds — verify Y offsets
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: expectedH})

	r1 := b1.Bounds()
	r2 := b2.Bounds()
	r3 := b3.Bounds()

	// Block 1 starts at Y=0
	if r1.Y != 0 {
		t.Errorf("block 1: expected Y=0, got %d", r1.Y)
	}

	// Block 2 starts after block 1 + spacing
	expectedY2 := r1.H + 1 // spacing
	if r2.Y != expectedY2 {
		t.Errorf("block 2: expected Y=%d, got %d", expectedY2, r2.Y)
	}

	// Block 3 starts after block 2 + spacing
	expectedY3 := r2.Y + r2.H + 1
	if r3.Y != expectedY3 {
		t.Errorf("block 3: expected Y=%d, got %d", expectedY3, r3.Y)
	}

	// Verify total height matches
	totalBottom := r3.Y + r3.H
	if totalBottom != expectedH {
		t.Errorf("total bottom: expected %d, got %d", expectedH, totalBottom)
	}

	// Paint and verify content at correct Y positions
	buf := buffer.NewBuffer(80, expectedH)
	container.Paint(buf)

	// Block 1 content at Y=0
	cell := buf.GetCell(0, 0)
	if cell.Rune != 'L' {
		t.Errorf("Y=0: expected 'L' (Line one), got %q", string(cell.Rune))
	}

	// Block 2 content at expectedY2
	cell = buf.GetCell(0, expectedY2)
	if cell.Rune != 'L' {
		t.Errorf("Y=%d: expected 'L' (Line two), got %q", expectedY2, string(cell.Rune))
	}
}

// === Integration Test 3: ThinkingBlock toggle changes container height ===

func TestIntegrationThinkingToggle(t *testing.T) {
	container := block.NewBlockContainer()
	container.SetSpacing(0) // no spacing for simpler math

	thinking := block.NewThinkingBlock("t1")
	// Append multi-line content so expanded is larger
	thinking.AppendDelta("This is a longer thinking content that should span multiple lines when rendered at a width of about 40 characters or so.")
	container.AddBlock(thinking)

	cs := component.Constraints{MaxWidth: 40}

	// Collapsed: 1 line
	collapsedSize := container.Measure(cs)
	if collapsedSize.H != 1 {
		t.Errorf("collapsed: expected height 1, got %d", collapsedSize.H)
	}

	// Expand
	thinking.Toggle()
	if thinking.Collapsed() {
		t.Error("expected thinking.Collapsed() == false (expanded) after Toggle")
	}

	expandedSize := container.Measure(cs)
	// Expanded: 1 header + content lines (should be > 1)
	if expandedSize.H <= 1 {
		t.Errorf("expanded: expected height > 1, got %d", expandedSize.H)
	}
	if expandedSize.H <= collapsedSize.H {
		t.Errorf("expanded (%d) should be taller than collapsed (%d)", expandedSize.H, collapsedSize.H)
	}

	// SetBounds and paint both states
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: expandedSize.H})
	buf := buffer.NewBuffer(40, expandedSize.H)
	container.Paint(buf)

	// First line should have header marker
	headerCell := buf.GetCell(0, 0)
	if headerCell.Rune == 0 || headerCell.Rune == ' ' {
		t.Error("expected header content at Y=0 after expand")
	}

	// Collapse again
	thinking.Toggle()
	if !thinking.Collapsed() {
		t.Error("expected thinking.Collapsed() == true after second Toggle")
	}
	reCollapsedSize := container.Measure(cs)
	if reCollapsedSize.H != 1 {
		t.Errorf("re-collapsed: expected height 1, got %d", reCollapsedSize.H)
	}

	// Verify dirty flag was set by Toggle
	if !thinking.IsDirty() {
		t.Error("expected dirty flag after Toggle")
	}
}

// === Integration Test 4: Streaming simulation ===

func TestIntegrationStreamingSimulation(t *testing.T) {
	container := block.NewBlockContainer()

	assistant := block.NewAssistantTextBlock("a1")
	container.AddBlock(assistant)

	cs := component.Constraints{MaxWidth: 80}

	// Initial state: empty content
	initialSize := assistant.Measure(cs)
	if initialSize.H != 1 {
		t.Errorf("empty: expected height 1, got %d", initialSize.H)
	}
	if !assistant.IsDirty() {
		t.Error("expected dirty flag on new block")
	}

	// Stream first delta
	assistant.AppendDelta("Hello")
	if !assistant.IsDirty() {
		t.Error("expected dirty after AppendDelta")
	}
	assistant.ClearDirty()
	if assistant.IsDirty() {
		t.Error("expected clean after ClearDirty")
	}

	size1 := assistant.Measure(cs)
	if size1.H != 1 {
		t.Errorf("after 'Hello': expected height 1, got %d", size1.H)
	}

	// Stream more content — should stay 1 line at width 80
	assistant.AppendDelta(" world, this is a test of streaming text.")
	size2 := assistant.Measure(cs)
	if size2.H != 1 {
		t.Errorf("after long delta: expected height 1 at width 80, got %d", size2.H)
	}

	// Stream enough to wrap to multiple lines at width 20
	assistant.ClearDirty()
	assistant2 := block.NewAssistantTextBlock("a2")
	container.AddBlock(assistant2)

	longText := "This is a very long piece of text that will definitely wrap across multiple lines when rendered at a narrow width of only twenty characters."
	assistant2.AppendDelta(longText)

	narrowCS := component.Constraints{MaxWidth: 20}
	size3 := assistant2.Measure(narrowCS)
	if size3.H < 3 {
		t.Errorf("long text at width 20: expected height >= 3, got %d", size3.H)
	}

	// Verify dirty was set by streaming
	if !assistant2.IsDirty() {
		t.Error("expected dirty after streaming delta")
	}

	// Container dirty propagates from child
	if !container.IsDirty() {
		t.Error("expected container dirty when child is dirty")
	}

	// Clear all dirty
	container.ClearDirty()
	if container.IsDirty() {
		t.Error("expected container clean after ClearDirty")
	}
	if assistant.IsDirty() {
		t.Error("expected assistant clean after container.ClearDirty")
	}

	// Complete the block
	assistant2.Complete()
	if assistant2.State() != block.BlockComplete {
		t.Errorf("expected complete state, got %s", assistant2.State())
	}
	if !assistant2.IsDirty() {
		t.Error("expected dirty after Complete")
	}

	// Simulate incremental re-measure and paint (render cycle)
	containerSize := container.Measure(narrowCS)
	container.SetBounds(component.Rect{X: 0, Y: 0, W: 20, H: containerSize.H})
	buf := buffer.NewBuffer(20, containerSize.H)
	container.Paint(buf)

	// Verify the buffer has content
	hasContent := false
	for y := 0; y < containerSize.H && !hasContent; y++ {
		for x := 0; x < 20; x++ {
			cell := buf.GetCell(x, y)
			if cell.Rune != ' ' && cell.Rune != 0 {
				hasContent = true
				break
			}
		}
	}
	if !hasContent {
		t.Error("buffer should have content after paint")
	}

	// Verify content includes streamed text
	allText := ""
	for y := 0; y < containerSize.H; y++ {
		for x := 0; x < 20; x++ {
			allText += string(buf.GetCell(x, y).Rune)
		}
	}
	if !contains(allText, "streaming") {
		t.Error("expected to find 'streaming' in rendered output")
	}
}

// === Helper: contains ===

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
