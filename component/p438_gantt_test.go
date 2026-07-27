package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestGanttChart_New_P438(t *testing.T) {
	gc := NewGanttChart()
	if gc.TaskCount() != 0 {
		t.Errorf("TaskCount = %d, want 0", gc.TaskCount())
	}
	if gc.MaxUnit() != 60 {
		t.Errorf("MaxUnit = %d, want 60", gc.MaxUnit())
	}
	if gc.LabelWidth() != 12 {
		t.Errorf("LabelWidth = %d, want 12", gc.LabelWidth())
	}
}

func TestGanttChart_AddTask_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "Task A", Start: 0, End: 10})
	gc.AddTask(GanttTask{Label: "Task B", Start: 5, End: 20})
	if gc.TaskCount() != 2 {
		t.Errorf("TaskCount = %d, want 2", gc.TaskCount())
	}
	// maxUnit should expand to end of last task
	if gc.MaxUnit() < 20 {
		t.Errorf("MaxUnit = %d, want >= 20", gc.MaxUnit())
	}
}

func TestGanttChart_SetTasks_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "X", Start: 0, End: 5})
	gc.SetTasks([]GanttTask{
		{Label: "A", Start: 0, End: 10},
		{Label: "B", Start: 10, End: 20},
		{Label: "C", Start: 20, End: 30},
	})
	if gc.TaskCount() != 3 {
		t.Errorf("TaskCount = %d, want 3", gc.TaskCount())
	}
}

func TestGanttChart_SetMaxUnit_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.SetMaxUnit(100)
	if gc.MaxUnit() != 100 {
		t.Errorf("MaxUnit = %d, want 100", gc.MaxUnit())
	}
}

func TestGanttChart_SetLabelWidth_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.SetLabelWidth(20)
	if gc.LabelWidth() != 20 {
		t.Errorf("LabelWidth = %d, want 20", gc.LabelWidth())
	}
	// Too small should be ignored
	gc.SetLabelWidth(2)
	if gc.LabelWidth() != 20 {
		t.Errorf("LabelWidth = %d, want 20 (ignored)", gc.LabelWidth())
	}
}

func TestGanttChart_Clear_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "X", Start: 0, End: 5})
	gc.Clear()
	if gc.TaskCount() != 0 {
		t.Error("should have 0 tasks after Clear")
	}
}

func TestGanttChart_Tasks_P438(t *testing.T) {
	gc := NewGanttChart()
	tasks := []GanttTask{{Label: "A", Start: 0, End: 10}}
	gc.SetTasks(tasks)
	got := gc.Tasks()
	if len(got) != 1 || got[0].Label != "A" {
		t.Errorf("Tasks mismatch: %v", got)
	}
}

func TestGanttChart_Style_P438(t *testing.T) {
	gc := NewGanttChart()
	st := DefaultGanttChartStyle()
	gc.SetStyle(st)
	if gc.Style().Header.Fg != st.Header.Fg {
		t.Error("style mismatch")
	}
}

func TestGanttChart_Measure_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "X", Start: 0, End: 5})
	sz := gc.Measure(Constraints{})
	if sz.H < 3 {
		t.Errorf("H = %d, want >= 3", sz.H)
	}
	if sz.W < 20 {
		t.Errorf("W = %d, want >= 20", sz.W)
	}
}

func TestGanttChart_Paint_NoPanic_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "Design", Start: 0, End: 14, Color: buffer.Cyan, Progress: 0.5})
	gc.AddTask(GanttTask{Label: "Build", Start: 10, End: 40, Color: buffer.Green, Progress: 0.3})
	gc.AddTask(GanttTask{Label: "Test", Start: 35, End: 55, Color: buffer.Yellow, Progress: 0.0})
	gc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	gc.Paint(buf)
}

func TestGanttChart_Paint_ZeroBounds_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	gc.Paint(buf)
}

func TestGanttChart_Paint_TinyBounds_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.SetBounds(Rect{X: 0, Y: 0, W: 5, H: 1})
	buf := buffer.NewBuffer(5, 1)
	gc.Paint(buf)
}

func TestGanttChart_String_P438(t *testing.T) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "X", Start: 0, End: 5})
	s := gc.String()
	if s == "" {
		t.Error("String should not be empty")
	}
}

func TestGanttChart_Children_P438(t *testing.T) {
	if NewGanttChart().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkGanttChart_Paint_P438(b *testing.B) {
	gc := NewGanttChart()
	gc.AddTask(GanttTask{Label: "Design", Start: 0, End: 14, Color: buffer.Cyan, Progress: 0.5})
	gc.AddTask(GanttTask{Label: "Build", Start: 10, End: 40, Color: buffer.Green, Progress: 0.3})
	gc.AddTask(GanttTask{Label: "Test", Start: 35, End: 55, Color: buffer.Yellow, Progress: 0.0})
	gc.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 10})
	buf := buffer.NewBuffer(60, 10)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gc.Paint(buf)
	}
}
