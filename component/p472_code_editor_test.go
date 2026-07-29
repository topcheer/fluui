package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestCodeEditorBasic(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetLanguage("go")
	ce.SetCode("package main\n\nfunc main() {\n    println(\"hello\")\n}\n")

	if ce.Language() != "go" {
		t.Errorf("Language = %q, want go", ce.Language())
	}
	if ce.LineCount() != 5 {
		t.Errorf("LineCount = %d, want 5", ce.LineCount())
	}
}

func TestCodeEditorSetCode(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetCode("line1\nline2\nline3")
	if ce.Code() != "line1\nline2\nline3" {
		t.Errorf("Code = %q", ce.Code())
	}
	if ce.LineCount() != 3 {
		t.Errorf("LineCount = %d, want 3", ce.LineCount())
	}
}

func TestCodeEditorEmptyCode(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetCode("")
	if ce.LineCount() != 0 {
		t.Errorf("LineCount = %d, want 0", ce.LineCount())
	}
}

func TestCodeEditorLineNumbersToggle(t *testing.T) {
	ce := NewCodeEditor()
	if !ce.ShowLineNumbers() {
		t.Error("should show line numbers by default")
	}
	ce.SetShowLineNumbers(false)
	if ce.ShowLineNumbers() {
		t.Error("should not show line numbers after toggle off")
	}
}

func TestCodeEditorCursor(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetCode("a\nb\nc")
	ce.SetCursorLine(1)
	if ce.CursorLine() != 1 {
		t.Errorf("CursorLine = %d, want 1", ce.CursorLine())
	}
	ce.SetCursorLine(-1)
	if ce.CursorLine() != -1 {
		t.Errorf("CursorLine = %d, want -1", ce.CursorLine())
	}
}

func TestCodeEditorMeasure(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetCode("line1\nline2\nline3")
	s := ce.Measure(Constraints{})
	if s.W < 20 {
		t.Errorf("W = %d, want >= 20", s.W)
	}
	if s.H < 5 {
		t.Errorf("H = %d, want >= 5", s.H)
	}
}

func TestCodeEditorPaint(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetLanguage("go")
	ce.SetCode("package main\n\nfunc main() {\n}\n")
	ce.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 8})

	buf := buffer.NewBuffer(60, 8)
	ce.Paint(buf)

	// Border
	if buf.GetCell(0, 0).Rune != '┌' {
		t.Error("top-left corner missing")
	}

	// Line number "1" should exist somewhere near top-left
	foundLineNum := false
	for x := 1; x < 6; x++ {
		if buf.GetCell(x, 1).Rune == '1' {
			foundLineNum = true
			break
		}
	}
	if !foundLineNum {
		t.Error("line number 1 not found")
	}

	// "package" keyword should be on line 1
	foundKeyword := false
	for x := 4; x < 60; x++ {
		if buf.GetCell(x, 1).Rune == 'p' {
			// Check if it's styled as keyword (violet bold)
			cell := buf.GetCell(x, 1)
			if cell.Flags&buffer.Bold != 0 {
				foundKeyword = true
				break
			}
		}
	}
	if !foundKeyword {
		t.Error("keyword 'package' not highlighted")
	}
}

func TestCodeEditorPaintWithCursor(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetCode("line1\nline2\nline3")
	ce.SetCursorLine(1)
	ce.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 6})

	buf := buffer.NewBuffer(60, 6)
	ce.Paint(buf) // should not panic

	// Cursor line should have underline flag
	foundUnderline := false
	for x := 5; x < 60; x++ {
		cell := buf.GetCell(x, 2) // line 2 (0-indexed = 1)
		if cell.Flags&buffer.Underline != 0 {
			foundUnderline = true
			break
		}
	}
	if !foundUnderline {
		t.Error("cursor underline not found on line 2")
	}
}

func TestCodeEditorPaintEmpty(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	ce.Paint(buf) // should not panic with no code
}

func TestCodeEditorPaintNoLineNumbers(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetShowLineNumbers(false)
	ce.SetCode("hello world")
	ce.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 4})
	buf := buffer.NewBuffer(60, 4)
	ce.Paint(buf) // should not panic
}

func TestCodeEditorChildren(t *testing.T) {
	ce := NewCodeEditor()
	if ce.Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestCodeEditorMultipleLanguages(t *testing.T) {
	ce := NewCodeEditor()
	// Python
	ce.SetLanguage("python")
	ce.SetCode("def hello():\n    return True")
	if ce.LineCount() != 2 {
		t.Errorf("Python LineCount = %d, want 2", ce.LineCount())
	}

	// JavaScript
	ce.SetLanguage("javascript")
	ce.SetCode("const x = 1;\nfunction foo() {}")
	if ce.LineCount() != 2 {
		t.Errorf("JS LineCount = %d, want 2", ce.LineCount())
	}
}

func TestCodeEditorStyle(t *testing.T) {
	ce := NewCodeEditor()
	ce.SetStyle(CodeEditorStyle{
		Normal:     buffer.Style{Fg: buffer.RGB(200, 200, 200)},
		Keyword:    buffer.Style{Fg: buffer.RGB(255, 0, 0), Flags: buffer.Bold},
		String:     buffer.Style{Fg: buffer.RGB(0, 255, 0)},
		Comment:    buffer.Style{Fg: buffer.RGB(100, 100, 100)},
		Number:     buffer.Style{Fg: buffer.RGB(255, 165, 0)},
		LineNumber: buffer.Style{Fg: buffer.RGB(128, 128, 128)},
		Cursor:     buffer.Style{Fg: buffer.RGB(255, 255, 0), Flags: buffer.Underline},
		Border:     buffer.Style{Fg: buffer.RGB(64, 64, 64)},
	})
	ce.SetCode("// test\nvar x = 42")
	ce.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	ce.Paint(buf)
}
