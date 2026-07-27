package component

import (
	"testing"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestProgressTimeline_New_P449(t *testing.T) {
	pt := NewProgressTimeline()
	if pt.MilestoneCount() != 0 {
		t.Errorf("MilestoneCount = %d, want 0", pt.MilestoneCount())
	}
	if pt.Progress() != 0 {
		t.Errorf("Progress = %v, want 0", pt.Progress())
	}
}

func TestProgressTimeline_AddMilestone_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "A", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "B", Status: MilestoneActive})
	if pt.MilestoneCount() != 2 {
		t.Errorf("MilestoneCount = %d, want 2", pt.MilestoneCount())
	}
}

func TestProgressTimeline_SetMilestones_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.SetMilestones([]TimelineMilestone{
		{Label: "A", Status: MilestoneDone},
		{Label: "B", Status: MilestoneDone},
		{Label: "C", Status: MilestonePending},
	})
	if pt.MilestoneCount() != 3 {
		t.Errorf("MilestoneCount = %d, want 3", pt.MilestoneCount())
	}
	if pt.DoneCount() != 2 {
		t.Errorf("DoneCount = %d, want 2", pt.DoneCount())
	}
}

func TestProgressTimeline_Progress_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "A", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "B", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "C", Status: MilestonePending})
	pt.AddMilestone(TimelineMilestone{Label: "D", Status: MilestonePending})
	if pt.Progress() != 0.5 {
		t.Errorf("Progress = %v, want 0.5", pt.Progress())
	}
}

func TestProgressTimeline_DoneCount_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "A", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "B", Status: MilestoneActive})
	if pt.DoneCount() != 1 {
		t.Errorf("DoneCount = %d, want 1", pt.DoneCount())
	}
}

func TestProgressTimeline_Milestones_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "X", Status: MilestoneDone})
	ms := pt.Milestones()
	if len(ms) != 1 || ms[0].Label != "X" {
		t.Errorf("Milestones mismatch: %v", ms)
	}
}

func TestProgressTimeline_Clear_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "X", Status: MilestoneDone})
	pt.Clear()
	if pt.MilestoneCount() != 0 {
		t.Error("should have 0 milestones after Clear")
	}
}

func TestProgressTimeline_Style_P449(t *testing.T) {
	pt := NewProgressTimeline()
	st := DefaultProgressTimelineStyle()
	pt.SetStyle(st)
	if pt.Style().Done.Fg != st.Done.Fg {
		t.Error("style mismatch")
	}
}

func TestProgressTimeline_Measure_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "A", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "B", Status: MilestonePending})
	sz := pt.Measure(Constraints{})
	if sz.H < 3 {
		t.Errorf("H = %d, want >= 3", sz.H)
	}
}

func TestProgressTimeline_Paint_NoPanic_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "Design", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "Build", Status: MilestoneActive})
	pt.AddMilestone(TimelineMilestone{Label: "Ship", Status: MilestonePending})
	pt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	pt.Paint(buf)
}

func TestProgressTimeline_Paint_AllDone_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "A", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "B", Status: MilestoneDone})
	pt.AddMilestone(TimelineMilestone{Label: "C", Status: MilestoneDone})
	pt.SetBounds(Rect{X: 0, Y: 0, W: 30, H: 4})
	buf := buffer.NewBuffer(30, 4)
	pt.Paint(buf)
}

func TestProgressTimeline_Paint_Empty_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.SetBounds(Rect{X: 0, Y: 0, W: 20, H: 3})
	buf := buffer.NewBuffer(20, 3)
	pt.Paint(buf)
}

func TestProgressTimeline_Paint_ZeroBounds_P449(t *testing.T) {
	pt := NewProgressTimeline()
	pt.AddMilestone(TimelineMilestone{Label: "X", Status: MilestoneDone})
	pt.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(1, 1)
	pt.Paint(buf)
}

func TestProgressTimeline_Children_P449(t *testing.T) {
	if NewProgressTimeline().Children() != nil {
		t.Error("Children should be nil")
	}
}

func BenchmarkProgressTimeline_Paint_P449(b *testing.B) {
	pt := NewProgressTimeline()
	for i := 0; i < 8; i++ {
		pt.AddMilestone(TimelineMilestone{Label: "M", Status: MilestoneStatus(i % 3)})
	}
	pt.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 5})
	buf := buffer.NewBuffer(60, 5)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pt.Paint(buf)
	}
}
