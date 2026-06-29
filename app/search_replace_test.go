package app

import (
	"testing"

	"github.com/topcheer/fluui/block"
)

func TestReplaceMode_Getters(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("hello")
	rm.SetReplace("world")

	if rm.Find() != "hello" {
		t.Errorf("Find() = %q, want 'hello'", rm.Find())
	}
	if rm.Replace() != "world" {
		t.Errorf("Replace() = %q, want 'world'", rm.Replace())
	}
}

func TestReplaceMode_CaseSensitive(t *testing.T) {
	rm := NewReplaceMode()
	if rm.CaseSensitive() {
		t.Error("CaseSensitive should be false by default")
	}
	rm.SetCaseSensitive(true)
	if !rm.CaseSensitive() {
		t.Error("CaseSensitive should be true after setting")
	}
}

func TestReplaceMode_ReplaceAll(t *testing.T) {
	rm := NewReplaceMode()
	if !rm.ReplaceAll() {
		t.Error("ReplaceAll should be true by default")
	}
	rm.SetReplaceAll(false)
	if rm.ReplaceAll() {
		t.Error("ReplaceAll should be false after setting")
	}
}

func TestReplaceMode_ReplaceFirst_CaseInsensitive(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("foo")
	rm.SetReplace("bar")
	rm.SetReplaceAll(false)

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "foo Foo FOO foo"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Replacements != 1 {
		t.Errorf("Replacements = %d, want 1 (first only)", results[0].Replacements)
	}
	if results[0].NewText != "bar Foo FOO foo" {
		t.Errorf("NewText = %q, want 'bar Foo FOO foo'", results[0].NewText)
	}
}

func TestReplaceMode_ReplaceAll_CaseInsensitive(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("foo")
	rm.SetReplace("bar")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "foo Foo FOO foo"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Replacements != 4 {
		t.Errorf("Replacements = %d, want 4 (all occurrences)", results[0].Replacements)
	}
	if results[0].NewText != "bar bar bar bar" {
		t.Errorf("NewText = %q, want 'bar bar bar bar'", results[0].NewText)
	}
}

func TestReplaceMode_CaseSensitive_All(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("Foo")
	rm.SetReplace("Bar")
	rm.SetCaseSensitive(true)

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "foo Foo FOO"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Replacements != 1 {
		t.Errorf("Replacements = %d, want 1 (only exact-case 'Foo')", results[0].Replacements)
	}
	if results[0].NewText != "foo Bar FOO" {
		t.Errorf("NewText = %q, want 'foo Bar FOO'", results[0].NewText)
	}
}

func TestReplaceMode_NoMatch(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("xyz")
	rm.SetReplace("abc")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "hello world"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0 (no matches)", len(results))
	}
}

func TestReplaceMode_EmptyFind(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("")
	rm.SetReplace("test")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "hello"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0 (empty find)", len(results))
	}
}

func TestReplaceMode_MultipleBlocks(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("cat")
	rm.SetReplace("dog")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "I have a cat"))
	c.AddBlock(block.NewUserMessageBlock("u2", "no match here"))
	c.AddBlock(block.NewUserMessageBlock("u3", "another cat here"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2 (u1 and u3)", len(results))
	}
	if results[0].BlockID != "u1" {
		t.Errorf("results[0].BlockID = %s, want u1", results[0].BlockID)
	}
	if results[1].BlockID != "u3" {
		t.Errorf("results[1].BlockID = %s, want u3", results[1].BlockID)
	}
}

func TestReplaceMode_AssistantTextBlock(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("hello")
	rm.SetReplace("goodbye")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewAssistantTextBlock("a1"))
	atb := c.Blocks()[0].(*block.AssistantTextBlock)
	atb.AppendDelta("hello world hello")

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Replacements != 2 {
		t.Errorf("Replacements = %d, want 2", results[0].Replacements)
	}
	if results[0].NewText != "goodbye world goodbye" {
		t.Errorf("NewText = %q, want 'goodbye world goodbye'", results[0].NewText)
	}
	if atb.Content() != "goodbye world goodbye" {
		t.Errorf("Content() = %q, want 'goodbye world goodbye'", atb.Content())
	}
}

func TestReplaceMode_ThinkingBlock(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("think")
	rm.SetReplace("ponder")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewThinkingBlock("t1"))
	tb := c.Blocks()[0].(*block.ThinkingBlock)
	tb.AppendDelta("I think about thinking")

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Replacements != 2 {
		t.Errorf("Replacements = %d, want 2", results[0].Replacements)
	}
	if tb.Content() != "I ponder about pondering" {
		t.Errorf("Content() = %q, want 'I ponder about pondering'", tb.Content())
	}
}

func TestReplaceMode_SkipsNonContentBlocks(t *testing.T) {
	rm := NewReplaceMode()
	rm.SetFind("test")
	rm.SetReplace("exam")

	c := block.NewBlockContainer()
	c.AddBlock(block.NewToolCallBlock("tc1", "bash", "echo test"))

	results := rm.ReplaceInBlocks(c.Blocks())
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0 (tool calls skipped)", len(results))
	}
}

func TestTotalReplacements(t *testing.T) {
	results := []ReplaceResult{
		{BlockID: "u1", Replacements: 3},
		{BlockID: "u2", Replacements: 2},
		{BlockID: "u3", Replacements: 0},
	}
	total := TotalReplacements(results)
	if total != 5 {
		t.Errorf("TotalReplacements = %d, want 5", total)
	}
}

func TestReplaceInString_CaseInsensitiveAll(t *testing.T) {
	text, count := replaceInString("Go go GO", "go", "stop", false, true)
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if text != "stop stop stop" {
		t.Errorf("text = %q, want 'stop stop stop'", text)
	}
}

func TestReplaceInString_CaseSensitiveFirst(t *testing.T) {
	text, count := replaceInString("Foo foo FOO", "foo", "bar", true, false)
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	if text != "Foo bar FOO" {
		t.Errorf("text = %q, want 'Foo bar FOO'", text)
	}
}

func TestReplaceInString_EmptyFind(t *testing.T) {
	text, count := replaceInString("hello", "", "x", false, true)
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
	if text != "hello" {
		t.Errorf("text = %q, want 'hello'", text)
	}
}

func TestCanReplaceBlock(t *testing.T) {
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "hello"))
	c.AddBlock(block.NewAssistantTextBlock("a1"))
	c.AddBlock(block.NewThinkingBlock("t1"))

	blocks := c.Blocks()
	for i := 0; i < 3; i++ {
		if !CanReplaceBlock(blocks[i]) {
			t.Errorf("blocks[%d] should support replacement", i)
		}
	}
}

func TestCanReplaceBlock_ToolCall(t *testing.T) {
	b := block.NewToolCallBlock("tc1", "bash", "echo test")
	if CanReplaceBlock(b) {
		t.Error("ToolCallBlock should NOT support replacement")
	}
}

func TestReplaceAll_Function(t *testing.T) {
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "hello hello world"))
	c.AddBlock(block.NewUserMessageBlock("u2", "hello there"))

	total := ReplaceAll(c.Blocks(), "hello", "hi")
	if total != 3 {
		t.Errorf("ReplaceAll total = %d, want 3", total)
	}

	u1 := c.Blocks()[0].(*block.UserMessageBlock)
	if u1.Content() != "hi hi world" {
		t.Errorf("u1.Content() = %q, want 'hi hi world'", u1.Content())
	}
}

func TestReplaceAll_EmptyQuery(t *testing.T) {
	c := block.NewBlockContainer()
	c.AddBlock(block.NewUserMessageBlock("u1", "hello"))
	total := ReplaceAll(c.Blocks(), "", "x")
	if total != 0 {
		t.Errorf("ReplaceAll with empty query = %d, want 0", total)
	}
}

func TestReplaceInBlock_Function(t *testing.T) {
	newText, offset, ok := ReplaceInBlock("hello world hello", "hello", "hi", 0)
	if !ok {
		t.Fatal("ReplaceInBlock should find a match")
	}
	if newText != "hi world hello" {
		t.Errorf("newText = %q, want 'hi world hello'", newText)
	}
	if offset != 2 {
		t.Errorf("offset = %d, want 2", offset)
	}
}

func TestReplaceInBlock_FromOffset(t *testing.T) {
	newText, offset, ok := ReplaceInBlock("hello world hello", "hello", "hi", 6)
	if !ok {
		t.Fatal("ReplaceInBlock should find a match at offset 6")
	}
	if newText != "hello world hi" {
		t.Errorf("newText = %q, want 'hello world hi'", newText)
	}
	if offset != 14 {
		t.Errorf("offset = %d, want 14", offset)
	}
}

func TestReplaceInBlock_NoMatch(t *testing.T) {
	_, _, ok := ReplaceInBlock("hello world", "xyz", "abc", 0)
	if ok {
		t.Error("ReplaceInBlock should return false for no match")
	}
}

func TestReplaceInBlock_EmptyQuery(t *testing.T) {
	_, _, ok := ReplaceInBlock("hello world", "", "x", 0)
	if ok {
		t.Error("ReplaceInBlock should return false for empty query")
	}
}
