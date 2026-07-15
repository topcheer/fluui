package ansi

import (
	"testing"
)

func TestStyleConfigFields_P200(t *testing.T) {
	sc := StyleConfig{}
	// Verify all fields that ggcode accesses are present
	accent := "#7aa2f7"
	sc.Code.Color = &accent
	sc.Code.BackgroundColor = nil
	sc.CodeBlock.Color = &accent
	// Heading access
	_ = sc.Heading[0]
	// Hr
	_ = sc.Hr
	// BlockQuote
	_ = sc.BlockQuote
}

func TestChromaStyleConfigFields_P200(t *testing.T) {
	c := ChromaStyleConfig{}
	accent := "#7aa2f7"
	text := "#a9b1d6"
	c.Text.Color = &text
	c.Error.BackgroundColor = nil
	c.Comment.Color = &accent
	c.Keyword.Color = &accent
	c.KeywordReserved.Color = &accent
	c.LiteralString.Color = &text
}

func TestPtrHelpers_P200(t *testing.T) {
	s := Ptr("test")
	if *s != "test" {
		t.Error("Ptr failed")
	}
	b := PtrBool(true)
	if !*b {
		t.Error("PtrBool failed")
	}
	u := PtrUint(5)
	if *u != 5 {
		t.Error("PtrUint failed")
	}
}