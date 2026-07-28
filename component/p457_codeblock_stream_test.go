package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestCodeBlockStream_New_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	if cbs.Language() != "go" {
		t.Errorf("Language = %q", cbs.Language())
	}
	if cbs.IsStreaming() {
		t.Error("should not be streaming initially")
	}
}

func TestCodeBlockStream_Start_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.Start()
	if !cbs.IsStreaming() {
		t.Error("should be streaming after Start")
	}
}

func TestCodeBlockStream_Append_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.Start()
	cbs.Append("package main\n")
	cbs.Append("func main() {}")
	if cbs.Code() != "package main\nfunc main() {}" {
		t.Errorf("Code = %q", cbs.Code())
	}
	if cbs.LineCount() != 2 {
		t.Errorf("LineCount = %d, want 2", cbs.LineCount())
	}
}

func TestCodeBlockStream_SetCode_P457(t *testing.T) {
	cbs := NewCodeBlockStream("python")
	cbs.SetCode("print('hello')")
	if cbs.Code() != "print('hello')" {
		t.Errorf("Code = %q", cbs.Code())
	}
}

func TestCodeBlockStream_Complete_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.Start()
	cbs.Append("x := 1")
	cbs.Complete()
	if cbs.IsStreaming() {
		t.Error("should not be streaming after Complete")
	}
	if !cbs.IsCompleted() {
		t.Error("should be completed")
	}
}

func TestCodeBlockStream_Language_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetLanguage("rust")
	if cbs.Language() != "rust" {
		t.Errorf("Language = %q", cbs.Language())
	}
}

func TestCodeBlockStream_LineNumbers_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	if !cbs.ShowLineNumbers() {
		t.Error("should show line numbers by default")
	}
	cbs.SetShowLineNumbers(false)
	if cbs.ShowLineNumbers() {
		t.Error("should be false")
	}
}

func TestCodeBlockStream_SetCursor_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetCursor(true) // no panic
}

func TestCodeBlockStream_SetStyle_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	st := DefaultCodeBlockStreamStyle()
	cbs.SetStyle(st)
}

func TestCodeBlockStream_Measure_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetCode("line1\nline2\nline3")
	sz := cbs.Measure(Constraints{})
	if sz.H < 3 {
		t.Errorf("H = %d, want >= 3", sz.H)
	}
}

func TestCodeBlockStream_Paint_Streaming_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.Start()
	cbs.Append("package main\n\nfunc main() {\n\tprintln(\"hello\") // greet\n}")
	cbs.SetCursor(true)
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	cbs.Paint(buf)
}

func TestCodeBlockStream_Paint_Completed_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetCode("x := 42")
	cbs.Complete()
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cbs.Paint(buf)
}

func TestCodeBlockStream_Paint_NoLineNumbers_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetCode("x := 1")
	cbs.SetShowLineNumbers(false)
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cbs.Paint(buf)
}

func TestCodeBlockStream_Paint_Empty_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	cbs.Paint(buf)
}

func TestCodeBlockStream_Paint_ZeroBounds_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	cbs.SetCode("x := 1")
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	cbs.Paint(buf)
}

func TestCodeBlockStream_Paint_AutoScroll_P457(t *testing.T) {
	cbs := NewCodeBlockStream("go")
	code := ""
	for i := 0; i < 30; i++ {
		code += "line\n"
	}
	cbs.SetCode(code)
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 40, H: 5})
	buf := buffer.NewBuffer(40, 5)
	cbs.Paint(buf) // should show last few lines
}

func TestCodeBlockStream_Children_P457(t *testing.T) {
	if NewCodeBlockStream("go").Children() != nil {
		t.Error("Children should be nil")
	}
}

func TestCountLinesInStr_P457(t *testing.T) {
	if countLinesInStr("") != 1 {
		t.Error("empty should be 1")
	}
	if countLinesInStr("a\nb\nc") != 3 {
		t.Error("3 lines expected")
	}
	if countLinesInStr("a") != 1 {
		t.Error("1 line expected")
	}
}

func BenchmarkCodeBlockStream_Paint_P457(b *testing.B) {
	cbs := NewCodeBlockStream("go")
	cbs.Start()
	cbs.Append("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n\tfor i := 0; i < 10; i++ {\n\t\tfmt.Println(i)\n\t}\n}\n")
	cbs.SetCursor(true)
	cbs.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 12})
	buf := buffer.NewBuffer(60, 12)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbs.Paint(buf)
	}
}
