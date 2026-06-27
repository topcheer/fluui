package block

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSaveContainer_Empty(t *testing.T) {
	c := NewBlockContainer()
	r := NewDefaultRegistry()

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	var sc SerializedContainer
	if err := json.Unmarshal(data, &sc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if sc.Version != 1 {
		t.Errorf("Version = %d, want 1", sc.Version)
	}
	if len(sc.Blocks) != 0 {
		t.Errorf("Blocks len = %d, want 0", len(sc.Blocks))
	}
}

func TestLoadContainer_Empty(t *testing.T) {
	r := NewDefaultRegistry()
	data := []byte(`{"version":1,"blocks":[]}`)

	c, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}
	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}
}

func TestSerializeRoundTrip_AllTypes(t *testing.T) {
	r := NewDefaultRegistry()
	c := NewBlockContainer()

	// 1. ThinkingBlock
	tb := NewThinkingBlock("think-1")
	tb.AppendDelta("Let me analyze")
	tb.AppendDelta(" this problem")
	tb.Toggle() // expand

	// 2. AssistantTextBlock
	at := NewAssistantTextBlock("asst-1")
	at.AppendDelta("Here is **bold** text")
	at.Complete()

	// 3. ToolCallBlock
	tc := NewToolCallBlock("tool-1", "read_file", `{"path":"/etc/hosts"}`)
	tc.Complete()

	// 4. ToolResultBlock
	tr := NewToolResultBlock("result-1")
	tr.AppendDelta("127.0.0.1 localhost")
	tr.Complete()

	// 5. ErrorBlock
	eb := NewErrorBlockWithMessage("err-1", "file not found")

	// 6. UserMessageBlock
	um := NewUserMessageBlock("user-1", "What is the hostname?")

	c.AddBlock(tb)
	c.AddBlock(at)
	c.AddBlock(tc)
	c.AddBlock(tr)
	c.AddBlock(eb)
	c.AddBlock(um)

	// Save
	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	// Load
	loaded, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}

	if loaded.Len() != 6 {
		t.Fatalf("loaded Len = %d, want 6", loaded.Len())
	}

	blocks := loaded.Blocks()

	// 1. ThinkingBlock
	gotTb, ok := blocks[0].(*ThinkingBlock)
	if !ok {
		t.Fatalf("block[0] type = %T, want *ThinkingBlock", blocks[0])
	}
	if gotTb.ID() != "think-1" {
		t.Errorf("thinking ID = %q, want %q", gotTb.ID(), "think-1")
	}
	if gotTb.Content() != "Let me analyze this problem" {
		t.Errorf("thinking content = %q", gotTb.Content())
	}
	if gotTb.Collapsed() {
		t.Error("thinking should be expanded (toggled from default collapsed)")
	}

	// 2. AssistantTextBlock
	gotAt, ok := blocks[1].(*AssistantTextBlock)
	if !ok {
		t.Fatalf("block[1] type = %T, want *AssistantTextBlock", blocks[1])
	}
	if gotAt.ID() != "asst-1" {
		t.Errorf("asst ID = %q", gotAt.ID())
	}
	if gotAt.Content() != "Here is **bold** text" {
		t.Errorf("asst content = %q", gotAt.Content())
	}

	// 3. ToolCallBlock
	gotTc, ok := blocks[2].(*ToolCallBlock)
	if !ok {
		t.Fatalf("block[2] type = %T, want *ToolCallBlock", blocks[2])
	}
	if gotTc.ID() != "tool-1" {
		t.Errorf("tool ID = %q", gotTc.ID())
	}
	if gotTc.ToolName() != "read_file" {
		t.Errorf("tool name = %q", gotTc.ToolName())
	}
	if gotTc.RawArgs() != `{"path":"/etc/hosts"}` {
		t.Errorf("tool args = %q", gotTc.RawArgs())
	}

	// 4. ToolResultBlock
	gotTr, ok := blocks[3].(*ToolResultBlock)
	if !ok {
		t.Fatalf("block[3] type = %T, want *ToolResultBlock", blocks[3])
	}
	if gotTr.ID() != "result-1" {
		t.Errorf("result ID = %q", gotTr.ID())
	}
	if gotTr.Output() != "127.0.0.1 localhost" {
		t.Errorf("result output = %q", gotTr.Output())
	}

	// 5. ErrorBlock
	gotEb, ok := blocks[4].(*ErrorBlock)
	if !ok {
		t.Fatalf("block[4] type = %T, want *ErrorBlock", blocks[4])
	}
	if gotEb.ID() != "err-1" {
		t.Errorf("error ID = %q", gotEb.ID())
	}
	if gotEb.Message() != "file not found" {
		t.Errorf("error message = %q", gotEb.Message())
	}

	// 6. UserMessageBlock
	gotUm, ok := blocks[5].(*UserMessageBlock)
	if !ok {
		t.Fatalf("block[5] type = %T, want *UserMessageBlock", blocks[5])
	}
	if gotUm.ID() != "user-1" {
		t.Errorf("user ID = %q", gotUm.ID())
	}
	if gotUm.Content() != "What is the hostname?" {
		t.Errorf("user content = %q", gotUm.Content())
	}
}

func TestSerialize_ThinkingBlock(t *testing.T) {
	b := NewThinkingBlock("t1")
	b.AppendDelta("Analyzing data")
	b.Complete() // finalize

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var s struct {
		Collapsed bool   `json:"collapsed"`
		Content   string `json:"content"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.Content != "Analyzing data" {
		t.Errorf("content = %q", s.Content)
	}
}

func TestSerialize_AssistantTextBlock(t *testing.T) {
	b := NewAssistantTextBlock("a1")
	b.AppendDelta("Hello ")
	b.AppendDelta("world")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var s struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.Content != "Hello world" {
		t.Errorf("content = %q, want %q", s.Content, "Hello world")
	}
}

func TestSerialize_ToolCallBlock(t *testing.T) {
	b := NewToolCallBlock("c1", "grep", `{"pattern":"TODO"}`)

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var s struct {
		ToolName string `json:"tool_name"`
		Args     string `json:"args"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.ToolName != "grep" {
		t.Errorf("tool_name = %q", s.ToolName)
	}
	if s.Args != `{"pattern":"TODO"}` {
		t.Errorf("args = %q", s.Args)
	}
}

func TestSerialize_ToolResultBlock(t *testing.T) {
	b := NewToolResultBlock("r1")
	b.AppendDelta("line1\n")
	b.AppendDelta("line2")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var s struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.Result != "line1\nline2" {
		t.Errorf("result = %q", s.Result)
	}
}

func TestSerialize_ErrorBlock(t *testing.T) {
	b := NewErrorBlockWithMessage("e1", "permission denied")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var s struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.Message != "permission denied" {
		t.Errorf("message = %q", s.Message)
	}
}

func TestSerialize_UserMessageBlock(t *testing.T) {
	b := NewUserMessageBlock("u1", "Hello assistant")

	data, err := b.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var s struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if s.Content != "Hello assistant" {
		t.Errorf("content = %q", s.Content)
	}
}

func TestLoadContainer_UnknownType(t *testing.T) {
	r := NewDefaultRegistry()
	data := []byte(`{"version":1,"blocks":[{"type":"nonexistent","id":"x","state":"complete"}]}`)

	_, err := LoadContainer(data, r)
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("error should mention 'not registered', got: %v", err)
	}
}

func TestLoadContainer_InvalidJSON(t *testing.T) {
	r := NewDefaultRegistry()
	data := []byte(`{invalid json}`)

	_, err := LoadContainer(data, r)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestSaveLoad_SingleBlock(t *testing.T) {
	r := NewDefaultRegistry()
	c := NewBlockContainer()

	tb := NewThinkingBlock("think-x")
	tb.AppendDelta("Reasoning here")
	c.AddBlock(tb)

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	// The JSON should contain the type and content
	str := string(data)
	if !strings.Contains(str, "thinking") {
		t.Errorf("JSON should contain 'thinking': %s", str)
	}
	if !strings.Contains(str, "Reasoning here") {
		t.Errorf("JSON should contain content: %s", str)
	}
	if !strings.Contains(str, "think-x") {
		t.Errorf("JSON should contain ID: %s", str)
	}

	loaded, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}
	if loaded.Len() != 1 {
		t.Fatalf("loaded Len = %d, want 1", loaded.Len())
	}

	got := loaded.Blocks()[0]
	if got.ID() != "think-x" {
		t.Errorf("ID = %q", got.ID())
	}
	if got.Type() != TypeThinking {
		t.Errorf("Type = %v", got.Type())
	}
}

func TestSaveContainer_NilRegistry(t *testing.T) {
	c := NewBlockContainer()
	tb := NewThinkingBlock("t1")
	c.AddBlock(tb)

	// Save with nil registry should still work (SaveContainer doesn't use the registry)
	_, err := SaveContainer(c, nil)
	if err != nil {
		t.Fatalf("SaveContainer with nil registry should work: %v", err)
	}
}

func TestSerialize_BlockStatePreserved(t *testing.T) {
	r := NewDefaultRegistry()
	c := NewBlockContainer()

	tb := NewThinkingBlock("think-1")
	tb.AppendDelta("content")
	tb.Complete()
	c.AddBlock(tb)

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	var sc SerializedContainer
	if err := json.Unmarshal(data, &sc); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if len(sc.Blocks) != 1 {
		t.Fatalf("blocks len = %d", len(sc.Blocks))
	}
	if sc.Blocks[0].State != "complete" {
		t.Errorf("state = %q, want %q", sc.Blocks[0].State, "complete")
	}

	// Load and verify state
	loaded, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}
	got := loaded.Blocks()[0]
	if got.State() != BlockComplete {
		t.Errorf("loaded state = %v, want %v", got.State(), BlockComplete)
	}
}

func TestSerialize_MultipleBlocksOrderPreserved(t *testing.T) {
	r := NewDefaultRegistry()
	c := NewBlockContainer()

	c.AddBlock(NewUserMessageBlock("u1", "first"))
	c.AddBlock(NewAssistantTextBlock("a1"))
	c.AddBlock(NewUserMessageBlock("u2", "second"))
	c.AddBlock(NewAssistantTextBlock("a2"))

	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	loaded, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}

	blocks := loaded.Blocks()
	if len(blocks) != 4 {
		t.Fatalf("blocks len = %d, want 4", len(blocks))
	}

	expectedIDs := []string{"u1", "a1", "u2", "a2"}
	for i, want := range expectedIDs {
		if blocks[i].ID() != want {
			t.Errorf("blocks[%d].ID() = %q, want %q", i, blocks[i].ID(), want)
		}
	}
}

func TestSerialize_DeserializeRestoresContent(t *testing.T) {
	r := NewDefaultRegistry()
	c := NewBlockContainer()

	// Add a tool call with specific args
	tc := NewToolCallBlock("tc1", "calculate", `{"x":42,"y":99}`)
	tc.Complete()
	c.AddBlock(tc)

	// Save + Load
	data, _ := SaveContainer(c, r)
	loaded, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}

	got := loaded.Blocks()[0]
	gotTc, ok := got.(*ToolCallBlock)
	if !ok {
		t.Fatalf("type = %T, want *ToolCallBlock", got)
	}
	if gotTc.ToolName() != "calculate" {
		t.Errorf("toolName = %q, want %q", gotTc.ToolName(), "calculate")
	}
	if gotTc.RawArgs() != `{"x":42,"y":99}` {
		t.Errorf("rawArgs = %q", gotTc.RawArgs())
	}
}
