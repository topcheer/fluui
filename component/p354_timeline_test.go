package component

import (
	"testing"
	"time"

	"github.com/topcheer/fluui/internal/buffer"
)

func TestP354_Timeline_Create(t *testing.T) {
	tl := NewTimeline([]TimelineEvent{
		{Title: "Started", Time: time.Now()},
		{Title: "Processing", Time: time.Now()},
	})
	if tl.EventCount() != 2 {
		t.Errorf("count = %d", tl.EventCount())
	}
}

func TestP354_Timeline_AddEvent(t *testing.T) {
	tl := NewTimeline(nil)
	tl.AddEvent(TimelineEvent{Title: "New", Time: time.Now()})
	if tl.EventCount() != 1 {
		t.Errorf("count = %d", tl.EventCount())
	}
}

func TestP354_Timeline_AddEvent_AutoTime(t *testing.T) {
	tl := NewTimeline(nil)
	tl.AddEvent(TimelineEvent{Title: "NoTime"})
	evs := tl.Events()
	if evs[0].Time.IsZero() {
		t.Error("time should be auto-set")
	}
}

func TestP354_Timeline_SetEvents(t *testing.T) {
	tl := NewTimeline(nil)
	tl.SetEvents([]TimelineEvent{{Title: "A"}, {Title: "B"}})
	if tl.EventCount() != 2 {
		t.Errorf("count = %d", tl.EventCount())
	}
}

func TestP354_Timeline_Clear(t *testing.T) {
	tl := NewTimeline([]TimelineEvent{{Title: "A"}})
	tl.Clear()
	if tl.EventCount() != 0 {
		t.Error("should be empty after clear")
	}
}

func TestP354_Timeline_SetMaxView(t *testing.T) {
	tl := NewTimeline(make([]TimelineEvent, 10))
	tl.SetMaxView(3)
	// Just verify no panic
}

func TestP354_Timeline_Measure(t *testing.T) {
	tl := NewTimeline([]TimelineEvent{{Title: "A"}, {Title: "B"}})
	s := tl.Measure(Constraints{MaxWidth: 50, MaxHeight: 10})
	if s.H < 3 {
		t.Errorf("height = %d, expected at least 3", s.H)
	}
}

func TestP354_Timeline_Paint(t *testing.T) {
	now := time.Date(2026, 2, 28, 14, 30, 0, 0, time.UTC)
	tl := NewTimeline([]TimelineEvent{
		{Title: "Request received", Time: now, Type: TimelineEventInfo},
		{Title: "Model loaded", Time: now, Type: TimelineEventSuccess},
		{Title: "Generating", Description: "Streaming response...", Time: now},
		{Title: "Failed", Time: now, Type: TimelineEventError},
	})
	tl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 15})
	buf := buffer.NewBuffer(50, 15)
	tl.Paint(buf)
	if buf.GetCell(0, 0).Rune == 0 {
		t.Error("expected non-empty cell")
	}
}

func TestP354_Timeline_Paint_Empty(t *testing.T) {
	tl := NewTimeline(nil)
	tl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 10})
	buf := buffer.NewBuffer(50, 10)
	tl.Paint(buf)
}

func TestP354_Timeline_Paint_ZeroBounds(t *testing.T) {
	tl := NewTimeline([]TimelineEvent{{Title: "A", Time: time.Now()}})
	tl.SetBounds(Rect{X: 0, Y: 0, W: 0, H: 0})
	buf := buffer.NewBuffer(50, 10)
	tl.Paint(buf)
}

func TestP354_Timeline_Paint_MaxView(t *testing.T) {
	events := make([]TimelineEvent, 20)
	for i := range events {
		events[i] = TimelineEvent{Title: "Event", Time: time.Now()}
	}
	tl := NewTimeline(events)
	tl.SetMaxView(3)
	tl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 20})
	buf := buffer.NewBuffer(50, 20)
	tl.Paint(buf)
}

func TestP354_Timeline_Paint_LongText(t *testing.T) {
	tl := NewTimeline([]TimelineEvent{
		{Title: "This is a very long title that needs truncation", Description: "And a very long description too", Time: time.Now()},
	})
	tl.SetBounds(Rect{X: 0, Y: 0, W: 15, H: 5})
	buf := buffer.NewBuffer(15, 5)
	tl.Paint(buf)
}

func TestP354_Timeline_EventMarkers(t *testing.T) {
	// Just verify no panic for all event types
	for _, et := range []TimelineEventType{
		TimelineEventDefault, TimelineEventSuccess,
		TimelineEventWarning, TimelineEventError, TimelineEventInfo,
	} {
		eventMarker(et)
	}
}

func TestP354_Timeline_AllTypes(t *testing.T) {
	now := time.Now()
	tl := NewTimeline([]TimelineEvent{
		{Title: "Default", Time: now, Type: TimelineEventDefault},
		{Title: "Success", Time: now, Type: TimelineEventSuccess},
		{Title: "Warning", Time: now, Type: TimelineEventWarning},
		{Title: "Error", Time: now, Type: TimelineEventError},
		{Title: "Info", Time: now, Type: TimelineEventInfo},
	})
	tl.SetBounds(Rect{X: 0, Y: 0, W: 50, H: 20})
	buf := buffer.NewBuffer(50, 20)
	tl.Paint(buf)
}

func BenchmarkTimeline_Paint(b *testing.B) {
	now := time.Now()
	events := make([]TimelineEvent, 10)
	for i := range events {
		events[i] = TimelineEvent{
			Title:       "Event",
			Description: "Description",
			Time:        now,
			Type:        TimelineEventInfo,
		}
	}
	tl := NewTimeline(events)
	tl.SetBounds(Rect{X: 0, Y: 0, W: 60, H: 25})
	buf := buffer.NewBuffer(60, 25)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tl.Paint(buf)
	}
}
