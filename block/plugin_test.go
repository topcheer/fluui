package block

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

// --- Test fixtures ---

// dividerPlugin registers a custom DividerBlock.
type dividerPlugin struct{}

func (d *dividerPlugin) Name() string { return "divider-plugin" }
func (d *dividerPlugin) Init(r *Registry) error {
	r.Register("divider", func(id string) Block {
		return NewDividerBlock(id)
	})
	return nil
}

// DividerBlock is a custom block that renders a horizontal divider line.
type DividerBlock struct {
	BaseBlock
	char rune
}

// NewDividerBlock creates a custom divider block.
func NewDividerBlock(id string) *DividerBlock {
	return &DividerBlock{
		BaseBlock: NewBaseBlock(id, BlockType(99)),
		char:      '\u2550', // double horizontal line
	}
}

func (b *DividerBlock) Measure(cs component.Constraints) component.Size {
	w := cs.MaxWidth
	if w <= 0 {
		w = 40
	}
	return component.Size{W: w, H: 1}
}

func (b *DividerBlock) Paint(buf *buffer.Buffer) {
	r := b.Bounds()
	style := buffer.Style{Fg: buffer.RGB(0xBD, 0x93, 0xF9)}
	for x := r.X; x < r.X+r.W; x++ {
		buf.SetCell(x, r.Y, buffer.Cell{Rune: b.char, Width: 1, Fg: style.Fg})
	}
}

func (b *DividerBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{"char": b.char})
}

func (b *DividerBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Char rune `json:"char"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s.Char != 0 {
		b.char = s.Char
	}
	return nil
}

// TypeName returns the registry type name for serialization.
func (b *DividerBlock) TypeName() string { return "divider" }

// codeBlockPlugin registers a custom CodeBlock with language field.
type codeBlockPlugin struct{}

func (c *codeBlockPlugin) Name() string { return "code-block-plugin" }
func (c *codeBlockPlugin) Init(r *Registry) error {
	r.Register("code_block", func(id string) Block {
		return NewCodeBlock(id)
	})
	return nil
}

// CodeBlock is a custom block for code snippets with language annotation.
type CodeBlock struct {
	BaseBlock
	language string
	code     string
}

func NewCodeBlock(id string) *CodeBlock {
	return &CodeBlock{
		BaseBlock: NewBaseBlock(id, BlockType(100)),
	}
}

func (b *CodeBlock) Measure(cs component.Constraints) component.Size {
	lines := strings.Count(b.code, "\n") + 1
	if b.code == "" {
		lines = 1
	}
	return component.Size{W: cs.MaxWidth, H: lines + 1} // +1 for language header
}

func (b *CodeBlock) Paint(buf *buffer.Buffer) {
	r := b.Bounds()
	buf.DrawText(r.X, r.Y, "["+b.language+"]", buffer.Style{
		Fg:    buffer.RGB(0x50, 0xFA, 0x7B),
		Flags: buffer.Bold,
	})
}

func (b *CodeBlock) SerializeState() (json.RawMessage, error) {
	return json.Marshal(map[string]any{
		"language": b.language,
		"code":     b.code,
	})
}

func (b *CodeBlock) DeserializeState(data json.RawMessage) error {
	var s struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	b.language = s.Language
	b.code = s.Code
	return nil
}

// TypeName returns the registry type name for serialization.
func (b *CodeBlock) TypeName() string { return "code_block" }

// errorPlugin always fails during Init.
type errorPlugin struct{}

func (e *errorPlugin) Name() string { return "error-plugin" }
func (e *errorPlugin) Init(r *Registry) error {
	return errPluginInitFailed
}

var errPluginInitFailed = errString("plugin init failed")

type errString string

func (e errString) Error() string { return string(e) }

// --- Tests ---

func TestPluginLoad(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	err := pm.Load(&dividerPlugin{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Verify the custom block type is registered.
	b, err := r.Create("divider", "div-1")
	if err != nil {
		t.Fatalf("Create divider: %v", err)
	}
	if b.ID() != "div-1" {
		t.Errorf("ID = %q, want div-1", b.ID())
	}
	if _, ok := b.(*DividerBlock); !ok {
		t.Errorf("expected *DividerBlock, got %T", b)
	}
}

func TestPluginManagerLoadMultiple(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	plugins := []Plugin{
		&dividerPlugin{},
		&codeBlockPlugin{},
	}
	if err := pm.LoadAll(plugins); err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if pm.Count() != 2 {
		t.Errorf("Count = %d, want 2", pm.Count())
	}

	// Both types should be registered.
	if _, err := r.Create("divider", "d1"); err != nil {
		t.Errorf("Create divider: %v", err)
	}
	if _, err := r.Create("code_block", "c1"); err != nil {
		t.Errorf("Create code_block: %v", err)
	}
}

func TestPluginInit(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	plugin := &codeBlockPlugin{}
	if err := pm.Load(plugin); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Init should have registered the type — verify via Registry.
	b, err := r.Create("code_block", "cb-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	cb, ok := b.(*CodeBlock)
	if !ok {
		t.Fatalf("expected *CodeBlock, got %T", b)
	}
	if cb.ID() != "cb-1" {
		t.Errorf("ID = %q", cb.ID())
	}
}

func TestPluginSerialization(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)
	if err := pm.Load(&codeBlockPlugin{}); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Create and populate a custom block.
	original, _ := r.Create("code_block", "cb-1")
	cb := original.(*CodeBlock)
	cb.language = "go"
	cb.code = "func main() {}"
	cb.Complete()

	// Serialize.
	data, err := cb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	// Deserialize into a fresh block.
	loaded, _ := r.Create("code_block", "cb-1")
	cb2 := loaded.(*CodeBlock)
	if err := cb2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}
	if cb2.language != "go" {
		t.Errorf("language = %q, want 'go'", cb2.language)
	}
	if cb2.code != "func main() {}" {
		t.Errorf("code = %q", cb2.code)
	}
}

func TestPluginSerialization_ContainerRoundTrip(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)
	if err := pm.Load(&codeBlockPlugin{}); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if err := pm.Load(&dividerPlugin{}); err != nil {
		t.Fatalf("Load divider: %v", err)
	}

	// Build container with custom blocks.
	c := NewBlockContainer()
	cb, _ := r.Create("code_block", "cb-1")
	cb.(*CodeBlock).language = "rust"
	cb.(*CodeBlock).code = "fn main() {}"
	cb.Complete()

	div, _ := r.Create("divider", "div-1")
	div.Complete()

	c.AddBlock(cb)
	c.AddBlock(div)

	// Save.
	data, err := SaveContainer(c, r)
	if err != nil {
		t.Fatalf("SaveContainer: %v", err)
	}

	// Load.
	loaded, err := LoadContainer(data, r)
	if err != nil {
		t.Fatalf("LoadContainer: %v", err)
	}
	if loaded.Len() != 2 {
		t.Fatalf("loaded Len = %d, want 2", loaded.Len())
	}

	// Verify the code block.
	b0 := loaded.Blocks()[0]
	cb2, ok := b0.(*CodeBlock)
	if !ok {
		t.Fatalf("block[0] type = %T, want *CodeBlock", b0)
	}
	if cb2.language != "rust" {
		t.Errorf("language = %q, want 'rust'", cb2.language)
	}
	if cb2.code != "fn main() {}" {
		t.Errorf("code = %q", cb2.code)
	}

	// Verify the divider block.
	b1 := loaded.Blocks()[1]
	if _, ok := b1.(*DividerBlock); !ok {
		t.Fatalf("block[1] type = %T, want *DividerBlock", b1)
	}
}

func TestPluginName(t *testing.T) {
	p := &dividerPlugin{}
	if p.Name() != "divider-plugin" {
		t.Errorf("Name = %q, want 'divider-plugin'", p.Name())
	}

	p2 := &codeBlockPlugin{}
	if p2.Name() != "code-block-plugin" {
		t.Errorf("Name = %q, want 'code-block-plugin'", p2.Name())
	}
}

func TestPluginManagerPlugins(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	_ = pm.Load(&dividerPlugin{})
	_ = pm.Load(&codeBlockPlugin{})

	list := pm.Plugins()
	if len(list) != 2 {
		t.Fatalf("Plugins len = %d, want 2", len(list))
	}
	if list[0].Name() != "divider-plugin" {
		t.Errorf("plugins[0].Name = %q", list[0].Name())
	}
	if list[1].Name() != "code-block-plugin" {
		t.Errorf("plugins[1].Name = %q", list[1].Name())
	}
}

func TestPluginError(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	err := pm.Load(&errorPlugin{})
	if err == nil {
		t.Fatal("expected error from errorPlugin")
	}
	if !strings.Contains(err.Error(), "init failed") {
		t.Errorf("error should contain 'init failed': %v", err)
	}
	// Error plugin should NOT be in the loaded list.
	if pm.HasPlugin("error-plugin") {
		t.Error("error plugin should not be in loaded list")
	}
	if pm.Count() != 0 {
		t.Errorf("Count = %d, want 0", pm.Count())
	}
}

func TestPluginDuplicateLoad(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	if err := pm.Load(&dividerPlugin{}); err != nil {
		t.Fatalf("first Load: %v", err)
	}
	err := pm.Load(&dividerPlugin{})
	if err == nil {
		t.Fatal("expected error loading duplicate plugin")
	}
	if !strings.Contains(err.Error(), "already loaded") {
		t.Errorf("error should contain 'already loaded': %v", err)
	}
}

func TestCustomBlockExample_CodeBlock(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)
	if err := pm.Load(&codeBlockPlugin{}); err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Create via registry.
	b, err := r.Create("code_block", "example-1")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	cb := b.(*CodeBlock)
	cb.language = "python"
	cb.code = "print('hello')"
	cb.Complete()

	// Measure works.
	cs := component.Constraints{MaxWidth: 80}
	size := cb.Measure(cs)
	if size.H < 2 {
		t.Errorf("Measure height = %d, want >= 2", size.H)
	}

	// Paint doesn't panic.
	buf := buffer.NewBuffer(80, 10)
	cb.SetBounds(component.Rect{X: 0, Y: 0, W: 80, H: 2})
	cb.Paint(buf)

	// Serialize.
	data, _ := cb.SerializeState()
	var s struct {
		Language string `json:"language"`
		Code     string `json:"code"`
	}
	json.Unmarshal(data, &s)
	if s.Language != "python" {
		t.Errorf("language = %q, want 'python'", s.Language)
	}
	if s.Code != "print('hello')" {
		t.Errorf("code = %q", s.Code)
	}
}

func TestPluginManager_NilRegistry(t *testing.T) {
	pm := NewPluginManager(nil)
	if pm.Registry() == nil {
		t.Error("Registry should be non-nil even with nil input")
	}
}

func TestPluginManager_LoadNil(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)
	err := pm.Load(nil)
	if err == nil {
		t.Fatal("expected error loading nil plugin")
	}
}

func TestPluginManager_LoadAllPartialFailure(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	plugs := []Plugin{
		&dividerPlugin{},   // succeeds
		&errorPlugin{},     // fails
		&codeBlockPlugin{}, // should not be loaded
	}
	err := pm.LoadAll(plugs)
	if err == nil {
		t.Fatal("expected error")
	}
	// Only the first plugin should be loaded.
	if pm.Count() != 1 {
		t.Errorf("Count = %d, want 1", pm.Count())
	}
	if !pm.HasPlugin("divider-plugin") {
		t.Error("divider-plugin should be loaded")
	}
	if pm.HasPlugin("code-block-plugin") {
		t.Error("code-block-plugin should NOT be loaded after failure")
	}
}

func TestPluginManager_HasPlugin(t *testing.T) {
	r := NewDefaultRegistry()
	pm := NewPluginManager(r)

	_ = pm.Load(&dividerPlugin{})

	if !pm.HasPlugin("divider-plugin") {
		t.Error("HasPlugin should return true for loaded plugin")
	}
	if pm.HasPlugin("nonexistent") {
		t.Error("HasPlugin should return false for unknown plugin")
	}
}
