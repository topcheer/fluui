package component

import (
	"sync"
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

// ─── StyleDecl Tests ───

func TestStyleSheet_StyleDecl_ToStyle_Full(t *testing.T) {
	decl := StyleDecl{
		Fg:            buffer.NamedColor(buffer.NamedRed),
		Bg:            buffer.NamedColor(buffer.NamedBlue),
		Bold:          boolPtr(true),
		Italic:        boolPtr(true),
		Underline:     boolPtr(true),
		Strikethrough: boolPtr(true),
	}
	s := decl.ToStyle(buffer.Style{})
	if !s.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("Fg should be red")
	}
	if !s.Bg.Equal(buffer.NamedColor(buffer.NamedBlue)) {
		t.Error("Bg should be blue")
	}
	if s.Flags&buffer.Bold == 0 {
		t.Error("Bold should be set")
	}
	if s.Flags&buffer.Italic == 0 {
		t.Error("Italic should be set")
	}
	if s.Flags&buffer.Underline == 0 {
		t.Error("Underline should be set")
	}
	if s.Flags&buffer.Strikethrough == 0 {
		t.Error("Strikethrough should be set")
	}
}

func TestStyleSheet_StyleDecl_ToStyle_Unset(t *testing.T) {
	// All nil → should inherit from base
	base := buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedGreen),
		Flags: buffer.Bold,
	}
	decl := StyleDecl{}
	s := decl.ToStyle(base)
	if !s.Fg.Equal(buffer.NamedColor(buffer.NamedGreen)) {
		t.Error("Fg should inherit green from base")
	}
	if s.Flags&buffer.Bold == 0 {
		t.Error("Bold should inherit from base")
	}
}

func TestStyleSheet_StyleDecl_ToStyle_ClearFlag(t *testing.T) {
	// Set bold=false → should clear the flag even if base has it
	base := buffer.Style{Flags: buffer.Bold | buffer.Italic}
	decl := StyleDecl{Bold: boolPtr(false)}
	s := decl.ToStyle(base)
	if s.Flags&buffer.Bold != 0 {
		t.Error("Bold should be cleared")
	}
	if s.Flags&buffer.Italic == 0 {
		t.Error("Italic should be inherited")
	}
}

func TestStyleSheet_StyleDecl_Padding(t *testing.T) {
	decl := StyleDecl{
		PaddingTop:    intPtr(1),
		PaddingBottom: intPtr(2),
		PaddingLeft:   intPtr(3),
		PaddingRight:  intPtr(4),
	}
	top, bottom, left, right := decl.Padding()
	if top != 1 || bottom != 2 || left != 3 || right != 4 {
		t.Errorf("padding mismatch: %d,%d,%d,%d", top, bottom, left, right)
	}

	// No padding
	decl2 := StyleDecl{}
	top, bottom, left, right = decl2.Padding()
	if top != 0 || bottom != 0 || left != 0 || right != 0 {
		t.Error("empty decl should have zero padding")
	}
}

// ─── StyleSheet Registration Tests ───

func TestStyleSheet_Add(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed), Bold: boolPtr(true)})

	if ss.Count() != 1 {
		t.Errorf("expected 1 class, got %d", ss.Count())
	}
	if !ss.Has(".error") {
		t.Error("Has should return true")
	}
	if ss.Has(".nonexistent") {
		t.Error("Has should return false for nonexistent")
	}
}

func TestStyleSheet_Add_WithoutDot(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add("error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed)})

	// Should work with or without dot
	if !ss.Has(".error") {
		t.Error("Has with dot should work")
	}
	if !ss.Has("error") {
		t.Error("Has without dot should work")
	}
}

func TestStyleSheet_Add_Overwrite(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed)})
	ss.Add(".error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedBlue)})

	decl, ok := ss.Get(".error")
	if !ok {
		t.Fatal("expected .error to exist")
	}
	if !decl.Fg.Equal(buffer.NamedColor(buffer.NamedBlue)) {
		t.Error("should have blue (overwritten)")
	}
	if ss.Count() != 1 {
		t.Errorf("expected 1 class after overwrite, got %d", ss.Count())
	}
}

func TestStyleSheet_Remove(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed)})

	if !ss.Remove(".error") {
		t.Error("Remove should return true")
	}
	if ss.Has(".error") {
		t.Error("Has should return false after Remove")
	}
	if ss.Remove(".error") {
		t.Error("Remove should return false for nonexistent")
	}
}

func TestStyleSheet_Get(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed), Bold: boolPtr(true)})

	decl, ok := ss.Get(".error")
	if !ok {
		t.Fatal("expected .error to exist")
	}
	if !decl.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("Fg should be red")
	}
	if decl.Bold == nil || !*decl.Bold {
		t.Error("Bold should be true")
	}

	_, ok = ss.Get(".nonexistent")
	if ok {
		t.Error("Get should return false for nonexistent")
	}
}

func TestStyleSheet_Classes(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{})
	ss.Add(".success", StyleDecl{})
	ss.Add(".warning", StyleDecl{})

	classes := ss.Classes()
	if len(classes) != 3 {
		t.Errorf("expected 3 classes, got %d", len(classes))
	}
	// Should be sorted
	if classes[0] != ".error" {
		t.Errorf("expected .error first, got %q", classes[0])
	}
}

// ─── Resolve Tests ───

func TestStyleSheet_Resolve_SingleClass(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{
		Fg:   buffer.NamedColor(buffer.NamedRed),
		Bold: boolPtr(true),
	})

	style := ss.Resolve(".error")
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("Fg should be red")
	}
	if style.Flags&buffer.Bold == 0 {
		t.Error("Bold should be set")
	}
}

func TestStyleSheet_Resolve_MultipleClasses(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed)})
	ss.Add(".bold", StyleDecl{Bold: boolPtr(true)})

	// Merge .error + .bold
	style := ss.Resolve("error bold")
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("Fg should be red")
	}
	if style.Flags&buffer.Bold == 0 {
		t.Error("Bold should be set")
	}
}

func TestStyleSheet_Resolve_LaterOverrides(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".red", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed)})
	ss.Add(".blue", StyleDecl{Fg: buffer.NamedColor(buffer.NamedBlue)})

	// .blue comes later → should override
	style := ss.Resolve("red blue")
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedBlue)) {
		t.Error("Fg should be blue (later class wins)")
	}
}

func TestStyleSheet_Resolve_WithBase(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".bold", StyleDecl{Bold: boolPtr(true)})

	base := buffer.Style{
		Fg:    buffer.NamedColor(buffer.NamedGreen),
		Flags: buffer.Italic,
	}
	style := ss.ResolveWithBase(".bold", base)
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedGreen)) {
		t.Error("Fg should inherit green from base")
	}
	if style.Flags&buffer.Bold == 0 {
		t.Error("Bold should be set from class")
	}
	if style.Flags&buffer.Italic == 0 {
		t.Error("Italic should be inherited from base")
	}
}

func TestStyleSheet_Resolve_NonexistentClass(t *testing.T) {
	ss := NewStyleSheet()
	style := ss.Resolve(".nonexistent")
	// Should return default style
	if style.Flags != 0 {
		t.Error("nonexistent class should return default style")
	}
}

func TestStyleSheet_Resolve_EmptyString(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".red", StyleDecl{Fg: buffer.NamedColor(buffer.NamedRed)})
	style := ss.Resolve("")
	if style.Flags != 0 || style.Fg.Type != 0 {
		t.Error("empty class string should return default style")
	}
}

func TestStyleSheet_ResolveDecl(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{
		Fg:          buffer.NamedColor(buffer.NamedRed),
		PaddingTop:  intPtr(2),
		PaddingLeft: intPtr(1),
	})

	decl := ss.ResolveDecl(".error")
	if decl.PaddingTop == nil || *decl.PaddingTop != 2 {
		t.Error("PaddingTop should be 2")
	}
	if decl.PaddingLeft == nil || *decl.PaddingLeft != 1 {
		t.Error("PaddingLeft should be 1")
	}
}

// ─── Apply Tests ───

type testStylable struct {
	style buffer.Style
}

func (ts *testStylable) SetStyle(s buffer.Style) { ts.style = s }

func TestStyleSheet_Apply(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".error", StyleDecl{
		Fg:   buffer.NamedColor(buffer.NamedRed),
		Bold: boolPtr(true),
	})

	ts := &testStylable{}
	style := ss.Apply(ts, ".error")
	if !ts.style.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("component style should have red Fg")
	}
	if ts.style.Flags&buffer.Bold == 0 {
		t.Error("component style should have Bold")
	}
	if !style.Fg.Equal(ts.style.Fg) {
		t.Error("returned style should match component style")
	}
}

// ─── Theme Integration Tests ───

func TestStyleSheet_FromTheme(t *testing.T) {
	colorFn := func(name string) buffer.Color {
		switch name {
		case "error":
			return buffer.NamedColor(buffer.NamedRed)
		case "success":
			return buffer.NamedColor(buffer.NamedGreen)
		case "accent":
			return buffer.NamedColor(buffer.NamedCyan)
		default:
			return buffer.Color{}
		}
	}
	ss := StyleSheetFromTheme(colorFn)

	if !ss.Has(".error") {
		t.Error("should have .error class")
	}
	if !ss.Has(".bold") {
		t.Error("should have .bold class")
	}
	if !ss.Has(".italic") {
		t.Error("should have .italic class")
	}

	// .error should resolve to red
	style := ss.Resolve(".error")
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error(".error should be red")
	}

	// .bold should have bold flag
	style = ss.Resolve(".bold")
	if style.Flags&buffer.Bold == 0 {
		t.Error(".bold should set Bold flag")
	}
}

func TestStyleSheet_FromTheme_NilFn(t *testing.T) {
	ss := StyleSheetFromTheme(nil)
	// Should still have attribute classes
	if !ss.Has(".bold") {
		t.Error("should have .bold class even with nil colorFn")
	}
}

// ─── Parse Tests ───

func TestStyleSheet_Parse(t *testing.T) {
	text := `
.error {
	fg: red
	bold: true
}
.success {
	fg: green
	underline: true
}
`
	ss, err := ParseStyleSheet(text)
	if err != nil {
		t.Fatalf("ParseStyleSheet failed: %v", err)
	}
	if ss.Count() != 2 {
		t.Errorf("expected 2 classes, got %d", ss.Count())
	}

	style := ss.Resolve(".error")
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error(".error fg should be red")
	}
	if style.Flags&buffer.Bold == 0 {
		t.Error(".error should have bold")
	}

	style = ss.Resolve(".success")
	if !style.Fg.Equal(buffer.NamedColor(buffer.NamedGreen)) {
		t.Error(".success fg should be green")
	}
	if style.Flags&buffer.Underline == 0 {
		t.Error(".success should have underline")
	}
}

func TestStyleSheet_Parse_HexColors(t *testing.T) {
	text := `
.custom {
	fg: #ff8800
	bg: #000033
}
`
	ss, err := ParseStyleSheet(text)
	if err != nil {
		t.Fatalf("ParseStyleSheet failed: %v", err)
	}
	style := ss.Resolve(".custom")
	// Just verify it parses without error and sets something
	if style.Fg.Type == 0 {
		t.Error("hex fg should be set")
	}
}

func TestStyleSheet_Parse_Padding(t *testing.T) {
	text := `
.padded {
	padding-top: 2
	padding-bottom: 3
}
`
	ss, _ := ParseStyleSheet(text)
	decl := ss.ResolveDecl(".padded")
	if decl.PaddingTop == nil || *decl.PaddingTop != 2 {
		t.Error("padding-top should be 2")
	}
	if decl.PaddingBottom == nil || *decl.PaddingBottom != 3 {
		t.Error("padding-bottom should be 3")
	}
}

func TestStyleSheet_Parse_Empty(t *testing.T) {
	ss, err := ParseStyleSheet("")
	if err != nil {
		t.Fatalf("ParseStyleSheet failed: %v", err)
	}
	if ss.Count() != 0 {
		t.Errorf("expected 0 classes, got %d", ss.Count())
	}
}

// ─── Helper Tests ───

func TestStyleSheet_NormalizeClassName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{".error", "error"},
		{"Error", "error"},
		{"  .warning  ", "warning"},
		{"success", "success"},
	}
	for _, tt := range tests {
		got := normalizeClassName(tt.input)
		if got != tt.want {
			t.Errorf("normalizeClassName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStyleSheet_ParseClasses(t *testing.T) {
	result := parseClasses("error bold underline")
	if len(result) != 3 {
		t.Errorf("expected 3 classes, got %d", len(result))
	}
	if result[0] != "error" || result[1] != "bold" || result[2] != "underline" {
		t.Errorf("unexpected classes: %v", result)
	}
}

func TestStyleSheet_ParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  *bool
	}{
		{"true", boolPtr(true)},
		{"false", boolPtr(false)},
		{"yes", boolPtr(true)},
		{"no", boolPtr(false)},
		{"1", boolPtr(true)},
		{"0", boolPtr(false)},
		{"on", boolPtr(true)},
		{"off", boolPtr(false)},
		{"invalid", nil},
	}
	for _, tt := range tests {
		got := parseBool(tt.input)
		if got == nil && tt.want != nil {
			t.Errorf("parseBool(%q) = nil, want %v", tt.input, *tt.want)
		} else if got != nil && tt.want == nil {
			t.Errorf("parseBool(%q) = %v, want nil", tt.input, *got)
		} else if got != nil && tt.want != nil && *got != *tt.want {
			t.Errorf("parseBool(%q) = %v, want %v", tt.input, *got, *tt.want)
		}
	}
}

func TestStyleSheet_ParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
		nil_  bool
	}{
		{"42", 42, false},
		{"0", 0, false},
		{"-5", -5, false},
		{"abc", 0, true},
	}
	for _, tt := range tests {
		got := parseInt(tt.input)
		if tt.nil_ {
			if got != nil {
				t.Errorf("parseInt(%q) = %v, want nil", tt.input, *got)
			}
		} else {
			if got == nil {
				t.Errorf("parseInt(%q) = nil, want %d", tt.input, tt.want)
			} else if *got != tt.want {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, *got, tt.want)
			}
		}
	}
}

func TestStyleSheet_MergeDecls(t *testing.T) {
	base := StyleDecl{
		Fg:     buffer.NamedColor(buffer.NamedRed),
		Bold:   boolPtr(true),
		PaddingTop: intPtr(1),
	}
	overlay := StyleDecl{
		Bg:         buffer.NamedColor(buffer.NamedBlue),
		PaddingTop: intPtr(5),
	}
	merged := mergeDecls(base, overlay)
	if !merged.Fg.Equal(buffer.NamedColor(buffer.NamedRed)) {
		t.Error("Fg should inherit from base")
	}
	if !merged.Bg.Equal(buffer.NamedColor(buffer.NamedBlue)) {
		t.Error("Bg should come from overlay")
	}
	if merged.PaddingTop == nil || *merged.PaddingTop != 5 {
		t.Error("PaddingTop should be overwritten to 5")
	}
}

func TestStyleSheet_String(t *testing.T) {
	ss := NewStyleSheet()
	ss.Add(".a", StyleDecl{})
	ss.Add(".b", StyleDecl{})
	s := ss.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}

// ─── Concurrency Tests ───

func TestStyleSheet_Concurrent(t *testing.T) {
	ss := NewStyleSheet()

	var wg sync.WaitGroup
	// Concurrent writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			ss.Add("class"+string(rune('a'+n)), StyleDecl{
				Fg: buffer.NamedColor(n % 15),
			})
		}(i)
	}
	// Concurrent readers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ss.Resolve("a b c")
			ss.Classes()
			ss.Count()
			_ = ss.Has(".test")
		}()
	}
	wg.Wait()
}
