package recorder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// --- Recorder tests ---

func TestRecorder_New(t *testing.T) {
	r := NewRecorder()
	if r == nil {
		t.Fatal("expected non-nil recorder")
	}
	if r.IsRecording() {
		t.Error("expected not recording by default")
	}
	if r.EventCount() != 0 {
		t.Error("expected 0 events")
	}
}

func TestRecorder_StartStop(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	if !r.IsRecording() {
		t.Error("expected recording after Start")
	}
	r.Stop()
	if r.IsRecording() {
		t.Error("expected not recording after Stop")
	}
}

func TestRecorder_RecordKey(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordKey('a', 0, 0)
	r.RecordKey(0, 13, 0) // Enter

	events := r.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != EventKey {
		t.Error("expected EventKey type")
	}
	if events[0].KeyRune != 'a' {
		t.Errorf("expected rune 'a', got %q", string(events[0].KeyRune))
	}
	if events[1].KeyCode != 13 {
		t.Errorf("expected key code 13, got %d", events[1].KeyCode)
	}
}

func TestRecorder_RecordMouse(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordMouse(10, 5, 0, 1) // left click at (10,5)

	events := r.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventMouse {
		t.Error("expected EventMouse type")
	}
	if events[0].MouseX != 10 || events[0].MouseY != 5 {
		t.Errorf("expected (10,5), got (%d,%d)", events[0].MouseX, events[0].MouseY)
	}
}

func TestRecorder_RecordResize(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordResize(100, 40)

	events := r.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != EventResize {
		t.Error("expected EventResize type")
	}
	if events[0].Width != 100 || events[0].Height != 40 {
		t.Errorf("expected 100x40, got %dx%d", events[0].Width, events[0].Height)
	}
}

func TestRecorder_RecordFocus(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordFocus(true)
	r.RecordFocus(false)

	events := r.Events()
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if !events[0].Focused {
		t.Error("expected focused=true for first event")
	}
	if events[1].Focused {
		t.Error("expected focused=false for second event")
	}
}

func TestRecorder_RecordPaste(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordPaste("hello world")

	events := r.Events()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].PasteText != "hello world" {
		t.Errorf("expected 'hello world', got %q", events[0].PasteText)
	}
}

func TestRecorder_MixedEvents(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordKey('a', 0, 0)
	r.RecordMouse(5, 3, 0, 1)
	r.RecordResize(100, 50)
	r.RecordFocus(true)
	r.RecordPaste("text")

	events := r.Events()
	if len(events) != 5 {
		t.Fatalf("expected 5 events, got %d", len(events))
	}
}

func TestRecorder_NotRecording(t *testing.T) {
	r := NewRecorder()
	// Don't call Start
	r.RecordKey('a', 0, 0)
	if r.EventCount() != 0 {
		t.Error("expected 0 events when not recording")
	}
}

func TestRecorder_StopThenRecord(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)
	r.Stop()

	// Recording after Stop should be ignored
	r.RecordKey('b', 0, 0)
	if r.EventCount() != 1 {
		t.Errorf("expected 1 event after stop, got %d", r.EventCount())
	}
}

func TestRecorder_Timestamps(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	r.RecordKey('a', 0, 0)
	time.Sleep(2 * time.Millisecond)
	r.RecordKey('b', 0, 0)

	events := r.Events()
	if events[0].Timestamp >= events[1].Timestamp {
		t.Error("expected timestamps to be increasing")
	}
	if events[0].Timestamp < 0 {
		t.Error("expected non-negative timestamp")
	}
}

func TestRecorder_Events_DefensiveCopy(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)

	events1 := r.Events()
	events1[0].KeyRune = 'z'

	events2 := r.Events()
	if events2[0].KeyRune == 'z' {
		t.Error("expected defensive copy to prevent mutation")
	}
}

func TestRecorder_Clear(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)
	r.RecordKey('b', 0, 0)
	r.Clear()
	if r.EventCount() != 0 {
		t.Error("expected 0 events after clear")
	}
}

// --- Save/Load tests ---

func TestRecorder_SaveLoad(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)
	r.RecordMouse(10, 5, 0, 1)
	r.RecordResize(120, 40)

	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")

	err := r.Save(path)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists and is valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if rec.Version != 1 {
		t.Errorf("expected version 1, got %d", rec.Version)
	}
	if rec.Width != 80 || rec.Height != 24 {
		t.Errorf("expected 80x24, got %dx%d", rec.Width, rec.Height)
	}
	if len(rec.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(rec.Events))
	}
}

func TestRecorder_SaveBytes(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('x', 0, 0)

	data, err := r.SaveBytes()
	if err != nil {
		t.Fatalf("SaveBytes failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty data")
	}

	// Should be valid JSON
	var rec Recording
	if err := json.Unmarshal(data, &rec); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
}

func TestPlayer_Load(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)

	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")
	_ = r.Save(path)

	player, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if player.EventCount() != 1 {
		t.Errorf("expected 1 event, got %d", player.EventCount())
	}
	if player.Width() != 80 || player.Height() != 24 {
		t.Errorf("expected 80x24, got %dx%d", player.Width(), player.Height())
	}
}

func TestPlayer_LoadFromBytes(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)
	r.RecordKey('a', 0, 0)
	r.RecordMouse(5, 5, 0, 1)

	data, _ := r.SaveBytes()
	player, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}
	if player.EventCount() != 2 {
		t.Errorf("expected 2 events, got %d", player.EventCount())
	}
}

// --- Player tests ---

func TestPlayer_Basic(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Width:   80,
		Height:  24,
		Events: []RecordedEvent{
			{Type: EventKey, KeyRune: 'a', Timestamp: 0},
			{Type: EventKey, KeyRune: 'b', Timestamp: 1000000},
			{Type: EventKey, KeyRune: 'c', Timestamp: 2000000},
		},
	}
	p := NewPlayer(rec)

	if p.EventCount() != 3 {
		t.Errorf("expected 3 events, got %d", p.EventCount())
	}
	if !p.HasNext() {
		t.Error("expected HasNext=true")
	}
}

func TestPlayer_Next(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, KeyRune: 'a'},
			{Type: EventKey, KeyRune: 'b'},
		},
	}
	p := NewPlayer(rec)

	ev, ok := p.Next()
	if !ok || ev.KeyRune != 'a' {
		t.Errorf("expected first event 'a', got %q ok=%v", string(ev.KeyRune), ok)
	}

	p.Advance()

	ev, ok = p.Next()
	if !ok || ev.KeyRune != 'b' {
		t.Errorf("expected second event 'b', got %q ok=%v", string(ev.KeyRune), ok)
	}

	p.Advance()

	_, ok = p.Next()
	if ok {
		t.Error("expected ok=false after all events consumed")
	}
}

func TestPlayer_Reset(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, KeyRune: 'a'},
			{Type: EventKey, KeyRune: 'b'},
		},
	}
	p := NewPlayer(rec)

	p.Advance()
	p.Advance()
	if p.HasNext() {
		t.Error("expected no next after advancing past all")
	}

	p.Reset()
	if !p.HasNext() {
		t.Error("expected HasNext after Reset")
	}

	ev, _ := p.Next()
	if ev.KeyRune != 'a' {
		t.Errorf("expected 'a' after reset, got %q", string(ev.KeyRune))
	}
}

func TestPlayer_Cursor(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey},
			{Type: EventKey},
			{Type: EventKey},
		},
	}
	p := NewPlayer(rec)

	if p.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", p.Cursor())
	}
	p.Advance()
	if p.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", p.Cursor())
	}
	p.Advance()
	p.Advance()
	if p.Cursor() != 3 {
		t.Errorf("expected cursor 3, got %d", p.Cursor())
	}
}

func TestPlayer_Remaining(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey},
			{Type: EventKey},
			{Type: EventKey},
		},
	}
	p := NewPlayer(rec)

	if p.Remaining() != 3 {
		t.Errorf("expected 3 remaining, got %d", p.Remaining())
	}
	p.Advance()
	if p.Remaining() != 2 {
		t.Errorf("expected 2 remaining, got %d", p.Remaining())
	}
}

func TestPlayer_NextDelay(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, Timestamp: 0},
			{Type: EventKey, Timestamp: 5000000}, // 5ms after first
			{Type: EventKey, Timestamp: 8000000}, // 3ms after second
		},
	}
	p := NewPlayer(rec)

	// First event has no delay
	delay := p.NextDelay()
	if delay != 0 {
		t.Errorf("expected 0 delay for first event, got %v", delay)
	}

	p.Advance()

	// Second event: 5ms delay
	delay = p.NextDelay()
	if delay != 5*time.Millisecond {
		t.Errorf("expected 5ms delay, got %v", delay)
	}

	p.Advance()

	// Third event: 3ms delay
	delay = p.NextDelay()
	if delay != 3*time.Millisecond {
		t.Errorf("expected 3ms delay, got %v", delay)
	}
}

func TestPlayer_SetSpeed(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, Timestamp: 0},
			{Type: EventKey, Timestamp: 10000000}, // 10ms
		},
	}
	p := NewPlayer(rec)

	if p.Speed() != 1.0 {
		t.Errorf("expected default speed 1.0, got %f", p.Speed())
	}

	p.SetSpeed(2.0)
	if p.Speed() != 2.0 {
		t.Errorf("expected speed 2.0, got %f", p.Speed())
	}

	p.Advance()

	// At 2x speed, 10ms delay becomes 5ms
	delay := p.NextDelay()
	if delay != 5*time.Millisecond {
		t.Errorf("expected 5ms at 2x speed, got %v", delay)
	}
}

func TestPlayer_SetSpeed_HalfSpeed(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, Timestamp: 0},
			{Type: EventKey, Timestamp: 10000000}, // 10ms
		},
	}
	p := NewPlayer(rec)
	p.SetSpeed(0.5)
	p.Advance()

	// At 0.5x speed, 10ms delay becomes 20ms
	delay := p.NextDelay()
	if delay != 20*time.Millisecond {
		t.Errorf("expected 20ms at 0.5x speed, got %v", delay)
	}
}

func TestPlayer_TotalDuration(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, Timestamp: 0},
			{Type: EventKey, Timestamp: 1500000000}, // 1.5s
		},
	}
	p := NewPlayer(rec)

	dur := p.TotalDuration()
	if dur != 1500*time.Millisecond {
		t.Errorf("expected 1.5s total duration, got %v", dur)
	}
}

func TestPlayer_TotalDuration_Empty(t *testing.T) {
	rec := &Recording{Version: 1, Events: []RecordedEvent{}}
	p := NewPlayer(rec)
	if p.TotalDuration() != 0 {
		t.Error("expected 0 duration for empty recording")
	}
}

func TestPlayer_Play(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, KeyRune: 'a', Timestamp: 0},
			{Type: EventKey, KeyRune: 'b', Timestamp: 1000000}, // 1ms
			{Type: EventKey, KeyRune: 'c', Timestamp: 2000000}, // 1ms
		},
	}
	p := NewPlayer(rec)
	p.SetSpeed(100.0) // fast playback for test

	var received []rune
	p.Play(func(e RecordedEvent) bool {
		received = append(received, e.KeyRune)
		return true
	})

	if len(received) != 3 {
		t.Fatalf("expected 3 events played, got %d", len(received))
	}
	if received[0] != 'a' || received[1] != 'b' || received[2] != 'c' {
		t.Errorf("expected abc, got %q", string(received))
	}
}

func TestPlayer_Play_StopEarly(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, KeyRune: 'a'},
			{Type: EventKey, KeyRune: 'b'},
			{Type: EventKey, KeyRune: 'c'},
		},
	}
	p := NewPlayer(rec)

	count := 0
	p.PlayFast(func(e RecordedEvent) bool {
		count++
		if count >= 2 {
			return false // stop early
		}
		return true
	})

	if count != 2 {
		t.Errorf("expected 2 events before stopping, got %d", count)
	}
}

func TestPlayer_PlayFast(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events: []RecordedEvent{
			{Type: EventKey, KeyRune: 'a', Timestamp: 0},
			{Type: EventKey, KeyRune: 'b', Timestamp: 5000000000}, // 5s
		},
	}
	p := NewPlayer(rec)

	var received []rune
	p.PlayFast(func(e RecordedEvent) bool {
		received = append(received, e.KeyRune)
		return true
	})

	if len(received) != 2 {
		t.Errorf("expected 2 events, got %d", len(received))
	}
}

func TestPlayer_Events(t *testing.T) {
	events := []RecordedEvent{
		{Type: EventKey, KeyRune: 'x'},
	}
	rec := &Recording{Version: 1, Events: events}
	p := NewPlayer(rec)

	got := p.Events()
	if len(got) != 1 || got[0].KeyRune != 'x' {
		t.Errorf("unexpected events: %+v", got)
	}
}

// --- Round-trip tests ---

func TestRecorder_Player_RoundTrip(t *testing.T) {
	r := NewRecorder()
	r.Start(100, 30)

	r.RecordKey('h', 0, 0)
	r.RecordKey('i', 0, 0)
	r.RecordMouse(5, 10, 1, 0)
	r.RecordResize(120, 40)
	r.RecordFocus(true)
	r.RecordPaste("hello")

	data, err := r.SaveBytes()
	if err != nil {
		t.Fatalf("SaveBytes failed: %v", err)
	}

	player, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("LoadFromBytes failed: %v", err)
	}

	if player.Width() != 100 || player.Height() != 30 {
		t.Errorf("expected 100x30, got %dx%d", player.Width(), player.Height())
	}
	if player.EventCount() != 6 {
		t.Errorf("expected 6 events, got %d", player.EventCount())
	}

	// Verify event types are preserved
	events := player.Events()
	expectedTypes := []EventType{EventKey, EventKey, EventMouse, EventResize, EventFocus, EventPaste}
	for i, et := range expectedTypes {
		if events[i].Type != et {
			t.Errorf("event %d: expected type %d, got %d", i, et, events[i].Type)
		}
	}
}

// --- Concurrency tests ---

func TestRecorder_Concurrent(t *testing.T) {
	r := NewRecorder()
	r.Start(80, 24)

	done := make(chan struct{})
	// Concurrent recorders
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				r.RecordKey('a', 0, 0)
			}
			done <- struct{}{}
		}()
	}

	// Concurrent reader
	go func() {
		for j := 0; j < 100; j++ {
			_ = r.Events()
			_ = r.EventCount()
		}
		done <- struct{}{}
	}()

	for i := 0; i < 6; i++ {
		<-done
	}

	if r.EventCount() != 500 {
		t.Errorf("expected 500 events, got %d", r.EventCount())
	}
}

func TestPlayer_Concurrent(t *testing.T) {
	rec := &Recording{
		Version: 1,
		Events:  make([]RecordedEvent, 100),
	}
	for i := range rec.Events {
		rec.Events[i] = RecordedEvent{Type: EventKey, KeyRune: 'a', Timestamp: int64(i)}
	}
	p := NewPlayer(rec)

	done := make(chan struct{})
	// Concurrent Next/Advance
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 20; j++ {
				p.Next()
				p.Advance()
				p.Remaining()
				p.HasNext()
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 5; i++ {
		<-done
	}
}
