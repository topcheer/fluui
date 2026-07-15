package tree

import (
	"strings"
	"testing"

	"github.com/topcheer/fluui/compat/lipgloss"
)

func TestRootAndChildren(t *testing.T) {
	tr := Root("header")
	tr.Child("child1")
	tr.Child("child2")
	result := tr.String()
	if !strings.Contains(result, "header") {
		t.Error("should contain root")
	}
	if !strings.Contains(result, "child1") {
		t.Error("should contain child1")
	}
	if !strings.Contains(result, "child2") {
		t.Error("should contain child2")
	}
}

func TestMultiLineChild(t *testing.T) {
	tr := Root("root")
	tr.Child("line1\nline2")
	result := tr.String()
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d: %q", len(lines), result)
	}
}

func TestEnumeratorStyle(t *testing.T) {
	tr := Root("root")
	tr.EnumeratorStyle(lipgloss.NewStyle())
	_ = tr.String() // should not panic
}

func TestIndent(t *testing.T) {
	tr := Root("root")
	tr.Child("child")
	tr.Indent(4)
	_ = tr.String() // should not panic
}

func TestEmptyTree(t *testing.T) {
	tr := Root("root")
	result := tr.String()
	if result != "root" {
		t.Errorf("empty tree should return just root, got %q", result)
	}
}