package lipgloss

import "testing"

// P215: cover remaining branches in parseColor, namedColorIndex, JoinVertical

func TestColorToBuffer_None_P215(t *testing.T) {
	c := Color("") // empty = ColorNone
	result := parseColor(c)
	if result.Type != 0 {
		t.Error("empty color should map to buffer.ColorNone")
	}
}

func TestNamedColorIndex_AllBright_P215(t *testing.T) {
	brights := []string{"brightblack", "bright-black", "gray", "grey",
		"brightred", "bright-red", "brightgreen", "bright-green",
		"brightyellow", "bright-yellow", "brightblue", "bright-blue",
		"brightmagenta", "bright-magenta", "brightcyan", "bright-cyan",
		"brightwhite", "bright-white"}
	for _, name := range brights {
		idx := namedColorIndex(name)
		if idx == 0 && name != "black" {
			t.Errorf("namedColorIndex(%q) should return non-zero", name)
		}
	}
}

func TestNamedColorIndex_CaseInsensitive_P215(t *testing.T) {
	if namedColorIndex("RED") != namedColorIndex("red") {
		t.Error("should be case-insensitive")
	}
	if namedColorIndex("Green") != namedColorIndex("green") {
		t.Error("should be case-insensitive")
	}
}

func TestNamedColorIndex_Purple_P215(t *testing.T) {
	if namedColorIndex("purple") != namedColorIndex("magenta") {
		t.Error("purple should equal magenta")
	}
}

func TestJoinVertical_EmptyAndSingle_P215(t *testing.T) {
	if JoinVertical(Left) != "" {
		t.Error("empty should return empty")
	}
	if JoinVertical(Left, "only") != "only" {
		t.Error("single should return as-is")
	}
}