package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"
	"time"
)

// --- Recorder tests ---

func TestRecorderStartStop(t *testing.T) {
	r := NewRecorder()

	if r.IsActive() {
		t.Error("Recorder should be inactive initially")
	}

	r.Start()
	if !r.IsActive() {
		t.Error("Recorder should be active after Start")
	}

	r.Stop()
	if r.IsActive() {
		t.Error("Recorder should be inactive after Stop")
	}
}

func TestRecorderStartResets(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordUserInput("first")
	r.Stop()

	// Restart — should clear old entries
	r.Start()
	if r.EntryCount() != 0 {
		t.Errorf("EntryCount after restart = %d, want 0", r.EntryCount())
	}
}

func TestRecordUserInput(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordUserInput("hello world")

	entries := r.Entries()
	if len(entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(entries))
	}

	if entries[0].Type != string(EntryUserInput) {
		t.Errorf("Type = %q, want %q", entries[0].Type, EntryUserInput)
	}

	var data UserInputData
	if err := json.Unmarshal(entries[0].Data, &data); err != nil {
		t.Fatalf("Unmarshal data: %v", err)
	}
	if data.Text != "hello world" {
		t.Errorf("Text = %q, want 'hello world'", data.Text)
	}
}

func TestRecordAIResponse(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordAIResponse("delta1")
	r.RecordAIResponse("delta2")

	entries := r.Entries()
	if len(entries) != 2 {
		t.Fatalf("Entries len = %d, want 2", len(entries))
	}

	for i, e := range entries {
		if e.Type != string(EntryAIResponse) {
			t.Errorf("Entry %d Type = %q, want %q", i, e.Type, EntryAIResponse)
		}
		var data AIResponseData
		if err := json.Unmarshal(e.Data, &data); err != nil {
			t.Fatalf("Entry %d unmarshal: %v", i, err)
		}
		want := "delta1"
		if i == 1 {
			want = "delta2"
		}
		if data.Delta != want {
			t.Errorf("Entry %d Delta = %q, want %q", i, data.Delta, want)
		}
	}
}

func TestRecordToolCall(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordToolCall("read_file", `{"path": "/tmp/test.go"}`)

	entries := r.Entries()
	if len(entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(entries))
	}

	if entries[0].Type != string(EntryToolCall) {
		t.Errorf("Type = %q, want %q", entries[0].Type, EntryToolCall)
	}

	var data ToolCallData
	if err := json.Unmarshal(entries[0].Data, &data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if data.Name != "read_file" {
		t.Errorf("Name = %q, want read_file", data.Name)
	}
	if data.Args != `{"path": "/tmp/test.go"}` {
		t.Errorf("Args = %q", data.Args)
	}
}

func TestRecordError(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordError(errors.New("connection refused"))

	entries := r.Entries()
	if len(entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(entries))
	}

	if entries[0].Type != string(EntryError) {
		t.Errorf("Type = %q, want %q", entries[0].Type, EntryError)
	}

	var data ErrorData
	if err := json.Unmarshal(entries[0].Data, &data); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if data.Message != "connection refused" {
		t.Errorf("Message = %q, want 'connection refused'", data.Message)
	}
}

func TestRecordNilError(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordError(nil) // should not panic

	entries := r.Entries()
	if len(entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(entries))
	}
	var data ErrorData
	json.Unmarshal(entries[0].Data, &data)
	if data.Message != "" {
		t.Errorf("Message = %q, want empty", data.Message)
	}
}

func TestRecordWhenInactive(t *testing.T) {
	r := NewRecorder()
	// Don't call Start
	r.RecordUserInput("test")

	if r.EntryCount() != 0 {
		t.Errorf("EntryCount = %d, want 0 when inactive", r.EntryCount())
	}
}

func TestTimestampOrdering(t *testing.T) {
	r := NewRecorder()
	r.Start()

	r.RecordUserInput("a")
	time.Sleep(2 * time.Millisecond)
	r.RecordAIResponse("b")
	time.Sleep(2 * time.Millisecond)
	r.RecordToolCall("tool", "{}")

	entries := r.Entries()
	if len(entries) != 3 {
		t.Fatalf("Entries len = %d, want 3", len(entries))
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			t.Errorf("Entry %d timestamp before entry %d", i, i-1)
		}
	}
}

// --- Save/Load round-trip ---

func TestSaveLoadRoundTrip(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.RecordUserInput("test input")
	r.RecordAIResponse("AI response")
	r.RecordToolCall("search", `{"q": "hello"}`)
	r.RecordError(errors.New("oops"))
	r.Stop()

	// Save
	var buf bytes.Buffer
	if err := r.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Verify JSON structure
	var file RecordedFile
	if err := json.Unmarshal(buf.Bytes(), &file); err != nil {
		t.Fatalf("Unmarshal saved file: %v", err)
	}
	if file.Version != 1 {
		t.Errorf("Version = %d, want 1", file.Version)
	}
	if len(file.Entries) != 4 {
		t.Errorf("Entries len = %d, want 4", len(file.Entries))
	}

	// Load
	loaded, err := LoadRecording(&buf)
	if err != nil {
		t.Fatalf("LoadRecording: %v", err)
	}
	if len(loaded) != 4 {
		t.Fatalf("Loaded entries len = %d, want 4", len(loaded))
	}

	// Verify types preserved
	expectedTypes := []string{
		string(EntryUserInput),
		string(EntryAIResponse),
		string(EntryToolCall),
		string(EntryError),
	}
	for i, e := range loaded {
		if e.Type != expectedTypes[i] {
			t.Errorf("Entry %d Type = %q, want %q", i, e.Type, expectedTypes[i])
		}
	}

	// Verify user input data round-trip
	var uid UserInputData
	json.Unmarshal(loaded[0].Data, &uid)
	if uid.Text != "test input" {
		t.Errorf("User input text = %q, want 'test input'", uid.Text)
	}

	// Verify tool call data round-trip
	var tcd ToolCallData
	json.Unmarshal(loaded[2].Data, &tcd)
	if tcd.Name != "search" {
		t.Errorf("Tool name = %q, want 'search'", tcd.Name)
	}
}

func TestSaveLoadEmpty(t *testing.T) {
	r := NewRecorder()
	r.Start()
	r.Stop()

	var buf bytes.Buffer
	if err := r.Save(&buf); err != nil {
		t.Fatalf("Save empty: %v", err)
	}

	loaded, err := LoadRecording(&buf)
	if err != nil {
		t.Fatalf("LoadRecording empty: %v", err)
	}
	if len(loaded) != 0 {
		t.Errorf("Loaded entries len = %d, want 0", len(loaded))
	}
}

// --- Player tests ---

func TestPlayerNext(t *testing.T) {
	entries := []RecordingEntry{
		{Type: string(EntryUserInput), Data: json.RawMessage(`{}`)},
		{Type: string(EntryAIResponse), Data: json.RawMessage(`{}`)},
		{Type: string(EntryToolCall), Data: json.RawMessage(`{}`)},
	}
	p := NewPlayer(entries)

	for i := 0; i < 3; i++ {
		if !p.HasNext() {
			t.Fatalf("HasNext() = false at iteration %d, want true", i)
		}
		entry, err := p.Next()
		if err != nil {
			t.Fatalf("Next() error at iteration %d: %v", i, err)
		}
		if entry.Type != entries[i].Type {
			t.Errorf("Entry %d Type = %q, want %q", i, entry.Type, entries[i].Type)
		}
	}

	// After all entries
	if p.HasNext() {
		t.Error("HasNext() = true after all entries, want false")
	}
	_, err := p.Next()
	if err != io.EOF {
		t.Errorf("Next() error = %v, want io.EOF", err)
	}
}

func TestPlayerReset(t *testing.T) {
	entries := []RecordingEntry{
		{Type: "a"},
		{Type: "b"},
		{Type: "c"},
	}
	p := NewPlayer(entries)

	// Consume all
	for i := 0; i < 3; i++ {
		p.Next()
	}
	if p.HasNext() {
		t.Error("Should be at end")
	}

	// Reset
	p.Reset()
	if !p.HasNext() {
		t.Error("Should have entries after Reset")
	}
	if p.CurrentIndex() != 0 {
		t.Errorf("CurrentIndex after Reset = %d, want 0", p.CurrentIndex())
	}

	// First entry should be "a"
	entry, err := p.Next()
	if err != nil {
		t.Fatalf("Next after Reset: %v", err)
	}
	if entry.Type != "a" {
		t.Errorf("First entry after Reset = %q, want 'a'", entry.Type)
	}
}

func TestPlayerProgress(t *testing.T) {
	entries := make([]RecordingEntry, 4)
	p := NewPlayer(entries)

	if p.Progress() != 0.0 {
		t.Errorf("Initial Progress = %f, want 0.0", p.Progress())
	}

	p.Next()
	if p.Progress() != 0.25 {
		t.Errorf("After 1/4 Progress = %f, want 0.25", p.Progress())
	}

	p.Next()
	if p.Progress() != 0.5 {
		t.Errorf("After 2/4 Progress = %f, want 0.5", p.Progress())
	}

	p.Next()
	p.Next()
	if p.Progress() != 1.0 {
		t.Errorf("After 4/4 Progress = %f, want 1.0", p.Progress())
	}
}

func TestPlayerProgressEmpty(t *testing.T) {
	p := NewPlayer(nil)
	if p.Progress() != 0.0 {
		t.Errorf("Empty Progress = %f, want 0.0", p.Progress())
	}
}

func TestPlayerSetSpeed(t *testing.T) {
	p := NewPlayer(nil)

	if p.Speed() != 1.0 {
		t.Errorf("Default Speed = %f, want 1.0", p.Speed())
	}

	p.SetSpeed(2.0)
	if p.Speed() != 2.0 {
		t.Errorf("Speed = %f, want 2.0", p.Speed())
	}

	p.SetSpeed(0.5)
	if p.Speed() != 0.5 {
		t.Errorf("Speed = %f, want 0.5", p.Speed())
	}

	// Zero or negative should default to 1.0
	p.SetSpeed(0)
	if p.Speed() != 1.0 {
		t.Errorf("Speed after SetSpeed(0) = %f, want 1.0", p.Speed())
	}

	p.SetSpeed(-1)
	if p.Speed() != 1.0 {
		t.Errorf("Speed after SetSpeed(-1) = %f, want 1.0", p.Speed())
	}
}

func TestPlayerPeek(t *testing.T) {
	entries := []RecordingEntry{
		{Type: "first"},
		{Type: "second"},
	}
	p := NewPlayer(entries)

	// Peek without advancing
	e := p.Peek()
	if e == nil {
		t.Fatal("Peek returned nil")
	}
	if e.Type != "first" {
		t.Errorf("Peek Type = %q, want 'first'", e.Type)
	}

	// Peek again — still at same position
	e = p.Peek()
	if e.Type != "first" {
		t.Errorf("Peek again Type = %q, want 'first'", e.Type)
	}

	// Advance
	p.Next()
	e = p.Peek()
	if e == nil {
		t.Fatal("Peek returned nil after Next")
	}
	if e.Type != "second" {
		t.Errorf("Peek after Next Type = %q, want 'second'", e.Type)
	}

	// At end
	p.Next()
	e = p.Peek()
	if e != nil {
		t.Error("Peek at end should return nil")
	}
}

func TestPlayerDuration(t *testing.T) {
	base := time.Now()
	entries := []RecordingEntry{
		{Timestamp: base},
		{Timestamp: base.Add(5 * time.Second)},
		{Timestamp: base.Add(12 * time.Second)},
	}
	p := NewPlayer(entries)

	d := p.Duration()
	if d != 12*time.Second {
		t.Errorf("Duration = %v, want 12s", d)
	}
}

func TestPlayerDurationEmpty(t *testing.T) {
	p := NewPlayer(nil)
	if p.Duration() != 0 {
		t.Errorf("Empty Duration = %v, want 0", p.Duration())
	}

	p2 := NewPlayer([]RecordingEntry{{Timestamp: time.Now()}})
	if p2.Duration() != 0 {
		t.Errorf("Single entry Duration = %v, want 0", p2.Duration())
	}
}

func TestPlayerDelay(t *testing.T) {
	base := time.Now()
	entries := []RecordingEntry{
		{Timestamp: base},
		{Timestamp: base.Add(2 * time.Second)},
		{Timestamp: base.Add(5 * time.Second)},
	}
	p := NewPlayer(entries)

	// Before first Next, delay = 0
	if d := p.Delay(); d != 0 {
		t.Errorf("Initial Delay = %v, want 0", d)
	}

	// After first Next, delay to second = 2s
	p.Next()
	d := p.Delay()
	if d != 2*time.Second {
		t.Errorf("Delay after 1st = %v, want 2s", d)
	}

	// With 2x speed, delay should be halved
	p.SetSpeed(2.0)
	d = p.Delay()
	if d != 1*time.Second {
		t.Errorf("Delay at 2x speed = %v, want 1s", d)
	}
}

// --- Integration: full record → save → load → play cycle ---

func TestRecordSaveLoadPlay(t *testing.T) {
	// Record
	r := NewRecorder()
	r.Start()
	r.RecordUserInput("What is Go?")
	r.RecordAIResponse("Go is a programming language.")
	r.RecordToolCall("search_docs", `{"q": "go"}`)
	r.RecordError(errors.New("timeout"))
	r.Stop()

	// Save
	var buf bytes.Buffer
	if err := r.Save(&buf); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load
	loaded, err := LoadRecording(&buf)
	if err != nil {
		t.Fatalf("LoadRecording: %v", err)
	}

	// Play
	p := NewPlayer(loaded)
	if p.TotalEntries() != 4 {
		t.Fatalf("TotalEntries = %d, want 4", p.TotalEntries())
	}

	var playedTypes []string
	for p.HasNext() {
		entry, err := p.Next()
		if err != nil {
			t.Fatalf("Next: %v", err)
		}
		playedTypes = append(playedTypes, entry.Type)
	}

	expected := []string{
		string(EntryUserInput),
		string(EntryAIResponse),
		string(EntryToolCall),
		string(EntryError),
	}
	for i, typ := range playedTypes {
		if typ != expected[i] {
			t.Errorf("Played %d = %q, want %q", i, typ, expected[i])
		}
	}
}
